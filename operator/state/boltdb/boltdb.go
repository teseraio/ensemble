package boltdb

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"

	gproto "github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"

	"github.com/boltdb/bolt"
	"github.com/teseraio/ensemble/lib/uuid"
	"github.com/teseraio/ensemble/operator/proto"
)

var (
	// clustersBucket    = []byte("clusters")
	deploymentsBucket = []byte("deployments")
	componentsBucket  = []byte("components")

	// deployment key
	metaKey = []byte("meta")
	depKey  = []byte("meta")

	// keys for the components table
	nextAppliedKey = []byte("next")
	lastSeqKey     = []byte("lastSeq")
)

// Factory is the factory method for the Boltdb backend
func Factory(config map[string]interface{}) (*BoltDB, error) {
	pathRaw, ok := config["path"]
	if !ok {
		return nil, fmt.Errorf("field 'path' not found")
	}
	path, ok := pathRaw.(string)
	if !ok {
		return nil, fmt.Errorf("field 'path' is not string")
	}

	db, err := bolt.Open(path, 0600, nil)
	if err != nil {
		return nil, err
	}
	b := &BoltDB{
		path:   path,
		db:     db,
		queue2: newTaskQueue(),
	}
	if err := b.initialize(); err != nil {
		return nil, err
	}
	return b, nil
}

// BoltDB is a boltdb state implementation
type BoltDB struct {
	path       string
	db         *bolt.DB
	queue2     *taskQueue
	waitChLock sync.Mutex
	waitCh     map[string]chan struct{}
}

func (b *BoltDB) Wait(id string) chan struct{} {
	b.waitChLock.Lock()
	defer b.waitChLock.Unlock()

	if b.waitCh == nil {
		b.waitCh = map[string]chan struct{}{}
	}
	ch := make(chan struct{})
	b.waitCh[id] = ch

	return ch
}

func (b *BoltDB) initialize() error {
	buckets := [][]byte{
		deploymentsBucket,
		componentsBucket,
	}
	err := b.db.Update(func(tx *bolt.Tx) error {
		for _, i := range buckets {
			if _, err := tx.CreateBucketIfNotExists(i); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	err = b.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(deploymentsBucket)

		return bkt.ForEach(func(k, v []byte) error {
			clusterBkt := bkt.Bucket(k)

			comps := clusterBkt.Bucket([]byte("components"))
			if comps == nil {
				return nil
			}
			reqs := clusterBkt.Bucket([]byte("requests"))
			if reqs == nil {
				return nil
			}

			// get the next key
			num, err := getSeqNumber(reqs, nextAppliedKey)
			if err != nil {
				return nil
			}
			if num == 0 {
				return nil
			}
			data := reqs.Get(seqID(num))
			if data != nil {
				// there is a pending item for this cluster
				comp, err := b.getComponentFromBucket(string(data), comps)
				if err != nil {
					return err
				}
				b.addTask(string(k), comp)
			}
			return nil
		})
	})
	if err != nil {
		return err
	}
	return nil
}

// Close implements the BoltDB interface
func (b *BoltDB) Close() error {
	return b.db.Close()
}

func putSeqNumber(bkt *bolt.Bucket, key []byte, num int64) error {
	if err := bkt.Put(key, []byte(fmt.Sprintf("%d", num))); err != nil {
		return err
	}
	return nil
}

func getSeqNumber(bkt *bolt.Bucket, key []byte) (int64, error) {
	seq := 0
	var err error
	if raw := bkt.Get(key); raw != nil {
		if seq, err = strconv.Atoi(string(raw)); err != nil {
			return 0, err
		}
	}
	return int64(seq), nil
}

func seqID2(seq int) []byte {
	return []byte(fmt.Sprintf("seq-%d", seq))
}

func seqID(seq int64) []byte {
	return []byte(fmt.Sprintf("seq-%d", seq))
}

func (b *BoltDB) GetComponentByID2(clusterName, ref string, sequence int64) (*proto.Component, error) {
	tx, err := b.db.Begin(false)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	bkt := tx.Bucket(deploymentsBucket).Bucket([]byte(clusterName))
	if bkt == nil {
		return nil, fmt.Errorf("bucket not found %s", clusterName)
	}
	comps := bkt.Bucket([]byte("components"))
	if err != nil {
		return nil, fmt.Errorf("components bucket not found")
	}

	comp, err := b.getComponentFromBucket(fmt.Sprintf("%s#%d", ref, sequence), comps)
	if err != nil {
		return nil, err
	}
	return comp, nil
}

func (b *BoltDB) getComponentFromBucket(name string, bkt *bolt.Bucket) (*proto.Component, error) {
	var ref string
	var seqID int

	spl := strings.Split(string(name), "#")
	if len(spl) == 2 {
		seqIDRaw, err := strconv.Atoi(spl[1])
		if err != nil {
			return nil, err
		}
		seqID = seqIDRaw
		ref = spl[0]
	} else {
		// pick the latest applied component (is at least the last valid one)
		num, err := getSeqNumber(bkt, lastSeqKey)
		if err != nil {
			return nil, err
		}
		if num == 0 {
			num = 1
		}
		ref = spl[0]
		seqID = int(num)
	}

	compBkt := bkt.Bucket([]byte(ref))
	if compBkt == nil {
		return nil, fmt.Errorf("ref bucket not found: %s", ref)
	}

	comp := proto.Component{}
	if err := dbGet(compBkt, seqID2(seqID), &comp); err != nil {
		return nil, err
	}
	return &comp, nil
}

func (b *BoltDB) updateComponentStatus(compsBkt *bolt.Bucket, compRef string, status proto.Component_Status) (*proto.Component, error) {
	spl := strings.Split(string(compRef), "#")
	ref, seqIDRaw := spl[0], spl[1]
	seqID, err := strconv.Atoi(seqIDRaw)
	if err != nil {
		return nil, err
	}

	compBkt := compsBkt.Bucket([]byte(ref))
	if compBkt == nil {
		return nil, fmt.Errorf("ref bucket not found %s", ref)
	}
	key := seqID2(seqID)

	comp := proto.Component{}
	if err := dbGet(compBkt, key, &comp); err != nil {
		return nil, err
	}
	// finalize this one
	comp.Status = status
	if err := dbPut(compBkt, key, &comp); err != nil {
		return nil, err
	}
	return &comp, nil
}

func (b *BoltDB) GetHistory(deploymentID string) ([]*proto.Component, error) {
	tx, err := b.db.Begin(false)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	bkt := tx.Bucket(deploymentsBucket).Bucket([]byte(deploymentID))
	if bkt == nil {
		return nil, fmt.Errorf("deployment bucket for %s not found", deploymentID)
	}
	comps := bkt.Bucket([]byte("components"))
	if err != nil {
		return nil, fmt.Errorf("components bucket not found")
	}
	reqs := bkt.Bucket([]byte("requests"))
	if reqs == nil {
		return nil, fmt.Errorf("requests bucket not found")
	}

	result := []*proto.Component{}
	c := reqs.Cursor()
	for k, v := c.Seek([]byte("seq-")); k != nil; k, v = c.Next() {
		component, err := b.getComponentFromBucket(string(v), comps)
		if err != nil {
			return nil, err
		}
		result = append(result, component)
	}
	return result, nil
}

func (b *BoltDB) GetComponentVersions(deploymentID string, id string) ([]*proto.Component, error) {
	tx, err := b.db.Begin(false)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	bkt := tx.Bucket(deploymentsBucket).Bucket([]byte(deploymentID))
	if bkt == nil {
		return nil, fmt.Errorf("deployment %s not found", deploymentID)
	}
	compsBkt := bkt.Bucket([]byte("components"))
	if compsBkt == nil {
		return nil, fmt.Errorf("components bucket not found")
	}

	compBkt := compsBkt.Bucket([]byte(id))
	if compBkt == nil {
		return nil, fmt.Errorf("bucket for component %s not found", id)
	}

	result := []*proto.Component{}
	err = compBkt.ForEach(func(k, v []byte) error {
		component := proto.Component{}
		if err := gproto.Unmarshal(v, &component); err != nil {
			return err
		}
		result = append(result, &component)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

// GetComponents returns all the available components for the deployment
func (b *BoltDB) GetComponents(deploymentID string) ([]*proto.Component, error) {
	tx, err := b.db.Begin(false)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	bkt := tx.Bucket(deploymentsBucket).Bucket([]byte(deploymentID))
	if bkt == nil {
		return nil, fmt.Errorf("deployment %s not found", deploymentID)
	}
	comps := bkt.Bucket([]byte("components"))
	if err != nil {
		return nil, fmt.Errorf("components bucket not found")
	}

	result := []*proto.Component{}
	err = comps.ForEach(func(k, v []byte) error {
		elem, err := b.readLatestComponent(comps.Bucket(k))
		if err != nil {
			return err
		}
		result = append(result, elem)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (b *BoltDB) Finalize(deploymentID string) error {
	tx, err := b.db.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, ok := b.queue2.finalize(deploymentID); !ok {
		return fmt.Errorf("task not found for deployment %s", deploymentID)
	}

	bkt := tx.Bucket(deploymentsBucket).Bucket([]byte(deploymentID))
	if bkt == nil {
		return fmt.Errorf("deployment %s not found", deploymentID)
	}
	compsBkt := bkt.Bucket([]byte("components"))
	if compsBkt == nil {
		return fmt.Errorf("components bucket not found")
	}
	reqs := bkt.Bucket([]byte("requests"))
	if reqs == nil {
		return fmt.Errorf("requests bucket not found")
	}

	num, err := getSeqNumber(reqs, nextAppliedKey)
	if err != nil {
		return err
	}

	compRef := reqs.Get(seqID(num))
	if compRef == nil {
		return fmt.Errorf("bucket not found for seq %d", num)
	}
	comp, err := b.updateComponentStatus(compsBkt, string(compRef), proto.Component_APPLIED)
	if err != nil {
		return err
	}

	// update the next key to apply
	if err := putSeqNumber(reqs, nextAppliedKey, num+1); err != nil {
		return err
	}

	// check the next one
	data := reqs.Get(seqID(num + 1))
	if data != nil {
		// there is a new one, eval it
		nextComp, err := b.updateComponentStatus(compsBkt, string(data), proto.Component_QUEUED)
		if err != nil {
			return err
		}
		b.addTask(deploymentID, nextComp)
	}
	if err := tx.Commit(); err != nil {
		return err
	}

	b.waitChLock.Lock()
	if ch, ok := b.waitCh[comp.Id]; ok {
		close(ch)
		delete(b.waitCh, comp.Id)
	}
	b.waitChLock.Unlock()

	return nil
}

func (b *BoltDB) addTask(deploymentID string, comp *proto.Component) {
	b.queue2.add(&proto.Task{
		DeploymentID: deploymentID,
		ComponentID:  comp.Id,
		Sequence:     comp.Sequence,
	})
}

func (b *BoltDB) applyComponent(bkt *bolt.Bucket, c *proto.Component) (*proto.Component, error) {
	if c.Id == "" {
		return nil, fmt.Errorf("component id is empty")
	}
	c.Timestamp = ptypes.TimestampNow()

	bkt, err := bkt.CreateBucketIfNotExists([]byte(c.Id))
	if err != nil {
		return nil, err
	}

	nextSeq, err := getLatestSequence(bkt)
	if err != nil {
		return nil, err
	}

	if nextSeq != 1 {
		prev := &proto.Component{}
		if err := dbGet(bkt, seqID(int64(nextSeq-1)), prev); err != nil {
			return nil, err
		}
		if prev.Action == proto.Component_DELETE {
			// a delete object cannot be created again
			return nil, fmt.Errorf("the object was deleted")
		} else if prev.Action == proto.Component_CREATE {
			// if its not delete, we need to make sure we dont try to
			// save the same spec
			if c.Action != proto.Component_DELETE {
				equal, err := proto.Cmp(prev.Spec, c.Spec)
				if err != nil {
					return nil, err
				}
				if equal {
					return nil, nil
				}
			}
		}
	} else {
		if c.Action == proto.Component_DELETE {
			return nil, fmt.Errorf("cannot remove non created object")
		}
	}

	// update the sequence number in the component
	c.Sequence = int64(nextSeq)

	if err := dbPut(bkt, seqID(c.Sequence), c); err != nil {
		return nil, err
	}
	return c, nil
}

func addSequenceComponent2(bkt *bolt.Bucket) (int, bool, error) {
	// get the latest sequence number
	nextSeq, err := getLatestSequence(bkt)
	if err != nil {
		return 0, false, err
	}

	nextKeyToApply, err := getSeqNumber(bkt, nextAppliedKey)
	if err != nil {
		return 0, false, err
	}
	if nextKeyToApply == 0 {
		// first entry, add the key
		if err := putSeqNumber(bkt, nextAppliedKey, 1); err != nil {
			return 0, false, err
		}
		return nextSeq, true, nil
	}

	var addEval bool
	if nextSeq == int(nextKeyToApply) {
		addEval = true
	}
	return nextSeq, addEval, nil
}

func (b *BoltDB) readLatestComponent(bkt *bolt.Bucket) (*proto.Component, error) {
	c := bkt.Cursor()

	var past *proto.Component
	for k, v := c.Last(); k != nil; k, v = c.Prev() {
		component := proto.Component{}
		if err := gproto.Unmarshal(v, &component); err != nil {
			return nil, err
		}
		if component.Status == proto.Component_APPLIED {
			if past == nil {
				return &component, nil
			}
			return past, nil
		}
		past = &component
	}

	// send the latest found
	if past != nil {
		return past, nil
	}
	return nil, fmt.Errorf("not found")
}

func (b *BoltDB) ReadDeployment(id string) (*proto.Component, error) {
	tx, err := b.db.Begin(false)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	return b.readDeploymentCluster(tx, id)
}

func (b *BoltDB) readDeploymentCluster(tx *bolt.Tx, id string) (*proto.Component, error) {
	bkt := tx.Bucket(deploymentsBucket)

	depBkt := bkt.Bucket([]byte(id))
	if depBkt == nil {
		// deployment does not exists
		return nil, nil
	}

	compBkt := depBkt.Bucket([]byte("components"))
	if compBkt == nil {
		// components bucket does not exists
		return nil, nil
	}

	// since the components are included using ULID in lexicographic order,
	// the first item in the components bucket will always be the cluster
	c := compBkt.Cursor()
	k, _ := c.First()

	clusterBkt := compBkt.Bucket(k)
	comp, err := b.readLatestComponent(clusterBkt)
	if err != nil {
		return nil, err
	}
	return comp, nil
}

func (b *BoltDB) NameToDeployment(name string) (string, error) {
	tx, err := b.db.Begin(false)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	return b.nameToDeploymentID(tx, name)
}

func (b *BoltDB) nameToDeploymentID(tx *bolt.Tx, name string) (string, error) {
	var deploymentID string

	err := tx.Bucket(deploymentsBucket).ForEach(func(k, v []byte) error {
		comp, err := b.readDeploymentCluster(tx, string(k))
		if err != nil {
			return err
		}
		if comp.Action == proto.Component_DELETE {
			return nil
		}
		if comp.Name == name {
			deploymentID = string(k)
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	return deploymentID, nil
}

func (b *BoltDB) Apply(comp *proto.Component) (*proto.Component, error) {
	tx, err := b.db.Begin(true)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	msg, err := proto.UnmarshalAny(comp.Spec)
	if err != nil {
		return nil, err
	}
	clusterRef := msg.(proto.ClusterRef)

	isCluster := clusterRef.GetCluster() == ""
	clusterName := ""
	if isCluster {
		clusterName = comp.Name
	} else {
		clusterName = clusterRef.GetCluster()
	}

	deploymentID, err := b.nameToDeploymentID(tx, clusterName)
	if err != nil {
		return nil, err
	}
	if deploymentID == "" {
		// only create a new deployment if its a cluster
		if clusterRef.GetCluster() != "" {
			return nil, fmt.Errorf("cluster not found")
		}
		deploymentID = uuid.UUID()
	}

	bkt, err := tx.Bucket(deploymentsBucket).CreateBucketIfNotExists([]byte(deploymentID))
	if err != nil {
		return nil, err
	}

	for _, name := range []string{"components", "requests"} {
		if _, err := bkt.CreateBucketIfNotExists([]byte(name)); err != nil {
			return nil, err
		}
	}

	// bucket to store the components
	comps := bkt.Bucket([]byte("components"))

	// bucket for the pending requests to be applied
	reqs := bkt.Bucket([]byte("requests"))

	// find the resource with the same name (if any)
	var resourceID string

	err = comps.ForEach(func(k, v []byte) error {
		bkt = comps.Bucket(k)
		component, err := b.readLatestComponent(bkt)
		if err != nil {
			return err
		}
		if component.Action == proto.Component_DELETE {
			return nil
		}
		if component.Name == comp.Name {
			resourceID = string(k)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if resourceID == "" {
		id, err := uuid.ULID()
		if err != nil {
			return nil, err
		}
		resourceID = id
	}

	comp.Id = resourceID

	// check if there is any pending requests to be applied, otherwise,
	// this component is the new item to be queued.
	nextSeqNum, addEval, err := addSequenceComponent2(reqs)
	if err != nil {
		return nil, err
	}

	if addEval {
		comp.Status = proto.Component_QUEUED
	} else {
		comp.Status = proto.Component_PENDING
	}
	comp, err = b.applyComponent(comps, comp)
	if err != nil {
		return nil, err
	}
	if comp == nil {
		return nil, nil
	}

	// add the new entry to requests
	reqVal := fmt.Sprintf("%s#%d", comp.Id, comp.Sequence)
	if err := reqs.Put(seqID2(int(nextSeqNum)), []byte(reqVal)); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	if addEval {
		b.addTask(deploymentID, comp)
	}
	return comp, nil
}

func getLatestSequence(bkt *bolt.Bucket) (int, error) {
	nextSeq := int(1)
	c := bkt.Cursor()
	if k, _ := c.Last(); k != nil {
		var err error
		if nextSeq, err = strconv.Atoi(strings.TrimPrefix(string(k), "seq-")); err != nil {
			return 0, err
		}
		nextSeq++
	}
	return nextSeq, nil
}

func (b *BoltDB) GetTask(ctx context.Context) *proto.Task {
	task := b.queue2.pop(ctx)
	if task == nil {
		return nil
	}
	return task.Task
}

func (b *BoltDB) LoadInstance(cluster, id string) (*proto.Instance, error) {
	tx, err := b.db.Begin(false)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	depsBkt := tx.Bucket(deploymentsBucket)

	// find the sub-bucket for the cluster
	depBkt := depsBkt.Bucket([]byte(cluster))
	if depBkt == nil {
		return nil, fmt.Errorf("bad")
	}

	nodeID := "node-" + id
	instance := proto.Instance{}
	if err := dbGet(depBkt, []byte(nodeID), &instance); err != nil {
		return nil, err
	}
	return &instance, nil
}

func (b *BoltDB) ListDeployments() ([]*proto.Deployment, error) {
	tx, err := b.db.Begin(false)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	depsBkt := tx.Bucket(deploymentsBucket)

	var deps []*proto.Deployment
	depsBkt.ForEach(func(k, v []byte) error {
		dep, err := b.loadDeploymentImpl(tx, depsBkt, string(k), false)
		if err != nil {
			return err
		}
		deps = append(deps, dep)
		return nil
	})
	return deps, nil
}

func (b *BoltDB) LoadDeployment(id string) (*proto.Deployment, error) {
	tx, err := b.db.Begin(false)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	depsBkt := tx.Bucket(deploymentsBucket)

	dep, err := b.loadDeploymentImpl(tx, depsBkt, id, true)
	if err != nil {
		return nil, err
	}
	return dep, nil
}

func (b *BoltDB) loadDeploymentImpl(tx *bolt.Tx, depsBkt *bolt.Bucket, id string, includeInstances bool) (*proto.Deployment, error) {
	// find the sub-bucket for the cluster
	depBkt := depsBkt.Bucket([]byte(id))
	if depBkt == nil {
		return nil, nil
	}

	// load the cluster meta
	c := &proto.Deployment{
		Instances: []*proto.Instance{},
		Id:        id,
	}

	// load the resource to assign the name
	comp, err := b.readDeploymentCluster(tx, id)
	if err != nil {
		return nil, nil
	}
	if err := dbGet(depBkt, depKey, c); err != nil {
		if err == errNotFound {
			c.Name = comp.Name
			return c, nil
		}
		return nil, err
	}

	// load the nodes under node-<id>
	if includeInstances {
		nodeCursor := depBkt.Cursor()
		for k, _ := nodeCursor.First(); k != nil; k, _ = nodeCursor.Next() {
			if !strings.HasPrefix(string(k), "node-") {
				continue
			}
			n := &proto.Instance{}
			if err := dbGet(depBkt, k, n); err != nil {
				return nil, err
			}
			if n.Status != proto.Instance_OUT {
				c.Instances = append(c.Instances, n)
			} else {
				// instance is out
			}
		}
	}
	return c, nil
}

func (b *BoltDB) UpdateDeployment(d *proto.Deployment) error {
	tx, err := b.db.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	depsBkt := tx.Bucket(deploymentsBucket)

	// find the sub-bucket for the cluster
	depBkt, err := depsBkt.CreateBucketIfNotExists([]byte(d.Id))
	if err != nil {
		return err
	}
	dd := d.Copy()
	dd.Instances = nil

	if err := dbPut(depBkt, depKey, dd); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

// UpsertNode implements the BoltDB interface
func (b *BoltDB) UpsertNode(n *proto.Instance) error {
	tx, err := b.db.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	depsBkt := tx.Bucket(deploymentsBucket)

	depID, err := b.nameToDeploymentID(tx, n.ClusterName)
	if err != nil {
		return err
	}
	// find the sub-bucket for the cluster
	depBkt := depsBkt.Bucket([]byte(depID))
	if depBkt == nil {
		return fmt.Errorf("deployment does not exists")
	}

	// upsert under node-<id>
	id := "node-" + n.ID
	if err := dbPut(depBkt, []byte(id), n); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

func dbPut(b *bolt.Bucket, id []byte, m gproto.Message) error {
	raw, err := gproto.Marshal(m)
	if err != nil {
		return err
	}
	if err := b.Put(id, raw); err != nil {
		return err
	}
	return err
}

var errNotFound = fmt.Errorf("not found")

func dbGet(b *bolt.Bucket, id []byte, m gproto.Message) error {
	raw := b.Get(id)
	if raw == nil {
		return errNotFound
	}
	if err := gproto.Unmarshal(raw, m); err != nil {
		return err
	}
	return nil
}

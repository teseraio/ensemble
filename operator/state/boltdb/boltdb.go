package boltdb

import (
	"bytes"
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

	componentsBucket = []byte("components")

	// versioned indexes
	//clusterIndex = []byte("indx_cluster")

	//indexKey = []byte("index")
	seqKey = []byte("seq")

	// deployment key
	depKey = []byte("meta")

	// keys for the components table
	lastAppliedKey = []byte("meta")
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
		db:     db,
		queue:  newTaskQueue(),
		queue2: newTaskQueue2(),
	}
	if err := b.initialize(); err != nil {
		return nil, err
	}
	return b, nil
}

// BoltDB is a boltdb state implementation
type BoltDB struct {
	db     *bolt.DB
	queue  *taskQueue
	queue2 *taskQueue2

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
					panic(err)
				}
				b.addTask(string(k), comp)
			}
			return nil
		})
	})
	if err != nil {
		return err
	}

	/*
		// load the indexes into the task
		err = b.db.View(func(tx *bolt.Tx) error {
			bkt := tx.Bucket(componentsBucket)

			return bkt.ForEach(func(k, v []byte) error {
				compBkt := bkt.Bucket(k)

				return compBkt.ForEach(func(k, v []byte) error {
					bkt := compBkt.Bucket(k)
					seqBkt := bkt.Bucket(seqKey)

					seq, err := getSeqNumber(bkt, lastAppliedKey)
					if err != nil {
						return err
					}
					comp := proto.Component{}
					if err := dbGet(seqBkt, seqID(seq), &comp); err != nil {
						return err
					}
					if comp.Status == proto.Component_PENDING {
						clusterID, err := proto.ClusterIDFromComponent(&comp)
						if err != nil {
							return err
						}
						b.queue.add(clusterID, &comp)
					}
					return nil
				})
			})
		})
		if err != nil {
			return err
		}
	*/

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

func (b *BoltDB) getComponentIndexes(namespace, name string) (int64, int64, error) {
	tx, err := b.db.Begin(false)
	if err != nil {
		return 0, 0, err
	}
	defer tx.Rollback()

	componentsBkt := tx.Bucket(componentsBucket)
	namespaceBkt := componentsBkt.Bucket([]byte(namespace))
	if namespaceBkt == nil {
		return 0, 0, fmt.Errorf("namespace %s does not exists", namespace)
	}
	componentBucket := namespaceBkt.Bucket([]byte(name))
	if componentBucket == nil {
		return 0, 0, fmt.Errorf("component %s does not exists", name)
	}

	lastSeq, err := getSeqNumber(componentBucket, lastSeqKey)
	if err != nil {
		return 0, 0, err
	}
	lastAppliedSeq, err := getSeqNumber(componentBucket, lastAppliedKey)
	if err != nil {
		return 0, 0, err
	}
	return lastSeq, lastAppliedSeq, nil
}

func (b *BoltDB) Apply222(c *proto.Component) (int64, error) {
	namespace := []byte(getProtoNamespace(c))

	tx, err := b.db.Begin(true)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	componentsBkt := tx.Bucket(componentsBucket)

	// append current timestamp
	c.Timestamp = ptypes.TimestampNow()
	if c.Id == "" {
		return 0, fmt.Errorf("empty id")
	}

	// create the bucket to store this specific namespace
	namespaceBkt, err := componentsBkt.CreateBucketIfNotExists(namespace)
	if err != nil {
		return 0, err
	}

	// find the bucket for the specific id
	componentBucket, err := namespaceBkt.CreateBucketIfNotExists([]byte(c.Name))
	if err != nil {
		return 0, err
	}

	// get the last sequence number for the component
	lastSeq, err := getSeqNumber(componentBucket, lastSeqKey)
	if err != nil {
		return 0, err
	}

	// reference to the bucket to store the historical sequences
	seqBkt, err := componentBucket.CreateBucketIfNotExists(seqKey)
	if err != nil {
		return 0, err
	}

	// get the current version
	old := proto.Component{}
	if lastSeq != 0 {
		if err := dbGet(seqBkt, seqID(int64(lastSeq)), &old); err != nil {
			return 0, err
		}
		if old.Action == proto.Component_DELETE {
			// we only expect create object
			if c.Action != proto.Component_CREATE {
				return 0, fmt.Errorf("the component is deleted, only create expected but found %s", old.Action)
			}
		} else if old.Action == proto.Component_CREATE {
			// if its not delete, we need to make sure we dont try to
			// save the same spec
			if c.Action != proto.Component_DELETE {
				if bytes.Equal(old.Spec.Value, c.Spec.Value) {
					return 0, nil
				}
			}
		}
	} else {
		if c.Action == proto.Component_DELETE {
			return 0, fmt.Errorf("cannot remove non created object")
		}
	}

	// update the sequence number in the component
	c.Sequence = int64(lastSeq) + 1

	if old.Status != proto.Component_PENDING && old.Status != proto.Component_QUEUED {
		// push this component to the pending queue
		clusterID, err := proto.ClusterIDFromComponent(c)
		if err != nil {
			return 0, err
		}
		b.queue.add(clusterID, c)

		c.Status = proto.Component_PENDING

		// update the new componetn being applied
		if err := putSeqNumber(componentBucket, lastAppliedKey, c.Sequence); err != nil {
			return 0, err
		}
	} else {
		c.Status = proto.Component_QUEUED
	}

	// store the object and the next sequence index
	if err := dbPut(seqBkt, seqID(c.Sequence), c); err != nil {
		return 0, err
	}
	if err := putSeqNumber(componentBucket, lastSeqKey, c.Sequence); err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return c.Sequence, nil
}

func (b *BoltDB) GetComponentByID(namespace, name string, compID string) (*proto.Component, error) {
	// TODO: Change dblayout to make this more optimal
	tx, err := b.db.Begin(false)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	componentsBkt := tx.Bucket(componentsBucket)
	namespaceBkt := componentsBkt.Bucket([]byte(namespace))

	compBkt := namespaceBkt.Bucket([]byte(name))
	seqBkt := compBkt.Bucket(seqKey)

	res := &proto.Component{}
	seqBkt.ForEach(func(k, v []byte) error {
		if res.Id != "" {
			return nil
		}
		comp := proto.Component{}
		if err := gproto.Unmarshal(v, &comp); err != nil {
			return err
		}
		if comp.Id == compID {
			res = &comp
		}
		return nil
	})
	if res == nil {
		return nil, fmt.Errorf("not found")
	}
	return res, nil
}

func (b *BoltDB) GetComponent(namespace, name string, sequence int64) (*proto.Component, error) {
	tx, err := b.db.Begin(false)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	componentsBkt := tx.Bucket(componentsBucket)
	namespaceBkt := componentsBkt.Bucket([]byte(namespace))

	compBkt := namespaceBkt.Bucket([]byte(name))
	seqBkt := compBkt.Bucket(seqKey)

	comp := proto.Component{}
	if err := dbGet(seqBkt, seqID(sequence), &comp); err != nil {
		return nil, err
	}

	return &comp, nil
}

func seqID2(seq int) []byte {
	return []byte(fmt.Sprintf("seq-%d", seq))
}

func seqID(seq int64) []byte {
	return []byte(fmt.Sprintf("seq-%d", seq))
}

func (b *BoltDB) GetTask(ctx context.Context) *proto.Component {
	tt := b.queue.pop(ctx)
	if tt == nil {
		return nil
	}
	return tt.Component
}

func (b *BoltDB) GetPending(id string) (*proto.Component, error) {
	tt, ok := b.queue.get(id)
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return tt.Component, nil
}

func getProtoNamespace(c *proto.Component) string {
	return strings.Replace(c.Spec.TypeUrl, ".", "-", -1)
}

func (b *BoltDB) Finalizxxxe(id string) error {
	tt, ok := b.queue.finalize(id)
	if !ok {
		return fmt.Errorf("task for id '%s' not found", id)
	}

	namespace := []byte(getProtoNamespace(tt.Component))

	tx, err := b.db.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	componentsBkt := tx.Bucket(componentsBucket)
	namespaceBkt := componentsBkt.Bucket(namespace)
	componentBkt := namespaceBkt.Bucket([]byte(tt.Component.Name))

	seq, err := getSeqNumber(componentBkt, lastAppliedKey)
	if err != nil {
		return err
	}

	seqBkt := componentBkt.Bucket(seqKey)

	current := proto.Component{}
	if err := dbGet(seqBkt, seqID(int64(seq)), &current); err != nil {
		return err
	}
	if current.Id != tt.Component.Id {
		return fmt.Errorf("wrong component")
	}
	if current.Status != proto.Component_PENDING {
		return fmt.Errorf("state should be pending")
	}

	// change status to complete
	current.Status = proto.Component_APPLIED
	if err := dbPut(seqBkt, seqID(seq), &current); err != nil {
		return err
	}

	// check if the next item is available
	nextComp := proto.Component{}
	if err := dbGet(seqBkt, seqID(seq+1), &nextComp); err != nil {
		if err != errNotFound {
			return err
		}
	} else {
		b.queue.add(tt.clusterID, &nextComp)

		// update the sequence number
		if err := putSeqNumber(componentBkt, lastAppliedKey, nextComp.Sequence); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	// notify any wait channels
	b.waitChLock.Lock()
	if ch, ok := b.waitCh[tt.Component.Id]; ok {
		close(ch)
		delete(b.waitCh, tt.Component.Id)
	}
	b.waitChLock.Unlock()

	return nil
}

func (b *BoltDB) GetComponentByID2(clusterName, ref string, sequence int64) (*proto.Component, error) {
	tx, err := b.db.Begin(false)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	fmt.Println("__ REF F__")
	fmt.Println(ref)
	fmt.Println(clusterName)

	bkt := tx.Bucket(deploymentsBucket).Bucket([]byte(clusterName))
	if bkt == nil {
		panic("bad1 1")
	}

	comps := bkt.Bucket([]byte("components"))
	if err != nil {
		panic("bad3")
	}

	comp, err := b.getComponentFromBucket(fmt.Sprintf("%s#%d", ref, sequence), comps)
	if err != nil {
		panic(err)
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
			panic(err)
		}
		seqID = seqIDRaw
		ref = spl[0]
	} else {
		// pick the latest applied component (is at least the last valid one)
		num, err := getSeqNumber(bkt, lastSeqKey)
		if err != nil {
			panic(err)
		}
		if num == 0 {
			num = 1
		}
		ref = spl[0]
		seqID = int(num)
	}

	compBkt := bkt.Bucket([]byte(ref))
	if compBkt == nil {
		panic("bad")
	}

	comp := proto.Component{}
	if err := dbGet(compBkt, seqID2(seqID), &comp); err != nil {
		panic(err)
	}
	return &comp, nil
}

func (b *BoltDB) updateComponentStatus(bkt *bolt.Bucket, name string, status proto.Component_Status) (*proto.Component, error) {
	spl := strings.Split(string(name), "#")
	ref, seqIDRaw := spl[0], spl[1]
	seqID, err := strconv.Atoi(seqIDRaw)
	if err != nil {
		panic(err)
	}

	compBkt := bkt.Bucket([]byte(ref))
	if compBkt == nil {
		panic("bad")
	}

	comp := proto.Component{}
	if err := dbGet(compBkt, seqID2(seqID), &comp); err != nil {
		panic(err)
	}

	// finalize this one
	comp.Status = status
	if err := dbPut(compBkt, seqID2(seqID), &comp); err != nil {
		panic(err)
	}
	return &comp, nil
}

func (b *BoltDB) GetHistory(name string) []*proto.Component {
	tx, err := b.db.Begin(false)
	if err != nil {
		panic(err)
	}
	defer tx.Rollback()

	deploymentID, err := b.nameToDeploymentID(tx, name)
	if err != nil {
		panic(err)
	}

	bkt := tx.Bucket(deploymentsBucket).Bucket([]byte(deploymentID))
	if bkt == nil {
		panic("bad1 2")
	}
	comps := bkt.Bucket([]byte("components"))
	if err != nil {
		panic("bad3")
	}
	reqs := bkt.Bucket([]byte("requests"))
	if reqs == nil {
		panic("bad2")
	}

	result := []*proto.Component{}
	c := reqs.Cursor()
	for k, v := c.Seek([]byte("seq-")); k != nil; k, v = c.Next() {
		component, err := b.getComponentFromBucket(string(v), comps)
		if err != nil {
			panic(err)
		}
		result = append(result, component)
	}
	return result
}

func (b *BoltDB) GetComponents(name string) ([]*proto.Component, error) {
	tx, err := b.db.Begin(false)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	deploymentID, err := b.nameToDeploymentID(tx, name)
	if err != nil {
		return nil, err
	}

	bkt := tx.Bucket(deploymentsBucket).Bucket([]byte(deploymentID))
	if bkt == nil {
		panic("bad1 3")
	}

	comps := bkt.Bucket([]byte("components"))
	if err != nil {
		panic("bad3")
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

func (b *BoltDB) Finalize2(clusterName string) error {
	fmt.Println("XXXXXXXXXXXXXXXx")
	fmt.Println(clusterName)

	tx, err := b.db.Begin(true)
	if err != nil {
		panic(err)
	}
	defer tx.Rollback()

	deploymentID, err := b.nameToDeploymentID(tx, clusterName)
	if err != nil {
		return err
	}
	if _, ok := b.queue2.finalize(deploymentID); !ok {
		panic("X")
	}

	bkt := tx.Bucket(deploymentsBucket).Bucket([]byte(deploymentID))
	if bkt == nil {
		panic("bad1 3")
	}
	comps := bkt.Bucket([]byte("components"))
	if comps == nil {
		panic("bad3")
	}
	reqs := bkt.Bucket([]byte("requests"))
	if reqs == nil {
		panic("bad2")
	}

	num, err := getSeqNumber(reqs, nextAppliedKey)
	if err != nil {
		panic(err)
	}

	data := reqs.Get(seqID(num))
	if data == nil {
		panic("bad")
	}

	fmt.Println("-- dat a--")
	fmt.Println(string(data))

	if _, err := b.updateComponentStatus(comps, string(data), proto.Component_APPLIED); err != nil {
		panic(err)
	}

	// update the next key to apply
	if err := putSeqNumber(reqs, nextAppliedKey, num+1); err != nil {
		panic(err)
	}

	// check the next one
	data = reqs.Get(seqID(num + 1))
	if data != nil {
		// there is a new one, eval it
		nextComp, err := b.updateComponentStatus(comps, string(data), proto.Component_QUEUED)
		if err != nil {
			panic(err)
		}

		//fmt.Println("-- next comp --")
		//fmt.Println(nextComp)
		b.addTask(deploymentID, nextComp)
	}

	if err := tx.Commit(); err != nil {
		panic(err)
	}
	return nil
}

func componentKey(namespace, name string) string {
	res := "comp-" + namespace
	if name != "" {
		res += "-" + name
	}
	return res
}

func (b *BoltDB) addTask(deploymentID string, comp *proto.Component) {
	b.queue2.add(&proto.Task{
		DeploymentID: deploymentID,
		ComponentID:  comp.Id,
		Sequence:     comp.Sequence,
	})
}

func (b *BoltDB) applyComponent2(c *proto.Component) (*proto.Component, error) {
	tx, err := b.db.Begin(true)
	if err != nil {
		panic(err)
	}
	defer tx.Rollback()

	c, err = b.applyComponent(tx.Bucket(componentsBucket), c)
	if err != nil {
		panic(err)
	}

	if err := tx.Commit(); err != nil {
		panic(err)
	}
	return c, nil
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
			// we only expect create object
			if c.Action != proto.Component_CREATE {
				return nil, fmt.Errorf("the component is deleted, only create expected but found %s", prev.Action)
			}
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
	/*
		if err := putSeqNumber(bkt, lastSeqKey, c.Sequence); err != nil {
			return nil, err
		}
	*/
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

/*
type kvIndex struct {
	bkt *bolt.Bucket
}

func (k *kvIndex) Exists(key string) bool {
	return k.Get(key) == ""
}

func (k *kvIndex) Get(key string) string {
	return string(k.bkt.Get([]byte(key)))
}

func (k *kvIndex) Put(key, val string) error {
	return k.bkt.Put([]byte(key), []byte(val))
}

func (k *kvIndex) List() (res []string) {
	k.bkt.ForEach(func(k, v []byte) error {
		res = append(res, string(k))
		return nil
	})
	return
}
*/

type bucket struct {
	bkt *bolt.Bucket
}

func (b *bucket) CreateBuckets(names []string) error {
	for _, name := range names {
		if _, err := b.bkt.CreateBucketIfNotExists([]byte(name)); err != nil {
			return err
		}
	}
	return nil
}

/*
func (b *bucket) Last(path string) string {
	bkt := b.bkt.Bucket([]byte(path))
	c := bkt.Cursor()
	if k, _ := c.Last(); k != nil {
		return string(k)
	}
	return ""
}

func (b *bucket) GetInt(path string, key string) (int, error) {
	data := b.Get(path, key)
	if len(data) == 0 {
		return 0, nil
	}
	num, err := strconv.Atoi(string(data))
	if err != nil {
		return 0, err
	}
	return num, nil
}

func (b *bucket) Get(path string, key string) string {
	bkt := b.bkt.Bucket([]byte(path))
	return string(bkt.Get([]byte(key)))
}

func (b *bucket) Put(path string, k string, v interface{}) error {
	bkt := b.bkt.Bucket([]byte(path))
	var raw string
	switch obj := v.(type) {
	case string:
		raw = obj
	default:
		return fmt.Errorf("type not found %s", reflect.TypeOf(v))
	}
	return bkt.Put([]byte(k), []byte(raw))
}
*/

func (b *BoltDB) readLatestComponent(bkt *bolt.Bucket) (*proto.Component, error) {
	c := bkt.Cursor()
	k, _ := c.Last()
	if k == nil {
		panic("BAD, there should be at least one sequence")
	}
	component := proto.Component{}
	if err := gproto.Unmarshal(bkt.Get(k), &component); err != nil {
		return nil, err
	}
	return &component, nil
}

func (b *BoltDB) ReadDeployment(tx *bolt.Tx, id string) (*proto.Component, error) {
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
	if k == nil {
		// Cannot happen
		panic("BUG")
	}

	clusterBkt := compBkt.Bucket(k)
	return b.readLatestComponent(clusterBkt)
}

func (b *BoltDB) nameToDeploymentID(tx *bolt.Tx, name string) (string, error) {
	var deploymentID string

	err := tx.Bucket(deploymentsBucket).ForEach(func(k, v []byte) error {
		comp, err := b.ReadDeployment(tx, string(k))
		if err != nil {
			return err
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

func (b *BoltDB) Apply2(comp *proto.Component) (*proto.Component, error) {
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
		// generate a new deployment id
		deploymentID = uuid.UUID()
	}

	bkt, err := tx.Bucket(deploymentsBucket).CreateBucketIfNotExists([]byte(deploymentID))
	if err != nil {
		return nil, err
	}

	bb := &bucket{bkt}
	err = bb.CreateBuckets([]string{
		"components",
		"requests",
	})
	if err != nil {
		return nil, err
	}

	// bucket to store the components
	comps, err := bkt.CreateBucketIfNotExists([]byte("components"))
	if err != nil {
		return nil, err
	}
	// bucket for the pending requests to be applied
	reqs, err := bkt.CreateBucketIfNotExists([]byte("requests"))
	if err != nil {
		return nil, err
	}

	// find the resource with the same name (if any)
	var resourceID string

	err = comps.ForEach(func(k, v []byte) error {
		bkt = comps.Bucket(k)
		component, err := b.readLatestComponent(bkt)
		if err != nil {
			return err
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
		fmt.Println("add eval", clusterName)
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

func (b *BoltDB) addSequenceComponent(bkt *bolt.Bucket, key string) (bool, error) {
	// get the latest sequence number
	nextSeq, err := getLatestSequence(bkt)
	if err != nil {
		return false, err
	}

	//fmt.Println("-- key --")
	//fmt.Println(key)

	nextKeyToApply, err := getSeqNumber(bkt, nextAppliedKey)
	if err != nil {
		return false, err
	}
	if nextKeyToApply == 0 {
		// first entry, add the key
		nextKeyToApply = 1
		if err := putSeqNumber(bkt, nextAppliedKey, 1); err != nil {
			return false, err
		}
	}

	var addEval bool
	if nextSeq == int(nextKeyToApply) {
		addEval = true
	}
	if err := bkt.Put(seqID2(nextSeq), []byte(key)); err != nil {
		return false, err
	}
	return addEval, nil
}

func (b *BoltDB) GetTask2(ctx context.Context) *proto.Task {
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
		dep, err := b.loadDeploymentImpl(depsBkt, string(k))
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

	dep, err := b.loadDeploymentImpl(depsBkt, id)
	if err != nil {
		return nil, err
	}
	return dep, nil
}

func (b *BoltDB) loadDeploymentImpl(depsBkt *bolt.Bucket, id string) (*proto.Deployment, error) {
	// find the sub-bucket for the cluster
	depBkt := depsBkt.Bucket([]byte(id))
	if depBkt == nil {
		return nil, nil
	}

	// load the cluster meta
	c := &proto.Deployment{
		Instances: []*proto.Instance{},
	}
	if err := dbGet(depBkt, depKey, c); err != nil {
		if err == errNotFound {
			return nil, nil
		}
		return nil, err
	}

	// load the nodes under node-<id>
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
			//fmt.Println("IS OUT!!")
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
	depBkt, _ := depsBkt.CreateBucketIfNotExists([]byte(d.Name))

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

	fmt.Println(n.Cluster)

	// find the sub-bucket for the cluster
	depBkt := depsBkt.Bucket([]byte(n.Cluster))
	if depBkt == nil {
		// create hte bucket, later on we add an step to do this
		if depBkt, err = depsBkt.CreateBucket([]byte(n.Cluster)); err != nil {
			return err
		}
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
		panic(err)
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

/*
func SetupFn(t *testing.T) (state.State, func()) {
	path := "/tmp/db-" + uuid.UUID()

	st, err := Factory(map[string]interface{}{
		"path": path,
	})
	if err != nil {
		t.Fatal(err)
	}
	closeFn := func() {
		if err := os.Remove(path); err != nil {
			t.Fatal(err)
		}
	}
	return st, closeFn
}
*/

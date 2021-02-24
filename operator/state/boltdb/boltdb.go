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
	"github.com/teseraio/ensemble/operator/proto"
	"github.com/teseraio/ensemble/operator/state"
)

var (
	// clustersBucket    = []byte("clusters")
	deploymentsBucket = []byte("deployments")
	evaluationsBucket = []byte("evaluations")

	componentsBucket = []byte("components")

	// versioned indexes
	clusterIndex = []byte("indx_cluster")

	metaKey  = []byte("meta")
	indexKey = []byte("index")
	seqKey   = []byte("seq")
)

// Factory is the factory method for the Boltdb backend
func Factory(config map[string]interface{}) (state.State, error) {
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
		db:    db,
		queue: newTaskQueue(),
	}
	if err := b.initialize(); err != nil {
		return nil, err
	}
	return b, nil
}

// BoltDB is a boltdb state implementation
type BoltDB struct {
	db    *bolt.DB
	queue *taskQueue

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
		evaluationsBucket,
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

	// load the indexes into the task
	err = b.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(componentsBucket)

		return bkt.ForEach(func(k, v []byte) error {
			compBkt := bkt.Bucket(k)

			return compBkt.ForEach(func(k, v []byte) error {
				bkt := compBkt.Bucket(k)
				seqBkt := bkt.Bucket(seqKey)

				seq, err := getSeqNumber(bkt)
				if err != nil {
					return err
				}
				comp := proto.Component{}
				if err := dbGet(seqBkt, seqID(seq), &comp); err != nil {
					return err
				}
				if comp.Status == proto.Component_PENDING {
					clusterID, err := clusterIDFromComponent(&comp)
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

	return nil
}

// Close implements the BoltDB interface
func (b *BoltDB) Close() error {
	return b.db.Close()
}

func getSeqNumber(bkt *bolt.Bucket) (int64, error) {
	seq := 0
	var err error
	if raw := bkt.Get(metaKey); raw != nil {
		if seq, err = strconv.Atoi(string(raw)); err != nil {
			return 0, err
		}
	}
	return int64(seq), nil
}

func (b *BoltDB) Apply(c *proto.Component) (int64, error) {
	namespace := []byte(c.Spec.TypeUrl)

	tx, err := b.db.Begin(true)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	componentsBkt := tx.Bucket(componentsBucket)

	// append current timestamp
	c.Timestamp = ptypes.TimestampNow()

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

	// get sequence number
	seq, err := getSeqNumber(componentBucket)
	if err != nil {
		return 0, err
	}

	// reference to the bucket to store the historical sequences
	seqBkt, err := componentBucket.CreateBucketIfNotExists(seqKey)
	if err != nil {
		return 0, err
	}

	// get the current version and check if it has changed
	old := proto.Component{}
	if seq != 0 {
		if err := dbGet(seqBkt, seqID(int64(seq)), &old); err != nil {
			return 0, err
		}
		if bytes.Equal(old.Spec.Value, c.Spec.Value) {
			return 0, nil
		}
	}

	// update the sequence number in the component
	c.Sequence = int64(seq) + 1
	if err := dbPut(seqBkt, seqID(c.Sequence), c); err != nil {
		return 0, err
	}

	// update the sequence
	if err := componentBucket.Put(metaKey, []byte(fmt.Sprintf("%d", c.Sequence))); err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}

	if old.Status != proto.Component_PENDING {
		// decode the clusterID of the component and push it to the queue
		clusterID, err := clusterIDFromComponent(c)
		if err != nil {
			return 0, err
		}
		b.queue.add(clusterID, c)
	}
	return c.Sequence, nil
}

func (b *BoltDB) GetComponent(id string, generation int64) (*proto.Component, *proto.Component, error) {
	tx, err := b.db.Begin(false)
	if err != nil {
		return nil, nil, err
	}
	defer tx.Rollback()

	generalBkt := tx.Bucket(clusterIndex)

	// find the bucket for the specific id
	compBkt := generalBkt.Bucket([]byte(id))

	seqBkt := compBkt.Bucket(seqKey)

	// get sequence number
	seq := 0
	if raw := compBkt.Get(metaKey); raw != nil {
		if seq, err = strconv.Atoi(string(raw)); err != nil {
			return nil, nil, err
		}
	}
	if seq == 0 {
		// it does not exists
		return nil, nil, nil
	}

	// read current object
	current := proto.Component{}
	if err := dbGet(seqBkt, seqID(int64(seq)), &current); err != nil {
		return nil, nil, err
	}

	// read old object is seq != 1
	var old *proto.Component
	if seq != 1 {
		cc := proto.Component{}
		if err := dbGet(seqBkt, seqID(int64(seq-1)), &cc); err != nil {
			return nil, nil, err
		}
		old = &cc
	}

	return &current, old, nil
}

func seqID(seq int64) []byte {
	return []byte(fmt.Sprintf("seq-%d", seq))
}

type clusterItem interface {
	gproto.Message
	GetClusterID() string
}

var specs = map[string]clusterItem{
	"proto.ResourceSpec": &proto.ResourceSpec{},
}

func clusterIDFromComponent(c *proto.Component) (string, error) {
	var clusterID string
	if c.Spec.TypeUrl == "proto.ClusterSpec2" {
		// the name of the component is the id of the cluster
		clusterID = c.Name
	} else {
		item, ok := specs[c.Spec.TypeUrl]
		if !ok {
			return "", fmt.Errorf("bad")
		}
		if err := gproto.Unmarshal(c.Spec.Value, item); err != nil {
			return "", err
		}
		clusterID = item.GetClusterID()
	}
	return clusterID, nil
}

func (b *BoltDB) Finalize(id string) error {
	tt, ok := b.queue.finalize(id)
	if !ok {
		return fmt.Errorf("task for id '%s' not found", id)
	}

	namespace := []byte(tt.Component.Spec.TypeUrl)

	tx, err := b.db.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	componentsBkt := tx.Bucket(componentsBucket)
	namespaceBkt := componentsBkt.Bucket(namespace)
	componentBkt := namespaceBkt.Bucket([]byte(tt.Component.Name))

	seq, err := getSeqNumber(componentBkt)
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
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	// notify any wait channels
	b.waitChLock.Lock()
	if ch, ok := b.waitCh[id]; ok {
		close(ch)
		delete(b.waitCh, id)
	}
	b.waitChLock.Unlock()

	return nil
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

func (b *BoltDB) LoadDeployment(id string) (*proto.Deployment, error) {
	tx, err := b.db.Begin(false)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	depsBkt := tx.Bucket(deploymentsBucket)

	// find the sub-bucket for the cluster
	depBkt := depsBkt.Bucket([]byte(id))
	if depBkt == nil {
		// cannot create it coz txn is not writtable
		return &proto.Deployment{Id: id, Instances: []*proto.Instance{}}, nil
	}

	// load the cluster meta
	c := &proto.Deployment{
		Instances: []*proto.Instance{},
	}
	if err := dbGet(depBkt, metaKey, c); err != nil {
		if err == errNotFound {
			return nil, fmt.Errorf("meta key not found")
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
	depBkt, _ := depsBkt.CreateBucketIfNotExists([]byte(d.Id))

	dd := d.Copy()
	dd.Instances = nil

	if err := dbPut(depBkt, metaKey, dd); err != nil {
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

/*
func (b *BoltDB) GetCluster(name string) (*proto.Cluster, error) {
	tx, err := b.db.Begin(false)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	clustersBkt := tx.Bucket(clustersBucket)

	c := proto.Cluster{}
	if err := dbGet(clustersBkt, []byte(name), &c); err != nil {
		if err == errNotFound {
			return nil, state.ErrClusterNotFound
		}
		return nil, err
	}
	return &c, nil
}

// UpsertCluster implements the BoltDB interface
func (b *BoltDB) UpsertCluster(c *proto.Cluster) error {
	tx, err := b.db.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	clustersBkt := tx.Bucket(clustersBucket)
	if err := dbPut(clustersBkt, []byte(c.Name), c); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}
*/

func (b *BoltDB) AddEvaluation(eval *proto.Evaluation) error {
	/*
		tx, err := b.db.Begin(true)
		if err != nil {
			return err
		}
		defer tx.Rollback()

		evalBkt := tx.Bucket(evaluationsBucket)
		if err := dbPut(evalBkt, []byte(eval.Id), eval); err != nil {
			return err
		}
		if err := tx.Commit(); err != nil {
			return err
		}

		tt := &task{
			Evaluation: eval,
			timestamp:  time.Now(),
		}
		b.queue2.add(tt)
	*/
	return nil
}

func (b *BoltDB) GetTask2(ctx context.Context) (*proto.Evaluation, error) {
	/*
		t := b.queue2.pop(ctx)
		if t == nil {
			return nil, nil
		}
		return t.Evaluation, nil
	*/
	return nil, nil
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

package boltdb

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"

	gproto "github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"

	"github.com/boltdb/bolt"
	"github.com/teseraio/ensemble/lib/uuid"
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
	namespace := []byte(getProtoNamespace(c))

	tx, err := b.db.Begin(true)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	componentsBkt := tx.Bucket(componentsBucket)

	// append current timestamp
	c.Timestamp = ptypes.TimestampNow()
	c.Id = fmt.Sprintf("%s.%s", namespace, c.Name)

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

	// get sequence number, TODO: this only allows two values stored
	seq, err := getSeqNumber(componentBucket)
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
	if seq != 0 {
		if err := dbGet(seqBkt, seqID(int64(seq)), &old); err != nil {
			return 0, err
		}
		if c.Action == proto.Component_CREATE {
			if bytes.Equal(old.Spec.Value, c.Spec.Value) {
				return 0, nil
			}
		}
	} else {
		if c.Action == proto.Component_DELETE {
			return 0, fmt.Errorf("cannot remove non created object")
		}
	}

	// update the sequence number in the component
	c.Sequence = int64(seq) + 1

	if old.Status != proto.Component_PENDING {
		// push this component to the pending queue
		clusterID, err := proto.ClusterIDFromComponent(c)
		if err != nil {
			return 0, err
		}
		b.queue.add(clusterID, c)

		c.Status = proto.Component_PENDING
		if err := componentBucket.Put(metaKey, []byte(fmt.Sprintf("%d", c.Sequence))); err != nil {
			return 0, err
		}
	} else {
		c.Status = proto.Component_QUEUED
	}

	if err := dbPut(seqBkt, seqID(c.Sequence), c); err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return c.Sequence, nil
}

func (b *BoltDB) GetComponentWithSequence(id string, sequence int64) (*proto.Component, error) {
	tx, err := b.db.Begin(false)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	parts := strings.Split(id, ".")
	namespace, name := parts[0], parts[1]

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

func (b *BoltDB) GetComponent(id string) (*proto.Component, error) {
	tx, err := b.db.Begin(false)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	parts := strings.Split(id, ".")
	namespace, name := parts[0], parts[1]

	componentsBkt := tx.Bucket(componentsBucket)
	namespaceBkt := componentsBkt.Bucket([]byte(namespace))

	compBkt := namespaceBkt.Bucket([]byte(name))
	seqBkt := compBkt.Bucket(seqKey)

	// read current object
	seq, err := getSeqNumber(compBkt)
	if err != nil {
		return nil, err
	}

	comp := proto.Component{}
	if err := dbGet(seqBkt, seqID(seq), &comp); err != nil {
		return nil, err
	}
	return &comp, nil
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

func (b *BoltDB) Finalize(id string) error {
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
		return nil, nil
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

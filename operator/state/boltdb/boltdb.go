package boltdb

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"

	gproto "github.com/golang/protobuf/proto"

	"github.com/boltdb/bolt"
	"github.com/teseraio/ensemble/operator/proto"
	"github.com/teseraio/ensemble/operator/state"
)

var (
	clustersBucket = []byte("clusters")

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
	indexes := [][]byte{
		clusterIndex,
	}
	buckets := [][]byte{
		clustersBucket,
	}
	err := b.db.Update(func(tx *bolt.Tx) error {
		for _, i := range append(buckets, indexes...) {
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
		for _, i := range indexes {
			generalBkt := tx.Bucket(i)
			err = generalBkt.ForEach(func(k, v []byte) error {
				itemBkt := generalBkt.Bucket(k)

				// Find the first version in the sequence bucket that is pending and push
				// it to the task queue
				cursor := itemBkt.Bucket(seqKey).Cursor()

				for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
					c := proto.Component{}
					if err := gproto.Unmarshal(v, &c); err != nil {
						return err
					}
					if c.Status == proto.Component_PENDING {
						b.queue.add(&c)
						break
					}
				}
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
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

func (b *BoltDB) applyImpl(bucket []byte, c *proto.Component) error {
	tx, err := b.db.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	generalBkt := tx.Bucket(bucket)

	// find the bucket for the specific id
	compBkt, err := generalBkt.CreateBucketIfNotExists([]byte(c.Name))
	if err != nil {
		return err
	}

	// get sequence number
	seq := 0
	if raw := compBkt.Get(metaKey); raw != nil {
		if seq, err = strconv.Atoi(string(raw)); err != nil {
			return err
		}
	}

	// append to the sequence bucket
	seqBkt, err := compBkt.CreateBucketIfNotExists(seqKey)
	if err != nil {
		return err
	}

	// get the current version and check if it has changed
	{
		if seq != 0 {
			c0 := proto.Component{}
			if err := dbGet(seqBkt, []byte(fmt.Sprintf("seq-%d", seq)), &c0); err != nil {
				return err
			}
			if bytes.Equal(c0.Spec.Value, c.Spec.Value) {
				return nil
			}
		}
	}

	// update the sequence number in the component
	c.Sequence = int64(seq) + 1

	seqID := []byte(fmt.Sprintf("seq-%d", c.Sequence))
	if err := dbPut(seqBkt, seqID, c); err != nil {
		return err
	}

	// update the sequence
	if err := compBkt.Put(metaKey, []byte(fmt.Sprintf("%d", c.Sequence))); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}

	// update the index if there is no task for this id
	if !b.queue.existsByName(c.Name) {
		b.queue.add(c)
	}
	return nil
}

func (b *BoltDB) getImpl(tx *bolt.Tx, name string) (*proto.Component, error) {
	bucket := clusterIndex
	generalBkt := tx.Bucket(bucket)

	// find the bucket for the specific id
	compBkt := generalBkt.Bucket([]byte(name))
	if compBkt == nil {
		return nil, fmt.Errorf("bad")
	}

	// get sequence number
	seq := 0
	if raw := compBkt.Get(metaKey); raw != nil {
		var err error
		if seq, err = strconv.Atoi(string(raw)); err != nil {
			return nil, err
		}
	}
	if seq == 0 {
		return nil, fmt.Errorf("not found x")
	}

	seqID := []byte(fmt.Sprintf("seq-%d", seq))

	comp := &proto.Component{}
	if err := dbGet(compBkt.Bucket(seqKey), seqID, comp); err != nil {
		return nil, err
	}
	return comp, nil
}

func (b *BoltDB) Get(name string) (*proto.Component, error) {
	var comp *proto.Component

	err := b.db.View(func(tx *bolt.Tx) error {
		var err error
		comp, err = b.getImpl(tx, name)
		return err
	})
	return comp, err
}

func (b *BoltDB) Apply(c *proto.Component) error {
	return b.applyImpl(clusterIndex, c)
}

func (b *BoltDB) GetTask(ctx context.Context) (*proto.Component, error) {
	t := b.queue.pop(ctx)
	if t == nil {
		return nil, nil
	}
	return t.Component, nil
}

func seqID(seq int64) []byte {
	return []byte(fmt.Sprintf("seq-%d", seq))
}

func (b *BoltDB) Finalize(id string) error {
	task, ok := b.queue.get(id)
	if !ok {
		return fmt.Errorf("task for id '%s' not found", id)
	}

	tx, err := b.db.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	comp := task.Component

	generalBkt := tx.Bucket(clusterIndex)

	// find the bucket for the specific id
	seqBkt := generalBkt.Bucket([]byte(comp.Name)).Bucket(seqKey)

	comp.Status = proto.Component_APPLIED
	if err := dbPut(seqBkt, seqID(comp.Sequence), comp); err != nil {
		return err
	}

	// find the next possible sequence
	nextComp := proto.Component{}
	if err := dbGet(seqBkt, seqID(comp.Sequence+1), &nextComp); err != nil {
		if err != errNotFound {
			return err
		}
	} else {
		// there exists a next component
		b.queue.add(&nextComp)
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

func (b *BoltDB) LoadCluster(id string) (*proto.Cluster, error) {
	tx, err := b.db.Begin(false)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	clustersBkt := tx.Bucket(clustersBucket)

	// find the sub-bucket for the cluster
	clusterBkt := clustersBkt.Bucket([]byte(id))
	if clusterBkt == nil {
		return nil, state.ErrClusterNotFound
	}

	// load the cluster meta
	c := &proto.Cluster{
		Nodes: []*proto.Node{},
	}
	if err := dbGet(clusterBkt, metaKey, c); err != nil {
		if err == errNotFound {
			return nil, fmt.Errorf("meta key not found")
		}
		return nil, err
	}

	// load the nodes under node-<id>
	nodeCursor := clusterBkt.Cursor()
	for k, _ := nodeCursor.First(); k != nil; k, _ = nodeCursor.Next() {
		if !strings.HasPrefix(string(k), "node-") {
			continue
		}
		n := &proto.Node{}
		if err := dbGet(clusterBkt, k, n); err != nil {
			return nil, err
		}
		c.Nodes = append(c.Nodes, n)
	}
	return c, nil
}

// UpsertNode implements the BoltDB interface
func (b *BoltDB) UpsertNode(n *proto.Node) error {
	tx, err := b.db.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	clustersBkt := tx.Bucket(clustersBucket)

	// find the sub-bucket for the cluster
	clusterBkt := clustersBkt.Bucket([]byte(n.Cluster))
	if clusterBkt == nil {
		return state.ErrClusterNotFound
	}

	// upsert under node-<id>
	id := "node-" + n.ID
	if err := dbPut(clusterBkt, []byte(id), n); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

// UpsertCluster implements the BoltDB interface
func (b *BoltDB) UpsertCluster(c *proto.Cluster) error {
	tx, err := b.db.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	clustersBkt := tx.Bucket(clustersBucket)

	// find the sub-bucket for the cluster
	clusterBkt := clustersBkt.Bucket([]byte(c.Name))
	if clusterBkt == nil {
		// cluster does not exists, create it
		clusterBkt, err = clustersBkt.CreateBucket([]byte(c.Name))
		if err != nil {
			return err
		}
	}

	c = c.Copy()
	c.Nodes = nil // we do not store the nodes here

	if err := dbPut(clusterBkt, metaKey, c); err != nil {
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

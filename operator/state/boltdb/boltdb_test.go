package boltdb

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/teseraio/ensemble/lib/uuid"
	"github.com/teseraio/ensemble/operator/proto"
	"github.com/teseraio/ensemble/operator/state"
)

func setupFn(t *testing.T) (state.State, func()) {
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

func TestSuite(t *testing.T) {
	state.TestSuite(t, setupFn)
}

func TestBoltdbReindexPending(t *testing.T) {
	config := map[string]interface{}{
		"path": "/tmp/db-" + uuid.UUID(),
	}
	st, err := Factory(config)
	if err != nil {
		t.Fatal(err)
	}
	db := st.(*BoltDB)

	// append two distinct components
	st.Apply(&proto.Component{
		Id:   "id1",
		Name: "A",
		Spec: proto.MustMarshalAny(&proto.ClusterSpec2{}),
	})
	st.Apply(&proto.Component{
		Id:   "id2",
		Name: "B",
		Spec: proto.MustMarshalAny(&proto.ClusterSpec2{}),
	})

	// two values expected in the queue
	assert.Equal(t, db.queue.popImpl().Component.Id, "id1")
	assert.Equal(t, db.queue.popImpl().Component.Id, "id2")

	assert.Nil(t, db.Close())

	// reload the database
	st, err = Factory(config)
	assert.NoError(t, err)

	fmt.Println(st)
}

func TestBoltdbFinalizeMultipleResources(t *testing.T) {
	config := map[string]interface{}{
		"path": "/tmp/db-" + uuid.UUID(),
	}
	st, err := Factory(config)
	if err != nil {
		t.Fatal(err)
	}
	db := st.(*BoltDB)

	rID, _ := db.Apply(&proto.Component{
		Id:   "id1",
		Name: "B",
		Spec: proto.MustMarshalAny(&proto.ResourceSpec{
			Cluster: "A",
		}),
	})
	assert.Equal(t, rID, int64(1))

	rID2, _ := db.Apply(&proto.Component{
		Id:   "id2",
		Name: "B",
		Spec: proto.MustMarshalAny(&proto.ResourceSpec{
			Cluster: "A",
			Params:  "{2}",
		}),
	})
	assert.Equal(t, rID2, int64(2))

	comp := db.queue.popImpl().Component

	assert.Equal(t, comp.Id, "id1")
	assert.Nil(t, db.queue.popImpl())

	db.Finalize(comp.Id)
	assert.Equal(t, db.queue.popImpl().Component.Id, "id2")
}

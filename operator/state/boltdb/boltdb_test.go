package boltdb

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/teseraio/ensemble/lib/uuid"
	"github.com/teseraio/ensemble/operator/proto"
	"github.com/teseraio/ensemble/operator/state"
)

func TestSuite(t *testing.T) {
	state.TestSuite(t, SetupFn)
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
		Spec: proto.MustMarshalAny(&proto.ClusterSpec{}),
	})
	st.Apply(&proto.Component{
		Id:   "id2",
		Name: "B",
		Spec: proto.MustMarshalAny(&proto.ClusterSpec{}),
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
		Name: "B",
		Spec: proto.MustMarshalAny(&proto.ResourceSpec{
			Cluster: "A",
		}),
	})
	assert.Equal(t, rID, int64(1))

	rID2, _ := db.Apply(&proto.Component{
		Name: "B",
		Spec: proto.MustMarshalAny(&proto.ResourceSpec{
			Cluster: "A",
			Params:  "{2}",
		}),
	})
	assert.Equal(t, rID2, int64(2))

	comp := db.queue.popImpl().Component
	assert.Equal(t, comp.Id, "proto-ResourceSpec.B")
	assert.Equal(t, comp.Sequence, int64(1))
	assert.Nil(t, db.queue.popImpl())

	db.Finalize(comp.Id)

	comp = db.queue.popImpl().Component
	assert.Equal(t, comp.Id, "proto-ResourceSpec.B")
	assert.Equal(t, comp.Sequence, int64(2))
}

func TestBoltdbApply(t *testing.T) {
	config := map[string]interface{}{
		"path": "/tmp/db-" + uuid.UUID(),
	}
	st, err := Factory(config)
	if err != nil {
		t.Fatal(err)
	}
	db := st.(*BoltDB)

	cID, _ := db.Apply(&proto.Component{
		Id:     "id1",
		Name:   "A",
		Action: proto.Component_CREATE,
		Spec: proto.MustMarshalAny(&proto.ClusterSpec{
			Backend: "backend1",
		}),
	})
	assert.Equal(t, cID, int64(1))

	// the sequence is not updated
	cID2, _ := db.Apply(&proto.Component{
		Id:     "id2",
		Name:   "A",
		Action: proto.Component_CREATE,
		Spec: proto.MustMarshalAny(&proto.ClusterSpec{
			Backend: "backend1",
		}),
	})
	assert.Equal(t, cID2, int64(0))

	// remove the component
	cID3, _ := db.Apply(&proto.Component{
		Id:     "id2",
		Name:   "A",
		Action: proto.Component_DELETE,
		Spec: proto.MustMarshalAny(&proto.ClusterSpec{
			Backend: "backend1",
		}),
	})
	assert.Equal(t, cID3, int64(2))
}

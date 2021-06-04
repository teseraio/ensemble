package boltdb

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/teseraio/ensemble/lib/uuid"
	"github.com/teseraio/ensemble/operator/proto"
)

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

	db = st.(*BoltDB)

	// we expect the same values
	assert.Equal(t, db.queue.popImpl().Component.Id, "id1")
	assert.Equal(t, db.queue.popImpl().Component.Id, "id2")
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
	assert.Equal(t, comp.Sequence, int64(1))
	assert.Nil(t, db.queue.popImpl())

	assert.NoError(t, db.Finalize("A"))

	comp = db.queue.popImpl().Component
	assert.Equal(t, comp.Id, "id2")
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

func TestGetComponent(t *testing.T) {
	config := map[string]interface{}{
		"path": "/tmp/db-" + uuid.UUID(),
	}
	st, err := Factory(config)
	if err != nil {
		t.Fatal(err)
	}
	db := st.(*BoltDB)

	seq, _ := db.Apply(&proto.Component{
		Id:     "id1",
		Name:   "A",
		Action: proto.Component_CREATE,
		Spec: proto.MustMarshalAny(&proto.ClusterSpec{
			Backend: "backend1",
		}),
	})

	comp, err := db.GetComponent("proto-ClusterSpec", "A", seq)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, comp.Id, "id1")
}

func TestListDeployments(t *testing.T) {
	config := map[string]interface{}{
		"path": "/tmp/db-" + uuid.UUID(),
	}
	st, err := Factory(config)
	if err != nil {
		t.Fatal(err)
	}

	err = st.UpdateDeployment(&proto.Deployment{
		Name: "a",
	})
	assert.NoError(t, err)

	err = st.UpdateDeployment(&proto.Deployment{
		Name: "b",
	})
	assert.NoError(t, err)

	deps, err := st.ListDeployments()
	assert.NoError(t, err)
	assert.Len(t, deps, 2)

	assert.Equal(t, deps[0].Name, "a")
	assert.Equal(t, deps[1].Name, "b")
}

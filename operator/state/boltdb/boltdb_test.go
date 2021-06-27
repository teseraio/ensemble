package boltdb

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/teseraio/ensemble/lib/uuid"
	"github.com/teseraio/ensemble/operator/proto"
)

func testBoltdb(t *testing.T) *BoltDB {
	config := map[string]interface{}{
		"path": "/tmp/db-" + uuid.UUID(),
	}
	st, err := Factory(config)
	if err != nil {
		t.Fatal(err)
	}
	return st
}

func TestListDeployments(t *testing.T) {
	config := map[string]interface{}{
		"path": "/tmp/db-" + uuid.UUID(),
	}
	st, err := Factory(config)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(st.LoadDeployment("a"))

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

func TestApplyMultipleTimesWithoutChange(t *testing.T) {
	// apply the same element without changes should not increase the sequence
	db := testBoltdb(t)

	spec := &proto.ClusterSpec{
		Groups: []*proto.ClusterSpec_Group{
			{
				Count: 1,
			},
		},
	}

	comp0, err := db.Apply2(&proto.Component{
		Name: "name1",
		Spec: proto.MustMarshalAny(spec),
	})
	assert.NoError(t, err)
	assert.Equal(t, comp0.Sequence, int64(1))

	comp1, err := db.Apply2(&proto.Component{
		Name: "name1",
		Spec: proto.MustMarshalAny(spec),
	})
	assert.NoError(t, err)
	assert.Nil(t, comp1)
}

func TestApplyDeleteFirstTime(t *testing.T) {
	// first action cannot be a delete action
	db := testBoltdb(t)

	_, err := db.Apply2(&proto.Component{
		Name:   "name1",
		Action: proto.Component_DELETE,
		Spec:   proto.MustMarshalAny(&proto.ClusterSpec{}),
	})
	assert.Error(t, err)
}

func TestApplyMultipleTimesWithDelete(t *testing.T) {
	// if the component is deleted, we can only create it
	// apply the same element without changes should not increase the sequence
	db := testBoltdb(t)

	spec := &proto.ClusterSpec{
		Groups: []*proto.ClusterSpec_Group{
			{
				Count: 1,
			},
		},
	}

	comp0, err := db.Apply2(&proto.Component{
		Name: "name1",
		Spec: proto.MustMarshalAny(spec),
	})
	assert.NoError(t, err)
	assert.Equal(t, comp0.Sequence, int64(1))

	comp1, err := db.Apply2(&proto.Component{
		Name:   "name1",
		Action: proto.Component_DELETE,
		Spec:   proto.MustMarshalAny(&proto.ClusterSpec{}),
	})
	assert.NoError(t, err)
	assert.Equal(t, comp1.Sequence, int64(2))

	// cannot delete, need to create first
	_, err = db.Apply2(&proto.Component{
		Name:   "name1",
		Action: proto.Component_DELETE,
		Spec:   proto.MustMarshalAny(&proto.ClusterSpec{}),
	})
	assert.Error(t, err)

	// we can create with the original spec again
	comp2, err := db.Apply2(&proto.Component{
		Name: "name1",
		Spec: proto.MustMarshalAny(spec),
	})
	assert.NoError(t, err)
	assert.Equal(t, comp2.Sequence, int64(3))
}

func TestApplyMultipleResources(t *testing.T) {
	// multiple resources for the same cluster are serialized in the tasks
	db := testBoltdb(t)

	comp1, err := db.Apply2(&proto.Component{
		Name: "name1",
		Spec: proto.MustMarshalAny(&proto.ClusterSpec{}),
	})
	assert.NoError(t, err)
	assert.Equal(t, comp1.Sequence, int64(1))

	comp2, err := db.Apply2(&proto.Component{
		Name: "user1",
		Spec: proto.MustMarshalAny(&proto.ResourceSpec{
			Cluster: "name1",
		}),
	})
	assert.NoError(t, err)
	assert.Equal(t, comp2.Sequence, int64(1))

	comp3, err := db.Apply2(&proto.Component{
		Name: "user1",
		Spec: proto.MustMarshalAny(&proto.ResourceSpec{
			Cluster: "name1",
		}),
	})
	assert.NoError(t, err)
	assert.Nil(t, comp3)
}

func TestApplyWithFinalized(t *testing.T) {
	// when a task is finalized the next needs to be triggered
	db := testBoltdb(t)

	_, err := db.Apply2(&proto.Component{
		Name: "name1",
		Spec: proto.MustMarshalAny(&proto.ClusterSpec{}),
	})
	assert.NoError(t, err)

	_, err = db.Apply2(&proto.Component{
		Name: "name1",
		Spec: proto.MustMarshalAny(&proto.ClusterSpec{
			Groups: []*proto.ClusterSpec_Group{
				{Count: 1},
			},
		}),
	})
	assert.NoError(t, err)

	// you can only pop one value
	task1 := db.queue2.popImpl()
	assert.NotNil(t, task1)
	assert.Nil(t, db.queue2.popImpl())

	assert.NoError(t, db.Finalize2("name1"))

	// you can pop a second value
	task2 := db.queue2.popImpl()
	assert.NotNil(t, task2)

	comps, err := db.GetComponents("name1")
	fmt.Println(comps)

	db.GetHistory("name1")

	//assert.NoError(t, err)
	//assert.Equal(t, comps[0].Status, proto.Component_APPLIED)
	//assert.Equal(t, comps[1].Status, proto.Component_QUEUED)
}

func TestApplyWhenQueued(t *testing.T) {
	// apply when the previous component is in queue
	db := testBoltdb(t)

	comp, err := db.Apply2(&proto.Component{
		Name: "name1",
		Spec: proto.MustMarshalAny(&proto.ClusterSpec{}),
	})
	assert.NoError(t, err)
	//panic(err)

	//task1 := db.queue2.popImpl()
	//assert.NotNil(t, task1)

	comp1, err := db.Apply2(&proto.Component{
		Name: "name1",
		Spec: proto.MustMarshalAny(&proto.ClusterSpec{
			Groups: []*proto.ClusterSpec_Group{
				{Count: 1},
			},
		}),
	})
	assert.NoError(t, err)

	fmt.Println(comp)
	fmt.Println(comp1)

	comps, err := db.GetComponents("name1")
	assert.NoError(t, err)
	fmt.Println(comps)

	db.GetHistory("name1")
}

func TestApplySecond(t *testing.T) {
	// second item being applied
	db := testBoltdb(t)

	_, err := db.Apply2(&proto.Component{
		Name: "name1",
		Spec: proto.MustMarshalAny(&proto.ClusterSpec{}),
	})
	assert.NoError(t, err)

	/*
		task1 := db.queue2.popImpl()
		assert.NotNil(t, task1)
		assert.NoError(t, db.Finalize2("name1"))
	*/

	_, err = db.Apply2(&proto.Component{
		Name: "name1",
		Spec: proto.MustMarshalAny(&proto.ClusterSpec{
			Groups: []*proto.ClusterSpec_Group{
				{Count: 1},
			},
		}),
	})
	assert.NoError(t, err)

	/*
		task2 := db.queue2.popImpl()
		assert.NotNil(t, task2)

		comps, err := db.GetComponents("name1")
		assert.NoError(t, err)
		assert.Equal(t, comps[0].Status, proto.Component_APPLIED)
		assert.Equal(t, comps[1].Status, proto.Component_QUEUED)
	*/
}

// watcher (tests)
// Reindex
// List the stubs

func TestApplyComponent(t *testing.T) {
	db := testBoltdb(t)

	c0, _ := db.applyComponent2(&proto.Component{
		Id:   "a",
		Spec: proto.MustMarshalAny(&proto.ClusterSpec{}),
	})
	c1, _ := db.applyComponent2(&proto.Component{
		Id: "a",
		Spec: proto.MustMarshalAny(&proto.ClusterSpec{
			Groups: []*proto.ClusterSpec_Group{
				{Count: 1},
			},
		}),
	})
	c2, _ := db.applyComponent2(&proto.Component{
		Id:   "a",
		Spec: proto.MustMarshalAny(&proto.ClusterSpec{}),
	})
	fmt.Println(c0)
	fmt.Println(c1)
	fmt.Println(c2)
}

package boltdb

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/teseraio/ensemble/lib/uuid"
	"github.com/teseraio/ensemble/operator/proto"
)

func testBoltdb(t *testing.T, pathRaw ...string) *BoltDB {
	path := "/tmp/db-" + uuid.UUID()
	if len(pathRaw) == 1 {
		path = pathRaw[0]
	}
	config := map[string]interface{}{
		"path": path,
	}
	st, err := Factory(config)
	if err != nil {
		t.Fatal(err)
	}
	return st
}

func TestListDeployments(t *testing.T) {
	db := testBoltdb(t)

	comp1, err := db.Apply(&proto.Component{
		Name: "name1",
		Spec: proto.MustMarshalAny(&proto.ClusterSpec{}),
	})
	assert.NoError(t, err)
	assert.Equal(t, comp1.Sequence, int64(1))

	depID, err := db.NameToDeployment("name1")
	assert.NoError(t, err)

	dep, err := db.LoadDeployment(depID)
	assert.NoError(t, err)

	dep.Backend = "xxx"
	err = db.UpdateDeployment(dep)
	assert.NoError(t, err)

	deps, err := db.ListDeployments()
	assert.NoError(t, err)
	assert.Len(t, deps, 1)

	assert.Equal(t, deps[0].Id, depID)
	assert.Equal(t, deps[0].Name, "name1")

	dep2, err := db.LoadDeployment(depID)
	assert.NoError(t, err)
	assert.Equal(t, dep2.Backend, "xxx")
}

func TestApplyFirst_CannotDelete(t *testing.T) {
	// first action cannot be delete
	db := testBoltdb(t)

	_, err := db.Apply(&proto.Component{
		Name:   "name1",
		Action: proto.Component_DELETE,
		Spec:   proto.MustMarshalAny(&proto.ClusterSpec{}),
	})
	assert.Error(t, err)
}

func TestApplySecond_SameComponent(t *testing.T) {
	// apply the same element without changes should not increase the sequence
	db := testBoltdb(t)

	spec := &proto.ClusterSpec{
		Groups: []*proto.ClusterSpec_Group{
			{
				Count: 1,
			},
		},
	}

	comp0, err := db.Apply(&proto.Component{
		Name: "name1",
		Spec: proto.MustMarshalAny(spec),
	})
	assert.NoError(t, err)
	assert.Equal(t, comp0.Sequence, int64(1))

	comp1, err := db.Apply(&proto.Component{
		Name: "name1",
		Spec: proto.MustMarshalAny(spec),
	})
	assert.NoError(t, err)
	assert.Nil(t, comp1)
}

func TestApplySecond_FinalizeFirst(t *testing.T) {
	// when a task is finalized the next needs to be triggered
	db := testBoltdb(t)

	comp0, err := db.Apply(&proto.Component{
		Name: "name1",
		Spec: proto.MustMarshalAny(&proto.ClusterSpec{}),
	})
	assert.NoError(t, err)

	depID, err := db.NameToDeployment("name1")
	assert.NoError(t, err)

	_, err = db.Apply(&proto.Component{
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
	assert.Equal(t, task1.ComponentID, comp0.Id)
	assert.Nil(t, db.queue2.popImpl())

	assert.NoError(t, db.Finalize(depID))

	// pop the second evaluation
	task2 := db.queue2.popImpl()
	assert.NotNil(t, task2)
	assert.Equal(t, task2.ComponentID, comp0.Id)
	assert.Equal(t, task2.Sequence, int64(2))

	vers, err := db.GetComponentVersions(depID, comp0.Id)
	assert.NoError(t, err)
	assert.Equal(t, vers[0].Status, proto.Component_APPLIED)
	assert.Equal(t, vers[1].Status, proto.Component_QUEUED)
}

func TestApplySecond_FirstQueued(t *testing.T) {
	// apply when the previous component is in queue
	db := testBoltdb(t)

	comp0, err := db.Apply(&proto.Component{
		Name: "name1",
		Spec: proto.MustMarshalAny(&proto.ClusterSpec{}),
	})
	assert.NoError(t, err)
	assert.Equal(t, comp0.Sequence, int64(1))

	task1 := db.queue2.popImpl()
	assert.Equal(t, task1.ComponentID, comp0.Id)

	comp1, err := db.Apply(&proto.Component{
		Name: "name1",
		Spec: proto.MustMarshalAny(&proto.ClusterSpec{
			Groups: []*proto.ClusterSpec_Group{
				{Count: 1},
			},
		}),
	})
	assert.NoError(t, err)
	assert.Equal(t, comp1.Sequence, int64(2))
	assert.Equal(t, comp1.Status, proto.Component_PENDING)
}

func TestApplySecond_FirstAlreadyDone(t *testing.T) {
	// apply when the previous component is finalized
	db := testBoltdb(t)

	comp0, err := db.Apply(&proto.Component{
		Name: "name1",
		Spec: proto.MustMarshalAny(&proto.ClusterSpec{}),
	})
	assert.NoError(t, err)
	assert.Equal(t, comp0.Sequence, int64(1))

	depID, err := db.NameToDeployment("name1")
	assert.NoError(t, err)

	task0 := db.queue2.popImpl()
	assert.Equal(t, task0.ComponentID, comp0.Id)
	assert.Equal(t, depID, task0.DeploymentID)

	// Finish task sequence 1
	assert.NoError(t, db.Finalize(depID))
	comp1, err := db.Apply(&proto.Component{
		Name: "name1",
		Spec: proto.MustMarshalAny(&proto.ClusterSpec{
			Groups: []*proto.ClusterSpec_Group{
				{Count: 1},
			},
		}),
	})
	assert.NoError(t, err)
	assert.Equal(t, comp1.Sequence, int64(2))
	assert.Equal(t, comp1.Status, proto.Component_QUEUED)
}

func TestApplySecond_RevertComponent(t *testing.T) {
	db := testBoltdb(t)

	comp0, err := db.Apply(&proto.Component{
		Spec: proto.MustMarshalAny(&proto.ClusterSpec{}),
	})
	assert.NoError(t, err)
	assert.Equal(t, comp0.Sequence, int64(1))

	comp1, err := db.Apply(&proto.Component{
		Spec: proto.MustMarshalAny(&proto.ClusterSpec{
			Groups: []*proto.ClusterSpec_Group{
				{Count: 1},
			},
		}),
	})
	assert.NoError(t, err)
	assert.Equal(t, comp1.Sequence, int64(2))

	comp2, err := db.Apply(&proto.Component{
		Spec: proto.MustMarshalAny(&proto.ClusterSpec{}),
	})
	assert.NoError(t, err)
	assert.Equal(t, comp2.Sequence, int64(3))
}

func TestApplyDelete_ClusterReuseName(t *testing.T) {
	db := testBoltdb(t)

	compA1, err := db.Apply(&proto.Component{
		Name: "name1",
		Spec: proto.MustMarshalAny(&proto.ClusterSpec{}),
	})
	assert.NoError(t, err)
	assert.Equal(t, compA1.Sequence, int64(1))

	compA2, err := db.Apply(&proto.Component{
		Name:   "name1",
		Action: proto.Component_DELETE,
		Spec:   proto.MustMarshalAny(&proto.ClusterSpec{}),
	})
	assert.NoError(t, err)
	assert.Equal(t, compA2.Sequence, int64(2))

	// create again with the same name, the new component should start
	// with sequence 1 again
	_, err = db.Apply(&proto.Component{
		Name: "name1",
		Spec: proto.MustMarshalAny(&proto.ClusterSpec{}),
	})
	assert.Error(t, err)
}

func TestApplyDelete_ResourceReuseName(t *testing.T) {
	db := testBoltdb(t)

	compA1, err := db.Apply(&proto.Component{
		Name: "name1",
		Spec: proto.MustMarshalAny(&proto.ClusterSpec{}),
	})
	assert.NoError(t, err)
	assert.Equal(t, compA1.Sequence, int64(1))

	compB1, err := db.Apply(&proto.Component{
		Name: "r1",
		Spec: proto.MustMarshalAny(&proto.ResourceSpec{
			Cluster: "name1",
		}),
	})
	assert.NoError(t, err)
	assert.Equal(t, compB1.Sequence, int64(1))

	compB2, err := db.Apply(&proto.Component{
		Name:   "r1",
		Action: proto.Component_DELETE,
		Spec: proto.MustMarshalAny(&proto.ResourceSpec{
			Cluster: "name1",
		}),
	})
	assert.NoError(t, err)
	assert.Equal(t, compB2.Sequence, int64(2))

	// apply again, we should create a new name
	_, err = db.Apply(&proto.Component{
		Name: "r1",
		Spec: proto.MustMarshalAny(&proto.ResourceSpec{
			Cluster: "name1",
		}),
	})
	assert.Error(t, err)
}

func TestApplyResourceUnknownCluster(t *testing.T) {
	db := testBoltdb(t)

	_, err := db.Apply(&proto.Component{
		Name: "r1",
		Spec: proto.MustMarshalAny(&proto.ResourceSpec{
			Cluster: "name1",
		}),
	})
	assert.Error(t, err)
}

func TestComponentsReindex(t *testing.T) {
	db := testBoltdb(t)

	comp, err := db.Apply(&proto.Component{
		Name: "name1",
		Spec: proto.MustMarshalAny(&proto.ClusterSpec{}),
	})
	assert.NoError(t, err)

	db.Close()
	db2 := testBoltdb(t, db.path)

	tt := db2.GetTask(context.Background())
	assert.NotNil(t, tt)
	assert.Equal(t, tt.ComponentID, comp.Id)
	assert.Equal(t, tt.Sequence, int64(1))
}

func TestReadDeployment(t *testing.T) {
	db := testBoltdb(t)

	comp1, err := db.Apply(&proto.Component{
		Name: "name1",
		Spec: proto.MustMarshalAny(&proto.ClusterSpec{}),
	})
	assert.NoError(t, err)
	assert.Equal(t, comp1.Sequence, int64(1))

	deps, _ := db.ListDeployments()
	depID := deps[0].Id

	comp2, err := db.Apply(&proto.Component{
		Name: "name1",
		Spec: proto.MustMarshalAny(&proto.ClusterSpec{
			Groups: []*proto.ClusterSpec_Group{
				{
					Count: 1,
				},
			},
		}),
	})
	assert.NoError(t, err)
	assert.Equal(t, comp2.Sequence, int64(2))

	comp3, err := db.Apply(&proto.Component{
		Name: "name1",
		Spec: proto.MustMarshalAny(&proto.ClusterSpec{
			Groups: []*proto.ClusterSpec_Group{
				{
					Count: 2,
				},
			},
		}),
	})
	assert.NoError(t, err)
	assert.Equal(t, comp3.Sequence, int64(3))

	// sequence=1 is the current pending deployment
	compR1, err := db.ReadDeployment(depID)
	assert.NoError(t, err)
	assert.Equal(t, compR1.Sequence, int64(1))

	// finalize sequence=1
	assert.NoError(t, db.Finalize(depID))

	// sequence=2 is the new pending deployment
	compR2, err := db.ReadDeployment(depID)
	assert.NoError(t, err)
	assert.Equal(t, compR2.Sequence, int64(2))
}

func TestLoadDeployments(t *testing.T) {
	db := testBoltdb(t)

	comp1, err := db.Apply(&proto.Component{
		Name: "name1",
		Spec: proto.MustMarshalAny(&proto.ClusterSpec{}),
	})
	assert.NoError(t, err)
	assert.Equal(t, comp1.Sequence, int64(1))

	deps, err := db.ListDeployments()
	assert.NoError(t, err)
	assert.Equal(t, deps[0].Name, "name1")
}

func TestDependsOn_PendingComponent(t *testing.T) {
	db := testBoltdb(t)

	comp1, err := db.Apply(&proto.Component{
		Name: "name1",
		Spec: proto.MustMarshalAny(&proto.ClusterSpec{}),
	})
	assert.NoError(t, err)
	assert.Equal(t, comp1.Status, proto.Component_QUEUED)

	comp2, err := db.Apply(&proto.Component{
		Name: "name2",
		Spec: proto.MustMarshalAny(&proto.ClusterSpec{
			DependsOn: []string{
				"name1",
			},
		}),
	})
	assert.NoError(t, err)
	assert.Equal(t, comp2.Status, proto.Component_BLOCKED)

	task1 := db.queue2.popImpl()
	assert.NotNil(t, task1)
	assert.Equal(t, task1.ComponentID, comp1.Id)
	assert.Nil(t, db.queue2.popImpl())

	// it should trigger 'name2' component
	assert.NoError(t, db.Finalize(task1.DeploymentID))

	task2 := db.queue2.popImpl()
	assert.NotNil(t, task2)
	assert.Equal(t, task2.ComponentID, comp2.Id)

	assert.NoError(t, db.Finalize(task2.DeploymentID))

	// write another component and it should be queued inmediately
	comp3, err := db.Apply(&proto.Component{
		Name: "name3",
		Spec: proto.MustMarshalAny(&proto.ClusterSpec{
			DependsOn: []string{
				"name2",
			},
		}),
	})
	assert.NoError(t, err)
	assert.Equal(t, comp3.Status, proto.Component_QUEUED)
}

func TestDependsOn_CompletedComponent(t *testing.T) {
	db := testBoltdb(t)

	comp1, err := db.Apply(&proto.Component{
		Name: "name1",
		Spec: proto.MustMarshalAny(&proto.ClusterSpec{}),
	})
	assert.NoError(t, err)
	assert.Equal(t, comp1.Status, proto.Component_QUEUED)

	task1 := db.queue2.popImpl()
	assert.NotNil(t, task1)
	assert.NoError(t, db.Finalize(task1.DeploymentID))

	comp2, err := db.Apply(&proto.Component{
		Name: "name2",
		Spec: proto.MustMarshalAny(&proto.ClusterSpec{
			DependsOn: []string{
				"name1",
			},
		}),
	})
	assert.NoError(t, err)
	assert.Equal(t, comp2.Status, proto.Component_QUEUED)

	task2 := db.queue2.popImpl()
	assert.NotNil(t, task2)
}

func TestDependsOn_ComponentDoesNotExists(t *testing.T) {
	db := testBoltdb(t)

	_, err := db.Apply(&proto.Component{
		Name: "name1",
		Spec: proto.MustMarshalAny(&proto.ClusterSpec{
			DependsOn: []string{
				"comp-1",
			},
		}),
	})
	assert.Error(t, err)
}

func TestDeployment_UpsertInstance(t *testing.T) {
	db := testBoltdb(t)

	dep := &proto.Deployment{
		Id: "dep1",
	}
	assert.NoError(t, db.UpdateDeployment(dep))

	i0 := &proto.Instance{
		ID:           "i0",
		DeploymentID: "dep1",
	}
	assert.NoError(t, db.UpsertNode(i0))

	i0Res, err := db.LoadNode("i0")
	assert.NoError(t, err)
	assert.Equal(t, i0Res.ID, i0.ID)

	depRes, err := db.LoadDeployment("dep1")
	assert.NoError(t, err)
	assert.Len(t, depRes.Instances, 1)
	assert.Equal(t, depRes.Instances[0].ID, "i0")
}

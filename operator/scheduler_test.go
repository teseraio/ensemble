package operator

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/teseraio/ensemble/lib/uuid"
	"github.com/teseraio/ensemble/operator/proto"
)

type mockDeployment struct {
	*proto.Deployment
}

func newMockDeployment() *mockDeployment {
	return &mockDeployment{
		Deployment: &proto.Deployment{
			Instances: []*proto.Instance{},
		},
	}
}

func (p *mockDeployment) Copy() *mockDeployment {
	pp := &mockDeployment{}
	pp.Deployment = p.Deployment.Copy()
	return pp
}

func mockClusterSpec() *proto.ClusterSpec {
	return &proto.ClusterSpec{
		Groups: []*proto.ClusterSpec_Group{
			{},
		},
	}
}

type expectedReconciler struct {
	place      int
	update     int
	reschedule int
	stop       int
	ready      int
	out        int
	done       bool
}

func testExpectReconcile(t *testing.T, reconciler *reconciler, expect expectedReconciler) {
	var place, update, reschedule int
	for _, i := range reconciler.res.place {
		if i.update {
			update++
		} else if i.reschedule {
			reschedule++
		} else {
			place++
		}
	}
	if place != expect.place {
		t.Fatalf("wrong place: %d %d", place, expect.place)
	}
	if update != expect.update {
		t.Fatalf("wrong update: %d %d", update, expect.update)
	}
	if reschedule != expect.reschedule {
		t.Fatalf("wrong reschedule: %d %d", reschedule, expect.reschedule)
	}
	if expect.stop != len(reconciler.res.stop) {
		t.Fatalf("wrong stop: %d %d", expect.stop, len(reconciler.res.stop))
	}
	if expect.ready != len(reconciler.res.ready) {
		t.Fatalf("wrong promote: %d %d", expect.ready, len(reconciler.res.ready))
	}
	if expect.out != len(reconciler.res.out) {
		t.Fatalf("wrong out: %d %d", expect.out, len(reconciler.res.out))
	}
	if expect.done != reconciler.res.done {
		t.Fatalf("incorrect done: %v %v", expect.done, reconciler.res.done)
	}
}

func TestReconciler_Place_Empty(t *testing.T) {
	spec := mockClusterSpec()
	spec.Groups[0].Count = 5

	dep := newMockDeployment()

	rec := &reconciler{
		dep:  dep.Deployment,
		spec: spec,
	}
	rec.Compute()

	testExpectReconcile(t, rec, expectedReconciler{
		place: 5,
	})
}

func TestReconciler_ScaleUp_Blocked(t *testing.T) {
	spec := mockClusterSpec()
	spec.Groups[0].Count = 10

	dep := newMockDeployment()
	for i := 0; i < 5; i++ {
		ii := &proto.Instance{}
		if i%2 == 0 {
			ii.Status = proto.Instance_PENDING
		} else {
			ii.Status = proto.Instance_SCHEDULED
		}
		ii.ID = uuid.UUID()
		ii.Group = spec.Groups[0]
		ii.Healthy = false // healthy is only possible is the instance is running
		dep.Instances = append(dep.Instances, ii)
	}

	rec := &reconciler{
		dep:  dep.Deployment,
		spec: spec,
	}
	rec.Compute()

	// we cannot scale up until all the instances are healthy
	testExpectReconcile(t, rec, expectedReconciler{
		place: 0,
	})
}

func TestReconciler_ScaleUp(t *testing.T) {
	spec := mockClusterSpec()
	spec.Groups[0].Count = 10

	dep := newMockDeployment()
	for i := 0; i < 5; i++ {
		ii := &proto.Instance{}
		ii.Status = proto.Instance_RUNNING
		ii.ID = uuid.UUID()
		ii.Group = spec.Groups[0]
		ii.Healthy = true
		dep.Instances = append(dep.Instances, ii)
	}

	rec := &reconciler{
		dep:  dep.Deployment,
		spec: spec,
	}
	rec.Compute()

	testExpectReconcile(t, rec, expectedReconciler{
		place: 5,
	})
}

func TestReconciler_ScaleDown(t *testing.T) {
	// 10 -> 5
	spec := mockClusterSpec()
	spec.Groups[0].Count = 5

	dep := newMockDeployment()
	for i := 0; i < 10; i++ {
		ii := &proto.Instance{}
		ii.Status = proto.Instance_RUNNING
		ii.ID = uuid.UUID()
		ii.Group = spec.Groups[0]
		ii.Healthy = true
		dep.Instances = append(dep.Instances, ii)
	}

	rec := &reconciler{
		dep:  dep.Deployment,
		spec: spec,
	}
	rec.Compute()

	testExpectReconcile(t, rec, expectedReconciler{
		stop: 5,
		done: false,
	})

	// second eval
	// update half the instances to pending stopped
	for i := 0; i < 5; i++ {
		dep.Instances[i].Status = proto.Instance_TAINTED
	}
	rec.Compute()

	testExpectReconcile(t, rec, expectedReconciler{
		stop: 0,
		done: false,
	})

	// third eval
	// update half the instances to stopped, they should
	// be removed of the ensemble
	for i := 0; i < 3; i++ {
		dep.Instances[i].Status = proto.Instance_STOPPED
	}
	rec.Compute()

	testExpectReconcile(t, rec, expectedReconciler{
		out:  3,
		done: false,
	})
}

func TestReconciler_ScaleDown_Complete(t *testing.T) {
	// 10 -> 5
	spec := mockClusterSpec()
	spec.Groups[0].Count = 5

	dep := newMockDeployment()
	for i := 0; i < 10; i++ {
		ii := &proto.Instance{}
		ii.Status = proto.Instance_RUNNING
		ii.ID = uuid.UUID()
		ii.Group = spec.Groups[0]
		ii.Healthy = true
		dep.Instances = append(dep.Instances, ii)
	}

	// stop 5 instances
	for i := 0; i < 5; i++ {
		dep.Instances[i].Status = proto.Instance_STOPPED
	}

	rec := &reconciler{
		dep:  dep.Deployment,
		spec: spec,
	}
	rec.Compute()

	testExpectReconcile(t, rec, expectedReconciler{
		out:  5,
		done: true,
	})
}

func TestReconciler_ScaleDown_Zero(t *testing.T) {
	// group to zero
	spec := mockClusterSpec()
	spec.Groups[0].Count = 0

	dep := newMockDeployment()
	for i := 0; i < 5; i++ {
		ii := &proto.Instance{}
		ii.Status = proto.Instance_RUNNING
		ii.ID = uuid.UUID()
		ii.Group = spec.Groups[0]
		ii.Healthy = true
		dep.Instances = append(dep.Instances, ii)
	}

	rec := &reconciler{
		dep:  dep.Deployment,
		spec: spec,
	}
	rec.Compute()

	testExpectReconcile(t, rec, expectedReconciler{
		stop: 5,
	})
}

func TestReconciler_MultipleGroup_Unblock(t *testing.T) {
	// unblock the next group
	spec := mockClusterSpec()
	spec.Groups[0].Count = 5
	spec.Groups[0].Type = "typ1"

	// add a second group
	spec.Groups = append(spec.Groups, &proto.ClusterSpec_Group{
		Count: 5,
		Type:  "typ2",
	})

	dep := newMockDeployment()

	for i := 0; i < 5; i++ {
		ii := &proto.Instance{}
		ii.ID = uuid.UUID()
		ii.Group = spec.Groups[0]
		ii.Healthy = true
		dep.Instances = append(dep.Instances, ii)
	}

	rec := &reconciler{
		dep:  dep.Deployment,
		spec: spec,
	}
	rec.Compute()

	// one last item promoted
	testExpectReconcile(t, rec, expectedReconciler{
		place: 5,
	})
}

func TestReconciler_Purge(t *testing.T) {
	// First eval: Remove the whole cluster
	spec := mockClusterSpec()

	dep := newMockDeployment()
	for i := 0; i < 5; i++ {
		ii := &proto.Instance{}
		ii.Status = proto.Instance_RUNNING
		ii.ID = uuid.UUID()
		ii.Group = spec.Groups[0]
		ii.Healthy = true
		dep.Instances = append(dep.Instances, ii)
	}

	rec := &reconciler{
		delete: true,
		dep:    dep.Deployment,
		spec:   spec,
	}
	rec.Compute()

	testExpectReconcile(t, rec, expectedReconciler{
		stop: 5,
	})

	// Second eval. Do not remove tainted instances
	for i := 0; i < 5; i++ {
		dep.Instances[i].Status = proto.Instance_TAINTED
	}

	rec = &reconciler{
		delete: true,
		dep:    dep.Deployment,
		spec:   spec,
	}
	rec.Compute()

	testExpectReconcile(t, rec, expectedReconciler{
		stop: 0,
	})

	// Third eval: Done=true
	for i := 0; i < 5; i++ {
		dep.Instances[i].Status = proto.Instance_STOPPED
	}

	rec = &reconciler{
		delete: true,
		dep:    dep.Deployment,
		spec:   spec,
	}
	rec.Compute()

	testExpectReconcile(t, rec, expectedReconciler{
		done: true,
	})
}

func TestReconciler_RollingUpgradeX(t *testing.T) {
	// 5 (1) -> 5 (2)
	spec0 := mockClusterSpec()
	spec0.Groups[0].Count = 5

	spec1 := spec0.Copy()
	spec1.Sequence++
	spec1.Groups[0].Resources = map[string]string{"A": "B"}

	dep := newMockDeployment()
	for i := 0; i < 5; i++ {
		ii := &proto.Instance{}
		ii.Status = proto.Instance_RUNNING
		ii.ID = uuid.UUID()
		ii.Group = spec0.Groups[0]
		ii.Healthy = true
		dep.Instances = append(dep.Instances, ii)
	}

	rec := &reconciler{
		dep:  dep.Deployment,
		spec: spec1,
	}
	rec.Compute()

	// two instance are stopped
	testExpectReconcile(t, rec, expectedReconciler{
		stop: 2,
	})

	// another reconcile should not stop the pending instances
	for i := 0; i < 2; i++ {
		dep.Instances[i].Status = proto.Instance_TAINTED
		dep.Instances[i].Canary = true
	}
	rec.Compute()

	testExpectReconcile(t, rec, expectedReconciler{
		stop: 0,
	})
}

func TestReconciler_RollingUpgrade_SecondEval(t *testing.T) {
	// Second evaluation for the rolling update
	spec0 := mockClusterSpec()
	spec0.Groups[0].Count = 5

	spec1 := spec0.Copy()
	spec1.Sequence++
	spec1.Groups[0].Resources = map[string]string{"A": "B"}

	dep := newMockDeployment()
	for i := 0; i < 3; i++ {
		ii := &proto.Instance{}
		ii.ID = uuid.UUID()
		ii.Status = proto.Instance_RUNNING
		ii.Group = spec0.Groups[0]
		ii.Status = proto.Instance_RUNNING
		dep.Instances = append(dep.Instances, ii)
	}

	// 2 instance stopped and canary
	for i := 3; i < 5; i++ {
		ii := &proto.Instance{}
		ii.ID = uuid.UUID()
		ii.Name = uuid.UUID()
		ii.Status = proto.Instance_STOPPED
		ii.Canary = true
		ii.Group = spec0.Groups[0]
		dep.Instances = append(dep.Instances, ii)
	}

	rec := &reconciler{
		dep:  dep.Deployment,
		spec: spec1,
	}
	rec.Compute()

	// two instance are stopped
	testExpectReconcile(t, rec, expectedReconciler{
		out:    2,
		update: 2,
		done:   false,
	})
}

func TestReconciler_RollingUpgrade_PartialPromote(t *testing.T) {
	spec0 := mockClusterSpec()
	spec0.Groups[0].Count = 5

	spec1 := spec0.Copy()
	spec1.Sequence++
	spec1.Groups[0].Resources = map[string]string{"A": "B"}

	dep := newMockDeployment()

	old := 3

	// 3 instances with the old revision and 2 with the new but only
	// one of them is healthy and can be promoted
	for i := 0; i < 5; i++ {
		ii := &proto.Instance{}
		ii.ID = uuid.UUID()
		ii.Status = proto.Instance_RUNNING

		if i < old {
			ii.Group = spec0.Groups[0]
		} else {
			ii.Group = spec1.Groups[0]
			ii.Canary = true
		}
		ii.Healthy = true
		dep.Instances = append(dep.Instances, ii)
	}

	// only one canary instance is healthy
	dep.Instances[4].Healthy = false

	rec := &reconciler{
		dep:  dep.Deployment,
		spec: spec1,
	}
	rec.Compute()

	// wait until the two instances are running to unblock
	// the next updates
	testExpectReconcile(t, rec, expectedReconciler{
		update: 0,
		ready:  1,
	})

	// Second eval: the second canary instance is healthy
	// it triggers the new rolling updates

	dep2 := dep.Copy()
	dep2.Instances[3].Canary = false // promote
	dep2.Instances[4].Healthy = true // healthy

	rec = &reconciler{
		dep:  dep2.Deployment,
		spec: spec1,
	}
	rec.Compute()

	testExpectReconcile(t, rec, expectedReconciler{
		stop:  2,
		ready: 1,
	})
}

func TestReconciler_RollingUpgrade_Complete(t *testing.T) {
	spec0 := mockClusterSpec()
	spec0.Groups[0].Count = 5

	dep := newMockDeployment()

	for i := 0; i < 5; i++ {
		ii := &proto.Instance{}
		ii.Status = proto.Instance_RUNNING
		ii.ID = uuid.UUID()
		ii.Group = spec0.Groups[0]
		ii.Healthy = true
		dep.Instances = append(dep.Instances, ii)
	}

	// one of the instances is a canary to be promoted
	dep.Instances[4].Canary = true

	rec := &reconciler{
		dep:  dep.Deployment,
		spec: spec0,
	}
	rec.Compute()

	// one last item promoted
	testExpectReconcile(t, rec, expectedReconciler{
		ready: 1,
		done:  true,
	})
}

func TestReconciler_RollingUpgrade_ScaleUp(t *testing.T) {
	// First we do rolling update and then scale
	spec0 := mockClusterSpec()
	spec0.Groups[0].Count = 5

	spec1 := spec0.Copy()
	spec1.Sequence++
	spec1.Groups[0].Count = 8
	spec1.Groups[0].Resources = map[string]string{"A": "B"}

	dep := newMockDeployment()

	for i := 0; i < 5; i++ {
		ii := &proto.Instance{}
		ii.Status = proto.Instance_RUNNING
		ii.ID = uuid.UUID()
		ii.Group = spec0.Groups[0]
		ii.Healthy = true
		dep.Instances = append(dep.Instances, ii)
	}

	rec := &reconciler{
		dep:  dep.Deployment,
		spec: spec1,
	}
	rec.Compute()

	testExpectReconcile(t, rec, expectedReconciler{
		stop: 2,
	})
}

func TestReconciler_RollingUpgrade_ScaleDown(t *testing.T) {
	// stop the instances before the rolling update
	spec0 := mockClusterSpec()
	spec0.Groups[0].Count = 5

	spec1 := spec0.Copy()
	spec1.Sequence++
	spec1.Groups[0].Count = 3
	spec1.Groups[0].Resources = map[string]string{"A": "B"}

	dep := newMockDeployment()

	for i := 0; i < 5; i++ {
		ii := &proto.Instance{}
		ii.Status = proto.Instance_RUNNING
		ii.ID = uuid.UUID()
		ii.Group = spec0.Groups[0]
		ii.Healthy = true
		dep.Instances = append(dep.Instances, ii)
	}

	rec := &reconciler{
		dep:  dep.Deployment,
		spec: spec1,
	}
	rec.Compute()

	testExpectReconcile(t, rec, expectedReconciler{
		stop: 2,
	})
	// assert.False(t, rec.done)
}

func TestReconciler_InstanceFailed_Restart(t *testing.T) {
	spec := mockClusterSpec()
	spec.Groups[0].Count = 5

	dep := newMockDeployment()

	for i := 0; i < 5; i++ {
		ii := &proto.Instance{}
		ii.Status = proto.Instance_RUNNING
		ii.ID = uuid.UUID()
		ii.Group = spec.Groups[0]
		ii.Healthy = true
		dep.Instances = append(dep.Instances, ii)
	}

	// one instance has failed
	dep.Instances[0].Status = proto.Instance_FAILED

	rec := &reconciler{
		dep:  dep.Deployment,
		spec: spec,
	}
	rec.Compute()

	testExpectReconcile(t, rec, expectedReconciler{
		reschedule: 1,
	})
}

func TestReconciler_InstanceFailed_MaxAttempts(t *testing.T) {
	spec := mockClusterSpec()
	spec.Groups[0].Count = 5

	dep := newMockDeployment()

	for i := 0; i < 5; i++ {
		ii := &proto.Instance{}
		ii.Status = proto.Instance_RUNNING
		ii.ID = uuid.UUID()
		ii.Group = spec.Groups[0]
		ii.Healthy = true
		dep.Instances = append(dep.Instances, ii)
	}

	// one instance has failed
	dep.Instances[0].Status = proto.Instance_FAILED
	dep.Instances[0].Reschedule = &proto.Instance_Reschedule{
		Attempts: 3,
	}

	rec := &reconciler{
		dep:  dep.Deployment,
		spec: spec,
	}
	rec.Compute()

	// the lost instance is not rescheduled
	testExpectReconcile(t, rec, expectedReconciler{
		reschedule: 0,
	})
}

func TestReconciler_InstanceFailed_ScaleDown(t *testing.T) {
	// scale down will pick one of the failed instances first
	spec := mockClusterSpec()
	spec.Groups[0].Count = 3

	dep := newMockDeployment()

	for i := 0; i < 5; i++ {
		ii := &proto.Instance{}
		ii.Status = proto.Instance_RUNNING
		ii.ID = uuid.UUID()
		ii.Group = spec.Groups[0]
		ii.Healthy = true
		dep.Instances = append(dep.Instances, ii)
	}

	// three instance failed, one is rescheduled, two are removed
	dep.Instances[2].Status = proto.Instance_FAILED
	dep.Instances[3].Status = proto.Instance_FAILED
	dep.Instances[4].Status = proto.Instance_FAILED

	rec := &reconciler{
		dep:  dep.Deployment,
		spec: spec,
	}
	rec.Compute()

	testExpectReconcile(t, rec, expectedReconciler{
		stop:       2,
		reschedule: 1,
	})

	assert.Equal(t, rec.res.stop[0].instance.ID, dep.Instances[2].ID)
	assert.Equal(t, rec.res.stop[1].instance.ID, dep.Instances[3].ID)
}

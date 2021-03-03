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

func (p *mockDeployment) union(ii []*proto.Instance) {
	for _, i := range ii {
		found := false
		for indx, j := range p.Instances {
			if j.ID == i.ID {
				// remove the current instance
				p.Instances[indx] = i
				found = true
				break
			}
		}
		if !found {
			p.Instances = append(p.Instances, i)
		}
	}
}

func mockClusterSpec() *proto.ClusterSpec {
	return &proto.ClusterSpec{
		Groups: []*proto.ClusterSpec_Group{
			{},
		},
	}
}

func TestReconciler_Place_Empty(t *testing.T) {
	spec := mockClusterSpec()
	spec.Groups[0].Count = 5

	dep := newMockDeployment()

	reconciler := &reconciler{
		dep:  dep.Deployment,
		spec: spec,
	}
	reconciler.Compute()

	assert.Equal(t, reconciler.check("add"), 5)
}

func TestReconciler_ScaleUp(t *testing.T) {
	spec := mockClusterSpec()
	spec.Groups[0].Count = 10

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

	place := rec.gather("add")
	assert.Len(t, place, 5)

	dep2 := dep.Copy()
	dep2.Instances = append(dep2.Instances, place...)

	rec = &reconciler{
		dep:  dep.Deployment,
		spec: spec,
	}
	assert.Len(t, rec.res, 0)
}

func TestReconciler_ScaleDownX(t *testing.T) {
	// 10 -> 5
	spec := mockClusterSpec()
	spec.Groups[0].Count = 5

	dep := newMockDeployment()
	for i := 0; i < 10; i++ {
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

	stop := rec.gather("stop")
	assert.Len(t, stop, 5)

	// the desired state has changed
	for _, i := range stop {
		if i.Desired != "Stop" {
			t.Fatal("bad")
		}
	}

	dep2 := dep.Copy()
	dep2.union(stop)

	rec = &reconciler{
		dep:  dep2.Deployment,
		spec: spec,
	}
	rec.Compute()

	assert.Len(t, rec.res, 0)
}

func TestReconciler_ScaleDown_Zero(t *testing.T) {
	// group to zero
	spec := mockClusterSpec()
	spec.Groups[0].Count = 0

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
	assert.Len(t, rec.gather("stop"), 5)
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
	assert.Len(t, rec.gather("add"), 5)
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

	stopping := rec.gather("stop")
	assert.Len(t, stopping, 5)

	// Second eval: wait for all the instance to stop
	dep2 := dep.Copy()
	dep2.union(stopping)

	rec = &reconciler{
		delete: true,
		dep:    dep2.Deployment,
		spec:   spec,
	}
	rec.Compute()

	assert.Len(t, rec.res, 0)
	assert.False(t, rec.done)

	// Third eval: the deployment is done when there are no more running nodes
	dep3 := dep.Copy()
	dep3.Instances = []*proto.Instance{}

	rec = &reconciler{
		delete: true,
		dep:    dep3.Deployment,
		spec:   spec,
	}
	rec.Compute()

	assert.True(t, rec.done)
}

func TestReconciler_RollingUpgrade(t *testing.T) {
	// 5 (1) -> 5 (2)
	spec0 := mockClusterSpec()
	spec0.Groups[0].Count = 5

	spec1 := spec0.Copy()
	spec1.Sequence++
	spec1.Groups[0].Resources = map[string]string{"A": "B"}

	dep := newMockDeployment()
	for i := 0; i < 5; i++ {
		ii := &proto.Instance{}
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

	update := rec.gather("update")
	assert.Len(t, update, 2)

	for _, i := range update {
		if !i.Canary {
			t.Fatal("bad")
		}
	}

	// Second eval: we have to wait for the healthy instances
	// to unblock more updates

	dep2 := dep.Copy()
	dep2.union(update)

	rec = &reconciler{
		dep:  dep2.Deployment,
		spec: spec1,
	}
	rec.Compute()

	update = rec.gather("update")
	assert.Len(t, update, 0)
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
	update := rec.gather("update")
	assert.Len(t, update, 0)

	promote := rec.gather("promote")
	assert.Len(t, promote, 1)

	// Second eval: the second canary instance is healthy
	// it triggers the new rolling updates

	dep2 := dep.Copy()
	dep2.union(promote)
	dep2.Instances[4].Healthy = true

	rec = &reconciler{
		dep:  dep2.Deployment,
		spec: spec1,
	}
	rec.Compute()

	assert.Len(t, rec.gather("update"), 2)
	assert.Len(t, rec.gather("promote"), 1)

	assert.False(t, rec.done)
}

func TestReconciler_RollingUpgrade_Complete(t *testing.T) {
	spec0 := mockClusterSpec()
	spec0.Groups[0].Count = 5

	dep := newMockDeployment()

	for i := 0; i < 5; i++ {
		ii := &proto.Instance{}
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
	assert.Len(t, rec.gather("promote"), 1)
	assert.True(t, rec.done)
}

func TestReconciler_RollingUpgrade_ScaleUp(t *testing.T) {
	// we need to figure out what to do here
}

func TestReconciler_RollingUpgrade_ScaleDown(t *testing.T) {
	// similar to now
}

func TestReconciler_InstanceFailed_Restart(t *testing.T) {
	spec := mockClusterSpec()
	spec.Groups[0].Count = 5

	dep := newMockDeployment()

	for i := 0; i < 5; i++ {
		ii := &proto.Instance{}
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

	assert.Len(t, rec.gather("reschedule"), 1)
}

func TestReconciler_InstanceFailed_MaxAttempts(t *testing.T) {
	spec := mockClusterSpec()
	spec.Groups[0].Count = 5

	dep := newMockDeployment()

	for i := 0; i < 5; i++ {
		ii := &proto.Instance{}
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
	assert.Len(t, rec.gather("reschedule"), 0)
}

func TestReconciler_InstanceFailed_ScaleDown(t *testing.T) {
	// scale down will pick one of the failed instances first
	spec := mockClusterSpec()
	spec.Groups[0].Count = 3

	dep := newMockDeployment()

	for i := 0; i < 5; i++ {
		ii := &proto.Instance{}
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

	stop := rec.gather("stop")
	assert.Len(t, stop, 2)

	// two failed instances are lost
	assert.Equal(t, stop[0].ID, dep.Instances[2].ID)
	assert.Equal(t, stop[1].ID, dep.Instances[3].ID)

	// one instance needs to be rescheduled
	reschedule := rec.gather("reschedule")
	assert.Len(t, reschedule, 1)
}

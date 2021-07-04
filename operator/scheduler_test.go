package operator

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/teseraio/ensemble/lib/uuid"
	"github.com/teseraio/ensemble/operator/proto"
	"github.com/teseraio/ensemble/schema"
)

func TestScheduler_EvalInstanceFailed(t *testing.T) {
	spec := mockClusterSpec()
	spec.Groups[0].Count = 3

	dep := newMockDeployment()
	for i := 0; i < 2; i++ {
		ii := &proto.Instance{}
		ii.Status = proto.Instance_RUNNING
		ii.ID = uuid.UUID()
		ii.Group = spec.Groups[0]
		ii.Healthy = true
		dep.Instances = append(dep.Instances, ii)
	}
	// one instance has failed
	dep.Instances[0].Status = proto.Instance_FAILED
	dep.CompId = "a"

	harness := NewHarness(t)
	harness.Deployment = dep.Deployment
	harness.Handler = &nullHandler{}

	harness.AddComponent(&proto.Component{
		Id:   "a",
		Spec: proto.MustMarshalAny(spec),
	})

	sched := NewScheduler(harness)

	plan, err := sched.Process(&proto.Evaluation{
		Id: uuid.UUID(),
	})
	assert.NoError(t, err)
	assert.Equal(t, plan.NodeUpdate[0].Name, dep.Instances[0].Name)
}

// TODO: We need to update the sequence of the other instances even if there are not changes

func TestScheduler_InstanceUpdateSequence(t *testing.T) {
	// 5 (1) -> 5 (2)
	spec0 := mockClusterSpec()
	spec0.Groups[0].Count = 5

	spec1 := spec0.Copy()
	spec1.Sequence++
	spec1.Groups[0].Resources = schema.MapToSpec(map[string]interface{}{"A": "B"})

	dep := newMockDeployment()
	for i := 0; i < 3; i++ {
		ii := &proto.Instance{}
		ii.Status = proto.Instance_RUNNING
		ii.ID = uuid.UUID()
		ii.Group = spec0.Groups[0]
		ii.Healthy = true
		dep.Instances = append(dep.Instances, ii)
	}

	// 4 and 5 have stopped and ready to be replaced
	for i := 3; i < 5; i++ {
		ii := &proto.Instance{}
		ii.ID = uuid.UUID()
		ii.Name = uuid.UUID()
		ii.Status = proto.Instance_STOPPED
		ii.Canary = true
		ii.Group = spec0.Groups[0]
		dep.Instances = append(dep.Instances, ii)
	}

	harness := NewHarness(t)
	harness.Deployment = dep.Deployment
	harness.Handler = &nullHandler{}

	harness.AddComponent(&proto.Component{
		Id:       "a",
		Sequence: 1,
		Spec:     proto.MustMarshalAny(spec1),
	})
	sched := NewScheduler(harness)

	plan, err := sched.Process(&proto.Evaluation{
		Id: uuid.UUID(),
	})
	assert.NoError(t, err)

	// nodes updated should have the new sequence=1
	for _, n := range plan.NodeUpdate {
		if n.Status == proto.Instance_PENDING {
			assert.Equal(t, n.Sequence, int64(1))
		}
	}
}

func TestScheduler_NewInstancesSequence(t *testing.T) {
	// 5 (1) -> 5 (2)
	spec0 := mockClusterSpec()
	spec0.Groups[0].Count = 5

	spec1 := spec0.Copy()
	spec1.Sequence++
	spec1.Groups[0].Count = 8

	dep := newMockDeployment()
	for i := 0; i < 3; i++ {
		ii := &proto.Instance{}
		ii.Status = proto.Instance_RUNNING
		ii.ID = uuid.UUID()
		ii.Group = spec0.Groups[0]
		ii.Healthy = true
		dep.Instances = append(dep.Instances, ii)
	}

	harness := NewHarness(t)
	harness.Deployment = dep.Deployment
	harness.Handler = &nullHandler{}

	harness.AddComponent(&proto.Component{
		Id:       "a",
		Sequence: 1,
		Spec:     proto.MustMarshalAny(spec1),
	})
	sched := NewScheduler(harness)

	plan, err := sched.Process(&proto.Evaluation{
		Id: uuid.UUID(),
	})
	assert.NoError(t, err)

	for _, i := range plan.NodeUpdate {
		assert.Equal(t, i.Sequence, int64(1))
	}
}

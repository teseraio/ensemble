package operator

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/teseraio/ensemble/lib/uuid"
	"github.com/teseraio/ensemble/operator/proto"
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

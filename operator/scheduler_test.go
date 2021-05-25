package operator

import (
	"testing"

	"github.com/teseraio/ensemble/operator/proto"
)

func TestScheduler_Place(t *testing.T) {
	eval := &proto.Evaluation{}

	sched := &scheduler{
		state: &mockState{},
	}
	sched.Process(eval)
}

type mockState struct {
	plan *proto.Plan
}

func (m *mockState) SubmitPlan(plan *proto.Plan) error {
	m.plan = plan
	return nil
}

func (m *mockState) LoadDeployment(id string) (*proto.Deployment, error) {
	return nil, nil
}

func (m *mockState) GetHandler(id string) (Handler, error) {
	return nil, nil
}

func (m *mockState) GetClusterSpec(id string, sequence int64) (*proto.ClusterSpec, *proto.Component, error) {
	return nil, nil, nil
}

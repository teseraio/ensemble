package operator

import (
	"fmt"

	gproto "github.com/golang/protobuf/proto"
	"github.com/teseraio/ensemble/lib/uuid"
	"github.com/teseraio/ensemble/operator/proto"
)

type schedState interface {
	LoadDeployment(id string) (*proto.Deployment, error)
	GetComponentByID(deployment string, id string, sequence int64) (*proto.Component, error)
	GetHandler(id string) (Handler, error)
}

type Scheduler interface {
	Process(eval *proto.Evaluation) (*proto.Plan, error)
}

func NewScheduler(state schedState) Scheduler {
	return &scheduler{state: state}
}

type scheduler struct {
	state schedState
}

func (s *scheduler) Process(eval *proto.Evaluation) (*proto.Plan, error) {
	// get the deployment
	dep, err := s.state.LoadDeployment(eval.DeploymentID)
	if err != nil {
		return nil, err
	}

	handler, err := s.state.GetHandler(dep.Backend)
	if err != nil {
		return nil, err
	}

	// TODO: XXX
	comp, err := s.state.GetComponentByID(eval.DeploymentID, dep.CompId, dep.Sequence)
	if err != nil {
		return nil, err
	}
	spec := &proto.ClusterSpec{}
	if err := gproto.Unmarshal(comp.Spec.Value, spec); err != nil {
		return nil, err
	}

	// we need this here because is not set before in the spec
	spec.Sequence = comp.Sequence

	r := &reconciler{
		delete: comp.Action == proto.Component_DELETE,
		dep:    dep,
		spec:   spec,
	}
	r.Compute()

	plan := &proto.Plan{
		EvalID:     eval.Id,
		NodeUpdate: []*proto.Instance{},
	}

	// out instances
	for _, i := range r.res.out {
		ii := i.Copy()
		ii.Status = proto.Instance_OUT

		plan.NodeUpdate = append(plan.NodeUpdate, ii)
	}

	// promote instances
	for _, i := range r.res.ready {
		ii := i.Copy()
		ii.Canary = false

		fmt.Println("-- ready --")
		fmt.Println(ii.Status)

		plan.NodeUpdate = append(plan.NodeUpdate, ii)
	}

	// stop instances
	for _, i := range r.res.stop {
		ii := i.instance.Copy()
		ii.Status = proto.Instance_TAINTED
		ii.DesiredStatus = proto.Instance_STOP
		ii.Canary = i.update

		plan.NodeUpdate = append(plan.NodeUpdate, ii)
	}

	if len(r.res.place) != 0 {
		// create a cluster object to initialize the instances
		placeInstances := []*proto.Instance{}
		for _, i := range r.res.place {
			var name string
			if i.instance == nil {
				name = i.name
			} else {
				name = i.instance.Name
			}

			ii := &proto.Instance{}
			ii.ID = uuid.UUID()
			ii.Group = i.group
			ii.Spec = &proto.NodeSpec{}
			ii.ClusterName = dep.Name
			ii.DeploymentID = dep.Id
			ii.DnsSuffix = dep.DnsSuffix
			ii.Name = name
			ii.Status = proto.Instance_PENDING
			ii.Canary = i.update

			placeInstances = append(placeInstances, ii)
		}

		placeInstances, err := handler.ApplyNodes(placeInstances, dep.Instances)
		if err != nil {
			return nil, err
		}
		plan.NodeUpdate = append(plan.NodeUpdate, placeInstances...)
	}

	if r.res.completed {
		plan.Done = true
		plan.Status = proto.DeploymentCompleted
	} else if r.res.done {
		if dep.Status != proto.DeploymentDone {
			plan.Done = true
			plan.Status = proto.DeploymentDone
		}
	} else {
		plan.Status = proto.DeploymentRunning
	}

	plan.Deployment = dep
	return plan, nil
}

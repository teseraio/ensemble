package operator

import (
	gproto "github.com/golang/protobuf/proto"
	"github.com/teseraio/ensemble/lib/uuid"
	"github.com/teseraio/ensemble/operator/proto"
)

type schedState interface {
	SubmitPlan(*proto.Plan) error
	LoadDeployment(id string) (*proto.Deployment, error)
	GetComponentByID(id string) (*proto.Component, error)
	GetHandler(id string) (Handler, error)
}

type Harness struct {
	Plan       *proto.Plan
	Deployment *proto.Deployment
	Handler    Handler
	Spec       *proto.ClusterSpec
}

type NodeExpect struct {
	Env map[string]string
}

func (h *Harness) ExpectNodeUpdate(e []NodeExpect) {

}

func (h *Harness) SubmitPlan(plan *proto.Plan) error {
	h.Plan = plan
	return nil
}

func (h *Harness) LoadDeployment(id string) (*proto.Deployment, error) {
	return h.Deployment, nil
}

func (h *Harness) GetHandler(id string) (Handler, error) {
	return h.Handler, nil
}

func (h *Harness) GetClusterSpec(id string, sequence int64) (*proto.ClusterSpec, *proto.Component, error) {
	return h.Spec, &proto.Component{}, nil
}

type Scheduler interface {
	Process(eval *proto.Evaluation) error
}

func NewScheduler(state schedState) Scheduler {
	return &scheduler{state: state}
}

type scheduler struct {
	state schedState
}

func (s *scheduler) Process(eval *proto.Evaluation) error {
	// get the deployment
	dep, err := s.state.LoadDeployment(eval.ClusterID)
	if err != nil {
		return err
	}
	handler, err := s.state.GetHandler(dep.Backend)
	if err != nil {
		return err
	}

	comp, err := s.state.GetComponentByID(dep.CompId)
	if err != nil {
		return err
	}
	spec := &proto.ClusterSpec{}
	if err := gproto.Unmarshal(comp.Spec.Value, spec); err != nil {
		return err
	}

	// we need this here because is not set before in the spec
	spec.Name = eval.ClusterID
	spec.Sequence = dep.Sequence

	r := &reconciler{
		delete: comp.Action == proto.Component_DELETE,
		dep:    dep,
		spec:   spec,
	}
	r.Compute()

	plan := &proto.Plan{
		EvalID:      eval.Id,
		NodeUpdate:  []*proto.Instance{},
		NodeInplace: []*proto.Instance{},
	}

	// out instances
	for _, i := range r.res.out {
		ii := i.Copy()
		ii.Status = proto.Instance_OUT

		plan.NodeInplace = append(plan.NodeInplace, ii)
	}

	// promote instances
	for _, i := range r.res.ready {
		ii := i.Copy()
		ii.Canary = false

		plan.NodeInplace = append(plan.NodeInplace, ii)
	}

	// stop instances
	for _, i := range r.res.stop {
		ii := i.instance.Copy()
		ii.Status = proto.Instance_TAINTED
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
			ii.Cluster = spec.Name
			ii.Name = name
			ii.Status = proto.Instance_PENDING
			ii.Canary = i.update

			placeInstances = append(placeInstances, ii)
		}

		placeInstances, err := handler.ApplyNodes(placeInstances, dep.Instances)
		if err != nil {
			return err
		}
		plan.NodeUpdate = append(plan.NodeUpdate, placeInstances...)
	}

	if r.res.done {
		if dep.Status != proto.DeploymentDone {
			plan.Status = proto.DeploymentDone
		}
	} else {
		plan.Status = proto.DeploymentRunning
	}

	if err := s.state.SubmitPlan(plan); err != nil {
		return err
	}
	return nil
}

package operator

import (
	"bytes"
	"fmt"
	"text/template"

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
	Scheduler  Scheduler
	Component  *proto.Component
}

type NodeExpect struct {
	Env map[string]string
}

func Assert(dep *proto.Deployment, n *proto.Instance, expected NodeExpect) {

	applyTmpl := func() func(v string) string {
		obj := map[string]interface{}{}
		for _, i := range dep.Instances {
			indx, err := proto.ParseIndex(i.Name)
			if err != nil {
				panic(err)
			}
			obj[fmt.Sprintf("Node%d", indx)] = i.Name
		}
		return func(v string) string {
			t, err := template.New("").Parse(v)
			if err != nil {
				panic(err)
			}
			buf1 := new(bytes.Buffer)
			if err = t.Execute(buf1, obj); err != nil {
				panic(err)
			}
			return buf1.String()
		}
	}

	tmpl := applyTmpl()

	// env vars
	for k, v := range expected.Env {
		if n.Spec.Env[k] != tmpl(v) {
			panic("BAD")
		}
	}
}

func (h *Harness) ApplyDep() *proto.Deployment {
	dep := h.Deployment.Copy()

	for _, n := range h.Plan.NodeUpdate {
		dep.Instances = append(dep.Instances, n)
	}

	return dep
}

func (h *Harness) GetComponentByID(id string) (*proto.Component, error) {
	return h.Component, nil
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

func (h *Harness) ApplySched(comp *proto.Component) *proto.Deployment {
	h.Component = comp
	h.Scheduler.Process(&proto.Evaluation{})
	h.Deployment = h.ApplyDep()
	return h.Deployment
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
	spec.Sequence = comp.Sequence // XXXXXXXXXXXXXXXXX

	fmt.Println("-- com p")
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

		plan.NodeUpdate = append(plan.NodeUpdate, ii)
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

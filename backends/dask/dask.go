package dask

import (
	"github.com/teseraio/ensemble/operator"
	"github.com/teseraio/ensemble/operator/proto"
)

const (
	schedulerKey = "scheduler"
)

type backend struct {
}

// Factory is a factory method for the zookeeper backend
func Factory() operator.Handler {
	return &backend{}
}

func (b *backend) PostHook(*operator.HookCtx) error {
	return nil
}

func (b *backend) Ready(t *proto.Instance) bool {
	return true
}

func (b *backend) Initialize(n []*proto.Instance, target *proto.Instance) (*proto.NodeSpec, error) {

	if target.Group.Type == "scheduler" {
		// start as a dask-scheduler
		target.Spec.Cmd = []string{
			"dask-scheduler",
		}
	} else if target.Group.Type == "worker" {
		// start the workers
		// find master
		var schedTarget string
		for _, m := range n {
			if m.Group.Type == "scheduler" {
				schedTarget = m.FullName()
			}
		}
		target.Spec.Cmd = []string{
			"dask-worker",
			"tcp://" + schedTarget + ":8786",
		}
	}
	return nil, nil
}

// Client implements the Handler interface
func (b *backend) Client(node *proto.Instance) (interface{}, error) {
	panic("X")
}

/*
// EvaluatePlan implements the Handler interface
func (b *backend) EvaluatePlan(ctx *operator.PlanCtx) error {

	for _, plan := range ctx.Plan.Sets {
		if plan.Type == "scheduler" {
			if len(plan.AddNodes) != 1 {
				return fmt.Errorf("only one node expected")
			}

			scheduler := plan.AddNodes[0]
			scheduler.Set(schedulerKey, "true")
			scheduler.Spec.Cmd = []string{
				"dask-scheduler",
			}

		} else if plan.Type == "worker" {
			scheduler := ""
			for _, n := range ctx.Cluster.Nodes {
				if n.Get(schedulerKey) == "true" {
					scheduler = n.FullName()
				}
			}
			if scheduler == "" {
				return fmt.Errorf("scheduler not found")
			}
			for _, node := range plan.AddNodes {
				node.Spec.Cmd = []string{
					"dask-worker",
					"tcp://" + scheduler + ":8786",
				}
			}
		}
	}
	return nil
}
*/

// Spec implements the Handler interface
func (b *backend) Spec() *operator.Spec {
	return &operator.Spec{
		Name: "Dask",
		Nodetypes: map[string]operator.Nodetype{
			"scheduler": {
				Image:   "daskdev/dask",
				Version: "2.30.0",
				Volumes: []*operator.Volume{},
				Ports:   []*operator.Port{},
			},
			"worker": {
				Image:   "daskdev/dask",
				Version: "2.30.0",
				Volumes: []*operator.Volume{},
				Ports:   []*operator.Port{},
			},
		},
	}
}

/*
// Client implements the Handler interface
func (b *backend) Client(node *proto.Node) (interface{}, error) {
	return nil, nil
}
*/

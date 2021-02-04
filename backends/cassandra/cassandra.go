package cassandra

import (
	"github.com/teseraio/ensemble/operator"
	"github.com/teseraio/ensemble/operator/proto"
)

var (
	seedKey = "seed"
)

type backend struct {
}

// Factory returns a factory method for the zookeeper backend
func Factory() operator.Handler {
	return &backend{}
}

func (b *backend) PostHook(*operator.HookCtx) error {
	return nil
}

// EvaluatePlan implements the Handler interface
func (b *backend) EvaluatePlan(ctx *operator.PlanCtx) error {
	plan := ctx.Plan.Sets[0]

	if len(plan.AddNodes) != 0 {
		var seed *proto.Node
		for _, n := range ctx.Cluster.Nodes {
			if n.Get(seedKey) == "ok" {
				seed = n
			}
		}
		if seed == nil {
			seed = plan.AddNodes[0]
			// take the first node as the seed
			seed.Set(seedKey, "ok")
		}
		for _, n := range plan.AddNodes {
			if n.Get(seedKey) == "" {
				// is not the seed node
				n.Spec.AddEnv("CASSANDRA_SEEDS", seed.ID)
			}
		}
	}

	if plan.DelNodesNum != 0 {
		set := ctx.Plan.Sets[0]
		// pick the last elements
		for _, n := range ctx.Cluster.Nodes[:set.DelNodesNum] {
			set.DelNodes = append(set.DelNodes, n.ID)
		}
	}
	return nil
}

// Spec implements the Handler interface
func (b *backend) Spec() *operator.Spec {
	return &operator.Spec{
		Name: "Cassandra",
		Nodetypes: map[string]operator.Nodetype{
			"": {
				Image:   "cassandra",
				Version: "latest", // TODO
				Volumes: []*operator.Volume{},
				Ports:   []*operator.Port{},
			},
		},
		Resources: []operator.Resource{},
	}
}

// Client implements the Handler interface
func (b *backend) Client(node *proto.Node) (interface{}, error) {
	return nil, nil
}

/*
// Reconcile implements the Handler interface
func (b *backend) Reconcile(executor operator.Executor, e *proto.Cluster, node *proto.Node, ctx *proto.Context) error {
	switch node.State {
	case proto.Node_INITIALIZED:
		b.recocileNodeInitialized(executor, e, node)
	}
	return nil
}
*/

func (b *backend) recocileNodeInitialized(executor operator.Executor, e *proto.Cluster, node *proto.Node) error {
	if len(e.Nodes) != 0 {
		// node joining a cluster (there should be another which is the seed)
		node.Spec.AddEnv("CASSANDRA_SEEDS", e.Nodes[0].ID)
	} else {
		// its the seed node
	}
	return nil
}

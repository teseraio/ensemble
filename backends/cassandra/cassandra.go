package cassandra

import (
	"github.com/teseraio/ensemble/operator"
	"github.com/teseraio/ensemble/operator/proto"
)

type backend struct {
}

// Factory returns a factory method for the zookeeper backend
func Factory() operator.Handler {
	return &backend{}
}

// EvaluatePlan implements the Handler interface
func (b *backend) EvaluatePlan(ctx *proto.Context) error {
	if ctx.Plan.Sets[0].DelNodesNum != 0 {
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
			"": operator.Nodetype{
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

// Reconcile implements the Handler interface
func (b *backend) Reconcile(executor operator.Executor, e *proto.Cluster, node *proto.Node, ctx *proto.Context) error {
	switch node.State {
	case proto.Node_INITIALIZED:
		b.recocileNodeInitialized(executor, e, node)
	}
	return nil
}

func (b *backend) delNodes(plan *proto.Plan) error {

	return nil
}

func (b *backend) recocileNodeInitialized(executor operator.Executor, e *proto.Cluster, node *proto.Node) error {
	if len(e.Nodes) != 0 {
		// node joining a cluster (there should be another which is the seed)
		node.Spec.AddEnv("CASSANDRA_SEEDS", e.Nodes[0].ID)
	} else {
		// its the seed node
	}
	return nil
}

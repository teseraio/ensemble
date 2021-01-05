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
func (b *backend) EvaluatePlan(plan *proto.Plan) error {
	if plan.DelNodesNum != 0 {
		return b.delNodes(plan)
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
func (b *backend) Reconcile(executor operator.Executor, e *proto.Cluster, node *proto.Node, plan *proto.Plan) error {
	switch node.State {
	case proto.Node_INITIALIZED:
		b.recocileNodeInitialized(executor, e, node, plan)
	}
	return nil
}

func (b *backend) delNodes(plan *proto.Plan) error {
	// pick the last elements
	for _, n := range plan.Cluster.Nodes[:plan.DelNodesNum] {
		plan.DelNodes = append(plan.DelNodes, n.ID)
	}
	return nil
}

func (b *backend) recocileNodeInitialized(executor operator.Executor, e *proto.Cluster, node *proto.Node, plan *proto.Plan) error {
	if len(e.Nodes) != 0 {
		// node joining a cluster (there should be another which is the seed)
		node.Spec.AddEnv("CASSANDRA_SEEDS", e.Nodes[0].ID)
	} else {
		// its the seed node
	}
	return nil
}

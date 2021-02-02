package dask

import (
	"github.com/teseraio/ensemble/operator"
	"github.com/teseraio/ensemble/operator/proto"
)

type backend struct {
}

// Factory is a factory method for the zookeeper backend
func Factory() operator.Handler {
	return &backend{}
}

// Reconcile implements the Handler interface
func (b *backend) Reconcile(_ operator.Executor, e *proto.Cluster, node *proto.Node, plan *proto.Context) error {
	switch node.State {
	case proto.Node_INITIALIZED:

		if node.Nodetype == "scheduler" {
			node.Spec.Cmd = []string{
				"dask-scheduler",
			}
		} else if node.Nodetype == "worker" {
			// This is always executed after the scheduler
			node.Spec.Cmd = []string{
				"dask-worker",
				"tcp://" + e.Nodes[0].FullName() + ":8786",
			}
		}
	}
	return nil
}

// EvaluatePlan implements the Handler interface
func (b *backend) EvaluatePlan(plan *proto.Context) error {
	return nil
}

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

// Client implements the Handler interface
func (b *backend) Client(node *proto.Node) (interface{}, error) {
	return nil, nil
}

package spark

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
func (b *backend) Reconcile(_ operator.Executor, e *proto.Cluster, node *proto.Node, plan *proto.Plan) error {
	switch node.State {
	case proto.Node_INITIALIZED:

		//
		node.Spec.Cmd = []string{
			"dask-scheduler",
		}

		// worker
		/*
			node.Spec.Cmd = []string{
				"dast-worker",
				"tcp://scheduler:8786",
			}
		*/
	}
	return nil
}

// PRIORITY
// DIFFERENT NODE SETS

// EvaluatePlan implements the Handler interface
func (b *backend) EvaluatePlan(plan *proto.Plan) error {
	return nil
}

// Spec implements the Handler interface
func (b *backend) Spec() *operator.Spec {
	return &operator.Spec{
		Name: "Dask",
		Nodetypes: map[string]operator.Nodetype{
			"": {
				Image:   "daskdev/dask",
				Version: "2.30.0",
				Volumes: []*operator.Volume{},
				Ports:   []*operator.Port{},
			},
			"1": {
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

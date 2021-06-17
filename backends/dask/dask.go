package dask

import (
	"github.com/teseraio/ensemble/operator"
	"github.com/teseraio/ensemble/operator/proto"
)

type backend struct {
	*operator.BaseOperator
}

// Factory is a factory method for the zookeeper backend
func Factory() operator.Handler {
	b := &backend{}
	b.BaseOperator = &operator.BaseOperator{}
	b.BaseOperator.SetHandler(b)
	return b
}

func (b *backend) Hooks() []operator.Hook {
	return []operator.Hook{}
}

func (b *backend) Name() string {
	return "Dask"
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

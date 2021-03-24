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

func (b *backend) Ready(t *proto.Instance) bool {
	return true
}

func (b *backend) Initialize(n []*proto.Instance, target *proto.Instance) (*proto.NodeSpec, error) {
	return nil, nil
}

// Spec implements the Handler interface
func (b *backend) Spec() *operator.Spec {
	return &operator.Spec{
		Name: "Clickhouse",
		Nodetypes: map[string]operator.Nodetype{
			"": {
				Image:   "yandex/clickhouse-server",
				Version: "latest",
				Volumes: []*operator.Volume{},
				Ports:   []*operator.Port{},
			},
		},
		Resources: []operator.Resource{},
	}
}

// Client implements the Handler interface
func (b *backend) Client(node *proto.Instance) (interface{}, error) {
	return nil, nil
}

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

func (b *backend) Ready(t *proto.Instance) bool {
	return true
}

func (b *backend) Initialize(n []*proto.Instance, target *proto.Instance) (*proto.NodeSpec, error) {
	// check if there is any seed node on the set already
	var seedNode *proto.Instance
	for _, i := range n {
		if i.GetTrue(seedKey) {
			seedNode = i
		}
	}

	if seedNode == nil {
		// we are the seed node
		target.SetTrue(seedKey)
	} else {
		// connect to the seed node
		target.Spec.AddEnv("CASSANDRA_SEEDS", seedNode.FullName())
	}
	return nil, nil
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
func (b *backend) Client(node *proto.Instance) (interface{}, error) {
	return nil, nil
}

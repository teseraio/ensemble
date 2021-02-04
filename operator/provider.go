package operator

import (
	"github.com/teseraio/ensemble/operator/proto"
)

// NodeUpdate is an update from a node (TODO: move to proto)
type NodeUpdate struct {
	// id of the node that has failed
	ID        string
	ClusterID string
}

// Provider is the entity that holds the state of the infrastructure. Both
// for the computing resources and the general resources.
type Provider interface {
	// Setup setups the provider (Maybe do this on the factory)
	Setup() error

	// Start starts the provider
	Start() error

	// CreateResource creates the computational resource
	CreateResource(*proto.Node) (*proto.Node, error)

	// DeleteResource deletes the computational resource
	DeleteResource(*proto.Node) (*proto.Node, error)

	// WatchUpdates watches for updates from nodes
	WatchUpdates() chan *NodeUpdate

	// Exec executes a shell script
	Exec(handler string, path string, args ...string) error
}

// ProviderFactory is a factory method to create factories
type ProviderFactory func(map[string]interface{}) error

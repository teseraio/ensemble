package operator

import (
	"github.com/teseraio/ensemble/operator/proto"
)

// TODO: Split provider between state and resource methods

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

	// Exec executes a shell script
	Exec(handler string, path string, args ...string) error

	// Update the status of the node
	UpdateNodeStatus(*proto.Node) (*proto.Node, error)

	// LoadCluster loads a Cluster at the specific resourceVersion
	LoadCluster(name string) (*proto.Cluster, error)

	// GetTask receives updates for ensembles
	GetTask() (*proto.Task, error)

	// FinalizeTask is used to notify the provider that the given task is done
	FinalizeTask(uuid string) error
}

// ProviderFactory is a factory method to create factories
type ProviderFactory func(map[string]interface{}) error

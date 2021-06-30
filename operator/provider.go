package operator

import (
	"fmt"

	"github.com/teseraio/ensemble/operator/proto"
	"github.com/teseraio/ensemble/schema"
)

var (
	ErrInstanceAlreadyRunning  = fmt.Errorf("instance already running")
	ErrProviderNameAlreadyUsed = fmt.Errorf("name already used")
)

// Provider is the entity that holds the state of the infrastructure. Both
// for the computing resources and the general resources.
type Provider interface {
	// Setup setups the provider (Maybe do this on the factory)
	Setup() error

	// Start starts the provider
	Start() error

	// CreateResource creates the computational resource
	CreateResource(*proto.Instance) (*proto.Instance, error)

	// DeleteResource deletes the computational resource
	DeleteResource(*proto.Instance) (*proto.Instance, error)

	// WatchUpdates watches for updates from nodes
	WatchUpdates() chan *proto.InstanceUpdate

	// Exec executes a shell script
	Exec(handler string, path string, args ...string) (string, error)

	// Resources returns a struct that defines the node resources
	// that can be configured for this provider
	Resources() ProviderResources

	Name() string
}

type ProviderResources struct {
	Nodeset schema.Schema2
}

// ProviderFactory is a factory method to create factories
type ProviderFactory func(map[string]interface{}) error

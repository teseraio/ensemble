package state

import (
	"context"
	"fmt"

	"github.com/teseraio/ensemble/operator/proto"
)

// Factory is the method to initialize the state
type Factory func(map[string]interface{}) (State, error)

// State stores the state of the Ensemble server
type State interface {
	Apply(*proto.Component) (int64, error)

	// GetComponent(id string) (*proto.Component, error)
	GetComponent(namespace, id string, sequence int64) (*proto.Component, error)

	Finalize(id string) error
	GetPending(id string) (*proto.Component, error)
	GetTask(ctx context.Context) *proto.Component

	ListDeployments() ([]*proto.Deployment, error)
	UpdateDeployment(d *proto.Deployment) error
	UpsertNode(n *proto.Instance) error
	LoadDeployment(id string) (*proto.Deployment, error)
	LoadInstance(cluster, id string) (*proto.Instance, error)

	// Close closes the state
	Close() error
}

var (
	ErrClusterNotFound = fmt.Errorf("cluster not found")
)

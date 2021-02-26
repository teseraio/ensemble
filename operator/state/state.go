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
	// UpsertNode updates the node. CHANGE TO UPSERTINSTANCE
	// UpsertNode(*proto.Node) error

	// UpsertCluster upserts the cluster
	// UpsertCluster(*proto.Cluster) error
	// GetCluster(name string) (*proto.Cluster, error)

	// Apply changes to a resource
	Apply(*proto.Component) (int64, error)
	GetComponent(namespace, id string, generation int64) (*proto.Component, error)
	Finalize(id string) error
	GetPending(id string) (*proto.Component, error)
	GetTask(ctx context.Context) *proto.Component

	// Get returns a component
	// Get(name string) (*proto.Component, error)

	// GetTask returns a new task to apply
	// GetTask(ctx context.Context) (*proto.ComponentTask, error)

	// LoadCluster loads a cluster from memory
	// LoadCluster(id string) (*proto.Cluster, error)

	UpdateDeployment(d *proto.Deployment) error
	UpsertNode(n *proto.Instance) error
	LoadDeployment(id string) (*proto.Deployment, error)
	LoadInstance(cluster, id string) (*proto.Instance, error)

	// Close closes the state
	Close() error

	// Experimental, move to operator off memory
	//AddEvaluation(eval *proto.Evaluation) error
	//GetTask2(ctx context.Context) (*proto.Evaluation, error)
}

var (
	ErrClusterNotFound = fmt.Errorf("cluster not found")
)

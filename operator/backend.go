package operator

import "github.com/teseraio/ensemble/operator/proto"

// HandlerFactory is a factory for Handlers
type HandlerFactory func() Handler

type PlanCtx struct {
	Plan    *proto.Plan
	Cluster *proto.Cluster
}

// Handler is the interface that needs to be implemented by the backend
type Handler interface {
	// Reconcile is called whenever there is an internal state change in the cluster
	Reconcile(executor Executor, e *proto.Cluster, node *proto.Node, plan *proto.Plan_Set) error

	// EvaluatePlan evaluates and modifies the execution plan
	EvaluatePlan(plan *PlanCtx) error

	// Spec returns the specification for the cluster
	Spec() *Spec

	// Client returns a connection with a specific node in the cluster
	Client(node *proto.Node) (interface{}, error)
}

// Executor is the interface required by the backends to execute
type Executor interface {
	Exec(n *proto.Node, path string, cmd ...string) error
}

// Spec returns the backend specification
type Spec struct {
	Name      string
	Nodetypes map[string]Nodetype
	Resources []Resource
}

// Nodetype is a type of node for the Backend
type Nodetype struct {
	// Image is the default docker image for the node
	Image string

	// TODO
	// Config interface{}

	// Version is the default docker image for the node
	Version string

	// Volume is a list volumes for this node type
	Volumes []*Volume

	// Ports is a list of ports for this node type
	Ports []*Port
}

// Port is an exposed port for the node
type Port struct {
	Name        string
	Port        uint64
	Description string
}

// Volume is a mounted path for the node
type Volume struct {
	Name        string
	Path        string
	Description string
}

// Resource is a resource in the cluster
type Resource interface {
	GetName() string
	Delete(conn interface{}) error
	Reconcile(conn interface{}) error
	Init(spec map[string]interface{}) error
}

// BaseResource is a resource that can have multiple instances
type BaseResource struct {
	ID string `schema:"id"`
}

// SetID sets the id of the specific resource
func (b *BaseResource) SetID(id string) {
	b.ID = id
}

// Init implements the Resource interface
func (b *BaseResource) Init(spec map[string]interface{}) error {
	return nil
}

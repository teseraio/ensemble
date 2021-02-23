package operator

import "github.com/teseraio/ensemble/operator/proto"

// HandlerFactory is a factory for Handlers
type HandlerFactory func() Handler

type HookCtx struct {
	Cluster  *proto.Cluster
	Node     *proto.Instance
	Executor Executor
}

type NodeRes struct {
	Config interface{}
}

type PlanCtx struct {
	Cluster *proto.Cluster
	Plan    *proto.Plan_Step
}

type BaseHandler struct {
}

func (b *BaseHandler) PostHook(*HookCtx) error {
	return nil
}

// Handler is the interface that needs to be implemented by the backend
type Handler interface {
	// EvaluatePlan evaluates and modifies the execution plan
	EvaluatePlan(n []*proto.Instance) error

	EvaluateConfig(spec *proto.NodeSpec, config map[string]string) error

	Initialize(grp *proto.Group, n []*proto.Instance, target *proto.Instance) (*proto.NodeSpec, error)
	// A(clr *proto.Cluster, n []*proto.Node) error

	// PostHook is executed when a node changes the state
	// PostHook(*HookCtx) error

	// Spec returns the specification for the cluster
	Spec() *Spec

	// Client returns a connection with a specific node in the cluster
	Client(node *proto.Instance) (interface{}, error)
}

// Executor is the interface required by the backends to execute
type Executor interface {
	Exec(n *proto.Instance, path string, cmd ...string) error
}

// Spec returns the backend specification
type Spec struct {
	Name      string
	Nodetypes map[string]Nodetype
	Resources []Resource
	Handlers  map[string]func(spec *proto.NodeSpec, grp *proto.ClusterSpec2_Group)
}

// Nodetype is a type of node for the Backend
type Nodetype struct {
	// Image is the default docker image for the node
	Image string

	// Config is the configuration fields for this node type
	Config interface{}

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

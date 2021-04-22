package operator

import (
	"fmt"

	"github.com/teseraio/ensemble/operator/proto"
	"github.com/teseraio/ensemble/schema"
)

// HandlerFactory is a factory for Handlers
type HandlerFactory func() Handler

type Handler2 interface {
	Spec() *Spec
	Initialize(n []*proto.Instance, target *proto.Instance) (*proto.NodeSpec, error)
}

// Handler is the interface that needs to be implemented by the backend
type Handler interface {
	// Initialize(n []*proto.Instance, target *proto.Instance) (*proto.NodeSpec, error)
	Ready(t *proto.Instance) bool

	// Spec returns the specification for the cluster
	// Spec() *Spec

	// Name returns the name of the handler (TODO. it can be removed later)
	Name() string

	// Client returns a connection with a specific node in the cluster
	// Client(node *proto.Instance) (interface{}, error)

	// GetSchemas returns the schemas for the backend
	GetSchemas() GetSchemasResponse

	// ApplyNodes applies to the spec the changes required
	ApplyNodes(n []*proto.Instance, cluster []*proto.Instance) ([]*proto.Instance, error)
}

type GetSchemasResponse struct {
	Nodes     map[string]schema.Schema2
	Resources map[string]schema.Schema2
}

// Spec returns the backend specification
type Spec struct {
	Name      string // out
	Nodetypes map[string]Nodetype
	Resources []Resource
	Handlers  map[string]func(spec *proto.NodeSpec, grp *proto.ClusterSpec_Group, data *schema.ResourceData)
}

func (s *Spec) GetResource(name string) (res Resource) {
	for _, i := range s.Resources {
		if i.GetName() == name {
			res = i
		}
	}
	return
}

// Nodetype is a type of node for the Backend
type Nodetype struct {
	// Image is the default docker image for the node
	Image string

	// Config is the configuration fields for this node type
	Config interface{} // out

	Schema schema.Schema2

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

type BaseOperator struct {
	handler Handler2
}

func (b *BaseOperator) SetHandler(h Handler2) {
	b.handler = h
}

func (b *BaseOperator) GetSchemas() GetSchemasResponse {
	resp := GetSchemasResponse{
		Nodes:     map[string]schema.Schema2{},
		Resources: map[string]schema.Schema2{},
	}
	// build the node types
	for k, v := range b.handler.Spec().Nodetypes {
		resp.Nodes[k] = v.Schema
	}
	// build the resources
	return resp
}

func (b *BaseOperator) ApplyNodes(place []*proto.Instance, cluster []*proto.Instance) ([]*proto.Instance, error) {
	// initialite the instances with the group specs
	placeInstances := []*proto.Instance{}
	for _, ii := range place {
		ii = ii.Copy()
		grpSpec := b.handler.Spec().Nodetypes[ii.Group.Type]

		ii.Spec.Image = grpSpec.Image
		ii.Spec.Version = grpSpec.Version

		fmt.Println("-- grp params --")
		fmt.Println(ii.Group.Params)

		hh, ok := b.handler.Spec().Handlers[ii.Group.Type]
		if ok {
			hh(ii.Spec, ii.Group, schema.NewResourceData(&grpSpec.Schema, ii.Group.Params))
		}
		placeInstances = append(placeInstances, ii)
	}

	// add the place instances too
	cluster = append(cluster, placeInstances...)

	// initialize each node
	for _, i := range placeInstances {
		if _, err := b.handler.Initialize(cluster, i); err != nil {
			panic(err)
		}
	}
	return placeInstances, nil
}

// BaseResource is a resource that can have multiple instances
type BaseResource struct {
}

// Init implements the Resource interface
func (b *BaseResource) Init(spec map[string]interface{}) error {
	return nil
}

var ErrResourceNotFound = fmt.Errorf("resource not found")

type CallbackRequest struct {
	Client interface{}
}

func (c *CallbackRequest) Get(s string) interface{} {
	return nil
}

type Resource2 struct {
	Name     string
	Schema   *schema.Record
	DeleteFn func(req *CallbackRequest) error
	ApplyFn  func(req *CallbackRequest) error
}

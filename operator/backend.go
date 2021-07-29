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
	Client(node *proto.Instance) (interface{}, error)
}

type nullHandler struct {
}

func (n *nullHandler) Setup(cplane ControlPlane) {
}

func (n *nullHandler) Name() string {
	return ""
}

func (n *nullHandler) WatchUpdates() chan *proto.InstanceUpdate {
	return nil
}

func (n *nullHandler) Evaluate(comp *proto.Component) (*proto.Component, error) {
	return comp, nil
}

func (n *nullHandler) GetSchemas() GetSchemasResponse {
	return GetSchemasResponse{}
}

func (n *nullHandler) ApplyNodes(nn []*proto.Instance, cluster []*proto.Instance) ([]*proto.Instance, error) {
	return nn, nil
}

func (n *nullHandler) ApplyResource(req *ApplyResourceRequest) error {
	return nil
}

// Handler is the interface that needs to be implemented by the backend
type Handler interface {
	// Name returns the name of the handler (TODO. it can be removed later)
	Name() string

	// Setup starts the backend and passes the control plane reference
	Setup(cplane ControlPlane)

	// Evaluate evaluates a component schema
	Evaluate(comp *proto.Component) (*proto.Component, error)

	// GetSchemas returns the schemas for the backend
	GetSchemas() GetSchemasResponse

	// ApplyNodes applies to the spec the changes required
	ApplyNodes(n []*proto.Instance, cluster []*proto.Instance) ([]*proto.Instance, error)

	// ApplyResource applies a resource change
	ApplyResource(req *ApplyResourceRequest) error
}

const (
	ApplyResourceRequestDelete    = "delete"
	ApplyResourceRequestReconcile = "reconcile"
)

type ApplyResourceRequest struct {
	Deployment *proto.Deployment
	Resource   *proto.ResourceSpec
	Action     string
}

type GetSchemasResponse struct {
	Nodes     map[string]schema.Schema2
	Resources map[string]schema.Schema2
}

// Spec returns the backend specification
type Spec struct {
	Name      string // out
	Nodetypes map[string]Nodetype
	Resources []*Resource2
	Validate  func(comp *proto.Component) (*proto.Component, error)
	Handlers  map[string]func(spec *proto.NodeSpec, grp *proto.ClusterSpec_Group, data *schema.ResourceData)
	Startup   func(i *proto.Instance) error
}

// Nodetype is a type of node for the Backend
type Nodetype struct {
	// Image is the default docker image for the node
	Image string

	// Config is the configuration fields for this node type
	Config interface{} // out

	Schema schema.Schema2

	// Version is the default docker image for the node
	DefaultVersion string

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
	cplane  ControlPlane
	// ch      chan *proto.InstanceUpdate
}

func (b *BaseOperator) SetHandler(h Handler2) {
	b.handler = h
}

func (b *BaseOperator) Setup(cplane ControlPlane) {
	// b.handler.Setup2()
	b.cplane = cplane
	stream := cplane.SubscribeInstanceUpdates()

	go func() {
		for {
			msg := <-stream
			if err := b.handleMsg(msg); err != nil {
				fmt.Printf("[ERR]: failed to handle backend message: %s", err.Error())
			}
		}
	}()
}

func (b *BaseOperator) handleMsg(msg *InstanceUpdate) error {
	instance, err := b.cplane.GetInstance(msg.InstanceID)
	if err != nil {
		return err
	}
	if instance.Status == proto.Instance_RUNNING && !instance.Healthy {
		ii := instance.Copy()
		if err := b.handler.Spec().Startup(ii); err != nil {
			// move to a failing state directly, or maybe ist just to reuquee the msg
			return err
		}

		// the new instance has to update
		if !ii.Healthy {
			return fmt.Errorf("it should be healtyh after startup time")
		}
		b.cplane.UpsertInstance(ii)
	}
	return nil
}

func (b *BaseOperator) Evaluate(comp *proto.Component) (*proto.Component, error) {
	valFunc := b.handler.Spec().Validate
	if valFunc != nil {
		return valFunc(comp)
	}
	return comp, nil
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
	for _, res := range b.handler.Spec().Resources {
		resp.Resources[res.Name] = res.Schema
	}
	return resp
}

func (b *BaseOperator) ApplyNodes(place []*proto.Instance, cluster []*proto.Instance) ([]*proto.Instance, error) {
	// initialite the instances with the group specs
	placeInstances := []*proto.Instance{}
	for _, ii := range place {
		ii = ii.Copy()
		grpSpec := b.handler.Spec().Nodetypes[ii.Group.Type]

		version := ii.Group.Version
		if version == "" {
			version = grpSpec.DefaultVersion
		}
		ii.Image = grpSpec.Image
		ii.Version = version

		hh, ok := b.handler.Spec().Handlers[ii.Group.Type]
		if ok {
			params := ii.Group.Params
			if params == nil {
				params = schema.MapToSpec(map[string]interface{}{})
			}
			hh(ii.Spec, ii.Group, schema.NewResourceData(&grpSpec.Schema, params))
		}
		placeInstances = append(placeInstances, ii)
	}

	// add the place instances too
	cluster = append(cluster, placeInstances...)

	postHook := b.handler.Spec().Startup != nil

	// initialize each node
	for _, i := range placeInstances {
		if _, err := b.handler.Initialize(cluster, i); err != nil {
			return nil, err
		}
		if !postHook {
			// there is no posthook so the scheduler can move ahead without problem
			i.Healthy = true
		}
	}
	return placeInstances, nil
}

func (b *BaseOperator) Client(node *proto.Instance) (interface{}, error) {
	return nil, nil
}

func (b *BaseOperator) ApplyResource(req *ApplyResourceRequest) error {
	// get one of the clients
	clt, err := b.handler.Client(req.Deployment.Instances[0])
	if err != nil {
		return err
	}

	// get resource
	var resource *Resource2
	for _, res := range b.handler.Spec().Resources {
		if res.Name == req.Resource.Resource {
			resource = res
			break
		}
	}

	// build the request
	handlerReq := &CallbackRequest{
		Client: clt,
		Data:   schema.NewResourceData(&resource.Schema, req.Resource.Params),
	}
	if req.Action == ApplyResourceRequestReconcile {
		err = resource.ApplyFn(handlerReq)
	} else if req.Action == ApplyResourceRequestDelete {
		err = resource.DeleteFn(handlerReq)
	} else {
		return fmt.Errorf("action not found '%s'", req.Action)
	}
	if err != nil {
		return err
	}
	return nil
}

var ErrResourceNotFound = fmt.Errorf("resource not found")

type CallbackRequest struct {
	Client interface{}
	Data   *schema.ResourceData
}

func (c *CallbackRequest) Get(s string) interface{} {
	return c.Data.Get(s)
}

type Resource2 struct {
	Name     string
	Schema   schema.Schema2
	DeleteFn func(req *CallbackRequest) error
	ApplyFn  func(req *CallbackRequest) error
}

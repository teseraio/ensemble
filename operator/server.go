package operator

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"reflect"
	"strings"

	gproto "github.com/golang/protobuf/proto"

	"github.com/hashicorp/go-hclog"
	"github.com/teseraio/ensemble/operator/proto"
	"github.com/teseraio/ensemble/operator/state"
	"github.com/teseraio/ensemble/schema"
	"google.golang.org/grpc"
)

// Config is the parametrization of the operator server
type Config struct {
	// Provider is the provider used by the operator
	Provider Provider

	// State is the state access
	State state.State

	// Backends are the list of backends handled by the operator
	HandlerFactories []HandlerFactory

	// GRPCAddr is the address of the grpc server
	GRPCAddr *net.TCPAddr
}

// Server is the operator server
type Server struct {
	config *Config
	logger hclog.Logger

	Provider Provider
	State    state.State

	handlers   map[string]Handler
	grpcServer *grpc.Server
	stopCh     chan struct{}

	service proto.EnsembleServiceServer
}

// NewServer starts an instance of the operator server
func NewServer(logger hclog.Logger, config *Config) (*Server, error) {
	s := &Server{
		config:   config,
		logger:   logger,
		Provider: config.Provider,
		State:    config.State,
		stopCh:   make(chan struct{}),
		handlers: map[string]Handler{},
	}

	for _, factory := range config.HandlerFactories {
		handler := factory()
		s.handlers[strings.ToLower(handler.Spec().Name)] = handler
	}

	s.service = &service{s}

	s.grpcServer = grpc.NewServer()
	proto.RegisterEnsembleServiceServer(s.grpcServer, s.service)

	// grpc address
	if err := s.setupGRPCServer(s.config.GRPCAddr.String()); err != nil {
		return nil, err
	}

	s.logger.Info("Start provider")
	if err := s.Provider.Start(); err != nil {
		return nil, err
	}

	go s.taskQueue()
	return s, nil
}

func (s *Server) setupGRPCServer(addr string) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	go func() {
		if err := s.grpcServer.Serve(lis); err != nil {
			panic(err)
		}
	}()

	s.logger.Info("Server started", "addr", addr)
	return nil
}

func (s *Server) Config() *Config {
	return s.config
}

// Stop stops the server
func (s *Server) Stop() {
	s.grpcServer.Stop()
	close(s.stopCh)
}

func (s *Server) applyChange(handler Handler, c *proto.Cluster, n *proto.Node, newState proto.Node_NodeState, plan *proto.Plan) (*proto.Cluster, error) {
	indx := c.NodeAtIndex(n.ID)

BACK:
	nn := n.Copy()
	nn.State = newState

	var err error
	var nextState proto.Node_NodeState

	// call the reconcile function
	if err := handler.Reconcile(s, c, nn, plan); err != nil {
		return c, err
	}

	s.logger.Debug("Apply state change", "id", nn.ID, "state", nn.State.String())

	if nn.State == proto.Node_INITIALIZED {
		// include the default image and version
		typ, ok := handler.Spec().Nodetypes[nn.Nodetype]
		if !ok {
			return nil, fmt.Errorf("nodetype %s does not exist", nn.Nodetype)
		}

		if nn.Spec == nil {
			nn.Spec = &proto.Node_NodeSpec{}
		}
		nn.Spec.Image = typ.Image
		nn.Spec.Version = typ.Version

		// create
		if nn, err = s.Provider.CreateResource(nn); err != nil {
			return nil, err
		}
	} else if nn.State == proto.Node_TAINTED {
		// delete
		if nn, err = s.Provider.DeleteResource(nn); err != nil {
			return nil, err
		}
	}

	// state transitions
	switch nn.State {
	case proto.Node_INITIALIZED:
		nextState = proto.Node_PENDING

	case proto.Node_PENDING:
		nextState = proto.Node_RUNNING

	case proto.Node_RUNNING:
		// No transition

	case proto.Node_TAINTED:
		nextState = proto.Node_DOWN

	case proto.Node_DOWN:
		// No transition
	}

	// update the cluster
	cc := c.Copy()
	if indx == -1 {
		cc.Nodes = append(cc.Nodes, nn)
		indx = len(cc.Nodes) - 1
	} else {
		cc.Nodes[indx] = nn
	}

	// save the new node state
	if err = s.State.UpsertNode(nn); err != nil {
		return nil, err
	}

	if nextState != proto.Node_UNKNOWN {
		c = cc
		n = nn
		newState = nextState

		// call the fsm again if there is another transition
		goto BACK
	}

	return cc, nil
}

// Exec implements the Activator interface
func (s *Server) Exec(n *proto.Node, path string, cmd ...string) error {
	return s.Provider.Exec(n.Handle, path, cmd...)
}

func (s *Server) taskQueue() {
	s.logger.Info("Starting task worker")

	for {
		task, err := s.State.GetTask(context.Background())
		if err != nil {
			s.logger.Error("failed to get task", "err", err)
			continue
		}

		s.logger.Info("New task", "id", task.Id)
		if err := s.handleTask(task); err != nil {
			s.logger.Error("failed to handle task", "id", task.Id, "err", err)
		}

		s.logger.Info("Finalize task", "id", task.Id)
		s.State.Finalize(task.Id)
	}
}

func (s *Server) getHandler(name string) (Handler, bool) {
	h, ok := s.handlers[strings.ToLower(name)]
	return h, ok
}

func (s *Server) handleResourceTask(eval *proto.Component) error {
	var spec proto.ResourceSpec
	if err := gproto.Unmarshal(eval.Spec.Value, &spec); err != nil {
		return err
	}

	cluster, err := s.State.LoadCluster(spec.Cluster)
	if err != nil {
		return err
	}
	handler, ok := s.getHandler(cluster.Backend)
	if !ok {
		return fmt.Errorf("handler not found %s", cluster.Backend)
	}

	// take any of the nodes in the cluster to connect
	clt, err := handler.Client(cluster.Nodes[0])
	if err != nil {
		return err
	}

	var resource Resource
	for _, r := range handler.Spec().Resources {
		if r.GetName() == spec.Resource {
			resource = r
		}
	}
	if resource == nil {
		return fmt.Errorf("resource not found %s", spec.Resource)
	}

	var params map[string]interface{}
	if err := json.Unmarshal([]byte(spec.Params), &params); err != nil {
		return err
	}
	val := reflect.New(reflect.TypeOf(resource)).Elem().Interface()
	if err := schema.Decode(params, &val); err != nil {
		return err
	}
	resource = val.(Resource)

	if err := resource.Init(params); err != nil {
		return err
	}

	if eval.Action == proto.Component_DELETE {
		if err := resource.Delete(clt); err != nil {
			return err
		}
	} else {
		if err := resource.Reconcile(clt); err != nil {
			return err
		}
	}
	return nil
}

func (s *Server) handleClusterTask(eval *proto.Component) error {
	var spec proto.ClusterSpec
	if err := gproto.Unmarshal(eval.Spec.Value, &spec); err != nil {
		return err
	}

	cluster, err := s.State.LoadCluster(eval.Name)
	if err != nil {
		if err == state.ErrClusterNotFound {
			// bootstrap
			cluster = &proto.Cluster{
				Name:    eval.Name,
				Backend: spec.Backend,
				Nodes:   []*proto.Node{},
			}
			if err := s.State.UpsertCluster(cluster); err != nil {
				return err
			}
		} else {
			return err
		}
	}
	if cluster.Backend != spec.Backend {
		return fmt.Errorf("trying to use a different backend")
	}

	handler, ok := s.getHandler(spec.Backend)
	if !ok {
		return fmt.Errorf("handler not found %s", spec.Backend)
	}

	plan, err := evaluateCluster(eval, &spec, cluster, handler)
	if err != nil {
		return err
	}
	if plan == nil {
		// no more plans to apply for this cluster
		return nil
	}

	if plan.DelNodes != nil {
		if cluster, err = s.deleteNodes(handler, cluster, plan); err != nil {
			return err
		}
	}
	if plan.AddNodes != nil {
		if cluster, err = s.addNodes(handler, cluster, plan); err != nil {
			return err
		}
	}
	return nil
}

func (s *Server) handleTask(task *proto.Component) error {

	var err error
	if task.Spec.TypeUrl == "ensembleoss.io/proto.ClusterSpec" {
		err = s.handleClusterTask(task)
	} else if task.Spec.TypeUrl == "ensembleoss.io/proto.ResourceSpec" {
		err = s.handleResourceTask(task)
	} else {
		return fmt.Errorf("type url not found '%s'", task.Spec.TypeUrl)
	}
	if err != nil {
		return err
	}

	return nil
}

func (s *Server) deleteNode(handler Handler, e *proto.Cluster, n *proto.Node, plan *proto.Plan) (*proto.Cluster, error) {
	return s.applyChange(handler, e, n, proto.Node_TAINTED, plan)
}

func (s *Server) addNode(handler Handler, e *proto.Cluster, n *proto.Node, plan *proto.Plan) (*proto.Cluster, error) {
	return s.applyChange(handler, e, n, proto.Node_INITIALIZED, plan)
}

func clusterDiff(c *proto.Cluster, spec *proto.ClusterSpec, eval *proto.Component) (*proto.Plan, error) {
	oldNum := int64(len(c.Nodes))
	newNum := spec.Replicas

	// scale down
	if oldNum > newNum {
		return &proto.Plan{DelNodesNum: oldNum - newNum}, nil
	}

	// scale up
	if oldNum < newNum {
		plan := &proto.Plan{
			Bootstrap: oldNum == 0,
		}
		for i := int64(0); i < newNum-oldNum; i++ {
			plan.Add(c.NewNode())
		}
		return plan, nil
	}

	return nil, nil
}

func evaluateCluster(eval *proto.Component, spec *proto.ClusterSpec, c *proto.Cluster, handler Handler) (*proto.Plan, error) {
	plan, err := clusterDiff(c, spec, eval)
	if err != nil {
		return nil, err
	}
	if plan == nil {
		return nil, nil
	}

	// make a copy of the cluster
	plan.Cluster = c.Copy()

	// call the handler in case it wants to do something
	if err := handler.EvaluatePlan(plan); err != nil {
		return nil, err
	}
	return plan, nil
}

func (s *Server) deleteNodes(handler Handler, e *proto.Cluster, plan *proto.Plan) (*proto.Cluster, error) {
	s.logger.Info("Scale down", "num", len(plan.DelNodes))

	var err error
	for _, nodeID := range plan.DelNodes {
		indx := e.NodeAtIndex(nodeID)
		n := e.Nodes[indx]

		if e, err = s.deleteNode(handler, e, n, plan); err != nil {
			return nil, err
		}

		// delete the node from the cluster
		e.DelNodeAtIndx(indx)
	}
	return e, nil
}

func (s *Server) addNodes(handler Handler, e *proto.Cluster, plan *proto.Plan) (*proto.Cluster, error) {
	s.logger.Info("Scale up", "num", len(plan.AddNodes))

	// write the cluster now
	for _, n := range plan.AddNodes {
		ee, err := s.addNode(handler, e, n, plan)
		if err != nil {
			return ee, err
		}
		e = ee
	}
	return e, nil
}

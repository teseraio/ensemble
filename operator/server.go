package operator

import (
	"encoding/json"
	"fmt"
	"net"
	"reflect"
	"strings"

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
	config     *Config
	logger     hclog.Logger
	Provider   Provider
	handlers   map[string]Handler
	grpcServer *grpc.Server
	stopCh     chan struct{}
}

// NewServer starts an instance of the operator server
func NewServer(logger hclog.Logger, config *Config) (*Server, error) {
	s := &Server{
		logger:   logger,
		Provider: config.Provider,
		stopCh:   make(chan struct{}),
		handlers: map[string]Handler{},
	}

	for _, factory := range config.HandlerFactories {
		handler := factory()
		s.handlers[strings.ToLower(handler.Spec().Name)] = handler
	}

	s.logger.Info("Start provider")
	if err := s.Provider.Start(); err != nil {
		return nil, err
	}

	go s.taskQueue()
	return s, nil
}

func (s *Server) setupGRPCServer() error {
	lis, err := net.Listen("tcp", s.config.GRPCAddr.String())
	if err != nil {
		return err
	}

	s.grpcServer = grpc.NewServer()
	proto.RegisterEnsembleServiceServer(s.grpcServer, &service{s})

	go func() {
		if err := s.grpcServer.Serve(lis); err != nil {
			panic(err)
		}
	}()

	s.logger.Info("Server started", "addr", s.config.GRPCAddr.String())
	return nil
}

// Stop stops the server
func (s *Server) Stop() {
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
	if nn, err = s.Provider.UpdateNodeStatus(nn); err != nil {
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
	s.logger.Info("Starting task queue")

	for {
		task, err := s.Provider.GetTask()
		if err != nil {
			s.logger.Error("failed to get task", "err", err)
			continue
		}

		s.logger.Info("New task", "id", task.ID)
		if err := s.handleTask(task); err != nil {
			s.logger.Error("failed to handle task", "id", task.ID, "err", err)
		}

		s.logger.Info("Finalize task", "id", task.ID)
		s.Provider.FinalizeTask(task.ID)
	}
}

func (s *Server) getHandler(name string) (Handler, bool) {
	h, ok := s.handlers[name]
	return h, ok
}

func HandleResource(eval *proto.Evaluation, handler Handler, e *proto.Cluster) error {
	// take any of the nodes in the cluster to connect
	clt, err := handler.Client(e.Nodes[0])
	if err != nil {
		return err
	}

	var resource Resource
	for _, r := range handler.Spec().Resources {
		if r.GetName() == eval.Resource {
			resource = r
		}
	}
	if resource == nil {
		return fmt.Errorf("resource not found %s", eval.Resource)
	}

	val := reflect.New(reflect.TypeOf(resource)).Elem().Interface()
	if err := schema.DecodeString(eval.Spec, &val); err != nil {
		return err
	}

	resource = val.(Resource)
	if eval.State == proto.EvaluationState_DELETED {
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

func (s *Server) handleClusterTask(eval *proto.Evaluation, handler Handler, e *proto.Cluster) error {
	plan, err := evaluateCluster(eval, e, handler)
	if err != nil {
		return err
	}
	if plan == nil {
		// no more plans to apply for this cluster
		return nil
	}

	if plan.DelNodes != nil {
		if e, err = s.deleteNodes(handler, e, plan); err != nil {
			return err
		}
	}
	if plan.AddNodes != nil {
		if e, err = s.addNodes(handler, e, plan); err != nil {
			return err
		}
	}
	return nil
}

func (s *Server) handleTask(task *proto.Task) error {
	eval := task.Evaluation

	clusterName := eval.Cluster
	if clusterName == "" {
		clusterName = eval.Name
	}
	cluster, err := s.Provider.LoadCluster(clusterName)
	if err != nil {
		return err
	}

	handler, ok := s.getHandler(eval.Backend)
	if !ok {
		return fmt.Errorf("handler not found %s", eval.Backend)
	}

	// execute the evaluation by type
	if eval.Resource == "" {
		err = s.handleClusterTask(eval, handler, cluster)
	} else {
		err = HandleResource(eval, handler, cluster)
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

type clusterSpec struct {
	Replicas int64
}

func clusterDiff(c *proto.Cluster, eval *proto.Evaluation) (*proto.Plan, error) {
	var spec clusterSpec
	if err := json.Unmarshal([]byte(eval.Spec), &spec); err != nil {
		return nil, err
	}

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

func evaluateCluster(eval *proto.Evaluation, c *proto.Cluster, handler Handler) (*proto.Plan, error) {
	plan, err := clusterDiff(c, eval)
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

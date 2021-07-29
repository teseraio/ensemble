package operator

import (
	"context"
	"fmt"
	"net"
	"reflect"
	"strings"
	"sync"
	"time"

	gproto "github.com/golang/protobuf/proto"

	"github.com/hashicorp/go-hclog"
	"github.com/teseraio/ensemble/lib/uuid"
	"github.com/teseraio/ensemble/operator/proto"
	"github.com/teseraio/ensemble/operator/state/boltdb"
	"google.golang.org/grpc"
)

// Config is the parametrization of the operator server
type Config struct {
	// Provider is the provider used by the operator
	Provider Provider

	// State is the state access
	State *boltdb.BoltDB

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
	State    *boltdb.BoltDB

	handlers   map[string]Handler
	grpcServer *grpc.Server
	stopCh     chan struct{}

	evalQueue *evalQueue
	service   proto.EnsembleServiceServer

	// subscriptions
	lock sync.Mutex
	subs []chan *InstanceUpdate
}

// NewServer starts an instance of the operator server
func NewServer(logger hclog.Logger, config *Config) (*Server, error) {
	s := &Server{
		config:    config,
		logger:    logger,
		Provider:  config.Provider,
		State:     config.State,
		stopCh:    make(chan struct{}),
		handlers:  map[string]Handler{},
		evalQueue: newEvalQueue(),
		subs:      []chan *InstanceUpdate{},
	}

	for _, factory := range config.HandlerFactories {
		handler := factory()
		s.handlers[strings.ToLower(handler.Name())] = handler
	}

	s.service = &service{s: s}

	s.grpcServer = grpc.NewServer(s.withLoggingUnaryInterceptor())
	proto.RegisterEnsembleServiceServer(s.grpcServer, s.service)

	// grpc address
	if err := s.setupGRPCServer(s.config.GRPCAddr.String()); err != nil {
		return nil, err
	}

	// setup the node watcher in the provider
	s.Provider.Setup(s)

	// setup watcher for the different backends
	for _, i := range s.handlers {
		i.Setup(s)
	}

	s.logger.Info("Start provider")
	if err := s.Provider.Start(); err != nil {
		return nil, err
	}

	go s.taskQueue4()
	go s.taskQueue5()

	go s.instanceWatcher()

	return s, nil
}

func (s *Server) withLoggingUnaryInterceptor() grpc.ServerOption {
	return grpc.UnaryInterceptor(s.loggingServerInterceptor)
}

func (s *Server) loggingServerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()
	h, err := handler(ctx, req)
	s.logger.Trace("Request", "method", info.FullMethod, "duration", time.Since(start), "error", err)
	return h, err
}

func (s *Server) instanceWatcher() {
	//fmt.Println("INSTANCE WATCHER START")

	stream := s.SubscribeInstanceUpdates()
	for {
		msg := <-stream
		if err := s.handleInstanceUpdate(msg); err != nil {
			s.logger.Error("failed to handle instance update", "err", err)
		}
	}
}

func (s *Server) handleInstanceUpdate(msg *InstanceUpdate) error {
	instance, err := s.GetInstance(msg.Id, msg.Cluster)
	if err != nil {
		return err
	}
	if instance.Status == proto.Instance_RUNNING || instance.Status == proto.Instance_STOPPED {
		// add eval

		if instance.Status == proto.Instance_STOPPED && instance.DesiredStatus == proto.Instance_RUN {
			// update the deployment to running in case it is not already
			dep, err := s.LoadDeployment(instance.DeploymentID)
			if err != nil {
				return err
			}
			if dep.Status != proto.DeploymentRunning {
				dep = dep.Copy()
				dep.Status = proto.DeploymentRunning
				if err := s.State.UpdateDeployment(dep); err != nil {
					return err
				}
			}
		}

		eval := &proto.Evaluation{
			Id:           uuid.UUID(),
			Status:       proto.Evaluation_PENDING,
			TriggeredBy:  proto.Evaluation_NODECHANGE,
			DeploymentID: msg.Cluster, // this is the deployment id
			Type:         proto.EvaluationTypeCluster,
		}
		s.evalQueue.add(eval)
	}
	return nil
}

func (s *Server) setupGRPCServer(addr string) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	go func() {
		if err := s.grpcServer.Serve(lis); err != nil {
			s.logger.Error("failed to serve grpc server", "err", err)
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

// Exec implements the Activator interface
func (s *Server) Exec(n *proto.Instance, path string, cmd ...string) (string, error) {
	return s.Provider.Exec("n.Handle", path, cmd...)
}

func (s *Server) taskQueue5() {
	s.logger.Info("Starting spec change worker")

	for {
		task := s.State.GetTask(context.Background())

		// pre-load the component
		comp, err := s.State.GetComponentByID2(task.DeploymentID, task.ComponentID, task.Sequence)
		if err != nil {
			s.logger.Error("failed to get component", "deployment", task.DeploymentID, "component", task.ComponentID, "sequence", task.Sequence)
			continue
		}
		typ := comp.Type()

		dep, err := s.State.LoadDeployment(task.DeploymentID)
		if err != nil {
			s.logger.Error(err.Error())
			continue
		}

		switch typ {
		case proto.EvaluationTypeCluster:
			err = s.handleCluster(task, dep, comp)
		case proto.EvaluationTypeResource:
			err = s.handleResource(task, dep, task.DeploymentID, comp)
		}
		if err != nil {
			s.logger.Error("failed to handle scheduler", "err", err)
			continue
		}
	}
}

func (s *Server) handleResource(task *proto.Task, dep *proto.Deployment, clusterID string, comp *proto.Component) error {
	var spec proto.ResourceSpec
	if err := gproto.Unmarshal(comp.Spec.Value, &spec); err != nil {
		return err
	}
	if dep == nil {
		return fmt.Errorf("deployment does not exists y")
	}

	// create a specChange evaluation
	s.evalQueue.add(&proto.Evaluation{
		Id:           uuid.UUID(),
		Status:       proto.Evaluation_PENDING,
		TriggeredBy:  proto.Evaluation_SPECCHANGE,
		DeploymentID: clusterID,
		Type:         proto.EvaluationTypeResource,
		Sequence:     comp.Sequence,
		ComponentID:  task.ComponentID,
	})
	return nil
}

func (s *Server) handleCluster(task *proto.Task, dep *proto.Deployment, comp *proto.Component) error {
	var spec proto.ClusterSpec
	if err := gproto.Unmarshal(comp.Spec.Value, &spec); err != nil {
		return err
	}

	if dep.Backend == "" {
		// new deployment
		dep.Backend = spec.Backend

		// get the dns suffix if any from the provider
		if s.Provider.Name() == "Kubernetes" {
			dep.DnsSuffix = ".default.svc.cluster.local"
		}
	} else {
		if dep.Backend != spec.Backend {
			return fmt.Errorf("the backend is not the same")
		}
	}

	// update the deployment
	dep.Status = proto.DeploymentRunning
	dep.Sequence = comp.Sequence
	dep.CompId = task.ComponentID

	if err := s.State.UpdateDeployment(dep); err != nil {
		return err
	}

	// create a specChange evaluation
	s.evalQueue.add(&proto.Evaluation{
		Id:           uuid.UUID(),
		Status:       proto.Evaluation_PENDING,
		TriggeredBy:  proto.Evaluation_SPECCHANGE,
		DeploymentID: task.DeploymentID,
		Type:         proto.EvaluationTypeCluster,
		Sequence:     comp.Sequence,
		ComponentID:  task.ComponentID,
	})
	return nil
}

func (s *Server) GetComponentByID(deployment, compID string, sequence int64) (*proto.Component, error) {
	return s.State.GetComponentByID2(deployment, compID, sequence)
}

func (s *Server) LoadDeployment(id string) (*proto.Deployment, error) {
	return s.State.LoadDeployment(id)
}

func (s *Server) GetHandler(id string) (Handler, error) {
	h, ok := s.handlers[strings.ToLower(id)]
	if !ok {
		return nil, fmt.Errorf("handler not found")
	}
	return h, nil
}

func (s *Server) newScheduler(typ string) Scheduler {
	if typ == proto.EvaluationTypeResource {
		return &ResourceScheduler{state: s}
	} else if typ == proto.EvaluationTypeCluster {
		return &scheduler{state: s}
	}
	panic("not found")
}

func (s *Server) taskQueue4() {
	s.logger.Info("Starting evaluation worker")

	for {
		eval := s.evalQueue.pop(context.Background())
		if eval == nil {
			return
		}

		s.logger.Debug("handle eval", "type", eval.Type, "id", eval.Id, "cluster", eval.DeploymentID, "trigger", eval.TriggeredBy.String())

		sched := s.newScheduler(eval.Type)
		plan, err := sched.Process(eval)
		if err != nil {
			s.logger.Error("failed to process", "err", err)
		} else {
			if err := s.SubmitPlan(eval, plan); err != nil {
				s.logger.Error("cannot submit plan", "err", err)
			}
		}

		s.logger.Trace("finalize eval", "id", eval.Id)
		s.evalQueue.finalize(eval.Id)
	}
}

func (s *Server) SubmitPlan(eval *proto.Evaluation, p *proto.Plan) error {
	// update the state
	for _, i := range p.NodeUpdate {
		if err := s.UpsertInstance(i); err != nil {
			return err
		}
	}

	if p.Deployment != nil {
		dep := p.Deployment.Copy()
		dep.Status = p.Status

		// update the state of the deployment if there is any change
		if len(p.NodeUpdate) != 0 || p.Deployment.Status != p.Status {
			if err := s.State.UpdateDeployment(dep); err != nil {
				return err
			}
		}
	}

	// if its done, finalize the component
	if p.Done {
		s.logger.Info("finalize task", "cluster", eval.DeploymentID)
		if err := s.State.Finalize(eval.DeploymentID); err != nil {
			return err
		}
	}
	return nil
}

func (s *Server) validateComponent(component *proto.Component) (*proto.Component, error) {
	msg, err := proto.UnmarshalAny(component.Spec)
	if err != nil {
		return nil, err
	}

	var handler Handler

	switch obj := msg.(type) {
	case *proto.ClusterSpec:
		handler, err = s.GetHandler(obj.Backend)
		if err != nil {
			return nil, err
		}
		if len(obj.Groups) == 0 {
			return nil, fmt.Errorf("no groups found")
		}
		for indx, grp := range obj.Groups {
			if grp.Count == 0 {
				return nil, fmt.Errorf("count 0 for group %d", indx)
			}
		}
	case *proto.ResourceSpec:
		// make sure the deployment exists
		depID, err := s.State.NameToDeployment(obj.Cluster)
		if err != nil {
			return nil, err
		}
		dep, err := s.LoadDeployment(depID)
		if err != nil {
			return nil, err
		}
		if dep == nil {
			return nil, fmt.Errorf("deployment does not exists '%s'", depID)
		}
		handler, err := s.GetHandler(dep.Backend)
		if err != nil {
			return nil, err
		}
		schemas := handler.GetSchemas()
		resource, ok := schemas.Resources[obj.Resource]
		if !ok {
			return nil, fmt.Errorf("resource %s does not exists", obj.Resource)
		}
		if err := resource.Validate(obj.Params); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("cannot validate spec: %s", reflect.TypeOf(msg))
	}

	component, err = handler.Evaluate(component)
	if err != nil {
		return nil, err
	}
	return component, nil
}

func (s *Server) UpsertInstance(n *proto.Instance) error {
	fmt.Printf("Upsert instance %s %s\n", n.ID, n.Status)

	s.lock.Lock()
	defer s.lock.Unlock()

	if err := s.State.UpsertNode(n); err != nil {
		return err
	}
	update := &InstanceUpdate{
		Id:      n.ID,
		Cluster: n.DeploymentID,
	}

	for _, ch := range s.subs {
		select {
		case ch <- update:
		default:
		}
	}
	return nil
}

func (s *Server) GetInstance(id, cluster string) (*proto.Instance, error) {
	return s.State.LoadInstance(cluster, id)
}

func (s *Server) SubscribeInstanceUpdates() <-chan *InstanceUpdate {
	fmt.Println("=====>")

	s.lock.Lock()
	defer s.lock.Unlock()

	if s.subs == nil {
		s.subs = []chan *InstanceUpdate{}
	}
	ch := make(chan *InstanceUpdate, 10)
	s.subs = append(s.subs, ch)

	return ch
}

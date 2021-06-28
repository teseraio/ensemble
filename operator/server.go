package operator

import (
	"context"
	"fmt"
	"net"
	"reflect"
	"strings"
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
	go s.startWatcher(s.Provider)

	// setup watcher for the different backends
	for _, i := range s.handlers {
		go s.startWatcher(i)
	}

	s.logger.Info("Start provider")
	if err := s.Provider.Start(); err != nil {
		return nil, err
	}

	go s.taskQueue4()
	go s.taskQueue5()

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

func (s *Server) upsertNode(i *proto.Instance) error {
	i = i.Copy()
	if err := s.State.UpsertNode(i); err != nil {
		return err
	}

	depID, err := s.State.NameToDeployment(i.ClusterName)
	if err != nil {
		return err
	}
	// TODO: Aggregate this in the same function or provide it from
	// the caller function as a context
	dep, err := s.LoadDeployment(depID)
	if err != nil {
		return err
	}
	handler, err := s.GetHandler(dep.Backend)
	if err != nil {
		return err
	}

	handler.ApplyHook(ApplyHookRequest{Instance: i, Deployment: dep})
	return nil
}

func (s *Server) upsertNodeAndEval(deploymentID string, i *proto.Instance) error {
	if err := s.upsertNode(i); err != nil {
		return err
	}
	eval := &proto.Evaluation{
		Id:           uuid.UUID(),
		Status:       proto.Evaluation_PENDING,
		TriggeredBy:  proto.Evaluation_NODECHANGE,
		DeploymentID: deploymentID,
		Type:         proto.EvaluationTypeCluster,
	}
	s.evalQueue.add(eval)
	return nil
}

func (s *Server) updateStatus(op *proto.InstanceUpdate) error {
	s.logger.Debug("update instance status", "id", op.ID, "cluster", op.ClusterName, "op", op)

	deploymentID, err := s.State.NameToDeployment(op.ClusterName)
	if err != nil {
		return err
	}
	i, err := s.State.LoadInstance(deploymentID, op.ID)
	if err != nil {
		return err
	}

	i = i.Copy()

	switch obj := op.Event.(type) {
	case *proto.InstanceUpdate_Running_:
		i.Ip = obj.Running.Ip
		i.Handler = obj.Running.Handler
		i.Status = proto.Instance_RUNNING

	case *proto.InstanceUpdate_Healthy_:
		i.Healthy = true

	case *proto.InstanceUpdate_Killing_:
		if i.Status == proto.Instance_TAINTED {
			// expected to be down
			i.Status = proto.Instance_STOPPED // It is moved to out by reconciler
			// dont do evaluation now
			return s.upsertNodeAndEval(deploymentID, i)
		} else {
			// the node is not expected to fail
			i.Status = proto.Instance_FAILED
			return s.upsertNodeAndEval(deploymentID, i)
		}
	}

	// update in the db
	if err := s.upsertNodeAndEval(deploymentID, i); err != nil {
		return err
	}
	return nil
}

type Watcher interface {
	WatchUpdates() chan *proto.InstanceUpdate
}

func (s *Server) startWatcher(w Watcher) {
	watchCh := w.WatchUpdates()
	for {
		select {
		case op := <-watchCh:
			go func() {
				if err := s.updateStatus(op); err != nil {
					panic(err)
				}
			}()

		case <-s.stopCh:
			return
		}
	}
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

// Exec implements the Activator interface
func (s *Server) Exec(n *proto.Instance, path string, cmd ...string) (string, error) {
	return s.Provider.Exec("n.Handle", path, cmd...)
}

func (s *Server) taskQueue5() {
	s.logger.Info("Starting spec change worker")

	for {
		task := s.State.GetTask2(context.Background())

		// pre-load the component
		comp, err := s.State.GetComponentByID2(task.DeploymentID, task.ComponentID, task.Sequence)
		if err != nil {
			panic(err)
		}
		typ := comp.Type()

		dep, err := s.State.LoadDeployment(task.DeploymentID)
		if err != nil {
			panic(err)
		}
		switch typ {
		case proto.EvaluationTypeCluster:
			err = s.handleCluster(task, dep, comp)
		case proto.EvaluationTypeResource:
			err = s.handleResource(task, dep, task.DeploymentID, comp)
		}
		if err != nil {
			panic(err)
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
				panic(err)
			}
		}

		s.logger.Trace("finalize eval", "id", eval.Id)
		s.evalQueue.finalize(eval.Id)
	}
}

func (s *Server) SubmitPlan(eval *proto.Evaluation, p *proto.Plan) error {
	// update the state
	for _, i := range p.NodeUpdate {
		if err := s.upsertNode(i); err != nil {
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
		if err := s.State.Finalize2(eval.DeploymentID); err != nil {
			return err
		}
	}

	// send the instances for update
	for _, i := range p.NodeUpdate {
		// create the instance
		if i.Status == proto.Instance_OUT {
			// instance is out
		} else {
			// Provider updates concurrently
			go func(i *proto.Instance) {
				if i.Status == proto.Instance_TAINTED {
					if _, err := s.Provider.DeleteResource(i); err != nil {
						panic(err)
					}
				} else if i.Status == proto.Instance_PENDING {
					if _, err := s.Provider.CreateResource(i); err != nil {
						panic(err)
					}
				}
			}(i)
		}
	}

	return nil
}

func (s *Server) validateComponent(component *proto.Component) error {
	msg, err := proto.UnmarshalAny(component.Spec)
	if err != nil {
		return err
	}
	switch obj := msg.(type) {
	case *proto.ClusterSpec:
		if len(obj.Groups) == 0 {
			return fmt.Errorf("no groups found")
		}
		for indx, grp := range obj.Groups {
			if grp.Count == 0 {
				return fmt.Errorf("count 0 for group %d", indx)
			}
			if grp.Storage == nil {
				grp.Storage = proto.EmptySpec()
			}
			if grp.Resources == nil {
				grp.Resources = proto.EmptySpec()
			}
		}
	case *proto.ResourceSpec:
		// make sure the deployment exists
		depID, err := s.State.NameToDeployment(obj.Cluster)
		if err != nil {
			return err
		}
		dep, err := s.LoadDeployment(depID)
		if err != nil {
			return err
		}
		if dep == nil {
			return fmt.Errorf("deployment does not exists %s", depID)
		}
		handler, err := s.GetHandler(dep.Backend)
		if err != nil {
			return err
		}
		schemas := handler.GetSchemas()
		resource, ok := schemas.Resources[obj.Resource]
		if !ok {
			return fmt.Errorf("resource %s does not exists", obj.Resource)
		}
		if err := resource.Validate(obj.Params); err != nil {
			return err
		}
	default:
		return fmt.Errorf("cannot validate spec: %s", reflect.TypeOf(msg))
	}
	return nil
}

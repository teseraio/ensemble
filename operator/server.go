package operator

import (
	"context"
	"fmt"
	"net"
	"strings"

	gproto "github.com/golang/protobuf/proto"

	"github.com/hashicorp/go-hclog"
	"github.com/teseraio/ensemble/lib/uuid"
	"github.com/teseraio/ensemble/operator/proto"
	"github.com/teseraio/ensemble/operator/state"
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

	s.grpcServer = grpc.NewServer()
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

	go s.taskQueue3()
	go s.taskQueue4()

	return s, nil
}

func (s *Server) upsertNode(i *proto.Instance) error {
	i = i.Copy()
	if err := s.State.UpsertNode(i); err != nil {
		return err
	}

	// TODO: Aggregate this in the same function or provide it from
	// the caller function as a context
	dep, err := s.LoadDeployment(i.Cluster)
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

func (s *Server) upsertNodeAndEval(i *proto.Instance) error {
	if err := s.upsertNode(i); err != nil {
		return err
	}
	eval := &proto.Evaluation{
		Id:          uuid.UUID(),
		Status:      proto.Evaluation_PENDING,
		TriggeredBy: proto.Evaluation_NODECHANGE,
		ClusterID:   i.Cluster,
	}
	s.evalQueue.add(eval)
	return nil
}

func (s *Server) updateStatus(op *proto.InstanceUpdate) error {
	s.logger.Debug("update instance status", "id", op.ID, "cluster", op.Cluster, "op", op)

	i, err := s.State.LoadInstance(op.Cluster, op.ID)
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
			return s.upsertNodeAndEval(i)
		} else {
			// the node is not expected to fail
			i.Status = proto.Instance_FAILED
			return s.upsertNodeAndEval(i)
		}
	}

	// update in the db
	if err := s.upsertNodeAndEval(i); err != nil {
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

func (s *Server) taskQueue3() {
	s.logger.Info("Starting spec change worker")

	for {
		comp := s.State.GetTask(context.Background())

		// get the name of the cluster and load the deployment
		clusterID, err := proto.ClusterIDFromComponent(comp)
		if err != nil {
			panic(err)
		}
		dep, err := s.State.LoadDeployment(clusterID)
		if err != nil {
			panic(err)
		}

		switch comp.Spec.TypeUrl {
		case "proto.ClusterSpec":
			err = s.handleCluster(dep, comp)

		case "proto.ResourceSpec":
			err = s.handleResource(dep, comp)
		}
		if err != nil {
			panic(err)
		}
	}
}

func (s *Server) handleResource(dep *proto.Deployment, comp *proto.Component) error {
	var spec proto.ResourceSpec
	if err := gproto.Unmarshal(comp.Spec.Value, &spec); err != nil {
		return err
	}
	if dep == nil {
		return fmt.Errorf("deployment does not exists")
	}

	handler, ok := s.getHandler(dep.Backend)
	if !ok {
		panic("bad")
	}
	dep, err := s.State.LoadDeployment(spec.Cluster)
	if err != nil {
		return err
	}

	schema, ok := handler.GetSchemas().Resources[spec.Resource]
	if !ok {
		return fmt.Errorf("resource not found %s", spec.Resource)
	}
	if err := schema.Validate(spec.Params); err != nil {
		panic(err)
	}

	if comp.Sequence != 1 {
		pastComp, err := s.State.GetComponent("proto-ResourceSpec", comp.Name, comp.Sequence-1)
		if err != nil {
			return err
		}
		var oldSpec proto.ResourceSpec
		if err := gproto.Unmarshal(pastComp.Spec.Value, &oldSpec); err != nil {
			return err
		}

		diff := schema.Diff(spec.Params, oldSpec.Params)

		// check if any of the diffs requires force-new
		forceNew := false
		for name := range diff {
			field, err := schema.Get(name)
			if err != nil {
				return err
			}
			if field.ForceNew {
				forceNew = true
			}
		}
		if forceNew {
			req := &ApplyResourceRequest{
				Deployment: dep,
				Action:     ApplyResourceRequestDelete,
				Resource:   &spec,
			}
			if err := handler.ApplyResource(req); err != nil {
				return err
			}
		}
	}

	var action string
	if comp.Action == proto.Component_DELETE {
		action = ApplyResourceRequestDelete
	} else {
		action = ApplyResourceRequestReconcile
	}
	req := &ApplyResourceRequest{
		Deployment: dep,
		Action:     action,
		Resource:   &spec,
	}
	handler.ApplyResource(req)

	if err := s.State.Finalize(dep.Name); err != nil {
		return err
	}
	return nil
}

func (s *Server) handleCluster(dep *proto.Deployment, comp *proto.Component) error {
	var spec proto.ClusterSpec
	if err := gproto.Unmarshal(comp.Spec.Value, &spec); err != nil {
		return err
	}

	if dep == nil {
		// new deployment
		dep = &proto.Deployment{
			Name:    comp.Name,
			Backend: spec.Backend,
		}
	} else {
		if dep.Backend != spec.Backend {
			return fmt.Errorf("the backend is not the same")
		}
	}

	// this is a change in the spec
	dep.Status = proto.DeploymentRunning
	dep.Sequence = comp.Sequence
	dep.CompId = comp.Id

	if err := s.State.UpdateDeployment(dep); err != nil {
		return err
	}

	// create a specChange evaluation
	s.evalQueue.add(&proto.Evaluation{
		Id:          uuid.UUID(),
		Status:      proto.Evaluation_PENDING,
		TriggeredBy: proto.Evaluation_SPECCHANGE,
		ClusterID:   dep.Name,
	})
	return nil
}

func (s *Server) GetComponentByID(deployment, compID string) (*proto.Component, error) {
	comp, err := s.State.GetComponentByID("proto-ClusterSpec", deployment, compID)
	if err != nil {
		return nil, err
	}
	return comp, nil
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

func (s *Server) taskQueue4() {
	s.logger.Info("Starting evaluation worker")

	for {
		eval := s.evalQueue.pop(context.Background())
		if eval == nil {
			return
		}

		s.logger.Debug("handle eval", "id", eval.Id, "cluster", eval.ClusterID, "trigger", eval.TriggeredBy.String())

		sched := NewScheduler(s)
		plan, err := sched.Process(eval)
		if err != nil {
			s.logger.Error("failed to process", "err", err)
		} else {
			if err := s.SubmitPlan(plan); err != nil {
				s.logger.Error("failed to submit plan", "err", err)
			}
		}

		s.logger.Trace("finalize eval", "id", eval.Id)
		s.evalQueue.finalize(eval.Id)
	}
}

func (s *Server) SubmitPlan(p *proto.Plan) error {
	// update the state
	for _, i := range p.NodeUpdate {
		if err := s.upsertNode(i); err != nil {
			return err
		}
	}

	dep := p.Deployment.Copy()
	dep.Status = p.Status

	// update the state of the deployment if there is any change
	if len(p.NodeUpdate) != 0 || p.Deployment.Status != p.Status {
		if err := s.State.UpdateDeployment(dep); err != nil {
			return err
		}
	}

	// if its done, finalize the component
	if dep.Status == proto.DeploymentDone {
		if err := s.State.Finalize(dep.Name); err != nil {
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

func (s *Server) getHandler(name string) (Handler, bool) {
	h, ok := s.handlers[strings.ToLower(name)]
	return h, ok
}

package operator

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"reflect"
	"strings"
	"time"

	gproto "github.com/golang/protobuf/proto"
	"github.com/mitchellh/mapstructure"

	"github.com/hashicorp/go-hclog"
	"github.com/teseraio/ensemble/lib/uuid"
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

	evalQueue *evalQueue
	service   proto.EnsembleServiceServer
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
		// deployments: map[string]*deploymentWatcher{},
		evalQueue: newEvalQueue(),
	}

	for _, factory := range config.HandlerFactories {
		handler := factory()
		s.handlers[strings.ToLower(handler.Spec().Name)] = handler
	}

	s.service = &service{s: s}

	s.grpcServer = grpc.NewServer()
	proto.RegisterEnsembleServiceServer(s.grpcServer, s.service)

	// grpc address
	if err := s.setupGRPCServer(s.config.GRPCAddr.String()); err != nil {
		return nil, err
	}

	// setup the node watcher in the provider
	go s.setupWatcher()

	s.logger.Info("Start provider")
	if err := s.Provider.Start(); err != nil {
		return nil, err
	}

	go s.taskQueue3()
	go s.taskQueue4()

	return s, nil
}

func (s *Server) readiness(i *proto.Instance) {
	dep, err := s.State.LoadDeployment(i.Cluster)
	if err != nil {
		panic(err)
	}

	handler, ok := s.getHandler(dep.Backend)
	if !ok {
		panic("bad")
	}
	for {
		if handler.Ready(i) {
			break
		}
		time.Sleep(1 * time.Second)
	}

	i.Healthy = true
	s.upsertNodeAndEval(i)
}

func (s *Server) upsertNodeAndEval(i *proto.Instance) error {
	if err := s.State.UpsertNode(i); err != nil {
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
	s.logger.Debug("update instance status", "id", op.ID, "cluster", op.Cluster)

	i, err := s.State.LoadInstance(op.Cluster, op.ID)
	if err != nil {
		return err
	}

	switch obj := op.Event.(type) {
	case *proto.InstanceUpdate_Running_:
		i.Ip = obj.Running.Ip
		i.Handler = obj.Running.Handler
		i.Status = proto.Instance_RUNNING

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

	if i.Status == proto.Instance_RUNNING && !i.Healthy {
		// if it is not healthy yet run the readiness function
		go s.readiness(i.Copy())
	}
	return nil
}

func (s *Server) setupWatcher() {
	watchCh := s.Provider.WatchUpdates()
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
func (s *Server) Exec(n *proto.Instance, path string, cmd ...string) error {
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

	// take any of the nodes in the cluster to connect
	clt, err := handler.Client(dep.Instances[0])
	if err != nil {
		return err
	}

	resSpec := handler.Spec().GetResource(spec.Resource)
	if resSpec == nil {
		return fmt.Errorf("resource not found %s", spec.Resource)
	}
	newResource, params, err := decodeResource(resSpec, spec.Params)
	if err != nil {
		return err
	}
	if err := newResource.Init(params); err != nil {
		return err
	}

	if comp.Sequence != 1 {
		pastComp, err := s.State.GetComponent("proto.ResourceSpec", comp.Id, comp.Sequence-1)
		if err != nil {
			return err
		}
		var oldSpec proto.ResourceSpec
		if err := gproto.Unmarshal(pastComp.Spec.Value, &oldSpec); err != nil {
			return err
		}

		forceNew, err := isForceNew(resSpec, &spec, &oldSpec)
		if err != nil {
			return err
		}
		if forceNew {
			// delete object
			removeResource, _, err := decodeResource(resSpec, oldSpec.Params)
			if err != nil {
				return err
			}
			if err := removeResource.Delete(clt); err != nil {
				return err
			}
		}
	}

	if comp.Action == proto.Component_DELETE {
		if err := newResource.Delete(clt); err != nil {
			return err
		}
	} else {
		if err := newResource.Reconcile(clt); err != nil {
			return err
		}
	}

	if err := s.State.Finalize(comp.Id); err != nil {
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

func (s *Server) taskQueue4() {
	s.logger.Info("Starting evaluation worker")

	for {
		eval := s.evalQueue.pop(context.Background())
		if eval == nil {
			return
		}

		s.logger.Debug("handle eval", "id", eval.Id, "cluster", eval.ClusterID, "trigger", eval.TriggeredBy.String())

		// get the deployment
		dep, err := s.State.LoadDeployment(eval.ClusterID)
		if err != nil {
			panic(err)
		}

		handler, ok := s.getHandler(dep.Backend)
		if !ok {
			panic("bad")
		}

		// get the spec for the cluster
		comp, err := s.State.GetComponent("proto-ClusterSpec", dep.Name, dep.Sequence)
		if err != nil {
			panic(err)
		}
		var spec proto.ClusterSpec
		if err := gproto.Unmarshal(comp.Spec.Value, &spec); err != nil {
			panic(err)
		}

		// we need this here because is not set before in the spec
		spec.Name = eval.ClusterID
		spec.Sequence = dep.Sequence

		r := &reconciler{
			delete: comp.Action == proto.Component_DELETE,
			dep:    dep,
			spec:   &spec,
		}
		r.Compute()
		// r.res.print()

		updates := []*proto.Instance{}

		// out instances
		for _, i := range r.res.out {
			ii := i.Copy()

			ii.Status = proto.Instance_OUT
			updates = append(updates, ii)
		}

		// promote instances
		for _, i := range r.res.ready {
			ii := i.Copy()
			ii.Canary = false

			updates = append(updates, ii)
		}

		// stop instances
		for _, i := range r.res.stop {
			ii := i.instance.Copy()
			ii.Status = proto.Instance_TAINTED
			ii.Canary = i.update

			updates = append(updates, ii)
		}

		if len(r.res.place) != 0 {
			// create a cluster object to initialize the instances
			cluster := []*proto.Instance{}
			for _, i := range dep.Instances {
				cluster = append(cluster, i)
			}

			// initialite the instances with the group specs
			placeInstances := []*proto.Instance{}
			for _, i := range r.res.place {

				var name string
				if i.instance == nil {
					name = i.name
				} else {
					name = i.instance.Name
				}

				ii := &proto.Instance{}
				ii.ID = uuid.UUID()
				ii.Group = i.group
				ii.Spec = &proto.NodeSpec{}
				ii.Cluster = spec.Name
				ii.Name = name
				ii.Status = proto.Instance_PENDING
				ii.Canary = i.update

				grpSpec := handler.Spec().Nodetypes[ii.Group.Type]

				ii.Spec.Image = grpSpec.Image
				ii.Spec.Version = grpSpec.Version

				hh, ok := handler.Spec().Handlers[ii.Group.Type]
				if ok {
					hh(ii.Spec, ii.Group)
				}
				placeInstances = append(placeInstances, ii)
			}

			// add the place instances too
			cluster = append(cluster, placeInstances...)

			// initialize each node
			for _, i := range placeInstances {
				if _, err := handler.Initialize(cluster, i); err != nil {
					panic(err)
				}
				updates = append(updates, i)
			}
		}

		plan := &schedulerPlan{
			deployment: dep,
			updates:    updates,
		}
		if r.res.done {
			if dep.Status != proto.DeploymentDone {
				plan.status = proto.DeploymentDone
			}
		} else {
			plan.status = proto.DeploymentRunning
		}

		if err := s.submitPlan(plan); err != nil {
			panic(err)
		}

		s.logger.Debug("finalize eval", "id", eval.Id)
		s.evalQueue.finalize(eval.Id)
	}
}

type schedulerPlan struct {
	deployment *proto.Deployment
	updates    []*proto.Instance
	status     string
}

// TODO: Serialize both handleEval and clusterSpec change because the spec can change
// and handleEval might rewrite with a wrong component id.

func (s *Server) submitPlan(p *schedulerPlan) error {
	// update the state
	for _, i := range p.updates {
		if err := s.State.UpsertNode(i); err != nil {
			return err
		}
	}

	dep := p.deployment.Copy()
	dep.Status = p.status

	// update the state of the deployment if there is any change
	if len(p.updates) != 0 || p.deployment.Status != p.status {
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
	for _, i := range p.updates {
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

func isForceNew(r Resource, old, new *proto.ResourceSpec) (bool, error) {
	var oldParams map[string]interface{}
	if err := json.Unmarshal([]byte(old.Params), &oldParams); err != nil {
		return false, err
	}
	var newParams map[string]interface{}
	if err := json.Unmarshal([]byte(new.Params), &newParams); err != nil {
		return false, err
	}

	// determine which fields are correct
	forcedFields := schema.ReadByTag(r, "force-new")

	for _, field := range forcedFields {
		oldVal, _ := schema.GetKey(oldParams, field)
		newVal, _ := schema.GetKey(newParams, field)

		if !reflect.DeepEqual(oldVal, newVal) {
			return true, nil
		}
	}
	return false, nil
}

func getResourceInstance(resource Resource) Resource {
	val := reflect.New(reflect.TypeOf(resource)).Elem().Interface()
	schema.Decode(map[string]interface{}{}, &val) // this id done to create the pointer with a value
	return val.(Resource)
}

func decodeResource(resource Resource, rawParams string) (Resource, map[string]interface{}, error) {
	var params map[string]interface{}
	if err := json.Unmarshal([]byte(rawParams), &params); err != nil {
		return nil, nil, err
	}
	val := reflect.New(reflect.TypeOf(resource)).Elem().Interface()
	if err := schema.Decode(params, &val); err != nil {
		return nil, nil, err
	}
	resource = val.(Resource)
	return resource, params, nil
}

func validateResources(output interface{}, input map[string]string) error {
	val := reflect.New(reflect.TypeOf(output)).Elem().Interface()
	var md mapstructure.Metadata
	config := &mapstructure.DecoderConfig{
		Metadata:         &md,
		Result:           &val,
		WeaklyTypedInput: true,
	}
	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return err
	}
	if err := decoder.Decode(input); err != nil {
		return err
	}
	if len(md.Unused) != 0 {
		return fmt.Errorf("unused keys %s", strings.Join(md.Unused, ","))
	}
	return nil
}

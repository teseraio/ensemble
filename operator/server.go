package operator

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"reflect"
	"strings"
	"sync"
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

	service proto.EnsembleServiceServer

	lock        sync.Mutex
	deployments map[string]*deploymentWatcher
}

// NewServer starts an instance of the operator server
func NewServer(logger hclog.Logger, config *Config) (*Server, error) {
	s := &Server{
		config:      config,
		logger:      logger,
		Provider:    config.Provider,
		State:       config.State,
		stopCh:      make(chan struct{}),
		handlers:    map[string]Handler{},
		deployments: map[string]*deploymentWatcher{},
		evalQueue:   newEvalQueue(),
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

type deploymentWatcher struct {
	s *Server

	// list of instances
	lock sync.Mutex
	ii   map[string]*proto.Instance
}

func (d *deploymentWatcher) upsertNodeAndEval(i *proto.Instance) {
	if err := d.s.State.UpsertNode(i); err != nil {
		panic(err)
	}
	eval := &proto.Evaluation{
		Id:          uuid.UUID(),
		Status:      proto.Evaluation_PENDING,
		TriggeredBy: proto.Evaluation_NODECHANGE,
		ClusterID:   i.Cluster,
	}
	d.s.evalQueue.add(eval)
}

func (d *deploymentWatcher) readiness(i *proto.Instance) {
	dep, err := d.s.State.LoadDeployment(i.Cluster)
	if err != nil {
		panic(err)
	}

	handler, ok := d.s.getHandler(dep.Backend)
	if !ok {
		panic("bad")
	}

	c := 0
	for {
		fmt.Printf("Ready: %s %d\n", i.FullName(), c)
		if handler.Ready(i) {
			break
		}
		time.Sleep(1 * time.Second)
	}

	fmt.Println("_ HEALTHY DONE _")
	i.Healthy = true
	d.upsertNodeAndEval(i)
}

func (d *deploymentWatcher) updateStatus(op *proto.InstanceUpdate) {
	d.s.logger.Debug("update instance status", "id", op.ID, "cluster", op.Cluster)

	i, ok := d.ii[op.ID]
	if !ok {
		panic("bad")
	}

	switch obj := op.Event.(type) {
	case *proto.InstanceUpdate_Conf:
		fmt.Println("_ A _")
		i.Ip = obj.Conf.Ip
		i.Handler = obj.Conf.Handler
		i.Status = proto.Instance_RUNNING

	case *proto.InstanceUpdate_Status:

		fmt.Println("-- i --")
		fmt.Println(i)
		fmt.Println(i.Desired)

		if i.Desired == "DOWN" || i.Desired == "Stop" {
			// expected to be down
			i.Status = proto.Instance_OUT
			// dont do evaluation now
			d.upsertNodeAndEval(i)
			return
		}
	}

	// update in the db
	d.upsertNodeAndEval(i)

	if i.Status == proto.Instance_RUNNING {
		go d.readiness(i)
	}
}

func (d *deploymentWatcher) Update(instance *proto.Instance) {
	d.s.logger.Debug("update instance", "id", instance.ID, "name", instance.Name, "cluster", instance.Cluster)

	d.lock.Lock()
	defer d.lock.Unlock()

	fmt.Printf("-- update instance from spec %s %s --\n", instance.ID, instance.Name)
	//fmt.Println(instance)
	//fmt.Println(instance.Canary)

	// do stuff here
	if instance.Desired == "Stop" {
		fmt.Println("- stop -")
		fmt.Println(instance.ID)

		if _, err := d.s.Provider.DeleteResource(instance); err != nil {
			panic(err)
		}
		fmt.Println("- done stop -")
	} else {
		fmt.Println("- create resource -")
		if _, err := d.s.Provider.CreateResource(instance); err != nil {
			panic(err)
		}
	}

	d.ii[instance.ID] = instance
}

func (s *Server) getDeployment(name string) *deploymentWatcher {
	s.lock.Lock()
	defer s.lock.Unlock()

	if _, ok := s.deployments[name]; !ok {
		fmt.Println("_ CREATE DEPLOYMENT _")
		fmt.Println(name)

		s.deployments[name] = &deploymentWatcher{s: s, ii: map[string]*proto.Instance{}}
	}
	return s.deployments[name]
}

func (s *Server) setupWatcher() {
	watchCh := s.Provider.WatchUpdates()
	for {
		select {
		case op := <-watchCh:
			dep := s.getDeployment(op.Cluster)
			go dep.updateStatus(op)

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
		pastComp, err := s.State.GetComponent(comp.Spec.TypeUrl, comp.Name, comp.Sequence-1)
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

	fmt.Printf("Handle cluster: %s (%d)\n", comp.Id, comp.Sequence)

	if dep == nil {
		// new deployment
		dep = &proto.Deployment{
			Id:      comp.Name,
			Backend: spec.Backend,
		}
	} else {
		if dep.Backend != spec.Backend {
			return fmt.Errorf("the backend is not the same")
		}
	}

	// this is a change in the spec
	dep.Status = proto.DeploymentRunning
	dep.CompID = comp.Id
	dep.Sequence = comp.Sequence

	if err := s.State.UpdateDeployment(dep); err != nil {
		return err
	}

	// create a specChange evaluation
	s.evalQueue.add(&proto.Evaluation{
		Id:          uuid.UUID(),
		Status:      proto.Evaluation_PENDING,
		TriggeredBy: proto.Evaluation_SPECCHANGE,
		ClusterID:   dep.Id,
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

		fmt.Println("-- comp id --")
		fmt.Println(dep.CompID)

		// get the spec for the cluster
		comp, err := s.State.GetPending(dep.CompID)
		if err != nil {
			panic(err)
		}

		var spec proto.ClusterSpec
		if err := gproto.Unmarshal(comp.Spec.Value, &spec); err != nil {
			panic(err)
		}

		spec.Name = eval.ClusterID
		spec.Sequence = dep.Sequence

		fmt.Println("- dep -")
		fmt.Println(dep)
		fmt.Println("- spec -")
		fmt.Println(spec)

		r := &reconciler{
			delete: comp.Action == proto.Component_DELETE,
			dep:    dep,
			spec:   &spec,
		}
		r.Compute()
		r.print()

		fmt.Println("- res -")
		fmt.Println(r.res)

		if !r.delete {
			// this is the initializtion step

			nn := []*proto.Instance{}
			for _, ii := range dep.Instances {
				nn = append(nn, ii)
			}
			for _, i := range r.res {
				nn = append(nn, i.instance)

				instance := i.instance

				grpSpec := handler.Spec().Nodetypes[instance.Group.Type]
				instance.Spec.Image = grpSpec.Image
				instance.Spec.Version = grpSpec.Version

				hh, ok := handler.Spec().Handlers[instance.Group.Type]
				if ok {
					hh(instance.Spec, instance.Group)
				}
			}

			// reconcile the init nodes
			for _, i := range r.res {
				handler.Initialize(nn, i.instance)

				fmt.Println("-- spec --")
				fmt.Println(i.instance.Spec)
			}
		}

		// we need to add this values to the db
		for _, i := range r.res {
			if err := s.State.UpsertNode(i.instance); err != nil {
				panic(err)
			}
		}

		depW := s.getDeployment(eval.ClusterID)
		for _, i := range r.res {
			// create the instance
			go depW.Update(i.instance)
		}

		if r.done {
			fmt.Println("___ DONE ___")
			// Update the status of the component and finalize the component
			dep.Status = proto.DeploymentDone
			if err := s.State.UpdateDeployment(dep); err != nil {
				panic(err)
			}
			if err := s.State.Finalize(dep.CompID); err != nil {
				panic(err)
			}
		}

		s.logger.Debug("finalize eval", "id", eval.Id)
		s.evalQueue.finalize(eval.Id)
	}
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

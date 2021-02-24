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

	go s.taskQueue2()
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
	if err := d.s.State.AddEvaluation(eval); err != nil {
		panic(err)
	}
}

func (d *deploymentWatcher) readiness(i *proto.Instance) {
	handler, ok := d.s.getHandler("Rabbitmq")
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
	i, ok := d.ii[op.ID]
	if !ok {
		panic("bad")
	}

	switch obj := op.Event.(type) {
	case *proto.InstanceUpdate_Conf:
		i.Ip = obj.Conf.Ip
		i.Handler = obj.Conf.Handler
		i.Status = proto.Instance_RUNNING

	case *proto.InstanceUpdate_Status:

		fmt.Println("-- i --")
		fmt.Println(i)
		fmt.Println(i.Desired)

		if i.Desired == "DOWN" {
			// expected to be down
			i.Status = proto.Instance_OUT
			// dont do evaluation now
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
	d.lock.Lock()
	defer d.lock.Unlock()

	fmt.Printf("-- update instance from spec %s %s --\n", instance.ID, instance.Name)
	//fmt.Println(instance)
	//fmt.Println(instance.Canary)

	// do stuff here
	if _, err := d.s.Provider.CreateResource(instance); err != nil {
		panic(err)
	}

	d.ii[instance.ID] = instance
}

func (s *Server) getDeployment(name string) *deploymentWatcher {
	s.lock.Lock()
	defer s.lock.Unlock()

	if _, ok := s.deployments[name]; !ok {
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

func (s *Server) taskQueue2() {
	s.logger.Info("Starting task worker")

	for {
		eval, err := s.State.GetTask2(context.Background())
		if err != nil {
			panic(err)
		}

		// get the component
		new, _, err := s.State.GetComponent(eval.ClusterID, 0)
		if err != nil {
			panic(err)
		}

		fmt.Println("##### EVAL #####")

		//fmt.Println(new)
		//fmt.Println(handler)

		switch new.Spec.GetTypeUrl() {
		case "proto.ResourceSpec":
			fmt.Println("_XXXX_")
			var res proto.ResourceSpec
			if err := gproto.Unmarshal(new.Spec.Value, &res); err != nil {
				panic(err)
			}
			if err := s.handleResourceTask(&res); err != nil {
				panic(err)
			}
			continue
		}

		handler, ok := s.getHandler("Rabbitmq")
		if !ok {
			panic("bad")
		}

		var spec proto.ClusterSpec2
		if err := gproto.Unmarshal(new.Spec.Value, &spec); err != nil {
			panic(err)
		}
		spec.Name = eval.ClusterID
		spec.Sequence = new.Sequence

		// get the deployment
		dep, err := s.State.LoadDeployment(eval.ClusterID)
		if err != nil {
			panic(err)
		}

		fmt.Println("#############################")
		fmt.Println(dep)
		fmt.Println(spec)

		dep.Sequence = new.Sequence
		dep.CompID = new.Id

		r := &reconciler{
			dep:  dep,
			spec: &spec,
		}
		r.Compute()
		r.print()

		// Update the status of the deployment
		if r.done {
			fmt.Println("____ DONE ____")
			dep.Status = proto.DeploymentDone
			// notify status
		} else {
			fmt.Println("____ RUNNING ____")
			dep.Status = proto.DeploymentRunning
		}
		if err := s.State.UpdateDeployment(dep); err != nil {
			panic(err)
		}

		fmt.Println("-- reconcile res --")
		fmt.Println(r.res)
		for _, i := range r.res {
			fmt.Println(i.instance, i.status)
		}

		// for the reconcile
		nn := []*proto.Instance{}
		for _, ii := range dep.Instances {
			nn = append(nn, ii)
		}

		for _, i := range r.res {
			if i.status == "promote" {
				continue
			}

			nn = append(nn, i.instance)

			instance := i.instance

			// make the backend reconcile
			fmt.Println(handler)
			fmt.Println(instance.Group.Type)

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
			if i.status == "promote" {
				continue
			}

			handler.Initialize(nn, i.instance)
		}

		// we need to add this values to the db
		for _, i := range r.res {
			if i.status != "promote" {
				if err := s.State.UpsertNode(i.instance); err != nil {
					panic(err)
				}
			}
		}

		//
		depW := s.getDeployment(eval.ClusterID)
		for _, i := range r.res {
			if i.status == "promote" {
				continue
			}

			// create the instance
			go depW.Update(i.instance)
		}

		if r.done {
			if err := s.State.Finalize(dep.CompID); err != nil {
				panic(err)
			}
		}
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

func (s *Server) handleResourceTask(spec *proto.ResourceSpec) error {
	handler, ok := s.getHandler("Rabbitmq")
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

	resource := handler.Spec().GetResource(spec.Resource)
	if resource == nil {
		return fmt.Errorf("resource not found %s", spec.Resource)
	}

	/*
		// Check if we have to destroy the current object if a force-new field
		// has changed
		if !isFirst {
			forceNew, err := isForceNew(resource, &spec, &oldSpec)
			if err != nil {
				return err
			}
			if forceNew {
				// delete object
				removeResource, _, err := decodeResource(resource, oldSpec.Params)
				if err != nil {
					return err
				}
				if err := removeResource.Delete(clt); err != nil {
					return err
				}
			}
		}
	*/

	resource, params, err := decodeResource(resource, spec.Params)
	if err != nil {
		return err
	}
	if err := resource.Init(params); err != nil {
		return err
	}

	fmt.Println("- reconcile -")
	fmt.Println(resource)

	if err := resource.Reconcile(clt); err != nil {
		return err
	}

	/*
		// check current value for the resource
		if eval.Action == proto.Component_DELETE {
			if err := resource.Delete(clt); err != nil {
				return err
			}
		} else {
		}
	*/

	return nil
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

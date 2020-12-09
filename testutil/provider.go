package testutil

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/teseraio/ensemble/operator/proto"
	"github.com/teseraio/ensemble/schema"
)

func WithProxy(config *TestProviderConfig) {
	config.Proxy = true
}

type TestProviderConfig struct {
	Proxy bool
}

type TestProviderCallback func(config *TestProviderConfig)

func NewTestProvider(t *testing.T, backend string, callback TestProviderCallback) (*Provider, func()) {
	config := &TestProviderConfig{}
	if callback != nil {
		callback(config)
	}
	clt, err := NewClient()
	if err != nil {
		t.Fatal(err)
	}
	p := &Provider{
		t:       t,
		docker:  clt,
		config:  config,
		backend: backend,
		tasks:   map[string]*task{},
		taskCh:  make(chan *proto.Task, 2),
		state:   &state{},
	}
	closeFn := func() { p.Stop() }
	return p, closeFn
}

type task struct {
	*proto.Task
	waitCh chan struct{}
}

// Provider is a mock provider for the swarm
type Provider struct {
	t       *testing.T
	config  *TestProviderConfig
	docker  *Client
	backend string

	tasks  map[string]*task
	taskCh chan *proto.Task

	// for the state
	state *state
}

func (p *Provider) Setup() error {
	return nil
}

func (p *Provider) Start() error {
	return nil
}

func (p *Provider) Stop() {
	p.docker.Clean()
}

func (p *Provider) Resource() schema.Schema {
	return schema.Schema{}
}

func (p *Provider) DeleteResource(node *proto.Node) (*proto.Node, error) {
	if err := p.docker.Remove(node.Handle); err != nil {
		return nil, err
	}
	return p.UpdateNodeStatus(node)
}

func (p *Provider) UpdateNodeStatus(node *proto.Node) (*proto.Node, error) {
	res, err := node.Marshal()
	if err != nil {
		return nil, err
	}

	eval := &proto.Evaluation{
		Name: node.ID,
		Spec: string(res),
	}
	eval = p.state.addObj(nodeNamespace, eval)

	nn := node.Copy()
	nn.ResourceVersion = eval.ResourceVersion
	return nn, nil
}

func (p *Provider) CreateResource(node *proto.Node) (*proto.Node, error) {
	id, err := p.docker.Create(context.TODO(), node)
	if err != nil {
		return nil, err
	}

	ip := p.docker.GetIP(id)

	nn := node.Copy()
	nn.Addr = ip
	nn.Handle = id
	return nn, nil
}

func (p *Provider) Exec(handler string, path string, args ...string) error {
	execCmd := []string{path}
	execCmd = append(execCmd, args...)

	return p.docker.Exec(context.Background(), handler, execCmd)
}

func (p *Provider) Remove(node *proto.Node) {
	p.docker.Remove(node.Handle)
}

func (p *Provider) GetTask() (*proto.Task, error) {
	task := <-p.taskCh
	return task, nil
}

func (p *Provider) FinalizeTask(uuid string) error {
	task, ok := p.tasks[uuid]
	if !ok {
		return fmt.Errorf("task %s not found", uuid)
	}
	close(task.waitCh)
	return nil
}

type TestTask struct {
	Name     string
	Resource string
	Input    string
	Delete   bool
}

func (p *Provider) LoadCluster(id string) (*proto.Cluster, error) {
	cluster := p.state.findByID(id)
	if cluster == nil {
		return nil, fmt.Errorf("not found")
	}

	c := &proto.Cluster{
		Nodes: []*proto.Node{},
	}
	for _, obj := range p.state.findByNamespace(nodeNamespace) {
		nn := new(proto.Node)
		if err := nn.Unmarshal([]byte(obj.evaluation().Spec)); err != nil {
			return nil, err
		}
		if nn.State != proto.NodeState_TAINTED && nn.State != proto.NodeState_DOWN {
			c.Nodes = append(c.Nodes, nn)
		}
	}
	return c, nil
}

// Apply applies a change. If the resource does not exists its a Create op
// if the resource exists its an Update op
func (p *Provider) Apply(t *TestTask) string {

	eval := &proto.Evaluation{
		Name:     t.Name,
		Cluster:  "A", // fixed cluster
		Resource: t.Resource,
		Spec:     t.Input,
		Backend:  p.backend,
	}

	if t.Delete {
		eval.State = proto.EvaluationState_DELETED
	}

	eval = p.state.addObj(objsNamespace, eval)

	id := uuid.New().String()
	pTask := &proto.Task{
		ID:         id,
		Evaluation: eval,
	}

	p.taskCh <- pTask

	p.tasks[id] = &task{
		Task:   pTask,
		waitCh: make(chan struct{}),
	}
	return id
}

// WaitForTask waits for a task to finish
func (p *Provider) WaitForTask(id string) {
	task, ok := p.tasks[id]
	if !ok {
		panic(fmt.Sprintf("BUG: Task %s not found", id))
	}
	<-task.waitCh
}

func (p *Provider) GetTaskByID(id string) *proto.Task {
	return p.tasks[id].Task
}

const (
	nodeNamespace = "node"
	objsNamespace = "objs"
)

type object struct {
	id    string
	evals []*proto.Evaluation
}

func (o *object) evaluation() *proto.Evaluation {
	return o.evals[len(o.evals)-1]
}

type state struct {
	version int
	objs    map[string]*object
}

func (s *state) findByID(id string) *object {
	for _, obj := range s.objs {
		if obj.id == id {
			return obj
		}
	}
	return nil
}

func (s *state) findByNamespace(namespace string) (result []*object) {
	prefix := fmt.Sprintf("/%s/", namespace)
	for id, obj := range s.objs {
		if strings.HasPrefix(id, prefix) {
			result = append(result, obj)
		}
	}
	return
}

func (s *state) nextVersion() string {
	s.version++
	return strconv.Itoa(s.version)
}

func (s *state) addObj(namespace string, eval *proto.Evaluation) *proto.Evaluation {
	key := fmt.Sprintf("/%s/%s", namespace, eval.Name)
	if len(s.objs) == 0 {
		s.objs = map[string]*object{}
	}

	obj, ok := s.objs[key]
	if !ok {
		obj = &object{
			id:    eval.Name,
			evals: []*proto.Evaluation{},
		}
	}

	eval.Generation = int64(len(obj.evals))
	eval.ResourceVersion = s.nextVersion()

	if eval.State == proto.EvaluationState_DELETED {
		// add the Spec from the last execution
		if !ok {
			panic("BUG: Cannot delete a non created object")
		}
		eval.Spec = obj.evals[len(obj.evals)-1].Spec
	} else {
		if eval.Generation == 0 {
			eval.State = proto.EvaluationState_CREATED
		} else {
			eval.State = proto.EvaluationState_UPDATED
		}
	}

	obj.evals = append(obj.evals, eval)
	s.objs[key] = obj
	return eval
}

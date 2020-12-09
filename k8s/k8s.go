package k8s

import (
	"container/heap"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/google/uuid"
	"github.com/hashicorp/go-hclog"
	"github.com/mitchellh/mapstructure"
	"github.com/teseraio/ensemble/operator/proto"
)

// Provider is a Provider implementation for kubernetes.
type Provider struct {
	client    *KubeClient
	logger    hclog.Logger
	stopCh    chan struct{}
	taskQueue *taskQueue
}

// Stop stops the kubernetes provider
func (p *Provider) Stop() {
	close(p.stopCh)
}

func (p *Provider) Setup() error {
	validClusters := map[string]struct{}{}
	err := p.listFull("/apis/ensembleoss.io/v1/namespaces/{namespace}/clusters", func(i *Item) {
		validClusters[i.Metadata.Name] = struct{}{}
	})
	if err != nil {
		return err
	}

	existingClusters := map[string]struct{}{}
	err = p.listFull("/api/v1/namespaces/{namespace}/services?labelSelector=ensemble%3Dok", func(i *Item) {
		existingClusters[i.Metadata.Name] = struct{}{}
	})
	if err != nil {
		return err
	}

	// check if there is any existing cluster that was deleted
	for name := range existingClusters {
		if _, ok := validClusters[name]; !ok {
			p.logger.Debug("Purge cluster", "name", name)
			if err := p.purgeCluster(name); err != nil {
				return err
			}
		}
	}
	return nil
}

func (p *Provider) contextWithClose() context.Context {
	ctx, cancelFn := context.WithCancel(context.Background())
	go func() {
		<-p.stopCh
		cancelFn()
	}()
	return ctx
}

func (p *Provider) trackEnsembleObjects(url string) {
	p.logger.Debug("track ensemble resource", "endpoint", url)

	itemCh, _ := p.client.Track(p.contextWithClose(), transformURL(url))
	for item := range itemCh {
		// generate a task for the item and enqueue it
		t, err := buildTask(item)
		if err != nil {
			p.logger.Error("failed to build task", "err", err)
		} else if t != nil {
			p.taskQueue.add(t)
		}
	}
}

const (
	// clustersURL is the k8s url for the cluster objects
	clustersURL = "/apis/ensembleoss.io/v1/namespaces/{namespace}/clusters"

	// resourcesURL is the k8s url for the resource objects
	resourcesURL = "/apis/ensembleoss.io/v1/namespaces/{namespace}/resources"
)

// Start starts the kubernetes provider
func (p *Provider) Start() error {

	// watch clusters
	go p.trackEnsembleObjects(clustersURL)

	// watch resources
	go p.trackEnsembleObjects(resourcesURL)

	return nil
}

func completeEvalForResourceKind(spec map[string]interface{}, eval *proto.Evaluation) error {
	var specFormat struct {
		Backend  string
		Cluster  string
		Resource string
		Params   map[string]interface{}
	}
	if err := mapstructure.Decode(spec, &specFormat); err != nil {
		return err
	}
	eval.Cluster = specFormat.Cluster
	eval.Resource = specFormat.Resource
	eval.Backend = specFormat.Backend

	params, err := json.Marshal(specFormat.Params)
	if err != nil {
		return err
	}
	eval.Spec = string(params)
	return nil
}

func completeEvalForClusterKind(spec map[string]interface{}, eval *proto.Evaluation) error {
	var specFormat struct {
		Backend struct {
			Name string
		}
	}
	if err := mapstructure.Decode(spec, &specFormat); err != nil {
		return err
	}
	eval.Backend = specFormat.Backend.Name
	return nil
}

func buildTask(i *Item) (*task, error) {
	// decode the observed generation
	var stat struct {
		ObservedGeneration int `mapstructure:"observedGeneration"`
	}
	if i.Status != nil {
		if err := mapstructure.Decode(i.Status, &stat); err != nil {
			return nil, err
		}
	}
	// The observed generation is the last generation, we can skip this value
	if stat.ObservedGeneration == i.Metadata.Generation {
		return nil, nil
	}

	// we need to serialize the input (TODO: Serialize state)
	spec, err := json.Marshal(i.Spec)
	if err != nil {
		return nil, err
	}

	eval := &proto.Evaluation{
		Name:            i.Metadata.Name,
		Generation:      int64(i.Metadata.Generation),
		Spec:            string(spec),
		ResourceVersion: i.Metadata.ResourceVersion,
	}

	if i.Kind == "Cluster" {
		if err := completeEvalForClusterKind(i.Spec, eval); err != nil {
			return nil, err
		}
	} else if i.Kind == "Resource" {
		if err := completeEvalForResourceKind(i.Spec, eval); err != nil {
			return nil, err
		}
	}

	oTask := &proto.Task{
		ID:         uuid.New().String(),
		Evaluation: eval,
		Timestamp:  ptypes.TimestampNow(),
	}

	tt, err := oTask.Time()
	if err != nil {
		return nil, err
	}
	k8sTask := &task{
		APIEndpoint: i.Metadata.SelfLink,
		Task:        oTask,
		timestamp:   tt,
	}
	return k8sTask, nil
}

// FinalizeTask implements the Provider interface
func (p *Provider) FinalizeTask(uuid string) error {
	task, ok := p.taskQueue.get(uuid)
	if !ok {
		return fmt.Errorf("Task %s not found", uuid)
	}

	// we need the latest resource version to update the object
	var item *Item
	if _, err := p.get(strings.TrimSuffix(task.APIEndpoint, "/status"), &item); err != nil {
		return err
	}

	obj := map[string]interface{}{
		"Domain": "ensembleoss.io/v1",
		"Kind":   item.Kind,
		"Status": map[string]interface{}{
			"observedGeneration": task.Evaluation.Generation,
		},
		"Name":            task.Evaluation.Name,
		"ResourceVersion": item.Metadata.ResourceVersion,
	}
	req, err := RunTmpl2("generic", obj)
	if err != nil {
		return err
	}

	if _, _, err := p.put(task.APIEndpoint+"/status", req); err != nil {
		return err
	}
	p.taskQueue.finalize(uuid)
	return nil
}

var emptyDel = []byte("{}")

// DeleteResource implements the Provider interface
func (p *Provider) DeleteResource(node *proto.Node) (*proto.Node, error) {
	if err := p.delete("/api/v1/namespaces/{namespace}/pods/"+node.ID, emptyDel); err != nil {
		return nil, err
	}

	var res struct {
		Status struct {
			Phase string
			PodIP string
		}
	}
	// wait for the pod to be removed
	for {
		if _, err := p.get("/api/v1/namespaces/{namespace}/pods/"+node.ID, &res); err != nil {
			if err == errNotFound {
				break
			}
			return nil, err
		}
		p.logger.Debug("Shutting down", "id", node.ID, "status", res.Status.Phase)
		time.Sleep(1 * time.Second)
	}

	// upsert the node spec
	nn, err := p.UpdateNodeStatus(node)
	if err != nil {
		return nil, err
	}
	return nn, nil
}

func (p *Provider) Exec(handler string, path string, cmdArgs ...string) error {
	out, err := p.client.Exec(handler, "main-container", path, cmdArgs...)
	if err != nil {
		return err
	}
	p.logger.Trace("exec", "handler", handler, "out", string(out))
	return nil
}

func transformURL(rawURL string) string {
	url := rawURL
	url = strings.Replace(url, "{namespace}", "default", -1)
	return url
}

func (p *Provider) loadNode(id string) (*proto.Node, error) {
	var item *Item
	if _, err := p.get("/apis/ensembleoss.io/v1/namespaces/{namespace}/nodes/"+id, &item); err != nil {
		return nil, err
	}
	node, err := unmarshalNode(item)
	if err != nil {
		return nil, err
	}
	return node, nil
}

// LoadCluster implements the Provider interface
func (p *Provider) LoadCluster(name string) (*proto.Cluster, error) {
	c := &proto.Cluster{
		Name:  name,
		Nodes: []*proto.Node{},
	}

	// Query all the nodes that belong to this cluster
	var res struct {
		Items []*Item
	}
	if _, err := p.get("/apis/ensembleoss.io/v1/namespaces/{namespace}/nodes?labelSelector=ensemble%3D"+name, &res); err != nil {
		return nil, err
	}
	for _, i := range res.Items {
		nn, err := unmarshalNode(i)
		if err != nil {
			return nil, err
		}
		c.Nodes = append(c.Nodes, nn)
	}
	return c, nil
}

// UpdateNodeStatus updates the status of the node
func (p *Provider) UpdateNodeStatus(node *proto.Node) (*proto.Node, error) {
	if node.ResourceVersion == "" {
		return nil, fmt.Errorf("Resource version must be set")
	}

	nodeRaw, err := marshalNode(node)
	if err != nil {
		return nil, err
	}

	resp, _, err := p.put("/apis/ensembleoss.io/v1/namespaces/{namespace}/nodes/"+node.ID+"/status", nodeRaw)
	if err != nil {
		return nil, err
	}
	nn, _, err := unmarshalNodeBytes(resp)
	if err != nil {
		return nil, err
	}
	return nn, nil
}

func (p *Provider) upsertNodeSpec(node *proto.Node) (*proto.Node, error) {
	// we assume the object has not been created for now
	isUpdated := node.ResourceVersion != ""

	nodeRaw, err := marshalNode(node)
	if err != nil {
		return nil, err
	}

	var resp []byte
	if isUpdated {
		resp, _, err = p.put("/apis/ensembleoss.io/v1/namespaces/{namespace}/nodes/"+node.ID+"/status", nodeRaw)
	} else {
		resp, _, err = p.post("/apis/ensembleoss.io/v1/namespaces/{namespace}/nodes", nodeRaw)
	}
	if err != nil {
		return nil, err
	}
	n, _, err := unmarshalNodeBytes(resp)
	if err != nil {
		return nil, err
	}

	// We cannot use n directly because only a status change changes the
	// kv, so the return value lacks some values
	node.ResourceVersion = n.ResourceVersion

	nn, err := p.UpdateNodeStatus(node)
	if err != nil {
		return nil, err
	}
	return nn, nil
}

func (p *Provider) purgeCluster(name string) error {
	p.logger.Info("Remove cluster", "name", name)

	// get the list of nodes for the ensemble
	var res struct {
		Items []*Item
	}
	if _, err := p.get("/apis/ensembleoss.io/v1/namespaces/{namespace}/nodes?labelSelector=ensemble%3D"+name, &res); err != nil {
		p.logger.Debug("nodes not found", "name", name, "err", err)
	}

	for _, node := range res.Items {
		name := node.Metadata.Name

		// delete pod
		if err := p.delete("/api/v1/namespaces/{namespace}/pods/"+name, emptyDel); err != nil {
			p.logger.Debug("failed to delete pod", "name", name, "err", err)
		}
		// delete crd
		if err := p.delete("/apis/ensembleoss.io/v1/namespaces/{namespace}/nodes/"+name, emptyDel); err != nil {
			p.logger.Debug("failed to delete crd node", "name", name, "err", err)
		}
	}

	// delete headless service
	if err := p.delete("/api/v1/namespaces/{namespace}/services/"+name, emptyDel); err != nil {
		p.logger.Debug("failed to delete service", "name", name, "err", err)
	}
	return nil
}

// GetTask implements the Provider interface
func (p *Provider) GetTask() (*proto.Task, error) {
	task := p.taskQueue.pop(p.stopCh)
	return task.Task, nil
}

func (p *Provider) upsertConfigMap(name string, data map[string]string) error {
	// check if the config map exists
	var item *Item

	if _, err := p.get("/api/v1/namespaces/{namespace}/configmaps/"+name, &item); err != nil {
		if err != errNotFound {
			return err
		}
	}

	exists := item != nil

	if exists {
		// Check if both maps are equal
		var previousData map[string]string
		if err := mapstructure.Decode(item.Data, &previousData); err != nil {
			return err
		}
		if reflect.DeepEqual(data, previousData) {
			return nil
		}
	}

	parts := []string{}
	for k, v := range data {
		parts = append(parts, fmt.Sprintf("\"%s\": \"%s\"", cleanPath(k), v))
	}
	obj := map[string]interface{}{
		"Name": name,
		"Data": strings.Join(parts, ","),
	}
	if exists {
		obj["ResourceVersion"] = item.Metadata.ResourceVersion
	}

	res, err := RunTmpl2("config-map", obj)
	if err != nil {
		return err
	}

	if exists {
		_, _, err = p.put("/api/v1/namespaces/{namespace}/configmaps/"+name, res)
	} else {
		_, _, err = p.post("/api/v1/namespaces/{namespace}/configmaps", res)
	}
	if err != nil {
		return err
	}
	return nil
}

// CreateResource implements the Provider interface
func (p *Provider) CreateResource(node *proto.Node) (*proto.Node, error) {
	node = node.Copy()

	// In order to do pod communication we need to create a headless service
	// TODO: Lets find a better place for this.
	obj := map[string]string{
		"Ensemble": node.Cluster,
		"Backend":  "",
	}
	data, err := RunTmpl2("headless-service", obj)
	if err != nil {
		return nil, err
	}
	if _, _, err := p.post("/api/v1/namespaces/{namespace}/services", data); err != nil {
		if err != errAlreadyExists {
			return nil, err
		}
	}

	if len(node.Spec.Files) > 0 {
		// store all the files under the '-files' prefix
		if err := p.upsertConfigMap(node.ID+"-files", node.Spec.Files); err != nil {
			return nil, err
		}
	}

	pod := &Pod{
		Name:     node.ID,
		Builder:  node.Spec,
		Ensemble: node.Cluster,
	}
	data, err = MarshalPod(pod)
	if err != nil {
		return nil, err
	}

	// create the Pod resource
	if _, _, err = p.post("/api/v1/namespaces/{namespace}/pods", data); err != nil {
		return nil, err
	}

	// wait for the resource to be running
	var res struct {
		Status struct {
			Phase string
			PodIP string
		}
	}
	for {
		if _, err := p.get("/api/v1/namespaces/{namespace}/pods/"+node.ID, &res); err != nil {
			return nil, err
		}
		if res.Status.Phase == "Running" {
			break
		}
		p.logger.Trace("create resource pod status", "id", node.ID, "status", res.Status.Phase)
		time.Sleep(1 * time.Second)
	}

	// for now we use the same id for the handle
	node.Handle = node.ID
	node.Addr = res.Status.PodIP

	// upsert the node spec
	nn, err := p.upsertNodeSpec(node)
	if err != nil {
		return nil, err
	}
	return nn, nil
}

func K8sFactory(logger hclog.Logger, c map[string]interface{}) (*Provider, error) {
	config, err := GetConfig()
	if err != nil {
		return nil, err
	}
	p := &Provider{
		client:    NewKubeClient(config),
		logger:    logger.Named("k8s"),
		stopCh:    make(chan struct{}),
		taskQueue: newTaskQueue(),
	}
	return p, nil
}

type task struct {
	*proto.Task

	// internal fields for the sort heap
	ready     bool
	index     int
	timestamp time.Time

	APIEndpoint string
}

type taskQueue struct {
	heap     taskQueueImpl
	lock     sync.Mutex
	items    map[string]*task
	updateCh chan struct{}
}

func newTaskQueue() *taskQueue {
	return &taskQueue{
		heap:     taskQueueImpl{},
		items:    map[string]*task{},
		updateCh: make(chan struct{}),
	}
}

func (t *taskQueue) get(id string) (*task, bool) {
	t.lock.Lock()
	defer t.lock.Unlock()

	tt, ok := t.items[id]
	return tt, ok
}

func (t *taskQueue) add(tt *task) {
	t.lock.Lock()
	defer t.lock.Unlock()

	tt.ready = true

	t.items[tt.ID] = tt
	heap.Push(&t.heap, tt)

	select {
	case t.updateCh <- struct{}{}:
	default:
	}
}

func (t *taskQueue) pop(stopCh chan struct{}) *task {
POP:
	t.lock.Lock()
	if len(t.heap) != 0 && t.heap[0].ready {
		// pop the first value
		tt := t.heap[0]
		tt.ready = false
		heap.Fix(&t.heap, tt.index)
		t.lock.Unlock()

		return tt
	}
	t.lock.Unlock()

	select {
	case <-t.updateCh:
		goto POP
	case <-stopCh:
		return nil
	}
}

func (t *taskQueue) finalize(id string) {
	t.lock.Lock()
	defer t.lock.Unlock()

	i, ok := t.items[id]
	if ok {
		heap.Remove(&t.heap, i.index)
	}
}

type taskQueueImpl []*task

func (t taskQueueImpl) Len() int { return len(t) }

func (t taskQueueImpl) Less(i, j int) bool {
	iNoReady, jNoReady := !t[i].ready, !t[j].ready
	if iNoReady && jNoReady {
		return false
	} else if iNoReady {
		return false
	} else if jNoReady {
		return true
	}
	return t[i].timestamp.Before(t[j].timestamp)
}

func (t taskQueueImpl) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
	t[i].index = i
	t[j].index = j
}

func (t *taskQueueImpl) Push(x interface{}) {
	n := len(*t)
	item := x.(*task)
	item.index = n
	*t = append(*t, item)
}

func (t *taskQueueImpl) Pop() interface{} {
	old := *t
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	*t = old[0 : n-1]
	return item
}

// Http methods

func (p *Provider) httpReq(method string, rawURL string, obj []byte) ([]byte, *Metadata, error) {
	url := transformURL(rawURL)

	resp, err := p.client.HTTPReq(method, url, obj)
	if err != nil {
		return nil, nil, err
	}
	if err := isError(resp); err != nil {
		return nil, nil, err
	}

	var item *Item
	if err := json.Unmarshal(resp, &item); err != nil {
		return nil, nil, err
	}
	return resp, item.Metadata, nil
}

func (p *Provider) listFull(rawURL string, callback func(i *Item)) error {
	return p.client.listFull(transformURL(rawURL), callback)
}

func (p *Provider) put(rawURL string, obj []byte) ([]byte, *Metadata, error) {
	return p.httpReq(http.MethodPut, rawURL, obj)
}

func (p *Provider) post(rawURL string, obj []byte) ([]byte, *Metadata, error) {
	return p.httpReq(http.MethodPost, rawURL, obj)
}

func (p *Provider) delete(rawURL string, obj []byte) error {
	_, _, err := p.httpReq(http.MethodDelete, rawURL, obj)
	return err
}

func (p *Provider) get(rawURL string, out interface{}) ([]byte, error) {
	url := transformURL(rawURL)

	resp, err := p.client.Get(url)
	if err != nil {
		return nil, err
	}
	if err := isError(resp); err != nil {
		return nil, err
	}
	if out == nil {
		return resp, nil
	}
	if err := json.Unmarshal(resp, out); err != nil {
		return nil, err
	}
	return resp, nil
}

var (
	errAlreadyExists = fmt.Errorf("Already exists")

	errInvalidResourceVersion = fmt.Errorf("Invalid resource version")

	errNotFound = fmt.Errorf("Not found")

	err404NotFound = fmt.Errorf("404 not found")
)

func isError(resp []byte) error {
	if strings.Contains(string(resp), "404 page not found") {
		return err404NotFound
	}
	type Status struct {
		Kind    string
		Status  string
		Reason  string
		Message string
	}
	var res Status
	if err := json.Unmarshal(resp, &res); err != nil {
		return nil
	}
	if res.Status != "Failure" {
		return nil
	}
	switch res.Reason {
	case "NotFound":
		return errNotFound
	case "AlreadyExists":
		return errAlreadyExists
	case "Invalid":
		return fmt.Errorf(res.Message)
	default:
		return fmt.Errorf("Undefined Error: '%s'", string(resp))
	}
}

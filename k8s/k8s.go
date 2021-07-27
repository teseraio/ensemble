package k8s

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/teseraio/ensemble/lib/mount"
	"github.com/teseraio/ensemble/operator"
	"github.com/teseraio/ensemble/operator/proto"
	"github.com/teseraio/ensemble/schema"
	"google.golang.org/grpc"
)

const crdURL = "/apis/apiextensions.k8s.io/v1/customresourcedefinitions"

// Provider is a Provider implementation for kubernetes.
type Provider struct {
	client *KubeClient
	logger hclog.Logger
	stopCh chan struct{}
	cplane operator.ControlPlane
	queue  *eventQueue
	// watchCh chan *proto.InstanceUpdate
}

// Stop stops the kubernetes provider
func (p *Provider) Stop() {
	close(p.stopCh)
}

var eventsURL = "/apis/events.k8s.io/v1/events"

func getIDFromRef(ref string) string {
	spl := strings.Split(ref, ".")
	return spl[0]
}

func (p *Provider) Setup(cplane operator.ControlPlane) error {
	p.cplane = cplane

	w, err := NewWatcher(p.logger, p.client, eventsURL, &Event{})
	if err != nil {
		return err
	}
	w.Run(p.stopCh)
	w.ForEach(func(task *WatchEntry, i interface{}) {
		if err := p.handleEvent(i.(*Event)); err != nil {
			p.logger.Error("failed to handle event", "err", err)
		}
	})

	ch := cplane.SubscribeInstanceUpdates()
	go func() {
		for {
			msg := <-ch
			if msg == nil {
				return
			}
			instance, err := cplane.GetInstance(msg.Id, msg.Cluster)
			if err != nil {
				p.logger.Error("failed to get instance", "err", err)
				continue
			}

			if instance.Status == proto.Instance_PENDING {
				p.queue.add(instance)
			} else if instance.Status == proto.Instance_TAINTED {
				p.queue.add(instance)
			}
		}
	}()

	go p.queueRun()
	return nil
}

func (p *Provider) handleEvent(event *Event) error {
	id := getIDFromRef(event.GetMetadata().Name)
	cluster := p.getPodCluster(id)

	if cluster == "" {
		return fmt.Errorf("cluster not found")
	}

	i, err := p.cplane.GetInstance(id, "")
	if err != nil {
		return err
	}
	i = i.Copy()

	if event.Reason == "Started" {
		// spetial, set running once we have the ip
		i.Set("started", "true")
		if err := p.cplane.UpsertInstance(i); err != nil {
			return err
		}
	}

	if event.Reason == "Killing" {
		// How to differentiate from a failed instance
		i.Status = proto.Instance_STOPPED
		if err := p.cplane.UpsertInstance(i); err != nil {
			return err
		}
	}

	return nil
}

func (p *Provider) createCRD(crdDefinition []byte) error {
	_, _, err := p.post(crdURL, crdDefinition)
	return err
}

// Start starts the kubernetes provider
func (p *Provider) Start() error {
	// create the protocol
	conn, err := grpc.Dial("127.0.0.1:6001", grpc.WithInsecure())
	if err != nil {
		return err
	}

	clt := proto.NewEnsembleServiceClient(conn)
	if err := p.trackCRDs(clt); err != nil {
		return err
	}
	return nil
}

func (p *Provider) Resources() operator.ProviderResources {
	return operator.ProviderResources{
		Resources: schema.Schema2{},
		Storage:   schema.Schema2{},
	}
}

/*
func (p *Provider) WatchUpdates() chan *proto.InstanceUpdate {
	return p.watchCh
}
*/

var emptyDel = []byte("{}")

// DeleteResource implements the Provider interface
func (p *Provider) DeleteResource(node *proto.Instance) (*proto.Instance, error) {
	if err := p.delete("/api/v1/namespaces/{namespace}/pods/"+node.ID, emptyDel); err != nil {
		return nil, err
	}
	return node, nil
}

// Exec implements the Provider interface
func (p *Provider) Exec(handler string, path string, cmdArgs ...string) (string, error) {
	out, err := p.client.Exec(handler, "main-container", path, cmdArgs...)
	if err != nil {
		return "", err
	}
	p.logger.Trace("exec", "handler", handler, "out", string(out))
	return string(out), nil
}

func transformURL(rawURL string) string {
	url := rawURL
	url = strings.Replace(url, "{namespace}", "default", -1)
	return url
}

func cleanPath(path string) string {
	// clean the path to k8s format
	return strings.Replace(strings.Trim(path, "/"), "/", ".", -1)
}

func (p *Provider) upsertConfigMap(name string, files *mount.MountPoint) error {
	// check if the config map exists
	var item *Item

	if _, err := p.get("/api/v1/namespaces/{namespace}/configmaps/"+name, &item); err != nil {
		if err != errNotFound {
			return err
		}
	}

	exists := item != nil

	parts := []string{}
	for name, content := range files.Files {
		content = strings.Replace(content, "\n", "\\n", -1)
		content = strings.Replace(content, "\"", "\\\"", -1)
		parts = append(parts, fmt.Sprintf("\"%s\": \"%s\"", cleanPath(name), content))
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

func (p *Provider) createHeadlessService(cluster string) error {
	// In order to do pod communication we need to create a headless service
	// TODO: Lets find a better place for this.
	obj := map[string]string{
		"Ensemble": cluster,
	}
	data, err := RunTmpl2("headless-service", obj)
	if err != nil {
		return err
	}
	if _, _, err := p.post("/api/v1/namespaces/{namespace}/services", data); err != nil {
		if err != errAlreadyExists {
			return err
		}
	}
	return nil
}

func (p *Provider) getPodCluster(id string) string {
	var obj *Item
	for i := 0; i < 10; i++ {
		if _, err := p.get("/api/v1/namespaces/{namespace}/pods/"+id, &obj); err != nil {
			if err != errNotFound {
				panic(err)
			}
		} else {
			break
		}
	}
	if obj == nil {
		return ""
	}
	return obj.Metadata.Labels["ensemble"]
}

func (p *Provider) getPodIP(id string) string {
	// wait for the resource to be running
	var res struct {
		Status struct {
			Phase string
			PodIP string
		}
	}
	for {
		if _, err := p.get("/api/v1/namespaces/{namespace}/pods/"+id, &res); err != nil {
			panic(err)
		}
		if res.Status.Phase == "Running" {
			break
		}
		p.logger.Trace("create resource pod status", "id", id, "status", res.Status.Phase)
		time.Sleep(50 * time.Millisecond)
	}
	return res.Status.PodIP
}

func (p *Provider) createVolume(instance *proto.Instance, m *proto.Instance_Mount) error {
	claimName := instance.Name + "-" + m.Name

	res, err := RunTmpl2("volume-claim", map[string]interface{}{
		"Name":        claimName,
		"StorageName": "local-path",
		"Storage":     "1Gi", // TODO
	})
	if err != nil {
		return err
	}
	if _, _, err = p.post("/api/v1/namespaces/{namespace}/persistentvolumeclaims", res); err != nil {
		if err != errAlreadyExists {
			return err
		}
	}
	return nil
}

func (p *Provider) Name() string {
	return "Kubernetes"
}

func (p *Provider) queueRun() {
	for {
		eval := p.queue.pop(p.stopCh)
		if eval == nil {
			return
		}

		instance := eval.(*proto.Instance)
		if err := p.handleQueueEvent(instance); err != nil {
			p.logger.Error("failed to handle event", "err", err)
		}
	}
}

func (p *Provider) handleQueueEvent(instance *proto.Instance) error {
	if instance.Status == proto.Instance_PENDING {
		if instance.Get("created") == "" {
			if err := p.createImpl(instance); err != nil {
				return err
			}
			instance.Set("created", "true")
			if err := p.cplane.UpsertInstance(instance); err != nil {
				return err
			}
		}
		if instance.Get("started") == "true" {
			// its already started, wait for the ip
			ip := p.getIpImpl(instance.ID)
			if ip == "" {
				// enqueue again
				time.Sleep(1 * time.Second)
				p.queue.add(instance)
			} else {
				instance.Status = proto.Instance_RUNNING
				instance.Ip = ip
				if err := p.cplane.UpsertInstance(instance); err != nil {
					return err
				}
			}
		}
	} else if instance.Status == proto.Instance_TAINTED {
		if _, err := p.DeleteResource(instance); err != nil {
			return err
		}
	}
	return nil
}

func (p *Provider) getIpImpl(id string) string {
	var res struct {
		Status struct {
			Phase string
			PodIP string
		}
	}
	if _, err := p.get("/api/v1/namespaces/{namespace}/pods/"+id, &res); err != nil {
		panic(err)
	}
	if res.Status.Phase == "Running" {
		return res.Status.PodIP
	}
	return ""
}

func (p *Provider) createImpl(node *proto.Instance) error {
	node = node.Copy()

	// create headless service for dns resolving
	if err := p.createHeadlessService(node.ClusterName); err != nil {
		return fmt.Errorf("failed to upsert headless service: %v", err)
	}

	// files to be mounted on the pod
	if len(node.Spec.Files) != 0 {
		mountPoints, err := mount.CreateMountPoints(node.Spec.Files)
		if err != nil {
			return err
		}
		for indx, mountPoint := range mountPoints {
			name := node.ID + "-file-data-" + strconv.Itoa(indx)

			if err := p.upsertConfigMap(name, mountPoint); err != nil {
				return fmt.Errorf("failed to upsert config map: %v", err)
			}
		}
	}

	// volume mount
	if len(node.Mounts) != 0 {
		for _, m := range node.Mounts {
			if err := p.createVolume(node, m); err != nil {
				return err
			}
		}
	}

	data, err := MarshalPod(node)
	if err != nil {
		return fmt.Errorf("failed to marshal pod: %v", err)
	}
	if _, _, err = p.post("/api/v1/namespaces/{namespace}/pods", data); err != nil {
		return fmt.Errorf("failed to create pod: %v", err)
	}
	return nil
}

func K8sFactory(logger hclog.Logger, c map[string]interface{}) (*Provider, error) {
	config, err := GetConfig()
	if err != nil {
		return nil, err
	}
	p := &Provider{
		client: NewKubeClient(config),
		logger: logger.Named("k8s"),
		stopCh: make(chan struct{}),
		queue:  newEventQueue(),
		//watchCh: make(chan *proto.InstanceUpdate, 5),
	}
	return p, nil
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
	errAlreadyExists = fmt.Errorf("already exists")

	errInvalidResourceVersion = fmt.Errorf("invalid resource version")

	errNotFound = fmt.Errorf("not found")

	err404NotFound = fmt.Errorf("404 not found")

	errExpired = fmt.Errorf("expired")

	errFutureVersion = fmt.Errorf("future version")
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
		Code    int
	}
	var res Status
	if err := json.Unmarshal(resp, &res); err != nil {
		return nil
	}
	if res.Kind == "" {
		// try to parse as error
		var errMsg struct {
			Type   string
			Object *Status
		}
		if err := json.Unmarshal(resp, &errMsg); err != nil {
			return nil
		}
		if errMsg.Type != "ERROR" {
			return nil
		}
		res = *errMsg.Object
	}
	if res.Status != "Failure" {
		return nil
	}
	switch res.Reason {
	case "NotFound":
		return errNotFound
	case "AlreadyExists":
		return errAlreadyExists
	case "Expired":
		return errExpired
	case "Invalid":
		return fmt.Errorf(res.Message)
	default:
		// filter by code
		switch res.Code {
		case 500:
			return errFutureVersion
		}
		return fmt.Errorf("undefined Error: '%s'", string(resp))
	}
}

type eventQueue struct {
	lock     sync.Mutex
	events   []interface{}
	updateCh chan struct{}
}

func newEventQueue() *eventQueue {
	return &eventQueue{
		events:   []interface{}{},
		updateCh: make(chan struct{}),
	}
}

func (e *eventQueue) add(task interface{}) {
	e.lock.Lock()
	defer e.lock.Unlock()

	e.events = append(e.events, task)

	select {
	case e.updateCh <- struct{}{}:
	default:
	}
}

func (e *eventQueue) popImpl() interface{} {
	e.lock.Lock()
	if len(e.events) != 0 {
		// pop the first value and remove it from the heap
		var pop interface{}
		pop, e.events = e.events[0], e.events[1:]
		e.lock.Unlock()

		return pop
	}
	e.lock.Unlock()
	return nil
}

func (e *eventQueue) pop(stopCh chan struct{}) interface{} {
POP:
	tt := e.popImpl()
	if tt != nil {
		return tt
	}

	select {
	case <-e.updateCh:
		goto POP
	case <-stopCh:
		return nil
	}
}

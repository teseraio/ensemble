package k8s

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/mitchellh/mapstructure"
	"github.com/teseraio/ensemble/operator/proto"
	"google.golang.org/grpc"
)

const crdURL = "/apis/apiextensions.k8s.io/v1/customresourcedefinitions"

// Provider is a Provider implementation for kubernetes.
type Provider struct {
	client  *KubeClient
	logger  hclog.Logger
	stopCh  chan struct{}
	watchCh chan *proto.InstanceUpdate
}

// Stop stops the kubernetes provider
func (p *Provider) Stop() {
	close(p.stopCh)
}

var eventsURL = "/apis/events.k8s.io/v1/events"

var podsURL2 = "/api/v1/namespaces/default/pods"

func getIDFromRef(ref string) string {
	spl := strings.Split(ref, ".")
	return spl[0]
}

func (p *Provider) Setup() error {
	go func() {
		store := newStore()
		newWatcher(store, p.client, eventsURL, &Event{}, false)

		for {
			task := store.pop(context.Background())
			event := task.item.(*Event)

			id := getIDFromRef(event.GetMetadata().Name)
			cluster := p.getPodCluster(id)

			if cluster == "" {
				continue
			}

			if event.Reason == "Started" {
				// query the node and get the ip

				ip := p.getPodIP(id)

				p.watchCh <- &proto.InstanceUpdate{
					ID:      id,
					Cluster: cluster,
					Event: &proto.InstanceUpdate_Conf{
						Conf: &proto.InstanceUpdate_ConfDone{
							Ip: ip,
						},
					},
				}
			}

			if event.Reason == "Killing" {
				p.watchCh <- &proto.InstanceUpdate{
					ID:      id,
					Cluster: cluster,
					Event: &proto.InstanceUpdate_Status{
						Status: &proto.InstanceUpdate_StatusChange{
							Status: proto.InstanceUpdate_StatusChange_FAILED,
						},
					},
				}
			}
		}
	}()

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

type Resource struct {
}

func (p *Provider) Resources() interface{} {
	return &Resource{}
}

func (p *Provider) WatchUpdates() chan *proto.InstanceUpdate {
	return p.watchCh
}

var emptyDel = []byte("{}")

// DeleteResource implements the Provider interface
func (p *Provider) DeleteResource(node *proto.Instance) (*proto.Instance, error) {
	if err := p.delete("/api/v1/namespaces/{namespace}/pods/"+node.ID, emptyDel); err != nil {
		return nil, err
	}
	return node, nil
}

// Exec implements the Provider interface
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

	fmt.Println("-- headless --")
	fmt.Println(string(data))

	if _, _, err := p.post("/api/v1/namespaces/{namespace}/services", data); err != nil {
		if err != errAlreadyExists {
			return err
		}
	}
	return nil
}

func (p *Provider) getPodCluster(id string) string {
	var obj *Item
	if _, err := p.get("/api/v1/namespaces/{namespace}/pods/"+id, &obj); err != nil {
		panic(err)
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
		time.Sleep(1 * time.Second)
	}
	return res.Status.PodIP
}

// CreateResource implements the Provider interface
func (p *Provider) CreateResource(node *proto.Instance) (*proto.Instance, error) {
	p.logger.Debug("upsert instance", "id", node.ID, "cluster", node.Cluster, "name", node.Name)

	node = node.Copy()
	if err := p.createHeadlessService(node.Cluster); err != nil {
		return nil, err
	}
	if len(node.Spec.Files) > 0 {
		// store all the files under the '-files' prefix
		if err := p.upsertConfigMap(node.ID+"-files", node.Spec.Files); err != nil {
			return nil, err
		}
	}

	fmt.Println("-- cmd --")
	fmt.Println(node.Spec.Cmd)

	data, err := MarshalPod(node)
	if err != nil {
		return nil, err
	}

	fmt.Println("-- data --")
	fmt.Println(string(data))

	// create the Pod resource
	if _, _, err = p.post("/api/v1/namespaces/{namespace}/pods", data); err != nil {
		return nil, err
	}
	return node, nil
}

func K8sFactory(logger hclog.Logger, c map[string]interface{}) (*Provider, error) {
	config, err := GetConfig()
	if err != nil {
		return nil, err
	}
	p := &Provider{
		client:  NewKubeClient(config),
		logger:  logger.Named("k8s"),
		stopCh:  make(chan struct{}),
		watchCh: make(chan *proto.InstanceUpdate, 5),
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

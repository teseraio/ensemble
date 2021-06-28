package k8s

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
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

			if event.Reason == "Failed" {
				p.watchCh <- &proto.InstanceUpdate{
					ID:          id,
					ClusterName: cluster,
					Event: &proto.InstanceUpdate_Failed_{
						Failed: &proto.InstanceUpdate_Failed{},
					},
				}
			}

			if event.Reason == "Scheduled" {
				// is being started
				p.watchCh <- &proto.InstanceUpdate{
					ID:          id,
					ClusterName: cluster,
					Event: &proto.InstanceUpdate_Scheduled_{
						Scheduled: &proto.InstanceUpdate_Scheduled{},
					},
				}
			}

			if event.Reason == "Started" {
				// query the node and get the ip

				ip := p.getPodIP(id)

				p.watchCh <- &proto.InstanceUpdate{
					ID:          id,
					ClusterName: cluster,
					Event: &proto.InstanceUpdate_Running_{
						Running: &proto.InstanceUpdate_Running{
							Ip: ip,
						},
					},
				}
			}

			if event.Reason == "Killing" {
				p.watchCh <- &proto.InstanceUpdate{
					ID:          id,
					ClusterName: cluster,
					Event: &proto.InstanceUpdate_Killing_{
						Killing: &proto.InstanceUpdate_Killing{},
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

func (p *Provider) Resources() operator.ProviderResources {
	return operator.ProviderResources{
		Resources: schema.Schema2{},
		Storage:   schema.Schema2{},
	}
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
		"StorageName": "standard",
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

// CreateResource implements the Provider interface
func (p *Provider) CreateResource(node *proto.Instance) (*proto.Instance, error) {
	p.logger.Debug("upsert instance", "id", node.ID, "cluster", node.ClusterName, "name", node.Name)
	node = node.Copy()

	// create headless service for dns resolving
	if err := p.createHeadlessService(node.ClusterName); err != nil {
		return nil, err
	}

	// files to be mounted on the pod
	if len(node.Spec.Files) != 0 {
		mountPoints, err := mount.CreateMountPoints(node.Spec.Files)
		if err != nil {
			return nil, err
		}
		for indx, mountPoint := range mountPoints {
			name := node.ID + "-file-data-" + strconv.Itoa(indx)

			if err := p.upsertConfigMap(name, mountPoint); err != nil {
				return nil, err
			}
		}
	}

	// volume mount
	if len(node.Mounts) != 0 {
		for _, m := range node.Mounts {
			if err := p.createVolume(node, m); err != nil {
				return nil, err
			}
		}
	}

	data, err := MarshalPod(node)
	if err != nil {
		return nil, err
	}

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

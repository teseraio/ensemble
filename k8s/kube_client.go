package k8s

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/teseraio/ensemble/k8s/spdy"
)

// Config is the configuration for KubeClient
type Config struct {
	Host        string
	TLSConfig   *tls.Config
	BearerToken string
}

// KubeClient is a kubernetes client
type KubeClient struct {
	config *Config
	client *http.Client
}

// NewKubeClient creates a new KubeClient
func NewKubeClient(config *Config) *KubeClient {
	client := &http.Client{
		Transport: &http.Transport{TLSClientConfig: config.TLSConfig},
	}
	k8sClient := &KubeClient{
		client: client,
		config: config,
	}
	return k8sClient
}

type ListResponse struct {
	Items    []*Item
	Metadata ListMetadata
}

type ListMetadata struct {
	Continue        string
	ResourceVersion string
}

type ListOpts struct {
	Continue string
	Limit    int
}

func (c *KubeClient) List(path string, opts *ListOpts) (*ListResponse, error) {
	if !strings.Contains(path, "?") {
		path = path + "?"
	} else {
		path = path + "&"
	}
	if opts != nil {
		if opts.Continue != "" {
			path += "continue=" + opts.Continue + "&"
		}
		if opts.Limit != 0 {
			path += "limit=" + strconv.Itoa(opts.Limit)
		}
	}

	data, err := c.Get(path)
	if err != nil {
		return nil, err
	}

	var result *ListResponse
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *KubeClient) Track(ctx context.Context, path string) (chan *Item, chan struct{}) {
	itemCh := make(chan *Item)
	watchCh := make(chan struct{})

	go func() {
		if err := c.trackImpl(ctx, path, itemCh, watchCh); err != nil {
			// TODO
			fmt.Printf("ERR: %v", err)
		}
	}()

	return itemCh, watchCh
}

func (c *KubeClient) listFull(path string, callback func(i *Item)) error {
	// list
	opts := &ListOpts{
		Limit: 10,
	}
	for {
		listResp, err := c.List(path, opts)
		if err != nil {
			return err
		}
		for _, i := range listResp.Items {
			callback(i)
		}
		if listResp.Metadata.Continue == "" {
			break
		}
		opts.Continue = listResp.Metadata.Continue
	}
	return nil
}

func (c *KubeClient) trackImpl(ctx context.Context, path string, itemCh chan *Item, watchCh chan struct{}) error {

	// list
	opts := &ListOpts{
		Limit: 10,
	}
	var resourceVersion string
	for {
		listResp, err := c.List(path, opts)
		if err != nil {
			return err
		}
		for _, i := range listResp.Items {
			itemCh <- i
		}
		if listResp.Metadata.Continue == "" {
			resourceVersion = listResp.Metadata.ResourceVersion
			break
		}
		opts.Continue = listResp.Metadata.Continue
	}

	// notify that we start the watch, weird
	close(watchCh)

	// watch
	if err := c.watchImpl(path, resourceVersion, itemCh); err != nil {
		return err
	}

	return nil
}

func (c *KubeClient) watchImpl(path string, resourceVersion string, itemCh chan *Item) error {
	path = path + "?watch=true&resourceVersion=" + resourceVersion
	resp, err := c.HTTPReqWithResponse(http.MethodGet, path, nil)
	if err != nil {
		return err
	}
	buffer := bufio.NewReader(resp.Body)
	for {
		res, err := buffer.ReadBytes(byte('\n'))
		if err != nil {
			if err == io.EOF {
				continue
			}
			return err
		}
		var evnt watchEvent
		if err := json.Unmarshal(res, &evnt); err != nil {
			return err
		}
		itemCh <- evnt.Object
	}
}

type watchEvent struct {
	Type   string
	Object *Item
}

func (c *KubeClient) Get(path string) ([]byte, error) {
	return c.HTTPReq(http.MethodGet, path, nil)
}

func (c *KubeClient) Delete(path string, obj []byte) ([]byte, error) {
	return c.HTTPReq(http.MethodDelete, path, obj)
}

func (c *KubeClient) Put(path string, obj []byte) ([]byte, error) {
	return c.HTTPReq(http.MethodPut, path, obj)
}

func (c *KubeClient) Post(path string, obj []byte) ([]byte, error) {
	return c.HTTPReq(http.MethodPost, path, obj)
}

// HTTPReqWithResponse is a generic method to make http requests that returns the response object
func (c *KubeClient) HTTPReqWithResponse(method string, path string, obj []byte) (*http.Response, error) {
	var reader io.Reader
	if obj != nil {
		reader = bytes.NewReader(obj)
	}
	req, err := http.NewRequest(method, c.config.Host+path, reader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("content-type", "application/json")
	if c.config.BearerToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.config.BearerToken)
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *KubeClient) HTTPReq(method string, path string, obj []byte) ([]byte, error) {
	resp, err := c.HTTPReqWithResponse(method, path, obj)
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (c *KubeClient) Exec(pod, container string, path string, cmdArgs ...string) ([]byte, error) {
	cmd := []string{path}
	cmd = append(cmd, cmdArgs...)

	url := c.config.Host + "/api/v1/namespaces/default/pods/" + pod + "/exec"

	remoteCommand := spdy.RemoteCommand{
		URL:         url,
		TLSConfig:   c.config.TLSConfig,
		BearerToken: c.config.BearerToken,
	}
	args := &spdy.Args{
		Container: container,
		Command:   cmd,
	}
	out, err := remoteCommand.Execute(args)
	if err != nil {
		return nil, err
	}
	return out, nil
}

type Metadata struct {
	Name            string
	ResourceVersion string            `json:"resourceVersion"`
	UID             string            `json:"uid"`
	Labels          map[string]string `json:"labels"`
	Generation      int               `json:"generation"`
	Annotations     map[string]string `json:"annotations"`
	SelfLink        string            `json:"selfLink"`
}

type Item struct {
	Metadata *Metadata
	Kind     string
	Spec     map[string]interface{}
	Status   map[string]interface{}
	Data     map[string]interface{}
}

// InCluster returns whether we are running inside a Kubernetes pod
func InCluster() bool {
	return os.Getenv("KUBERNETES_SERVICE_HOST") != "" && os.Getenv("KUBERNETES_SERVICE_PORT") != ""
}

const (
	inClusterTokenFile  = "/var/run/secrets/kubernetes.io/serviceaccount/token"
	inClusterRootCAFile = "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"
)

// InClusterConfig returns the auth from inside the pod
func InClusterConfig() (*Config, error) {
	if !InCluster() {
		return nil, fmt.Errorf("not in cluster")
	}

	token, err := ioutil.ReadFile(inClusterTokenFile)
	if err != nil {
		return nil, err
	}

	caFile, err := ioutil.ReadFile(inClusterRootCAFile)
	if err != nil {
		return nil, err
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM([]byte(caFile)) {
		return nil, fmt.Errorf("bad")
	}

	tlsConf := &tls.Config{
		RootCAs: certPool,
	}

	host := fmt.Sprintf("https://%s:%s", os.Getenv("KUBERNETES_SERVICE_HOST"), os.Getenv("KUBERNETES_SERVICE_PORT"))
	c := &Config{
		Host:        host,
		TLSConfig:   tlsConf,
		BearerToken: string(token),
	}
	return c, nil
}

func GetConfig() (*Config, error) {
	if InCluster() {
		return InClusterConfig()
	}

	// use kube config
	kubeConfig, err := NewKubeConfig("", "")
	if err != nil {
		return nil, err
	}
	config, err := kubeConfig.ToConfig()
	if err != nil {
		return nil, err
	}
	return config, nil
}

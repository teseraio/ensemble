package k8s

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"

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

type WatchEvent struct {
	Type   string
	Object interface{}
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
	if method == "PATCH" {
		req.Header.Set("content-type", "application/json-patch+json")
	} else {
		req.Header.Set("content-type", "application/json")
	}
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
	Data     map[string]interface{}
}

func (i *Item) GetMetadata() *Metadata {
	return i.Metadata
}

type EventRegarding struct {
	Kind string
}

type Event struct {
	Metadata  *Metadata
	Reason    string
	Regarding *EventRegarding
	Note      string
	Type      string
}

func (e *Event) GetMetadata() *Metadata {
	return e.Metadata
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

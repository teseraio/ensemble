package testutil

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/docker/client"
	"github.com/teseraio/ensemble/operator"
	"github.com/teseraio/ensemble/operator/proto"
)

var _ operator.Provider = &Client{}

type resource struct {
	nodeID    string
	handle    string
	clusterID string
}

// Client is a sugarcoat version of the docker client
type Client struct {
	client    *client.Client
	resources []*resource

	// TODO: Not here
	volumes  map[string]string
	updateCh chan *operator.NodeUpdate
}

// NewClient creates a new docker Client
func NewDockerClient() (*Client, error) {
	cli, err := client.NewEnvClient()
	if err != nil {
		return nil, err
	}
	clt := &Client{
		client:    cli,
		resources: []*resource{},
		volumes:   map[string]string{},
		updateCh:  make(chan *operator.NodeUpdate),
	}
	return clt, nil
}

func (c *Client) Start() error {
	return nil
}

func (c *Client) Setup() error {
	return nil
}

func (c *Client) Remove(id string) error {
	if err := c.client.ContainerStop(context.Background(), id, nil); err != nil {
		return err
	}
	return nil
}

func (c *Client) WatchUpdates() chan *operator.NodeUpdate {
	return c.updateCh
}

// PullImage pulls a docker image
func (c *Client) PullImage(ctx context.Context, image string) error {
	if strings.HasPrefix(image, "quay.io") {
		return nil
	}
	if strings.HasPrefix(image, "docker.elastic.co") {
		return nil
	}

	canonicalName := "docker.io/"
	if strings.Contains(image, "/") {
		canonicalName += image
	} else {
		canonicalName += "library/" + image
	}

	_, _, err := c.client.ImageInspectWithRaw(ctx, canonicalName)
	if err != nil {
		reader, err := c.client.ImagePull(ctx, canonicalName, types.ImagePullOptions{})
		if err != nil {
			return err
		}
		_, err = io.Copy(os.Stdout, reader)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) Clean() {
	for _, res := range c.resources {
		if err := c.client.ContainerStop(context.Background(), res.handle, nil); err != nil {
			panic(err)
		}
		if err := c.client.ContainerRemove(context.Background(), res.handle, types.ContainerRemoveOptions{}); err != nil {
			panic(err)
		}
	}
}

func (c *Client) GetIP(id string) string {
	res, err := c.client.ContainerInspect(context.Background(), id)
	if err != nil {
		panic(err)
	}
	return res.NetworkSettings.Networks["net1"].IPAddress
}

func createIfNotExists(path string) error {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return os.MkdirAll(path, 0755)
	}
	return err
}

func (c *Client) execImpl(ctx context.Context, id string, execCmd []string) error {
	ec := types.ExecConfig{
		User:         "",
		AttachStderr: true,
		AttachStdout: true,
		Cmd:          execCmd,
	}
	execID, err := c.client.ContainerExecCreate(ctx, id, ec)
	if err != nil {
		return err
	}

	conn, err := c.client.ContainerExecAttach(ctx, execID.ID, types.ExecConfig{})
	if err != nil {
		return err
	}
	defer conn.Close()

	o, err := ioutil.ReadAll(conn.Reader)
	if err != nil {
		return err
	}
	if _, err := c.client.ContainerExecInspect(ctx, execID.ID); err != nil {
		return err
	}

	fmt.Println(string(o))
	return nil
}

// Create creates a docker container
func (c *Client) Create(ctx context.Context, node *proto.Instance) (string, error) {
	// We will use the 'net1' network interface for dns resolving

	builder := node.Spec

	image := builder.Image + ":" + builder.Version
	name := node.FullName()

	// Build the volumes
	binds := []string{}
	if len(builder.Files) != 0 {
		dir, err := ioutil.TempDir("/tmp", "builder-")
		if err != nil {
			return "", err
		}
		for path, content := range builder.Files {
			fullPath := filepath.Join(dir, path)
			if err := createIfNotExists(filepath.Dir(fullPath)); err != nil {
				return "", err
			}
			if err := ioutil.WriteFile(fullPath, []byte(content), 0755); err != nil {
				return "", err
			}
			binds = append(binds, fmt.Sprintf("%s:%s", fullPath, path))
		}
	}

	if err := c.PullImage(ctx, image); err != nil {
		return "", err
	}

	env := []string{}
	for k, v := range builder.Env {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}
	config := &container.Config{
		Hostname: name,
		Image:    image,
		Env:      env,
		Cmd:      strslice.StrSlice(builder.Cmd),
	}
	hostConfig := &container.HostConfig{
		Binds: binds,
	}

	/*
		// decode computational resources
		resConfig := c.Resources().(*Resource)
		if err := mapstructure.WeakDecode(node.Resources.Spec, &resConfig); err != nil {
			return "", err
		}
		if resConfig != nil {
			hostConfig.Resources = container.Resources{
				CPUShares: int64(resConfig.CPUShares),
				CPUCount:  int64(resConfig.CPUCount),
			}
		}
	*/

	netConfig := &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			"net1": {},
		},
	}

	body, err := c.client.ContainerCreate(ctx, config, hostConfig, netConfig, name)
	if err != nil {
		panic(err)
	}

	c.resources = append(c.resources, &resource{
		nodeID:    node.ID,
		handle:    body.ID,
		clusterID: node.Cluster,
	})
	if err := c.client.ContainerStart(ctx, body.ID, types.ContainerStartOptions{}); err != nil {
		return "", err
	}

	// watch for updates in the node
	go func() {
		c.client.ContainerWait(context.Background(), body.ID)
		c.updateCh <- &operator.NodeUpdate{ID: node.ID, ClusterID: node.Cluster}
	}()
	return body.ID, nil
}

func (c *Client) Destroy(indx int) {
	res := c.resources[indx]

	// stop + remove does not work, so just remove with force
	if err := c.client.ContainerRemove(context.Background(), res.handle, types.ContainerRemoveOptions{Force: true}); err != nil {
		panic(err)
	}
}

type Resource struct {
	CPUShares uint64 `mapstructure:"cpuShares"`
	CPUCount  uint64 `mapstructure:"cpuCount"`
}

func (c *Client) Resources() interface{} {
	return &Resource{}
}

func (c *Client) CreateResource(node *proto.Instance) (*proto.Instance, error) {
	id, err := c.Create(context.TODO(), node)
	if err != nil {
		return nil, err
	}

	ip := c.GetIP(id)

	nn := node.Copy()

	fmt.Println("_ FORGOT _")
	fmt.Println(ip)

	/*
		nn.Addr = ip
		nn.Handle = id
	*/

	return nn, nil
}

func (c *Client) Exec(handler string, path string, args ...string) error {
	execCmd := []string{path}
	execCmd = append(execCmd, args...)

	return c.execImpl(context.Background(), handler, execCmd)
}

func (c *Client) DeleteResource(node *proto.Instance) (*proto.Instance, error) {
	panic("TODO")
	if err := c.Remove("node.Handle"); err != nil {
		return nil, err
	}
	return node, nil
}

package testutil

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/docker/client"
	"github.com/mitchellh/mapstructure"
	"github.com/teseraio/ensemble/operator"
	"github.com/teseraio/ensemble/operator/proto"
)

const networkName = "net1"

var _ operator.Provider = &Client{}

type resource struct {
	nodeID    string
	handle    string
	clusterID string
	instance  *proto.Instance
}

// Client is a sugarcoat version of the docker client
type Client struct {
	client *client.Client

	resources     map[string]*resource
	resourcesLock sync.Mutex

	workCh chan *proto.Instance

	// TODO: Not here
	volumes  map[string]string
	updateCh chan *proto.InstanceUpdate
}

// NewClient creates a new docker Client
func NewDockerClient() (*Client, error) {
	cli, err := client.NewEnvClient()
	if err != nil {
		return nil, err
	}
	clt := &Client{
		client:    cli,
		resources: map[string]*resource{},
		volumes:   map[string]string{},
		updateCh:  make(chan *proto.InstanceUpdate),
		workCh:    make(chan *proto.Instance, 5),
	}

	// upsert internal docker network 'net1' required for DNS support
	if _, err := cli.NetworkInspect(context.Background(), networkName); err != nil {
		if strings.Contains(err.Error(), "No such network") {
			if _, err := cli.NetworkCreate(context.Background(), networkName, types.NetworkCreate{CheckDuplicate: true}); err != nil {
				panic(err)
			}
		} else {
			return nil, err
		}
	}

	go clt.run()
	return clt, nil
}

func (c *Client) Start() error {
	return nil
}

func (c *Client) Setup() error {
	return nil
}

func (c *Client) Remove(id string) error {
	fmt.Printf("REMOVE==================>%s\n", id)
	if err := c.client.ContainerStop(context.Background(), id, nil); err != nil {
		return err
	}
	// do not remove the container name here but on the wait step because this part
	// is async
	return nil
}

func (c *Client) WatchUpdates() chan *proto.InstanceUpdate {
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
	return res.NetworkSettings.Networks[networkName].IPAddress
}

func createIfNotExists(path string) error {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return os.MkdirAll(path, 0755)
	}
	return err
}

func (c *Client) run() {
	worker := func(id string) {
		for instance := range c.workCh {
			if _, err := c.createImpl(context.Background(), instance); err != nil {
				panic(err)
				// send failed event
				c.updateCh <- &proto.InstanceUpdate{
					ID:      instance.ID,
					Cluster: instance.Cluster,
					Event: &proto.InstanceUpdate_Failed_{
						Failed: &proto.InstanceUpdate_Failed{},
					},
				}
			}
		}
	}
	numWorkers := 2
	for i := 0; i < numWorkers; i++ {
		go worker(strconv.Itoa(i))
	}
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

func (c *Client) removeByName(n string) error {
	c.resourcesLock.Lock()
	defer c.resourcesLock.Unlock()

	fmt.Println("_ REMOVE CANARY _")

	for name, i := range c.resources {
		if i.instance.FullName() == n {
			if err := c.Remove(i.handle); err != nil {
				return err
			}
			delete(c.resources, name)
		}
	}
	return nil
}

// Create creates a docker container
func (c *Client) createImpl(ctx context.Context, node *proto.Instance) (string, error) {
	c.resourcesLock.Lock()
	defer c.resourcesLock.Unlock()

	c.updateCh <- &proto.InstanceUpdate{
		ID:      node.ID,
		Cluster: node.Cluster,
		Event: &proto.InstanceUpdate_Scheduled_{
			Scheduled: &proto.InstanceUpdate_Scheduled{},
		},
	}

	// We will use the 'net1' network interface for dns resolving
	builder := node.Spec

	if builder.Version == "" {
		builder.Version = "latest"
	}
	image := builder.Image + ":" + builder.Version
	name := node.FullName()

	fmt.Printf("=====> CREATE: %s\n", name)
	fmt.Println(image)

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

	// decode computational resources
	if node.Group != nil {
		resConfig := c.Resources().(*Resource)
		if err := mapstructure.WeakDecode(node.Group.Resources, &resConfig); err != nil {
			return "", err
		}
		if resConfig != nil {
			hostConfig.Resources = container.Resources{
				CPUShares: int64(resConfig.CPUShares),
				CPUCount:  int64(resConfig.CPUCount),
			}
		}
	}

	netConfig := &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			networkName: {},
		},
	}

	body, err := c.client.ContainerCreate(ctx, config, hostConfig, netConfig, name)
	if err != nil {
		fmt.Println("- bad -")
		panic(err)
	}

	c.resources[node.ID] = &resource{
		nodeID:    node.ID,
		handle:    body.ID,
		clusterID: node.Cluster,
		instance:  node,
	}
	if err := c.client.ContainerStart(ctx, body.ID, types.ContainerStartOptions{}); err != nil {
		return "", err
	}

	// watch for updates in the node
	go func() {
		fmt.Println("> WAIT <")
		fmt.Println(body.ID)

		status, err := c.client.ContainerWait(context.Background(), body.ID)
		if err != nil {
			panic(err)
		}

		fmt.Println("====> status ====>")
		fmt.Println(status)
		fmt.Println(body.ID)

		// we need to remove it here so that we can reuse the name
		if err := c.client.ContainerRemove(context.Background(), body.ID, types.ContainerRemoveOptions{}); err != nil {
			panic(err)
		}

		c.updateCh <- &proto.InstanceUpdate{
			ID:      node.ID,
			Cluster: node.Cluster,
			Event: &proto.InstanceUpdate_Killing_{
				Killing: &proto.InstanceUpdate_Killing{},
			},
		}
	}()

	fmt.Println("XX")
	fmt.Println(body.ID)

	ip := c.GetIP(body.ID)

	c.updateCh <- &proto.InstanceUpdate{
		ID:      node.ID,
		Cluster: node.Cluster,
		Event: &proto.InstanceUpdate_Running_{
			Running: &proto.InstanceUpdate_Running{
				Ip:      ip,
				Handler: body.ID,
			},
		},
	}
	return body.ID, nil
}

func (c *Client) DestroyAt() {
	var handle string
	for _, j := range c.resources {
		handle = j.handle
	}
	if err := c.client.ContainerRemove(context.Background(), handle, types.ContainerRemoveOptions{Force: true}); err != nil {
		panic(err)
	}
}

func (c *Client) Destroy() {
	var resource *resource
	for _, v := range c.resources {
		resource = v
	}
	// stop + remove does not work, so just remove with force
	if err := c.client.ContainerRemove(context.Background(), resource.handle, types.ContainerRemoveOptions{Force: true}); err != nil {
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
	fmt.Printf("Create resource: %s\n", node.Name)

	// validation
	for _, r := range c.resources {
		if r.instance.ID == node.ID {
			return nil, operator.ErrInstanceAlreadyRunning
		}
		if r.instance.FullName() == node.FullName() {
			return nil, operator.ErrProviderNameAlreadyUsed
		}
	}

	// async serialize the execution
	c.workCh <- node

	return nil, nil
}

func (c *Client) Exec(handler string, path string, args ...string) error {
	execCmd := []string{path}
	execCmd = append(execCmd, args...)

	return c.execImpl(context.Background(), handler, execCmd)
}

func (c *Client) DeleteResource(node *proto.Instance) (*proto.Instance, error) {
	fmt.Printf("Delete resource: %s %s\n", node.Name, node.Handler)

	go func() {
		if err := c.Remove(node.Handler); err != nil {
			fmt.Printf("[ERR]: Failed to delete resource: %v\n", err)
		}
	}()
	return nil, nil
}

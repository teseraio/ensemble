package testutil

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/teseraio/ensemble/lib/mount"
	"github.com/teseraio/ensemble/lib/uuid"
	"github.com/teseraio/ensemble/operator"
	"github.com/teseraio/ensemble/operator/proto"
	"github.com/teseraio/ensemble/schema"
)

const networkName = "net1"

var _ operator.Provider = &Client{}

type resource struct {
	nodeID    string
	handle    string
	clusterID string
	instance  *proto.Instance
	active    bool
}

// Client is a sugarcoat version of the docker client
type Client struct {
	client *client.Client

	resources     map[string]*resource
	resourcesLock sync.Mutex

	//workCh chan *proto.Instance

	// TODO: Not here
	//volumes map[string]string
	//backendClient proto.BackendServiceClient

	controlPlane operator.ControlPlane

	//updateCh chan *proto.InstanceUpdate
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
	}

	// upsert internal docker network 'net1' required for DNS support
	if _, err := cli.NetworkInspect(context.Background(), networkName); err != nil {
		if strings.Contains(err.Error(), "No such network") {
			if _, err := cli.NetworkCreate(context.Background(), networkName, types.NetworkCreate{CheckDuplicate: true}); err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	return clt, nil
}

func (c *Client) Name() string {
	return "Docker"
}

func (c *Client) Start() error {
	return nil
}

func (c *Client) Setup(cc operator.ControlPlane) error {
	c.controlPlane = cc

	c.runProvider()
	return nil
}

func (c *Client) Remove(id string) error {
	if err := c.client.ContainerStop(context.Background(), id, nil); err != nil {
		return err
	}
	// do not remove the container name here but on the wait step because this part
	// is async
	// add a busybox image every time we delete something so that new instances
	// do not repeat the same ip addresses. Important for clickhouse that requires
	// a dns reset every time an item from the db gets removed.
	if err := c.dummyContainer(); err != nil {
		return err
	}
	return nil
}

func (c *Client) dummyContainer() error {
	config := &container.Config{
		Image: "busybox",
	}
	netConfig := &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			networkName: {},
		},
	}
	ctx := context.Background()
	body, err := c.client.ContainerCreate(ctx, config, &container.HostConfig{}, netConfig, uuid.UUID())
	if err != nil {
		return err
	}
	if err := c.client.ContainerStart(ctx, body.ID, types.ContainerStartOptions{}); err != nil {
		return err
	}
	return nil
}

func (c *Client) WatchUpdates() chan *proto.InstanceUpdate {
	return make(chan *proto.InstanceUpdate)
}

func (c *Client) runProvider() {

	fmt.Println("- run provider -")
	fmt.Println(c.controlPlane)

	ch := c.controlPlane.SubscribeInstanceUpdates()

	go func() {
		for {
			msg := <-ch
			if msg == nil {
				panic("bad")
			}

			//fmt.Println("-- docker msg --")
			//fmt.Println(msg)

			instance, err := c.controlPlane.GetInstance(msg.Id, msg.Cluster)
			if err != nil {
				panic(err)
			}

			//fmt.Println("-- docker instance --")
			//fmt.Println(instance)

			if instance.Status == proto.Instance_PENDING {
				fmt.Println("_ CREATE INSTANCE _", instance.ID, instance.Name, instance.Status)
				// we can work on this
				if _, err := c.createImpl(context.Background(), instance); err != nil {

					ii := instance.Copy()
					ii.Status = proto.Instance_FAILED

					if err := c.controlPlane.UpsertInstance(ii); err != nil {
						panic(err)
					}
				}

			} else if instance.Status == proto.Instance_TAINTED {
				fmt.Printf("_ DELETE INSTANCE: %s %s _\n", instance.ID, instance.Name)

				if err := c.client.ContainerKill(context.Background(), instance.Handler, "SIGINT"); err != nil {
					panic(err)
				}
				if err := c.client.ContainerStop(context.Background(), instance.Handler, nil); err != nil {
					panic(err)
				}
				fmt.Println("- done -")
			}
		}
	}()
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
	panic("CLEAN")

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

/*
func (c *Client) run() {
	worker := func(id string) {
		for instance := range c.workCh {
			if _, err := c.createImpl(context.Background(), instance); err != nil {
				panic(err)
			}
		}
	}
	numWorkers := 2
	for i := 0; i < numWorkers; i++ {
		go worker(strconv.Itoa(i))
	}
}
*/

// Create creates a docker container
func (c *Client) createImpl(ctx context.Context, node *proto.Instance) (string, error) {
	c.resourcesLock.Lock()
	defer c.resourcesLock.Unlock()

	// check if there is another container with the same name. Three possibilities
	// 1. The Provider did not fully removed the container after being removed by the scheduler.
	// 2. The Scheduler is trying to run another instance with the same name.
	// 3. There is a container from a previous execution that has not being purged.
	//
	// Since the docker provider keeps a local copy of all the valid running instances, cases 1 and 2
	// can be easily validated (we panic for now), otherwise its case 3 and we remove the name.

	filters := filters.NewArgs()
	filters.Add("name", node.Name)

	containers, err := c.client.ContainerList(ctx, types.ContainerListOptions{All: true, Filters: filters})
	if err != nil {
		return "", err
	}
	if size := len(containers); size != 0 {
		if size != 1 {
			return "", fmt.Errorf("more than one container expected")
		}

		prevInstance := c.findInstanceLocked(func(i *proto.Instance) bool {
			return i.Name == node.Name
		})
		if prevInstance != nil {
			// case 1 and 2
			panic(fmt.Errorf("instance with the same node name %s allocated twice: Now %s Before %s", node.Name, node.ID, prevInstance.ID))
		} else {
			// case 3
			if err := c.client.ContainerRemove(context.Background(), containers[0].ID, types.ContainerRemoveOptions{Force: true}); err != nil {
				return "", err
			}
		}
	}

	// We will use the 'net1' network interface for dns resolving
	builder := node.Spec

	version := node.Version
	if version == "" {
		version = "latest"
	}
	if node.Image == "" {
		return "", fmt.Errorf("node image empty")
	}
	image := node.Image + ":" + version

	name := node.FullName()

	binds := []string{}

	// mount files
	if len(builder.Files) != 0 {
		buildDir, err := ioutil.TempDir("/tmp", "builder-")
		if err != nil {
			return "", err
		}
		mountPoints, err := mount.CreateMountPoints(builder.Files)
		if err != nil {
			return "", err
		}
		for indx, mountPoint := range mountPoints {
			mountPath := filepath.Join(buildDir, fmt.Sprintf("%d", indx))
			for path, content := range mountPoint.Files {
				subPath := strings.TrimPrefix(path, mountPoint.Path)
				subPath = filepath.Join(mountPath, subPath)

				if err := createIfNotExists(filepath.Dir(subPath)); err != nil {
					return "", err
				}
				if err := ioutil.WriteFile(subPath, []byte(content), 0755); err != nil {
					return "", err
				}
			}
			binds = append(binds, fmt.Sprintf("%s:%s", mountPath, mountPoint.Path))
		}
	}

	// mount paths
	if len(node.Mounts) != 0 {
		dataDir := "/tmp/ensemble-" + node.ClusterName + "-" + node.Name
		for _, mount := range node.Mounts {
			localPath := dataDir + "-" + mount.Name
			if err := createIfNotExists(localPath); err != nil {
				return "", err
			}
			binds = append(binds, fmt.Sprintf("%s:%s", localPath, mount.Path))
		}
	}

	if err := c.PullImage(ctx, image); err != nil {
		return "", err
	}

	extraHostsMap := map[string]string{}
	for _, r := range c.resources {
		if r.instance.Ip == "" {
			continue
		}
		extraHostsMap[r.instance.ClusterName] = r.instance.Ip
	}
	extraHost := []string{}
	for k, v := range extraHostsMap {
		extraHost = append(extraHost, k+":"+v)
	}

	env := []string{}
	for k, v := range builder.Env {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	cmd := []string{}
	if builder.Cmd != "" {
		cmd = append(cmd, builder.Cmd)
	}
	cmd = append(cmd, builder.Args...)

	config := &container.Config{
		Hostname: name,
		Image:    image,
		Env:      env,
		Cmd:      strslice.StrSlice(cmd),
	}
	hostConfig := &container.HostConfig{
		Binds:      binds,
		ExtraHosts: extraHost,
	}

	netConfig := &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			networkName: {},
		},
	}

	body, err := c.client.ContainerCreate(ctx, config, hostConfig, netConfig, name)
	if err != nil {
		return "", err
	}

	c.resources[node.ID] = &resource{
		nodeID:    node.ID,
		handle:    body.ID,
		clusterID: node.ClusterName,
		instance:  node,
	}
	if err := c.client.ContainerStart(ctx, body.ID, types.ContainerStartOptions{}); err != nil {
		return "", err
	}

	// watch for updates in the node
	go func() {
		_, err := c.client.ContainerWait(context.Background(), body.ID)
		if err != nil {
			panic(err)
		}
		// we need to remove it here so that we can reuse the name
		if err := c.client.ContainerRemove(context.Background(), body.ID, types.ContainerRemoveOptions{}); err != nil {
			panic(err)
		}

		c.resources[node.ID].active = false

		// panic("bad")

		ii, err := c.controlPlane.GetInstance(node.ID, node.DeploymentID)
		if err != nil {
			panic(err)
		}
		// we need all the data stored here so we need to get the instance from db
		ii = ii.Copy()
		ii.Status = proto.Instance_FAILED

		if err := c.controlPlane.UpsertInstance(ii); err != nil {
			panic(err)
		}
	}()

	ip := c.GetIP(body.ID)

	nn := node.Copy()
	nn.Ip = ip
	nn.Handler = body.ID
	nn.Status = proto.Instance_RUNNING
	// nn.Healthy = true

	fmt.Printf("IP: %s %s\n", nn.Name, ip)

	if err := c.controlPlane.UpsertInstance(nn); err != nil {
		panic(err)
	}

	c.resources[node.ID].instance.Ip = ip
	return body.ID, nil
}

func (c *Client) Resources() operator.ProviderResources {
	return operator.ProviderResources{
		Node: schema.Schema2{
			Spec: &schema.Record{
				Fields: map[string]*schema.Field{
					"cpuShares": {
						Type:     schema.TypeInt,
						ForceNew: true,
					},
					"cpuCount": {
						Type:     schema.TypeInt,
						ForceNew: true,
					},
				},
			},
		},
	}
}

/*
func (c *Client) CreateResource(node *proto.Instance) (*proto.Instance, error) {
	// fmt.Printf("Create resource: %s %s\n", node.ID, node.Name)

	// validation
	for _, r := range c.resources {
		if r.instance.ID == node.ID {
			return nil, operator.ErrInstanceAlreadyRunning
		}
		if r.active {
			if r.instance.FullName() == node.FullName() {
				return nil, operator.ErrProviderNameAlreadyUsed
			}
		}
	}

	// async serialize the execution
	c.workCh <- node

	return nil, nil
}
*/

func (c *Client) Exec(id string, path string, args ...string) (string, error) {
	execCmd := []string{path}
	execCmd = append(execCmd, args...)

	var handler string
	for _, r := range c.resources {
		if r.instance.ID == id {
			handler = r.handle
		}
	}
	if handler == "" {
		return "", fmt.Errorf("not found")
	}

	return c.execImpl(context.Background(), handler, execCmd)
}

func (c *Client) execImpl(ctx context.Context, id string, execCmd []string) (string, error) {
	ec := types.ExecConfig{
		User:         "",
		AttachStderr: true,
		AttachStdout: true,
		Cmd:          execCmd,
	}
	execID, err := c.client.ContainerExecCreate(ctx, id, ec)
	if err != nil {
		return "", err
	}

	aresp, err := c.client.ContainerExecAttach(ctx, execID.ID, types.ExecConfig{})
	if err != nil {
		return "", err
	}
	defer aresp.Close()

	var outBuf, errBuf bytes.Buffer
	outputDone := make(chan error)

	go func() {
		_, err = stdcopy.StdCopy(&outBuf, &errBuf, aresp.Reader)
		outputDone <- err
	}()

	select {
	case err := <-outputDone:
		if err != nil {
			return "", err
		}
		break

	case <-ctx.Done():
		return "", ctx.Err()
	}

	// get the exit code
	if _, err := c.client.ContainerExecInspect(ctx, execID.ID); err != nil {
		return "", err
	}

	if errBuf.Len() != 0 {
		return "", fmt.Errorf(errBuf.String())
	}
	return outBuf.String(), nil
}

func (c *Client) findInstanceLocked(handler func(i *proto.Instance) bool) *proto.Instance {
	for _, i := range c.resources {
		if handler(i.instance) {
			return i.instance
		}
	}
	return nil
}

func (c *Client) DeleteResource(node *proto.Instance) (*proto.Instance, error) {
	// fmt.Printf("Delete resource: %s %s\n", node.Name, node.Handler)

	go func() {
		if err := c.Remove(node.Handler); err != nil {
			fmt.Printf("[ERR]: Failed to delete resource: %v\n", err)
		}
	}()
	return nil, nil
}

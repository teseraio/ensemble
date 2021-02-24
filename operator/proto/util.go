package proto

import (
	"bytes"
	"encoding/json"
	"strings"

	"google.golang.org/protobuf/proto"
)

const (
	DeploymentDone    = "done"
	DeploymentRunning = "running"
)

const (
	InstanceDesiredRunning = "running"
	InstanceDesiredStopped = "stopped"
)

func Equal(p0, p1 proto.Message) bool {
	m0, err := proto.Marshal(p0)
	if err != nil {
		panic("BAD")
	}
	m1, err := proto.Marshal(p1)
	if err != nil {
		panic("BAD")
	}
	return bytes.Equal(m0, m1)
}

/*
func (c *Cluster) Size() int {
	return len(c.Nodes)
}
*/

/*
func (c *Cluster) DelNodeAtIndx(i int) {
	c.Nodes = append(c.Nodes[:i], c.Nodes[i+1:]...)
}

func (c *Cluster) AddNode(n *Instance) {
	c.Nodes = append(c.Nodes, n)
}

func (c *Cluster) NodeByID(ID string) (*Node, bool) {
	for _, n := range c.Nodes {
		if n.ID == ID {
			return n, true
		}
	}
	return nil, false
}

func (c *Cluster) NodeAtIndex(ID string) int {
	for indx, n := range c.Nodes {
		if n.ID == ID {
			return indx
		}
	}
	return -1
}
*/

func (d *Deployment) Copy() *Deployment {
	return proto.Clone(d).(*Deployment)
}

/*
func (m *Node_Mount) Copy() *Node_Mount {
	return proto.Clone(m).(*Node_Mount)
}
*/

func (n *Instance) FullName() string {
	if n.Cluster != "" {
		return n.Name + "." + n.Cluster
	}
	return n.Name
}

func (n *Instance) Get(k string) string {
	v, _ := n.GetOk(k)
	return v
}

func (n *Instance) GetOk(k string) (string, bool) {
	v, ok := n.KV[k]
	return v, ok
}

func (n *Instance) Set(k, v string) {
	if n.KV == nil {
		n.KV = map[string]string{}
	}
	n.KV[k] = v
}

/*
func (n *Instance) Equal(nn *Instance) bool {
	// check the state
	if n.State != nn.State {
		return false
	}
	// check the kv store
	if !reflect.DeepEqual(n.KV, nn.KV) {
		// TODO: Do better than this
		if len(n.KV) == len(nn.KV) && len(n.KV) == 0 {
			return true
		}
		return false
	}
	// check the mounts
	if !reflect.DeepEqual(n.Mounts, nn.Mounts) {
		return false
	}
	// check spec
	if !reflect.DeepEqual(n.Spec, nn.Spec) {
		return false
	}
	return true
}
*/

// TODO: Use protobuf for this
func (n *Instance) Unmarshal(src []byte) error {
	return json.Unmarshal(src, &n)
}

func (n *Instance) Marshal() ([]byte, error) {
	return json.Marshal(n)
}

func (n *Instance) Copy() *Instance {
	return proto.Clone(n).(*Instance)
}

func (b *NodeSpec) AddFile(path string, content string) {
	if b.Files == nil {
		b.Files = map[string]string{}
	}
	b.Files[path] = content
}

func (b *NodeSpec) AddEnvList(l []string) {
	for _, i := range l {
		indx := strings.Index(i, "=")
		if indx == -1 {
			panic("BUG")
		}
		b.AddEnv(i[:indx], i[indx+1:])
	}
}

func (b *NodeSpec) AddEnvMap(m map[string]string) {
	for k, v := range m {
		b.AddEnv(k, v)
	}
}

func (b *NodeSpec) AddEnv(k, v string) {
	if b.Env == nil {
		b.Env = map[string]string{}
	}
	b.Env[k] = v
}

func (b *NodeSpec) Copy() *NodeSpec {
	return proto.Clone(b).(*NodeSpec)
}

/*
func (t *Task) Time() (time.Time, error) {
	return ptypes.Timestamp(t.Timestamp)
}

type XX struct {
	Replicas int64
	Config   string
	Resource string
}
*/

/*
func (p *Plan_Set) Add(n *Instance) {
	if p.AddNodes == nil {
		p.AddNodes = make([]*Node, 0)
	}
	p.AddNodes = append(p.AddNodes, n)
}
*/

func (r *ClusterSpec2) GetClusterID() string {
	return r.Name
}

func (r *ResourceSpec) GetClusterID() string {
	return r.Cluster
}

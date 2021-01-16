package proto

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"

	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
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

func (c *Cluster) Size() int {
	return len(c.Nodes)
}

func (c *Cluster) NewNode() *Node {
	return &Node{
		ID:      uuid.New().String(),
		State:   Node_UNKNOWN,
		Cluster: c.Name,
		Spec:    &Node_NodeSpec{},
	}
}

func (c *Cluster) DelNodeAtIndx(i int) {
	c.Nodes = append(c.Nodes[:i], c.Nodes[i+1:]...)
}

func (c *Cluster) AddNode(n *Node) {
	c.Nodes = append(c.Nodes, n)
}

func (c *Cluster) NodeAtIndex(ID string) int {
	for indx, n := range c.Nodes {
		if n.ID == ID {
			return indx
		}
	}
	return -1
}

func (c *Cluster) Copy() *Cluster {
	return proto.Clone(c).(*Cluster)
}

func (m *Node_Mount) Copy() *Node_Mount {
	return proto.Clone(m).(*Node_Mount)
}

func (n *Node) FullName() string {
	if n.Cluster != "" {
		return n.ID + "." + n.Cluster
	}
	return n.ID
}

func (n *Node) Get(k string) string {
	v, _ := n.GetOk(k)
	return v
}

func (n *Node) GetOk(k string) (string, bool) {
	v, ok := n.KV[k]
	return v, ok
}

func (n *Node) Set(k, v string) {
	if n.KV == nil {
		n.KV = map[string]string{}
	}
	n.KV[k] = v
}

func (n *Node) Equal(nn *Node) bool {
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

// TODO: Use protobuf for this
func (n *Node) Unmarshal(src []byte) error {
	return json.Unmarshal(src, &n)
}

func (n *Node) Marshal() ([]byte, error) {
	return json.Marshal(n)
}

func (n *Node) Copy() *Node {
	return proto.Clone(n).(*Node)
}

func (b *Node_NodeSpec) AddFile(path string, content string) {
	if b.Files == nil {
		b.Files = map[string]string{}
	}
	b.Files[path] = content
}

func (b *Node_NodeSpec) AddEnvList(l []string) {
	for _, i := range l {
		indx := strings.Index(i, "=")
		if indx == -1 {
			panic("BUG")
		}
		b.AddEnv(i[:indx], i[indx+1:])
	}
}

func (b *Node_NodeSpec) AddEnvMap(m map[string]string) {
	for k, v := range m {
		b.AddEnv(k, v)
	}
}

func (b *Node_NodeSpec) AddEnv(k, v string) {
	if b.Env == nil {
		b.Env = map[string]string{}
	}
	b.Env[k] = v
}

func (b *Node_NodeSpec) Copy() *Node_NodeSpec {
	return proto.Clone(b).(*Node_NodeSpec)
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

func (p *Plan) Add(n *Node) {
	p.AddNodes = append(p.AddNodes, n)
}

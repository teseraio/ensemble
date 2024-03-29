package proto

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"google.golang.org/protobuf/proto"
)

func (c *Component) IsDone() bool {
	return c.Status == Component_APPLIED
}

const (
	PlanStatusDone = "done"
)

type ClusterRef interface {
	GetCluster() string
}

func (c *ClusterSpec) GetCluster() string {
	return ""
}

const (
	EvaluationTypeCluster  = "cluster"
	EvaluationTypeResource = "resource"
)

const (
	DeploymentDone      = "done"
	DeploymentRunning   = "running"
	DeploymentCompleted = "completed"
	DeploymentFailed    = "failed"
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

func (c *Component) Copy() *Component {
	return proto.Clone(c).(*Component)
}

func (c *ClusterSpec) Copy() *ClusterSpec {
	return proto.Clone(c).(*ClusterSpec)
}

func (c *ClusterSpec_Group) Copy() *ClusterSpec_Group {
	return proto.Clone(c).(*ClusterSpec_Group)
}

func (d *Deployment) Copy() *Deployment {
	return proto.Clone(d).(*Deployment)
}

func (n *Instance) FullName() string {
	if n.ClusterName != "" {
		//if n.DnsSuffix != "" {
		//	return n.Name + "." + n.ClusterName + n.DnsSuffix
		//}
		return n.Name + "." + n.ClusterName
	}
	return n.Name
}

var okKey = "ok"

func (n *Instance) SetTrue(k string) {
	n.Set(k, okKey)
}

func (n *Instance) GetTrue(k string) bool {
	return n.Get(k) == "ok"
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

func (n *Instance) GetInt(k string) (int, error) {
	raw := n.Get(k)
	return strconv.Atoi(raw)
}

func (n *Instance) SetInt(k string, i int) {
	n.Set(k, fmt.Sprintf("%d", i))
}

func (n *Instance) Unmarshal(src []byte) error {
	return json.Unmarshal(src, &n)
}

func (n *Instance) Marshal() ([]byte, error) {
	return json.Marshal(n)
}

func (n *Instance) Copy() *Instance {
	return proto.Clone(n).(*Instance)
}

func (n *Instance) IsHealthy() bool {
	return n.Status == Instance_RUNNING && n.Healthy
}

func (e *Instance_ExitResult) Failed() bool {
	return e.Code != 0
}

func (e *Instance_ExitResult) Complete() bool {
	return e.Code == 0
}

func (b *NodeSpec) AddFile(path string, content string) {
	if b.Files == nil {
		b.Files = []*NodeSpec_File{}
	}
	b.Files = append(b.Files, &NodeSpec_File{
		Name:    path,
		Content: content,
	})
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
func (r *ClusterSpec) GetClusterID() string {
	return r.Name
}
*/

func (r *ResourceSpec) GetClusterID() string {
	return r.Cluster
}

type clusterItem interface {
	proto.Message
	GetClusterID() string
}

var specs = map[string]clusterItem{
	"proto.ResourceSpec": &ResourceSpec{},
}

func ClusterIDFromComponent(c *Component) (string, error) {
	var clusterID string
	if c.Spec.TypeUrl == "proto.ClusterSpec" {
		// the name of the component is the id of the cluster
		clusterID = c.Name
	} else {
		item, ok := specs[c.Spec.TypeUrl]
		if !ok {
			return "", fmt.Errorf("bad")
		}
		if err := proto.Unmarshal(c.Spec.Value, item); err != nil {
			return "", err
		}
		clusterID = item.GetClusterID()
	}
	return clusterID, nil
}

func ParseIndex(n string) (uint64, error) {
	parts := strings.Split(n, "-")
	if len(parts) != 2 && len(parts) != 3 {
		return 0, fmt.Errorf("wrong number of parts")
	}

	// the index is always the last element
	indexStr := parts[len(parts)-1]
	index, err := strconv.Atoi(indexStr)
	if err != nil {
		return 0, err
	}
	return uint64(index), nil
}

func BlockSpec(block *Spec_Block) *Spec {
	return &Spec{
		Block: &Spec_BlockValue{
			BlockValue: block,
		},
	}
}

// LiteralSpec wraps the literal and returns a spec.
func LiteralSpec(l *Spec_Literal) *Spec {
	return &Spec{
		Block: &Spec_Literal_{
			Literal: l,
		},
	}
}

func ArraySpec(values []*Spec) *Spec {
	return &Spec{
		Block: &Spec_Array_{
			Array: &Spec_Array{
				Values: values,
			},
		},
	}
}

func EmptySpec() *Spec {
	return &Spec{
		Block: &Spec_BlockValue{
			BlockValue: &Spec_Block{},
		},
	}
}

/// -- instance

func (i *Instance) Update(event isInstanceUpdate_Event) *InstanceUpdate {
	return &InstanceUpdate{
		ID:          i.ID,
		ClusterName: i.ClusterName,
		Event:       event,
	}
}

/// -- deployment functional

func (d *Deployment) Filter(filter func(n *Instance) bool) (res []*Instance) {
	for _, i := range d.Instances {
		if filter(i) {
			res = append(res, i)
		}
	}
	return res
}

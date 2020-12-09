package k8s

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/teseraio/ensemble/operator/proto"
	"github.com/teseraio/ensemble/schema"
)

// Pod is a type to create a k8s pod
type Pod struct {
	Name     string
	Ensemble string
	Builder  *proto.NodeSpec
}

type volumeMount struct {
	// Name is the name of the volume in the pod description
	Name string

	// Path is the path of the volume in the pod
	Path string

	// Items are specific files to mount from the config file
	Items map[string]string
}

func cleanPath(path string) string {
	// clean the path to k8s format
	return strings.Replace(strings.Trim(path, "/"), "/", ".", -1)
}

func convertFiles(paths []string) *volumeMount {
	// this function takes a group of paths and defines
	// how many mounts it has to create and the specific
	// volumes assuming that all the contents are from the same
	// config map

	if len(paths) > 1 {
		panic("We only do now one single file sorry")
	}

	pp := paths[0]

	dir, file := filepath.Split(pp)
	v := &volumeMount{
		Name: "config", // same by default, change when more than one, prefix with config
		Path: dir,
		Items: map[string]string{
			// we need to use the full path in k8s format to reference it
			cleanPath(pp): file,
		},
	}

	return v
}

// MarshalPod marshals a pod
func MarshalPod(p *Pod) ([]byte, error) {
	if p.Ensemble == "" {
		return nil, fmt.Errorf("ensemble not defined")
	}

	obj := map[string]interface{}{
		"Name":     p.Name,
		"Image":    p.Builder.Image,
		"Version":  p.Builder.Version,
		"Env":      p.Builder.Env,
		"Files":    p.Builder.Files,
		"Ensemble": p.Ensemble,
	}

	if len(p.Builder.Files) > 0 {
		paths := []string{}
		for k := range p.Builder.Files {
			paths = append(paths, k)
		}
		v := convertFiles(paths)
		obj["Volume"] = v
	}

	if len(p.Builder.Cmd) != 0 {
		obj["Command"] = "'" + strings.Join(p.Builder.Cmd, "', '") + "'"
	}

	return RunTmpl2("pod", obj)
}

func decodeInt(c map[string]interface{}, k string) (int, error) {
	raw, ok := c[k]
	if !ok {
		return 0, fmt.Errorf("key '%s' not found", k)
	}
	rawInt, ok := raw.(float64)
	if !ok {
		return 0, fmt.Errorf("key is not of type int but %s", reflect.TypeOf(raw).String())
	}
	return int(rawInt), nil
}

func decodeString(c map[string]interface{}, k string) (string, error) {
	raw, ok := c[k]
	if !ok {
		return "", fmt.Errorf("key '%s' not found", k)
	}
	rawStr, ok := raw.(string)
	if !ok {
		return "", fmt.Errorf("key is not of type string but %s", reflect.TypeOf(rawStr).String())
	}
	return rawStr, nil
}

type kvEntry struct {
	Key string `json:"key"`
	Val string `json:"val"`
}

func marshalNode(node *proto.Node) ([]byte, error) {
	isUpdated := node.ResourceVersion != ""

	kvs := []*kvEntry{}
	for k, v := range node.KV {
		kvs = append(kvs, &kvEntry{
			Key: k,
			Val: v,
		})
	}

	mounts := []*proto.Mount{}
	if node.Mounts != nil {
		mounts = node.Mounts
	}
	obj := map[string]interface{}{
		"Domain": "ensembleoss.io/v1",
		"Kind":   "Node",
		"Name":   node.ID,
		"Labels": map[string]interface{}{
			"ensemble": node.Cluster,
		},
		"Metadata": map[string]interface{}{
			"ensemble": node.Cluster,
		},
		"Spec": map[string]interface{}{
			"id":       node.ID,
			"nodeset":  node.Nodeset,
			"nodetype": node.Nodetype,
			"spec":     node.Spec,
			"mounts":   mounts,
		},
		// The status part will only be included when calling the /status endpoint
		"Status": map[string]interface{}{
			"podIP":  node.Addr,
			"handle": node.Handle,
			"status": proto.NodeState_name[int32(node.State)],
			"kv":     kvs,
		},
	}
	if isUpdated {
		obj["ResourceVersion"] = node.ResourceVersion
	}

	res, err := RunTmpl2("generic", obj)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func unmarshalNodeBytes(resp []byte) (*proto.Node, *Metadata, error) {
	var item *Item
	if err := json.Unmarshal(resp, &item); err != nil {
		return nil, nil, err
	}
	node, err := unmarshalNode(item)
	if err != nil {
		return nil, nil, err
	}
	return node, item.Metadata, nil
}

func unmarshalNode(i *Item) (*proto.Node, error) {
	n := new(proto.Node)

	ensemble, ok := i.Metadata.Labels["ensemble"]
	if !ok {
		return nil, fmt.Errorf("ensemble not found")
	}

	mnts, ok := i.Spec["mounts"]
	if ok {
		if err := mapstructure.Decode(mnts, &n.Mounts); err != nil {
			return nil, err
		}
	}

	specRaw, ok := i.Spec["spec"]
	if !ok {
		return nil, fmt.Errorf("spec not found")
	}
	if err := mapstructure.Decode(specRaw, &n.Spec); err != nil {
		return nil, err
	}

	var err error
	if n.Nodeset, err = decodeString(i.Spec, "nodeset"); err != nil {
		return nil, err
	}
	if n.Nodetype, err = decodeString(i.Spec, "nodetype"); err != nil {
		return nil, err
	}

	if len(i.Status) != 0 {
		if n.Handle, err = decodeString(i.Status, "handle"); err != nil {
			return nil, err
		}
		if n.Addr, err = decodeString(i.Status, "podIP"); err != nil {
			return nil, err
		}
		statusStr, err := decodeString(i.Status, "status")
		if err != nil {
			return nil, err
		}
		n.State = proto.NodeState(proto.NodeState_value[statusStr])

		var kvEntries []*kvEntry
		kv, ok := i.Status["kv"]
		if ok {
			if err := mapstructure.Decode(kv, &kvEntries); err != nil {
				return nil, err
			}
		}
		if len(kvEntries) != 0 {
			n.KV = map[string]string{}
			for _, i := range kvEntries {
				n.KV[i.Key] = i.Val
			}
		}
	}

	n.Cluster = ensemble
	n.ID = i.Metadata.Name
	n.ResourceVersion = i.Metadata.ResourceVersion
	return n, nil
}

// Crd is a k8s custom resource definition
type Crd struct {
	Kind     string
	Singular string
	Plural   string
	Group    string

	Schema  *schema.Schema
	SpecStr string
}

// MarshalCRD marshals a CRD object
func MarshalCRD(c *Crd) ([]byte, error) {
	if !strings.Contains(c.Group, ".") {
		return nil, fmt.Errorf("group domain needs at least one dot")
	}

	if c.Singular == "" {
		c.Singular = strings.ToLower(c.Kind)
	}
	if c.Plural == "" {
		c.Plural = c.Singular + "s"
	}

	if c.SpecStr == "" {
		spec, err := c.Schema.Spec.OpenAPIV3JSON()
		if err != nil {
			return nil, err
		}
		c.SpecStr = string(spec)
	}

	buf, err := RunTmpl2("crd", c)
	if err != nil {
		return nil, err
	}
	return buf, nil
}

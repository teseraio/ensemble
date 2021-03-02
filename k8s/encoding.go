package k8s

import (
	"path/filepath"
	"strings"

	"github.com/teseraio/ensemble/operator/proto"
)

/*
// Pod is a type to create a k8s pod
type Pod struct {
	Name     string
	Ensemble string
	Builder  *proto.NodeSpec
}
*/

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
func MarshalPod(i *proto.Instance) ([]byte, error) {
	builder := i.Spec

	obj := map[string]interface{}{
		"Name":     i.Name,
		"Image":    builder.Image,
		"Version":  builder.Version,
		"Env":      builder.Env,
		"Files":    builder.Files,
		"Ensemble": i.Cluster,
		"Hostname": i.Name,
	}

	if num := len(builder.Cmd); num != 0 {
		obj["Command"] = builder.Cmd[0]
		if num > 1 {
			obj["Args"] = builder.Cmd[1:]
		}
	}

	if len(builder.Files) > 0 {
		paths := []string{}
		for k := range builder.Files {
			paths = append(paths, k)
		}
		v := convertFiles(paths)
		obj["Volume"] = v
	}

	return RunTmpl2("pod", obj)
}

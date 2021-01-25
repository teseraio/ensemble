package k8s

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/teseraio/ensemble/operator/proto"
)

// Pod is a type to create a k8s pod
type Pod struct {
	Name     string
	Ensemble string
	Builder  *proto.Node_NodeSpec
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

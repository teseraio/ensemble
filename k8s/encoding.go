package k8s

import (
	"path/filepath"
	"strings"

	"github.com/teseraio/ensemble/operator/proto"
)

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
func MarshalPod(node *proto.Node) ([]byte, error) {
	// all the mount points for the pod
	type mount struct {
		Name     string
		Path     string
		ReadOnly bool
	}
	var volMounts []*mount

	// list of all the volumes in the pod
	var volumes []interface{}

	obj := map[string]interface{}{
		"Name":     node.ID,
		"Image":    node.Spec.Image,
		"Version":  node.Spec.Version,
		"Env":      node.Spec.Env,
		"Files":    node.Spec.Files,
		"Ensemble": node.Cluster,
	}

	// add the persistent volumes
	for _, m := range node.Mounts {
		volMounts = append(volMounts, &mount{
			Name:     m.Name,
			Path:     m.Path,
			ReadOnly: false,
		})
		volumes = append(volumes, map[string]interface{}{
			"name": m.Name,
			"persistentVolumeClaim": map[string]interface{}{
				"claimName": node.ID + "-" + m.Name,
			},
		})
	}

	// add mount volumes
	obj["Volumes"] = volumes

	// add mount points
	obj["VolumeMounts"] = volMounts

	if len(node.Spec.Cmd) != 0 {
		obj["Command"] = "'" + strings.Join(node.Spec.Cmd, "', '") + "'"
	}

	return RunTmpl2("pod", obj)
}

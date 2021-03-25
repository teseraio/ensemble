package k8s

import (
	"strconv"
	"strings"

	"github.com/teseraio/ensemble/lib/mount"
	"github.com/teseraio/ensemble/operator/proto"
)

// MarshalPod marshals a pod
func MarshalPod(i *proto.Instance) ([]byte, error) {
	builder := i.Spec

	version := builder.Version
	if version == "" {
		version = "latest"
	}

	type mountPoint struct {
		Name     string
		Path     string
		ReadOnly bool
	}
	var volMounts []*mountPoint

	// list of all the volumes in the pod
	var volumes []interface{}

	obj := map[string]interface{}{
		"ID":       i.ID,
		"Name":     i.Name,
		"Image":    builder.Image,
		"Version":  version,
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

	// add the persistent volumes
	for _, m := range i.Mounts {
		volMounts = append(volMounts, &mountPoint{
			Name:     m.Name,
			Path:     m.Path,
			ReadOnly: false,
		})
		volumes = append(volumes, map[string]interface{}{
			"name": m.Name,
			"persistentVolumeClaim": map[string]interface{}{
				"claimName": i.Name + "-" + m.Name,
			},
		})
	}

	if len(builder.Files) > 0 {
		mountPoints, err := mount.CreateMountPoints(builder.Files)
		if err != nil {
			return nil, err
		}
		for indx, pnt := range mountPoints {
			name := "file-data-" + strconv.Itoa(indx)

			// create the mount point
			volMounts = append(volMounts, &mountPoint{
				Name: name,
				Path: pnt.Path,
			})

			// we need to mount all the files to the specific locations
			items := []interface{}{}
			for name := range pnt.Files {
				relPath := strings.TrimPrefix(name, pnt.Path)
				relPath = strings.TrimPrefix(relPath, "/")

				items = append(items, map[string]interface{}{
					"key":  cleanPath(name),
					"path": relPath,
				})
			}

			// create the volume as a config reference
			volumes = append(volumes, map[string]interface{}{
				"name": name,
				"configMap": map[string]interface{}{
					"name":  i.ID + "-" + name,
					"items": items,
				},
			})
		}
	}

	// add mount volumes
	obj["Volumes"] = volumes

	// add mount points
	obj["VolumeMounts"] = volMounts

	return RunTmpl2("pod", obj)
}

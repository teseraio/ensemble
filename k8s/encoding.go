package k8s

import (
	"sort"
	"strconv"
	"strings"

	"github.com/teseraio/ensemble/lib/mount"
	"github.com/teseraio/ensemble/operator/proto"
)

type mountFile struct {
	Key  string `json:"key"`
	Path string `json:"path"`
}

type mountFiles []*mountFile

func (m mountFiles) Len() int {
	return len(m)
}

func (m mountFiles) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}

func (m mountFiles) Less(i, j int) bool {
	return m[i].Key < m[j].Key
}

// MarshalPod marshals a pod
func MarshalPod(i *proto.Instance) ([]byte, error) {
	builder := i.Spec

	version := i.Version
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
		"Image":    i.Image,
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
			items := mountFiles{}
			for name := range pnt.Files {
				relPath := strings.TrimPrefix(name, pnt.Path)
				relPath = strings.TrimPrefix(relPath, "/")

				items = append(items, &mountFile{
					Key:  cleanPath(name),
					Path: relPath,
				})
			}
			sort.Sort(items)

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

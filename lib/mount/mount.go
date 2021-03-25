package mount

import (
	"strings"

	"github.com/teseraio/ensemble/operator/proto"
)

type MountPoint struct {
	Path  string
	Files map[string]string
}

func CreateMountPoints(files []*proto.NodeSpec_File) ([]*MountPoint, error) {
	groups := []*MountPoint{}
	for _, file := range files {
		name := file.Name

		found := false
		for _, grp := range groups {
			prefix, ok := getPrefix(grp.Path, name)
			if ok {
				found = true
				// replace the group
				grp.Path = prefix
				grp.Files[name] = file.Content
				break
			}
		}
		if !found {
			// get absolute path
			groups = append(groups, &MountPoint{
				Path: getAbs(name),
				Files: map[string]string{
					name: file.Content,
				},
			})
		}
	}
	return groups, nil
}

func getAbs(path string) string {
	spl := strings.Split(path, "/")
	name := spl[:len(spl)-1]
	return strings.Join(name, "/")
}

func getPrefix(a, b string) (string, bool) {
	aSpl := strings.Split(a, "/")
	bSpl := strings.Split(b, "/")

	size := len(aSpl)
	if size > len(bSpl) {
		size = len(bSpl)
	}

	prefix := []string{}
	for i := 0; i < size; i++ {
		if aSpl[i] == bSpl[i] {
			prefix = append(prefix, aSpl[i])
		}
	}
	return strings.Join(prefix, "/"), len(prefix) != 1
}

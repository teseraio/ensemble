package artifacts

import (
	"strings"

	"github.com/teseraio/helm-template-engine/loader"
)

func GetArtifacts() (files []*loader.BufferedFile) {
	files = []*loader.BufferedFile{}

	prefix := "../charts/operator/"
	for _, name := range AssetNames() {
		files = append(files, &loader.BufferedFile{
			Name: strings.TrimPrefix(name, prefix),
			Data: MustAsset(name),
		})
	}
	return
}

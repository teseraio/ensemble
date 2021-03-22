package k8s

import (
	"fmt"

	"github.com/teseraio/ensemble/lib/template"
)

//go:generate go-bindata -pkg k8s -o ./bindata.go ./resources

func RunTmpl2(name string, obj interface{}) ([]byte, error) {
	content, err := Asset(fmt.Sprintf("resources/%s.template", name))
	if err != nil {
		return nil, err
	}
	return template.RunTmpl(string(content), obj)
}

package operator

import (
	"fmt"

	"github.com/teseraio/ensemble/lib/template"
	"github.com/teseraio/ensemble/operator/proto"
)

//go:generate go-bindata -pkg operator -o ./bindata.go ./resources

// Operator listens for new events in K8s and calls the Ensemble GRPC protocol
type Operator struct {
	EnsembleClient proto.EnsembleServiceClient
}

// Webhook validation!!

func runTmpl(name string, obj interface{}) ([]byte, error) {
	return template.RunTmpl(string(MustAsset(fmt.Sprintf("resources/%s.template", name))), obj)
}

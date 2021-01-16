package k8s

import (
	"encoding/json"

	"github.com/golang/protobuf/ptypes/any"
	"github.com/mitchellh/mapstructure"
	"github.com/teseraio/ensemble/operator/proto"
)

func decodeClusterSpec(item *Item) (*any.Any, error) {
	// it should correspond to the crd-cluster.json spec
	var spec struct {
		Replicas int64
		Backend  struct {
			Name string
		}
	}
	if err := mapstructure.Decode(item.Spec, &spec); err != nil {
		return nil, err
	}
	res := proto.MustMarshalAny(&proto.ClusterSpec{
		Backend:  spec.Backend.Name,
		Replicas: spec.Replicas,
	})
	return res, nil
}

func decodeResourceSpec(item *Item) (*any.Any, error) {
	// it should correspond to the crd-resource.json spec
	var spec struct {
		Backend  string
		Cluster  string
		Resource string
		Params   map[string]interface{}
	}
	if err := mapstructure.Decode(item.Spec, &spec); err != nil {
		return nil, err
	}

	raw, err := json.Marshal(spec.Params)
	if err != nil {
		return nil, err
	}
	res := proto.MustMarshalAny(&proto.ResourceSpec{
		Backend:  spec.Backend,
		Cluster:  spec.Cluster,
		Resource: spec.Resource,
		Params:   string(raw),
	})
	return res, nil
}

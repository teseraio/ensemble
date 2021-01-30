package k8s

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/golang/protobuf/ptypes/any"
	"github.com/mitchellh/mapstructure"
	"github.com/teseraio/ensemble/operator/proto"
)

const (
	// clustersURL is the k8s url for the cluster objects
	clustersURL = "/apis/ensembleoss.io/v1/namespaces/default/clusters"

	// resourcesURL is the k8s url for the resource objects
	resourcesURL = "/apis/ensembleoss.io/v1/namespaces/default/resources"
)

func (p *Provider) trackCRDs(clt proto.EnsembleServiceClient) {
	store := newStore()
	newWatcher(store, p.client, clustersURL)
	newWatcher(store, p.client, resourcesURL)

	for {
		task := store.pop(context.Background())
		item := task.item

		spec, err := decodeItem(item)
		if err != nil {
			panic(err)
		}

		c := &proto.Component{
			Name:     item.Metadata.Name,
			Spec:     spec,
			Metadata: item.Metadata.Labels,
		}
		if _, err := clt.Apply(context.Background(), c); err != nil {
			panic(err)
		}
	}
}

func decodeItem(item *Item) (*any.Any, error) {
	if item.Kind == "Cluster" {
		return decodeClusterSpec(item)
	}
	if item.Kind == "Resource" {
		return decodeResourceSpec(item)
	}
	return nil, fmt.Errorf("unknown type %s", item.Kind)
}

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
		Backend: spec.Backend.Name,
		Sets:    []*proto.ClusterSpec_Set{},
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
	err := mapstructure.Decode(item.Spec, &spec)
	if err != nil {
		return nil, err
	}

	var raw []byte
	if len(spec.Params) != 0 {
		raw, err = json.Marshal(spec.Params)
		if err != nil {
			return nil, err
		}
	}
	res := proto.MustMarshalAny(&proto.ResourceSpec{
		Backend:  spec.Backend,
		Cluster:  spec.Cluster,
		Resource: spec.Resource,
		Params:   string(raw),
	})
	return res, nil
}

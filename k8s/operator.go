package k8s

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/golang/protobuf/ptypes/any"
	"github.com/mitchellh/mapstructure"
	"github.com/teseraio/ensemble/operator/proto"
)

type APIGroupList struct {
	Groups []*APIGroup
}

func (a *APIGroupList) findGroup(name string) bool {
	for _, grp := range a.Groups {
		if grp.Name == name {
			return true
		}
	}
	return false
}

type APIGroup struct {
	Name string
}

type APIResourceList struct {
	Resources []*APIResource
}

func (a *APIResourceList) findResource(name string) bool {
	for _, res := range a.Resources {
		if res.Name == name {
			return true
		}
	}
	return false
}

type APIResource struct {
	Name string
}

var validResources = []string{
	"clusters",
	"resources",
}

func (p *Provider) trackCRDs(clt proto.EnsembleServiceClient) error {
	// check that the ensembleoss group exists
	var apiGroupList APIGroupList
	if _, err := p.get("/apis", &apiGroupList); err != nil {
		return err
	}
	if !apiGroupList.findGroup("ensembleoss.io") {
		return fmt.Errorf("ensemble group not found")
	}

	var apiResourceList APIResourceList
	if _, err := p.get("/apis/ensembleoss.io/v1", &apiResourceList); err != nil {
		return err
	}

	store := newStore()
	for _, res := range validResources {
		ok := apiResourceList.findResource(res)
		if ok {
			newWatcher(store, p.client, "/apis/ensembleoss.io/v1/namespaces/default/"+res, &Item{}, true)
			p.logger.Info("CRD tracker started", "name", res)
		} else {
			p.logger.Warn("CRD resource not defined", "name", res)
		}
	}

	go func() {
		for {
			task := store.pop(context.Background())
			item := task.item.(*Item)

			spec, err := decodeItem(item)
			if err != nil {
				panic(err)
			}

			action := proto.Component_CREATE
			if task.typ == "DELETED" {
				action = proto.Component_DELETE
			}

			c := &proto.Component{
				Name:     item.Metadata.Name,
				Spec:     spec,
				Metadata: item.Metadata.Labels,
				Action:   action,
			}

			p.logger.Debug("apply component", "name", item.Metadata.Name, "kind", item.Kind, "action", task.typ)
			if _, err := clt.Apply(context.Background(), c); err != nil {
				panic(err)
			}
		}
	}()

	return nil
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
		Backend struct {
			Name string
		}
		Groups []struct {
			Type     string
			Name     string
			Replicas uint64
			Params   map[string]string
		}
	}
	if err := mapstructure.Decode(item.Spec, &spec); err != nil {
		return nil, err
	}

	var groups []*proto.ClusterSpec_Group
	for _, s := range spec.Groups {
		groups = append(groups, &proto.ClusterSpec_Group{
			Count:  int64(s.Replicas),
			Type:   s.Type,
			Params: s.Params,
		})
	}
	res := proto.MustMarshalAny(&proto.ClusterSpec{
		Backend: spec.Backend.Name,
		Groups:  groups,
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
		Cluster:  spec.Cluster,
		Resource: spec.Resource,
		Params:   string(raw),
	})
	return res, nil
}

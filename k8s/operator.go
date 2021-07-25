package k8s

import (
	"context"
	"fmt"

	"github.com/golang/protobuf/ptypes/any"
	"github.com/mitchellh/mapstructure"
	"github.com/teseraio/ensemble/operator/proto"
	"github.com/teseraio/ensemble/schema"
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

	crdHandler := func(task *WatchEntry, i interface{}) {
		item := i.(*Item)

		spec, err := DecodeItem(item)
		if err != nil {
			p.logger.Error("failed to decode watch item", "err", err)
			return
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
			p.logger.Error("failed to apply component", "err", err)
		}
	}

	for _, res := range validResources {
		ok := apiResourceList.findResource(res)
		if ok {
			w, err := NewWatcher(p.logger, p.client, "/apis/ensembleoss.io/v1/namespaces/default/"+res, &Item{})
			if err != nil {
				return err
			}
			w.WithList().Run(p.stopCh)
			w.ForEach(crdHandler)

			p.logger.Info("CRD tracker started", "name", res)
		} else {
			p.logger.Warn("CRD resource not defined", "name", res)
		}
	}

	return nil
}

func DecodeItem(item *Item) (*any.Any, error) {
	if item.Kind == "Cluster" {
		return DecodeClusterSpec(item)
	}
	if item.Kind == "Resource" {
		return DecodeResourceSpec(item)
	}
	return nil, fmt.Errorf("unknown type %s", item.Kind)
}

func DecodeClusterSpec(item *Item) (*any.Any, error) {
	// it should correspond to the crd-cluster.json spec
	var spec struct {
		Backend struct {
			Name string
		}
		Groups []struct {
			Type     string
			Name     string
			Replicas uint64
			Params   map[string]interface{}
		}
		Depends []string
	}
	if err := mapstructure.Decode(item.Spec, &spec); err != nil {
		return nil, err
	}

	var groups []*proto.ClusterSpec_Group
	for _, s := range spec.Groups {
		grp := &proto.ClusterSpec_Group{
			Count: int64(s.Replicas),
			Type:  s.Type,
		}
		if len(s.Params) != 0 {
			grp.Params = schema.MapToSpec(s.Params)
		}
		groups = append(groups, grp)
	}
	res := proto.MustMarshalAny(&proto.ClusterSpec{
		Backend:   spec.Backend.Name,
		Groups:    groups,
		DependsOn: spec.Depends,
	})
	return res, nil
}

func DecodeResourceSpec(item *Item) (*any.Any, error) {
	// it should correspond to the crd-resource.json spec
	var spec struct {
		Cluster  string
		Resource string
		Params   map[string]interface{}
	}
	err := mapstructure.Decode(item.Spec, &spec)
	if err != nil {
		return nil, err
	}
	res := &proto.ResourceSpec{
		Cluster:  spec.Cluster,
		Resource: spec.Resource,
	}
	if len(spec.Params) != 0 {
		res.Params = schema.MapToSpec(spec.Params)
	} else {
		res.Params = proto.EmptySpec()
	}
	anyRes := proto.MustMarshalAny(res)
	return anyRes, nil
}

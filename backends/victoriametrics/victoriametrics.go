package clickhouse

import (
	"fmt"
	"strings"

	gproto "github.com/golang/protobuf/proto"
	"github.com/teseraio/ensemble/operator"
	"github.com/teseraio/ensemble/operator/proto"
	"github.com/teseraio/ensemble/schema"
)

type backend struct {
	*operator.BaseOperator
}

// Factory is a factory method for the zookeeper backend
func Factory() operator.Handler {
	b := &backend{}
	b.BaseOperator = &operator.BaseOperator{}
	b.BaseOperator.SetHandler(b)
	return b
}

func (b *backend) Hooks() []operator.Hook {
	return []operator.Hook{}
}

func (b *backend) Name() string {
	return "VictoriaMetrics"
}

func (b *backend) Ready(t *proto.Instance) bool {
	return true
}

func (b *backend) Initialize(nodes []*proto.Instance, target *proto.Instance) (*proto.NodeSpec, error) {

	if target.Group.Type == "storage" {
		target.Spec.Cmd = []string{
			"--storageDataPath=", "/storage",
		}
	} else {
		fmt.Println("__ FIND NODES __")
		// find the storage nodes
		storageNodes := []string{}
		for _, i := range nodes {
			if i.Group.Type == "storage" {
				storageNodes = append(storageNodes, i.FullName()+":8400")
			}
		}
		target.Spec.Cmd = []string{
			"--storageNode", strings.Join(storageNodes, ","),
		}
	}

	return nil, nil
}

// Spec implements the Handler interface
func (b *backend) Spec() *operator.Spec {
	return &operator.Spec{
		Name: "VictoriaMetrics",
		Config: schema.Schema2{
			Spec: &schema.Record{
				Fields: map[string]*schema.Field{},
			},
		},
		Nodetypes: map[string]operator.Nodetype{
			"storage": {
				Image:          "victoriametrics/vmstorage",
				DefaultVersion: "latest",
				Volumes:        []*operator.Volume{},
				Ports:          []*operator.Port{},
			},
			"insert": {
				Image:          "victoriametrics/vminsert",
				DefaultVersion: "latest",
				Volumes:        []*operator.Volume{},
				Ports:          []*operator.Port{},
				Schema: schema.Schema2{
					Spec: &schema.Record{
						Fields: map[string]*schema.Field{
							"storage_replicas": {
								Type:     schema.TypeInt,
								Computed: true,
							},
						},
					},
				},
			},
			"select": {
				Image:          "victoriametrics/vmselect",
				DefaultVersion: "latest",
				Volumes:        []*operator.Volume{},
				Ports:          []*operator.Port{},
				Schema: schema.Schema2{
					Spec: &schema.Record{
						Fields: map[string]*schema.Field{
							"storage_replicas": {
								Type:     schema.TypeInt,
								Computed: true,
							},
						},
					},
				},
			},
		},
		Validate: func(comp *proto.Component) (*proto.Component, error) {
			var spec proto.ClusterSpec
			if err := gproto.Unmarshal(comp.Spec.Value, &spec); err != nil {
				return nil, err
			}

			// num of storage replicas
			numStorage := 0
			for _, i := range spec.Groups {
				if i.Type == "storage" {
					numStorage = int(i.Count)
				}
			}

			for _, i := range spec.Groups {
				if i.Type != "storage" {
					i.Params = schema.MapToSpec(map[string]interface{}{
						"storage_replicas": numStorage,
					})
				}
			}
			/*
				for _, grp := range spec.Groups {
					grp.Params = schema.MapToSpec(map[string]interface{}{
						"replicas": int(grp.Count),
					})
				}
			*/
			comp.Spec = proto.MustMarshalAny(&spec)
			return comp, nil
		},
		Handlers: map[string]func(spec *proto.NodeSpec, grp *proto.ClusterSpec_Group, data *schema.ResourceData){
			"": func(spec *proto.NodeSpec, grp *proto.ClusterSpec_Group, data *schema.ResourceData) {
				// spec.AddEnv("ZOO_TICK_TIME", data.Get("tickTime").(string))
			},
		},
	}
}

// Client implements the Handler interface
func (b *backend) Client(node *proto.Instance) (interface{}, error) {
	return nil, nil
}

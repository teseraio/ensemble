package clickhouse

import (
	"fmt"

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
	return "clickhouse"
}

func (b *backend) Ready(t *proto.Instance) bool {
	return true
}

func (b *backend) Initialize(nodes []*proto.Instance, target *proto.Instance) (*proto.NodeSpec, error) {
	sch := b.Spec().Nodetypes[""].Schema
	data := schema.NewResourceData(&sch, target.Group.Params)
	zkNode := data.Get("zookeeper").(string)

	replicas := []*Replica{}

	uniqueNodes := map[string]struct{}{}
	for _, n := range nodes {
		uniqueNodes[n.FullName()] = struct{}{}
	}
	for node := range uniqueNodes {
		replicas = append(replicas, &Replica{
			Host: node,
			Port: 9009,
		})
	}

	obj := &Cluster{
		Name:      target.FullName(),
		Zookeeper: zkNode,
		Shards: []*Shard{
			{
				Replicas: replicas,
			},
		},
	}

	res, err := runTmpl("cluster", obj)
	if err != nil {
		panic(err)
	}

	target.Spec.AddFile("/etc/clickhouse-server/config.xml", string(res))
	target.Spec.AddFile("/etc/clickhouse-server/users.xml", users)

	return nil, nil
}

// Spec implements the Handler interface
func (b *backend) Spec() *operator.Spec {
	return &operator.Spec{
		Name: "Clickhouse",
		Nodetypes: map[string]operator.Nodetype{
			"": {
				Image:          "yandex/clickhouse-server",
				DefaultVersion: "20.4",
				Volumes:        []*operator.Volume{},
				Ports:          []*operator.Port{},
				Schema: schema.Schema2{
					Spec: &schema.Record{
						Fields: map[string]*schema.Field{
							// reference to the zookeeper so that its available in Initialized
							"zookeeper": {
								Type: schema.TypeString,
							},
							// use this field to force the update of the nodes
							"replicas": {
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
			if len(spec.Groups) != 1 {
				return nil, fmt.Errorf("only one group expected")
			}

			grp := spec.Groups[0]

			sch := b.Spec().Nodetypes[""].Schema
			data := schema.NewResourceData(&sch, grp.Params)
			zkNode := data.Get("zookeeper").(string)

			for _, grp := range spec.Groups {
				grp.Params = schema.MapToSpec(map[string]interface{}{
					"replicas":  int(grp.Count),
					"zookeeper": zkNode,
				})
			}

			spec.DependsOn = []string{
				zkNode,
			}
			comp.Spec = proto.MustMarshalAny(&spec)
			return comp, nil
		},
		Handlers: map[string]func(spec *proto.NodeSpec, grp *proto.ClusterSpec_Group, data *schema.ResourceData){
			"": func(spec *proto.NodeSpec, grp *proto.ClusterSpec_Group, data *schema.ResourceData) {
			},
		},
	}
}

// Client implements the Handler interface
func (b *backend) Client(node *proto.Instance) (interface{}, error) {
	return nil, nil
}

const users = `<?xml version="1.0"?>
<company>
    <profiles>
        <default>
            <max_memory_usage>10000000000</max_memory_usage>
            <use_uncompressed_cache>0</use_uncompressed_cache>
            <load_balancing>in_order</load_balancing>
            <log_queries>1</log_queries>
        </default>
    </profiles>
    <users>
        <default>
            <password></password>
            <profile>default</profile>
            <networks>
                <ip>::/0</ip>
            </networks>
            <quota>default</quota>
        </default>
        <admin>
            <password>123</password>
            <profile>default</profile>
            <networks>
                <ip>::/0</ip>
            </networks>
            <quota>default</quota>
        </admin>
    </users>
    <quotas>
        <default>
            <interval>
                <duration>3600</duration>
                <queries>0</queries>
                <errors>0</errors>
                <result_rows>0</result_rows>
                <read_rows>0</read_rows>
                <execution_time>0</execution_time>
            </interval>
        </default>
    </quotas>
</company>`

package clickhouse

import (
	"fmt"

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

	fmt.Println("// nodes //")
	fmt.Println(nodes)

	replicas := []*Replica{}
	for _, n := range nodes {
		replicas = append(replicas, &Replica{
			Host: n.FullName(),
			Port: 9000,
		})
	}
	obj := &Cluster{
		Name: target.FullName(),
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

	fmt.Println(string(res))

	target.Spec.AddFile("/etc/clickhouse-server/config.xml", string(res))
	target.Spec.AddFile("/etc/clickhouse-server/users.xml", users)

	//fmt.Println("// res //")
	//fmt.Println(string(res))

	// panic("X")

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
						Fields: map[string]*schema.Field{},
					},
				},
			},
		},
		/*
			Validate: func(comp *proto.Component) (*proto.Component, error) {
				return comp, nil
			},
		*/
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

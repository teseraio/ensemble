package kafka

import (
	"fmt"
	"strconv"

	gproto "github.com/golang/protobuf/proto"
	"github.com/teseraio/ensemble/operator"
	"github.com/teseraio/ensemble/operator/proto"
	"github.com/teseraio/ensemble/schema"
)

type backend struct {
	*operator.BaseOperator
}

// Factory is a factory method for the kafka backend
func Factory() operator.Handler {
	b := &backend{}
	b.BaseOperator = &operator.BaseOperator{}
	b.BaseOperator.SetHandler(b)
	return b
}

func (b *backend) Name() string {
	return "kafka"
}

func (b *backend) Ready(t *proto.Instance) bool {
	return true
}

func (b *backend) Initialize(nodes []*proto.Instance, target *proto.Instance) (*proto.NodeSpec, error) {

	localIndex, err := proto.ParseIndex(target.Name)
	if err != nil {
		return nil, err
	}

	sch := b.Spec().Nodetypes[""].Schema
	data := schema.NewResourceData(&sch, target.Group.Params)
	zkNode := data.Get("zookeeper").(string)

	target.Spec.AddEnv("KAFKA_BROKER_ID", strconv.Itoa(int(localIndex)))
	target.Spec.AddEnv("KAFKA_ZOOKEEPER_CONNECT", fmt.Sprintf("%s:2181", zkNode))
	target.Spec.AddEnv("KAFKA_ADVERTISED_LISTENERS", fmt.Sprintf("PLAINTEXT://%s:29092", target.FullName()))
	target.Spec.AddEnv("KAFKA_LISTENER_SECURITY_PROTOCOL_MAP", "PLAINTEXT:PLAINTEXT,PLAINTEXT_HOST:PLAINTEXT")
	target.Spec.AddEnv("KAFKA_INTER_BROKER_LISTENER_NAME", "PLAINTEXT")
	target.Spec.AddEnv("KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR", "1")

	return nil, nil
}

// Spec implements the Handler interface
func (b *backend) Spec() *operator.Spec {
	return &operator.Spec{
		Name: "Kafka",
		Nodetypes: map[string]operator.Nodetype{
			"": {
				Image:          "confluentinc/cp-kafka",
				DefaultVersion: "5.3.1",
				Volumes:        []*operator.Volume{},
				Ports:          []*operator.Port{},
				Schema: schema.Schema2{
					Spec: &schema.Record{
						Fields: map[string]*schema.Field{
							"zookeeper": {
								Type:     schema.TypeString,
								Required: true,
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

package zookeeper

import (
	"fmt"
	"strconv"
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

func (b *backend) Setup2() {
	fmt.Println("_ SETUP _")
}

func (b *backend) Hooks() []operator.Hook {
	return []operator.Hook{}
}

func (b *backend) Name() string {
	return "zookeeper"
}

func (b *backend) Ready(t *proto.Instance) bool {
	return true
}

func (b *backend) Initialize(nodes []*proto.Instance, target *proto.Instance) (*proto.NodeSpec, error) {
	// Id of the instance
	localIndex, err := proto.ParseIndex(target.Name)
	if err != nil {
		return nil, err
	}
	target.Spec.AddEnv("ZOO_MY_ID", strconv.Itoa(int(localIndex)))

	// list of the zookeeper instances
	res := []string{}
	for _, node := range nodes {
		remoteIndex, err := proto.ParseIndex(node.Name)
		if err != nil {
			return nil, err
		}

		if node.ID == target.ID {
			res = append(res, fmt.Sprintf("server.%d=0.0.0.0:2888:3888;2181", remoteIndex))
		} else {
			res = append(res, getZkNodeSpec(node, remoteIndex))
		}
	}

	//target.Healthy = true
	target.Spec.AddEnv("ZOO_SERVERS", strings.Join(res, " "))
	return nil, nil
}

func getZkNodeSpec(node *proto.Instance, index uint64) string {
	return fmt.Sprintf("server.%d=%s:2888:3888;2181", index, node.FullName())
}

// Spec implements the Handler interface
func (b *backend) Spec() *operator.Spec {
	return &operator.Spec{
		Name: "Zookeeper",
		Nodetypes: map[string]operator.Nodetype{
			"": {
				Image:          "zookeeper",
				DefaultVersion: "3.6",
				Volumes:        []*operator.Volume{},
				Ports:          []*operator.Port{},
				Schema: schema.Schema2{
					Spec: &schema.Record{
						Fields: map[string]*schema.Field{
							"tickTime": {
								Type:    schema.TypeString,
								Default: "2000",
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
			if grp.Count != 1 && grp.Count < 3 {
				if grp.Count != 1 {
					return nil, fmt.Errorf("either 1 or 3 expected")
				}
			}
			if grp.Count%2 == 0 {
				return nil, fmt.Errorf("odd number of nodes required")
			}
			return comp, nil
		},
		Handlers: map[string]func(spec *proto.NodeSpec, grp *proto.ClusterSpec_Group, data *schema.ResourceData){
			"": func(spec *proto.NodeSpec, grp *proto.ClusterSpec_Group, data *schema.ResourceData) {
				spec.AddEnv("ZOO_TICK_TIME", data.Get("tickTime").(string))
			},
		},
	}
}

// Client implements the Handler interface
func (b *backend) Client(node *proto.Instance) (interface{}, error) {
	return nil, nil
}

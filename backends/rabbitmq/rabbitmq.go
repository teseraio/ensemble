package rabbitmq

import (
	"fmt"
	"time"

	rabbithole "github.com/michaelklishin/rabbit-hole/v2"
	"github.com/teseraio/ensemble/lib/template"
	"github.com/teseraio/ensemble/operator"
	"github.com/teseraio/ensemble/operator/proto"
	"github.com/teseraio/ensemble/schema"
)

const (
	rabbitmqConf    = "/etc/rabbitmq/rabbitmq.conf"
	rabbitmqPlugins = "/etc/rabbitmq/enabled_plugins"
)

const (
	enabledPlugins = "[rabbitmq_management,rabbitmq_management_agent]."
)

type backend struct {
	*operator.BaseOperator
}

// Factory returns a factory method for the zookeeper backend
func Factory() operator.Handler {
	b := &backend{}
	b.BaseOperator = &operator.BaseOperator{}
	b.BaseOperator.SetHandler(b)
	return b
}

func loopRetry(timeout time.Duration, handler func() error) error {
	doneCh := make(chan struct{})
	go func() {
		<-time.After(timeout)
		doneCh <- struct{}{}
	}()

	timeInterval := 1 * time.Second
	for {
		select {
		case <-time.After(timeInterval):
		case <-doneCh:
			return fmt.Errorf("timeout")
		}

		err := handler()
		if err == nil {
			break
		}
	}
	return nil
}

func (b *backend) Hooks() []operator.Hook {
	return []operator.Hook{
		{
			Name:  "readiness",
			State: proto.Instance_RUNNING,
			Handler: func(emit func(i *proto.InstanceUpdate), req operator.ApplyHookRequest) error {
				instance := req.Instance
				if instance.Healthy {
					return nil
				}

				clt, err := rabbithole.NewClient("http://"+instance.Ip+":15672", "guest", "guest")
				if err != nil {
					return err
				}

				// check if rabbimq is running
				err = loopRetry(5*time.Minute, func() error {
					_, err = clt.Overview()
					fmt.Println(err)
					return err
				})
				if err != nil {
					return fmt.Errorf("timeout readiness probe")
				}

				nodesExpected := len(req.Deployment.Instances)

				// check if its syncer with others
				err = loopRetry(5*time.Minute, func() error {
					nodes, err := clt.ListNodes()
					if err != nil {
						return err
					}
					// fmt.Println(len(nodes), nodesExpected)
					if len(nodes) == nodesExpected {
						return nil
					}
					return fmt.Errorf("not yet")
				})
				if err != nil {
					return fmt.Errorf("failed cluster formation")
				}

				emit(instance.Update(&proto.InstanceUpdate_Healthy_{
					Healthy: &proto.InstanceUpdate_Healthy{},
				}))

				return nil
			},
		},
	}
}

func (b *backend) Name() string {
	return "Rabbitmq"
}

const rabbitmqConfFile = `
cluster_formation.peer_discovery_backend = classic_config

loopback_users = none

{{ if .Nodes }}
{{ range $i, $elem := .Nodes }}
cluster_formation.classic_config.nodes.{{ $i }} = rabbit@{{ $elem }}
{{ end }}
{{ end }}`

func (b *backend) Initialize(n []*proto.Instance, target *proto.Instance) (*proto.NodeSpec, error) {
	target.Spec.AddEnv("RABBITMQ_ERLANG_COOKIE", "TODO")
	target.Spec.AddEnv("RABBITMQ_USE_LONGNAME", "true")

	target.Spec.AddFile(rabbitmqPlugins, enabledPlugins)

	var nodes []string
	for _, i := range n {
		if i.ID != target.ID {
			nodes = append(nodes, i.FullName())
		}
	}
	configContent, err := template.RunTmpl(rabbitmqConfFile, map[string]interface{}{"Nodes": nodes})
	if err != nil {
		return nil, err
	}
	target.Spec.AddFile(rabbitmqConf, string(configContent))
	return nil, nil
}

func (b *backend) Ready(t *proto.Instance) bool {
	clt, err := rabbithole.NewClient("http://"+t.Ip+":15672", "guest", "guest")
	if err != nil {
		return false
	}
	if _, err := clt.Overview(); err != nil {
		return false
	}
	return true
}

// Spec implements the Handler interface
func (b *backend) Spec() *operator.Spec {
	return &operator.Spec{
		Name: "Rabbitmq",
		Nodetypes: map[string]operator.Nodetype{
			"": {
				Image:          "rabbitmq",
				DefaultVersion: "3.8.14-management",
				Volumes:        []*operator.Volume{},
				Ports:          []*operator.Port{
					// http-api 15672
				},
				Schema: schema.Schema2{
					Spec: &schema.Record{},
				},
			},
		},
		Handlers: map[string]func(spec *proto.NodeSpec, grp *proto.ClusterSpec_Group, data *schema.ResourceData){
			"": func(spec *proto.NodeSpec, grp *proto.ClusterSpec_Group, data *schema.ResourceData) {
			},
		},
		Resources: []*operator.Resource2{
			user(),
			exchange(),
			vhost(),
		},
	}
}

// Client implements the Handler interface
func (b *backend) Client(node *proto.Instance) (interface{}, error) {
	return rabbithole.NewClient("http://"+node.Ip+":15672", "guest", "guest")
}

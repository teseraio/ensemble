package rabbitmq

import (
	rabbithole "github.com/michaelklishin/rabbit-hole/v2"
	"github.com/teseraio/ensemble/lib/template"
	"github.com/teseraio/ensemble/operator"
	"github.com/teseraio/ensemble/operator/proto"
)

const (
	rabbitmqConf    = "/etc/rabbitmq/rabbitmq.conf"
	rabbitmqPlugins = "/etc/rabbitmq/enabled_plugins"
)

const (
	enabledPlugins = "[rabbitmq_management,rabbitmq_management_agent]."
)

type backend struct {
}

// Factory returns a factory method for the zookeeper backend
func Factory() operator.Handler {
	return &backend{}
}

const rabbitmqConfFile = `
cluster_formation.peer_discovery_backend = classic_config

loopback_users = none

{{ if .Nodes }}
{{ range $i, $elem := .Nodes }}
cluster_formation.classic_config.nodes.{{ $i }} = rabbit@{{ $elem }}
{{ end }}
{{ end }}`

/*
cluster_formation.classic_config.nodes.1 = rabbit@A0.A
cluster_formation.classic_config.nodes.2 = rabbit@A1.A
cluster_formation.classic_config.nodes.3 = rabbit@A2.A
*/

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
		panic(err)
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

/*
// EvaluatePlan implements the Handler interface
func (b *backend) EvaluatePlan(plan *proto.Context) error {
	if plan.Plan.Sets[0].DelNodesNum != 0 {
		set := plan.Plan.Sets[0]
		for _, n := range plan.Cluster.Nodes[:set.DelNodesNum] {
			set.DelNodes = append(set.DelNodes, n.ID)
		}
	}
	return nil
}
*/

// Spec implements the Handler interface
func (b *backend) Spec() *operator.Spec {
	return &operator.Spec{
		Name: "Rabbitmq",
		Nodetypes: map[string]operator.Nodetype{
			"": {
				Image:   "rabbitmq",
				Version: "latest", // TODO
				Volumes: []*operator.Volume{},
				Ports:   []*operator.Port{
					// http-api 15672
				},
			},
		},
		Handlers: map[string]func(spec *proto.NodeSpec, grp *proto.ClusterSpec_Group){
			"": func(spec *proto.NodeSpec, grp *proto.ClusterSpec_Group) {
				spec.Image = "rabbitmq"
				spec.Version = "latest"
			},
		},
		Resources: []operator.Resource{
			&User{},
			&Exchange{},
			&VHost{},
		},
	}
}

// Client implements the Handler interface
func (b *backend) Client(node *proto.Instance) (interface{}, error) {
	return rabbithole.NewClient("http://"+node.Ip+":15672", "guest", "guest")
}

/*
// Reconcile implements the Handler interface
func (b *backend) Reconcile(executor operator.Executor, e *proto.Cluster, node *proto.Node, plan *proto.Context) error {
	switch node.State {
	case proto.Node_INITIALIZED:
		recocileNodeInitialized(node)

	case proto.Node_PENDING:
		time.Sleep(10 * time.Second)

	case proto.Node_RUNNING:
		return b.reconcileNodeRunning(executor, e, node)

	case proto.Node_TAINTED:
		return reconcileNodeTainted(executor, node)
	}
	return nil
}

func reconcileNodeTainted(executor operator.Executor, node *proto.Node) error {
	return executor.Exec(node, rabbitmqctl, "shutdown")
}

func recocileNodeInitialized(node *proto.Node) {
	node.Spec.AddEnv("RABBITMQ_ERLANG_COOKIE", "TODO")
	node.Spec.AddEnv("RABBITMQ_USE_LONGNAME", "true")

	// enable the http management plugin by default
	node.Spec.AddEnv("RABBITMQ_ENABLED_PLUGINS_FILE", "/some/enabled_plugins")
	node.Spec.AddFile("/some/enabled_plugins", enabledPlugins)
}

func (b *backend) reconcileNodeRunning(executor operator.Executor, e *proto.Cluster, node *proto.Node) error {
	if len(e.Nodes) != 1 {
		// node joining a cluster
		target := e.Nodes[0]

		clt, err := b.Client(target)
		if err != nil {
			return err
		}

		info, err := clt.(*rabbithole.Client).Overview()
		if err != nil {
			return err
		}

		nodeName := info.Node
		if err := addNode(executor, node, nodeName); err != nil {
			return err
		}
	}
	return nil
}
*/

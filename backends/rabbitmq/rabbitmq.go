package rabbitmq

import (
	"time"

	rabbithole "github.com/michaelklishin/rabbit-hole/v2"
	"github.com/teseraio/ensemble/operator"
	"github.com/teseraio/ensemble/operator/proto"
)

var (
	masterKey = "master"
)

const (
	enabledPlugins = "[rabbitmq_management,rabbitmq_management_agent,rabbitmq_shovel]."
)

const rabbitmqctl = "rabbitmqctl"

func addNode(executor operator.Executor, node *proto.Node, master string) error {
	if err := executor.Exec(node, rabbitmqctl, "stop_app"); err != nil {
		return err
	}
	if err := executor.Exec(node, rabbitmqctl, "reset"); err != nil {
		return err
	}
	if err := executor.Exec(node, rabbitmqctl, "join_cluster", master); err != nil {
		return err
	}
	if err := executor.Exec(node, rabbitmqctl, "start_app"); err != nil {
		return err
	}
	return nil
}

type backend struct {
}

// Factory returns a factory method for the zookeeper backend
func Factory() operator.Handler {
	return &backend{}
}

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

// Spec implements the Handler interface
func (b *backend) Spec() *operator.Spec {
	return &operator.Spec{
		Name: "Rabbitmq",
		Nodetypes: map[string]operator.Nodetype{
			"": {
				Image:   "rabbitmq",
				Version: "latest", // TODO
				Volumes: []*operator.Volume{},
				Ports:   []*operator.Port{},
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
func (b *backend) Client(node *proto.Node) (interface{}, error) {
	return rabbithole.NewClient("http://"+node.Addr+":15672", "guest", "guest")
}

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

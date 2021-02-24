package zookeeper

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/teseraio/ensemble/operator"
	"github.com/teseraio/ensemble/operator/proto"
)

const (
	// keyIndx is the key to store the index node
	// in the cluster
	// keyIndx = "Indx"

	// keyRole is the key to store the role of the
	// node in the ensemble (observer, participant)
	keyRole = "Role"

	// roleParticipant is an active node in the ensemble
	roleParticipant = "participant"

	// roleObserver is a follower node in the ensemble
	// that does not form part of the ensemble
	roleObserver = "observer"
)

type backend struct {
	operator.BaseHandler
}

// Factory is a factory method for the zookeeper backend
func Factory() operator.Handler {
	return &backend{}
}

func (b *backend) Ready(t *proto.Instance) bool {
	return false
}

func (b *backend) PostHook(*operator.HookCtx) error {
	// TAINTED: TODO
	return nil
}

func (b *backend) EvaluateConfig(spec *proto.NodeSpec, cc map[string]string) error {
	spec.Image = "zookeeper"
	spec.Version = "3.6"

	// this should be pretty deterministic
	var c *config
	if err := mapstructure.WeakDecode(cc, &c); err != nil {
		return err
	}
	if c != nil {
		if c.TickTime != 0 {
			spec.AddEnv("ZOO_TICK_TIME", strconv.Itoa(int(c.TickTime)))
		}
	}
	return nil
}

func (b *backend) Initialize(nodes []*proto.Instance, target *proto.Instance) (*proto.NodeSpec, error) {
	// Id of the instance
	target.Spec.AddEnv("ZOO_MY_ID", strconv.Itoa(int(target.Index)))

	// list of the zookeeper instances
	res := []string{}
	for _, node := range nodes {
		res = append(res, getZkNodeSpec(node))
	}
	target.Spec.AddEnv("ZOO_SERVERS", strings.Join(res, " "))

	return nil, nil
}

// EvaluatePlan implements the Handler interface
func (b *backend) EvaluatePlan(n []*proto.Instance) error {
	// set the index of each node
	/*
		for indx, nn := range n {
			nn.Set(keyIndx, strconv.Itoa(indx))
		}
	*/
	/*
		switch obj := ctx.Plan.Action.(type) {
		case *proto.Plan_Step_ActionScale_:
			if obj.ActionScale.Direction == proto.Plan_Step_ActionScale_UP {
				return b.evaluatePlanScaleUp(ctx, obj.ActionScale)
			}
			return b.evaluatePlanScaleDown(ctx, obj.ActionScale)
		}
	*/
	return nil
}

/*
func (b *backend) evaluatePlanScaleUp(ctx *operator.PlanCtx, action *proto.Plan_Step_ActionScale) error {
	return nil
}

func (b *backend) evaluatePlanScaleDown(ctx *operator.PlanCtx, action *proto.Plan_Step_ActionScale) error {
	return nil
}
*/

// zookeeper only has one set
//plan := ctx.Plan.Sets[0]

//if plan.DelNodesNum != 0 {
// scale down

/*
	cc := ctx.Cluster.Copy()
	sort.Sort(sortedNodes(cc.Nodes))

	delNodes := []string{}
	for i := 0; i < int(plan.DelNodesNum); i++ {
		delNodes = append(delNodes, cc.Nodes[i].ID)
	}
	plan.DelNodes = delNodes
*/

//} else {
// scale up

/*
	// start the index in 1
	ogIndx := len(ctx.Cluster.Nodes) + 1

	// add a sequential index to each node
	for seqIndx, n := range plan.AddNodes {
		indx := strconv.Itoa(ogIndx + seqIndx)
		n.Set(keyIndx, indx)
		n.Spec.AddEnv("ZOO_MY_ID", indx)

		if ctx.Plan.Bootstrap {
			// participant
			n.Set(keyRole, roleParticipant)
		} else {
			// observer
			n.Set(keyRole, roleObserver)
		}
	}

	// get the cluster nodes
	var nodes []*proto.Node
	if ctx.Plan.Bootstrap {
		// join as participant
		nodes = plan.AddNodes
	} else {
		// join as observer
		nodes = ctx.Cluster.Nodes
	}

	// add the cluster nodes
	for _, n := range plan.AddNodes {
		var res []string
		for _, node := range nodes {
			res = append(res, getZkNodeSpec(node))
		}
		if !ctx.Plan.Bootstrap {
			// add yourself to the cluster too
			res = append(res, getZkNodeSpec(n))
		}

		// add the configuration
		config := ctx.NodeTypes[n.Nodetype].Config.(*config)

		if config != nil {
			if config.TickTime != 0 {
				n.Spec.AddEnv("ZOO_TICK_TIME", strconv.Itoa(int(config.TickTime)))
			}
		}
		n.Spec.AddEnv("ZOO_SERVERS", strings.Join(res, " "))
	}
*/
//}

func getZkNodeSpec(node *proto.Instance) string {
	return fmt.Sprintf("server.%d=%s:2888:3888;2181", node.Index, node.FullName())
}

type config struct {
	TickTime uint64 `mapstructure:"tickTime"`
}

// Spec implements the Handler interface
func (b *backend) Spec() *operator.Spec {
	return &operator.Spec{
		Name: "Zookeeper",
		Nodetypes: map[string]operator.Nodetype{
			"": {
				Image:   "zookeeper",
				Version: "3.6",
				Volumes: []*operator.Volume{},
				Ports:   []*operator.Port{},
				Config:  &config{},
			},
		},
		Handlers: map[string]func(spec *proto.NodeSpec, grp *proto.ClusterSpec2_Group){
			"": func(spec *proto.NodeSpec, grp *proto.ClusterSpec2_Group) {
				fmt.Println("X")
				spec.Image = "zookeeper"
				spec.Version = "3.6"

				var c *config
				if err := mapstructure.WeakDecode(grp.Config, &c); err != nil {
					panic(err)
				}
				if c != nil {
					if c.TickTime != 0 {
						spec.AddEnv("ZOO_TICK_TIME", strconv.Itoa(int(c.TickTime)))
					}
				}
			},
		},
	}
}

// Client implements the Handler interface
func (b *backend) Client(node *proto.Instance) (interface{}, error) {
	/*
		c, _, err := zk.Connect([]string{node.Addr}, time.Second)
		if err != nil {
			return nil, err
		}
		return c, nil
	*/
	panic("X")
	return nil, nil
}

type sortedNodes []*proto.Instance

func (s sortedNodes) Len() int      { return len(s) }
func (s sortedNodes) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s sortedNodes) Less(i, j int) bool {
	/*
		if s[i].Get(keyRole) == roleObserver {
			return true
		}
		if s[j].Get(keyRole) == roleObserver {
			return true
		}
	*/
	return false
}

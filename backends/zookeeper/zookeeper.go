package zookeeper

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-zookeeper/zk"
	"github.com/teseraio/ensemble/operator"
	"github.com/teseraio/ensemble/operator/proto"
)

const (
	// keyIndx is the key to store the index node
	// in the cluster
	keyIndx = "Indx"

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

func (b *backend) PostHook(*operator.HookCtx) error {
	// TAINTED: TODO
	return nil
}

// EvaluatePlan implements the Handler interface
func (b *backend) EvaluatePlan(ctx *operator.PlanCtx) error {

	// zookeeper only has one set
	plan := ctx.Plan.Sets[0]

	if plan.DelNodesNum != 0 {
		// scale down

		cc := ctx.Cluster.Copy()
		sort.Sort(sortedNodes(cc.Nodes))

		delNodes := []string{}
		for i := 0; i < int(plan.DelNodesNum); i++ {
			delNodes = append(delNodes, cc.Nodes[i].ID)
		}
		plan.DelNodes = delNodes

	} else {
		// scale up

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
			n.Spec.AddEnv("ZOO_SERVERS", strings.Join(res, " "))
		}
	}
	return nil
}

func getZkNodeSpec(node *proto.Node) string {
	return fmt.Sprintf("server.%s=%s:2888:3888:%s;2181", node.Get(keyIndx), node.FullName(), node.Get(keyRole))
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
			},
		},
	}
}

// Client implements the Handler interface
func (b *backend) Client(node *proto.Node) (interface{}, error) {
	c, _, err := zk.Connect([]string{node.Addr}, time.Second)
	if err != nil {
		return nil, err
	}
	return c, nil
}

type sortedNodes []*proto.Node

func (s sortedNodes) Len() int      { return len(s) }
func (s sortedNodes) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s sortedNodes) Less(i, j int) bool {
	if s[i].Get(keyRole) == roleObserver {
		return true
	}
	if s[j].Get(keyRole) == roleObserver {
		return true
	}
	return false
}

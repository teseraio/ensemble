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
}

// Factory is a factory method for the zookeeper backend
func Factory() operator.Handler {
	return &backend{}
}

// Reconcile implements the Handler interface
func (b *backend) Reconcile(_ operator.Executor, e *proto.Cluster, node *proto.Node, plan *proto.Context) error {
	if node.State == proto.Node_INITIALIZED {
		nodeInitialized(e, node, plan.Plan.Bootstrap, plan.Set)
	}
	return nil
}

func nodeInitialized(e *proto.Cluster, node *proto.Node, bootstrap bool, plan *proto.Plan_Set) {

	// enable reconfig for the API
	node.Spec.AddEnv("ZOO_CFG_EXTRA", "reconfigEnabled=true")

	res := []string{}
	if bootstrap {
		for indxInt, peer := range plan.AddNodes {
			// do not start the indexes with 0
			indx := strconv.Itoa(indxInt + 1)

			if peer.ID == node.ID {
				node.Set(keyIndx, indx)

				// nodes are included as participant during bootstrap
				node.Set(keyRole, roleParticipant)

				res = append(res, fmt.Sprintf("server.%s=0.0.0.0:2888:3888;2181", indx))
			} else {
				res = append(res, fmt.Sprintf("server.%s=%s:2888:3888;2181", indx, peer.FullName()))
			}
		}
	} else {
		// Nodes are joined as observers (TODO. handler to change them to participants)
		indx := strconv.Itoa(len(e.Nodes) + 1)

		node.Set(keyIndx, indx)
		node.Set(keyRole, roleObserver)

		// get a subset of the nodes in the cluster and join them as observer
		for _, n := range e.Nodes {
			res = append(res, getZkNodeSpec(n))
		}
		// add our node as observer
		res = append(res, getZkNodeSpec(node))
	}

	//	set the id of the node
	node.Spec.AddEnv("ZOO_MY_ID", node.Get(keyIndx))

	// set the list of servers
	node.Spec.AddEnv("ZOO_SERVERS", strings.Join(res, " "))
}

// EvaluatePlan implements the Handler interface
func (b *backend) EvaluatePlan(ctx *proto.Context) error {
	// there is only one set in the plan since there is only one node type
	set := ctx.Plan.Sets[0]
	if set.DelNodesNum != 0 {
		cc := ctx.Cluster.Copy()
		sort.Sort(sortedNodes(cc.Nodes))

		delNodes := []string{}
		for i := 0; i < int(set.DelNodesNum); i++ {
			delNodes = append(delNodes, cc.Nodes[i].ID)
		}
		set.DelNodes = delNodes
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
			"": operator.Nodetype{
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

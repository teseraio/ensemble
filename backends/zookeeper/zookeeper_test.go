package zookeeper

import (
	"fmt"
	"testing"

	"github.com/teseraio/ensemble/operator"
	"github.com/teseraio/ensemble/operator/proto"
	"github.com/teseraio/ensemble/schema"
)

func TestBootstrap(t *testing.T) {
	h := &operator.Harness{
		Deployment: &proto.Deployment{},
		Handler:    Factory(),
	}
	h.Scheduler = operator.NewScheduler(h)
	dep := h.ApplySched(&proto.Component{
		Spec: proto.MustMarshalAny(&proto.ClusterSpec{
			Groups: []*proto.ClusterSpec_Group{
				{
					Count:  3,
					Params: schema.MapToSpec(nil),
				},
			},
		}),
	})

	operator.Assert(dep, h.Plan.NodeUpdate[0], operator.NodeExpect{
		Env: map[string]string{
			"ZOO_MY_ID":   "1",
			"ZOO_SERVERS": "server.1=0.0.0.0:2888:3888;2181 server.2={{.Node2}}:2888:3888;2181 server.3={{.Node3}}:2888:3888;2181",
		},
	})
	operator.Assert(dep, h.Plan.NodeUpdate[1], operator.NodeExpect{
		Env: map[string]string{
			"ZOO_MY_ID":   "2",
			"ZOO_SERVERS": "server.1={{.Node1}}:2888:3888;2181 server.2=0.0.0.0:2888:3888;2181 server.3={{.Node3}}:2888:3888;2181",
		},
	})
	operator.Assert(dep, h.Plan.NodeUpdate[2], operator.NodeExpect{
		Env: map[string]string{
			"ZOO_MY_ID":   "3",
			"ZOO_SERVERS": "server.1={{.Node1}}:2888:3888;2181 server.2={{.Node2}}:2888:3888;2181 server.3=0.0.0.0:2888:3888;2181",
		},
	})

	// move all the nodes to running since they are pending
	for _, n := range h.Deployment.Instances {
		n.Status = proto.Instance_RUNNING
	}

	// apply the update
	h.ApplySched(&proto.Component{
		Sequence: 1,
		Spec: proto.MustMarshalAny(&proto.ClusterSpec{
			Groups: []*proto.ClusterSpec_Group{
				{
					Count: 3,
					Params: schema.MapToSpec(map[string]interface{}{
						"tickTime": "3000",
					}),
				},
			},
		}),
	})

	fmt.Println("-- plan --")
	for _, i := range h.Plan.NodeUpdate {
		fmt.Println(i)
	}
}

package zookeeper

import (
	"testing"

	"github.com/teseraio/ensemble/operator"
	"github.com/teseraio/ensemble/operator/proto"
	"github.com/teseraio/ensemble/schema"
	"github.com/teseraio/ensemble/testutil"
)

func TestBootstrap(t *testing.T) {
	h := operator.NewHarness(t)
	h.Handler = Factory()
	h.Scheduler = operator.NewScheduler(h)

	h.AddComponent(&proto.Component{
		Id: "a1",
		Spec: proto.MustMarshalAny(&proto.ClusterSpec{
			Groups: []*proto.ClusterSpec_Group{
				{
					Count:  3,
					Params: schema.MapToSpec(nil),
				},
			},
		}),
	})

	plan := h.Eval()
	h.Expect(plan, &operator.HarnessExpect{
		Nodes: []*operator.HarnessExpectInstance{
			{
				Spec: &proto.NodeSpec{
					Env: map[string]string{
						"ZOO_MY_ID":   "1",
						"ZOO_SERVERS": "server.1=0.0.0.0:2888:3888;2181 server.2={{.Node_2}}:2888:3888;2181 server.3={{.Node_3}}:2888:3888;2181",
					},
				},
			},
			{
				Spec: &proto.NodeSpec{
					Env: map[string]string{
						"ZOO_MY_ID":   "2",
						"ZOO_SERVERS": "server.1={{.Node_1}}:2888:3888;2181 server.2=0.0.0.0:2888:3888;2181 server.3={{.Node_3}}:2888:3888;2181",
					},
				},
			},
			{
				Spec: &proto.NodeSpec{
					Env: map[string]string{
						"ZOO_MY_ID":   "3",
						"ZOO_SERVERS": "server.1={{.Node_1}}:2888:3888;2181 server.2={{.Node_2}}:2888:3888;2181 server.3=0.0.0.0:2888:3888;2181",
					},
				},
			},
		},
	})

	h.ApplyDep(plan, func(n *proto.Instance) {
		n.Status = proto.Instance_RUNNING
		n.Healthy = true
	})

	// it should be done once all nodes are running
	plan = h.Eval()
	h.Expect(plan, &operator.HarnessExpect{
		Status: "done",
	})

	// Update tick time
	h.AddComponent(&proto.Component{
		Id:       "a2",
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

	plan = h.Eval()
	h.Expect(plan, &operator.HarnessExpect{
		Nodes: []*operator.HarnessExpectInstance{
			{
				Name:   "{{.Node_1}}",
				Status: proto.Instance_TAINTED,
			},
			{
				Name:   "{{.Node_2}}",
				Status: proto.Instance_TAINTED,
			},
		},
	})
}

func TestE2E(t *testing.T) {
	// testutil.IsE2EEnabled(t)

	srv := testutil.TestOperator(t, Factory)
	// defer srv.Close()

	uuid := srv.Apply(&proto.Component{
		Name: "A",
		Spec: proto.MustMarshalAny(&proto.ClusterSpec{
			Backend: "Zookeeper",
			Groups: []*proto.ClusterSpec_Group{
				{
					Count: 3,
					Params: schema.MapToSpec(
						map[string]interface{}{
							// "tickTime": "2000",
						},
					),
				},
			},
		}),
	})

	srv.WaitForTask(uuid)

	/*
		uuid = srv.Apply(&proto.Component{
			Name: "A",
			Spec: proto.MustMarshalAny(&proto.ClusterSpec{
				Backend: "Zookeeper",
				Groups: []*proto.ClusterSpec_Group{
					{
						Count: 3,
						Params: schema.MapToSpec(
							map[string]interface{}{
								"tickTime": "3000",
							},
						),
					},
				},
			}),
		})

		srv.WaitForTask(uuid)
	*/
}

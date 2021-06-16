package cassandra

import (
	"testing"

	"github.com/teseraio/ensemble/operator"
	"github.com/teseraio/ensemble/operator/proto"
	"github.com/teseraio/ensemble/testutil"
)

func TestBootstrap(t *testing.T) {
	h := operator.NewHarness(t)
	h.Handler = Factory()
	h.Scheduler = operator.NewScheduler(h)

	h.AddComponent(&proto.Component{
		Spec: proto.MustMarshalAny(&proto.ClusterSpec{
			Groups: []*proto.ClusterSpec_Group{
				{
					Count: 2,
				},
			},
		}),
	})

	plan := h.Eval()
	h.Expect(plan, &operator.HarnessExpect{
		Nodes: []*operator.HarnessExpectInstance{
			{
				KV: map[string]string{
					"seed": "ok",
				},
			},
			{
				Spec: &proto.NodeSpec{
					Env: map[string]string{
						"CASSANDRA_SEEDS": "{{.Node_1}}",
					},
				},
			},
		},
	})

	h.ApplyDep(plan, func(n *proto.Instance) {
		n.Status = proto.Instance_RUNNING
		n.Healthy = true
	})

	plan = h.Eval()
	h.Expect(plan, &operator.HarnessExpect{
		Status: "done",
	})
}

func TestE2E(t *testing.T) {
	testutil.IsE2EEnabled(t)

	srv := testutil.TestOperator(t, Factory)
	// defer srv.Close()

	uuid := srv.Apply(&proto.Component{
		Name: "A",
		Spec: proto.MustMarshalAny(&proto.ClusterSpec{
			Backend: "Cassandra",
			Groups: []*proto.ClusterSpec_Group{
				{
					Count: 2,
				},
			},
		}),
	})

	srv.WaitForTask(uuid)
}

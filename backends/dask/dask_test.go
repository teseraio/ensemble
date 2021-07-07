package dask

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
					Type:  "scheduler",
					Count: 1,
				},
				{
					Type:  "worker",
					Count: 1,
				},
			},
		}),
	})

	plan := h.Eval()
	h.Expect(plan, &operator.HarnessExpect{
		Nodes: []*operator.HarnessExpectInstance{
			{
				Spec: &proto.NodeSpec{
					Args: []string{},
				},
			},
		},
	})
	h.ApplyDep(plan, func(n *proto.Instance) {
		n.Status = proto.Instance_RUNNING
		n.Healthy = true
	})

	// create worker
	plan = h.Eval()
	h.Expect(plan, &operator.HarnessExpect{
		Nodes: []*operator.HarnessExpectInstance{
			{
				Spec: &proto.NodeSpec{
					Args: []string{
						"tcp://{{.Node_scheduler_1}}:8786",
					},
				},
			},
		},
	})
}

func TestE2E(t *testing.T) {
	testutil.IsE2EEnabled(t)

	srv := testutil.TestOperator(t, Factory)
	// defer srv.Close()

	uuid := srv.Apply(&proto.Component{
		Name: "A",
		Spec: proto.MustMarshalAny(&proto.ClusterSpec{
			Backend: "Dask",
			Groups: []*proto.ClusterSpec_Group{
				{
					Type:  "scheduler",
					Count: 1,
				},
				{
					Type:  "worker",
					Count: 1,
				},
			},
		}),
	})

	srv.WaitForTask(uuid)

	uuid = srv.Apply(&proto.Component{
		Name:   "A",
		Action: proto.Component_DELETE,
		Spec: proto.MustMarshalAny(&proto.ClusterSpec{
			Backend: "Dask",
		}),
	})

	srv.WaitForTask(uuid)
}

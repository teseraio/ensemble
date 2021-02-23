package zookeeper

import (
	"fmt"
	"testing"
	"time"

	"github.com/teseraio/ensemble/operator/proto"
	"github.com/teseraio/ensemble/testutil"
)

func TestBootstrap(t *testing.T) {
	srv := testutil.TestOperator(t, Factory)
	// defer srv.Close()

	srv.Apply(&proto.Component{
		Name: "A",
		Spec: proto.MustMarshalAny(&proto.ClusterSpec2{
			Backend: "Zookeeper",
			Groups: []*proto.ClusterSpec2_Group{
				{
					Count:    3,
					Revision: 1,
				},
			},
		}),
	})

	//srv.WaitForTask(uuid)

	time.Sleep(3 * time.Second)

	fmt.Printf("\n\n\nSTART REVISION\n\n\n")

	srv.Apply(&proto.Component{
		Name: "A",
		Spec: proto.MustMarshalAny(&proto.ClusterSpec2{
			Backend: "Zookeeper",
			Groups: []*proto.ClusterSpec2_Group{
				{
					Count:    3,
					Revision: 2,
					Config: map[string]string{
						"tickTime": "3000",
					},
				},
			},
		}),
	})

	time.Sleep(3 * time.Second)

	//srv.Destroy(0)

	//time.Sleep(3 * time.Second)
}

/*
func TestDeleteNodes(t *testing.T) {
	cases := []struct {
		cluster *proto.Cluster
		num     int
		delete  []string
	}{
		{
			&proto.Cluster{
				Nodes: []*proto.Node{
					{
						ID: "A",
						KV: map[string]string{
							keyRole: roleParticipant,
						},
					},
					{
						ID: "B",
						KV: map[string]string{
							keyRole: roleParticipant,
						},
					},
					{
						ID: "C",
						KV: map[string]string{
							keyRole: roleObserver,
						},
					},
				},
			},
			1,
			[]string{
				"C",
			},
		},
	}

	for _, c := range cases {
		ctx := &operator.PlanCtx{
			Plan: &proto.Plan{
				Sets: []*proto.Plan_Set{
					{
						DelNodesNum: int64(c.num),
					},
				},
			},
			Cluster: c.cluster,
		}
		b := &backend{}
		if err := b.EvaluatePlan(ctx); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(c.delete, ctx.Plan.Sets[0].DelNodes) {
			t.Fatal("bad")
		}
	}
}
*/

package zookeeper

import (
	"reflect"
	"testing"

	"github.com/teseraio/ensemble/operator/proto"
	"github.com/teseraio/ensemble/testutil"
)

func TestZookeeperBootstrap(t *testing.T) {
	srv := testutil.TestOperator(t, Factory)
	defer srv.Close()

	uuid := srv.Apply(&proto.Component{
		Name: "A",
		Spec: proto.MustMarshalAny(&proto.ClusterSpec{
			Backend:  "Zookeeper",
			Replicas: 3,
		}),
	})

	srv.WaitForTask(uuid)
}

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
		p := &proto.Plan{
			Cluster:     c.cluster,
			DelNodesNum: int64(c.num),
		}
		b := &backend{}
		if err := b.EvaluatePlan(p); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(c.delete, p.DelNodes) {
			t.Fatal("bad")
		}
	}
}

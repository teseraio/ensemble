package clickhouse

import (
	"fmt"
	"testing"

	"github.com/teseraio/ensemble/backends/zookeeper"
	"github.com/teseraio/ensemble/operator/proto"
	"github.com/teseraio/ensemble/schema"
	"github.com/teseraio/ensemble/testutil"
)

func TestE2E(t *testing.T) {
	// testutil.IsE2EEnabled(t)

	srv := testutil.TestOperator(t, Factory, zookeeper.Factory)
	// defer srv.Close()

	srv.Apply(&proto.Component{
		Name: "zk1",
		Spec: proto.MustMarshalAny(&proto.ClusterSpec{
			Backend: "Zookeeper",
			Groups: []*proto.ClusterSpec_Group{
				{
					Count: 1,
				},
			},
		}),
	})

	uuid := srv.Apply(&proto.Component{
		Name: "AB",
		Spec: proto.MustMarshalAny(&proto.ClusterSpec{
			Backend: "Clickhouse",
			Groups: []*proto.ClusterSpec_Group{
				{
					Count: 3,
					Params: schema.MapToSpec(
						map[string]interface{}{
							"zookeeper": "zk1",
						},
					),
				},
			},
		}),
	})

	srv.WaitForTask(uuid)

	fmt.Println("___ UPDATE ____")

	uuid = srv.Apply(&proto.Component{
		Name: "AB",
		Spec: proto.MustMarshalAny(&proto.ClusterSpec{
			Backend: "Clickhouse",
			Groups: []*proto.ClusterSpec_Group{
				{
					Count: 4,
					Params: schema.MapToSpec(
						map[string]interface{}{
							"zookeeper": "zk1",
						},
					),
				},
			},
		}),
	})

	srv.WaitForTask(uuid)
}

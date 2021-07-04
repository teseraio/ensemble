package clickhouse

import (
	"testing"

	"github.com/teseraio/ensemble/operator/proto"
	"github.com/teseraio/ensemble/schema"
	"github.com/teseraio/ensemble/testutil"
)

func TestE2E(t *testing.T) {
	// testutil.IsE2EEnabled(t)

	srv := testutil.TestOperator(t, Factory)
	// defer srv.Close()

	uuid := srv.Apply(&proto.Component{
		Name: "AB",
		Spec: proto.MustMarshalAny(&proto.ClusterSpec{
			Backend: "Clickhouse",
			Groups: []*proto.ClusterSpec_Group{
				{
					Count: 3,
					Params: schema.MapToSpec(
						map[string]interface{}{},
					),
				},
			},
			DependsOn: []string{
				"zk1",
			},
		}),
	})

	srv.WaitForTask(uuid)

	uuid = srv.Apply(&proto.Component{
		Name: "AB",
		Spec: proto.MustMarshalAny(&proto.ClusterSpec{
			Backend: "Clickhouse",
			Groups: []*proto.ClusterSpec_Group{
				{
					Count: 4,
					Params: schema.MapToSpec(
						map[string]interface{}{},
					),
				},
			},
			DependsOn: []string{
				"zk1",
			},
		}),
	})

	srv.WaitForTask(uuid)
}

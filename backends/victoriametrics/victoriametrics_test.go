package clickhouse

import (
	"fmt"
	"testing"
	"time"

	"github.com/teseraio/ensemble/operator/proto"
	"github.com/teseraio/ensemble/testutil"
)

func TestE2E(t *testing.T) {
	// testutil.IsE2EEnabled(t)

	srv := testutil.TestOperator(t, Factory)
	// defer srv.Close()

	uuid := srv.Apply(&proto.Component{
		Name: "AB",
		Spec: proto.MustMarshalAny(&proto.ClusterSpec{
			Backend: "VictoriaMetrics",
			Groups: []*proto.ClusterSpec_Group{
				{
					Type:  "storage",
					Count: 1,
				},
				{
					Type:  "insert",
					Count: 1,
				},
				{
					Type:  "select",
					Count: 1,
				},
			},
		}),
	})

	srv.WaitForTask(uuid)

	fmt.Println("_ NEXT _")

	uuid = srv.Apply(&proto.Component{
		Name: "AB",
		Spec: proto.MustMarshalAny(&proto.ClusterSpec{
			Backend: "VictoriaMetrics",
			Groups: []*proto.ClusterSpec_Group{
				{
					Type:  "storage",
					Count: 2,
				},
				{
					Type:  "insert",
					Count: 1,
				},
				{
					Type:  "select",
					Count: 1,
				},
			},
		}),
	})

	srv.WaitForTask(uuid)
	time.Sleep(3 * time.Second)
}

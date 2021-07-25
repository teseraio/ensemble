package victoriametrics

import (
	"testing"

	"github.com/teseraio/ensemble/operator/proto"
	"github.com/teseraio/ensemble/testutil"
)

func TestVictoriaMetrics_Initial(t *testing.T) {
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
}

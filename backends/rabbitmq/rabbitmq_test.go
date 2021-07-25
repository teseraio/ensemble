package rabbitmq

import (
	"testing"

	"github.com/teseraio/ensemble/operator/proto"
	"github.com/teseraio/ensemble/schema"
	"github.com/teseraio/ensemble/testutil"
)

func TestRabbitmq_Initial(t *testing.T) {
	// testutil.IsE2EEnabled(t)

	srv := testutil.TestOperator(t, Factory)
	// defer srv.Close()

	uuid := srv.Apply(&proto.Component{
		Name: "A",
		Spec: proto.MustMarshalAny(&proto.ClusterSpec{
			Backend: "Rabbitmq",
			Groups: []*proto.ClusterSpec_Group{
				{
					Count: 3,
					Params: schema.MapToSpec(map[string]interface{}{
						"cookie": "cookie",
					}),
				},
			},
		}),
	})

	srv.WaitForTask(uuid)

	uuid = srv.Apply(&proto.Component{
		Name: "A",
		Spec: proto.MustMarshalAny(&proto.ClusterSpec{
			Backend: "Rabbitmq",
			Groups: []*proto.ClusterSpec_Group{
				{
					Count: 4,
					Params: schema.MapToSpec(map[string]interface{}{
						"cookie": "cookie",
					}),
				},
			},
		}),
	})

	srv.WaitForTask(uuid)
}

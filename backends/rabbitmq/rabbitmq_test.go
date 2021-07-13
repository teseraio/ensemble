package rabbitmq

import (
	"testing"

	"github.com/teseraio/ensemble/operator/proto"
	"github.com/teseraio/ensemble/testutil"
)

func TestE2E(t *testing.T) {
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
				},
			},
		}),
	})

	srv.WaitForTask(uuid)

	/*
		// Scale up
		uuid = srv.Apply(&proto.Component{
			Name: "A",
			Spec: proto.MustMarshalAny(&proto.ClusterSpec{
				Backend: "Rabbitmq",
				Groups: []*proto.ClusterSpec_Group{
					{
						Count: 4,
					},
				},
			}),
		})

		srv.WaitForTask(uuid)
	*/
}

package cassandra

import (
	"testing"

	"github.com/teseraio/ensemble/operator/proto"
	"github.com/teseraio/ensemble/testutil"
)

func TestBootstrap(t *testing.T) {
	srv := testutil.TestOperator(t, Factory)
	defer srv.Close()

	uuid := srv.Apply(&proto.Component{
		Name: "A",
		Spec: proto.MustMarshalAny(&proto.ClusterSpec{
			Backend: "Cassandra",
			Sets: []*proto.ClusterSpec_Set{
				{Replicas: 2},
			},
		}),
	})

	srv.WaitForTask(uuid)
}

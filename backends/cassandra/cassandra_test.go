package cassandra

import (
	"testing"

	"github.com/teseraio/ensemble/operator/proto"
	"github.com/teseraio/ensemble/testutil"
)

func TestCassandraBootstrap(t *testing.T) {
	srv := testutil.TestOperator(t, Factory)
	defer srv.Close()

	uuid := srv.Apply(&proto.ApplyReq{
		Component: &proto.Component{
			Name: "A",
			Spec: proto.MustMarshalAny(&proto.ClusterSpec{
				Backend:  "Cassandra",
				Replicas: 2,
			}),
		},
	})

	srv.WaitForTask(uuid)
}

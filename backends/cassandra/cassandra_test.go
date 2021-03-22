package cassandra

import (
	"testing"

	"github.com/teseraio/ensemble/operator/proto"
	"github.com/teseraio/ensemble/testutil"
)

func TestBootstrap(t *testing.T) {
	srv := testutil.TestOperator(t, Factory)
	// defer srv.Close()

	uuid := srv.Apply(&proto.Component{
		Name: "A",
		Spec: proto.MustMarshalAny(&proto.ClusterSpec{
			Backend: "Cassandra",
			Groups: []*proto.ClusterSpec_Group{
				{
					Count: 2,
				},
			},
		}),
	})

	srv.WaitForTask(uuid)
}

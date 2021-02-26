package dask

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
			Backend: "Dask",
			Groups: []*proto.ClusterSpec_Group{
				{
					Type:  "scheduler",
					Count: 1,
				},
				{
					Type:  "worker",
					Count: 3,
				},
			},
		}),
	})

	srv.WaitForTask(uuid)
}

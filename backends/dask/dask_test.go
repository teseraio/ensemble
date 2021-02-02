package dask

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
			Backend: "Dask",
			Sets: []*proto.ClusterSpec_Set{
				{
					Type:     "scheduler",
					Replicas: 1,
				},
				{
					Type:     "worker",
					Replicas: 1,
				},
			},
		}),
	})

	srv.WaitForTask(uuid)
}

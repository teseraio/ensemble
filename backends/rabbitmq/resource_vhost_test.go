package rabbitmq

import (
	"testing"

	"github.com/teseraio/ensemble/operator/proto"
	"github.com/teseraio/ensemble/testutil"
)

func TestVHost(t *testing.T) {
	srv := testutil.TestOperator(t, Factory)
	defer srv.Close()

	uuid1 := srv.Apply(&proto.ApplyReq{
		Component: &proto.Component{
			Name: "A",
			Spec: proto.MustMarshalAny(&proto.ClusterSpec{
				Backend:  "Rabbitmq",
				Replicas: 1,
			}),
		},
	})

	srv.WaitForTask(uuid1)

	// create the vhost
	uuid2 := srv.Apply(&proto.ApplyReq{
		Component: &proto.Component{
			Name: "B",
			Spec: proto.MustMarshalAny(&proto.ResourceSpec{
				Cluster:  "A",
				Resource: "VHost",
				Params: `{
				"name": "B"
			}`,
			}),
		},
	})

	srv.WaitForTask(uuid2)
}

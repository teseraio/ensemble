package rabbitmq

import (
	"testing"

	"github.com/teseraio/ensemble/operator/proto"
	"github.com/teseraio/ensemble/schema"
	"github.com/teseraio/ensemble/testutil"
)

func TestVHost(t *testing.T) {
	testutil.IsE2EEnabled(t)

	srv := testutil.TestOperator(t, Factory)
	// defer srv.Close()

	uuid1 := srv.Apply(&proto.Component{
		Name: "A",
		Spec: proto.MustMarshalAny(&proto.ClusterSpec{
			Backend: "Rabbitmq",
			Groups: []*proto.ClusterSpec_Group{
				{Count: 1},
			},
		}),
	})

	srv.WaitForTask(uuid1)

	// create the vhost
	uuid2 := srv.Apply(&proto.Component{
		Name: "B",
		Spec: proto.MustMarshalAny(&proto.ResourceSpec{
			Cluster:  "A",
			Resource: "VHost",
			Params: schema.MapToSpec(map[string]interface{}{
				"name": "B",
			}),
		}),
	})

	srv.WaitForTask(uuid2)

	// force new the vhost
	uuid3 := srv.Apply(&proto.Component{
		Name: "B",
		Spec: proto.MustMarshalAny(&proto.ResourceSpec{
			Cluster:  "A",
			Resource: "VHost",
			Params: schema.MapToSpec(map[string]interface{}{
				"name": "C",
			}),
		}),
	})

	srv.WaitForTask(uuid3)
}

package rabbitmq

import (
	"testing"

	"github.com/teseraio/ensemble/operator/proto"
	"github.com/teseraio/ensemble/testutil"
)

func TestExchange(t *testing.T) {
	srv := testutil.TestOperator(t, Factory)
	defer srv.Close()

	uuid1 := srv.Apply(&proto.Component{
		Name: "A",
		Spec: proto.MustMarshalAny(&proto.ClusterSpec2{
			Name:    "A",
			Backend: "Rabbitmq",
			Groups: []*proto.ClusterSpec2_Group{
				{
					Count: 1,
				},
			},
		}),
	})

	srv.WaitForTask(uuid1)

	// create the vhost
	uuid2 := srv.Apply(&proto.Component{
		Name: "V",
		Spec: proto.MustMarshalAny(&proto.ResourceSpec{
			Cluster:  "A",
			Resource: "VHost",
			Params: `{
						"name": "v"
					}`,
		}),
	})

	srv.WaitForTask(uuid2)

	// create the exchange
	uuid3 := srv.Apply(&proto.Component{
		Name: "E",
		Spec: proto.MustMarshalAny(&proto.ResourceSpec{
			Cluster:  "A",
			Resource: "Exchange",
			Params: `{
						"name": "e",
						"vhost": "v",
						"settings": {
							"type": "fanout"
						}
					}`,
		}),
	})

	srv.WaitForTask(uuid3)
}

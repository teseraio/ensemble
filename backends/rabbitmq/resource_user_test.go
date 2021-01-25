package rabbitmq

import (
	"testing"

	"github.com/teseraio/ensemble/operator/proto"
	"github.com/teseraio/ensemble/testutil"
)

func TestUser(t *testing.T) {
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
				"name": "v"
			}`,
			}),
		},
	})

	srv.WaitForTask(uuid2)

	/*
		provider, _ := testutil.NewTestProvider(t, "rabbitmq", nil)

		srv := operator.TestOperator(t, provider, Factory)
		defer srv.Stop()

		uuid := provider.Apply(&testutil.TestTask{
			Name:  "A",
			Input: `{"replicas": 1}`,
		})
		provider.WaitForTask(uuid)

		// create the user
		uuid = provider.Apply(&testutil.TestTask{
			Name:     "B",
			Resource: "User",
			Input: `{
				"username": "B",
				"password": "xxx"
			}`,
		})
		provider.WaitForTask(uuid)

		// update the password
		uuid = provider.Apply(&testutil.TestTask{
			Name:     "B",
			Resource: "User",
			Input: `{
				"username": "B",
				"password": "yyy"
			}`,
		})
		provider.WaitForTask(uuid)
	*/
}

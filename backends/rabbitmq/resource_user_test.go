package rabbitmq

import (
	"testing"
)

func TestUser(t *testing.T) {
	/*
		srv := testutil.TestOperator(t, Factory)
		defer srv.Close()

		uuid1 := srv.Apply(&proto.Component{
			Name: "A",
			Spec: proto.MustMarshalAny(&proto.ClusterSpec{
				Backend: "Rabbitmq",
				Sets: []*proto.ClusterSpec_Set{
					{Replicas: 1},
				},
			}),
		})

		srv.WaitForTask(uuid1)

		// create the user
		uuid2 := srv.Apply(&proto.Component{
			Name: "B",
			Spec: proto.MustMarshalAny(&proto.ResourceSpec{
				Cluster:  "A",
				Resource: "User",
				Params: `{
					"username": "user",
					"password": "pass"
				}`,
			}),
		})

		srv.WaitForTask(uuid2)

		// change the name of the user
		uuid3 := srv.Apply(&proto.Component{
			Name: "B",
			Spec: proto.MustMarshalAny(&proto.ResourceSpec{
				Cluster:  "A",
				Resource: "User",
				Params: `{
					"username": "user2",
					"password": "pass"
				}`,
			}),
		})

		srv.WaitForTask(uuid3)
	*/
}

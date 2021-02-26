package rabbitmq

import (
	"fmt"
	"testing"

	"github.com/teseraio/ensemble/operator/proto"
	"github.com/teseraio/ensemble/testutil"
)

func TestUser(t *testing.T) {
	srv := testutil.TestOperator(t, Factory)
	defer srv.Close()

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

	fmt.Println("_ RECONCILE _")
	// create the user
	uuid2 := srv.Apply(&proto.Component{
		Name: "B",
		Spec: proto.MustMarshalAny(&proto.ResourceSpec{
			Cluster:  "A",
			Resource: "User",
			Params: `{
					"username": "user1",
					"password": "pass1"
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
}

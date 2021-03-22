package zookeeper

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
			Backend: "Zookeeper",
			Groups: []*proto.ClusterSpec_Group{
				{
					Count: 3,
				},
			},
		}),
	})

	srv.WaitForTask(uuid)

	uuid = srv.Apply(&proto.Component{
		Name: "A",
		Spec: proto.MustMarshalAny(&proto.ClusterSpec{
			Backend: "Zookeeper",
			Groups: []*proto.ClusterSpec_Group{
				{
					Count: 3,
					Config: map[string]string{
						"tickTime": "3000",
					},
				},
			},
		}),
	})

	srv.WaitForTask(uuid)
}

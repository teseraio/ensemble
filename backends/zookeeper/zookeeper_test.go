package zookeeper

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/teseraio/ensemble/operator"
	"github.com/teseraio/ensemble/operator/proto"
	"github.com/teseraio/ensemble/schema"
)

func TestBootstrap(t *testing.T) {

	h := &operator.Harness{
		Deployment: &proto.Deployment{},
		Handler:    Factory(),
		Spec: &proto.ClusterSpec{
			Groups: []*proto.ClusterSpec_Group{
				{
					Count: 3,
					Params: schema.MapToSpec(
						map[string]interface{}{
							// "tickTime": "2000",
						},
					),
				},
			},
		},
	}

	sc := operator.NewScheduler(h)
	assert.NoError(t, sc.Process(&proto.Evaluation{}))

	h.ExpectNodeUpdate([]operator.NodeExpect{
		{
			Env: map[string]string{
				"ZOO_MY_ID":   "1",
				"ZOO_SERVERS": "server.1=0.0.0.0:2888:3888;2181 server.2=8319a035-2:2888:3888;2181 server.3=8e253d74-3:2888:3888;2181",
			},
		},
		{
			Env: map[string]string{
				"ZOO_MY_ID":   "1",
				"ZOO_SERVERS": "server.1=033e1ede-1:2888:3888;2181 server.2=0.0.0.0:2888:3888;2181 server.3=8e253d74-3:2888:3888;2181",
			},
		},
		{
			Env: map[string]string{
				"ZOO_MY_ID":   "1",
				"ZOO_SERVERS": "server.1=033e1ede-1:2888:3888;2181 server.2=8319a035-2:2888:3888;2181 server.3=0.0.0.0:2888:3888;2181",
			},
		},
	})

	/*
		for _, i := range h.Plan.NodeUpdate {
			fmt.Println("-- i --")
			fmt.Println(i.Spec)
		}
	*/

	/*
		srv := testutil.TestOperator(t, Factory)
		// defer srv.Close()

		uuid := srv.Apply(&proto.Component{
			Name: "A",
			Spec: proto.MustMarshalAny(&proto.ClusterSpec{
				Backend: "Zookeeper",
				Groups: []*proto.ClusterSpec_Group{
					{
						Count: 3,
						Params: schema.MapToSpec(
							map[string]interface{}{
								"tickTime": "2000",
							},
						),
						Resources: schema.MapToSpec(
							map[string]interface{}{
								"cpuShares": "1000",
							},
						),
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
						Params: schema.MapToSpec(
							map[string]interface{}{
								"tickTime": "3000",
							},
						),
					},
				},
			}),
		})

		srv.WaitForTask(uuid)
	*/
}

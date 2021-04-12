package rabbitmq

/*
func TestVHost(t *testing.T) {
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

	// create the vhost
	uuid2 := srv.Apply(&proto.Component{
		Name: "B",
		Spec: proto.MustMarshalAny(&proto.ResourceSpec{
			Cluster:  "A",
			Resource: "VHost",
			Params: `{
				"name": "B"
			}`,
		}),
	})

	srv.WaitForTask(uuid2)
}
*/

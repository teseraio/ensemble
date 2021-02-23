package operator

import (
	"fmt"
	"testing"

	"github.com/teseraio/ensemble/operator/proto"
)

func TestReconcile(t *testing.T) {
	r := &reconciler2{
		dep: &proto.Deployment{
			Instances: []*proto.Instance{
				{
					Healthy: true,
					Group: &proto.ClusterSpec2_Group{
						Revision: 3,
					},
				},
			},
		},
		spec: &proto.ClusterSpec2{
			Groups: []*proto.ClusterSpec2_Group{
				{
					Count:    5,
					Revision: 3,
				},
			},
		},
	}
	r.Compute()
	for _, i := range r.res {
		fmt.Println(i.status, i.instance)
	}
}

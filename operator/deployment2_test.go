package operator

import (
	"fmt"
	"testing"

	"github.com/teseraio/ensemble/operator/proto"
)

func TestReconciler(t *testing.T) {
	r := &reconciler2{
		dep: &proto.Deployment{
			Instances: []*proto.Instance{
				{
					Revision: 3,
					Healthy:  true,
				},
				{
					Revision: 1,
				},
			},
		},
		spec: &proto.ClusterSpec2{
			Group: &proto.ClusterSpec2_Group{
				Count:    5,
				Revision: 3,
			},
		},
	}
	r.Compute()
	for _, i := range r.res {
		fmt.Println(i.status, i.instance)
	}
}

package operator

import (
	"fmt"
	"testing"

	"github.com/teseraio/ensemble/lib/uuid"
	"github.com/teseraio/ensemble/operator/proto"
)

func TestReconcileX(t *testing.T) {
	r := &reconciler{
		dep: &proto.Deployment{
			Instances: []*proto.Instance{
				{
					Healthy:  true,
					Group:    &proto.ClusterSpec_Group{},
					Sequence: 2,
				},
			},
		},
		spec: &proto.ClusterSpec{
			Groups: []*proto.ClusterSpec_Group{
				{
					Count: 5,
				},
			},
			Sequence: 2,
		},
	}
	r.Compute()
	for _, i := range r.res {
		fmt.Println(i.status, i.instance)
	}
}

func TestReconcileGroups(t *testing.T) {
	r := &reconciler{
		dep: &proto.Deployment{
			Instances: []*proto.Instance{
				{
					ID:      uuid.UUID(),
					Healthy: false,
					Group: &proto.ClusterSpec_Group{
						Type: "x",
					},
				},
			},
		},
		spec: &proto.ClusterSpec{
			Name: "cluster",
			Groups: []*proto.ClusterSpec_Group{
				{
					Type:  "x",
					Count: 1,
				},
				{
					Type:  "y",
					Count: 3,
				},
			},
		},
	}
	r.Compute()
	r.print()
}

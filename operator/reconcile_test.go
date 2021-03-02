package operator

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/teseraio/ensemble/lib/uuid"
	"github.com/teseraio/ensemble/operator/proto"
)

func TestReconcileScaleDown(t *testing.T) {
	r := &reconciler{
		dep: &proto.Deployment{
			Instances: []*proto.Instance{
				{
					Healthy: true,
					Group:   &proto.ClusterSpec_Group{},
				},
				{
					Healthy: true,
					Group:   &proto.ClusterSpec_Group{},
				},
			},
		},
		spec: &proto.ClusterSpec{
			Groups: []*proto.ClusterSpec_Group{
				{
					Count: 1,
				},
			},
		},
	}
	r.Compute()
	assert.Equal(t, r.check("stop"), 1)
}

func TestReconcileScaleUp(t *testing.T) {
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
	assert.Equal(t, r.check("add"), 4)
}

func TestReconcileGroups_CompleteFirstGroup(t *testing.T) {
	r := &reconciler{
		dep: &proto.Deployment{
			Instances: []*proto.Instance{},
		},
		spec: &proto.ClusterSpec{
			Name: "cluster",
			Groups: []*proto.ClusterSpec_Group{
				{
					Type:  "group1",
					Count: 1,
				},
				{
					Type:  "group2",
					Count: 3,
				},
			},
		},
	}
	r.Compute()
	assert.Equal(t, r.check("add"), 1)
}

func TestReconcileGroups_CompleteSecondGroup(t *testing.T) {
	r := &reconciler{
		dep: &proto.Deployment{
			Instances: []*proto.Instance{
				{
					ID:      uuid.UUID(),
					Healthy: true,
					Group: &proto.ClusterSpec_Group{
						Name: "group1",
					},
				},
			},
		},
		spec: &proto.ClusterSpec{
			Name: "cluster",
			Groups: []*proto.ClusterSpec_Group{
				{
					Type:  "group1",
					Count: 1,
				},
				{
					Type:  "group2",
					Count: 3,
				},
			},
		},
	}
	r.Compute()
	assert.Equal(t, r.check("add"), 1)
}

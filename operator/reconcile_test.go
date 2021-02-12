package operator

import (
	"fmt"
	"testing"

	"github.com/teseraio/ensemble/operator/proto"
)

func TestReconcile(t *testing.T) {
	a := &allocReconciler{
		c: &proto.Cluster{
			Groups: []*proto.Group{
				{
					Count:    2,
					Nodeset:  "a",
					Revision: 5,
				},
			},
		},
		nodes: []*proto.Instance{
			{
				ID:       "a",
				Group:    "a",
				Revision: 4,
			},
			{
				ID:       "b",
				Group:    "a",
				Revision: 4,
			},
		},
	}
	a.reconcile()
	fmt.Println(a.result)
}

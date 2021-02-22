package operator

import (
	"context"
	"testing"

	"github.com/teseraio/ensemble/operator/proto"
)

func TestApply(t *testing.T) {
	c := &proto.Component{
		Id: "A",
		Spec: proto.MustMarshalAny(&proto.ClusterSpec{
			Groups: []*proto.ClusterSpec_Group{
				{
					Count: 3,
				},
			},
		}),
	}

	s := &service{}
	s.Apply(context.Background(), c)
}

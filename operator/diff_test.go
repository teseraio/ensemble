package operator

import (
	"testing"

	"github.com/teseraio/ensemble/operator/proto"
	"github.com/teseraio/ensemble/schema"
)

func TestDiff(t *testing.T) {
	emptySchema := &schema.Schema2{
		Spec: &schema.Record{},
	}

	ff := buildDiffFunc(emptySchema, emptySchema, emptySchema)

	old := &proto.ClusterSpec_Group{
		// Params: map[string]string{},
	}
	new := &proto.ClusterSpec_Group{}

	ff(new, old)
}

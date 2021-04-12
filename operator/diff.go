package operator

import (
	"github.com/teseraio/ensemble/operator/proto"
	"github.com/teseraio/ensemble/schema"
)

func buildDiffFunc(params, resource, storage *schema.Schema2) updateFn {
	return func(new, old *proto.ClusterSpec_Group) bool {
		return false
	}
}

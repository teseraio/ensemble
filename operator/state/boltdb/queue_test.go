package boltdb

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/teseraio/ensemble/operator/proto"
)

func TestQueueSerializeByClusterID(t *testing.T) {
	q := newTaskQueue()

	q.add("A", &proto.Component{
		Id:   "id1",
		Spec: proto.MustMarshalAny(&proto.ClusterSpec{}),
	})

	q.add("A", &proto.Component{
		Id:   "id2",
		Spec: proto.MustMarshalAny(&proto.ClusterSpec{}),
	})

	q.add("A", &proto.Component{
		Id:   "id3",
		Spec: proto.MustMarshalAny(&proto.ResourceSpec{}),
	})

	assert.Equal(t, q.popImpl().Id, "id1")
	assert.Nil(t, q.popImpl())

	q.finalize("A")

	assert.Equal(t, q.popImpl().Id, "id2")

	/*
		assert.Nil(t, q.popImpl())

		q.add("B", &proto.Component{
			Id:   "id4",
			Spec: proto.MustMarshalAny(&proto.ResourceSpec{}),
		})

		assert.Equal(t, q.popImpl().Id, "id4")
		assert.Nil(t, q.popImpl())

		q.finalize("A")
		q.finalize("A")

		assert.Equal(t, q.popImpl().Id, "id3")
		assert.Nil(t, q.popImpl())
	*/
}

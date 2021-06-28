package boltdb

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/teseraio/ensemble/operator/proto"
)

func TestQueue(t *testing.T) {
	q := newTaskQueue()

	assert.Nil(t, q.popImpl())

	q.add(&proto.Task{
		DeploymentID: "id1",
	})

	assert.NotNil(t, q.popImpl())
	assert.Nil(t, q.popImpl())

	q.add(&proto.Task{
		DeploymentID: "id2",
	})

	assert.NotNil(t, q.popImpl())

	_, ok := q.finalize("id1")
	assert.True(t, ok)

	assert.Nil(t, q.popImpl())
}

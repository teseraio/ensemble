package boltdb

import (
	"context"
	"testing"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/teseraio/ensemble/lib/uuid"
	"github.com/teseraio/ensemble/operator/proto"
)

func TestTaskQueue(t *testing.T) {
	tt := newTaskQueue()

	mockTask := func() *proto.Component {
		timestamp := ptypes.TimestampNow()
		return &proto.Component{
			Id:        uuid.UUID(),
			Timestamp: timestamp,
		}
	}

	t0 := mockTask()
	t1 := mockTask()
	t2 := mockTask()

	tt.add(t0, nil)
	tt.add(t1, nil)
	tt.add(t2, nil)

	ctx := context.Background()

	if t0e := tt.pop(ctx); t0e.New.Id != t0.Id {
		t.Fatal("bad")
	}
	if t1e := tt.pop(ctx); t1e.New.Id != t1.Id {
		t.Fatal("bad")
	}
	if t2e := tt.pop(ctx); t2e.New.Id != t2.Id {
		t.Fatal("bad")
	}

	// it blocks if there are no more items till a new one arrives
	t3 := mockTask()

	taskCh := make(chan *task)
	go func() {
		taskCh <- tt.pop(ctx)
	}()

	select {
	case <-taskCh:
		t.Fatal("bad")
	case <-time.After(100 * time.Millisecond):
	}

	tt.add(t3, nil)

	select {
	case t3e := <-taskCh:
		if t3e.New.Id != t3.Id {
			t.Fatal("bad")
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("bad")
	}

	// Once finalized the task is removed
	if len(tt.heap) != 4 {
		t.Fatal("bad")
	}
	tt.finalize(t0.Id)

	if len(tt.heap) != 3 {
		t.Fatal("bad")
	}
}

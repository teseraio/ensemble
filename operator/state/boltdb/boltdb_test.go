package boltdb

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/golang/protobuf/ptypes/any"
	"github.com/teseraio/ensemble/lib/uuid"
	"github.com/teseraio/ensemble/operator/proto"
	"github.com/teseraio/ensemble/operator/state"
)

func setupFn(t *testing.T) (state.State, func()) {
	path := "/tmp/db-" + uuid.UUID()

	st, err := Factory(map[string]interface{}{
		"path": path,
	})
	if err != nil {
		t.Fatal(err)
	}
	closeFn := func() {
		if err := os.Remove(path); err != nil {
			t.Fatal(err)
		}
	}
	return st, closeFn
}

func TestSuite(t *testing.T) {
	state.TestSuite(t, setupFn)
}

func TestIndexUpdate(t *testing.T) {
	config := map[string]interface{}{
		"path": "/tmp/db-" + uuid.UUID(),
	}
	st, err := Factory(config)
	if err != nil {
		t.Fatal(err)
	}

	ca1 := &proto.Component{
		Id:     uuid.UUID(),
		Name:   "Item",
		Status: proto.Component_PENDING,
		Spec: &any.Any{
			TypeUrl: "1",
			Value:   []byte{0x1},
		},
	}
	if err := st.Apply(ca1); err != nil {
		t.Fatal(err)
	}

	ca2 := &proto.Component{
		Id:     uuid.UUID(),
		Name:   "Item",
		Status: proto.Component_PENDING,
		Spec: &any.Any{
			TypeUrl: "2",
			Value:   []byte{0x2},
		},
	}
	if err := st.Apply(ca2); err != nil {
		t.Fatal(err)
	}

	// it must return the first version
	cb, _ := st.GetTask(context.Background())
	if cb.New.Spec.TypeUrl != "1" {
		t.Fatal("bad")
	}

	// GetTask should block because task 2 of A only
	// be indexed after task 1 is done
	ctx, _ := context.WithDeadline(context.Background(), time.Now().Add(time.Second))
	cb, _ = st.GetTask(ctx)

	if cb != nil {
		t.Fatal("bad")
	}

	// Finalize the task 1
	if err := st.Finalize(ca1.Id); err != nil {
		t.Fatal(err)
	}
	// now we can retrieve the task 2
	cb, _ = st.GetTask(context.Background())
	if cb.New.Spec.TypeUrl != "2" {
		t.Fatal("bad")
	}
}

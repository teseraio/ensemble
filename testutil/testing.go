package testutil

import (
	"fmt"
	"testing"
	"time"

	"github.com/teseraio/ensemble/lib/uuid"
	"github.com/teseraio/ensemble/operator"
	"github.com/teseraio/ensemble/operator/proto"
)

// testing suite for the Provider
func TestProvider(t *testing.T, p operator.Provider) {
	//t.Run("", func(t *testing.T) {
	//TestPodLifecycle(t, p)
	//})
	//t.Run("", func(t *testing.T) {
	//TestPodBarArgs(t, p)
	//})
	//TestPodJobFailed(t, p)
	// TestDNS
	TestPodFiles(t, p)
}

func readEvent(p operator.Provider, t *testing.T) *proto.InstanceUpdate {
	select {
	case evnt := <-p.WatchUpdates():
		return evnt
	case <-time.After(10 * time.Second):
	}
	t.Fatal("timeout")
	return nil
}

func TestPodBarArgs(t *testing.T, p operator.Provider) {
	i := &proto.Instance{
		ID:      uuid.UUID(),
		Cluster: "xx11",
		Name:    "yy22",
		Spec: &proto.NodeSpec{
			Image: "busybox",
			Cmd:   []string{"xxx"},
		},
	}
	if _, err := p.CreateResource(i); err != nil {
		t.Fatal(err)
	}

	// the pod is scheduled
	evnt := readEvent(p, t)
	if _, ok := evnt.Event.(*proto.InstanceUpdate_Scheduled_); !ok {
		t.Fatal("expected scheduled")
	}

	// the pod fails
	evnt = readEvent(p, t)
	if _, ok := evnt.Event.(*proto.InstanceUpdate_Failed_); !ok {
		t.Fatal("expected failed")
	}
}

func TestPodJobFailed(t *testing.T, p operator.Provider) {
	// TODO
	i := &proto.Instance{
		ID:      uuid.UUID(),
		Cluster: "xx11",
		Name:    "yy22",
		Spec: &proto.NodeSpec{
			Image: "busybox",
			// it stops gracefully
			Cmd: []string{"sleep", "2"},
		},
	}
	if _, err := p.CreateResource(i); err != nil {
		t.Fatal(err)
	}

	// the pod is scheduled
	evnt := readEvent(p, t)
	if _, ok := evnt.Event.(*proto.InstanceUpdate_Scheduled_); !ok {
		t.Fatal("expected scheduled")
	}

	time.Sleep(10 * time.Second)
}

func TestPodLifecycle(t *testing.T, p operator.Provider) {
	id := uuid.UUID()

	i := &proto.Instance{
		ID:      id,
		Cluster: "c11",
		Name:    "d22",
		Spec: &proto.NodeSpec{
			Image: "nginx",
		},
	}

	if _, err := p.CreateResource(i); err != nil {
		t.Fatal(err)
	}

	// wait for the container to be running
	evnt := readEvent(p, t)
	if _, ok := evnt.Event.(*proto.InstanceUpdate_Scheduled_); !ok {
		t.Fatal("expected scheduled")
	}

	evnt = readEvent(p, t)
	obj, ok := evnt.Event.(*proto.InstanceUpdate_Running_)
	if !ok {
		t.Fatal("expected running")
	}
	i.Handler = obj.Running.Handler

	/*
		// try to create the same container again
		if _, err := p.CreateResource(i); err != nil {
			t.Fatal(err)
		}
	*/

	if _, err := p.DeleteResource(i); err != nil {
		t.Fatal(err)
	}

	// wait for termination event
	evnt = readEvent(p, t)
	if _, ok := evnt.Event.(*proto.InstanceUpdate_Killing_); !ok {
		t.Fatal("expected stopped")
	}
}

func TestPodFiles(t *testing.T, p operator.Provider) {
	id := uuid.UUID()

	i := &proto.Instance{
		ID:      id,
		Cluster: "c11",
		Name:    "d22",
		Spec: &proto.NodeSpec{
			Image: "nginx",
			Files2: []*proto.NodeSpec_File{
				{
					Name:    "/a/b/c.txt",
					Content: "abcd",
				},
			},
		},
		Mounts: []*proto.Instance_Mount{
			{
				Name: "data",
				Path: "/data",
			},
		},
	}

	if _, err := p.CreateResource(i); err != nil {
		t.Fatal(err)
	}

	// wait for it to be ready
	for {
		evnt := <-p.WatchUpdates()
		if _, ok := evnt.Event.(*proto.InstanceUpdate_Running_); ok {
			break
		}
	}
	fmt.Println("- ready -")
}

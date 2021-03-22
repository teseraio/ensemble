package testutil

import (
	"testing"
	"time"

	"github.com/teseraio/ensemble/lib/uuid"
	"github.com/teseraio/ensemble/operator"
	"github.com/teseraio/ensemble/operator/proto"
)

// testing suite for the Provider
func TestProvider(t *testing.T, p operator.Provider) {
	//TestPodLifecycle(t, p)
	//TestPodBarArgs(t, p)
	//TestPodJobFailed(t, p)
	//TODO: TestDNS
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

	if _, err := p.DeleteResource(i); err != nil {
		t.Fatal(err)
	}

	// wait for termination event
	evnt = readEvent(p, t)
	if _, ok := evnt.Event.(*proto.InstanceUpdate_Killing_); !ok {
		t.Fatal("expected stopped")
	}
}

package testutil

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/teseraio/ensemble/lib/uuid"
	"github.com/teseraio/ensemble/operator"
	"github.com/teseraio/ensemble/operator/proto"
)

// testing suite for the Provider
func TestProvider(t *testing.T, p operator.Provider) {
	t.Run("TestPodLifecycle", func(t *testing.T) {
		TestPodLifecycle(t, p)
	})
	t.Run("TestDNS", func(t *testing.T) {
		TestDNS(t, p)
	})
	t.Run("TestPodMount", func(t *testing.T) {
		TestPodMount(t, p)
	})
	t.Run("TestPodFiles", func(t *testing.T) {
		TestPodFiles(t, p)
	})
	// TODO
	//TestPodBarArgs(t, p)
	//TestPodJobFailed(t, p)
}

func readEvent(p operator.Provider, t *testing.T) *proto.InstanceUpdate {
	select {
	case evnt := <-p.WatchUpdates():
		return evnt
	case <-time.After(20 * time.Second):
	}
	t.Fatal("timeout")
	return nil
}

func waitForRunning(p operator.Provider, t *testing.T) *proto.InstanceUpdate_Running_ {
	evnt := waitForEvent(p, t, func(evnt *proto.InstanceUpdate) bool {
		_, ok := evnt.Event.(*proto.InstanceUpdate_Running_)
		return ok
	})
	return evnt.Event.(*proto.InstanceUpdate_Running_)
}

func waitForDeleted(p operator.Provider, t *testing.T) {
	waitForEvent(p, t, func(evnt *proto.InstanceUpdate) bool {
		_, ok := evnt.Event.(*proto.InstanceUpdate_Killing_)
		return ok
	})
}

func waitForEvent(p operator.Provider, t *testing.T, handler func(i *proto.InstanceUpdate) bool) *proto.InstanceUpdate {
	doneCh := make(chan struct{})
	go func() {
		<-time.After(20 * time.Second)
		close(doneCh)
	}()
	for {
		select {
		case evnt := <-p.WatchUpdates():
			if handler(evnt) {
				return evnt
			}
		case <-doneCh:
			t.Fatal("timeout")
		}
	}
}

func TestPodBarArgs(t *testing.T, p operator.Provider) {
	// TODO
	i := &proto.Instance{
		ID:          uuid.UUID(),
		ClusterName: "xx11",
		Name:        "yy22",
		Image:       "busybox",
		Spec: &proto.NodeSpec{
			Cmd: "xxx",
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
		ID:          uuid.UUID(),
		ClusterName: "xx11",
		Name:        "yy22",
		Image:       "busybox",
		Spec: &proto.NodeSpec{
			// it stops gracefully
			Cmd:  "sleep",
			Args: []string{"2"},
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
		ID:          id,
		ClusterName: "c11",
		Name:        "d22",
		Image:       "nginx",
		Spec:        &proto.NodeSpec{},
	}

	if _, err := p.CreateResource(i); err != nil {
		t.Fatal(err)
	}

	// wait for running event
	evnt := waitForEvent(p, t, func(evnt *proto.InstanceUpdate) bool {
		_, ok := evnt.Event.(*proto.InstanceUpdate_Running_)
		return ok
	})

	i.Handler = evnt.Event.(*proto.InstanceUpdate_Running_).Running.Handler
	if _, err := p.DeleteResource(i); err != nil {
		t.Fatal(err)
	}

	// wait for termination event
	waitForEvent(p, t, func(evnt *proto.InstanceUpdate) bool {
		_, ok := evnt.Event.(*proto.InstanceUpdate_Killing_)
		return ok
	})
}

func TestPodFiles(t *testing.T, p operator.Provider) {
	id := uuid.UUID()

	files := []*proto.NodeSpec_File{
		{
			Name:    "/a/b/c.txt",
			Content: "abcd",
		},
		{
			Name:    "/a/d.txt",
			Content: "efgh",
		},
		{
			Name: "/a/e.txt",
			Content: `Line1
Line2 "a"
Line3`,
		},
	}
	i := &proto.Instance{
		ID:          id,
		ClusterName: "c11",
		Name:        uuid.UUID(),
		Image:       "nginx",
		Spec: &proto.NodeSpec{
			Files: files,
		},
	}

	if _, err := p.CreateResource(i); err != nil {
		t.Fatal(err)
	}

	// wait for it to be ready
	waitForRunning(p, t)

	for _, file := range files {
		out, err := p.Exec(id, "cat", file.Name)
		assert.NoError(t, err)
		assert.Equal(t, out, file.Content)
	}
}

func TestPodMount(t *testing.T, p operator.Provider) {
	id := uuid.UUID()

	i := &proto.Instance{
		ID:          id,
		ClusterName: "c11",
		Name:        uuid.UUID(),
		Image:       "nginx",
		Spec:        &proto.NodeSpec{},
		Mounts: []*proto.Instance_Mount{
			{
				Name: "one",
				Path: "/data",
			},
		},
	}

	if _, err := p.CreateResource(i); err != nil {
		t.Fatal(err)
	}

	// wait for it to be ready
	evnt := waitForRunning(p, t)
	i.Handler = evnt.Running.Handler

	// /data/test.txt does not exists
	_, err := p.Exec(i.ID, "cat", "/data/test.txt")
	assert.Error(t, err)

	if _, err := p.Exec(i.ID, "touch", "/data/test.txt"); err != nil {
		t.Fatal(err)
	}

	// stop the container
	if _, err := p.DeleteResource(i); err != nil {
		t.Fatal(err)
	}

	// wait for the container to stop
	waitForDeleted(p, t)

	// "restart" the instance
	ii := i.Copy()
	ii.ID = uuid.UUID()

	if _, err := p.CreateResource(ii); err != nil {
		t.Fatal(err)
	}

	// wait for it to be ready
	evnt = waitForRunning(p, t)
	i.Handler = evnt.Running.Handler

	// /data/test.txt should be available
	_, err = p.Exec(ii.ID, "cat", "/data/test.txt")
	assert.NoError(t, err)
}

func TestDNS(t *testing.T, p operator.Provider) {
	target := &proto.Instance{
		ID:          uuid.UUID(),
		ClusterName: "c11",
		Name:        uuid.UUID(),
		Image:       "nginx",
		Spec:        &proto.NodeSpec{},
	}

	if _, err := p.CreateResource(target); err != nil {
		t.Fatal(err)
	}

	// wait for it to be ready
	evnt := waitForRunning(p, t)
	target.Handler = evnt.Running.Handler

	source := &proto.Instance{
		ID:          uuid.UUID(),
		ClusterName: "c11",
		Name:        uuid.UUID(),
		Image:       "nginx",
		Spec:        &proto.NodeSpec{},
	}

	if _, err := p.CreateResource(source); err != nil {
		t.Fatal(err)
	}

	// wait for it to be ready
	evnt = waitForRunning(p, t)
	source.Handler = evnt.Running.Handler

	// valid dns
	out, err := p.Exec(source.ID, "curl", "--fail", "--silent", "--show-error", target.Name+".c11")
	assert.NoError(t, err)
	assert.True(t, strings.HasPrefix(out, "<!DOCTYPE html>"))

	// invalid dns
	_, err = p.Exec(source.ID, "curl", "--fail", "--silent", "--show-error", target.Name+".c12")
	assert.Error(t, err)
}

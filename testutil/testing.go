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

func TestPodBarArgs(t *testing.T, p operator.Provider) {
	// TODO
	i := &proto.Instance{
		ID:      uuid.UUID(),
		Cluster: "xx11",
		Name:    "yy22",
		Image:   "busybox",
		Spec: &proto.NodeSpec{
			Cmd: []string{"xxx"},
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
		Image:   "busybox",
		Spec: &proto.NodeSpec{
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
		Image:   "nginx",
		Spec:    &proto.NodeSpec{},
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
Line2
Line3`,
		},
	}
	i := &proto.Instance{
		ID:      id,
		Cluster: "c11",
		Name:    uuid.UUID(),
		Image:   "nginx",
		Spec: &proto.NodeSpec{
			Files: files,
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

	for _, file := range files {
		out, err := p.Exec(id, "cat", file.Name)
		assert.NoError(t, err)
		assert.Equal(t, out, file.Content)
	}
}

func TestPodMount(t *testing.T, p operator.Provider) {
	id := uuid.UUID()

	i := &proto.Instance{
		ID:      id,
		Cluster: "c11",
		Name:    uuid.UUID(),
		Image:   "nginx",
		Spec:    &proto.NodeSpec{},
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
	for {
		evnt := <-p.WatchUpdates()
		if obj, ok := evnt.Event.(*proto.InstanceUpdate_Running_); ok {
			i.Handler = obj.Running.Handler
			break
		}
	}

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
	for {
		evnt := <-p.WatchUpdates()
		if _, ok := evnt.Event.(*proto.InstanceUpdate_Killing_); ok {
			break
		}
	}

	// "restart" the instance
	ii := i.Copy()
	ii.ID = uuid.UUID()

	if _, err := p.CreateResource(ii); err != nil {
		t.Fatal(err)
	}

	// wait for it to be ready
	for {
		evnt := <-p.WatchUpdates()
		if obj, ok := evnt.Event.(*proto.InstanceUpdate_Running_); ok {
			ii.Handler = obj.Running.Handler
			break
		}
	}

	// /data/test.txt should be available
	_, err = p.Exec(ii.ID, "cat", "/data/test.txt")
	assert.NoError(t, err)
}

func TestDNS(t *testing.T, p operator.Provider) {
	target := &proto.Instance{
		ID:      uuid.UUID(),
		Cluster: "c11",
		Name:    uuid.UUID(),
		Image:   "nginx",
		Spec:    &proto.NodeSpec{},
	}

	if _, err := p.CreateResource(target); err != nil {
		t.Fatal(err)
	}

	// wait for it to be ready
	for {
		evnt := <-p.WatchUpdates()
		if obj, ok := evnt.Event.(*proto.InstanceUpdate_Running_); ok {
			target.Handler = obj.Running.Handler
			break
		}
	}

	source := &proto.Instance{
		ID:      uuid.UUID(),
		Cluster: "c11",
		Name:    uuid.UUID(),
		Image:   "nginx",
		Spec:    &proto.NodeSpec{},
	}

	if _, err := p.CreateResource(source); err != nil {
		t.Fatal(err)
	}

	// wait for it to be ready
	for {
		evnt := <-p.WatchUpdates()
		if obj, ok := evnt.Event.(*proto.InstanceUpdate_Running_); ok {
			source.Handler = obj.Running.Handler
			break
		}
	}

	// valid dns
	out, err := p.Exec(source.ID, "curl", "--fail", "--silent", "--show-error", target.Name+".c11")
	assert.NoError(t, err)
	assert.True(t, strings.HasPrefix(out, "<!DOCTYPE html>"))

	// invalid dns
	_, err = p.Exec(source.ID, "curl", "--fail", "--silent", "--show-error", target.Name+".c12")
	assert.Error(t, err)
}

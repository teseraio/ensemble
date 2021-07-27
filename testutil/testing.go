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
	c := &operator.InmemControlPlane{}
	p.Setup(c)

	t.Run("TestPodLifecycle", func(t *testing.T) {
		TestPodLifecycle(t, c, p)
	})
	t.Run("TestDNS", func(t *testing.T) {
		TestDNS(t, c, p)
	})
	t.Run("TestPodMount", func(t *testing.T) {
		TestPodMount(t, c, p)
	})
	t.Run("TestPodFiles", func(t *testing.T) {
		TestPodFiles(t, c, p)
	})
	// TODO
	//TestPodBarArgs(t, p)
	//TestPodJobFailed(t, p)
}

func readEvent(p operator.ControlPlane, t *testing.T) *proto.Instance {
	ch := p.SubscribeInstanceUpdates()

	select {
	case msg := <-ch:
		instance, err := p.GetInstance(msg.Id, msg.Cluster)
		if err != nil {
			t.Fatal(err)
		}
		return instance
	case <-time.After(10 * time.Second):
		t.Fatal("timeout")
	}
	return nil
}

func waitForRunning(c operator.ControlPlane, t *testing.T) *proto.Instance {
	return waitForEvent(c, t, func(evnt *proto.Instance) bool {
		return evnt.Status == proto.Instance_RUNNING
	})
}

func waitForStopped(c operator.ControlPlane, t *testing.T) *proto.Instance {
	return waitForEvent(c, t, func(evnt *proto.Instance) bool {
		return evnt.Status == proto.Instance_STOPPED
	})
}

func waitForEvent(c operator.ControlPlane, t *testing.T, handler func(i *proto.Instance) bool) *proto.Instance {
	doneCh := make(chan struct{})
	go func() {
		<-time.After(20 * time.Second)
		close(doneCh)
	}()
	evnts := c.SubscribeInstanceUpdates()
	for {
		select {
		case evnt := <-evnts:
			instance, err := c.GetInstance(evnt.Id, evnt.Cluster)
			if err != nil {
				t.Fatal(err)
			}
			if handler(instance) {
				return instance.Copy()
			}
		case <-doneCh:
			t.Fatal("timeout")
		}
	}
}

/*
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
*/

func TestPodLifecycle(t *testing.T, c operator.ControlPlane, p operator.Provider) {
	id := uuid.UUID()

	i := &proto.Instance{
		ID:           id,
		DeploymentID: "c11",
		ClusterName:  "c11",
		Name:         "d22",
		Image:        "nginx",
		Spec:         &proto.NodeSpec{},
		Status:       proto.Instance_PENDING,
	}
	if err := c.UpsertInstance(i); err != nil {
		t.Fatal(err)
	}

	// wait for running event
	i = waitForRunning(c, t)

	i.Status = proto.Instance_TAINTED
	if err := c.UpsertInstance(i); err != nil {
		t.Fatal(err)
	}

	// wait for termination event
	waitForStopped(c, t)
}

func TestPodFiles(t *testing.T, c operator.ControlPlane, p operator.Provider) {
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
		ID:           id,
		DeploymentID: "c11",
		ClusterName:  "c11",
		Name:         uuid.UUID(),
		Image:        "nginx",
		Spec: &proto.NodeSpec{
			Files: files,
		},
		Status: proto.Instance_PENDING,
	}
	if err := c.UpsertInstance(i); err != nil {
		t.Fatal(err)
	}

	// wait for it to be ready
	waitForRunning(c, t)

	for _, file := range files {
		out, err := p.Exec(id, "cat", file.Name)
		assert.NoError(t, err)
		assert.Equal(t, out, file.Content)
	}
}

func TestPodMount(t *testing.T, c operator.ControlPlane, p operator.Provider) {
	id := uuid.UUID()

	initial := &proto.Instance{
		ID:           id,
		DeploymentID: "c11",
		ClusterName:  "c11",
		Name:         uuid.UUID(),
		Image:        "nginx",
		Spec:         &proto.NodeSpec{},
		Mounts: []*proto.Instance_Mount{
			{
				Name: "one",
				Path: "/data",
			},
		},
		Status: proto.Instance_PENDING,
	}
	if err := c.UpsertInstance(initial); err != nil {
		t.Fatal(err)
	}

	// wait for it to be ready
	i := waitForRunning(c, t)

	// /data/test.txt does not exists
	_, err := p.Exec(i.ID, "cat", "/data/test.txt")
	assert.Error(t, err)

	if _, err := p.Exec(i.ID, "touch", "/data/test.txt"); err != nil {
		t.Fatal(err)
	}

	// stop the container
	i.Status = proto.Instance_TAINTED
	if err := c.UpsertInstance(i); err != nil {
		t.Fatal(err)
	}

	// wait for the container to stop
	waitForStopped(c, t)

	// "restart" the instance
	initial = initial.Copy()
	initial.ID = uuid.UUID()

	if err := c.UpsertInstance(initial); err != nil {
		t.Fatal(err)
	}

	// wait for it to be ready
	ii := waitForRunning(c, t)

	// /data/test.txt should be available
	_, err = p.Exec(ii.ID, "cat", "/data/test.txt")
	assert.NoError(t, err)
}

func TestDNS(t *testing.T, c operator.ControlPlane, p operator.Provider) {
	target := &proto.Instance{
		ID:           uuid.UUID(),
		DeploymentID: "c11",
		ClusterName:  "c11",
		Name:         uuid.UUID(),
		Image:        "nginx",
		Spec:         &proto.NodeSpec{},
		Status:       proto.Instance_PENDING,
	}
	if err := c.UpsertInstance(target); err != nil {
		t.Fatal(err)
	}

	// wait for it to be ready
	target = waitForRunning(c, t)

	source := &proto.Instance{
		ID:           uuid.UUID(),
		DeploymentID: "c11",
		ClusterName:  "c11",
		Name:         uuid.UUID(),
		Image:        "nginx",
		Spec:         &proto.NodeSpec{},
		Status:       proto.Instance_PENDING,
	}
	if err := c.UpsertInstance(source); err != nil {
		t.Fatal(err)
	}

	// wait for it to be ready
	source = waitForRunning(c, t)

	// valid dns
	out, err := p.Exec(source.ID, "curl", "--fail", "--silent", "--show-error", target.Name+".c11")
	assert.NoError(t, err)
	assert.True(t, strings.HasPrefix(out, "<!DOCTYPE html>"))

	// invalid dns
	_, err = p.Exec(source.ID, "curl", "--fail", "--silent", "--show-error", target.Name+".c12")
	assert.Error(t, err)
}

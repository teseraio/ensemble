package testutil

import (
	"fmt"
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
	/*
		t.Run("TestDNS", func(t *testing.T) {
			TestDNS(t, c, p)
		})
		t.Run("TestPodMount", func(t *testing.T) {
			TestPodMount(t, c, p)
		})
		t.Run("TestPodFiles", func(t *testing.T) {
			TestPodFiles(t, c, p)
		})
		t.Run("TestPodBarArgs", func(t *testing.T) {
			TestPodBarArgs(t, c, p)
		})
		t.Run("TestPodJobFailed", func(t *testing.T) {
			TestPodJobFailed(t, c, p)
		})
	*/
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

func TestPodBarArgs(t *testing.T, c operator.ControlPlane, p operator.Provider) {
	i := &proto.Instance{
		ID:          uuid.UUID(),
		ClusterName: "xx11",
		Name:        "yy22",
		Image:       "busybox",
		Spec: &proto.NodeSpec{
			Cmd: "xxx",
		},
		Status: proto.Instance_PENDING,
	}
	if err := c.UpsertInstance(i); err != nil {
		t.Fatal(err)
	}

	ii := readEvent(c, t)
	assert.Equal(t, ii.Status, proto.Instance_FAILED)
}

func TestPodJobFailed(t *testing.T, c operator.ControlPlane, p operator.Provider) {
	i := &proto.Instance{
		ID:          uuid.UUID(),
		ClusterName: "xx11",
		Name:        "yyy22",
		Image:       "busybox",
		Spec: &proto.NodeSpec{
			// it stops gracefully
			Cmd:  "sleep",
			Args: []string{"2"},
		},
		Status: proto.Instance_PENDING,
	}
	if err := c.UpsertInstance(i); err != nil {
		t.Fatal(err)
	}

	ii := readEvent(c, t)
	assert.Equal(t, ii.Status, proto.Instance_RUNNING)

	ii = readEvent(c, t)
	assert.Equal(t, ii.Status, proto.Instance_FAILED)
}

func TestPodLifecycle(t *testing.T, c operator.ControlPlane, p operator.Provider) {
	id := uuid.UUID()

	// create the resource
	i := &proto.Instance{
		ID:          id,
		ClusterName: "c111",
		Name:        "d22",
		Image:       "nginx",
		Spec:        &proto.NodeSpec{},
		Status:      proto.Instance_PENDING,
	}
	if err := c.UpsertInstance(i); err != nil {
		t.Fatal(err)
	}

	for {
		ii := readEvent(c, t)

		fmt.Println("-- ii --")
		fmt.Println(ii)
	}

	/*
		assert.Equal(t, ii.Status, proto.Instance_RUNNING)
		assert.NotEmpty(t, ii.Ip, "")
		assert.NotEmpty(t, ii.Handler, "")
	*/

	return

	/*
		// delete the resource
		ii = ii.Copy()
		ii.Status = proto.Instance_TAINTED

		if err := c.UpsertInstance(ii); err != nil {
			t.Fatal(err)
		}

		ii = readEvent(c, t)
		assert.Equal(t, ii.Status, proto.Instance_FAILED)

		// recreate the resource with the same name and different id
		ii = ii.Copy()
		ii.ID = uuid.UUID()
		ii.Status = proto.Instance_PENDING

		if err := c.UpsertInstance(ii); err != nil {
			t.Fatal(err)
		}

		ii = readEvent(c, t)
		assert.Equal(t, ii.Status, proto.Instance_RUNNING)
	*/
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
		ID:          id,
		ClusterName: "c11",
		Name:        uuid.UUID(),
		Image:       "nginx",
		Spec: &proto.NodeSpec{
			Files: files,
		},
		Status: proto.Instance_PENDING,
	}
	if err := c.UpsertInstance(i); err != nil {
		t.Fatal(err)
	}

	ii := readEvent(c, t)
	assert.Equal(t, ii.Status, proto.Instance_RUNNING)

	for _, file := range files {
		out, err := p.Exec(id, "cat", file.Name)
		assert.NoError(t, err)
		assert.Equal(t, out, file.Content)
	}
}

func TestPodMount(t *testing.T, c operator.ControlPlane, p operator.Provider) {
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
		Status: proto.Instance_PENDING,
	}

	if err := c.UpsertInstance(i); err != nil {
		t.Fatal(err)
	}
	ii := readEvent(c, t)
	assert.Equal(t, ii.Status, proto.Instance_RUNNING)

	{
		// /data/test.txt does not exists
		_, err := p.Exec(i.ID, "cat", "/data/test.txt")
		assert.Error(t, err)

		if _, err := p.Exec(i.ID, "touch", "/data/test.txt"); err != nil {
			t.Fatal(err)
		}
	}

	{
		// stop the container
		ii = ii.Copy()
		ii.Status = proto.Instance_TAINTED

		if err := c.UpsertInstance(ii); err != nil {
			t.Fatal(err)
		}
		ii = readEvent(c, t)
		assert.Equal(t, ii.Status, proto.Instance_FAILED)
	}

	{
		// "restart" the instance with a different id
		ii := i.Copy()
		ii.ID = uuid.UUID()

		if err := c.UpsertInstance(ii); err != nil {
			t.Fatal(err)
		}

		ii = readEvent(c, t)
		assert.Equal(t, ii.Status, proto.Instance_RUNNING)

		// /data/test.txt should be available
		_, err := p.Exec(ii.ID, "cat", "/data/test.txt")
		assert.NoError(t, err)
	}
}

func TestDNS(t *testing.T, c operator.ControlPlane, p operator.Provider) {
	target := &proto.Instance{
		ID:          uuid.UUID(),
		ClusterName: "c11",
		Name:        "target",
		Image:       "nginx",
		Spec:        &proto.NodeSpec{},
		Status:      proto.Instance_PENDING,
	}
	if err := c.UpsertInstance(target); err != nil {
		t.Fatal(err)
	}

	// wait for it to be ready
	ii := readEvent(c, t)
	assert.Equal(t, ii.Status, proto.Instance_RUNNING)

	source := &proto.Instance{
		ID:          uuid.UUID(),
		ClusterName: "c11",
		Name:        uuid.UUID(),
		Image:       "nginx",
		Spec:        &proto.NodeSpec{},
		Status:      proto.Instance_PENDING,
	}
	if err := c.UpsertInstance(source); err != nil {
		t.Fatal(err)
	}

	// wait for it to be ready
	ii = readEvent(c, t)
	assert.Equal(t, ii.Status, proto.Instance_RUNNING)

	// valid dns
	out, err := p.Exec(source.ID, "curl", "--fail", "--silent", "--show-error", "target.c11")
	assert.NoError(t, err)
	assert.True(t, strings.HasPrefix(out, "<!DOCTYPE html>"))

	// invalid dns
	_, err = p.Exec(source.ID, "curl", "--fail", "--silent", "--show-error", "target.c12")
	assert.Error(t, err)
}

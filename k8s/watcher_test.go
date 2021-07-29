package k8s

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/assert"
	"github.com/teseraio/ensemble/lib/uuid"
	"github.com/teseraio/ensemble/operator/proto"
)

type crdSetup struct {
	name string

	t *testing.T
	p *Provider
}

func (c *crdSetup) Delete(name string) {
	err := c.p.delete(c.Path()+"/"+name, emptyDel)
	if err != nil {
		c.t.Fatal(err)
	}
}

func (c *crdSetup) Update(name string, resourceVersion string) *Metadata {
	obj, _ := RunTmpl2("mock", map[string]interface{}{
		"resource":        strings.Title(c.name),
		"name":            name,
		"value":           uuid.UUID(),
		"resourceVersion": resourceVersion,
	})
	_, metadata, err := c.p.put(c.Path()+"/"+name, obj)
	if err != nil {
		c.t.Fatal(err)
	}
	return metadata
}

func (c *crdSetup) Create(name string) *Metadata {
	obj, _ := RunTmpl2("mock", map[string]interface{}{
		"resource": strings.Title(c.name),
		"name":     name,
		"value":    uuid.UUID(),
	})
	_, metadata, err := c.p.post(c.Path(), obj)
	if err != nil {
		c.t.Fatal(err)
	}
	return metadata
}

func (c *crdSetup) Path() string {
	return "/apis/mock.io/v1/namespaces/default/" + c.name + "s"
}

func (c *crdSetup) Close() {
	if err := c.p.delete("/apis/apiextensions.k8s.io/v1/customresourcedefinitions/"+c.name+"s.mock.io", emptyDel); err != nil {
		c.t.Fatal(err)
	}
}

func setupCRD(t *testing.T, p *Provider, name string) *crdSetup {
	crdDefinition, _ := RunTmpl2("mock-crd", map[string]interface{}{
		"name": name,
	})
	if _, _, err := p.post(crdURL, []byte(crdDefinition)); err != nil {
		if err != errAlreadyExists {
			t.Fatal(err)
		}
	}

	// wait for the CRD to be created
	time.Sleep(2 * time.Second)

	crd := &crdSetup{
		name: name,
		t:    t,
		p:    p,
	}
	return crd
}

func TestWatcher_WrongPath(t *testing.T) {
	p, _ := K8sFactory(hclog.NewNullLogger(), nil)
	nullLogger := hclog.NewNullLogger()

	_, err := NewWatcher(nullLogger, p.client, "/apis/mock.io/v1/namespaces/default/test2s", &Item{})
	assert.Error(t, err)
}

func TestWatcher_ListImpl(t *testing.T) {
	p, _ := K8sFactory(hclog.NewNullLogger(), nil)

	crd := setupCRD(t, p, "testget")
	defer crd.Close()

	watcher, err := NewWatcher(hclog.NewNullLogger(), p.client, crd.Path(), &Item{})
	assert.NoError(t, err)
	watcher.WithLimit(2)

	// query the latest item
	pager, err := watcher.listImpl("")
	assert.NoError(t, err)
	assert.Zero(t, pager.items)

	metadata := []*Metadata{}
	for i := 0; i < 20; i++ {
		metadata = append(metadata, crd.Create(fmt.Sprintf("item-%d", i)))
	}

	// query from the beginning, IT DOES NOT respect limit and returns all the elements
	pager, err = watcher.listImpl("0")
	assert.NoError(t, err)
	assert.Len(t, pager.items, 20)

	// query from a specific resource version
	pager, err = watcher.listImpl(metadata[10].ResourceVersion)
	assert.NoError(t, err)
	assert.Len(t, pager.items, 11)
}

func TestWatcher_WatchExpire(t *testing.T) {
	p, _ := K8sFactory(hclog.NewNullLogger(), nil)

	crd := setupCRD(t, p, "testwatchererror")
	defer crd.Close()

	watcher, err := NewWatcher(hclog.NewNullLogger(), p.client, crd.Path(), &Item{})
	assert.NoError(t, err)

	// expires
	err = watcher.watchImpl("1", func(typ string, item itemObj) error {
		return nil
	})
	assert.ErrorIs(t, err, errExpired)
}

func TestWatcher_WatchImpl(t *testing.T) {
	p, _ := K8sFactory(hclog.NewNullLogger(), nil)

	crd := setupCRD(t, p, "testwatcher")
	defer crd.Close()

	watcher, err := NewWatcher(hclog.NewNullLogger(), p.client, crd.Path(), &Item{})
	assert.NoError(t, err)

	// create one item to have the resource version reference
	metadata := crd.Create("item")

	eventCh := make(chan string)
	go func() {
		watcher.watchImpl(metadata.ResourceVersion, func(typ string, item itemObj) error {
			eventCh <- typ
			return nil
		})
	}()

	// create
	metadata = crd.Create("item-2")
	evnt := <-eventCh
	assert.Equal(t, evnt, "ADDED")

	// update
	crd.Update("item-2", metadata.ResourceVersion)
	evnt = <-eventCh
	assert.Equal(t, evnt, "MODIFIED")

	// delete
	crd.Delete("item-2")
	evnt = <-eventCh
	assert.Equal(t, evnt, "DELETED")
}

func TestWatcher_Lifecycle(t *testing.T) {
	p, _ := K8sFactory(hclog.NewNullLogger(), nil)

	crd := setupCRD(t, p, "testlifecycle")
	defer crd.Close()

	watcher, err := NewWatcher(hclog.NewNullLogger(), p.client, crd.Path(), &Item{})
	assert.NoError(t, err)

	for i := 0; i < 20; i++ {
		crd.Create(fmt.Sprintf("item-%d", i))
	}

	// wait for the items to be available for the list
	time.Sleep(2 * time.Second)

	watcher.WithList(true)

	stopCh := make(chan struct{})
	watcher.Run(stopCh)

	popItem := func() *WatchEntry {
		ctx, cancelFn := context.WithCancel(context.Background())
		go func() {
			<-time.After(2 * time.Second)
			cancelFn()
		}()
		obj := watcher.store.pop(ctx)
		if obj == nil {
			t.Fatal("timeout")
		}
		return obj
	}

	for i := 0; i < 20; i++ {
		obj := popItem()
		item := obj.item.(*Item)
		assert.Equal(t, item.Metadata.Name, fmt.Sprintf("item-%d", i))
	}

	// update
	for i := 0; i < 10; i++ {
		crd.Delete(fmt.Sprintf("item-%d", i))
	}

	for i := 0; i < 10; i++ {
		obj := popItem()
		item := obj.item.(*Item)
		assert.Equal(t, item.Metadata.Name, fmt.Sprintf("item-%d", i))
	}
}

func TestWatcher_PodUpdate(t *testing.T) {
	p, _ := K8sFactory(hclog.NewNullLogger(), nil)

	watcher, err := NewWatcher(hclog.NewNullLogger(), p.client, podsURL, &PodItem{})
	assert.NoError(t, err)

	watcher.Run(nil)

	// create pod
	instance := &proto.Instance{
		ID:           uuid.UUID(),
		DeploymentID: "c11",
		ClusterName:  "c11",
		Name:         "d22",
		Image:        "nginx",
		Spec:         &proto.NodeSpec{},
		Status:       proto.Instance_PENDING,
	}
	assert.NoError(t, p.createPod(instance))

	// check manually if its running
	for i := 0; ; i++ {
		var item PodItem
		_, err := p.get(podsURL+"/"+instance.ID, &item)
		assert.NoError(t, err)

		if item.Status.Phase == "Running" {
			break
		}
		if i < 10 {
			time.Sleep(1 * time.Second)
		} else {
			t.Fatal("timeout")
		}
	}

	popItem := func() *WatchEntry {
		ctx, cancelFn := context.WithCancel(context.Background())
		go func() {
			<-time.After(2 * time.Second)
			cancelFn()
		}()
		obj := watcher.store.pop(ctx)
		if obj == nil {
			t.Fatal("timeout")
		}
		return obj
	}

	// there should be more than one entry
	for i := 0; i < 2; i++ {
		popItem()
	}
}

func TestWatcher_Queue(t *testing.T) {
	s := newStore()

	s.add("", &Item{
		Metadata: &Metadata{
			Name: "A",
		},
	})

	s.add("", &Item{
		Metadata: &Metadata{
			Name: "B",
		},
	})

	s.add("", &Item{
		Metadata: &Metadata{
			Name: "A",
		},
		Kind: "Update",
	})

	// 2 elements in the store
	if len(s.items) != len(s.heapImpl) && len(s.items) != 1 {
		t.Fatal("bad")
	}

	// A pops updated
	e := s.pop(context.Background())
	if e.item.(*Item).Metadata.Name != "A" {
		t.Fatal("B expected")
	}
	if e.item.(*Item).Kind != "Update" {
		t.Fatal("bad")
	}

	// 1 element left
	if len(s.items) != len(s.heapImpl) && len(s.items) != 1 {
		t.Fatal("bad")
	}

	// B pops
	e = s.pop(context.Background())
	if e.item.(*Item).Metadata.Name != "B" {
		t.Fatal("B expected")
	}
}

func TestWatcher_ListErrors(t *testing.T) {
	t.Skip("Not sure how to test this yet")

	p, _ := K8sFactory(hclog.NewNullLogger(), nil)

	crd := setupCRD(t, p, "testgeterror")
	defer crd.Close()

	watcher, err := NewWatcher(hclog.NewNullLogger(), p.client, crd.Path(), &Item{})
	assert.NoError(t, err)

	// call an expired resource version that is not available
	_, err = watcher.listImpl("1")
	assert.ErrorIs(t, err, errExpired)

	metadata := crd.Create("item")

	// call a future resource version
	_, err = watcher.listImpl(metadata.ResourceVersion + "0")
	assert.ErrorIs(t, err, errFutureVersion)
}

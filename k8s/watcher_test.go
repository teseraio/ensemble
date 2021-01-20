package k8s

import (
	"context"
	"testing"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/teseraio/ensemble/lib/uuid"
)

func TestWatcher(t *testing.T) {
	p, _ := K8sFactory(hclog.NewNullLogger(), nil)

	// Create test CRD objects
	crdDefinition, _ := RunTmpl2("mock-crd", map[string]interface{}{
		"name": "test1",
	})
	if _, _, err := p.post(crdURL, []byte(crdDefinition)); err != nil {
		if err != errAlreadyExists {
			t.Fatal(err)
		}
	}

	// wait for the CRD to be created
	time.Sleep(1 * time.Second)

	defer func() {
		if err := p.delete("/apis/apiextensions.k8s.io/v1/customresourcedefinitions/test1s.mock.io", emptyDel); err != nil {
			t.Fatal(err)
		}
	}()

	create := func(name string) {
		obj, _ := RunTmpl2("mock", map[string]interface{}{
			"resource": "Test1",
			"name":     name,
		})
		if _, _, err := p.post("/apis/mock.io/v1/namespaces/{namespace}/test1s", obj); err != nil {
			t.Fatal(err)
		}
	}

	for i := 0; i < 20; i++ {
		create(uuid.UUID())
	}

	store := newStore()
	newWatcher(store, p.client, "/apis/mock.io/v1/namespaces/default/test1s")
}

func TestWatcherStore(t *testing.T) {
	s := newStore()

	s.add(&Item{
		Metadata: &Metadata{
			Name: "A",
		},
	})

	s.add(&Item{
		Metadata: &Metadata{
			Name: "B",
		},
	})

	s.add(&Item{
		Metadata: &Metadata{
			Name: "A",
		},
		Kind: "Update",
	})

	// 2 elements in the store
	if len(s.items) != len(s.heapImpl) && len(s.items) != 1 {
		t.Fatal("bad")
	}

	// B pop first because A was modified
	e := s.pop(context.Background())
	if e.item.Metadata.Name != "B" {
		t.Fatal("B expected")
	}

	// 1 element left
	if len(s.items) != len(s.heapImpl) && len(s.items) != 1 {
		t.Fatal("bad")
	}

	// A pops
	e = s.pop(context.Background())
	if e.item.Metadata.Name != "A" {
		t.Fatal("B expected")
	}
	if e.item.Kind != "Update" {
		t.Fatal("bad")
	}
}

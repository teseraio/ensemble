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

	ids := []string{}
	for i := 0; i < 20; i++ {
		id := uuid.UUID()
		ids = append(ids, id)
		create(id)
	}

	store := newStore()
	newWatcher(store, p.client, "/apis/mock.io/v1/namespaces/default/test1s", &Item{}, true)

	for i := 0; i < 20; i++ {
		e := store.pop(context.Background())
		item := e.item.(*Item)
		if !contains(ids, item.Metadata.Name) {
			t.Fatal("bad")
		}
	}

	// delete an element
	if err := p.delete("/apis/mock.io/v1/namespaces/{namespace}/test1s/"+ids[0], emptyDel); err != nil {
		t.Fatal(err)
	}

	tt := store.pop(context.Background())
	if tt.item.(*Item).Metadata.Name != ids[0] {
		t.Fatal("bad")
	}
}

func contains(i []string, j string) bool {
	for _, o := range i {
		if o == j {
			return true
		}
	}
	return false
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
	if e.item.(*Item).Metadata.Name != "B" {
		t.Fatal("B expected")
	}

	// 1 element left
	if len(s.items) != len(s.heapImpl) && len(s.items) != 1 {
		t.Fatal("bad")
	}

	// A pops
	e = s.pop(context.Background())
	if e.item.(*Item).Metadata.Name != "A" {
		t.Fatal("B expected")
	}
	if e.item.(*Item).Kind != "Update" {
		t.Fatal("bad")
	}
}

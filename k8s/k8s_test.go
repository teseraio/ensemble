package k8s

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/teseraio/ensemble/lib/uuid"
	"github.com/teseraio/ensemble/operator/proto"
)

func TestUpsertConfigMap(t *testing.T) {
	p, _ := K8sFactory(hclog.NewNullLogger(), nil)

	id := uuid.UUID()

	// check that the config map was created
	checkData := func(data map[string]string) {
		var res struct {
			Data map[string]string
		}
		if _, err := p.get("/api/v1/namespaces/{namespace}/configmaps/"+id, &res); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(res.Data, data) {
			t.Fatal("bad")
		}
	}

	data := map[string]string{
		"A": "B",
		"C": "D",
	}
	if err := p.upsertConfigMap(id, data); err != nil {
		t.Fatal(err)
	}

	checkData(data)

	// change the values
	data = map[string]string{
		"E": "F",
		"G": "H",
	}
	if err := p.upsertConfigMap(id, data); err != nil {
		t.Fatal(err)
	}

	checkData(data)
}

func TestPodLifecycle(t *testing.T) {
	p, _ := K8sFactory(hclog.NewNullLogger(), nil)
	p.Setup()

	id := uuid.UUID()

	n0 := &proto.Instance{
		ID:      id,
		Cluster: "a",
		Status:  proto.Instance_RUNNING,
		Spec: &proto.NodeSpec{
			Image:   "redis",
			Version: "latest",
		},
	}

	if _, err := p.CreateResource(n0); err != nil {
		t.Fatal(err)
	}

	var evnt *proto.InstanceUpdate

	// wait for successful event
	select {
	case evnt = <-p.watchCh:
	case <-time.After(10 * time.Second):
		t.Fatal("timeout")
	}

	fmt.Println("- evnt -")
	fmt.Println(evnt)

	if _, err := p.DeleteResource(n0); err != nil {
		t.Fatal(err)
	}

	// wait for termination event
	select {
	case evnt = <-p.watchCh:
	case <-time.After(3 * time.Second):
		t.Fatal("timeout")
	}
}

func isNotFound(err error) bool {
	if err == nil {
		return false
	}
	return err == errNotFound
}

func TestOperatorLifecycle(t *testing.T) {
	p, _ := K8sFactory(hclog.NewNullLogger(), nil)
	p.Setup()
}

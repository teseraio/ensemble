package k8s

import (
	"testing"

	"github.com/hashicorp/go-hclog"
	"github.com/teseraio/ensemble/testutil"
)

func TestK8sProviderSpec(t *testing.T) {
	// TODO
	p, err := K8sFactory(hclog.NewNullLogger(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Setup(); err != nil {
		t.Fatal(err)
	}
	testutil.TestProvider(t, p)
}

/*
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

func readEvent(p *Provider, t *testing.T) *proto.InstanceUpdate {
	select {
	case evnt := <-p.watchCh:
		return evnt
	case <-time.After(10 * time.Second):
	}
	t.Fatal("timeout")
	return nil
}

func TestPodLifecycle(t *testing.T) {
	p, _ := K8sFactory(hclog.NewNullLogger(), nil)
	p.Setup()

	id := uuid.UUID()

	i := &proto.Instance{
		ID:      id,
		Cluster: "c11",
		Name:    "d22",
		Spec: &proto.NodeSpec{
			Image: "busybox",
			Cmd:   []string{"sleep", "30000"},
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
	if _, ok := evnt.Event.(*proto.InstanceUpdate_Running_); !ok {
		t.Fatal("expected running")
	}

	if _, err := p.DeleteResource(i); err != nil {
		t.Fatal(err)
	}

	// wait for termination event
	evnt = readEvent(p, t)
	if _, ok := evnt.Event.(*proto.InstanceUpdate_Killing_); !ok {
		t.Fatal("expected stopped")
	}
}

func TestPodDns(t *testing.T) {
	p, _ := K8sFactory(hclog.NewNullLogger(), nil)
	p.Setup()

	i := &proto.Instance{
		ID:      uuid.UUID(),
		Cluster: "c1",
		Name:    "d2",
		Spec: &proto.NodeSpec{
			Image: "nginx",
		},
	}

	if _, err := p.CreateResource(i); err != nil {
		t.Fatal(err)
	}

	// wait for the container to be ready
	for {
		evnt := <-p.watchCh
		if _, ok := evnt.Event.(*proto.InstanceUpdate_Running_); ok {
			break
		}
	}

	// create a curl container
	ii := &proto.Instance{
		ID:      uuid.UUID(),
		Cluster: "c2",
		Name:    "n2",
		Spec: &proto.NodeSpec{
			Image:   "curlimages/curl",
			Version: "7.75.0",
			Cmd:     []string{"sleep", "30000"},
		},
	}

	if _, err := p.CreateResource(ii); err != nil {
		t.Fatal(err)
	}

	// wait for the container to be ready
	for {
		evnt := <-p.watchCh
		if _, ok := evnt.Event.(*proto.InstanceUpdate_Running_); ok {
			break
		}
	}

	// TODO: exec curl container
	p.Exec(ii.ID, "curl", "d2.c1")
}

func TestPodBadArgs(t *testing.T) {
	p, _ := K8sFactory(hclog.NewNullLogger(), nil)
	p.Setup()

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
*/

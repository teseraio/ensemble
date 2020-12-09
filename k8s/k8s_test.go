package k8s

import (
	"reflect"
	"testing"
	"time"

	"github.com/golang/protobuf/ptypes"

	"github.com/google/uuid"
	"github.com/hashicorp/go-hclog"
	"github.com/teseraio/ensemble/operator/proto"
)

const crdURL = "/apis/apiextensions.k8s.io/v1/customresourcedefinitions"

func setupCRDs(t *testing.T, p *Provider) func() {
	createAssets := []string{
		"node",
		"cluster",
		"resource",
	}
	for _, asset := range createAssets {
		crdDefinition, err := Asset("resources/crd-" + asset + ".json")
		if err != nil {
			t.Fatal(err)
		}
		if _, _, err := p.post(crdURL, []byte(crdDefinition)); err != nil {
			if err != errAlreadyExists {
				t.Fatal(err)
			}
		}
	}

	// wait for the CRDs to be available
	time.Sleep(1 * time.Second)

	purge := func() {
		for _, name := range []string{"clusters", "nodes", "resources"} {
			if err := p.delete("/apis/apiextensions.k8s.io/v1/customresourcedefinitions/"+name+".ensembleoss.io", emptyDel); err != nil {
				t.Fatal(err)
			}
		}
	}
	return purge
}

func TestUpsertConfigMap(t *testing.T) {
	p := K8sFactory(hclog.NewNullLogger(), nil)

	id := uuid.New().String()

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
	p := K8sFactory(hclog.NewNullLogger(), nil)

	purgeFn := setupCRDs(t, p)
	defer purgeFn()

	p.Setup()

	id := uuid.New().String()

	n0 := &proto.Node{
		ID:      id,
		Cluster: "a",
		State:   proto.NodeState_RUNNING,
		Spec: &proto.NodeSpec{
			Image:   "redis", // TODO: Something better
			Version: "latest",
		},
	}
	n1, err := p.CreateResource(n0)
	if err != nil {
		t.Fatal(err)
	}

	// update the node
	n1.Set("A", "B")

	n2, err := p.UpdateNodeStatus(n1)
	if err != nil {
		t.Fatal(err)
	}
	// resource version has to increase
	if n1.ResourceVersion >= n2.ResourceVersion {
		t.Fatal("bad")
	}

	// try to delete an unknown node (it should fail)
	if _, err := p.DeleteResource(&proto.Node{ID: "A"}); err == nil {
		t.Fatal("bad")
	}

	n2.State = proto.NodeState_DOWN

	n3, err := p.DeleteResource(n2)
	if err != nil {
		t.Fatal(err)
	}
	// resource version has to increase
	if n2.ResourceVersion >= n3.ResourceVersion {
		t.Fatal("bad")
	}
}

func TestPurgeCluster(t *testing.T) {
	p := K8sFactory(hclog.NewNullLogger(), nil)

	purgeFn := setupCRDs(t, p)
	defer purgeFn()

	p.Setup()

	// Dns names must start with a char
	ensembleName := "srv" + uuid.New().String()

	mockNode := func() *proto.Node {
		return &proto.Node{
			ID:      uuid.New().String(),
			Cluster: ensembleName,
			State:   proto.NodeState_RUNNING,
			Spec: &proto.NodeSpec{
				Image:   "redis", // TODO: Something better
				Version: "latest",
			},
		}
	}

	n0, n1 := mockNode(), mockNode()

	if _, err := p.CreateResource(n0); err != nil {
		t.Fatal(err)
	}
	if _, err := p.CreateResource(n1); err != nil {
		t.Fatal(err)
	}

	// Purge the cluster
	if err := p.purgeCluster(ensembleName); err != nil {
		t.Fatal(err)
	}

	// there should not be any node for the ensemble
	var res struct {
		Items []*Item
	}
	if _, err := p.get("/apis/ensembleoss.io/v1/namespaces/{namespace}/nodes?labelSelector=ensemble%3D"+ensembleName, &res); err != nil {
		t.Fatal(err)
	}
	if len(res.Items) != 0 {
		t.Fatal("bad")
	}

	// the service has been removed
	if _, err := p.get("/api/v1/namespaces/{namespace}/services/"+ensembleName, nil); !isNotFound(err) {
		t.Fatal("bad")
	}
}

func isNotFound(err error) bool {
	if err == nil {
		return false
	}
	return err == errNotFound
}

func TestNodeEncoding(t *testing.T) {
	p := K8sFactory(hclog.NewNullLogger(), nil)

	purgeFn := setupCRDs(t, p)
	defer purgeFn()

	p.Setup()

	var cases = []*proto.Node{
		&proto.Node{
			ID:       uuid.New().String(),
			Nodeset:  "a",
			Nodetype: "b",
			Addr:     "c",
			Handle:   "d",
			State:    proto.NodeState_RUNNING,
			Spec: &proto.NodeSpec{
				Image: "h",
				Env: map[string]string{
					"a": "b",
				},
				Files: map[string]string{
					"c": "d",
				},
				Cmd: []string{"a", "b", "c"},
			},
			KV: map[string]string{
				"a": "b",
			},
			Mounts: []*proto.Mount{
				&proto.Mount{
					Id:   "id",
					Name: "name",
					Path: "path",
				},
			},
		},
	}

	for _, c := range cases {
		n0, err := p.upsertNodeSpec(c)
		if err != nil {
			t.Fatal(err)
		}

		// we need to query the same node we return from the upsert
		n1, err := p.loadNode(c.ID)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(n1, n0) {
			t.Fatal(err)
		}

		if n0.ResourceVersion == c.ResourceVersion {
			t.Fatal("bad")
		}

		// equal so that we can use reflect
		n0.ResourceVersion = c.ResourceVersion
		if !reflect.DeepEqual(c, n0) {
			t.Fatal("bad")
		}
	}
}

func TestLoadCluster(t *testing.T) {
	p := K8sFactory(hclog.NewNullLogger(), nil)

	purgeFn := setupCRDs(t, p)
	defer purgeFn()

	p.Setup()

	ensembleName := "srv" + uuid.New().String()

	mockNode := func() *proto.Node {
		return &proto.Node{
			ID:      uuid.New().String(),
			Cluster: ensembleName,
			State:   proto.NodeState_RUNNING,
			Spec: &proto.NodeSpec{
				Image: "redis",
			},
		}
	}

	n0, n1 := mockNode(), mockNode()

	if _, err := p.upsertNodeSpec(n0); err != nil {
		t.Fatal(err)
	}
	if _, err := p.upsertNodeSpec(n1); err != nil {
		t.Fatal(err)
	}

	c, err := p.LoadCluster(ensembleName)
	if err != nil {
		t.Fatal(err)
	}
	if len(c.Nodes) != 2 {
		t.Fatal("bad")
	}
}

func TestTrackerLifecycle(t *testing.T) {
	p := K8sFactory(hclog.NewNullLogger(), nil)

	purgeFn := setupCRDs(t, p)
	defer purgeFn()

	p.Start()

	createObject := func(backend string) string {
		id := uuid.New().String()

		obj := map[string]interface{}{
			"Name":     id,
			"Backend":  backend,
			"Replicas": 2,
		}
		req, err := RunTmpl2("kind-cluster", obj)
		if err != nil {
			t.Fatal(err)
		}
		if _, _, err := p.post("/apis/ensembleoss.io/v1/namespaces/{namespace}/clusters", req); err != nil {
			t.Fatal(err)
		}
		return id
	}

	createResource := func() string {
		id := uuid.New().String()

		obj := map[string]interface{}{
			"Name":     id,
			"Cluster":  "Cluster1",
			"Resource": "Resource1",
		}
		req, err := RunTmpl2("kind-resource", obj)
		if err != nil {
			t.Fatal(err)
		}
		if _, _, err := p.post("/apis/ensembleoss.io/v1/namespaces/{namespace}/resources", req); err != nil {
			t.Fatal(err)
		}
		return id
	}

	uuid0 := createObject("backend1")

	t0, err := p.GetTask()
	if err != nil {
		t.Fatal(err)
	}
	if t0.Evaluation.Name != uuid0 {
		t.Fatal("bad")
	}

	if err := p.FinalizeTask(t0.ID); err != nil {
		t.Fatal(err)
	}

	uuid1 := createResource()

	t1, err := p.GetTask()
	if err != nil {
		t.Fatal(err)
	}
	if t1.Evaluation.Name != uuid1 {
		t.Fatal("bad")
	}
	if t1.Evaluation.Cluster != "Cluster1" {
		t.Fatal("bad")
	}
	if t1.Evaluation.Resource != "Resource1" {
		t.Fatal("bad")
	}
}

func TestTaskQueue(t *testing.T) {
	tt := newTaskQueue()

	mockTask := func() *task {
		timestamp := ptypes.TimestampNow()
		return &task{
			Task: &proto.Task{
				ID:        uuid.New().String(),
				Timestamp: timestamp,
			},
			timestamp: timestamp.AsTime(),
		}
	}

	t0 := mockTask()
	t1 := mockTask()
	t2 := mockTask()

	tt.add(t0)
	tt.add(t1)
	tt.add(t2)

	stopCh := make(chan struct{})

	if t0e := tt.pop(stopCh); t0e.ID != t0.ID {
		t.Fatal("bad")
	}
	if t1e := tt.pop(stopCh); t1e.ID != t1.ID {
		t.Fatal("bad")
	}
	if t2e := tt.pop(stopCh); t2e.ID != t2.ID {
		t.Fatal("bad")
	}

	// it blocks if there are no more items till a new one arrives
	t3 := mockTask()

	taskCh := make(chan *task)
	go func() {
		taskCh <- tt.pop(stopCh)
	}()

	select {
	case <-taskCh:
		t.Fatal("bad")
	case <-time.After(100 * time.Millisecond):
	}

	tt.add(t3)

	select {
	case t3e := <-taskCh:
		if t3e.ID != t3.ID {
			t.Fatal("bad")
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("bad")
	}

	// Once finalized the task is removed
	if len(tt.heap) != 4 {
		t.Fatal("bad")
	}
	tt.finalize(t0.ID)

	if len(tt.heap) != 3 {
		t.Fatal("bad")
	}
}

package operator

import (
	"strconv"
	"testing"

	"github.com/teseraio/ensemble/lib/uuid"
	"github.com/teseraio/ensemble/operator/proto"
	"github.com/teseraio/ensemble/testutil"
)

func TestNodeHooks(t *testing.T) {
	provider, _ := testutil.NewTestProvider(t, "Mock", nil)
	srv := TestOperator(t, provider, nil)
	defer srv.Stop()

	id := uuid.UUID()

	c := &proto.Cluster{}
	n := &proto.Node{
		ID: id,
	}

	h := &mockHandler{}

	// add the node
	cc, err := srv.addNode(h, c, n, &proto.Plan{})
	if err != nil {
		t.Fatal(err)
	}
	if len(cc.Nodes) != 1 {
		t.Fatal("bad")
	}
	node := cc.Nodes[0]
	if node.ID != id {
		t.Fatal("bad")
	}
	if node.State != proto.Node_RUNNING {
		t.Fatal("bad")
	}

	// delete the node
	cc, err = srv.deleteNode(h, cc, node, &proto.Plan{})
	if err != nil {
		t.Fatal(err)
	}
	// it should not remove the node from the cluster (yet)
	if len(cc.Nodes) != 1 {
		t.Fatal("bad")
	}
	if cc.Nodes[0].State != proto.Node_DOWN {
		t.Fatal("bad")
	}
}

func TestEvaluateCluster(t *testing.T) {
	handler := &mockHandler{}

	cases := []struct {
		replicas int
		cluster  *proto.Cluster
		check    func(*testing.T, *proto.Plan)
	}{
		{
			// scale up
			3,
			&proto.Cluster{
				Nodes: []*proto.Node{},
			},
			func(t *testing.T, p *proto.Plan) {
				if !p.Bootstrap {
					t.Fatal("bad")
				}
				if len(p.AddNodes) != 3 {
					t.Fatal("bad")
				}
			},
			// scale down
		},
		{
			1,
			&proto.Cluster{
				Nodes: []*proto.Node{
					{
						ID: uuid.UUID(),
					},
					{
						ID: uuid.UUID(),
					},
					{
						ID: uuid.UUID(),
					},
				},
			},
			func(t *testing.T, p *proto.Plan) {
				if p.DelNodesNum != 2 {
					t.Fatal("bad")
				}
			},
		},
	}
	for _, c := range cases {
		evaluation := &proto.Evaluation{
			Spec: `{"replicas": ` + strconv.Itoa(c.replicas) + `}`,
		}
		found, _ := evaluateCluster(evaluation, c.cluster, handler)
		c.check(t, found)
	}
}

type mockHandler struct {
}

func (m *mockHandler) Reconcile(executor Executor, e *proto.Cluster, node *proto.Node, plan *proto.Plan) error {
	return nil
}

func (m *mockHandler) EvaluatePlan(plan *proto.Plan) error {
	return nil
}

func (m *mockHandler) Spec() *Spec {
	return &Spec{
		Nodetypes: map[string]Nodetype{
			"": Nodetype{
				Image:   "redis",
				Version: "latest", // TODO, mock provider
			},
		},
	}
}

func (m *mockHandler) Client(node *proto.Node) (interface{}, error) {
	return nil, nil
}

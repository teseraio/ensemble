package state

import (
	"testing"

	"github.com/teseraio/ensemble/operator/proto"
)

type setupFn func(*testing.T) (State, func())

// TestSuite has a suite of tests for state implementations
func TestSuite(t *testing.T, setup setupFn) {
	b, closeFn := setup(t)
	defer closeFn()

	/*
		c0 := &proto.Cluster{
			Name: "C",
		}
		if err := b.UpsertCluster(c0); err != nil {
			t.Fatal(err)
		}
		n0 := &proto.Node{
			ID:      "A",
			Cluster: "C",
		}
		if err := b.UpsertNode(n0); err != nil {
			t.Fatal(err)
		}

		c00, err := b.LoadCluster("C")
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println(c00)
	*/

	t.Run("Task get", func(t *testing.T) {
		c0 := &proto.Component{
			Id:   "A",
			Name: "A",
			Spec: proto.MustMarshalAny(&proto.ClusterSpec{
				Replicas: 3,
			}),
		}
		if err := b.Apply(c0); err != nil {
			t.Fatal(err)
		}

		// send the same task again, the sequence gets updated anyway
		if err := b.Apply(c0); err != nil {
			t.Fatal(err)
		}

		c00, err := b.Get("A")
		if err != nil {
			t.Fatal(err)
		}
		if c00.Sequence != 2 {
			t.Fatal("bad")
		}
	})
}

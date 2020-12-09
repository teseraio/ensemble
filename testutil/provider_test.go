package testutil

import (
	"testing"

	"github.com/teseraio/ensemble/operator/proto"
)

func TestState(t *testing.T) {
	s := &state{}

	e0 := s.addObj("e", &proto.Evaluation{})
	e1 := s.addObj("e", &proto.Evaluation{})
	a0 := s.addObj("a", &proto.Evaluation{})

	if e0.Generation != 0 {
		t.Fatal("bad")
	}
	if e1.Generation != 1 {
		t.Fatal("bad")
	}
	if a0.Generation != 0 {
		t.Fatal("bad")
	}
}

package proto

import (
	"reflect"
	"testing"
)

func TestNodeSpec(t *testing.T) {
	out := map[string]string{
		"A": "B",
	}

	b := &NodeSpec{}
	b.AddEnv("A", "B")
	if !reflect.DeepEqual(b.Env, out) {
		t.Fatal("bad")
	}

	b = &NodeSpec{}
	b.AddEnvList([]string{
		"A=B",
	})
	if !reflect.DeepEqual(b.Env, out) {
		t.Fatal("bad")
	}

	b = &NodeSpec{}
	b.AddEnvMap(map[string]string{
		"A": "B",
	})
	if !reflect.DeepEqual(b.Env, out) {
		t.Fatal("bad")
	}

	b = &NodeSpec{}
	b.AddEnvList([]string{
		"A=",
	})
	if b.Env["A"] != "" {
		t.Fatal("bad")
	}
}

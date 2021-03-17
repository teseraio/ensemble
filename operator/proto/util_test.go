package proto

import "testing"

/*
func TestNodeSpec(t *testing.T) {
	out := map[string]string{
		"A": "B",
	}

	b := &Node_NodeSpec{}
	b.AddEnv("A", "B")
	if !reflect.DeepEqual(b.Env, out) {
		t.Fatal("bad")
	}

	b = &Node_NodeSpec{}
	b.AddEnvList([]string{
		"A=B",
	})
	if !reflect.DeepEqual(b.Env, out) {
		t.Fatal("bad")
	}

	b = &Node_NodeSpec{}
	b.AddEnvMap(map[string]string{
		"A": "B",
	})
	if !reflect.DeepEqual(b.Env, out) {
		t.Fatal("bad")
	}

	b = &Node_NodeSpec{}
	b.AddEnvList([]string{
		"A=",
	})
	if b.Env["A"] != "" {
		t.Fatal("bad")
	}
}
*/

func TestParseIndex(t *testing.T) {
	cases := []struct {
		name  string
		index int64
	}{
		{
			name:  "name-5",
			index: 5,
		},
		{
			name:  "name-typ2-1",
			index: 1,
		},
	}

	for _, c := range cases {
		index, err := ParseIndex(c.name)
		if err != nil && c.index != -1 {
			t.Fatal(err)
		}
		if err == nil && c.index == -1 {
			t.Fatal("it should fail")
		}
		if index != uint64(c.index) {
			t.Fatal("bad")
		}
	}
}

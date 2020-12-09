package k8s

import (
	"testing"
)

func TestCleanPath(t *testing.T) {
	cases := map[string]string{
		"/a/b/c": "a.b.c",
		"a/b/c/": "a.b.c",
	}
	for k, v := range cases {
		res := cleanPath(k)
		if res != v {
			t.Fatalf("Expected '%s' but found '%s'", v, res)
		}
	}
}

package k8s

import (
	"fmt"
	"testing"

	"github.com/teseraio/ensemble/operator/proto"
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

func TestEncodingPod(t *testing.T) {

	x, err := MarshalPod(&proto.Instance{
		Spec: &proto.NodeSpec{
			Cmd: []string{
				"A",
				"B",
				"C",
			},
		},
	})
	fmt.Println(err)
	fmt.Println(x)
}

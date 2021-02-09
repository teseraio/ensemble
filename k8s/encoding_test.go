package k8s

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"reflect"
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

func TestEncodePod(t *testing.T) {
	raw, err := ioutil.ReadFile("./resources/fixtures/pod.json")
	if err != nil {
		t.Fatal(err)
	}
	var expected map[string]interface{}
	if err := json.Unmarshal(raw, &expected); err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		Node *proto.Node
		Name string
	}{
		{
			Node: &proto.Node{
				ID:      "a",
				Cluster: "b",
				Spec: &proto.Node_NodeSpec{
					Image:   "image",
					Version: "latest",
				},
			},
			Name: "example1",
		},
		{
			Node: &proto.Node{
				ID:      "a",
				Cluster: "b",
				Spec: &proto.Node_NodeSpec{
					Image:   "image",
					Version: "latest",
				},
				Mounts: []*proto.Node_Mount{
					{
						Name: "mount1",
						Path: "/data",
					},
				},
			},
			Name: "example2",
		},
	}
	for _, c := range cases {
		raw, err := MarshalPod(c.Node)
		if err != nil {
			t.Fatal(err)
		}
		var found map[string]interface{}
		if err := json.Unmarshal(raw, &found); err != nil {
			t.Fatal(err)
		}
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(expected[c.Name], found) {
			prettyPrint(expected[c.Name].(map[string]interface{}))
			prettyPrint(found)
			t.Fatal("bad")
		}

	}
}

func prettyPrint(data map[string]interface{}) {
	raw, _ := json.MarshalIndent(data, "", "    ")
	fmt.Println(string(raw))
}

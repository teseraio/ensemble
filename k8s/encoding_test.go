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
		Node *proto.Instance
		Name string
	}{
		{
			Node: &proto.Instance{
				ID:      "a",
				Cluster: "b",
				Name:    "a",
				Spec: &proto.NodeSpec{
					Image:   "image",
					Version: "latest",
				},
			},
			Name: "example1",
		},
		{
			Node: &proto.Instance{
				ID:      "id",
				Cluster: "b",
				Name:    "c",
				Spec: &proto.NodeSpec{
					Image:   "image",
					Version: "latest",
				},
				Mounts: []*proto.Instance_Mount{
					{
						Name: "mount1",
						Path: "/data",
					},
				},
			},
			Name: "example2",
		},
		{
			Node: &proto.Instance{
				ID:      "id",
				Cluster: "b",
				Name:    "c",
				Spec: &proto.NodeSpec{
					Image:   "image",
					Version: "latest",
					Files: []*proto.NodeSpec_File{
						{
							Name:    "/data/a",
							Content: "a",
						},
						{
							Name:    "/data/b",
							Content: "b",
						},
					},
				},
			},
			Name: "example3",
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

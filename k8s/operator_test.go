package k8s

import (
	"bytes"
	"strings"
	"testing"
	"time"

	gproto "github.com/golang/protobuf/proto"

	"github.com/hashicorp/go-hclog"
	"github.com/teseraio/ensemble/operator/proto"
)

func createOpCRDs(t *testing.T, p *Provider) func() {
	if err := p.createCRD(MustAsset("resources/crd-cluster.json")); err != nil {
		if err != errAlreadyExists {
			t.Fatal(err)
		}
	}
	if err := p.createCRD(MustAsset("resources/crd-resource.json")); err != nil {
		if err != errAlreadyExists {
			t.Fatal(err)
		}
	}

	closeFn := func() {
		if err := p.delete(crdURL+"/clusters.ensembleoss.io", emptyDel); err != nil {
			t.Fatal(err)
		}
		if err := p.delete(crdURL+"/resources.ensembleoss.io", emptyDel); err != nil {
			t.Fatal(err)
		}
	}
	return closeFn
}

func TestItemDecoding(t *testing.T) {
	p, _ := K8sFactory(hclog.NewNullLogger(), nil)
	p.Setup()

	// create CRDs for clusters and resources
	closeFn := createOpCRDs(t, p)
	defer closeFn()

	// wait for the CRD to be created
	time.Sleep(500 * time.Millisecond)

	var cases = []struct {
		item string
		spec gproto.Message
	}{
		{
			item: `{
				"backend": {
					"name": "a"
				},
				"replicas": 1
			}`,
			spec: &proto.ClusterSpec{
				Backend:  "a",
				Replicas: 1,
			},
		},
		{
			item: `{
				"backend": "b",
				"cluster": "c",
				"resource": "r"
			}`,
			spec: &proto.ResourceSpec{
				Backend:  "b",
				Cluster:  "c",
				Resource: "r",
			},
		},
		{
			item: `{
				"backend": "b",
				"cluster": "c",
				"resource": "r",
				"params": {
					"a": "b"
				}
			}`,
			spec: &proto.ResourceSpec{
				Backend:  "b",
				Cluster:  "c",
				Resource: "r",
				Params:   "{\"a\":\"b\"}",
			},
		},
	}
	for _, c := range cases {
		isErr := c.spec == nil

		var kind string
		if _, ok := c.spec.(*proto.ClusterSpec); ok {
			kind = "clusters"
		} else {
			kind = "resources"
		}

		obj, err := RunTmpl(`{
			"apiVersion": "ensembleoss.io/v1",
			"kind": "{{.kind}}",
			"metadata": {
				"name": "a"
			},
			"spec": {{.spec}}
		}`, map[string]interface{}{
			"kind": strings.Title(kind[:len(kind)-1]),
			"spec": c.item,
		})
		if err != nil {
			t.Fatal(err)
		}

		// Create object
		url := "/apis/ensembleoss.io/v1/namespaces/default/" + kind
		if _, _, err := p.post(url, obj); err != nil {
			if err == nil && isErr {
				t.Fatal("bad")
			}
			if err != nil && !isErr {
				t.Fatal(err)
			}
		}
		if isErr {
			continue
		}

		// Read object and compare
		var item *Item
		if _, err := p.get("/apis/ensembleoss.io/v1/namespaces/default/"+kind+"/a", &item); err != nil {
			t.Fatal(err)
		}
		spec, err := decodeItem(item)
		if err != nil {
			t.Fatal(err)
		}

		expected := proto.MustMarshalAny(c.spec)
		if !bytes.Equal(expected.Value, spec.Value) {
			t.Fatal("bad")
		}

		// Delete object
		if err := p.delete("/apis/ensembleoss.io/v1/namespaces/default/"+kind+"/a", emptyDel); err != nil {
			t.Fatal(err)
		}
	}
}

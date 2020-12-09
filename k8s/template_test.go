package k8s

import (
	"encoding/json"
	"io/ioutil"
	"reflect"
	"testing"
)

func TestTemplate(t *testing.T) {
	type Case struct {
		Name     string
		Template string
		Obj      interface{}
	}

	cases := []*Case{
		{
			"example1",
			"generic",
			map[string]interface{}{
				"Status": map[string]interface{}{
					"observedGeneration": "1",
				},
				"Metadata": map[string]interface{}{
					"a": "b",
				},
				"ResourceVersion": "2",
			},
		},
	}

	data, err := ioutil.ReadFile("./resources/fixtures/template.json")
	if err != nil {
		t.Fatal(err)
	}

	var fixtures map[string]interface{}
	if err := json.Unmarshal(data, &fixtures); err != nil {
		t.Fatal(err)
	}

	for _, c := range cases {
		expected := fixtures[c.Name]

		res, err := RunTmpl2(c.Template, c.Obj)
		if err != nil {
			t.Fatal(err)
		}

		var found map[string]interface{}
		if err := json.Unmarshal(res, &found); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(expected, found) {
			t.Fatal("bad")
		}
	}
}

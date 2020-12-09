package schema

import (
	"encoding/json"
	"io/ioutil"
	"testing"
)

func TestOpenAPIV3Marshal(t *testing.T) {
	cases := []struct {
		filename string
		schema   *Record
	}{
		{
			"openapiv3.json",
			&Record{
				Fields: map[string]*Field{
					"a": {
						Type: TypeString,
					},
					"b": {
						Type: &Record{
							Fields: map[string]*Field{
								"c": {
									Type:     TypeInt,
									Required: true,
								},
							},
						},
					},
					"d": {
						Type: TypeString,
					},
					"e": {
						Type: &Array{
							Elem: &Record{
								Fields: map[string]*Field{
									"e1": {
										Type: TypeString,
									},
									"e2": {
										Type: TypeString,
									},
								},
							},
						},
					},
				},
			},
		},
	}
	for _, c := range cases {
		raw, err := ioutil.ReadFile("./fixtures/" + c.filename)
		if err != nil {
			t.Fatal(err)
		}

		// prettify the json
		var aux map[string]interface{}
		if err := json.Unmarshal(raw, &aux); err != nil {
			t.Fatal(err)
		}
		expected, err := json.Marshal(aux)
		if err != nil {
			t.Fatal(err)
		}

		res, err := c.schema.OpenAPIV3JSON()
		if err != nil {
			t.Fatal(err)
		}

		if string(res) != string(expected) {
			t.Fatal("bad")
		}
	}
}

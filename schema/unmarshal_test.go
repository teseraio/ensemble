package schema

import (
	"testing"
)

func TestValidate(t *testing.T) {
	var struct1 struct {
		A string `schema:"a,required"`
		B struct {
			C string `schema:",required"`
		}
	}

	cases := []struct {
		obj   interface{}
		input map[string]interface{}
		err   bool
	}{
		{
			obj: struct1,
			input: map[string]interface{}{
				"a": "a",
				"B": map[string]interface{}{
					"C": "c",
				},
			},
			err: false,
		},
		{
			obj: struct1,
			input: map[string]interface{}{
				"a": "a",
				"B": map[string]interface{}{},
			},
			err: true,
		},
		{
			obj: struct {
				A string `schema:"a,required"`
			}{},
			input: map[string]interface{}{
				"a": "",
			},
			err: false,
		},
		{
			obj: struct {
				A string `schema:",required"`
			}{},
			input: map[string]interface{}{},
			err:   true,
		},
	}
	for _, c := range cases {
		err := ValidateRequired(c.input, c.obj)
		if err != nil && !c.err {
			t.Fatal(err)
		}
		if err == nil && c.err {
			t.Fatal("bad")
		}
	}
}

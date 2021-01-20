package schema

import (
	"fmt"
	"testing"
)

func TestValidate(t *testing.T) {
	cases := []struct {
		obj   interface{}
		input map[string]interface{}
		err   bool
	}{
		{
			obj: struct {
				A string `schema:"a,required"`
				B struct {
					C string `schema:",required"`
				}
			}{},
			input: map[string]interface{}{
				"a": "a",
				"B": map[string]interface{}{
					"C": "c",
				},
			},
			err: false,
		},
	}
	for _, c := range cases {
		fmt.Println(ValidateRequired(c.input, c.obj))
	}
}

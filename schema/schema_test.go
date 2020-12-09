package schema

import (
	"fmt"
	"testing"
)

func TestSchema(t *testing.T) {
	var s struct {
		A string `schema:",required"`
	}
	schema, err := GenerateSchema(s)
	if err != nil {
		t.Fatal(err)
	}

	out, err := schema.Spec.OpenAPIV3JSON()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(out))
}

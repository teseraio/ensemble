package schema

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/teseraio/ensemble/operator/proto"
)

func TestSchema2Get(t *testing.T) {
	schema := &Schema2{
		Spec: &Record{
			Fields: map[string]*Field{
				"a": {
					Type: TypeInt,
				},
				"b": {
					Type: &Record{
						Fields: map[string]*Field{
							"c": {
								Type: TypeString,
							},
						},
					},
				},
			},
		},
	}
	fmt.Println(schema.Get("b.c"))
}

func TestSchema2Diff(t *testing.T) {
	cases := []struct {
		Schema Schema2
		Old    map[string]interface{}
		New    map[string]interface{}
	}{
		{
			Schema2{
				Spec: &Record{
					Fields: map[string]*Field{
						"a": {
							Type: TypeString,
						},
					},
				},
			},
			map[string]interface{}{
				"a": "a",
			},
			map[string]interface{}{
				"a": "b",
			},
		},
	}
	fmt.Println(cases)

	for _, i := range cases {
		res := i.Schema.Diff(MapToSpec(i.Old), MapToSpec(i.New))
		fmt.Println(res)
	}
}

func TestMapToSpec(t *testing.T) {
	cases := []struct {
		Map  map[string]interface{}
		Spec *proto.Spec
	}{
		{
			map[string]interface{}{
				"a": "b",
				"c": 1,
			},
			proto.BlockSpec(&proto.Spec_Block{
				Attrs: map[string]*proto.Spec{
					"a": proto.LiteralSpec(&proto.Spec_Literal{
						Value: "b",
					}),
					"c": proto.LiteralSpec(&proto.Spec_Literal{
						Value: "1",
					}),
				},
			}),
		},
		{
			map[string]interface{}{
				"a": map[string]interface{}{
					"b": "c",
					"d": []string{
						"d1",
						"d2",
					},
				},
			},
			proto.BlockSpec(&proto.Spec_Block{
				Attrs: map[string]*proto.Spec{
					"a": proto.BlockSpec(&proto.Spec_Block{
						Attrs: map[string]*proto.Spec{
							"b": proto.LiteralSpec(&proto.Spec_Literal{
								Value: "c",
							}),
							"d": proto.ArraySpec([]*proto.Spec{
								proto.LiteralSpec(&proto.Spec_Literal{
									Value: "d1",
								}),
								proto.LiteralSpec(&proto.Spec_Literal{
									Value: "d2",
								}),
							}),
						},
					}),
				},
			}),
		},
	}

	for _, c := range cases {
		res := MapToSpec(c.Map)
		if !reflect.DeepEqual(res, c.Spec) {
			t.Fatal("bad")
		}
	}
}

func TestValidate(t *testing.T) {
	cases := []struct {
		rec  *Record
		spec map[string]interface{}
	}{
		{
			&Record{
				Fields: map[string]*Field{
					"a": {
						Type: &Record{
							Fields: map[string]*Field{
								"b": {
									Type: TypeString,
								},
							},
						},
					},
				},
			},
			map[string]interface{}{
				"a": map[string]interface{}{
					"b": "c",
				},
			},
		},
		{
			&Record{
				Fields: map[string]*Field{
					"a": {
						Type:     TypeString,
						Required: true,
					},
				},
			},
			map[string]interface{}{},
		},
		{
			&Record{
				Fields: map[string]*Field{
					"a": {
						Type: &Array{
							Elem: TypeString,
						},
					},
				},
			},
			map[string]interface{}{
				"a": []interface{}{
					1,
					2,
				},
			},
		},
	}

	for _, c := range cases {
		fmt.Println(validate(c.rec, MapToSpec(c.spec)))
	}
}

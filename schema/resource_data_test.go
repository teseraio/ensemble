package schema

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/teseraio/ensemble/operator/proto"
)

func TestResourceData(t *testing.T) {
	rsc := &ResourceData{
		sch: &Schema2{
			Spec: &Record{
				Fields: map[string]*Field{
					"a": {
						Type: TypeInt,
					},
					"b": {
						Type: &Array{
							Elem: TypeString,
						},
					},
					"c": {
						Type: &Record{
							Fields: map[string]*Field{
								"d": {
									Type: TypeString,
								},
								"e": {
									Type: &Array{
										Elem: TypeInt,
									},
								},
							},
						},
					},
				},
			},
		},
		flatmap: map[string]string{
			"a":     "1",
			"b.#":   "2",
			"b.0":   "b0",
			"b.1":   "b1",
			"c.d":   "d",
			"c.e.#": "2",
			"c.e.0": "1",
			"c.e.1": "2",
		},
	}
	// assert.Equal(t, rsc.Get("a").(string), "1")
	fmt.Println(rsc.Get("b").([]interface{}))
	fmt.Println(rsc.Get("c"))
}

func TestFlattenSpec(t *testing.T) {
	cases := []struct {
		spec    *proto.Spec
		flatten map[string]string
	}{
		{
			MapToSpec(map[string]interface{}{
				"a": map[string]interface{}{
					"b": "c",
				},
				"d": "e",
				"f": []interface{}{
					"g1",
					"g2",
				},
				"h": []interface{}{
					map[string]interface{}{
						"i": "j",
						"k": "l",
					},
				},
			}),
			map[string]string{
				"a.b":   "c",
				"d":     "e",
				"f.#":   "2",
				"f.0":   "g1",
				"f.1":   "g2",
				"h.#":   "1",
				"h.0.i": "j",
				"h.0.k": "l",
			},
		},
	}

	for _, c := range cases {
		assert.Equal(t, c.flatten, flatten(c.spec))
	}
}

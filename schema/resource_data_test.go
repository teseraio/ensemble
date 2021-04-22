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
				},
			},
		},
		flatmap: map[string]string{
			"a": "1",
		},
	}

	fmt.Println(rsc)

	fmt.Println(rsc.Get("a"))
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
			}),
			map[string]string{
				"d":   "e",
				"a.b": "c",
			},
		},
	}

	for _, c := range cases {
		assert.Equal(t, c.flatten, flatten(c.spec))
	}
}

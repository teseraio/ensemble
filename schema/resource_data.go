package schema

import (
	"strings"

	"github.com/teseraio/ensemble/operator/proto"
)

func NewResourceData(sch *Schema2, data *proto.Spec) *ResourceData {
	return &ResourceData{
		sch:     sch,
		flatmap: flatten(data),
	}
}

type ResourceData struct {
	sch     *Schema2
	flatmap map[string]string
}

func (r *ResourceData) Get(k string) interface{} {
	val, _ := r.GetOK(k)
	return val
}

func (r *ResourceData) GetOK(k string) (interface{}, bool) {
	// get the field in the schema
	_, err := r.sch.Get(k)
	if err != nil {
		return nil, false
	}
	val, ok := r.flatmap[k]
	if !ok {
		return nil, false
	}
	return val, true
}

func (r *ResourceData) Decode(obj interface{}) error {
	return nil
}

func flatten(s *proto.Spec) (res map[string]string) {
	if !isBlockValue(s) {
		panic("BUG: Only can flatten block values")
	}

	res = map[string]string{}
	var flat func(parent []string, s *proto.Spec)

	flat = func(parent []string, s *proto.Spec) {
		obj := s.Block.(*proto.Spec_BlockValue)
		for key, attr := range obj.BlockValue.Attrs {
			subKey := append(parent, key)
			if isBlockValue(attr) {
				flat(subKey, attr)
			} else {
				res[strings.Join(subKey, ".")] = attr.Block.(*proto.Spec_Literal_).Literal.Value
			}
		}
	}
	flat([]string{}, s)
	return
}

func isBlockValue(obj *proto.Spec) bool {
	_, ok := obj.Block.(*proto.Spec_BlockValue)
	return ok
}

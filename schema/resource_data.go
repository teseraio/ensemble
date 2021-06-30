package schema

import (
	"strconv"
	"strings"

	"github.com/teseraio/ensemble/operator/proto"
)

func NewResourceData(sch *Schema2, data *proto.Spec) *ResourceData {
	r := &ResourceData{
		sch:     sch,
		flatmap: flatten(data),
	}
	// fill in default values
	for k, f := range sch.Spec.Fields {
		if f.Default != "" {
			if _, ok := r.flatmap[k]; !ok {
				r.flatmap[k] = f.Default
			}
		}
	}
	return r
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
	field, err := r.sch.Get(k)
	if err != nil {
		return nil, false
	}
	return r.getField(field.Type, k)
}

func (r *ResourceData) getField(field Type, k string) (interface{}, bool) {
	switch obj := field.(type) {
	case *Array:
		return r.getArray(obj, k)

	case ScalarType:
		return r.getLiteral(obj, k)

	default:
		panic("BUG: getOk not found")
	}
}

func (r *ResourceData) getLiteral(field ScalarType, k string) (interface{}, bool) {
	val, ok := r.flatmap[k]
	if !ok {
		return nil, false
	}
	return val, true
}

func (r *ResourceData) getArray(field *Array, k string) (interface{}, bool) {
	num, err := strconv.Atoi(r.flatmap[k+".#"])
	if err != nil {
		panic(err)
	}
	values := []interface{}{}
	for i := 0; i < num; i++ {
		prefix := k + "." + strconv.Itoa(i)
		if val, ok := r.getField(field.Elem, prefix); ok {
			values = append(values, val)
		}
	}
	return values, true
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

		switch obj := s.Block.(type) {
		case *proto.Spec_BlockValue:
			for key, attr := range obj.BlockValue.Attrs {
				flat(append(parent, key), attr)
			}

		case *proto.Spec_Array_:
			res[strings.Join(append(parent, "#"), ".")] = strconv.Itoa(len(obj.Array.Values))
			for indx, val := range obj.Array.Values {
				flat(append(parent, strconv.Itoa(indx)), val)
			}

		case *proto.Spec_Literal_:
			res[strings.Join(parent, ".")] = obj.Literal.Value

		default:
			panic("BUG: Spec type not found")
		}

		/*
			obj := s.Block.(*proto.Spec_BlockValue)
			for key, attr := range obj.BlockValue.Attrs {
				subKey := append(parent, key)
				if isBlockValue(attr) {
					flat(subKey, attr)
				} else {
					res[strings.Join(subKey, ".")] = attr.Block.(*proto.Spec_Literal_).Literal.Value
				}
			}
		*/
	}
	flat([]string{}, s)
	return
}

func isBlockValue(obj *proto.Spec) bool {
	_, ok := obj.Block.(*proto.Spec_BlockValue)
	return ok
}

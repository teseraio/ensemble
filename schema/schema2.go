package schema

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/teseraio/ensemble/operator/proto"
)

type Schema2 struct {
	Spec *Record
}

func (s *Schema2) Get(k string) (*Field, error) {
	parts := strings.Split(k, ".")

	rec := s.Spec
	for {
		var key string
		key, parts = parts[0], parts[1:]

		field, ok := rec.Fields[key]
		if !ok {
			return nil, fmt.Errorf("field %s not found", key)
		}
		if len(parts) == 0 {
			return field, nil
		}
		// there are more parts, it must be a map
		obj, ok := field.Type.(*Record)
		if !ok {
			return nil, fmt.Errorf("its not a record")
		}
		rec = obj
	}
}

func (s *Schema2) Validate(spec *proto.Spec) error {
	return validate(s.Spec, spec)
}

func (s *Schema2) Diff(i, j *proto.Spec) Diff {
	d := Diff{}
	d.build(s.Spec, i, j)
	return d
}

type DiffField struct {
	Old interface{}
	New interface{}
}

type Diff map[string]DiffField

func (d *Diff) build(spec *Record, i, j *proto.Spec) Diff {
	d.diffImpl([]string{}, spec, i, j)
	return Diff{}
}

func (d *Diff) append(key []string, diff DiffField) {
	(*d)[strings.Join(key, ".")] = diff
}

func (d *Diff) diffImpl(key []string, t Type, iRaw, jRaw *proto.Spec) error {
	switch obj := t.(type) {
	case *Record:
		i := iRaw.Block.(*proto.Spec_BlockValue)
		j := jRaw.Block.(*proto.Spec_BlockValue)

		for name, field := range obj.Fields {
			subKey := append(key, name)
			d.diffImpl(subKey, field.Type, i.BlockValue.Attrs[name], j.BlockValue.Attrs[name])
		}

	case ScalarType:
		i := iRaw.Block.(*proto.Spec_Literal_)
		j := jRaw.Block.(*proto.Spec_Literal_)

		switch obj {
		case TypeString:
			if i.Literal.Value != j.Literal.Value {
				d.append(key, DiffField{
					Old: i,
					New: j,
				})
			}
		}
	default:
		panic(fmt.Sprintf("Not found: %s", t.Type()))
	}
	return nil
}

func MapToSpec(m map[string]interface{}) *proto.Spec {
	if m == nil {
		m = map[string]interface{}{}
	}
	var impl func(v reflect.Value) *proto.Spec

	impl = func(v reflect.Value) *proto.Spec {
		switch v.Type().Kind() {
		case reflect.Map:
			block := &proto.Spec_Block{
				Attrs: map[string]*proto.Spec{},
			}
			for _, k := range v.MapKeys() {
				elem := impl(v.MapIndex(k))
				block.Attrs[k.String()] = elem
			}
			return proto.BlockSpec(block)

		case reflect.Interface:
			return impl(v.Elem())

		case reflect.String:
			return proto.LiteralSpec(&proto.Spec_Literal{Value: v.String()})

		case reflect.Slice:
			values := []*proto.Spec{}
			for i := 0; i < v.Len(); i++ {
				elem := impl(v.Index(i))
				values = append(values, elem)
			}
			return proto.ArraySpec(values)

		case reflect.Int:
			return proto.LiteralSpec(&proto.Spec_Literal{Value: strconv.Itoa(int(v.Int()))})

		default:
			panic("NOT FOUND" + v.Type().Kind().String())
		}
	}

	res := impl(reflect.ValueOf(m))
	return res
}

func validate(t Type, s *proto.Spec) error {
	switch obj := t.(type) {
	case *Record:
		attrs := s.Block.(*proto.Spec_BlockValue).BlockValue.Attrs
		for k, field := range obj.Fields {
			val, ok := attrs[k]
			if ok {
				if err := validate(field.Type, val); err != nil {
					return err
				}
			} else {
				if field.Required {
					return fmt.Errorf("value '%s' not found", k)
				}
			}
		}
		return nil

	case *Array:
		arraySch, ok := s.Block.(*proto.Spec_Array_)
		if !ok {
			return fmt.Errorf("array expected")
		}
		for _, v := range arraySch.Array.Values {
			if err := validate(obj.Elem, v); err != nil {
				return err
			}
		}
		return nil

	case ScalarType:
		val := s.Block.(*proto.Spec_Literal_).Literal.Value
		switch obj {
		case TypeString:
			return nil

		case TypeInt:
			if _, err := strconv.Atoi(val); err != nil {
				return err
			}
		}

	default:
		panic(fmt.Sprintf("BUG: type not found %v", reflect.TypeOf(obj)))
	}
	return nil
}

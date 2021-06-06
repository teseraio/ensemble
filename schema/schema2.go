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
	var impl func(i interface{}) (*proto.Spec, error)

	impl = func(i interface{}) (*proto.Spec, error) {
		switch obj := i.(type) {
		case map[string]interface{}:
			block := &proto.Spec_Block{
				Attrs: map[string]*proto.Spec{},
			}
			for k, v := range obj {
				elem, err := impl(v)
				if err != nil {
					return nil, err
				}
				block.Attrs[k] = elem
			}
			return proto.BlockSpec(block), nil

		case string:
			return proto.LiteralSpec(&proto.Spec_Literal{Value: obj}), nil

		default:
			return nil, fmt.Errorf("type not found %s", reflect.TypeOf(obj))
		}
	}

	res, err := impl(m)
	if err != nil {
		panic(fmt.Sprintf("BUG: %v", err))
	}
	return res
}

func validate(t Type, s *proto.Spec) error {
	fmt.Println("-- validate --")

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

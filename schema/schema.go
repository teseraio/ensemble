package schema

import (
	"fmt"
	"reflect"
	"strings"
)

// Schema is an OpenAPIv3 schema of a resource
type Schema struct {
	Spec   *Record
	Status *Record
}

// GenerateSchema generates an OpenAPIv3 schema
func GenerateSchema(s interface{}) (*Schema, error) {
	t := reflect.TypeOf(s)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected struct")
	}

	out, err := generateStruct(t)
	if err != nil {
		return nil, err
	}

	// split the objects between spec and status given
	// the computed field
	schema := &Schema{
		Spec:   &Record{},
		Status: &Record{},
	}

	record := out.(*Record)
	for n, f := range record.Fields {
		if f.Computed {
			schema.Status.addField(n, f)
		} else {
			schema.Spec.addField(n, f)
		}
	}

	return schema, nil
}

func generateImpl(t reflect.Type) (Type, error) {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	switch t.Kind() {
	case reflect.Struct:
		return generateStruct(t)

	case reflect.String:
		return TypeString, nil

	case reflect.Bool:
		return TypeBool, nil

	case reflect.Map:
		return &Map{}, nil

	default:
		panic(fmt.Sprintf("BUG: Type not found %s", t.Kind()))
	}
}

func generateStruct(t reflect.Type) (Type, error) {
	r := &Record{
		Fields: map[string]*Field{},
	}

	fields := t.NumField()
	for i := 0; i < fields; i++ {
		elem := t.Field(i)

		typ, err := generateImpl(elem.Type)
		if err != nil {
			return nil, err
		}

		tag := newTag(elem.Tag)
		f := &Field{
			Type:     typ,
			Computed: tag.contains("computed"),
			Required: tag.contains("required"),
		}

		name := tag.name
		if name == "" {
			name = strings.ToLower(elem.Name)
		}
		if tag.contains("squash") {
			// Squash a record into this structure
			record, ok := f.Type.(*Record)
			if !ok {
				return nil, fmt.Errorf("cannot squash a non record type")
			}
			for name, f := range record.Fields {
				r.Fields[name] = f
			}
		} else {
			r.Fields[name] = f
		}
	}
	return r, nil
}

type tag struct {
	name string
	tags []string
}

func newTag(t reflect.StructTag) *tag {
	tt := &tag{}
	tags := strings.Split(t.Get("schema"), ",")
	if len(tags) != 0 {
		tt.name, tt.tags = tags[0], tags[1:]
	}
	return tt
}

func (t *tag) contains(j string) bool {
	for _, i := range t.tags {
		if i == j {
			return true
		}
	}
	return false
}

package schema

import "reflect"

//go:generate stringer -type=ScalarType -output=types_string.go

// Type describes a valid type for the schema
type Type interface {
	Type() string
	Kind() reflect.Type
}

// batch of predefined reflect types
var (
	boolT   = reflect.TypeOf(bool(false))
	stringT = reflect.TypeOf("")
	tupleT  = reflect.TypeOf(map[string]interface{}{})
)

const (
	// TypeString is a string type
	TypeString ScalarType = iota

	// TypeInt is a int type
	TypeInt

	// TypeBool is a bool type
	TypeBool
)

// ScalarType is a scalar type
type ScalarType int

// Type implements the Type interface
func (s ScalarType) Type() string {
	return s.String()
}

func (s ScalarType) Kind() reflect.Type {
	return stringT
}

// Array is an array of objects
type Array struct {
	Elem Type
}

// Type implements the Type interface
func (a *Array) Type() string {
	return "array"
}

func (a *Array) Kind() reflect.Type {
	panic("X")
}

// Map is a map<string, interface{}> without specific types
type Map struct {
	Elem Type
}

// Type implements the Type interface
func (m *Map) Type() string {
	return "map"
}

func (m *Map) Kind() reflect.Type {
	return tupleT
}

// Record is an object with several values
type Record struct {
	Fields map[string]*Field
}

func (r *Record) addField(name string, f *Field) {
	if len(r.Fields) == 0 {
		r.Fields = map[string]*Field{}
	}
	r.Fields[name] = f
}

// Field is a field in a record
type Field struct {
	Type Type

	// Computed describes whether the field is a status
	Computed bool

	// Required specifies if the field is required
	Required bool

	// ForceNew describes whether a change in the field requires delete the old one
	ForceNew bool

	// Default value for the field
	Default string
}

// Type implements the Type interface
func (r *Record) Type() string {
	return "record"
}

func (r *Record) Kind() reflect.Type {
	return tupleT
}

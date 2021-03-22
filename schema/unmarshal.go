package schema

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"unicode"

	"github.com/mitchellh/mapstructure"
)

// DecodeString decodes a string into an object
func DecodeString(input string, obj interface{}) error {
	var input2 map[string]interface{}
	if err := json.Unmarshal([]byte(input), &input2); err != nil {
		return err
	}
	return Decode(input2, obj)
}

// Decode decodes a map into an object
func Decode(input map[string]interface{}, obj interface{}) error {
	if err := ValidateRequired(input, obj); err != nil {
		return err
	}
	dc := &mapstructure.DecoderConfig{
		Result:  obj,
		TagName: "schema",
	}
	ms, err := mapstructure.NewDecoder(dc)
	if err != nil {
		return err
	}
	if err = ms.Decode(input); err != nil {
		return err
	}
	return nil
}

func Encode(obj interface{}) (map[string]interface{}, error) {
	var res map[string]interface{}
	dc := &mapstructure.DecoderConfig{
		Result:  &res,
		TagName: "schema",
	}
	ms, err := mapstructure.NewDecoder(dc)
	if err != nil {
		return nil, err
	}
	if err = ms.Decode(obj); err != nil {
		return nil, err
	}
	return res, nil
}

func ValidateRequired(input map[string]interface{}, obj interface{}) error {
	// we have to do this manually since mapstructure does not allow required tags.

	// read all the values in obj that are required
	requiredFields := ReadByTag(obj, "required")

	// check if any of these fields is nil
	for _, field := range requiredFields {
		if _, ok := GetKey(input, field); ok {
			return fmt.Errorf("bad")
		}
	}

	return nil
}

func GetKey(input map[string]interface{}, key string) (interface{}, bool) {
	keys := strings.Split(strings.Trim(key, "."), ".")
	if len(keys) == 0 {
		return nil, false
	}

	val := input

	var elem string
	for {
		// pop a value
		elem, keys = keys[0], keys[1:]

		elemVal, ok := val[elem]
		if !ok {
			if unicode.IsUpper(rune(elem[0])) {
				// try in lowercase
				elem = string(unicode.ToLower(rune(elem[0]))) + elem[1:]
				elemVal, ok = val[elem]
				if !ok {
					// the key does not exists
					return nil, false
				}
			}
		}

		if len(keys) == 0 {
			return elemVal, true
		}

		// there are some keys left, elemVal must be a map
		elemMap, ok := elemVal.(map[string]interface{})
		if !ok {
			return nil, false
		}
		val = elemMap
	}
}

func ReadByTag(obj interface{}, target string) []string {
	res := []string{}

	var impl func(parent string, v reflect.Value)

	impl = func(parent string, v reflect.Value) {
		if v.Kind() != reflect.Struct {
			switch v.Kind() {
			case reflect.Ptr:
				impl("", v.Elem())

			case reflect.Interface:
				impl("", v.Elem())
			}
			return
		}

		typ := v.Type()
		for i := 0; i < v.NumField(); i++ {
			f := typ.Field(i)

			name := f.Name
			tags := f.Tag.Get("schema")

			parts := strings.Split(tags, ",")
			if len(parts) == 0 {
				// no tags, skip
				continue
			}

			// the first element is the name if any
			if parts[0] != "" {
				name = parts[0]
			}

			fullName := parent + "." + name
			for _, tag := range parts[1:] {
				if tag == target {
					res = append(res, fullName)
				}
			}
			if f.Type.Kind() == reflect.Struct {
				impl(fullName, v.Field(i))
			}
		}
	}

	impl("", reflect.ValueOf(obj))
	return res
}

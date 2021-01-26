package schema

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

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

func ValidateRequired(input map[string]interface{}, obj interface{}) error {
	// we have to do this manually since mapstructure does not allow required tags.

	// read all the values in obj that are required
	requiredFields := readRequiredFlags(obj)

	// check if any of these fields is nil
	for _, field := range requiredFields {
		if !existsKey(input, field) {
			return fmt.Errorf("bad")
		}
	}

	return nil
}

func existsKey(input map[string]interface{}, key string) bool {
	keys := strings.Split(strings.Trim(key, "."), ".")
	if len(keys) == 0 {
		return false
	}

	val := input

	var elem string
	for {
		// pop a value
		elem, keys = keys[0], keys[1:]

		elemVal, ok := val[elem]
		if !ok {
			// the key does not exists
			return false
		}

		if len(keys) == 0 {
			return true
		}

		// there are some keys left, elemVal must be a map
		elemMap, ok := elemVal.(map[string]interface{})
		if !ok {
			return false
		}
		val = elemMap
	}
}

func readRequiredFlags(obj interface{}) []string {
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
				if tag == "required" {
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

package schema

import (
	"encoding/json"

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

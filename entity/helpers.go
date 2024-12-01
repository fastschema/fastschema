package entity

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/buger/jsonparser"
	"github.com/mitchellh/mapstructure"
)

// NewEntityFromJSON creates a new entity from a JSON string.
func NewEntityFromJSON(jsonData string) (*Entity, error) {
	entity := New()
	if err := entity.UnmarshalJSON([]byte(jsonData)); err != nil {
		return nil, err
	}

	return entity, nil
}

// NewEntityFromMap creates a new entity from a map.
func NewEntityFromMap(data map[string]any) *Entity {
	entity := New()

	for key, value := range data {
		if valueMap, ok := value.(map[string]any); ok {
			entity.Set(key, NewEntityFromMap(valueMap))
			continue
		}

		entity.Set(key, value)
	}

	return entity
}

// BindEntity binds the fields of the given Entity to a target struct value.
func BindEntity[T any](e *Entity) (result T, err error) {
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName: "json",
		Result:  &result,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(func(from, to reflect.Type, data any) (any, error) {
			if e, ok := data.(*Entity); ok {
				return e.ToMap(), nil
			}

			return data, nil
		}),
	})

	if err != nil {
		return
	}

	if err = decoder.Decode(e); err != nil {
		return
	}

	return result, nil
}

// UnmarshalJSONValue converts json bytes to a Go value.
func UnmarshalJSONValue(
	data []byte,
	valueData []byte,
	dataType jsonparser.ValueType,
	offset int,
) (any, error) {
	// jsonparser removes the enclosing quotes; we need to restore them to make a valid JSON
	if dataType == jsonparser.String {
		// valueData = data[offset-len(valueData)-2 : offset]
		valueData = []byte(fmt.Sprintf("\"%s\"", string(valueData)))
	}

	var value any

	if err := json.Unmarshal(valueData, &value); err != nil {
		return nil, err
	}

	return value, nil
}

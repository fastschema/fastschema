package schema

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"

	"github.com/buger/jsonparser"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/mitchellh/mapstructure"
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

// Entity represents a single entity.
type Entity struct {
	// name is the name of the schema/model that the entity represents.
	name string
	data *orderedmap.OrderedMap[string, any]
}

// NamedEntity creates a new entity with the given model.
func NamedEntity(name string) *Entity {
	return &Entity{
		name: name,
		data: orderedmap.New[string, any](),
	}
}

// NewEntity creates a new entity.
func NewEntity(ids ...uint64) *Entity {
	entity := &Entity{
		data: orderedmap.New[string, any](),
	}

	if len(ids) > 0 {
		entity.Set(FieldID, ids[0])
	}

	return entity
}

// NewEntityFromJSON creates a new entity from a JSON string.
func NewEntityFromJSON(jsonData string) (*Entity, error) {
	entity := NewEntity()
	if err := entity.UnmarshalJSON([]byte(jsonData)); err != nil {
		return nil, err
	}

	return entity, nil
}

// NewEntityFromMap creates a new entity from a map.
func NewEntityFromMap(data map[string]any) *Entity {
	entity := NewEntity()

	for key, value := range data {
		if valueMap, ok := value.(map[string]any); ok {
			entity.Set(key, NewEntityFromMap(valueMap))
			continue
		}

		entity.Set(key, value)
	}

	return entity
}

// Name returns the model name of the entity. If names is provided, it sets the model name to the first name in the list.
func (e *Entity) Name(names ...string) string {
	if len(names) > 0 {
		e.name = names[0]
	}
	return e.name
}

// Empty returns true if the entity is empty.
func (e *Entity) Empty() bool {
	return e.data.Len() == 0
}

// Keys returns the keys of the entity.
func (e *Entity) Keys() []string {
	keys := make([]string, 0, e.data.Len())
	for pair := e.data.Oldest(); pair != nil; pair = pair.Next() {
		keys = append(keys, pair.Key)
	}
	return keys
}

// String returns the string representation of the entity.
func (e *Entity) String() string {
	str, _ := e.ToJSON()
	return str
}

// First returns the oldest key/value pair in the entity.
func (e *Entity) First() *orderedmap.Pair[string, any] {
	return e.data.Oldest()
}

// Set sets a value in the entity.
func (e *Entity) Set(name string, value any) *Entity {
	e.data.Set(name, value)
	return e
}

// Get returns a value from the entity.
func (e *Entity) Get(name string, defaultValues ...any) any {
	if value, present := e.data.Get(name); present {
		return value
	}

	if len(defaultValues) > 0 {
		return defaultValues[0]
	}

	return nil
}

// Delete removes a value from the entity.
func (e *Entity) Delete(name string) *Entity {
	e.data.Delete(name)
	return e
}

// GetString returns a string value from the entity.
func (e *Entity) GetString(name string, defaultValues ...string) string {
	if value, present := e.data.Get(name); present {
		if stringValue, ok := value.(string); ok {
			return stringValue
		}
	}

	if len(defaultValues) > 0 {
		return defaultValues[0]
	}

	return ""
}

// GetUint64 returns the foreign key value (uint64) from the entity.
func (e *Entity) GetUint64(name string, optional bool) (uint64, error) {
	value, ok := e.data.Get(name)
	if !ok {
		return 0, utils.If(optional, nil, fmt.Errorf("cannot get uint64 value from entity: %s", name))
	}

	idValue, ok := value.(uint64)

	if !ok {
		// try converting the value to uint64
		var err error
		idValue, err = strconv.ParseUint(fmt.Sprintf("%v", value), 10, 64)
		if err != nil {
			return 0, utils.If(optional, nil, fmt.Errorf(`invalid uint64 value %s=%v (%T)`, name, value, value))
		}
	}

	return idValue, nil
}

// ID returns the ID of the entity.
func (e *Entity) ID() uint64 {
	idValue := e.Get(FieldID)

	if idValue == nil {
		return 0
	}

	if idUint64, err := strconv.ParseUint(fmt.Sprintf("%v", idValue), 10, 64); err == nil {
		return idUint64
	}

	return 0
}

// SetID sets the ID of the entity.
// value can be uint64, float64, or a string that can be converted to uint64.
// if the value is not a valid ID, do nothing.
// returns the entity.
func (e *Entity) SetID(value any) *Entity {
	if uint64ID, ok := value.(uint64); ok {
		e.data.Set(FieldID, uint64ID)
	} else if float64ID, ok := value.(float64); ok {
		e.data.Set(FieldID, uint64(float64ID))
	}

	if uint64Value, err := strconv.ParseUint(fmt.Sprintf("%v", value), 10, 64); err == nil {
		e.data.Set(FieldID, uint64Value)
	}

	return e
}

// MarshalJSON implements the json.Marshaler interface.
func (e *Entity) MarshalJSON() ([]byte, error) {
	return e.data.MarshalJSON()
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (e *Entity) UnmarshalJSON(data []byte) (err error) {
	return jsonparser.ObjectEach(
		data,
		func(keyData []byte, valueData []byte, dataType jsonparser.ValueType, offset int) error {
			if dataType == jsonparser.Object {
				e.Set(string(keyData), utils.Must(NewEntityFromJSON(string(valueData))))
				return nil
			}

			// Process array data
			if dataType == jsonparser.Array {
				entityValues := []*Entity{}
				nonEntityValues := []any{}

				// Process array data
				if _, err := jsonparser.ArrayEach(valueData, func(
					itemValue []byte,
					itemDataType jsonparser.ValueType,
					itemOffset int,
					err error,
				) {
					if itemDataType == jsonparser.Object {
						entityValues = append(entityValues, utils.Must(NewEntityFromJSON(string(itemValue))))
						return
					}

					iArrayItemValue := utils.Must(UnmarshalJSONValue(data, itemValue, itemDataType, itemOffset))
					nonEntityValues = append(nonEntityValues, iArrayItemValue)
				}); err != nil {
					return err
				}

				/** Process array data **/
				if len(entityValues) > 0 && len(nonEntityValues) > 0 {
					return fmt.Errorf(
						"cannot mix entities and non-entities in a slice: %s=%v",
						string(keyData),
						string(valueData),
					)
				}

				if len(entityValues) > 0 {
					e.Set(string(keyData), entityValues)
					return nil
				}

				if len(nonEntityValues) > 0 {
					e.Set(string(keyData), nonEntityValues)
				}
				/** End process array data **/

				return nil
			}

			e.Set(string(keyData), utils.Must(UnmarshalJSONValue(data, valueData, dataType, offset)))

			return nil
		},
	)
}

// UnmarshalJSONValue converts json bytes to a Go value.
func UnmarshalJSONValue(data []byte, valueData []byte, dataType jsonparser.ValueType, offset int) (any, error) {
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

// ToJSON converts the entity to a JSON string.
func (e *Entity) ToJSON() (string, error) {
	jsonData, err := e.MarshalJSON()
	if err != nil {
		return "", err
	}

	return string(jsonData), nil
}

// ToMap converts the entity to a map.
func (e *Entity) ToMap() map[string]any {
	data := map[string]any{}
	for pair := e.data.Oldest(); pair != nil; pair = pair.Next() {
		if entityValue, ok := pair.Value.(*Entity); ok {
			data[pair.Key] = entityValue.ToMap()
			continue
		}

		if entitiesValue, ok := pair.Value.([]*Entity); ok {
			entityValues := []map[string]any{}
			for _, entity := range entitiesValue {
				entityValues = append(entityValues, entity.ToMap())
			}

			data[pair.Key] = entityValues
			continue
		}

		data[pair.Key] = pair.Value
	}

	return data
}

// func EntityToStruct[T any](e *Entity) (T, error) {
// 	var j T
// 	jsonData, err := json.Marshal(e)
// 	if err != nil {
// 		return j, err
// 	}

// 	_ = json.Unmarshal(jsonData, &j)
// 	return j, nil
// }

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

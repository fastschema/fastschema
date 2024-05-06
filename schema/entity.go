package schema

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/buger/jsonparser"
	"github.com/fastschema/fastschema/pkg/utils"
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

// Entity represents a single entity.
type Entity struct {
	data *orderedmap.OrderedMap[string, any]
}

// Empty returns true if the entity is empty.
func (e *Entity) Empty() bool {
	return e.data.Len() == 0
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
func (e *Entity) SetID(value any) error {
	if uint64ID, ok := value.(uint64); ok {
		e.data.Set(FieldID, uint64ID)
		return nil
	} else if float64ID, ok := value.(float64); ok {
		e.data.Set(FieldID, uint64(float64ID))
		return nil
	}

	valueUint64, err := strconv.ParseUint(fmt.Sprintf("%v", value), 10, 64)
	if err != nil {
		return fmt.Errorf("cannot convert ID value %v to uint64", value)
	}
	e.data.Set(FieldID, valueUint64)
	return nil
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

		data[pair.Key] = pair.Value
	}

	return data
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

// EntityToStruct converts an entity to a struct
func (e *Entity) EntityToStruct(s interface{}) interface{} {
	fmt.Println("start")
	// Get the type of the struct
	structType := reflect.TypeOf(s).Elem()

	// Create a new instance of the struct
	structValue := reflect.New(structType).Elem()

	// Iterate over the fields of the struct
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)

		entityFieldName := field.Name
		jsonTag := strings.Split(field.Tag.Get("json"), ",")[0]

		// Check if the field has a json tag
		if jsonTag != "" {
			entityFieldName = jsonTag
		}

		// Get the corresponding value from the schema.Entity
		structField := structValue.FieldByName(field.Name)
		value := e.Get(entityFieldName)

		// Check if the value is not nil
		if value == nil {
			continue
		}

		// Check if the field is a struct
		if field.Type.Kind() == reflect.Struct {
			// value must be an entity, otherwise the conversion will fail
			entityValue, ok := value.(*Entity)
			if !ok {
				panic(fmt.Sprintf("value must be an entity, got %v", value))
			}

			// Recursively convert the nested struct
			nestedStruct := reflect.New(field.Type).Interface()
			nestedStruct = entityValue.EntityToStruct(nestedStruct)
			structField.Set(reflect.ValueOf(nestedStruct))
			continue
		}

		if field.Type.Kind() == reflect.Slice || field.Type.Kind() == reflect.Array {
			fmt.Println("field.Type.Kind()", field.Type.Kind())
			// Check if all items are primitive types and have the same type
			sliceValue := reflect.ValueOf(value)
			// sliceType := sliceValue.Type()
			if isPrimitiveSlice(sliceValue) {
				// Set the value of the struct field
				structField.Set(reflect.ValueOf(value))
				continue
			} else {
				// reflect.Struct, reflect.Map, reflect.Slice, reflect.Array
				// check if slice type is reflect.Struct
				fmt.Println("fieldfieldfield", field)
				if field.Type.Elem().Kind() == reflect.Struct {
					// value must be a slice of entities, otherwise the conversion will fail
					entitySlice, ok := value.([]*Entity)
					fmt.Println("entitySlice", entitySlice)
					if !ok {
						panic(fmt.Sprintf("value must be a slice of entities, got %v", value))
					}
					// Create a new slice to hold the converted structs
					structSlice := reflect.MakeSlice(field.Type, len(entitySlice), len(entitySlice))
					// Iterate over the entities and convert them to structs
					for i, entity := range entitySlice {
						fmt.Println("entityentity", entity)
						structValue := reflect.New(field.Type.Elem()).Interface()
						structValue = entity.EntityToStruct(structValue)
						structSlice.Index(i).Set(reflect.ValueOf(structValue).Elem())
					}
					// Set the value of the struct field
					structField.Set(structSlice)
					continue
				}

				// check if slice type is reflect.Slice, reflect.Array

			}

		}

		// Set the value of the struct field
		structField.Set(reflect.ValueOf(value))
	}

	return structValue.Interface()
}

func isPrimitiveSlice(slice interface{}) bool {
	val := reflect.ValueOf(slice)

	// Check if the input is a slice or array
	if val.Kind() != reflect.Slice && val.Kind() != reflect.Array {
		return false
	}

	// Get the type of the elements in the slice or array
	elemType := val.Type().Elem()

	// Check if the element type is a primitive type
	switch elemType.Kind() {
	case reflect.Bool, reflect.String:
		return true
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return true
	case reflect.Float32, reflect.Float64:
		return true
	case reflect.Complex64, reflect.Complex128:
		return true
	case reflect.Struct, reflect.Map, reflect.Slice, reflect.Array:
		return false
	default:
		return false
	}
}

package entity

import (
	"fmt"

	"github.com/fastschema/fastschema/pkg/utils"
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

const FieldID = "id"
const FieldCreatedAt = "created_at"
const FieldUpdatedAt = "updated_at"
const FieldDeletedAt = "deleted_at"

type Entity struct {
	data *orderedmap.OrderedMap[string, any]
}

// New creates a new entity.
func New(ids ...uint64) *Entity {
	entity := &Entity{
		data: orderedmap.New[string, any](),
	}

	if len(ids) > 0 {
		entity.Set(FieldID, ids[0])
	}

	return entity
}

func (e *Entity) Data() *orderedmap.OrderedMap[string, any] {
	return e.data
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
	str, err := e.ToJSON()
	if err != nil {
		return fmt.Errorf("cannot convert entity to string: %w", err).Error()
	}

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

// SetID sets the ID of the entity.
// if the value is not a valid ID, returns the error.
func (e *Entity) SetID(value any) error {
	uint64Value, err := utils.AnyToUint[uint64](value)
	if err != nil {
		return fmt.Errorf("cannot set entity id=%s (%T): %w", value, value, err)
	}

	e.data.Set(FieldID, uint64Value)

	return nil
}

// ID returns the ID of the entity.
func (e *Entity) ID() uint64 {
	idValue, _ := e.GetUint64(FieldID, true)
	return idValue
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
func (e *Entity) GetUint64(name string, optional bool) (uint64Value uint64, err error) {
	defer func() {
		if err != nil && optional {
			err = nil
		}
	}()

	value, ok := e.data.Get(name)
	if !ok {
		err = fmt.Errorf("cannot get uint64 value from entity: %s", name)
		return
	}

	uint64Value, err = utils.AnyToUint[uint64](value)
	if err != nil {
		err = fmt.Errorf("cannot get uint64 value from entity, %s=%v (%T), error: %w", name, value, value, err)
	}

	return
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

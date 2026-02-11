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
	data    *orderedmap.OrderedMap[string, any]
	idField string
}

// New creates a new entity.
func New(ids ...any) *Entity {
	entity := &Entity{
		data:    orderedmap.New[string, any](),
		idField: FieldID,
	}

	if len(ids) > 0 {
		_ = entity.SetID(ids[0])
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

// NewWithIDField creates a new entity with a specific id field.
func NewWithIDField(idField string, ids ...any) *Entity {
	e := New()
	e.SetIDField(idField)
	if len(ids) > 0 {
		_ = e.SetID(ids[0])
	}
	return e
}

// SetIDField configures which field should be treated as the id field for this entity.
func (e *Entity) SetIDField(fieldName string) *Entity {
	if e == nil {
		return e
	}
	if fieldName == "" {
		e.idField = FieldID
		return e
	}
	e.idField = fieldName
	return e
}

// GetIDField returns the configured id field name for the entity.
func (e *Entity) GetIDField() string {
	if e == nil {
		return FieldID
	}
	if e.idField == "" {
		return FieldID
	}
	return e.idField
}

// SetID sets the entity ID using the configured id field.
func (e *Entity) SetID(value any) error {
	if value == nil {
		return fmt.Errorf("cannot set entity id=<nil>")
	}

	idField := e.GetIDField()
	if idField != FieldID {
		e.data.Set(idField, value)
	}
	e.data.Set(FieldID, value)

	return nil
}

// ID returns the ID of the entity.
func (e *Entity) ID() any {
	idField := e.GetIDField()
	if idField != "" {
		if value, ok := e.data.Get(idField); ok {
			return value
		}
	}
	if idField != FieldID {
		if value, ok := e.data.Get(FieldID); ok {
			return value
		}
	}
	return nil
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

package schema

import (
	"fmt"

	"github.com/fastschema/fastschema/pkg/utils"
)

// Field define the data struct for a field
type Field struct {
	Type       FieldType      `json:"type"`
	Name       string         `json:"name"`
	Label      string         `json:"label"`
	IsMultiple bool           `json:"multiple,omitempty"` // Is a multiple field.
	Renderer   *FieldRenderer `json:"renderer,omitempty"` // renderer of the field.
	Size       int64          `json:"size,omitempty"`     // max size parameter for string, blob, etc.
	Unique     bool           `json:"unique,omitempty"`   // column with unique constraint.
	Optional   bool           `json:"optional,omitempty"` // null or not null attribute.
	Default    any            `json:"default,omitempty"`  // default value.
	Enums      []*FieldEnum   `json:"enums,omitempty"`    // enum values.
	Relation   *Relation      `json:"relation,omitempty"` // relation of the field.
	DB         *FieldDB       `json:"db,omitempty"`       // db config for the field.

	// Querier
	Sortable   bool `json:"sortable,omitempty"`   // Has a "sort" option in the tag.
	Filterable bool `json:"filterable,omitempty"` // Has a "filter" option in the tag.

	IsSystemField bool `json:"is_system_field,omitempty"` // Is a system field.
}

// Init initializes the field.
// schemaNames is only required for media field.
func (f *Field) Init(schemaNames ...string) {
	if f.DB == nil {
		f.DB = &FieldDB{}
	}

	if f.Type == TypeMedia {
		f.Relation = &Relation{
			Type:             utils.If(f.IsMultiple, M2M, O2M),
			Owner:            false,
			TargetSchemaName: "media",
			TargetFieldName:  fmt.Sprintf("%s_%s", schemaNames[0], f.Name),
			BackRef:          nil,
		}
	}
}

// Clone returns a copy of the field.
func (f *Field) Clone() *Field {
	newField := &Field{
		Type:          f.Type,
		Name:          f.Name,
		Label:         f.Label,
		Renderer:      f.Renderer,
		Size:          f.Size,
		IsMultiple:    f.IsMultiple,
		Unique:        f.Unique,
		Optional:      f.Optional,
		Default:       f.Default,
		Sortable:      f.Sortable,
		Filterable:    f.Filterable,
		IsSystemField: f.IsSystemField,
		Relation:      f.Relation.Clone(),
		DB:            f.DB.Clone(),
	}

	if f.Enums != nil {
		newField.Enums = make([]*FieldEnum, len(f.Enums))
		for i, enum := range f.Enums {
			newField.Enums[i] = enum.Clone()
		}
	}

	return newField
}

// IsValidValue returns true if the value is valid for the column
func (f *Field) IsValidValue(value any) bool {
	if valueArray, ok := value.([]any); ok {
		for _, v := range valueArray {
			if !f.IsValidValue(v) {
				return false
			}
		}
		return true
	}

	if value == nil {
		return true
	}

	switch f.Type {
	case TypeBool:
		_, ok := value.(bool)
		return ok

	case TypeTime:
		return utils.IsValidTime(value)

	case TypeJSON:
		jsonStringValue, ok := value.(string)
		if !ok {
			return false
		}

		_, err := NewEntityFromJSON(jsonStringValue)
		return err == nil

	case TypeUUID, TypeString, TypeText:
		_, ok := value.(string)
		return ok

	case TypeBytes:
		_, ok := value.([]byte)
		return ok

	case TypeEnum:
		enumStringValue, ok := value.(string)
		if !ok {
			return false
		}

		for _, enum := range f.Enums {
			if enum.Value == enumStringValue {
				return true
			}
		}

	case TypeInt, TypeInt8, TypeInt16, TypeInt32, TypeInt64:
		return utils.IsValidInt(value)

	case TypeUint, TypeUint8, TypeUint16, TypeUint32, TypeUint64:
		return utils.IsValidUInt(value)

	// case TypeUintptr:
	// 	return utils.IsValidUInt(value)

	case TypeFloat32, TypeFloat64:
		_, ok := value.(float64)
		return ok
	}

	return false
}

func CreateUint64Field(name string) *Field {
	return &Field{
		Type:  TypeUint64,
		Name:  name,
		Label: name,
		DB: &FieldDB{
			Attr: "UNSIGNED",
		},
		Unique:   false,
		Optional: false,
	}
}

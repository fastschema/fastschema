package schema

import (
	"fmt"
	"strconv"
	"time"

	"github.com/fastschema/fastschema/pkg/utils"
)

// Field define the data struct for a field
type Field struct {
	Type       FieldType `json:"type"`
	Name       string    `json:"name"`
	Label      string    `json:"label"`
	IsMultiple bool      `json:"multiple,omitempty"` // Is a multiple field.
	Size       int64     `json:"size,omitempty"`     // max size parameter for string, blob, etc.
	Unique     bool      `json:"unique,omitempty"`   // column with unique constraint.
	Optional   bool      `json:"optional,omitempty"` // null or not null attribute.
	Default    any       `json:"default,omitempty"`  // default value.
	// Querier
	Sortable   bool           `json:"sortable,omitempty"`   // Has a "sort" option in the tag.
	Filterable bool           `json:"filterable,omitempty"` // Has a "filter" option in the tag.
	Renderer   *FieldRenderer `json:"renderer,omitempty"`   // renderer of the field.
	Enums      []*FieldEnum   `json:"enums,omitempty"`      // enum values.
	Relation   *Relation      `json:"relation,omitempty"`   // relation of the field.
	DB         *FieldDB       `json:"db,omitempty"`         // db config for the field.

	IsSystemField bool `json:"is_system_field,omitempty"` // Is a system field.
}

// Init initializes the field.
// schemaNames is only required for file field.
func (f *Field) Init(schemaNames ...string) {
	if f.DB == nil {
		f.DB = &FieldDB{}
	}

	if f.Type == TypeFile {
		f.Relation = &Relation{
			Type:             utils.If(f.IsMultiple, M2M, O2M),
			Owner:            false,
			TargetSchemaName: "file",
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

// Merge Fields merge second field to the first field.
func MergeFields(f1, f2 *Field) {
	if f2.Type.Valid() {
		f1.Type = f2.Type
	}

	if f2.Name != "" {
		f1.Name = f2.Name
	}

	if f2.Label != "" {
		f1.Label = f2.Label
	}

	if f2.Size > 0 {
		f1.Size = f2.Size
	}

	if f2.Default != nil {
		f1.Default = f2.Default
	}

	if f2.Renderer != nil {
		f1.Renderer = f2.Renderer.Clone()
	}

	if f2.Enums != nil {
		newEnums := make([]*FieldEnum, len(f2.Enums))
		for i, enum := range f2.Enums {
			newEnums[i] = enum.Clone()
		}

		f1.Enums = newEnums
	}

	if f2.Relation != nil {
		f1.Relation = f2.Relation.Clone()
	}

	if f2.DB != nil {
		f1.DB = f2.DB.Clone()
	}

	f1.IsMultiple = f2.IsMultiple
	f1.Unique = f2.Unique
	f1.Optional = f2.Optional
	f1.Sortable = f2.Sortable
	f1.Filterable = f2.Filterable
	f1.IsSystemField = f2.IsSystemField
}

func ErrInvalidFieldValue(fieldName string, value any, errs ...error) error {
	fotmat := "invalid field value: %s=%#v"
	if len(errs) == 0 {
		return fmt.Errorf(fotmat, fieldName, value)
	}

	return fmt.Errorf(
		"invalid field value: %s=%#v - %w",
		fieldName, value, errs[0],
	)
}

// StringToFieldValue converts a string value to a specific type based on the provided field.
func StringToFieldValue[T any](field *Field, strValue string) (T, error) {
	var result T
	var value any
	var err error

	switch field.Type {
	case TypeBool:
		value, err = strconv.ParseBool(strValue)
		if err != nil {
			return result, ErrInvalidFieldValue(field.Name, strValue, err)
		}
	case TypeInt:
		value, err = strconv.Atoi(strValue)
		if err != nil {
			return result, ErrInvalidFieldValue(field.Name, strValue, err)
		}
	case TypeInt8:
		value, err = strconv.ParseInt(strValue, 10, 8)
		if err != nil {
			return result, ErrInvalidFieldValue(field.Name, strValue, err)
		}
		value = int8(value.(int64))
	case TypeInt16:
		value, err = strconv.ParseInt(strValue, 10, 16)
		if err != nil {
			return result, ErrInvalidFieldValue(field.Name, strValue, err)
		}
		value = int16(value.(int64))
	case TypeInt32:
		value, err = strconv.ParseInt(strValue, 10, 32)
		if err != nil {
			return result, ErrInvalidFieldValue(field.Name, strValue, err)
		}
		value = int32(value.(int64))
	case TypeInt64:
		value, err = strconv.ParseInt(strValue, 10, 64)
		if err != nil {
			return result, ErrInvalidFieldValue(field.Name, strValue, err)
		}
	case TypeUint:
		value, err = strconv.ParseUint(strValue, 10, 64)
		if err != nil {
			return result, ErrInvalidFieldValue(field.Name, strValue, err)
		}
		value = uint(value.(uint64))
	case TypeUint8:
		value, err = strconv.ParseUint(strValue, 10, 8)
		if err != nil {
			return result, ErrInvalidFieldValue(field.Name, strValue, err)
		}
		value = uint8(value.(uint64))
	case TypeUint16:
		value, err = strconv.ParseUint(strValue, 10, 16)
		if err != nil {
			return result, ErrInvalidFieldValue(field.Name, strValue, err)
		}
		value = uint16(value.(uint64))
	case TypeUint32:
		value, err = strconv.ParseUint(strValue, 10, 32)
		if err != nil {
			return result, ErrInvalidFieldValue(field.Name, strValue, err)
		}
		value = uint32(value.(uint64))
	case TypeUint64:
		value, err = strconv.ParseUint(strValue, 10, 64)
		if err != nil {
			return result, ErrInvalidFieldValue(field.Name, strValue, err)
		}
	case TypeFloat32:
		value, err = strconv.ParseFloat(strValue, 32)
		if err != nil {
			return result, ErrInvalidFieldValue(field.Name, strValue, err)
		}
		value = float32(value.(float64))
	case TypeFloat64:
		value, err = strconv.ParseFloat(strValue, 64)
		if err != nil {
			return result, ErrInvalidFieldValue(field.Name, strValue, err)
		}
	case TypeTime:
		if strValue == "NOW()" {
			value = "NOW()"
		} else {
			value, err = time.Parse(time.RFC3339, strValue)
			if err != nil {
				return result, ErrInvalidFieldValue(field.Name, strValue, err)
			}
		}
	default:
		value = strValue
	}

	tResult, ok := value.(T)
	if !ok {
		return result, ErrInvalidFieldValue(field.Name, strValue, fmt.Errorf("can't convert %#v to %T", value, result))
	}

	return tResult, nil
}

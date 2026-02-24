package schema

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"maps"
	"reflect"
	"slices"
	"time"
)

// FieldEnum define the data struct for an enum field
type FieldEnum struct {
	Value string `json:"value" yaml:"value"`
	Label string `json:"label" yaml:"label"`
}

// FieldRenderer define the renderer of a field
type FieldRenderer struct {
	Class    string         `json:"class,omitempty" yaml:"class,omitempty"`       // renderer class name
	Settings map[string]any `json:"settings,omitempty" yaml:"settings,omitempty"` // renderer settings.
}

func (fr *FieldRenderer) Clone() *FieldRenderer {
	if fr == nil {
		return nil
	}

	settings := make(map[string]any)
	maps.Copy(settings, fr.Settings)

	return &FieldRenderer{
		Class:    fr.Class,
		Settings: settings,
	}
}

// FieldDB define the db config for a field
type FieldDB struct {
	Attr      string `json:"attr,omitempty" yaml:"attr,omitempty"`           // extra attributes.
	Collation string `json:"collation,omitempty" yaml:"collation,omitempty"` // collation type (utf8mb4_unicode_ci, utf8mb4_general_ci)
	Increment bool   `json:"increment,omitempty" yaml:"increment,omitempty"` // auto increment
	Key       string `json:"key,omitempty" yaml:"key,omitempty"`             // key definition (PRI, UNI or MUL).
}

func (f *FieldEnum) Clone() *FieldEnum {
	return &FieldEnum{
		Value: f.Value,
		Label: f.Label,
	}
}

func (f *FieldDB) Clone() *FieldDB {
	if f == nil {
		return nil
	}

	return &FieldDB{
		Attr:      f.Attr,
		Collation: f.Collation,
		Increment: f.Increment,
		Key:       f.Key,
	}
}

// FieldType define the data type of a field
type FieldType int

const (
	TypeInvalid FieldType = iota
	TypeBool
	TypeTime
	TypeJSON
	TypeUUID
	TypeBytes
	TypeEnum
	TypeString
	TypeText
	TypeInt8
	TypeInt16
	TypeInt32
	TypeInt
	TypeInt64
	TypeUint8
	TypeUint16
	TypeUint32
	TypeUint
	TypeUint64
	TypeFloat32
	TypeFloat64
	TypeRelation
	TypeFile
	endFieldTypes
)

var (
	fieldTypeToStrings = [...]string{
		TypeInvalid: "invalid",
		TypeBool:    "bool",
		TypeTime:    "time",
		TypeJSON:    "json",
		TypeUUID:    "uuid",
		TypeBytes:   "bytes",
		TypeEnum:    "enum",
		TypeString:  "string",
		TypeText:    "text",
		TypeInt:     "int",
		TypeInt8:    "int8",
		TypeInt16:   "int16",
		TypeInt32:   "int32",
		TypeInt64:   "int64",
		TypeUint:    "uint",
		// TypeUintptr:  "uintptr",
		TypeUint8:    "uint8",
		TypeUint16:   "uint16",
		TypeUint32:   "uint32",
		TypeUint64:   "uint64",
		TypeFloat32:  "float32",
		TypeFloat64:  "float64",
		TypeRelation: "relation",
		TypeFile:     "file",
	}

	atomicTypes = []FieldType{
		TypeBool,
		TypeString,
		TypeText,
		TypeInt,
		TypeInt8,
		TypeInt16,
		TypeInt32,
		TypeInt64,
		TypeUint,
		TypeUint8,
		TypeUint16,
		TypeUint32,
		TypeUint64,
		TypeFloat32,
		TypeFloat64,
		TypeTime,
	}

	stringToFieldTypes = map[string]FieldType{
		"invalid": TypeInvalid,
		"bool":    TypeBool,
		"time":    TypeTime,
		"json":    TypeJSON,
		"uuid":    TypeUUID,
		"bytes":   TypeBytes,
		"enum":    TypeEnum,
		"string":  TypeString,
		"text":    TypeText,
		"int":     TypeInt,
		"int8":    TypeInt8,
		"int16":   TypeInt16,
		"int32":   TypeInt32,
		"int64":   TypeInt64,
		"uint":    TypeUint,
		// "uintptr":  TypeUintptr,
		"uint8":    TypeUint8,
		"uint16":   TypeUint16,
		"uint32":   TypeUint32,
		"uint64":   TypeUint64,
		"float32":  TypeFloat32,
		"float64":  TypeFloat64,
		"relation": TypeRelation,
		"file":     TypeFile,
	}

	fieldTypeToStringsToStructTypes = [...]reflect.Type{
		TypeInvalid: nil,
		TypeBool:    reflect.TypeFor[bool](),
		TypeTime:    reflect.TypeFor[time.Time](),
		TypeJSON:    reflect.TypeFor[[]byte](),
		TypeUUID:    reflect.TypeFor[[16]byte](),
		TypeBytes:   reflect.TypeFor[[]byte](),
		TypeEnum:    reflect.TypeFor[FieldEnum](),
		TypeString:  reflect.TypeFor[string](),
		TypeText:    reflect.TypeFor[string](),
		TypeInt:     reflect.TypeFor[int](),
		TypeInt8:    reflect.TypeFor[int8](),
		TypeInt16:   reflect.TypeFor[int16](),
		TypeInt32:   reflect.TypeFor[int32](),
		TypeInt64:   reflect.TypeFor[int64](),
		TypeUint:    reflect.TypeFor[uint](),
		// TypeUintptr:  reflect.TypeFor[uintptr](),
		TypeUint8:    reflect.TypeFor[uint8](),
		TypeUint16:   reflect.TypeFor[uint16](),
		TypeUint32:   reflect.TypeFor[uint32](),
		TypeUint64:   reflect.TypeFor[uint64](),
		TypeFloat32:  reflect.TypeFor[float32](),
		TypeFloat64:  reflect.TypeFor[float64](),
		TypeRelation: reflect.TypeFor[*Relation](),
		TypeFile:     reflect.TypeFor[*Relation](),
	}

	reflectTypesToFieldType = map[reflect.Type]FieldType{
		reflect.TypeFor[bool]():       TypeBool,
		reflect.TypeFor[time.Time]():  TypeTime,
		reflect.TypeFor[*time.Time](): TypeTime,
		reflect.TypeFor[[]byte]():     TypeJSON,
		reflect.TypeFor[[16]byte]():   TypeUUID,
		reflect.TypeFor[FieldEnum]():  TypeEnum,
		reflect.TypeFor[string]():     TypeString,
		reflect.TypeFor[int]():        TypeInt,
		reflect.TypeFor[int8]():       TypeInt8,
		reflect.TypeFor[int16]():      TypeInt16,
		reflect.TypeFor[int32]():      TypeInt32,
		reflect.TypeFor[int64]():      TypeInt64,
		reflect.TypeFor[uint]():       TypeUint,
		reflect.TypeFor[uint8]():      TypeUint8,
		reflect.TypeFor[uint16]():     TypeUint16,
		reflect.TypeFor[uint32]():     TypeUint32,
		reflect.TypeFor[uint64]():     TypeUint64,
		reflect.TypeFor[float32]():    TypeFloat32,
		reflect.TypeFor[float64]():    TypeFloat64,

		reflect.TypeFor[sql.NullString]():  TypeString,
		reflect.TypeFor[sql.NullInt64]():   TypeInt64,
		reflect.TypeFor[sql.NullInt32]():   TypeInt32,
		reflect.TypeFor[sql.NullInt16]():   TypeInt16,
		reflect.TypeFor[sql.NullByte]():    TypeInt8,
		reflect.TypeFor[sql.NullFloat64](): TypeFloat64,
		reflect.TypeFor[sql.NullBool]():    TypeBool,
		reflect.TypeFor[sql.NullTime]():    TypeTime,

		reflect.TypeFor[*sql.NullString]():  TypeString,
		reflect.TypeFor[*sql.NullInt64]():   TypeInt64,
		reflect.TypeFor[*sql.NullInt32]():   TypeInt32,
		reflect.TypeFor[*sql.NullInt16]():   TypeInt16,
		reflect.TypeFor[*sql.NullByte]():    TypeInt8,
		reflect.TypeFor[*sql.NullFloat64](): TypeFloat64,
		reflect.TypeFor[*sql.NullBool]():    TypeBool,
		reflect.TypeFor[*sql.NullTime]():    TypeTime,

		// reflect.TypeFor[*Relation]():  TypeRelation,
		// reflect.TypeFor[*Relation]():  TypeFile,
	}
)

func FieldTypeFromReflectType(t reflect.Type) FieldType {
	if f, ok := reflectTypesToFieldType[t]; ok {
		return f
	}

	return TypeInvalid
}

func FieldTypeFromString(s string) FieldType {
	if f, ok := stringToFieldTypes[s]; ok {
		return f
	}

	return TypeInvalid
}

func (t FieldType) IsRelationType() bool {
	return t == TypeRelation || t == TypeFile
}

// String returns the string representation of a type.
func (t FieldType) String() string {
	if t < endFieldTypes {
		return fieldTypeToStrings[t]
	}
	return fieldTypeToStrings[TypeInvalid]
}

// StructType returns the reflect.Type of the field type
func (t FieldType) StructType() reflect.Type {
	if t < endFieldTypes {
		return fieldTypeToStringsToStructTypes[t]
	}
	return fieldTypeToStringsToStructTypes[TypeInvalid]
}

// IsAtomic reports if the given type is an atomic type.
func (t FieldType) IsAtomic() bool {
	return slices.Contains(atomicTypes, t)
}

// Valid reports if the given type if known type.
func (t FieldType) Valid() bool {
	return t > TypeInvalid && t < endFieldTypes
}

// MarshalJSON marshal an enum value to the quoted json string value
func (t FieldType) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(fieldTypeToStrings[t])
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

// UnmarshalJSON unmashals a quoted json string to the enum value
func (t *FieldType) UnmarshalJSON(b []byte) error {
	var fieldType string
	if err := json.Unmarshal(b, &fieldType); err != nil {
		return err
	}
	*t = stringToFieldTypes[fieldType] // If the string can't be found, it will be set to the zero value: 'invalid'

	if *t == TypeInvalid {
		return fmt.Errorf("invalid field type %q", fieldType)
	}

	return nil
}

// UnmarshalYAML unmashals a YAML string to the enum value
func (t *FieldType) UnmarshalYAML(unmarshal func(any) error) error {
	var fieldType string
	if err := unmarshal(&fieldType); err != nil {
		return err
	}
	*t = stringToFieldTypes[fieldType] // If the string can't be found, it will be set to the zero value: 'invalid'

	if *t == TypeInvalid {
		return fmt.Errorf("invalid field type %q", fieldType)
	}

	return nil
}

// MarshalYAML marshals the enum value to a YAML string
func (t FieldType) MarshalYAML() (any, error) {
	return fieldTypeToStrings[t], nil
}

// RelationType define the relation type of a field
type RelationType int

const (
	RelationInvalid RelationType = iota
	O2O
	O2M
	M2M
	endRelationTypes
)

var (
	relationTypeToStrings = [...]string{
		RelationInvalid: "invalid",
		O2O:             "o2o",
		O2M:             "o2m",
		M2M:             "m2m",
	}

	stringToRelationTypes = map[string]RelationType{
		"invalid": RelationInvalid,
		"o2o":     O2O,
		"o2m":     O2M,
		"m2m":     M2M,
	}
)

func RelationTypeFromString(s string) RelationType {
	if f, ok := stringToRelationTypes[s]; ok {
		return f
	}

	return RelationInvalid
}

func (t RelationType) IsO2O() bool {
	return t == O2O
}

func (t RelationType) IsO2M() bool {
	return t == O2M
}

func (t RelationType) IsM2M() bool {
	return t == M2M
}

// String returns the string representation of a type.
func (t RelationType) String() string {
	if t < endRelationTypes {
		return relationTypeToStrings[t]
	}
	return relationTypeToStrings[RelationInvalid]
}

// Valid reports if the given type if known type.
func (t RelationType) Valid() bool {
	return t > RelationInvalid && t < endRelationTypes
}

// MarshalJSON marshal an enum value to the quoted json string value
func (t RelationType) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(relationTypeToStrings[t])
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

// UnmarshalJSON unmashals a quoted json string to the enum value
func (t *RelationType) UnmarshalJSON(b []byte) error {
	var j string
	if err := json.Unmarshal(b, &j); err != nil {
		return err
	}
	*t = stringToRelationTypes[j] // If the string can't be found, it will be set to the zero value: 'invalid'
	return nil
}

// UnmarshalYAML unmarshal a YAML string to the enum value
func (t *RelationType) UnmarshalYAML(unmarshal func(any) error) error {
	var relationType string
	if err := unmarshal(&relationType); err != nil {
		return err
	}
	*t = stringToRelationTypes[relationType] // If the string can't be found, it will be set to the zero value: 'invalid'
	return nil
}

// MarshalYAML marshals the enum value to a YAML string
func (t RelationType) MarshalYAML() (any, error) {
	return relationTypeToStrings[t], nil
}

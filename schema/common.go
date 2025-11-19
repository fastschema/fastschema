package schema

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"reflect"
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
	for k, v := range fr.Settings {
		settings[k] = v
	}

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
		TypeBool:    reflect.TypeOf(bool(false)),
		TypeTime:    reflect.TypeOf(time.Time{}),
		TypeJSON:    reflect.TypeOf([]byte{}),
		TypeUUID:    reflect.TypeOf([16]byte{}),
		TypeBytes:   reflect.TypeOf([]byte{}),
		TypeEnum:    reflect.TypeOf(FieldEnum{}),
		TypeString:  reflect.TypeOf(string("")),
		TypeText:    reflect.TypeOf(string("")),
		TypeInt:     reflect.TypeOf(int(0)),
		TypeInt8:    reflect.TypeOf(int8(0)),
		TypeInt16:   reflect.TypeOf(int16(0)),
		TypeInt32:   reflect.TypeOf(int32(0)),
		TypeInt64:   reflect.TypeOf(int64(0)),
		TypeUint:    reflect.TypeOf(uint(0)),
		// TypeUintptr:  reflect.TypeOf(uintptr(0)),
		TypeUint8:    reflect.TypeOf(uint8(0)),
		TypeUint16:   reflect.TypeOf(uint16(0)),
		TypeUint32:   reflect.TypeOf(uint32(0)),
		TypeUint64:   reflect.TypeOf(uint64(0)),
		TypeFloat32:  reflect.TypeOf(float32(0)),
		TypeFloat64:  reflect.TypeOf(float64(0)),
		TypeRelation: reflect.TypeOf(&Relation{}),
		TypeFile:     reflect.TypeOf(&Relation{}),
	}

	reflectTypesToFieldType = map[reflect.Type]FieldType{
		reflect.TypeOf(bool(false)):  TypeBool,
		reflect.TypeOf(time.Time{}):  TypeTime,
		reflect.TypeOf(&time.Time{}): TypeTime,
		reflect.TypeOf([]byte{}):     TypeJSON,
		reflect.TypeOf([16]byte{}):   TypeUUID,
		reflect.TypeOf([]byte{}):     TypeBytes,
		reflect.TypeOf(FieldEnum{}):  TypeEnum,
		reflect.TypeOf(string("")):   TypeString,
		reflect.TypeOf(int(0)):       TypeInt,
		reflect.TypeOf(int8(0)):      TypeInt8,
		reflect.TypeOf(int16(0)):     TypeInt16,
		reflect.TypeOf(int32(0)):     TypeInt32,
		reflect.TypeOf(int64(0)):     TypeInt64,
		reflect.TypeOf(uint(0)):      TypeUint,
		reflect.TypeOf(uint8(0)):     TypeUint8,
		reflect.TypeOf(uint16(0)):    TypeUint16,
		reflect.TypeOf(uint32(0)):    TypeUint32,
		reflect.TypeOf(uint64(0)):    TypeUint64,
		reflect.TypeOf(float32(0)):   TypeFloat32,
		reflect.TypeOf(float64(0)):   TypeFloat64,

		reflect.TypeOf(sql.NullString{}):  TypeString,
		reflect.TypeOf(sql.NullInt64{}):   TypeInt64,
		reflect.TypeOf(sql.NullInt32{}):   TypeInt32,
		reflect.TypeOf(sql.NullInt16{}):   TypeInt16,
		reflect.TypeOf(sql.NullByte{}):    TypeInt8,
		reflect.TypeOf(sql.NullFloat64{}): TypeFloat64,
		reflect.TypeOf(sql.NullBool{}):    TypeBool,
		reflect.TypeOf(sql.NullTime{}):    TypeTime,

		reflect.TypeOf(&sql.NullString{}):  TypeString,
		reflect.TypeOf(&sql.NullInt64{}):   TypeInt64,
		reflect.TypeOf(&sql.NullInt32{}):   TypeInt32,
		reflect.TypeOf(&sql.NullInt16{}):   TypeInt16,
		reflect.TypeOf(&sql.NullByte{}):    TypeInt8,
		reflect.TypeOf(&sql.NullFloat64{}): TypeFloat64,
		reflect.TypeOf(&sql.NullBool{}):    TypeBool,
		reflect.TypeOf(&sql.NullTime{}):    TypeTime,

		// reflect.TypeOf(&Relation{}):  TypeRelation,
		// reflect.TypeOf(&Relation{}):  TypeFile,
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
	for _, pt := range atomicTypes {
		if t == pt {
			return true
		}
	}
	return false
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
func (t *FieldType) UnmarshalYAML(unmarshal func(interface{}) error) error {
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
func (t FieldType) MarshalYAML() (interface{}, error) {
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

// UnmarshalYAML unmashals a YAML string to the enum value
func (t *RelationType) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var relationType string
	if err := unmarshal(&relationType); err != nil {
		return err
	}
	*t = stringToRelationTypes[relationType] // If the string can't be found, it will be set to the zero value: 'invalid'
	return nil
}

// MarshalYAML marshals the enum value to a YAML string
func (t RelationType) MarshalYAML() (interface{}, error) {
	return relationTypeToStrings[t], nil
}

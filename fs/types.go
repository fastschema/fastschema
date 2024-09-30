package fs

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// Map is a shortcut for map[string]any
type Map = map[string]any

func MapValue[V any](m Map, key string, defaultValues ...V) V {
	if v, ok := m[key]; ok {
		value, ok := v.(V)
		if ok {
			return value
		}
	}

	var defaultValue V
	if len(defaultValues) > 0 {
		defaultValue = defaultValues[0]
	}

	return defaultValue
}

var SystemSchemaTypes = []any{
	Role{},
	Permission{},
	User{},
	File{},
}

type Arg struct {
	Type        ArgType `json:"type"`
	Required    bool    `json:"required"`
	Description string  `json:"description"`
	Example     any     `json:"example"`
}

type Args map[string]Arg

// Meta hold extra data, ex: request method, path, etc
type Meta struct {
	// Http method empty means the method is not set/allowed
	Get     string `json:"get,omitempty"`     // Only use for restful method GET
	Head    string `json:"head,omitempty"`    // Only use for restful method HEAD
	Post    string `json:"post,omitempty"`    // Only use for restful method POST
	Put     string `json:"put,omitempty"`     // Only use for restful method PUT
	Delete  string `json:"delete,omitempty"`  // Only use for restful method DELETE
	Connect string `json:"connect,omitempty"` // Only use for restful method CONNECT
	Options string `json:"options,omitempty"` // Only use for restful method OPTIONS
	Trace   string `json:"trace,omitempty"`   // Only use for restful method TRACE
	Patch   string `json:"patch,omitempty"`   // Only use for restful method PATCH

	// WS
	WS string `json:"ws,omitempty"` // Only use for websocket

	Prefix     string     `json:"prefix,omitempty"` // Only use for group resource
	Args       Args       `json:"args,omitempty"`
	Public     bool       `json:"public,omitempty"`
	Signatures Signatures `json:"-"`
}

// Signature hold the information of a signature
type Signature struct {
	Type any
	Name string
}

// Signatures hold the input and output types of a resolver
//
//   - The first element is the input type
//
//   - The second element is the output type
//
//   - Each element is a type of any or a *Signature
//
// For example:
//
// - []any{&LoginData{}, &LoginResponse{}} // The input type is *LoginData and the output type is *LoginResponse
//
// - []any{int, string} // The input type is int and the output type is string
//
// - []any{&Signature{Type: dynamicStruct, Name: "Dynamic"}, int}: dynamic struct doesn't have name, use *Signature to define it
type Signatures = []any

func (a Args) Clone() Args {
	args := make(Args, len(a))
	for k, v := range a {
		args[k] = v
	}
	return args
}

func (m *Meta) Clone() *Meta {
	return &Meta{
		Get:     m.Get,
		Head:    m.Head,
		Post:    m.Post,
		Put:     m.Put,
		Delete:  m.Delete,
		Connect: m.Connect,
		Options: m.Options,
		Trace:   m.Trace,
		Patch:   m.Patch,

		WS: m.WS,

		Prefix: m.Prefix,
		Args:   m.Args.Clone(),
		Public: m.Public,
	}
}

// ArgType define the data type of a field
type ArgType int

const (
	TypeInvalid ArgType = iota
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
	endArgTypes
)

var (
	argTypeToStrings = [...]string{
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
		TypeUint8:   "uint8",
		TypeUint16:  "uint16",
		TypeUint32:  "uint32",
		TypeUint64:  "uint64",
		TypeFloat32: "float32",
		TypeFloat64: "float64",
	}

	stringToArgTypes = map[string]ArgType{
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
		"uint8":   TypeUint8,
		"uint16":  TypeUint16,
		"uint32":  TypeUint32,
		"uint64":  TypeUint64,
		"float32": TypeFloat32,
		"float64": TypeFloat64,
	}

	argTypeToOpenCommonType = map[ArgType]string{
		TypeInvalid: "invalid",
		TypeBool:    "boolean",
		TypeTime:    "string",
		TypeJSON:    "object",
		TypeUUID:    "string",
		TypeBytes:   "string",
		TypeEnum:    "string",
		TypeString:  "string",
		TypeText:    "string",
		TypeInt:     "integer",
		TypeInt8:    "integer",
		TypeInt16:   "integer",
		TypeInt32:   "integer",
		TypeInt64:   "integer",
		TypeUint:    "integer",
		TypeUint8:   "integer",
		TypeUint16:  "integer",
		TypeUint32:  "integer",
		TypeUint64:  "integer",
		TypeFloat32: "number",
		TypeFloat64: "number",
	}
)

// Common returns the common type of a field
func (t ArgType) Common() string {
	if t < endArgTypes {
		return argTypeToOpenCommonType[t]
	}
	return argTypeToOpenCommonType[TypeInvalid]
}

// String returns the string representation of a type.
func (t ArgType) String() string {
	if t < endArgTypes {
		return argTypeToStrings[t]
	}
	return argTypeToStrings[TypeInvalid]
}

// Valid reports if the given type if known type.
func (t ArgType) Valid() bool {
	return t > TypeInvalid && t < endArgTypes
}

// MarshalJSON marshal an enum value to the quoted json string value
func (t ArgType) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(argTypeToStrings[t])
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

// UnmarshalJSON unmashals a quoted json string to the enum value
func (t *ArgType) UnmarshalJSON(b []byte) error {
	var fieldType string
	if err := json.Unmarshal(b, &fieldType); err != nil {
		return err
	}
	*t = stringToArgTypes[fieldType] // If the string can't be found, it will be set to the zero value: 'invalid'

	if *t == TypeInvalid {
		return fmt.Errorf("invalid arg type %q", fieldType)
	}

	return nil
}

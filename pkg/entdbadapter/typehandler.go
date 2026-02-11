package entdbadapter

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/fastschema/fastschema/entity"
	"github.com/fastschema/fastschema/schema"
	"github.com/google/uuid"
)

// TypeHandler defines scan/assign pair for a field type.
// ScanValue returns a pointer to scan into.
// AssignValue converts the scanned value to the target type.
type TypeHandler struct {
	ScanValue   func() any
	AssignValue func(column string, value any, entity *entity.Entity) (any, error)
}

// typeHandlers maps field types to their handlers.
// This centralizes all type conversion logic in one place.
var typeHandlers = map[schema.FieldType]TypeHandler{
	schema.TypeBool:    {ScanValue: scanBool, AssignValue: assignBool},
	schema.TypeTime:    {ScanValue: scanTime, AssignValue: assignTime},
	schema.TypeJSON:    {ScanValue: scanBytes, AssignValue: assignJSON},
	schema.TypeUUID:    {ScanValue: scanUUID, AssignValue: assignUUID},
	schema.TypeBytes:   {ScanValue: scanBytes, AssignValue: assignBytes},
	schema.TypeEnum:    {ScanValue: scanString, AssignValue: assignString},
	schema.TypeString:  {ScanValue: scanString, AssignValue: assignString},
	schema.TypeText:    {ScanValue: scanString, AssignValue: assignString},
	schema.TypeInt8:    {ScanValue: scanInt64, AssignValue: assignInt8},
	schema.TypeInt16:   {ScanValue: scanInt64, AssignValue: assignInt16},
	schema.TypeInt32:   {ScanValue: scanInt64, AssignValue: assignInt32},
	schema.TypeInt:     {ScanValue: scanInt64, AssignValue: assignInt},
	schema.TypeInt64:   {ScanValue: scanInt64, AssignValue: assignInt64Value},
	schema.TypeUint8:   {ScanValue: scanInt64, AssignValue: assignUint8},
	schema.TypeUint16:  {ScanValue: scanInt64, AssignValue: assignUint16},
	schema.TypeUint32:  {ScanValue: scanInt64, AssignValue: assignUint32},
	schema.TypeUint:    {ScanValue: scanInt64, AssignValue: assignUint},
	schema.TypeUint64:  {ScanValue: scanInt64, AssignValue: assignUint64},
	schema.TypeFloat32: {ScanValue: scanFloat64, AssignValue: assignFloat32},
	schema.TypeFloat64: {ScanValue: scanFloat64, AssignValue: assignFloat64Value},
}

// GetTypeHandler returns the type handler for the given field type.
// Returns a default handler if the field type is not found.
func GetTypeHandler(fieldType schema.FieldType) TypeHandler {
	if handler, ok := typeHandlers[fieldType]; ok {
		return handler
	}
	return TypeHandler{ScanValue: scanAny, AssignValue: assignAny}
}

// =============================================================================
// Scan Value Functions
// =============================================================================

func scanBool() any    { return new(sql.NullBool) }
func scanTime() any    { return new(sql.NullTime) }
func scanBytes() any   { return new([]byte) }
func scanUUID() any    { return new(uuid.UUID) }
func scanString() any  { return new(sql.NullString) }
func scanInt64() any   { return new(sql.NullInt64) }
func scanFloat64() any { return new(sql.NullFloat64) }
func scanAny() any     { return new(any) }

// =============================================================================
// Assign Value Functions
// =============================================================================

func assignBool(_ string, value any, _ *entity.Entity) (any, error) {
	v, ok := value.(*sql.NullBool)
	if !ok {
		return nil, fieldTypeError("*sql.NullBool", value)
	}
	if v.Valid {
		return v.Bool, nil
	}
	return nil, nil
}

func assignTime(_ string, value any, _ *entity.Entity) (any, error) {
	v, ok := value.(*sql.NullTime)
	if !ok {
		return nil, fieldTypeError("*sql.NullTime", value)
	}
	if v.Valid {
		return v.Time, nil
	}
	return nil, nil
}

func assignJSON(column string, value any, e *entity.Entity) (any, error) {
	v, ok := value.(*[]byte)
	if !ok {
		return nil, fieldTypeError("*[]byte", value)
	}
	if v != nil && len(*v) > 0 {
		existing := e.Get(column)
		if err := json.Unmarshal(*v, &existing); err != nil {
			return nil, fmt.Errorf("unmarshal field field_type_JSON: %w", err)
		}
		return existing, nil
	}
	return nil, nil
}

func assignUUID(_ string, value any, _ *entity.Entity) (any, error) {
	v, ok := value.(*uuid.UUID)
	if !ok {
		return nil, fieldTypeError("*uuid.UUID", value)
	}
	if v != nil {
		return *v, nil
	}
	return nil, nil
}

func assignBytes(_ string, value any, _ *entity.Entity) (any, error) {
	v, ok := value.(*[]byte)
	if !ok {
		return nil, fieldTypeError("*[]byte", value)
	}
	if v != nil {
		return *v, nil
	}
	return nil, nil
}

func assignString(_ string, value any, _ *entity.Entity) (any, error) {
	v, ok := value.(*sql.NullString)
	if !ok {
		return nil, fieldTypeError("*sql.NullString", value)
	}
	if v.Valid {
		return v.String, nil
	}
	return nil, nil
}

// =============================================================================
// Integer Assignment Functions
// =============================================================================

func assignInt8(_ string, value any, _ *entity.Entity) (any, error) {
	v, ok := value.(*sql.NullInt64)
	if !ok {
		return nil, fieldTypeError("*sql.NullInt64", value)
	}
	if v.Valid {
		return int8(v.Int64), nil
	}
	return nil, nil
}

func assignInt16(_ string, value any, _ *entity.Entity) (any, error) {
	v, ok := value.(*sql.NullInt64)
	if !ok {
		return nil, fieldTypeError("*sql.NullInt64", value)
	}
	if v.Valid {
		return int16(v.Int64), nil
	}
	return nil, nil
}

func assignInt32(_ string, value any, _ *entity.Entity) (any, error) {
	v, ok := value.(*sql.NullInt64)
	if !ok {
		return nil, fieldTypeError("*sql.NullInt64", value)
	}
	if v.Valid {
		return int32(v.Int64), nil
	}
	return nil, nil
}

func assignInt(_ string, value any, _ *entity.Entity) (any, error) {
	v, ok := value.(*sql.NullInt64)
	if !ok {
		return nil, fieldTypeError("*sql.NullInt64", value)
	}
	if v.Valid {
		return int(v.Int64), nil
	}
	return nil, nil
}

func assignInt64Value(_ string, value any, _ *entity.Entity) (any, error) {
	v, ok := value.(*sql.NullInt64)
	if !ok {
		return nil, fieldTypeError("*sql.NullInt64", value)
	}
	if v.Valid {
		return v.Int64, nil
	}
	return nil, nil
}

// =============================================================================
// Unsigned Integer Assignment Functions
// =============================================================================

func assignUint8(_ string, value any, _ *entity.Entity) (any, error) {
	v, ok := value.(*sql.NullInt64)
	if !ok {
		return nil, fieldTypeError("*sql.NullInt64", value)
	}
	if v.Valid {
		return uint8(v.Int64), nil
	}
	return nil, nil
}

func assignUint16(_ string, value any, _ *entity.Entity) (any, error) {
	v, ok := value.(*sql.NullInt64)
	if !ok {
		return nil, fieldTypeError("*sql.NullInt64", value)
	}
	if v.Valid {
		return uint16(v.Int64), nil
	}
	return nil, nil
}

func assignUint32(_ string, value any, _ *entity.Entity) (any, error) {
	v, ok := value.(*sql.NullInt64)
	if !ok {
		return nil, fieldTypeError("*sql.NullInt64", value)
	}
	if v.Valid {
		return uint32(v.Int64), nil
	}
	return nil, nil
}

func assignUint(_ string, value any, _ *entity.Entity) (any, error) {
	v, ok := value.(*sql.NullInt64)
	if !ok {
		return nil, fieldTypeError("*sql.NullInt64", value)
	}
	// Note: returns 0 for invalid (null) values to maintain backward compatibility
	return uint(v.Int64), nil
}

func assignUint64(_ string, value any, _ *entity.Entity) (any, error) {
	v, ok := value.(*sql.NullInt64)
	if !ok {
		return nil, fieldTypeError("*sql.NullInt64", value)
	}
	if v.Valid {
		return uint64(v.Int64), nil
	}
	return nil, nil
}

// =============================================================================
// Float Assignment Functions
// =============================================================================

func assignFloat32(_ string, value any, _ *entity.Entity) (any, error) {
	v, ok := value.(*sql.NullFloat64)
	if !ok {
		return nil, fieldTypeError("*sql.NullFloat64", value)
	}
	if v.Valid {
		return float32(v.Float64), nil
	}
	return nil, nil
}

func assignFloat64Value(_ string, value any, _ *entity.Entity) (any, error) {
	v, ok := value.(*sql.NullFloat64)
	if !ok {
		return nil, fieldTypeError("*sql.NullFloat64", value)
	}
	if v.Valid {
		return v.Float64, nil
	}
	return nil, nil
}

// =============================================================================
// Default Assignment Function
// =============================================================================

func assignAny(_ string, value any, _ *entity.Entity) (any, error) {
	v, ok := value.(*any)
	if !ok {
		return nil, fieldTypeError("*any", value)
	}
	if v != nil {
		return *v, nil
	}
	return nil, nil
}

// =============================================================================
// Helper Functions for Integer Type Grouping
// =============================================================================

// IsIntegerType returns true if the field type is any integer type.
func IsIntegerType(fieldType schema.FieldType) bool {
	switch fieldType {
	case schema.TypeInt8, schema.TypeInt16, schema.TypeInt32, schema.TypeInt, schema.TypeInt64,
		schema.TypeUint8, schema.TypeUint16, schema.TypeUint32, schema.TypeUint, schema.TypeUint64:
		return true
	default:
		return false
	}
}

// IsStringType returns true if the field type is a string-like type.
func IsStringType(fieldType schema.FieldType) bool {
	switch fieldType {
	case schema.TypeString, schema.TypeText, schema.TypeEnum:
		return true
	default:
		return false
	}
}

// IsFloatType returns true if the field type is a float type.
func IsFloatType(fieldType schema.FieldType) bool {
	return fieldType == schema.TypeFloat32 || fieldType == schema.TypeFloat64
}

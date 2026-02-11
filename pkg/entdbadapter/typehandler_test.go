package entdbadapter

import (
	"database/sql"
	"testing"
	"time"

	"github.com/fastschema/fastschema/entity"
	"github.com/fastschema/fastschema/schema"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestGetTypeHandler(t *testing.T) {
	tests := []struct {
		name      string
		fieldType schema.FieldType
		wantType  string // expected scan value type
	}{
		{"Bool", schema.TypeBool, "*sql.NullBool"},
		{"Time", schema.TypeTime, "*sql.NullTime"},
		{"JSON", schema.TypeJSON, "*[]uint8"},
		{"UUID", schema.TypeUUID, "*uuid.UUID"},
		{"Bytes", schema.TypeBytes, "*[]uint8"},
		{"Enum", schema.TypeEnum, "*sql.NullString"},
		{"String", schema.TypeString, "*sql.NullString"},
		{"Text", schema.TypeText, "*sql.NullString"},
		{"Int8", schema.TypeInt8, "*sql.NullInt64"},
		{"Int16", schema.TypeInt16, "*sql.NullInt64"},
		{"Int32", schema.TypeInt32, "*sql.NullInt64"},
		{"Int", schema.TypeInt, "*sql.NullInt64"},
		{"Int64", schema.TypeInt64, "*sql.NullInt64"},
		{"Uint8", schema.TypeUint8, "*sql.NullInt64"},
		{"Uint16", schema.TypeUint16, "*sql.NullInt64"},
		{"Uint32", schema.TypeUint32, "*sql.NullInt64"},
		{"Uint", schema.TypeUint, "*sql.NullInt64"},
		{"Uint64", schema.TypeUint64, "*sql.NullInt64"},
		{"Float32", schema.TypeFloat32, "*sql.NullFloat64"},
		{"Float64", schema.TypeFloat64, "*sql.NullFloat64"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := GetTypeHandler(tt.fieldType)
			assert.NotNil(t, handler.ScanValue)
			assert.NotNil(t, handler.AssignValue)
		})
	}
}

func TestGetTypeHandlerDefault(t *testing.T) {
	// Test with a type that doesn't have a specific handler (uses default)
	// The default handler uses scanAny and assignAny
	handler := GetTypeHandler(schema.TypeInvalid)
	assert.NotNil(t, handler.ScanValue)
	assert.NotNil(t, handler.AssignValue)
}

func TestTypeHandlerAssignValues(t *testing.T) {
	e := entity.New()

	tests := []struct {
		name      string
		fieldType schema.FieldType
		scanValue any
		expected  any
	}{
		{
			name:      "Bool true",
			fieldType: schema.TypeBool,
			scanValue: &sql.NullBool{Valid: true, Bool: true},
			expected:  true,
		},
		{
			name:      "Bool null",
			fieldType: schema.TypeBool,
			scanValue: &sql.NullBool{Valid: false},
			expected:  nil,
		},
		{
			name:      "String valid",
			fieldType: schema.TypeString,
			scanValue: &sql.NullString{Valid: true, String: "test"},
			expected:  "test",
		},
		{
			name:      "String null",
			fieldType: schema.TypeString,
			scanValue: &sql.NullString{Valid: false},
			expected:  nil,
		},
		{
			name:      "Int8 valid",
			fieldType: schema.TypeInt8,
			scanValue: &sql.NullInt64{Valid: true, Int64: 127},
			expected:  int8(127),
		},
		{
			name:      "Int16 valid",
			fieldType: schema.TypeInt16,
			scanValue: &sql.NullInt64{Valid: true, Int64: 32767},
			expected:  int16(32767),
		},
		{
			name:      "Int32 valid",
			fieldType: schema.TypeInt32,
			scanValue: &sql.NullInt64{Valid: true, Int64: 2147483647},
			expected:  int32(2147483647),
		},
		{
			name:      "Int valid",
			fieldType: schema.TypeInt,
			scanValue: &sql.NullInt64{Valid: true, Int64: 42},
			expected:  int(42),
		},
		{
			name:      "Int64 valid",
			fieldType: schema.TypeInt64,
			scanValue: &sql.NullInt64{Valid: true, Int64: 9223372036854775807},
			expected:  int64(9223372036854775807),
		},
		{
			name:      "Uint8 valid",
			fieldType: schema.TypeUint8,
			scanValue: &sql.NullInt64{Valid: true, Int64: 255},
			expected:  uint8(255),
		},
		{
			name:      "Uint16 valid",
			fieldType: schema.TypeUint16,
			scanValue: &sql.NullInt64{Valid: true, Int64: 65535},
			expected:  uint16(65535),
		},
		{
			name:      "Uint32 valid",
			fieldType: schema.TypeUint32,
			scanValue: &sql.NullInt64{Valid: true, Int64: 4294967295},
			expected:  uint32(4294967295),
		},
		{
			name:      "Uint valid",
			fieldType: schema.TypeUint,
			scanValue: &sql.NullInt64{Valid: true, Int64: 42},
			expected:  uint(42),
		},
		{
			name:      "Uint null returns zero",
			fieldType: schema.TypeUint,
			scanValue: &sql.NullInt64{Valid: false},
			expected:  uint(0),
		},
		{
			name:      "Uint64 valid",
			fieldType: schema.TypeUint64,
			scanValue: &sql.NullInt64{Valid: true, Int64: 42},
			expected:  uint64(42),
		},
		{
			name:      "Float32 valid",
			fieldType: schema.TypeFloat32,
			scanValue: &sql.NullFloat64{Valid: true, Float64: 3.14},
			expected:  float32(3.14),
		},
		{
			name:      "Float64 valid",
			fieldType: schema.TypeFloat64,
			scanValue: &sql.NullFloat64{Valid: true, Float64: 3.14159265359},
			expected:  float64(3.14159265359),
		},
		{
			name:      "Time valid",
			fieldType: schema.TypeTime,
			scanValue: &sql.NullTime{Valid: true, Time: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)},
			expected:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:      "Time null",
			fieldType: schema.TypeTime,
			scanValue: &sql.NullTime{Valid: false},
			expected:  nil,
		},
		{
			name:      "Bytes valid",
			fieldType: schema.TypeBytes,
			scanValue: &[]byte{0x01, 0x02, 0x03},
			expected:  []byte{0x01, 0x02, 0x03},
		},
		{
			name:      "Bytes nil",
			fieldType: schema.TypeBytes,
			scanValue: (*[]byte)(nil),
			expected:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := GetTypeHandler(tt.fieldType)
			result, err := handler.AssignValue("test_col", tt.scanValue, e)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTypeHandlerUUID(t *testing.T) {
	e := entity.New()

	t.Run("UUID valid", func(t *testing.T) {
		testUUID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
		handler := GetTypeHandler(schema.TypeUUID)
		result, err := handler.AssignValue("uuid_col", &testUUID, e)
		assert.NoError(t, err)
		assert.Equal(t, testUUID, result)
	})

	t.Run("UUID nil", func(t *testing.T) {
		handler := GetTypeHandler(schema.TypeUUID)
		result, err := handler.AssignValue("uuid_col", (*uuid.UUID)(nil), e)
		assert.NoError(t, err)
		assert.Nil(t, result)
	})
}

func TestTypeHandlerJSON(t *testing.T) {
	t.Run("JSON valid", func(t *testing.T) {
		e := entity.New()
		e.Set("json_col", map[string]any{}) // Initialize with expected structure
		jsonBytes := []byte(`{"key": "value"}`)
		handler := GetTypeHandler(schema.TypeJSON)
		result, err := handler.AssignValue("json_col", &jsonBytes, e)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		// Check that the JSON was unmarshaled correctly
		resultMap, ok := result.(map[string]any)
		assert.True(t, ok)
		assert.Equal(t, "value", resultMap["key"])
	})

	t.Run("JSON empty", func(t *testing.T) {
		e := entity.New()
		emptyBytes := []byte{}
		handler := GetTypeHandler(schema.TypeJSON)
		result, err := handler.AssignValue("json_col", &emptyBytes, e)
		assert.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("JSON invalid", func(t *testing.T) {
		e := entity.New()
		e.Set("json_col", map[string]any{})
		invalidJSON := []byte(`{invalid}`)
		handler := GetTypeHandler(schema.TypeJSON)
		_, err := handler.AssignValue("json_col", &invalidJSON, e)
		assert.Error(t, err)
	})
}

func TestTypeHandlerErrors(t *testing.T) {
	e := entity.New()

	tests := []struct {
		name       string
		fieldType  schema.FieldType
		wrongValue any
	}{
		{"Bool wrong type", schema.TypeBool, "not a bool"},
		{"Time wrong type", schema.TypeTime, "not a time"},
		{"String wrong type", schema.TypeString, 123},
		{"Int64 wrong type", schema.TypeInt64, "not an int"},
		{"Float64 wrong type", schema.TypeFloat64, "not a float"},
		{"UUID wrong type", schema.TypeUUID, "not a uuid"},
		{"Bytes wrong type", schema.TypeBytes, 123},
		{"JSON wrong type", schema.TypeJSON, 123},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := GetTypeHandler(tt.fieldType)
			_, err := handler.AssignValue("test_col", tt.wrongValue, e)
			assert.Error(t, err)
		})
	}
}

func TestIsIntegerType(t *testing.T) {
	integerTypes := []schema.FieldType{
		schema.TypeInt8, schema.TypeInt16, schema.TypeInt32, schema.TypeInt, schema.TypeInt64,
		schema.TypeUint8, schema.TypeUint16, schema.TypeUint32, schema.TypeUint, schema.TypeUint64,
	}

	for _, ft := range integerTypes {
		t.Run(ft.String(), func(t *testing.T) {
			assert.True(t, IsIntegerType(ft))
		})
	}

	nonIntegerTypes := []schema.FieldType{
		schema.TypeString, schema.TypeBool, schema.TypeFloat32, schema.TypeFloat64,
		schema.TypeTime, schema.TypeJSON, schema.TypeUUID, schema.TypeBytes,
	}

	for _, ft := range nonIntegerTypes {
		t.Run(ft.String()+"_not_integer", func(t *testing.T) {
			assert.False(t, IsIntegerType(ft))
		})
	}
}

func TestIsStringType(t *testing.T) {
	stringTypes := []schema.FieldType{
		schema.TypeString, schema.TypeText, schema.TypeEnum,
	}

	for _, ft := range stringTypes {
		t.Run(ft.String(), func(t *testing.T) {
			assert.True(t, IsStringType(ft))
		})
	}

	nonStringTypes := []schema.FieldType{
		schema.TypeInt, schema.TypeBool, schema.TypeFloat64, schema.TypeTime,
	}

	for _, ft := range nonStringTypes {
		t.Run(ft.String()+"_not_string", func(t *testing.T) {
			assert.False(t, IsStringType(ft))
		})
	}
}

func TestIsFloatType(t *testing.T) {
	assert.True(t, IsFloatType(schema.TypeFloat32))
	assert.True(t, IsFloatType(schema.TypeFloat64))
	assert.False(t, IsFloatType(schema.TypeInt))
	assert.False(t, IsFloatType(schema.TypeString))
}

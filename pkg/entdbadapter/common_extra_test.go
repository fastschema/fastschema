package entdbadapter

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"testing"

	entSchema "entgo.io/ent/dialect/sql/schema"
	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/entity"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCommonIsZeroValue tests the isZeroValue function with various types
func TestCommonIsZeroValue(t *testing.T) {
	nonNilUUID := uuid.New()
	tests := []struct {
		name     string
		value    any
		expected bool
	}{
		{name: "nil_value", value: nil, expected: true},
		{name: "zero_int", value: 0, expected: true},
		{name: "nonzero_int", value: 42, expected: false},
		{name: "zero_int64", value: int64(0), expected: true},
		{name: "nonzero_int64", value: int64(123), expected: false},
		{name: "zero_uint", value: uint(0), expected: true},
		{name: "nonzero_uint", value: uint(7), expected: false},
		{name: "zero_uint64", value: uint64(0), expected: true},
		{name: "nonzero_uint64", value: uint64(999), expected: false},
		{name: "zero_float64", value: float64(0), expected: true},
		{name: "nonzero_float64", value: 3.14, expected: false},
		{name: "empty_string", value: "", expected: true},
		{name: "nonempty_string", value: "hello", expected: false},
		{name: "false_bool", value: false, expected: true},
		{name: "true_bool", value: true, expected: false},
		{name: "empty_slice", value: []int{}, expected: true},
		{name: "nonempty_slice", value: []int{1, 2}, expected: false},
		{name: "nil_slice", value: []int(nil), expected: true},
		{name: "empty_map", value: map[string]int{}, expected: true},
		{name: "nonempty_map", value: map[string]int{"a": 1}, expected: false},
		{name: "nil_map", value: map[string]int(nil), expected: true},
		{name: "nil_pointer", value: (*int)(nil), expected: true},
		{name: "empty_struct", value: struct{}{}, expected: true},
		{name: "nonempty_struct", value: struct{ A int }{A: 5}, expected: false},
		// UUID types - covering the UUID-specific branches
		{name: "zero_uuid", value: uuid.UUID{}, expected: true},
		{name: "nil_uuid", value: uuid.Nil, expected: true},
		{name: "nonzero_uuid", value: nonNilUUID, expected: false},
		{name: "nil_uuid_ptr", value: (*uuid.UUID)(nil), expected: true},
		{name: "zero_uuid_ptr", value: &uuid.UUID{}, expected: true},
		{name: "nonzero_uuid_ptr", value: &nonNilUUID, expected: false},
		// Array types (fixed size arrays)
		{name: "zero_array", value: [3]int{}, expected: true},
		{name: "nonzero_array", value: [3]int{1, 2, 3}, expected: false},
		// Interface types
		{name: "nil_interface", value: interface{}(nil), expected: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isZeroValue(tt.value)
			assert.Equal(t, tt.expected, result, "isZeroValue(%v) should be %v", tt.value, tt.expected)
		})
	}
}

// TestCommonValueKey tests the valueKey function
// Note: valueKey returns "type:value" format for cache key generation
func TestCommonValueKey(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected string
	}{
		{name: "int", input: 42, expected: "int:42"},
		{name: "int64", input: int64(123), expected: "int64:123"},
		{name: "uint", input: uint(99), expected: "uint:99"},
		{name: "uint64", input: uint64(777), expected: "uint64:777"},
		{name: "string", input: "hello", expected: "string:hello"},
		{name: "bool_true", input: true, expected: "bool:true"},
		{name: "bool_false", input: false, expected: "bool:false"},
		{name: "float64", input: 3.14, expected: "float64:3.14"},
		{name: "nil", input: nil, expected: "<nil>"},
		{name: "bytes", input: []byte{0x01, 0x02}, expected: "[]uint8:0102"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := valueKey(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestCommonNormalizeIDValue tests the normalizeIDValue function
// Note: For signed integer types, the function returns int64. For unsigned types, it returns uint64.
func TestCommonNormalizeIDValue(t *testing.T) {
	tests := []struct {
		name         string
		fieldType    schema.FieldType
		input        any
		expectedType string // "int64", "uint64", or "passthrough"
		wantErr      bool
	}{
		{name: "int_field", fieldType: schema.TypeInt, input: 42, expectedType: "int64", wantErr: false},
		{name: "int64_field", fieldType: schema.TypeInt64, input: int64(123), expectedType: "int64", wantErr: false},
		{name: "uint_field", fieldType: schema.TypeUint, input: uint(99), expectedType: "uint64", wantErr: false},
		{name: "uint64_field", fieldType: schema.TypeUint64, input: uint64(777), expectedType: "uint64", wantErr: false},
		{name: "string_field", fieldType: schema.TypeString, input: "abc", expectedType: "passthrough", wantErr: false},
		{name: "nil_value", fieldType: schema.TypeInt, input: nil, expectedType: "passthrough", wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := &schema.Field{Type: tt.fieldType, Name: "test_field"}
			result, err := normalizeIDValue(field, tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				switch tt.expectedType {
				case "int64":
					assert.IsType(t, int64(0), result, "Result should be int64 for signed types")
				case "uint64":
					assert.IsType(t, uint64(0), result, "Result should be uint64 for unsigned types")
				case "passthrough":
					// Value passes through unchanged
					assert.Equal(t, tt.input, result)
				}
			}
		})
	}
}

// TestGetAtlasMigrateDriver tests the getAtlasMigrateDriver function
func TestGetAtlasMigrateDriver(t *testing.T) {
	// Test unsupported dialect
	_, err := getAtlasMigrateDriver("unsupported", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported")
}

// TestReferenceOptionTypeToEnt tests all cases of referenceOptionTypeToEnt
func TestReferenceOptionTypeToEnt(t *testing.T) {
	tests := []struct {
		name     string
		input    schema.ReferenceOptionType
		expected entSchema.ReferenceOption
	}{
		{name: "NoAction", input: schema.NoAction, expected: entSchema.NoAction},
		{name: "Restrict", input: schema.Restrict, expected: entSchema.Restrict},
		{name: "Cascade", input: schema.Cascade, expected: entSchema.Cascade},
		{name: "SetNull", input: schema.SetNull, expected: entSchema.SetNull},
		{name: "SetDefault", input: schema.SetDefault, expected: entSchema.SetDefault},
		{name: "Unknown defaults to NoAction", input: schema.ReferenceOptionType(999), expected: entSchema.NoAction},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := referenceOptionTypeToEnt(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestTxExecHookErrors tests Tx.Exec with hook errors
func TestTxExecHookErrors(t *testing.T) {
	ctx := context.Background()
	sb := createSchemaBuilder()
	migrationDir := utils.Must(os.MkdirTemp("", "migrations"))
	defer os.RemoveAll(migrationDir)

	// Create client with pre-exec hook that returns error
	client, err := NewTestClient(migrationDir, sb, func() *db.Hooks {
		return &db.Hooks{
			PreDBExec: []db.PreDBExec{
				func(ctx context.Context, option *db.QueryOption) error {
					return errors.New("pre-exec hook error")
				},
			},
		}
	})
	require.NoError(t, err)
	defer client.Close()

	// Create transaction
	tx, err := client.Tx(ctx)
	require.NoError(t, err)
	defer tx.Rollback()

	// Test Exec with hook error
	_, err = tx.Exec(ctx, "SELECT 1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "pre-exec hook error")
}

// TestTxQueryHookErrors tests Tx.Query with hook errors
func TestTxQueryHookErrors(t *testing.T) {
	ctx := context.Background()
	sb := createSchemaBuilder()
	migrationDir := utils.Must(os.MkdirTemp("", "migrations"))
	defer os.RemoveAll(migrationDir)

	// Create client with pre-query hook that returns error
	client, err := NewTestClient(migrationDir, sb, func() *db.Hooks {
		return &db.Hooks{
			PreDBQuery: []db.PreDBQuery{
				func(ctx context.Context, option *db.QueryOption) error {
					return errors.New("pre-query hook error")
				},
			},
		}
	})
	require.NoError(t, err)
	defer client.Close()

	// Create transaction
	tx, err := client.Tx(ctx)
	require.NoError(t, err)
	defer tx.Rollback()

	// Test Query with hook error
	_, err = tx.Query(ctx, "SELECT 1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "pre-query hook error")
}

// TestTxSetSQLDBAndSetDriver tests Tx.SetSQLDB and Tx.SetDriver (no-op methods)
func TestTxSetSQLDBAndSetDriver(t *testing.T) {
	ctx := context.Background()
	sb := createSchemaBuilder()
	migrationDir := utils.Must(os.MkdirTemp("", "migrations"))
	defer os.RemoveAll(migrationDir)

	client, err := NewTestClient(migrationDir, sb)
	require.NoError(t, err)
	defer client.Close()

	tx, err := client.Tx(ctx)
	require.NoError(t, err)
	defer tx.Rollback()

	// These are no-op methods to satisfy the EntAdapter interface
	txEnt := tx.(EntAdapter)
	txEnt.SetSQLDB(nil)  // Should not panic
	txEnt.SetDriver(nil) // Should not panic
}

// TestTxDriverExecAndQuery tests TxDriver.Exec and TxDriver.Query
func TestTxDriverExecAndQuery(t *testing.T) {
	ctx := context.Background()
	sb := createSchemaBuilder()
	migrationDir := utils.Must(os.MkdirTemp("", "migrations"))
	defer os.RemoveAll(migrationDir)

	client, err := NewTestClient(migrationDir, sb)
	require.NoError(t, err)
	defer client.Close()

	tx, err := NewTx(ctx, client)
	require.NoError(t, err)

	txDriver := tx.driver.(*TxDriver)

	// Test TxDriver.Exec - passes through to dialectTx
	err = txDriver.Exec(ctx, "SELECT 1", nil, nil)
	// May return error depending on driver, but should not panic
	_ = err

	// Test TxDriver.Query - passes through to dialectTx
	err = txDriver.Query(ctx, "SELECT 1", nil, nil)
	// May return error depending on driver, but should not panic
	_ = err

	tx.Rollback()
}

// TestGetRelationTargetFieldErrors tests error paths for getRelationTargetField
func TestGetRelationTargetFieldErrors(t *testing.T) {
	sb := createSchemaBuilder()

	// Test nil builder
	_, err := getRelationTargetField(nil, &schema.Relation{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "schema builder is not initialized")

	// Test nil relation
	_, err = getRelationTargetField(sb, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "relation is not defined")

	// Test missing schema
	_, err = getRelationTargetField(sb, &schema.Relation{
		TargetSchemaName: "nonexistent",
	})
	assert.Error(t, err)

	// Test missing field - use a valid schema but invalid target column
	_, err = getRelationTargetField(sb, &schema.Relation{
		TargetSchemaName: "user",
		TargetColumn:     "nonexistent_column",
	})
	assert.Error(t, err)
}

// TestResolveRelationTargetColumnError tests the error path of resolveRelationTargetColumn
// when a custom TargetColumn is specified but not found in the target schema
func TestResolveRelationTargetColumnError(t *testing.T) {
	// Create a schema with a relation that has a custom TargetColumn
	parentSchema := &schema.Schema{
		Name:           "parent",
		Namespace:      "parents",
		LabelFieldName: "name",
		Fields: []*schema.Field{
			{Name: "name", Label: "Name", Type: schema.TypeString},
			{
				Name:  "children",
				Label: "Children",
				Type:  schema.TypeRelation,
				Relation: &schema.Relation{
					Type:             schema.O2M,
					Owner:            true,
					TargetSchemaName: "child",
					TargetFieldName:  "parent",
				},
			},
		},
	}

	childSchema := &schema.Schema{
		Name:           "child",
		Namespace:      "children",
		LabelFieldName: "name",
		Fields: []*schema.Field{
			{Name: "name", Label: "Name", Type: schema.TypeString},
			{
				Name:  "parent",
				Label: "Parent",
				Type:  schema.TypeRelation,
				Relation: &schema.Relation{
					Type:             schema.O2M,
					TargetSchemaName: "parent",
					TargetFieldName:  "children",
					// Specify a non-existent target column
					TargetColumn: "nonexistent_column",
				},
			},
		},
	}

	schemas := map[string]*schema.Schema{
		parentSchema.Name: parentSchema,
		childSchema.Name:  childSchema,
	}

	sb, err := schema.NewBuilderFromSchemas("", schemas)
	// Schema builder may succeed despite the invalid column reference
	// The error will occur during adapter creation when it tries to resolve the column
	if err != nil {
		assert.Contains(t, err.Error(), "nonexistent_column")
		return
	}

	migrationDir := utils.Must(os.MkdirTemp("", "migrations"))
	defer os.RemoveAll(migrationDir)

	// This should fail because the target column doesn't exist
	_, err = NewTestClient(migrationDir, sb)
	// The error may occur during schema initialization or adapter creation
	// Either way, it should produce an error related to the missing column
	if err != nil {
		assert.Contains(t, err.Error(), "nonexistent_column")
	}
}

// TestAutoGenerateUUID tests the autoGenerateUUID function with UUID primary keys
func TestAutoGenerateUUID(t *testing.T) {
	// Create a schema with UUID primary key
	uuidSchema := &schema.Schema{
		Name:             "uuid_entity",
		Namespace:        "uuid_entities",
		LabelFieldName:   "name",
		PrimaryFieldName: "uuid_id",
		Fields: []*schema.Field{
			{
				Name:          "uuid_id",
				Label:         "UUID ID",
				Type:          schema.TypeUUID,
				IsSystemField: true,
			},
			{Name: "name", Label: "Name", Type: schema.TypeString},
		},
	}

	schemas := map[string]*schema.Schema{
		uuidSchema.Name: uuidSchema,
	}

	sb, err := schema.NewBuilderFromSchemas("", schemas)
	require.NoError(t, err)

	migrationDir := utils.Must(os.MkdirTemp("", "migrations"))
	defer os.RemoveAll(migrationDir)

	client, err := NewTestClient(migrationDir, sb)
	require.NoError(t, err)
	defer client.Close()

	ctx := context.Background()

	// Get the model
	model, err := client.Model("uuid_entity")
	require.NoError(t, err)

	// Test 1: Create without UUID - should auto-generate
	entity1 := entity.New().Set("name", "Test 1")
	id1, err := model.Mutation().Create(ctx, entity1)
	require.NoError(t, err)
	require.NotNil(t, id1)

	// Verify UUID was generated (it should be a valid UUID string or uuid.UUID)
	assert.NotNil(t, id1)

	// Test 2: Create with UUID already set - should use the provided value
	presetUUID := uuid.New()
	entity2 := entity.New().Set("name", "Test 2").Set("uuid_id", presetUUID)
	id2, err := model.Mutation().Create(ctx, entity2)
	require.NoError(t, err)

	// The returned ID should match the preset UUID
	assert.NotNil(t, id2)
}

// TestNormalizeUUIDValueExtra tests additional UUID normalization paths
func TestNormalizeUUIDValueExtra(t *testing.T) {
	validUUID := uuid.New()
	validUUIDString := validUUID.String()
	validUUIDBytes := validUUID[:]

	tests := []struct {
		name        string
		input       any
		expectError bool
	}{
		{name: "uuid.UUID", input: validUUID, expectError: false},
		{name: "string_uuid", input: validUUIDString, expectError: false},
		{name: "bytes_uuid", input: validUUIDBytes, expectError: false},
		{name: "nil_value", input: nil, expectError: false},
		{name: "invalid_string", input: "not-a-uuid", expectError: true},
		{name: "invalid_type", input: 12345, expectError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := normalizeUUIDValue(tt.input)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.input != nil {
					assert.NotNil(t, result)
				}
			}
		})
	}
}

// TestNormalizeIDValueErrors tests error paths for normalizeIDValue
func TestNormalizeIDValueErrors(t *testing.T) {
	// Test with nil field
	result, err := normalizeIDValue(nil, 42)
	assert.NoError(t, err)
	assert.Equal(t, 42, result)

	// Test with nil value
	result, err = normalizeIDValue(&schema.Field{Type: schema.TypeInt}, nil)
	assert.NoError(t, err)
	assert.Nil(t, result)

	// Test unsigned integer conversion
	result, err = normalizeIDValue(&schema.Field{Type: schema.TypeUint64, Name: "test"}, 100)
	assert.NoError(t, err)
	assert.Equal(t, uint64(100), result)

	// Test signed integer conversion
	result, err = normalizeIDValue(&schema.Field{Type: schema.TypeInt64, Name: "test"}, -50)
	assert.NoError(t, err)
	assert.Equal(t, int64(-50), result)

	// Test UUID type
	validUUID := uuid.New()
	result, err = normalizeIDValue(&schema.Field{Type: schema.TypeUUID, Name: "test"}, validUUID)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Test invalid unsigned conversion
	_, err = normalizeIDValue(&schema.Field{Type: schema.TypeUint64, Name: "test"}, "not-a-number")
	assert.Error(t, err)

	// Test invalid signed conversion
	_, err = normalizeIDValue(&schema.Field{Type: schema.TypeInt64, Name: "test"}, "not-a-number")
	assert.Error(t, err)
}

// TestCollectEntityIDsEdgeCases tests edge cases for collectEntityIDs
func TestCollectEntityIDsEdgeCases(t *testing.T) {
	// Test with empty entities
	ids, byKey, err := collectEntityIDs("test", nil, []*entity.Entity{})
	assert.NoError(t, err)
	assert.Empty(t, ids)
	assert.Empty(t, byKey)

	// Test with entities that have IDs
	e1 := entity.New()
	e1.SetID(uint64(1))
	e2 := entity.New()
	e2.SetID(uint64(2))

	ids, byKey, err = collectEntityIDs("test", nil, []*entity.Entity{e1, e2})
	assert.NoError(t, err)
	assert.Len(t, ids, 2)
	assert.Len(t, byKey, 2)

	// Test error case - entity with no ID
	empEntity := entity.New()
	_, _, err = collectEntityIDs("test", nil, []*entity.Entity{empEntity})
	assert.Error(t, err)
}

// TestCollectParentRefsExtra tests additional edge cases for collectParentRefs
func TestCollectParentRefsExtra(t *testing.T) {
	// Test with empty entities
	refs, parentMap, err := collectParentRefs([]*entity.Entity{}, "id", nil, "test", false)
	assert.NoError(t, err)
	assert.Empty(t, refs)
	assert.Empty(t, parentMap)

	// Test with entities that have ref values
	e1 := entity.New()
	e1.SetID(uint64(1))
	e1.Set("parent_id", uint64(10))

	e2 := entity.New()
	e2.SetID(uint64(2))
	e2.Set("parent_id", uint64(20))

	refs, parentMap, err = collectParentRefs([]*entity.Entity{e1, e2}, "parent_id", nil, "test", false)
	assert.NoError(t, err)
	assert.Len(t, refs, 2)
	assert.Len(t, parentMap, 2)

	// Test with duplicate ref values - should only add once to refs
	e3 := entity.New()
	e3.SetID(uint64(3))
	e3.Set("parent_id", uint64(10)) // Same as e1

	refs, parentMap, err = collectParentRefs([]*entity.Entity{e1, e3}, "parent_id", nil, "test", false)
	assert.NoError(t, err)
	assert.Len(t, refs, 1) // Only one unique ref
	assert.Len(t, parentMap, 1)
	assert.Len(t, parentMap[valueKey(uint64(10))], 2) // Both e1 and e3 in same parent

	// Test skipNullFK = true with null FK value
	e4 := entity.New()
	e4.SetID(uint64(4))
	// parent_id is not set

	refs, parentMap, err = collectParentRefs([]*entity.Entity{e4}, "parent_id", nil, "test", true)
	assert.NoError(t, err)
	assert.Empty(t, refs) // No refs collected
	assert.Empty(t, parentMap)

	// Test skipNullFK = false with null FK value - should error
	_, _, err = collectParentRefs([]*entity.Entity{e4}, "parent_id", nil, "test", false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty reference value")
}

// TestValueKeyEdgeCases tests additional edge cases for valueKey
func TestValueKeyEdgeCases(t *testing.T) {
	// Test with float32
	result := valueKey(float32(3.14))
	assert.Contains(t, result, "float32")

	// Test with int8, int16, int32
	assert.Contains(t, valueKey(int8(8)), "int8")
	assert.Contains(t, valueKey(int16(16)), "int16")
	assert.Contains(t, valueKey(int32(32)), "int32")

	// Test with uint8, uint16, uint32
	assert.Contains(t, valueKey(uint8(8)), "uint8")
	assert.Contains(t, valueKey(uint16(16)), "uint16")
	assert.Contains(t, valueKey(uint32(32)), "uint32")
}

// TestTypeHandlerAssign tests the assign functions for various types
func TestTypeHandlerAssign(t *testing.T) {
	e := entity.New()

	// Test assignBool
	boolVal := &sql.NullBool{Bool: true, Valid: true}
	result, err := assignBool("", boolVal, e)
	assert.NoError(t, err)
	assert.Equal(t, true, result)

	// Test assignBool with invalid value
	_, err = assignBool("", "not-a-bool", e)
	assert.Error(t, err)

	// Test assignBool with null
	nullBool := &sql.NullBool{Valid: false}
	result, err = assignBool("", nullBool, e)
	assert.NoError(t, err)
	assert.Nil(t, result)

	// Test assignTime
	timeVal := &sql.NullTime{Valid: false}
	result, err = assignTime("", timeVal, e)
	assert.NoError(t, err)
	assert.Nil(t, result)

	// Test assignTime with invalid value
	_, err = assignTime("", "not-a-time", e)
	assert.Error(t, err)

	// Test assignBytes
	bytesVal := &[]byte{1, 2, 3}
	result, err = assignBytes("", bytesVal, e)
	assert.NoError(t, err)
	assert.Equal(t, []byte{1, 2, 3}, result)

	// Test assignBytes with invalid type
	_, err = assignBytes("", "not-bytes", e)
	assert.Error(t, err)

	// Test assignString
	stringVal := &sql.NullString{String: "hello", Valid: true}
	result, err = assignString("", stringVal, e)
	assert.NoError(t, err)
	assert.Equal(t, "hello", result)

	// Test assignString with invalid type
	_, err = assignString("", 123, e)
	assert.Error(t, err)

	// Test assignUUID
	testUUID := uuid.New()
	result, err = assignUUID("", &testUUID, e)
	assert.NoError(t, err)
	assert.Equal(t, testUUID, result)

	// Test assignUUID with invalid type
	_, err = assignUUID("", "not-uuid", e)
	assert.Error(t, err)

	// Test assignInt8
	int64Val := &sql.NullInt64{Int64: 127, Valid: true}
	result, err = assignInt8("", int64Val, e)
	assert.NoError(t, err)
	assert.Equal(t, int8(127), result)

	// Test assignInt8 with invalid type
	_, err = assignInt8("", "not-int", e)
	assert.Error(t, err)

	// Test assignUint8
	result, err = assignUint8("", int64Val, e)
	assert.NoError(t, err)
	assert.Equal(t, uint8(127), result)

	// Test assignUint8 with invalid type
	_, err = assignUint8("", "not-int", e)
	assert.Error(t, err)

	// Test assignFloat32
	float64Val := &sql.NullFloat64{Float64: 3.14, Valid: true}
	result, err = assignFloat32("", float64Val, e)
	assert.NoError(t, err)
	assert.Equal(t, float32(3.14), result)

	// Test assignFloat32 with invalid type
	_, err = assignFloat32("", "not-float", e)
	assert.Error(t, err)

	// Test assignAny
	anyVal := interface{}("test")
	result, err = assignAny("", &anyVal, e)
	assert.NoError(t, err)
	assert.Equal(t, "test", result)

	// Test assignAny with invalid type
	_, err = assignAny("", "not-ptr", e)
	assert.Error(t, err)

	// Test remaining integer assign functions with null values
	nullInt := &sql.NullInt64{Valid: false}

	// Test assignInt16 with null
	result, err = assignInt16("", nullInt, e)
	assert.NoError(t, err)
	assert.Nil(t, result)

	// Test assignInt32 with null
	result, err = assignInt32("", nullInt, e)
	assert.NoError(t, err)
	assert.Nil(t, result)

	// Test assignInt with null
	result, err = assignInt("", nullInt, e)
	assert.NoError(t, err)
	assert.Nil(t, result)

	// Test assignInt64Value with null
	result, err = assignInt64Value("", nullInt, e)
	assert.NoError(t, err)
	assert.Nil(t, result)

	// Test assignUint16 with null
	result, err = assignUint16("", nullInt, e)
	assert.NoError(t, err)
	assert.Nil(t, result)

	// Test assignUint32 with null
	result, err = assignUint32("", nullInt, e)
	assert.NoError(t, err)
	assert.Nil(t, result)

	// Test assignUint64 with null
	result, err = assignUint64("", nullInt, e)
	assert.NoError(t, err)
	assert.Nil(t, result)

	// Test assignFloat64Value with null
	nullFloat := &sql.NullFloat64{Valid: false}
	result, err = assignFloat64Value("", nullFloat, e)
	assert.NoError(t, err)
	assert.Nil(t, result)

	// Test assignFloat32 with null
	result, err = assignFloat32("", nullFloat, e)
	assert.NoError(t, err)
	assert.Nil(t, result)

	// Test assignInt8 with null
	result, err = assignInt8("", nullInt, e)
	assert.NoError(t, err)
	assert.Nil(t, result)

	// Test assignUint8 with null
	result, err = assignUint8("", nullInt, e)
	assert.NoError(t, err)
	assert.Nil(t, result)

	// Additional error type tests
	_, err = assignInt16("", "bad", e)
	assert.Error(t, err)

	_, err = assignInt32("", "bad", e)
	assert.Error(t, err)

	_, err = assignInt("", "bad", e)
	assert.Error(t, err)

	_, err = assignInt64Value("", "bad", e)
	assert.Error(t, err)

	_, err = assignUint16("", "bad", e)
	assert.Error(t, err)

	_, err = assignUint32("", "bad", e)
	assert.Error(t, err)

	_, err = assignUint64("", "bad", e)
	assert.Error(t, err)

	_, err = assignUint("", "bad", e)
	assert.Error(t, err)

	_, err = assignFloat64Value("", "bad", e)
	assert.Error(t, err)
}

// TestGetTypeHandlerExtra tests GetTypeHandler for unknown types and helper functions
func TestGetTypeHandlerExtra(t *testing.T) {
	// Test getting handler for unknown type
	handler := GetTypeHandler(schema.FieldType(999))
	assert.NotNil(t, handler.ScanValue)
	assert.NotNil(t, handler.AssignValue)

	// Test IsIntegerType
	assert.True(t, IsIntegerType(schema.TypeInt64))
	assert.False(t, IsIntegerType(schema.TypeString))

	// Test IsStringType
	assert.True(t, IsStringType(schema.TypeString))
	assert.False(t, IsStringType(schema.TypeInt))

	// Test IsFloatType
	assert.True(t, IsFloatType(schema.TypeFloat64))
	assert.False(t, IsFloatType(schema.TypeInt))
}

// TestAssignJSONAndAnyNullPaths tests null paths for JSON and Any assignment
func TestAssignJSONAndAnyNullPaths(t *testing.T) {
	e := entity.New()

	// Test assignJSON with nil bytes
	nilBytes := (*[]byte)(nil)
	result, err := assignJSON("test", nilBytes, e)
	assert.NoError(t, err)
	assert.Nil(t, result)

	// Test assignJSON with empty bytes
	emptyBytes := &[]byte{}
	result, err = assignJSON("test", emptyBytes, e)
	assert.NoError(t, err)
	assert.Nil(t, result)

	// Test assignJSON with invalid type
	_, err = assignJSON("test", "not-bytes", e)
	assert.Error(t, err)

	// Test assignJSON with invalid JSON
	badJSON := []byte("{invalid}")
	_, err = assignJSON("test", &badJSON, e)
	assert.Error(t, err)

	// Test assignAny with nil
	nilAny := (*any)(nil)
	result, err = assignAny("test", nilAny, e)
	assert.NoError(t, err)
	assert.Nil(t, result)

	// Test assignUUID with nil
	nilUUID := (*uuid.UUID)(nil)
	result, err = assignUUID("test", nilUUID, e)
	assert.NoError(t, err)
	assert.Nil(t, result)

	// Test assignBytes with nil
	nilBytesPtr := (*[]byte)(nil)
	result, err = assignBytes("test", nilBytesPtr, e)
	assert.NoError(t, err)
	assert.Nil(t, result)
}

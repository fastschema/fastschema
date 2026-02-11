package entdbadapter

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/fastschema/fastschema/entity"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
	"github.com/stretchr/testify/assert"
)

func TestScan(t *testing.T) {
	migrationDir := utils.Must(os.MkdirTemp("", "migrations"))
	adapter, err := NewTestClient(
		migrationDir,
		utils.Must(schema.NewBuilderFromDir(migrationDir, fs.SystemSchemaTypes...)),
	)

	ctx := context.Background()
	assert.NoError(t, err)
	assert.NotNil(t, adapter)
	_, err = adapter.Query(ctx, "SELECT FROM", []any{})
	assert.Error(t, err)
	_, err = adapter.Exec(ctx, "SELECT FROM", []any{})
	assert.Error(t, err)

	rows, err := adapter.Query(
		ctx,
		`SELECT
			? as bool_column,
			? as int_column,
			? as int64_column,
			? as uint64_column,
			? as time_column
		`,
		[]any{
			true,
			5,
			int64(1),
			uint64(1),
			time.Now(),
		}...,
	)
	assert.NoError(t, err)
	assert.NotNil(t, rows)
	assert.Len(t, rows, 1)

	json := rows[0].String()

	assert.Contains(t, json, `"bool_column":1`)
	assert.Contains(t, json, `"int_column":5`)
	assert.Contains(t, json, `"int64_column":1`)
	assert.Contains(t, json, `"uint64_column":1`)
	assert.Contains(t, json, `"time_column":`)

	result, err := adapter.Exec(ctx, "SELECT 1")
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

type testSQLColumnType struct {
	scanType         reflect.Type
	databaseTypeName string
}

func (t *testSQLColumnType) ScanType() reflect.Type {
	return t.scanType
}

func (t *testSQLColumnType) DatabaseTypeName() string {
	return t.databaseTypeName
}

func createSQLColumnType[T any](val T, dbTypeNames ...any) (string, SQLColumnType, T) {
	rtype := reflect.TypeOf(val)

	columnName := "column"
	if rtype != nil {
		columnName = rtype.Name()
	}

	return columnName, &testSQLColumnType{
		scanType:         rtype,
		databaseTypeName: fmt.Sprintf("%v", append(dbTypeNames, "")[0]),
	}, utils.CreateZeroValue(rtype).(T)
}

func TestCreateRowsScanValues(t *testing.T) {
	type testStruct struct{}

	testValues := [][]any{
		{true},
		{int(1)},
		{int8(1)},
		{int16(1)},
		{int32(1)},
		{int64(1)},
		{uint(1)},
		{uint8(1)},
		{uint16(1)},
		{uint32(1)},
		{uint64(1)},
		{uintptr(1)},
		{float32(1)},
		{float64(1)},
		{complex64(1)},
		{complex128(1)},
		{[1]int{1}, new([]any)},
		{[]any{1}},
		{map[string]string{"1": "1"}, new(map[string]any)},
		{[]int{1}, new([]any)},
		{"1"},
		{sql.NullString{}, new(string)},
		{sql.NullInt64{}, new(int64)},
		{sql.NullInt32{}, new(int32)},
		{sql.NullInt16{}, new(int16)},
		{sql.NullByte{}, new(uint8)},
		{sql.NullFloat64{}, new(float64)},
		{sql.NullBool{}, new(bool)},
		{sql.NullTime{}},
		{testStruct{}, new(any)},

		{time.Time{}, new(time.Time)},
	}

	columns := []SQLColumn{}
	expected := []any{}

	for _, testValue := range testValues {
		dbTypeName := ""
		if len(testValue) > 2 {
			dbTypeName = testValue[2].(string)
		}

		columnName, columnType, value := createSQLColumnType(testValue[0], dbTypeName)
		columns = append(columns, SQLColumn{
			Name:      columnName,
			Type:      columnType,
			FieldType: schema.FieldTypeFromReflectType(columnType.ScanType()),
		})

		if len(testValue) > 1 {
			expected = append(expected, testValue[1])
		} else {
			expected = append(expected, &value)
		}
	}

	values := rawRowsScanValues(columns)
	assert.NotNil(t, values)

	for i, e := range expected {
		expectedType := spew.Sprintf("%T", e)
		valuesType := spew.Sprintf("%T", values[i])
		assert.Equal(t, expectedType, valuesType)
	}
}

func TestIsDateTimeColumn(t *testing.T) {
	// Test for struct time type and DATETIME database type
	scanType := reflect.TypeOf(time.Time{})
	databaseTypeName := "DATETIME"
	result := isDateTimeColumn(scanType, databaseTypeName)
	assert.True(t, result)

	// Test for struct time type and non-DATETIME database type
	scanType = reflect.TypeOf(time.Time{})
	databaseTypeName = "VARCHAR"
	result = isDateTimeColumn(scanType, databaseTypeName)
	assert.True(t, result)

	// Test for non-struct time type and DATETIME database type
	scanType = reflect.TypeOf("")
	databaseTypeName = "DATETIME"
	result = isDateTimeColumn(scanType, databaseTypeName)
	assert.True(t, result)

	// Test for non-struct time type and non-DATETIME database type
	scanType = reflect.TypeOf("")
	databaseTypeName = "VARCHAR"
	result = isDateTimeColumn(scanType, databaseTypeName)
	assert.False(t, result)
}

// TestColumnScanValue tests the scan value creation for different field types
func TestColumnScanValue(t *testing.T) {
	tests := []struct {
		name      string
		fieldType schema.FieldType
		wantType  string
	}{
		{"JSON", schema.TypeJSON, "*[]uint8"},
		{"Bytes", schema.TypeBytes, "*[]uint8"},
		{"Bool", schema.TypeBool, "*sql.NullBool"},
		{"Float32", schema.TypeFloat32, "*sql.NullFloat64"},
		{"Float64", schema.TypeFloat64, "*sql.NullFloat64"},
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
		{"Enum", schema.TypeEnum, "*sql.NullString"},
		{"String", schema.TypeString, "*sql.NullString"},
		{"Text", schema.TypeText, "*sql.NullString"},
		{"Time", schema.TypeTime, "*sql.NullTime"},
		{"UUID", schema.TypeUUID, "*uuid.UUID"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := columnScanValue(tt.fieldType)
			gotType := reflect.TypeOf(result).String()
			assert.Equal(t, tt.wantType, gotType)
		})
	}
}

// TestSchemaAssignValues tests the assignment of scanned values to entity
func TestSchemaAssignValues(t *testing.T) {
	testSchema := &schema.Schema{
		Name:           "test",
		Namespace:      "tests",
		LabelFieldName: "name",
		Fields: []*schema.Field{
			{Name: "id", Type: schema.TypeUint64, DB: &schema.FieldDB{Attr: "PRIMARY_KEY"}},
			{Name: "name", Type: schema.TypeString},
			{Name: "age", Type: schema.TypeInt},
			{Name: "active", Type: schema.TypeBool},
			{Name: "score", Type: schema.TypeFloat64},
		},
	}
	assert.NoError(t, testSchema.Init(false))

	t.Run("assign string value", func(t *testing.T) {
		e := entity.New()
		nameValue := &sql.NullString{String: "John", Valid: true}

		err := schemaAssignValues(testSchema, e, []string{"name"}, []any{nameValue})
		assert.NoError(t, err)
		assert.Equal(t, "John", e.Get("name"))
	})

	t.Run("assign bool value", func(t *testing.T) {
		e := entity.New()
		activeValue := &sql.NullBool{Bool: true, Valid: true}

		err := schemaAssignValues(testSchema, e, []string{"active"}, []any{activeValue})
		assert.NoError(t, err)
		assert.Equal(t, true, e.Get("active"))
	})

	t.Run("assign float value", func(t *testing.T) {
		e := entity.New()
		scoreValue := &sql.NullFloat64{Float64: 95.5, Valid: true}

		err := schemaAssignValues(testSchema, e, []string{"score"}, []any{scoreValue})
		assert.NoError(t, err)
		assert.Equal(t, 95.5, e.Get("score"))
	})

	t.Run("assign null value", func(t *testing.T) {
		e := entity.New()
		nameValue := &sql.NullString{Valid: false}

		err := schemaAssignValues(testSchema, e, []string{"name"}, []any{nameValue})
		assert.NoError(t, err)
		// Null values should not be set
		assert.Nil(t, e.Get("name"))
	})

	t.Run("error on mismatched column count", func(t *testing.T) {
		e := entity.New()
		err := schemaAssignValues(testSchema, e, []string{"name", "age"}, []any{&sql.NullString{}})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "mismatch number of scan values")
	})

	t.Run("unknown column creates new any field", func(t *testing.T) {
		e := entity.New()
		err := schemaAssignValues(testSchema, e, []string{"unknown_col"}, []any{new(any)})
		assert.NoError(t, err)
		// Unknown columns should still be set
		assert.NotNil(t, e.Get("unknown_col"))
	})
}

// TestColumnAssignValueErrors tests error cases for column assignment
func TestColumnAssignValueErrors(t *testing.T) {
	e := entity.New()

	t.Run("bool type mismatch", func(t *testing.T) {
		_, err := columnAssignValue("test", schema.TypeBool, "not a bool", e)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "*sql.NullBool")
	})

	t.Run("time type mismatch", func(t *testing.T) {
		_, err := columnAssignValue("test", schema.TypeTime, "not a time", e)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "*sql.NullTime")
	})

	t.Run("JSON type mismatch", func(t *testing.T) {
		_, err := columnAssignValue("test", schema.TypeJSON, "not bytes", e)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "*[]byte")
	})

	t.Run("UUID type mismatch", func(t *testing.T) {
		_, err := columnAssignValue("test", schema.TypeUUID, "not uuid", e)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "*uuid.UUID")
	})
}

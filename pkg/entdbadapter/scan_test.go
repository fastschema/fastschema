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
			? as int64_column,
			? as uint64_column,
			? as time_column
		`,
		[]any{
			true,
			1,
			int64(1),
			uint64(1),
			time.Now(),
		},
	)
	assert.NoError(t, err)
	assert.NotNil(t, rows)
	assert.Len(t, rows, 1)

	json := rows[0].String()

	assert.Contains(t, json, `"bool_column":1`)
	assert.Contains(t, json, `"int64_column":1`)
	assert.Contains(t, json, `"uint64_column":1`)
	assert.Contains(t, json, `"time_column":`)

	result, err := adapter.Exec(ctx, "SELECT 1", []any{})
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

	columns := []string{}
	columnTypes := []SQLColumnType{}
	expected := []any{}

	for _, testValue := range testValues {
		dbTypeName := ""
		if len(testValue) > 2 {
			dbTypeName = testValue[2].(string)
		}

		columnName, columnType, value := createSQLColumnType(testValue[0], dbTypeName)
		columns = append(columns, columnName)
		columnTypes = append(columnTypes, columnType)

		if len(testValue) > 1 {
			expected = append(expected, testValue[1])
		} else {
			expected = append(expected, &value)
		}
	}

	values := createRowsScanValues(columns, columnTypes)
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

package entdbadapter

import (
	"database/sql"
	"fmt"
	"reflect"

	dialectsql "entgo.io/ent/dialect/sql"
	"github.com/fastschema/fastschema/entity"
	"github.com/fastschema/fastschema/schema"
)

type SQLColumnType interface {
	ScanType() reflect.Type
	DatabaseTypeName() string
}

type SQLColumn struct {
	Name      string
	Type      SQLColumnType
	FieldType schema.FieldType
}

func getRowsColumns(rows *dialectsql.Rows) ([]SQLColumn, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("[getRowsColumns] failed to get columns: %w", err)
	}

	columnTypes, err := rows.ColumnTypes()
	if err != nil {
		return nil, fmt.Errorf("[createRowsScanValues] failed to get column types: %w", err)
	}

	sqlColumns := make([]SQLColumn, len(columns))
	for i, column := range columns {
		var fieldType schema.FieldType
		if isDateTimeColumn(columnTypes[i].ScanType(), columnTypes[i].DatabaseTypeName()) {
			fieldType = schema.TypeTime
		} else {
			fieldType = schema.FieldTypeFromReflectType(columnTypes[i].ScanType())
		}
		sqlColumns[i] = SQLColumn{
			Name:      column,
			Type:      columnTypes[i],
			FieldType: fieldType,
		}
	}

	return sqlColumns, nil
}

func rawRowsScanValues(columns []SQLColumn) []any {
	values := make([]any, len(columns))
	for i, column := range columns {
		scanType := column.Type.ScanType()
		databaseTypeName := column.Type.DatabaseTypeName()

		if isDateTimeColumn(scanType, databaseTypeName) {
			values[i] = new(sql.NullTime)
			continue
		}

		if scanType == nil {
			values[i] = new(any)
			continue
		}

		values[i] = columnScanValue(schema.FieldTypeFromReflectType(scanType))
	}

	return values
}

// schemaScanValues create a slice of scan values for the given columns.
func schemaScanValues(s *schema.Schema, columns []string) (_ []any, err error) {
	values := make([]any, len(columns))
	for i := range columns {
		field := s.Field(columns[i])
		if field == nil {
			values[i] = new(any)
			continue
		}

		if values[i] = columnScanValue(field.Type); values[i] == nil {
			return nil, fmt.Errorf("unexpected column %q for schema %s", columns[i], s.Name)
		}
	}
	return values, nil
}

// schemaAssignValues assigns the given values to the entity.
func schemaAssignValues(
	s *schema.Schema,
	e *entity.Entity,
	columns []string,
	values []any,
) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}

	primaryKey := ""
	if s != nil {
		primaryKey = s.PrimaryKeyName()
	}

	for i := range columns {
		field := s.Field(columns[i])
		if field == nil { // No field found. Ignore it.
			e.Set(columns[i], new(any))
			continue
		}

		v, err := columnAssignValue(
			field.Name,
			field.Type,
			values[i],
			e,
		)

		if err != nil {
			return fmt.Errorf("getColumnAssignValue for field %s: %w", field.Name, err)
		}

		if v != nil {
			e.Set(columns[i], v)
		}
	}

	if primaryKey != "" {
		e.SetIDField(primaryKey)
	}

	return nil
}

// columnScanValue returns a pointer suitable for sql.Rows.Scan for the given field type.
// Uses the unified TypeHandler system for consistent type handling.
func columnScanValue(fieldType schema.FieldType) any {
	return GetTypeHandler(fieldType).ScanValue()
}

// columnAssignValue converts a scanned value to the appropriate Go type for the field.
// Uses the unified TypeHandler system for consistent type handling.
func columnAssignValue(
	column string,
	fieldType schema.FieldType,
	value any,
	entity *entity.Entity,
) (any, error) {
	return GetTypeHandler(fieldType).AssignValue(column, value, entity)
}

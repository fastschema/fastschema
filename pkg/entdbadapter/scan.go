package entdbadapter

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"reflect"

	dialectsql "entgo.io/ent/dialect/sql"
	"github.com/fastschema/fastschema/entity"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
	"github.com/google/uuid"
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
func schemaAssignValues(s *schema.Schema, entity *entity.Entity, columns []string, values []any) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		field := s.Field(columns[i])
		if field == nil { // No field found. Ignore it.
			entity.Set(columns[i], new(any))
			continue
		}

		v, err := columnAssignValue(
			field.Name,
			field.Type,
			values[i],
			entity,
		)
		if err != nil {
			return fmt.Errorf("getColumnAssignValue for field %s: %w", field.Name, err)
		}
		if v != nil {
			entity.Set(columns[i], v)
		}
	}

	return nil
}

func columnScanValue(fieldType schema.FieldType) any {
	switch fieldType {
	case schema.TypeJSON, schema.TypeBytes:
		return new([]byte)
	case schema.TypeBool:
		return new(sql.NullBool)
	case schema.TypeFloat32, schema.TypeFloat64:
		return new(sql.NullFloat64)
	case schema.TypeInt8, schema.TypeInt16, schema.TypeInt32, schema.TypeInt, schema.TypeInt64, schema.TypeUint8, schema.TypeUint16, schema.TypeUint32, schema.TypeUint, schema.TypeUint64:
		return new(sql.NullInt64)
	case schema.TypeEnum, schema.TypeString, schema.TypeText:
		return new(sql.NullString)
	case schema.TypeTime:
		return new(sql.NullTime)
	case schema.TypeUUID:
		return new(uuid.UUID)
	default:
		return new(any)
	}
}

func columnAssignValue(
	column string,
	fieldType schema.FieldType,
	value any,
	entity *entity.Entity,
) (any, error) {
	switch fieldType {
	case schema.TypeBool:
		if v, ok := value.(*sql.NullBool); !ok {
			return nil, fieldTypeError("*sql.NullBool", value)
		} else if v.Valid {
			return v.Bool, nil
		}
	case schema.TypeTime:
		if v, ok := value.(*sql.NullTime); !ok {
			return nil, fieldTypeError("*sql.NullTime", value)
		} else if v.Valid {
			return v.Time, nil
		}
	case schema.TypeJSON:
		if v, ok := value.(*[]byte); !ok {
			return nil, fieldTypeError("*[]byte", value)
		} else if v != nil && len(*v) > 0 {
			e := entity.Get(column)
			if err := json.Unmarshal(*v, &e); err != nil {
				return nil, fmt.Errorf("unmarshal field field_type_JSON: %w", err)
			}
			return e, nil
		}
	case schema.TypeUUID:
		if v, ok := value.(*uuid.UUID); !ok {
			return nil, fieldTypeError("*uuid.UUID", value)
		} else if v != nil {
			return *v, nil
		}
	case schema.TypeBytes:
		if v, ok := value.(*[]byte); !ok {
			return nil, fieldTypeError("*[]byte", value)
		} else if v != nil {
			return *v, nil
		}
	case schema.TypeEnum:
		if v, ok := value.(*sql.NullString); !ok {
			return nil, fieldTypeError("*sql.NullString", value)
		} else if v.Valid {
			return v.String, nil
		}
	case schema.TypeString:
		if v, ok := value.(*sql.NullString); !ok {
			return nil, fieldTypeError("*sql.NullString", value)
		} else if v.Valid {
			return v.String, nil
		}
	case schema.TypeText:
		if v, ok := value.(*sql.NullString); !ok {
			return nil, fieldTypeError("*sql.NullString", value)
		} else if v.Valid {
			return v.String, nil
		}
	case schema.TypeInt8:
		if v, ok := value.(*sql.NullInt64); !ok {
			return nil, fieldTypeError("*sql.NullInt64", value)
		} else if v.Valid {
			return int8(v.Int64), nil
		}
	case schema.TypeInt16:
		if v, ok := value.(*sql.NullInt64); !ok {
			return nil, fieldTypeError("*sql.NullInt64", value)
		} else if v.Valid {
			return int16(v.Int64), nil
		}
	case schema.TypeInt32:
		if v, ok := value.(*sql.NullInt64); !ok {
			return nil, fieldTypeError("*sql.NullInt64", value)
		} else if v.Valid {
			return int32(v.Int64), nil
		}
	case schema.TypeInt:
		if v, ok := value.(*sql.NullInt64); !ok {
			return nil, fieldTypeError("*sql.NullInt64", value)
		} else if v.Valid {
			return int(v.Int64), nil
		}
	case schema.TypeInt64:
		if v, ok := value.(*sql.NullInt64); !ok {
			return nil, fieldTypeError("*sql.NullInt64", value)
		} else if v.Valid {
			return v.Int64, nil
		}
	case schema.TypeUint8:
		if v, ok := value.(*sql.NullInt64); !ok {
			return nil, fieldTypeError("*sql.NullInt64", value)
		} else if v.Valid {
			return uint8(v.Int64), nil
		}
	case schema.TypeUint16:
		if v, ok := value.(*sql.NullInt64); !ok {
			return nil, fieldTypeError("*sql.NullInt64", value)
		} else if v.Valid {
			return uint16(v.Int64), nil
		}
	case schema.TypeUint32:
		if v, ok := value.(*sql.NullInt64); !ok {
			return nil, fieldTypeError("*sql.NullInt64", value)
		} else if v.Valid {
			return uint32(v.Int64), nil
		}
	case schema.TypeUint:
		if v, ok := value.(*sql.NullInt64); !ok {
			return nil, fieldTypeError("*sql.NullInt64", value)
		} else {
			return utils.If(v.Valid, uint(v.Int64), uint(0)), nil
		}
	case schema.TypeUint64:
		if v, ok := value.(*sql.NullInt64); !ok {
			return nil, fieldTypeError("*sql.NullInt64", value)
		} else if v.Valid {
			return uint64(v.Int64), nil
		}
	case schema.TypeFloat32:
		if v, ok := value.(*sql.NullFloat64); !ok {
			return nil, fieldTypeError("*sql.NullFloat64", value)
		} else if v.Valid {
			return float32(v.Float64), nil
		}
	case schema.TypeFloat64:
		if v, ok := value.(*sql.NullFloat64); !ok {
			return nil, fieldTypeError("*sql.NullFloat64", value)
		} else if v.Valid {
			return v.Float64, nil
		}
	default:
		if v, ok := value.(*any); !ok {
			return nil, fieldTypeError("*any", value)
		} else if v != nil {
			return *v, nil
		}
	}

	return nil, nil
}

package entdbadapter

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"reflect"

	dialectsql "entgo.io/ent/dialect/sql"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
	"github.com/google/uuid"
)

type SQLColumnType interface {
	ScanType() reflect.Type
	DatabaseTypeName() string
}

func getRowsColumns(rows *dialectsql.Rows) ([]string, []SQLColumnType, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, nil, fmt.Errorf("[getRowsColumns] failed to get columns: %w", err)
	}

	columnTypes, err := rows.ColumnTypes()
	if err != nil {
		return nil, nil, fmt.Errorf("[createRowsScanValues] failed to get column types: %w", err)
	}

	sqlColumnTypes := make([]SQLColumnType, len(columnTypes))
	for i, columnType := range columnTypes {
		sqlColumnTypes[i] = columnType
	}

	return columns, sqlColumnTypes, nil
}

func createRowsScanValues(columns []string, columnTypes []SQLColumnType) []any {
	values := make([]any, len(columns))
	for i := range columnTypes {
		scanType := columnTypes[i].ScanType()
		databaseTypeName := columnTypes[i].DatabaseTypeName()

		if isDateTimeColumn(scanType, databaseTypeName) {
			values[i] = new(sql.NullTime)
			continue
		}

		if scanType == nil {
			values[i] = new(any)

			continue
		}

		switch scanType.Kind() {
		case reflect.Bool:
			values[i] = new(bool)
		case reflect.Int:
			values[i] = new(int)
		case reflect.Int8:
			values[i] = new(int8)
		case reflect.Int16:
			values[i] = new(int16)
		case reflect.Int32:
			values[i] = new(int32)
		case reflect.Int64:
			values[i] = new(int64)
		case reflect.Uint:
			values[i] = new(uint)
		case reflect.Uint8:
			values[i] = new(uint8)
		case reflect.Uint16:
			values[i] = new(uint16)
		case reflect.Uint32:
			values[i] = new(uint32)
		case reflect.Uint64:
			values[i] = new(uint64)
		case reflect.Uintptr:
			values[i] = new(uintptr)
		case reflect.Float32:
			values[i] = new(float32)
		case reflect.Float64:
			values[i] = new(float64)
		case reflect.Complex64:
			values[i] = new(complex64)
		case reflect.Complex128:
			values[i] = new(complex128)
		case reflect.Array:
			values[i] = new([]any)
		case reflect.Interface:
			values[i] = new(any)
		case reflect.Map:
			values[i] = new(map[string]any)
		case reflect.Slice:
			values[i] = new([]any)
		case reflect.String:
			values[i] = new(string)
		case reflect.Struct:
			switch scanType.String() {
			case "sql.NullString":
				values[i] = new(string)
			case "sql.NullInt64":
				values[i] = new(int64)
			case "sql.NullInt32":
				values[i] = new(int32)
			case "sql.NullInt16":
				values[i] = new(int16)
			case "sql.NullByte":
				values[i] = new(byte)
			case "sql.NullFloat64":
				values[i] = new(float64)
			case "sql.NullBool":
				values[i] = new(bool)
			case "sql.NullTime":
				values[i] = new(sql.NullTime)
			default:
				values[i] = new(any)
			}
		default:
			values[i] = new(any)
		}
	}

	return values
}

// scanValues create a slice of scan values for the given columns.
func scanValues(s *schema.Schema, columns []string) ([]any, error) {
	values := make([]any, len(columns))
	for i := range columns {
		field := s.Field(columns[i])
		if field == nil { // No field found. Ignore it.
			values[i] = new(any)
			continue
		}
		switch field.Type {
		case schema.TypeJSON, schema.TypeBytes:
			values[i] = new([]byte)
		case schema.TypeBool:
			values[i] = new(sql.NullBool)
		case schema.TypeFloat32, schema.TypeFloat64:
			values[i] = new(sql.NullFloat64)
		case schema.TypeInt8, schema.TypeInt16, schema.TypeInt32, schema.TypeInt, schema.TypeInt64, schema.TypeUint8, schema.TypeUint16, schema.TypeUint32, schema.TypeUint, schema.TypeUint64:
			values[i] = new(sql.NullInt64)
		case schema.TypeEnum, schema.TypeString, schema.TypeText:
			values[i] = new(sql.NullString)
		case schema.TypeTime:
			values[i] = new(sql.NullTime)
		case schema.TypeUUID:
			values[i] = new(uuid.UUID)
		default:
			return nil, fmt.Errorf("unexpected column %q for schema %s", columns[i], s.Name)
		}
	}
	return values, nil
}

// assignValues assigns the given values to the entity.
func assignValues(s *schema.Schema, entity *schema.Entity, columns []string, values []any) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		field := s.Field(columns[i])
		if field == nil { // No field found. Ignore it.
			entity.Set(columns[i], new(any))
			continue
		}
		switch field.Type {
		case schema.TypeBool:
			if value, ok := values[i].(*sql.NullBool); !ok {
				return fieldTypeError("Bool", values[i])
			} else if value.Valid {
				entity.Set(field.Name, value.Bool)
			}
		case schema.TypeTime:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fieldTypeError("Time", values[i])
			} else if value.Valid {
				entity.Set(field.Name, value.Time)
			}
		case schema.TypeJSON:
			if value, ok := values[i].(*[]byte); !ok {
				return fieldTypeError("JSON", values[i])
			} else if value != nil && len(*value) > 0 {
				e := entity.Get(field.Name)
				if err := json.Unmarshal(*value, &e); err != nil {
					return fmt.Errorf("unmarshal field field_type_JSON: %w", err)
				}
				entity.Set(field.Name, e)
			}
		case schema.TypeUUID:
			if value, ok := values[i].(*uuid.UUID); !ok {
				return fieldTypeError("UUID", values[i])
			} else if value != nil {
				entity.Set(field.Name, *value)
			}
		case schema.TypeBytes:
			if value, ok := values[i].(*[]byte); !ok {
				return fieldTypeError("Bytes", values[i])
			} else if value != nil {
				entity.Set(field.Name, *value)
			}
		case schema.TypeEnum:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fieldTypeError("Enum", values[i])
			} else if value.Valid {
				entity.Set(field.Name, value.String)
			}
		case schema.TypeString:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fieldTypeError("String", values[i])
			} else if value.Valid {
				entity.Set(field.Name, value.String)
			}
		case schema.TypeText:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fieldTypeError("Text", values[i])
			} else if value.Valid {
				entity.Set(field.Name, value.String)
			}
		case schema.TypeInt8:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fieldTypeError("Int8", values[i])
			} else if value.Valid {
				entity.Set(field.Name, int8(value.Int64))
			}
		case schema.TypeInt16:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fieldTypeError("Int16", values[i])
			} else if value.Valid {
				entity.Set(field.Name, int16(value.Int64))
			}
		case schema.TypeInt32:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fieldTypeError("Int32", values[i])
			} else if value.Valid {
				entity.Set(field.Name, int32(value.Int64))
			}
		case schema.TypeInt:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fieldTypeError("Int", values[i])
			} else if value.Valid {
				entity.Set(field.Name, int(value.Int64))
			}
		case schema.TypeInt64:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fieldTypeError("Int64", values[i])
			} else if value.Valid {
				entity.Set(field.Name, value.Int64)
			}
		case schema.TypeUint8:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fieldTypeError("Uint8", values[i])
			} else if value.Valid {
				entity.Set(field.Name, uint8(value.Int64))
			}
		case schema.TypeUint16:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fieldTypeError("Uint16", values[i])
			} else if value.Valid {
				entity.Set(field.Name, uint16(value.Int64))
			}
		case schema.TypeUint32:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fieldTypeError("Uint32", values[i])
			} else if value.Valid {
				entity.Set(field.Name, uint32(value.Int64))
			}
		case schema.TypeUint:
			value, ok := values[i].(*sql.NullInt64)
			if !ok {
				return fieldTypeError("Uint", values[i])
			}

			entity.Set(field.Name, utils.If(value.Valid, uint(value.Int64), uint(0)))
		case schema.TypeUint64:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fieldTypeError("Uint64", values[i])
			} else if value.Valid {
				entity.Set(field.Name, uint64(value.Int64))
			}
		case schema.TypeFloat32:
			if value, ok := values[i].(*sql.NullFloat64); !ok {
				return fieldTypeError("Float32", values[i])
			} else if value.Valid {
				entity.Set(field.Name, float32(value.Float64))
			}
		case schema.TypeFloat64:
			if value, ok := values[i].(*sql.NullFloat64); !ok {
				return fieldTypeError("Float64", values[i])
			} else if value.Valid {
				entity.Set(field.Name, value.Float64)
			}
		}
	}

	return nil
}

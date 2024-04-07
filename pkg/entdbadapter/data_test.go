package entdbadapter

import (
	"os"
	"testing"

	"entgo.io/ent/dialect"
	dialectSql "entgo.io/ent/dialect/sql"
	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/fastschema/fastschema/app"
	"github.com/fastschema/fastschema/pkg/testutils"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
	"github.com/stretchr/testify/assert"
)

func createSchemaBuilder() *schema.Builder {
	return testutils.CreateSchemaBuilder("../../tests/data/schemas")
}

func createMockAdapter(t *testing.T) *Adapter {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)

	tmpDir, err := os.MkdirTemp("", "migrations")
	assert.NoError(t, err)

	sb := createSchemaBuilder()
	client := utils.Must(NewEntClient(&app.DBConfig{
		Driver:       "sqlmock",
		MigrationDir: tmpDir,
	}, sb, dialectSql.OpenDB(dialect.MySQL, db)))

	adapter, ok := client.(*Adapter)
	assert.True(t, ok)

	return adapter
}

var testUserSchemaJSON = `{
	"name": "user",
	"namespace": "users",
	"label_field": "name",
	"fields": [
		{
			"name": "name",
			"label": "Name",
			"type": "string",
			"sortable": true
		},
		{
			"name": "json_field",
			"label": "JSON Field",
			"type": "json"
		},
		{
			"name": "bytes_field",
			"label": "Bytes Field",
			"type": "bytes"
		},
		{
			"name": "bool_field",
			"label": "Bool Field",
			"type": "bool"
		},
		{
			"name": "float32_field",
			"label": "Float32 Field",
			"type": "float32"
		},
		{
			"name": "float64_field",
			"label": "Float64 Field",
			"type": "float64"
		},
		{
			"name": "int8_field",
			"label": "int8 Field",
			"type": "int8"
		},
		{
			"name": "int16_field",
			"label": "int16 Field",
			"type": "int16"
		},
		{
			"name": "int32_field",
			"label": "int32 Field",
			"type": "int32"
		},
		{
			"name": "int_field",
			"label": "int Field",
			"type": "int"
		},
		{
			"name": "int64_field",
			"label": "int64 Field",
			"type": "int64"
		},
		{
			"name": "uint8_field",
			"label": "uint8 Field",
			"type": "uint8"
		},
		{
			"label": "uint16 Field",
			"name": "uint16_field",
			"type": "uint16"
		},
		{
			"label": "uint32 Field",
			"name": "uint32_field",
			"type": "uint32"
		},
		{
			"label": "uint Field",
			"name": "uint_field",
			"type": "uint"
		},
		{
			"label": "uint64 Field",
			"name": "uint64_field",
			"type": "uint64"
		},
		{
			"label": "time Field",
			"name": "time_field",
			"type": "time"
		},
		{
			"label": "uuid Field",
			"name": "uuid_field",
			"type": "uuid"
		},
		{
			"label": "enum Field",
			"name": "enum_field",
			"type": "enum",
			"enums": [
				{
					"label": "Enum1",
					"value": "enum1"
				},
				{
					"label": "Enum2",
					"value": "enum2"
				}
			]
		},
		{
			"label": "string Field",
			"name": "string_field",
			"type": "string"
		},
		{
			"label": "text Field",
			"name": "text_field",
			"type": "text"
		}
	]
}`

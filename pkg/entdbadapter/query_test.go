package entdbadapter

import (
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"entgo.io/ent/dialect"
	dialectSql "entgo.io/ent/dialect/sql"
	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/fastschema/fastschema/app"
	"github.com/fastschema/fastschema/pkg/testutils"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

var sbq = testutils.CreateSchemaBuilder("../../tests/data/schemas")
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

func TestScanValues(t *testing.T) {
	userSchema := &schema.Schema{}
	assert.Nil(t, json.Unmarshal([]byte(testUserSchemaJSON), userSchema))
	assert.NoError(t, userSchema.Init(false))

	// _, err := ScanValues(userSchema, []string{"invalid"})
	// assert.Error(t, err)

	results := utils.Must(scanValues(userSchema, []string{
		"json_field",
		"bytes_field",
		"bool_field",
		"float32_field",
		"float64_field",
		"int8_field",
		"int16_field",
		"int32_field",
		"int_field",
		"int64_field",
		"uint8_field",
		"uint16_field",
		"uint32_field",
		"uint_field",
		"uint64_field",
		"time_field",
		"uuid_field",
		"enum_field",
		"string_field",
		"text_field",
	}))

	assert.Equal(t, []any{
		new([]byte),
		new([]byte),
		new(sql.NullBool),
		new(sql.NullFloat64),
		new(sql.NullFloat64),

		new(sql.NullInt64),
		new(sql.NullInt64),
		new(sql.NullInt64),
		new(sql.NullInt64),
		new(sql.NullInt64),
		new(sql.NullInt64),
		new(sql.NullInt64),
		new(sql.NullInt64),
		new(sql.NullInt64),
		new(sql.NullInt64),

		new(sql.NullTime),
		new(uuid.UUID),

		new(sql.NullString),
		new(sql.NullString),
		new(sql.NullString),
	}, results)
}

func TestAssignValues(t *testing.T) {
	userSchema := &schema.Schema{}
	assert.Nil(t, json.Unmarshal([]byte(testUserSchemaJSON), userSchema))
	assert.NoError(t, userSchema.Init(false))
	entity := schema.NewEntity(1)

	err := assignValues(userSchema, entity, []string{"id", "name"}, []any{1})
	assert.Equal(t, "mismatch number of scan values: 1 != 2", err.Error())

	// err = AssignValues(userSchema, entity, []string{"invalid"}, []any{1})
	// assert.Equal(t, "field user.invalid not found", err.Error())

	type args struct {
		column      string
		value       any
		expectError string
		expectValue any
	}

	now := time.Now()
	jsonValue := []byte(`{"a": 1}`)
	jsonValueError := []byte(`{"a": 1, "b"}`)
	byteValue := []byte("hello")
	uuidValue := uuid.New()
	tests := []args{
		{
			column:      "bool_field",
			value:       &sql.NullBool{Bool: true, Valid: true},
			expectValue: true,
		},
		{
			column:      "bool_field",
			value:       true,
			expectError: "unexpected type bool for field type Bool",
		},
		{
			column:      "time_field",
			value:       &sql.NullTime{Time: now, Valid: true},
			expectValue: now,
		},
		{
			column:      "time_field",
			value:       1,
			expectError: "unexpected type int for field type Time",
		},
		{
			column: "json_field",
			value:  &jsonValue,
			expectValue: map[string]any{
				"a": float64(1),
			},
		},
		{
			column:      "json_field",
			value:       &jsonValueError,
			expectError: "unmarshal field field_type_JSON: invalid character '}' after object key",
		},
		{
			column:      "json_field",
			value:       1,
			expectError: "unexpected type int for field type JSON",
		},
		{
			column:      "uuid_field",
			value:       &uuidValue,
			expectValue: uuidValue,
		},
		{
			column:      "uuid_field",
			value:       1,
			expectError: "unexpected type int for field type UUID",
		},
		{
			column:      "bytes_field",
			value:       &byteValue,
			expectValue: byteValue,
		},
		{
			column:      "bytes_field",
			value:       1,
			expectError: "unexpected type int for field type Bytes",
		},
		{
			column:      "enum_field",
			value:       &sql.NullString{String: "hello", Valid: true},
			expectValue: "hello",
		},
		{
			column:      "enum_field",
			value:       "hello",
			expectError: "unexpected type string for field type Enum",
		},
		{
			column:      "string_field",
			value:       &sql.NullString{String: "hello", Valid: true},
			expectValue: "hello",
		},
		{
			column:      "string_field",
			value:       "hello",
			expectError: "unexpected type string for field type String",
		},
		{
			column:      "text_field",
			value:       &sql.NullString{String: "hello", Valid: true},
			expectValue: "hello",
		},
		{
			column:      "text_field",
			value:       "hello",
			expectError: "unexpected type string for field type Text",
		},
		{
			column:      "int8_field",
			value:       &sql.NullInt64{Int64: 1, Valid: true},
			expectValue: int8(1),
		},
		{
			column:      "int8_field",
			value:       1,
			expectError: "unexpected type int for field type Int8",
		},
		{
			column:      "int16_field",
			value:       &sql.NullInt64{Int64: 1, Valid: true},
			expectValue: int16(1),
		},
		{
			column:      "int16_field",
			value:       1,
			expectError: "unexpected type int for field type Int16",
		},
		{
			column:      "int32_field",
			value:       &sql.NullInt64{Int64: 1, Valid: true},
			expectValue: int32(1),
		},
		{
			column:      "int32_field",
			value:       1,
			expectError: "unexpected type int for field type Int32",
		},
		{
			column:      "int_field",
			value:       &sql.NullInt64{Int64: 1, Valid: true},
			expectValue: int(1),
		},
		{
			column:      "int_field",
			value:       1,
			expectError: "unexpected type int for field type Int",
		},
		{
			column:      "int64_field",
			value:       &sql.NullInt64{Int64: 1, Valid: true},
			expectValue: int64(1),
		},
		{
			column:      "int64_field",
			value:       1,
			expectError: "unexpected type int for field type Int64",
		},
		{
			column:      "uint8_field",
			value:       &sql.NullInt64{Int64: 1, Valid: true},
			expectValue: uint8(1),
		},
		{
			column:      "uint8_field",
			value:       1,
			expectError: "unexpected type int for field type Uint8",
		},
		{
			column:      "uint16_field",
			value:       &sql.NullInt64{Int64: 1, Valid: true},
			expectValue: uint16(1),
		},
		{
			column:      "uint16_field",
			value:       1,
			expectError: "unexpected type int for field type Uint16",
		},
		{
			column:      "uint32_field",
			value:       &sql.NullInt64{Int64: 1, Valid: true},
			expectValue: uint32(1),
		},
		{
			column:      "uint32_field",
			value:       1,
			expectError: "unexpected type int for field type Uint32",
		},
		{
			column:      "uint_field",
			value:       &sql.NullInt64{Int64: 1, Valid: true},
			expectValue: uint(1),
		},
		{
			column:      "uint_field",
			value:       1,
			expectError: "unexpected type int for field type Uint",
		},
		{
			column:      "uint64_field",
			value:       &sql.NullInt64{Int64: 1, Valid: true},
			expectValue: uint64(1),
		},
		{
			column:      "uint64_field",
			value:       1,
			expectError: "unexpected type int for field type Uint64",
		},
		{
			column:      "float32_field",
			value:       &sql.NullFloat64{Float64: 1, Valid: true},
			expectValue: float32(1),
		},
		{
			column:      "float32_field",
			value:       1,
			expectError: "unexpected type int for field type Float32",
		},
		{
			column:      "float64_field",
			value:       &sql.NullFloat64{Float64: 1, Valid: true},
			expectValue: float64(1),
		},
		{
			column:      "float64_field",
			value:       1,
			expectError: "unexpected type int for field type Float64",
		},
	}

	for _, tt := range tests {
		err := assignValues(userSchema, entity, []string{tt.column}, []any{tt.value})
		if tt.expectError != "" {
			assert.Equal(t, tt.expectError, err.Error())
		} else {
			assert.NoError(t, err)
			assert.Equal(t, tt.expectValue, entity.Get(tt.column))
		}
	}
}

func TestCount(t *testing.T) {
	tests := []testutils.MockTestCountData{
		{
			Name:   "Count_with_no_filter",
			Schema: "user",
			Expect: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(utils.EscapeQuery("SELECT COUNT(`users`.`id`) FROM `users`")).
					WillReturnRows(mock.NewRows([]string{"count"}).AddRow(2))
			},
			ExpectCount: 2,
		},
		{
			Name:   "Count_with_filter",
			Schema: "user",
			Filter: `{
				"id": {
					"$gt": 1
				}
			}`,
			Expect: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(utils.EscapeQuery("SELECT COUNT(`users`.`id`) FROM `users` WHERE `id` > ?")).
					WithArgs(float64(1)).
					WillReturnRows(mock.NewRows([]string{"count"}).AddRow(11))
			},
			ExpectCount: 11,
		},
		{
			Name:   "Count_with_columns",
			Schema: "user",
			Filter: `{
				"id": {
					"$gt": 1
				}
			}`,
			Column: "name",
			Expect: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(utils.EscapeQuery("SELECT COUNT(`users`.`name`) FROM `users` WHERE `id` > ?")).
					WithArgs(float64(1)).
					WillReturnRows(mock.NewRows([]string{"count"}).AddRow(11))
			},
			ExpectCount: 11,
		},
		{
			Name:   "Count_with_unique",
			Schema: "user",
			Filter: `{
				"id": {
					"$gt": 1
				}
			}`,
			Unique: true,
			Expect: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(utils.EscapeQuery("SELECT COUNT(DISTINCT `users`.`id`) FROM `users` WHERE `id` > ?")).
					WithArgs(float64(1)).
					WillReturnRows(mock.NewRows([]string{"count"}).AddRow(11))
			},
			ExpectCount: 11,
		},
		{
			Name:   "Count_with_column_and_unique",
			Schema: "user",
			Filter: `{
				"id": {
					"$gt": 1
				}
			}`,
			Column: "status",
			Unique: true,
			Expect: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(utils.EscapeQuery("SELECT COUNT(DISTINCT `users`.`status`) FROM `users` WHERE `id` > ?")).
					WithArgs(float64(1)).
					WillReturnRows(mock.NewRows([]string{"count"}).AddRow(11))
			},
			ExpectCount: 11,
		},
	}

	testutils.MockRunCountTests(func(d *sql.DB) app.DBClient {
		client := utils.Must(NewEntClient(&app.DBConfig{
			Driver: "sqlmock",
		}, sbq, dialectSql.OpenDB(dialect.MySQL, d)))
		return client
	}, sbq, t, tests)
}

func TestQuery(t *testing.T) {
	tests := []testutils.MockTestQueryData{
		{
			Name:   "Query_with_no_filter",
			Schema: "user",
			Expect: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(utils.EscapeQuery("SELECT * FROM `users`")).
					WillReturnRows(mock.NewRows([]string{"id", "name"}).
						AddRow(1, "John").
						AddRow(2, "Doe"))
			},
			ExpectEntities: []*schema.Entity{
				schema.NewEntity(1).Set("name", "John"),
				schema.NewEntity(2).Set("name", "Doe"),
			},
		},
		{
			Name:   "Query_with_filter",
			Schema: "user",
			Filter: `{
				"age": {
					"$gt": 5
				}
			}`,
			Expect: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(utils.EscapeQuery("SELECT * FROM `users` WHERE `age` > ?")).
					WithArgs(float64(5)).
					WillReturnRows(mock.NewRows([]string{"id", "name"}).
						AddRow(1, "John"))
			},
			ExpectEntities: []*schema.Entity{
				schema.NewEntity(1).Set("name", "John"),
			},
		},
		{
			Name:   "Query_with_limit_offset_and_order",
			Schema: "car",
			Filter: `{
				"name": {
					"$like": "%car%"
				}
			}`,
			Limit:  10,
			Offset: 20,
			Order:  []string{"-id", "name"},
			Expect: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(utils.EscapeQuery("SELECT * FROM `cars` WHERE `name` LIKE ? ORDER BY `id` DESC, `name` ASC LIMIT 10 OFFSET 20")).
					WithArgs("%car%").
					WillReturnRows(mock.NewRows([]string{"id", "name"}).
						AddRow(1, "car1"))
			},
			ExpectEntities: []*schema.Entity{
				schema.NewEntity(1).Set("name", "car1"),
			},
		},
		{
			Name:   "Query_with_invalid_order",
			Schema: "car",
			Filter: `{
				"name": {
					"$like": "%car%"
				}
			}`,
			Limit:       10,
			Offset:      20,
			Order:       []string{"-invalid"},
			ExpectError: "column car.invalid not found",
		},
		{
			Name:   "Query_with_columns",
			Schema: "user",
			Columns: []string{
				"id",
				"name",
			},
			Expect: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(utils.EscapeQuery("SELECT `users`.`id`, `users`.`name` FROM `users`")).
					WillReturnRows(mock.NewRows([]string{"id", "name"}).
						AddRow(1, "John"))
			},
			ExpectEntities: []*schema.Entity{
				schema.NewEntity(1).Set("name", "John"),
			},
		},
		{
			Name:   "Query_with_invalid_columns",
			Schema: "user",
			Columns: []string{
				"id",
				"invalid",
			},
			Expect: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(utils.EscapeQuery("SELECT `users`.`id`, `users`.`name` FROM `users`")).
					WillReturnRows(mock.NewRows([]string{"id", "name"}).
						AddRow(1, "John"))
			},
			ExpectError: "column user.invalid not found",
		},
		{
			Name:   "Query_with_relation_filter",
			Schema: "car",
			Filter: `{
				"name": {
					"$like": "%car%"
				},
				"owner.groups.name": {
					"$like": "%admin%"
				}
			}`,
			Expect: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(utils.EscapeQuery("SELECT * FROM `cars` WHERE `name` LIKE ? AND `cars`.`owner_id` IN (SELECT `users`.`id` FROM `users` WHERE `users`.`id` IN (SELECT `groups_users`.`users` FROM `groups_users` JOIN `groups` AS `t1` ON `groups_users`.`groups` = `t1`.`id` WHERE `name` LIKE ?))")).
					WithArgs("%car%", "%admin%").
					WillReturnRows(mock.NewRows([]string{"id", "name"}).
						AddRow(1, "car1"))
			},
			ExpectEntities: []*schema.Entity{
				schema.NewEntity(1).Set("name", "car1"),
			},
		},
		{
			Name:   "Query_with_edges_O2M_two_types",
			Schema: "user",
			Columns: []string{
				"name",
				"pets",
			},
			Expect: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(utils.EscapeQuery("SELECT `users`.`id`, `users`.`name` FROM `users`")).
					WillReturnRows(mock.NewRows([]string{"id", "name"}).
						AddRow(1, "John"))
				mock.ExpectQuery(utils.EscapeQuery("SELECT * FROM `pets` WHERE `owner_id` IN (?)")).
					WillReturnRows(mock.NewRows([]string{"id", "name", "owner_id"}).
						AddRow(1, "Pet 1", uint64(1)).
						AddRow(2, "Pet 2", uint64(1)).
						AddRow(3, "Pet 3", uint64(1)))
			},
			ExpectEntities: []*schema.Entity{
				schema.NewEntity(1).Set("name", "John").Set("pets", []*schema.Entity{
					schema.NewEntity(1).Set("name", "Pet 1").Set("owner_id", uint64(1)),
					schema.NewEntity(2).Set("name", "Pet 2").Set("owner_id", uint64(1)),
					schema.NewEntity(3).Set("name", "Pet 3").Set("owner_id", uint64(1)),
				}),
			},
		},
		{
			Name:   "Query_with_edges_O2M_two_types_reverse",
			Schema: "pet",
			Columns: []string{
				"id",
				"name",
				"owner",
			},
			Expect: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(utils.EscapeQuery("SELECT `pets`.`id`, `pets`.`name`, `pets`.`owner_id` FROM `pets`")).
					WillReturnRows(mock.NewRows([]string{"id", "name", "owner_id"}).
						AddRow(1, "Pet 1", uint64(1)).
						AddRow(2, "Pet 2", uint64(1)).
						AddRow(3, "Pet 3", uint64(2)))
				mock.ExpectQuery(utils.EscapeQuery("SELECT * FROM `users` WHERE `id` IN (?, ?)")).
					WillReturnRows(mock.NewRows([]string{"id", "name"}).
						AddRow(1, "John").
						AddRow(2, "Jane"))
			},
			ExpectEntities: []*schema.Entity{
				schema.NewEntity(1).
					Set("name", "Pet 1").
					Set("owner_id", uint64(1)).
					Set("owner", schema.NewEntity(1).
						Set("name", "John")),
				schema.NewEntity(2).
					Set("name", "Pet 2").
					Set("owner_id", uint64(1)).
					Set("owner", schema.NewEntity(1).
						Set("name", "John")),
				schema.NewEntity(3).
					Set("name", "Pet 3").
					Set("owner_id", uint64(2)).
					Set("owner", schema.NewEntity(2).
						Set("name", "Jane")),
			},
		},
		{
			Name:    "Query_with_edges_O2M_same_type",
			Schema:  "node",
			Columns: []string{"name", "children"},
			Expect: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(utils.EscapeQuery("SELECT `nodes`.`id`, `nodes`.`name` FROM `nodes`")).
					WillReturnRows(mock.NewRows([]string{"id", "name"}).
						AddRow(1, "Node 1"))
				mock.ExpectQuery(utils.EscapeQuery("SELECT * FROM `nodes` WHERE `parent_id` IN (?)")).
					WillReturnRows(mock.NewRows([]string{"id", "name", "parent_id"}).
						AddRow(2, "Node 2", uint64(1)).
						AddRow(3, "Node 3", uint64(1)))
			},
			ExpectEntities: []*schema.Entity{
				schema.NewEntity(1).Set("name", "Node 1").Set("children", []*schema.Entity{
					schema.NewEntity(2).Set("name", "Node 2").Set("parent_id", uint64(1)),
					schema.NewEntity(3).Set("name", "Node 3").Set("parent_id", uint64(1)),
				}),
			},
		},
		{
			Name:    "Query_with_edges_O2M_same_type_reverse",
			Schema:  "node",
			Columns: []string{"name", "parent"},
			Expect: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(utils.EscapeQuery("SELECT `nodes`.`id`, `nodes`.`name`, `nodes`.`parent_id` FROM `nodes`")).
					WillReturnRows(mock.NewRows([]string{"id", "name", "parent_id"}).
						AddRow(3, "Node 3", 1).
						AddRow(4, "Node 4", 2))
				mock.ExpectQuery(utils.EscapeQuery("SELECT * FROM `nodes` WHERE `id` IN (?, ?)")).
					WillReturnRows(mock.NewRows([]string{"id", "name"}).
						AddRow(1, "Node 1").
						AddRow(2, "Node 2"))
			},
			ExpectEntities: []*schema.Entity{
				schema.NewEntity(3).Set("name", "Node 3").Set("parent_id", 1).Set("parent", schema.NewEntity(1).Set("name", "Node 1")),
				schema.NewEntity(4).Set("name", "Node 4").Set("parent_id", 2).Set("parent", schema.NewEntity(2).Set("name", "Node 2")),
			},
		},
		{
			Name:    "Query_with_edges_O2O_two_types",
			Schema:  "user",
			Columns: []string{"name", "card"},
			Expect: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(utils.EscapeQuery("SELECT `users`.`id`, `users`.`name` FROM `users`")).
					WillReturnRows(mock.NewRows([]string{"id", "name"}).
						AddRow(1, "John").
						AddRow(2, "Jane"))
				mock.ExpectQuery(utils.EscapeQuery("SELECT * FROM `cards` WHERE `owner_id` IN (?, ?)")).
					WithArgs(1, 2).
					WillReturnRows(mock.NewRows([]string{"id", "number", "owner_id"}).
						AddRow(1, "1234", 1))
			},
			ExpectEntities: []*schema.Entity{
				schema.NewEntity(1).Set("name", "John").Set("card", schema.NewEntity(1).Set("number", "1234").Set("owner_id", 1)),
				schema.NewEntity(2).Set("name", "Jane"),
			},
		},
		{
			Name:    "Query_with_edges_O2O_two_types_reverse",
			Schema:  "card",
			Columns: []string{"number", "owner"},
			Expect: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(utils.EscapeQuery("SELECT `cards`.`id`, `cards`.`number`, `cards`.`owner_id` FROM `cards`")).
					WillReturnRows(mock.NewRows([]string{"id", "number", "owner_id"}).
						AddRow(1, "1234", 1).
						AddRow(2, "5678", 2))
				mock.ExpectQuery(utils.EscapeQuery("SELECT * FROM `users` WHERE `id` IN (?, ?)")).
					WillReturnRows(mock.NewRows([]string{"id", "name"}).
						AddRow(1, "John"))
			},
			ExpectEntities: []*schema.Entity{
				schema.NewEntity(1).Set("number", "1234").Set("owner_id", 1).Set("owner", schema.NewEntity(1).Set("name", "John")),
				schema.NewEntity(2).Set("number", "5678").Set("owner_id", 2),
			},
		},
		{
			Name:    "Query_with_edges_O2O_same_type",
			Schema:  "node",
			Columns: []string{"name", "next"},
			Expect: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(utils.EscapeQuery("SELECT `nodes`.`id`, `nodes`.`name` FROM `nodes`")).
					WillReturnRows(mock.NewRows([]string{"id", "name"}).
						AddRow(1, "Node 1").
						AddRow(2, "Node 2"))
				mock.ExpectQuery(utils.EscapeQuery("SELECT * FROM `nodes` WHERE `prev_id` IN (?, ?)")).
					WillReturnRows(mock.NewRows([]string{"id", "name", "prev_id"}).
						AddRow(2, "Node 2", 1))
			},
			ExpectEntities: []*schema.Entity{
				schema.NewEntity(1).Set("name", "Node 1").Set("next", schema.NewEntity(2).Set("name", "Node 2").Set("prev_id", 1)),
				schema.NewEntity(2).Set("name", "Node 2"),
			},
		},
		{
			Name:    "Query_with_edges_O2O_same_type_reverse",
			Schema:  "node",
			Columns: []string{"name", "prev"},
			Expect: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(utils.EscapeQuery("SELECT `nodes`.`id`, `nodes`.`name`, `nodes`.`prev_id` FROM `nodes`")).
					WillReturnRows(mock.NewRows([]string{"id", "name", "prev_id"}).
						AddRow(1, "Node 1", nil).
						AddRow(2, "Node 2", 1))
				mock.ExpectQuery(utils.EscapeQuery("SELECT * FROM `nodes` WHERE `id` IN (?)")).
					WithArgs(1).
					WillReturnRows(mock.NewRows([]string{"id", "name"}).
						AddRow(1, "Node 1"))
			},
			ExpectEntities: []*schema.Entity{
				schema.NewEntity(1).Set("name", "Node 1"),
				schema.NewEntity(2).Set("name", "Node 2").Set("prev_id", 1).Set("prev", schema.NewEntity(1).Set("name", "Node 1")),
			},
		},
		{
			Name:    "Query_with_edges_O2O_bidi",
			Schema:  "user",
			Columns: []string{"name", "spouse"},
			Expect: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(utils.EscapeQuery("SELECT `users`.`id`, `users`.`name`, `users`.`spouse_id` FROM `users`")).
					WillReturnRows(mock.NewRows([]string{"id", "name", "spouse_id"}).
						AddRow(1, "John", 2).
						AddRow(2, "Jane", 1))
				mock.ExpectQuery(utils.EscapeQuery("SELECT * FROM `users` WHERE `id` IN (?, ?)")).
					WithArgs(2, 1).
					WillReturnRows(mock.NewRows([]string{"id", "name", "spouse_id"}).
						AddRow(2, "Jane", 1).
						AddRow(1, "John", 2))
			},
			ExpectEntities: []*schema.Entity{
				schema.NewEntity(1).Set("name", "John").Set("spouse_id", 2).Set("spouse", schema.NewEntity(2).Set("name", "Jane").Set("spouse_id", 1)),
				schema.NewEntity(2).Set("name", "Jane").Set("spouse_id", 1).Set("spouse", schema.NewEntity(1).Set("name", "John").Set("spouse_id", 2)),
			},
		},
		{
			Name:    "Query_with_edges_M2M_two_types",
			Schema:  "group",
			Columns: []string{"name", "users"},
			Expect: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(utils.EscapeQuery("SELECT `groups`.`id`, `groups`.`name` FROM `groups`")).
					WillReturnRows(mock.NewRows([]string{"id", "name"}).
						AddRow(11, "Group 11").
						AddRow(22, "Group 22"))
				mock.ExpectQuery(utils.EscapeQuery("SELECT `t1`.`groups` AS groups_id, `users`.`id`, `users`.`username`, `users`.`email`, `users`.`password`, `users`.`active`, `users`.`provider`, `users`.`provider_id`, `users`.`provider_username`, `users`.`name`, `users`.`status`, `users`.`approved`, `users`.`bio`, `users`.`age`, `users`.`json`, `users`.`deleted`, `users`.`spouse_id`, `users`.`partner_id`, `users`.`workplace_id`, `users`.`room_id`, `users`.`parent_id`, `users`.`created_at`, `users`.`updated_at`, `users`.`deleted_at` FROM `users` JOIN `groups_users` AS `t1` ON `t1`.`users` = `users`.`id` WHERE `t1`.`groups` IN (?, ?) ORDER BY `id` ASC")).
					WithArgs(11, 22).
					WillReturnRows(mock.NewRows([]string{"groups_id", "id", "name"}).
						AddRow(11, 1, "John").
						AddRow(11, 2, "Jane").
						AddRow(22, 3, "Bob"))
			},
			ExpectEntities: []*schema.Entity{
				schema.NewEntity(11).Set("name", "Group 11").Set("users", []*schema.Entity{
					schema.NewEntity(1).Set("name", "John"),
					schema.NewEntity(2).Set("name", "Jane"),
				}),
				schema.NewEntity(22).Set("name", "Group 22").Set("users", []*schema.Entity{
					schema.NewEntity(3).Set("name", "Bob"),
				}),
			},
		},
		{
			Name:    "Query_with_edges_M2M_two_types_reverse",
			Schema:  "user",
			Columns: []string{"name", "groups"},
			Expect: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(utils.EscapeQuery("SELECT `users`.`id`, `users`.`name` FROM `users`")).
					WillReturnRows(mock.NewRows([]string{"id", "name"}).
						AddRow(1, "John").
						AddRow(2, "Jane").
						AddRow(3, "Bob"))
				mock.ExpectQuery(utils.EscapeQuery("SELECT `t1`.`users` AS users_id, `groups`.`id`, `groups`.`name`, `groups`.`created_at`, `groups`.`updated_at`, `groups`.`deleted_at` FROM `groups` JOIN `groups_users` AS `t1` ON `t1`.`groups` = `groups`.`id` WHERE `t1`.`users` IN (?, ?, ?) ORDER BY `id` ASC")).
					WithArgs(1, 2, 3).
					WillReturnRows(mock.NewRows([]string{"users_id", "id", "name"}).
						AddRow(1, 11, "Group 11").
						AddRow(1, 22, "Group 22").
						AddRow(2, 11, "Group 11"))
			},
			ExpectEntities: []*schema.Entity{
				schema.NewEntity(1).Set("name", "John").Set("groups", []*schema.Entity{
					schema.NewEntity(11).Set("name", "Group 11"),
					schema.NewEntity(22).Set("name", "Group 22"),
				}),
				schema.NewEntity(2).Set("name", "Jane").Set("groups", []*schema.Entity{
					schema.NewEntity(11).Set("name", "Group 11"),
				}),
				schema.NewEntity(3).Set("name", "Bob").Set("groups", []*schema.Entity{}),
			},
		},
		{
			Name:    "Query_with_edges_M2M_same_type",
			Schema:  "user",
			Columns: []string{"name", "following"},
			Expect: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(utils.EscapeQuery("SELECT `users`.`id`, `users`.`name` FROM `users`")).
					WillReturnRows(mock.NewRows([]string{"id", "name"}).
						AddRow(1, "John").
						AddRow(2, "Jane").
						AddRow(3, "Bob"))
				mock.ExpectQuery(utils.EscapeQuery("SELECT `t1`.`followers` AS followers_id, `users`.`id`, `users`.`username`, `users`.`email`, `users`.`password`, `users`.`active`, `users`.`provider`, `users`.`provider_id`, `users`.`provider_username`, `users`.`name`, `users`.`status`, `users`.`approved`, `users`.`bio`, `users`.`age`, `users`.`json`, `users`.`deleted`, `users`.`spouse_id`, `users`.`partner_id`, `users`.`workplace_id`, `users`.`room_id`, `users`.`parent_id`, `users`.`created_at`, `users`.`updated_at`, `users`.`deleted_at` FROM `users` JOIN `followers_following` AS `t1` ON `t1`.`following` = `users`.`id` WHERE `t1`.`followers` IN (?, ?, ?) ORDER BY `id` ASC")).
					WithArgs(1, 2, 3).
					WillReturnRows(mock.NewRows([]string{"followers_id", "id", "name"}).
						AddRow(1, 2, "Jane").
						AddRow(1, 3, "Bob").
						AddRow(2, 3, "Bob"))
			},
			ExpectEntities: []*schema.Entity{
				schema.NewEntity(1).Set("name", "John").Set("following", []*schema.Entity{
					schema.NewEntity(2).Set("name", "Jane"),
					schema.NewEntity(3).Set("name", "Bob"),
				}),
				schema.NewEntity(2).Set("name", "Jane").Set("following", []*schema.Entity{
					schema.NewEntity(3).Set("name", "Bob"),
				}),
				schema.NewEntity(3).Set("name", "Bob").Set("following", []*schema.Entity{}),
			},
		},
		{
			Name:    "Query_with_edges_M2M_same_type_reverse",
			Schema:  "user",
			Columns: []string{"name", "followers"},
			Expect: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(utils.EscapeQuery("SELECT `users`.`id`, `users`.`name` FROM `users`")).
					WillReturnRows(mock.NewRows([]string{"id", "name"}).
						AddRow(1, "John").
						AddRow(2, "Jane").
						AddRow(3, "Bob"))
				mock.ExpectQuery(utils.EscapeQuery("SELECT `t1`.`following` AS following_id, `users`.`id`, `users`.`username`, `users`.`email`, `users`.`password`, `users`.`active`, `users`.`provider`, `users`.`provider_id`, `users`.`provider_username`, `users`.`name`, `users`.`status`, `users`.`approved`, `users`.`bio`, `users`.`age`, `users`.`json`, `users`.`deleted`, `users`.`spouse_id`, `users`.`partner_id`, `users`.`workplace_id`, `users`.`room_id`, `users`.`parent_id`, `users`.`created_at`, `users`.`updated_at`, `users`.`deleted_at` FROM `users` JOIN `followers_following` AS `t1` ON `t1`.`followers` = `users`.`id` WHERE `t1`.`following` IN (?, ?, ?) ORDER BY `id` ASC")).
					WithArgs(1, 2, 3).
					WillReturnRows(mock.NewRows([]string{"following_id", "id", "name"}).
						AddRow(1, 2, "Jane").
						AddRow(1, 3, "Bob").
						AddRow(2, 3, "Bob"))
			},
			ExpectEntities: []*schema.Entity{
				schema.NewEntity(1).Set("name", "John").Set("followers", []*schema.Entity{
					schema.NewEntity(2).Set("name", "Jane"),
					schema.NewEntity(3).Set("name", "Bob"),
				}),
				schema.NewEntity(2).Set("name", "Jane").Set("followers", []*schema.Entity{
					schema.NewEntity(3).Set("name", "Bob"),
				}),
				schema.NewEntity(3).Set("name", "Bob").Set("followers", []*schema.Entity{}),
			},
		},
		{
			Name:    "Query_with_edges_M2M_bidi",
			Schema:  "user",
			Columns: []string{"name", "friends"},
			Expect: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(utils.EscapeQuery("SELECT `users`.`id`, `users`.`name` FROM `users`")).
					WillReturnRows(mock.NewRows([]string{"id", "name"}).
						AddRow(1, "John").
						AddRow(2, "Jane").
						AddRow(3, "Bob"))
				mock.ExpectQuery(utils.EscapeQuery("SELECT `t1`.`friends` AS friends_id, `users`.`id`, `users`.`username`, `users`.`email`, `users`.`password`, `users`.`active`, `users`.`provider`, `users`.`provider_id`, `users`.`provider_username`, `users`.`name`, `users`.`status`, `users`.`approved`, `users`.`bio`, `users`.`age`, `users`.`json`, `users`.`deleted`, `users`.`spouse_id`, `users`.`partner_id`, `users`.`workplace_id`, `users`.`room_id`, `users`.`parent_id`, `users`.`created_at`, `users`.`updated_at`, `users`.`deleted_at` FROM `users` JOIN `friends_user` AS `t1` ON `t1`.`user` = `users`.`id` WHERE `t1`.`user` IN (?, ?, ?) ORDER BY `id` ASC")).
					WithArgs(1, 2, 3).
					WillReturnRows(mock.NewRows([]string{"friends_id", "id", "name"}).
						AddRow(1, 2, "Jane").
						AddRow(1, 3, "Bob").
						AddRow(2, 3, "Bob"))
			},
			ExpectEntities: []*schema.Entity{
				schema.NewEntity(1).Set("name", "John").Set("friends", []*schema.Entity{
					schema.NewEntity(2).Set("name", "Jane"),
					schema.NewEntity(3).Set("name", "Bob"),
				}),
				schema.NewEntity(2).Set("name", "Jane").Set("friends", []*schema.Entity{
					schema.NewEntity(3).Set("name", "Bob"),
				}),
				schema.NewEntity(3).Set("name", "Bob").Set("friends", []*schema.Entity{}),
			},
		},
		{
			Name:    "Query_with_edges_O2O_fields",
			Schema:  "user",
			Columns: []string{"name", "card.number"},
			Expect: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(utils.EscapeQuery("SELECT `users`.`id`, `users`.`name` FROM `users`")).
					WillReturnRows(mock.NewRows([]string{"id", "name"}).
						AddRow(1, "John").
						AddRow(2, "Jane"))
				mock.ExpectQuery(utils.EscapeQuery("SELECT `cards`.`id`, `cards`.`number`, `cards`.`owner_id` FROM `cards` WHERE `owner_id` IN (?, ?)")).
					WithArgs(1, 2).
					WillReturnRows(mock.NewRows([]string{"id", "number", "owner_id"}).
						AddRow(1, "1234", 1))
			},
			ExpectEntities: []*schema.Entity{
				schema.NewEntity(1).Set("name", "John").Set("card", schema.NewEntity(1).Set("number", "1234").Set("owner_id", 1)),
				schema.NewEntity(2).Set("name", "Jane"),
			},
		},
		{
			Name:    "Query_with_edges_O2O_reverse_fields",
			Schema:  "card",
			Columns: []string{"number", "owner.name", "owner.age"},
			Expect: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(utils.EscapeQuery("SELECT `cards`.`id`, `cards`.`number`, `cards`.`owner_id` FROM `cards`")).
					WillReturnRows(mock.NewRows([]string{"id", "number", "owner_id"}).
						AddRow(1, "1234", 1).
						AddRow(2, "5678", 2))
				mock.ExpectQuery(utils.EscapeQuery("SELECT `users`.`id`, `users`.`name`, `users`.`age` FROM `users` WHERE `id` IN (?, ?)")).
					WillReturnRows(mock.NewRows([]string{"id", "name", "age"}).
						AddRow(1, "John", 8))
			},
			ExpectEntities: []*schema.Entity{
				schema.NewEntity(1).Set("number", "1234").Set("owner_id", 1).Set("owner", schema.NewEntity(1).Set("name", "John").Set("age", 8)),
				schema.NewEntity(2).Set("number", "5678").Set("owner_id", 2),
			},
		},
		{
			Name:   "Query_with_edges_O2M_fields",
			Schema: "user",
			Columns: []string{
				"name",
				"pets.name",
				"pets.created_at",
			},
			Expect: func(mock sqlmock.Sqlmock) {
				createdAt := utils.Must(time.Parse(time.RFC3339, "2006-01-02T15:04:05Z"))
				mock.ExpectQuery(utils.EscapeQuery("SELECT `users`.`id`, `users`.`name` FROM `users`")).
					WillReturnRows(mock.NewRows([]string{"id", "name"}).
						AddRow(1, "John"))
				mock.ExpectQuery(utils.EscapeQuery("SELECT `pets`.`id`, `pets`.`name`, `pets`.`created_at`, `pets`.`owner_id` FROM `pets` WHERE `owner_id` IN (?)")).
					WillReturnRows(mock.NewRows([]string{"id", "name", "created_at", "owner_id"}).
						AddRow(1, "Pet 1", createdAt, uint64(1)).
						AddRow(2, "Pet 2", createdAt, uint64(1)).
						AddRow(3, "Pet 3", createdAt, uint64(1)))
			},
			ExpectEntities: []*schema.Entity{
				schema.NewEntity(1).Set("name", "John").Set("pets", []*schema.Entity{
					schema.NewEntity(1).Set("name", "Pet 1").Set("created_at", utils.Must(time.Parse(time.RFC3339, "2006-01-02T15:04:05Z"))).Set("owner_id", uint64(1)),
					schema.NewEntity(2).Set("name", "Pet 2").Set("created_at", utils.Must(time.Parse(time.RFC3339, "2006-01-02T15:04:05Z"))).Set("owner_id", uint64(1)),
					schema.NewEntity(3).Set("name", "Pet 3").Set("created_at", utils.Must(time.Parse(time.RFC3339, "2006-01-02T15:04:05Z"))).Set("owner_id", uint64(1)),
				}),
			},
		},
		{
			Name:   "Query_with_edges_O2M_reverse_fields",
			Schema: "pet",
			Columns: []string{
				"id",
				"name",
				"owner.name",
				"owner.age",
			},
			Expect: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(utils.EscapeQuery("SELECT `pets`.`id`, `pets`.`name`, `pets`.`owner_id` FROM `pets`")).
					WillReturnRows(mock.NewRows([]string{"id", "name", "owner_id"}).
						AddRow(1, "Pet 1", uint64(1)).
						AddRow(2, "Pet 2", uint64(1)).
						AddRow(3, "Pet 3", uint64(2)))
				mock.ExpectQuery(utils.EscapeQuery("SELECT `users`.`id`, `users`.`name`, `users`.`age` FROM `users` WHERE `id` IN (?, ?)")).
					WillReturnRows(mock.NewRows([]string{"id", "name", "age"}).
						AddRow(1, "John", 5).
						AddRow(2, "Jane", 8))
			},
			ExpectEntities: []*schema.Entity{
				schema.NewEntity(1).Set("name", "Pet 1").Set("owner_id", 1).Set("owner", schema.NewEntity(1).Set("name", "John").Set("age", 5)),
				schema.NewEntity(2).Set("name", "Pet 2").Set("owner_id", 1).Set("owner", schema.NewEntity(1).Set("name", "John").Set("age", 5)),
				schema.NewEntity(3).Set("name", "Pet 3").Set("owner_id", 2).Set("owner", schema.NewEntity(2).Set("name", "Jane").Set("age", 8)),
			},
		},
		{
			Name:    "Query_with_edges_M2M_fields",
			Schema:  "group",
			Columns: []string{"name", "users.name", "users.age"},
			Expect: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(utils.EscapeQuery("SELECT `groups`.`id`, `groups`.`name` FROM `groups`")).
					WillReturnRows(mock.NewRows([]string{"id", "name"}).
						AddRow(11, "Group 11").
						AddRow(22, "Group 22"))
				mock.ExpectQuery(utils.EscapeQuery("SELECT `t1`.`groups` AS groups_id, `users`.`id`, `users`.`name`, `users`.`age` FROM `users` JOIN `groups_users` AS `t1` ON `t1`.`users` = `users`.`id` WHERE `t1`.`groups` IN (?, ?) ORDER BY `id` ASC")).
					WithArgs(11, 22).
					WillReturnRows(mock.NewRows([]string{"groups_id", "id", "name"}).
						AddRow(11, 1, "John").
						AddRow(11, 2, "Jane").
						AddRow(22, 3, "Bob"))
			},
			ExpectEntities: []*schema.Entity{
				schema.NewEntity(11).Set("name", "Group 11").Set("users", []*schema.Entity{
					schema.NewEntity(1).Set("name", "John"),
					schema.NewEntity(2).Set("name", "Jane"),
				}),
				schema.NewEntity(22).Set("name", "Group 22").Set("users", []*schema.Entity{
					schema.NewEntity(3).Set("name", "Bob"),
				}),
			},
		},
	}

	testutils.MockRunQueryTests(func(d *sql.DB) app.DBClient {
		client := utils.Must(NewEntClient(&app.DBConfig{
			Driver: "sqlmock",
		}, sbq, dialectSql.OpenDB(dialect.MySQL, d)))
		return client
	}, sbq, t, tests)
}

func TestFirstOnly(t *testing.T) {
	d, mock, err := sqlmock.New()
	assert.NoError(t, err)
	assert.NotNil(t, d)
	assert.NotNil(t, mock)
	client := utils.Must(NewEntClient(&app.DBConfig{
		Driver: "sqlmock",
	}, sbq, dialectSql.OpenDB(dialect.MySQL, d)))

	mock.ExpectQuery(utils.EscapeQuery("SELECT * FROM `users` WHERE `name` = ? LIMIT 1")).
		WithArgs("user1").
		WillReturnRows(mock.NewRows([]string{"id", "name"}).
			AddRow(1, "user1"))
	mock.ExpectQuery(utils.EscapeQuery("SELECT * FROM `users` WHERE `name` = ? LIMIT 1")).
		WithArgs("user2").
		WillReturnRows(mock.NewRows([]string{"id", "name"}))

	mock.ExpectQuery(utils.EscapeQuery("SELECT * FROM `users` WHERE `name` = ?")).
		WithArgs("user3").
		WillReturnRows(mock.NewRows([]string{"id", "name"}).
			AddRow(3, "user3"))

	mock.ExpectQuery(utils.EscapeQuery("SELECT * FROM `users` WHERE `name` = ?")).
		WithArgs("user4").
		WillReturnRows(mock.NewRows([]string{"id", "name"}).
			AddRow(4, "user4").
			AddRow(44, "user44"))

	mock.ExpectQuery(utils.EscapeQuery("SELECT * FROM `users` WHERE `name` = ?")).
		WithArgs("user5").
		WillReturnRows(mock.NewRows([]string{"id", "name"}))

	query1 := utils.Must(client.Model("user")).Query(app.EQ("name", "user1"))
	user1, err := query1.First()
	assert.NoError(t, err)
	assert.Equal(t, "user1", user1.Get("name"))

	query2 := utils.Must(client.Model("user")).Query(app.EQ("name", "user2"))
	_, err = query2.First()
	assert.Equal(t, "no entities found", err.Error())

	query3 := utils.Must(client.Model("user")).Query(app.EQ("name", "user3"))
	user3, err := query3.Only()
	assert.NoError(t, err)
	assert.Equal(t, "user3", user3.Get("name"))

	query4 := utils.Must(client.Model("user")).Query(app.EQ("name", "user4"))
	_, err = query4.Only()
	assert.Equal(t, "more than one entity found", err.Error())

	query5 := utils.Must(client.Model("user")).Query(app.EQ("name", "user5"))
	_, err = query5.Only()
	assert.Equal(t, "no entities found", err.Error())
}

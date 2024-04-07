package entdbadapter

import (
	"database/sql"
	"encoding/json"
	"testing"

	"entgo.io/ent/dialect"
	dialectSql "entgo.io/ent/dialect/sql"
	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/fastschema/fastschema/app"
	"github.com/fastschema/fastschema/pkg/testutils"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var userSchemaJSON = `{
  "name": "user",
  "namespace": "users",
  "label_field": "name",
  "fields": [
    {
      "name": "name",
      "label": "Name",
      "type": "string",
      "unique": true
    },
    {
      "name": "age",
      "label": "Age",
      "type": "uint"
    }
  ]
}`

func TestModel(t *testing.T) {
	userSchema := &schema.Schema{}
	assert.Nil(t, json.Unmarshal([]byte(userSchemaJSON), userSchema))

	// idField := &schema.Field{
	// 	Name: "id",
	// 	Type: schema.TypeUint64,
	// 	DB: &schema.FieldDB{
	// 		Increment: true,
	// 	},
	// }
	// idEntColumn := entdbadapter.CreateEntColumn(idField)
	// idColumn := &entdbadapter.Column{field: idField, entColumn: idEntColumn}

	// nameField := &schema.Field{Name: "name"}
	// nameEntColumn := entdbadapter.CreateEntColumn(nameField)
	// nameColumn := &entdbadapter.Column{field: nameField, entColumn: nameEntColumn}

	// model := &entdbadapter.Model{
	// 	name:        "user",
	// 	schema:      userSchema,
	// 	entIDColumn: idEntColumn,
	// 	columns:     []*entdbadapter.Column{idColumn, nameColumn},
	// }

	// assert.Equal(t, userSchema, model.Schema())
	// assert.Equal(t, nameColumn, utils.Must(model.Column("name")))

	// query := model.Query()
	// assert.NotNil(t, query)

	// mutation, err := model.Mutation()
	// assert.NoError(t, err)
	// assert.NotNil(t, mutation)
}

func TestModelName(t *testing.T) {
	model := &Model{name: "user"}
	assert.Equal(t, "user", model.Name())
}
func TestModelCreate(t *testing.T) {
	userSchema := &schema.Schema{}
	assert.Nil(t, json.Unmarshal([]byte(userSchemaJSON), userSchema))

	sb := createSchemaBuilder()
	createMockClient := func(d *sql.DB) app.DBClient {
		driver := utils.Must(NewEntClient(&app.DBConfig{
			Driver: "sqlmock",
		}, sb, dialectSql.OpenDB(dialect.MySQL, d)))
		return driver
	}

	client, err := testutils.NewMockClient(
		createMockClient,
		sb,
		func(m sqlmock.Sqlmock) {
			m.ExpectExec(utils.EscapeQuery("INSERT INTO `users` (`name`, `age`) VALUES (?, ?)")).
				WithArgs("John", float64(30)).
				WillReturnResult(sqlmock.NewResult(1, 1))
		},
		false,
	)
	require.NoError(t, err)
	model, err := client.Model("user")
	require.NoError(t, err)
	entity, err := schema.NewEntityFromJSON(`{"name": "John", "age": 30}`)
	assert.NoError(t, err)
	id, err := model.Create(entity)
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), id)
}

func TestModelCreateFromJson(t *testing.T) {
	userSchema := &schema.Schema{}
	assert.Nil(t, json.Unmarshal([]byte(userSchemaJSON), userSchema))

	sb := createSchemaBuilder()
	createMockClient := func(d *sql.DB) app.DBClient {
		driver := utils.Must(NewEntClient(&app.DBConfig{
			Driver: "sqlmock",
		}, sb, dialectSql.OpenDB(dialect.MySQL, d)))
		return driver
	}

	client, err := testutils.NewMockClient(
		createMockClient,
		sb,
		func(m sqlmock.Sqlmock) {
			m.ExpectExec(utils.EscapeQuery("INSERT INTO `users` (`name`, `age`) VALUES (?, ?)")).
				WithArgs("John", float64(30)).
				WillReturnResult(sqlmock.NewResult(1, 1))
		},
		false,
	)
	require.NoError(t, err)

	model, err := client.Model("user")
	require.NoError(t, err)
	id, err := model.CreateFromJSON(`{"name": "John", "age": 30}`)
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), id)
}

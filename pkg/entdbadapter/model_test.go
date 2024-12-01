package entdbadapter

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"

	"entgo.io/ent/dialect"
	dialectSql "entgo.io/ent/dialect/sql"
	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/entity"
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
  ],
	"db": {
		"indexes": [
			{
				"name": "age",
				"unique": false,
				"columns": ["age"]
			}
		]
	}
}`

func TestModelName(t *testing.T) {
	model := &Model{name: "user"}
	assert.Equal(t, "user", model.Name())
}
func TestModelCreate(t *testing.T) {
	userSchema := &schema.Schema{}
	assert.Nil(t, json.Unmarshal([]byte(userSchemaJSON), userSchema))

	sb := createSchemaBuilder()
	createMockClient := func(d *sql.DB) db.Client {
		driver := utils.Must(NewEntClient(&db.Config{
			Driver: "sqlmock",
		}, sb, dialectSql.OpenDB(dialect.MySQL, d)))
		return driver
	}

	client, err := NewMockExpectClient(
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
	entity, err := entity.NewEntityFromJSON(`{"name": "John", "age": 30}`)
	assert.NoError(t, err)
	id, err := model.Create(context.Background(), entity)
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), id)
}

func TestModelCreateFromJson(t *testing.T) {
	userSchema := &schema.Schema{}
	assert.Nil(t, json.Unmarshal([]byte(userSchemaJSON), userSchema))

	sb := createSchemaBuilder()
	createMockClient := func(d *sql.DB) db.Client {
		driver := utils.Must(NewEntClient(&db.Config{
			Driver: "sqlmock",
		}, sb, dialectSql.OpenDB(dialect.MySQL, d)))
		return driver
	}

	client, err := NewMockExpectClient(
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

	_, err = model.CreateFromJSON(context.Background(), `{"name": "John", "age"}`)
	assert.Error(t, err)

	id, err := model.CreateFromJSON(context.Background(), `{"name": "John", "age": 30}`)
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), id)
}

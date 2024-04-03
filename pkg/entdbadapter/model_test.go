package entdbadapter

import (
	"encoding/json"
	"testing"

	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
	"github.com/stretchr/testify/assert"
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

	idField := &schema.Field{
		Name: "id",
		Type: schema.TypeUint64,
		DB: &schema.FieldDB{
			Increment: true,
		},
	}
	idEntColumn := CreateEntColumn(idField)
	idColumn := &Column{field: idField, entColumn: idEntColumn}

	nameField := &schema.Field{Name: "name"}
	nameEntColumn := CreateEntColumn(nameField)
	nameColumn := &Column{field: nameField, entColumn: nameEntColumn}

	model := &Model{
		name:        "user",
		schema:      userSchema,
		entIDColumn: idEntColumn,
		columns:     []*Column{idColumn, nameColumn},
	}

	assert.Equal(t, userSchema, model.Schema())
	assert.Equal(t, nameColumn, utils.Must(model.Column("name")))

	query := model.Query()
	assert.NotNil(t, query)

	mutation, err := model.Mutation()
	assert.NoError(t, err)
	assert.NotNil(t, mutation)
}

package schema

import (
	"testing"

	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestSchema(t *testing.T) {
	_, err := NewSchemaFromJSONFile("invalid_file.json")
	assert.Error(t, err)

	s, err := NewSchemaFromJSONFile("../tests/data/schemas/user.json")
	assert.NoError(t, err)
	assert.Equal(t, "user", s.Name)

	assert.NoError(t, s.Init(false))
	assert.NoError(t, s.Init(false))
	assert.True(t, s.initialized)
	// assert.Equal(t, map[string][]string{}, s.RelationsFKColumns)
	assert.Equal(t, &Field{
		Name:  FieldID,
		Type:  TypeUint64,
		Label: "ID",
		DB: &FieldDB{
			Attr:      "UNSIGNED",
			Key:       "UNI",
			Increment: true,
		},
		Unique:        true,
		Filterable:    true,
		Sortable:      true,
		IsSystemField: true,
	}, utils.Must(s.Field(FieldID)))
	assert.True(t, len(s.DBColumns) > 0)

	assert.NotNil(t, utils.Must(s.Field(FieldCreatedAt)))
	assert.NotNil(t, utils.Must(s.Field(FieldUpdatedAt)))
	assert.NotNil(t, utils.Must(s.Field(FieldDeletedAt)))

	s2, err := NewSchemaFromJSONFile("../tests/data/schemas/user.json")
	assert.NoError(t, err)
	assert.Equal(t, "user", s2.Name)

	s2.DisableTimestamp = true
	assert.NoError(t, s2.Init(true))
	_, err = s2.Field(FieldID)
	assert.Error(t, err)
}

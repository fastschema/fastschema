package schema

import (
	"encoding/json"
	"os"
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

func TestSchemaClone(t *testing.T) {
	s := &Schema{
		Name:             "user",
		Namespace:        "schema",
		LabelFieldName:   "name",
		DisableTimestamp: false,
		DBColumns:        []string{"column1", "column2"},
		IsSystemSchema:   true,
		IsJunctionSchema: false,
		Fields: []*Field{
			{
				Name:  "field1",
				Type:  TypeString,
				Label: "label1",
			},
			{
				Name:  "field2",
				Type:  TypeString,
				Label: "label2",
			},
		},
	}

	clone := s.Clone()

	// Check if the cloned schema has the same properties as the original schema
	assert.Equal(t, s.Name, clone.Name)
	assert.Equal(t, s.Namespace, clone.Namespace)
	assert.Equal(t, s.LabelFieldName, clone.LabelFieldName)
	assert.Equal(t, s.DisableTimestamp, clone.DisableTimestamp)
	assert.Equal(t, s.DBColumns, clone.DBColumns)
	assert.Equal(t, s.IsSystemSchema, clone.IsSystemSchema)
	assert.Equal(t, s.IsJunctionSchema, clone.IsJunctionSchema)

	// Check if the cloned schema has the same fields as the original schema
	assert.Equal(t, len(s.Fields), len(clone.Fields))
	for i := range s.Fields {
		assert.Equal(t, s.Fields[i].Name, clone.Fields[i].Name)
		assert.Equal(t, s.Fields[i].Type, clone.Fields[i].Type)
		assert.Equal(t, s.Fields[i].Label, clone.Fields[i].Label)
	}
}

func TestSaveToFile(t *testing.T) {
	s := &Schema{
		Name:             "user",
		Namespace:        "schema",
		LabelFieldName:   "name",
		DisableTimestamp: false,
		DBColumns:        []string{"column1", "column2"},
		IsSystemSchema:   true,
		IsJunctionSchema: false,
		Fields: []*Field{
			{
				Name:          "field1",
				Type:          TypeString,
				Label:         "label1",
				IsSystemField: true,
				Filterable:    true,
				Sortable:      true,
				Unique:        true,
			},
			{
				Name:          "field2",
				Type:          TypeString,
				Label:         "label2",
				IsSystemField: false,
				Filterable:    false,
				Sortable:      false,
				Unique:        false,
			},
		},
	}

	tmpDir, err := os.MkdirTemp("", "fastschema")
	assert.NoError(t, err)
	filename := tmpDir + "/test_schema.json"
	assert.NoError(t, s.SaveToFile(filename))

	// Read the saved file
	fileData, err := os.ReadFile(filename)
	assert.NoError(t, err)

	// Unmarshal the file data into a new Schema object
	var savedSchema Schema
	err = json.Unmarshal(fileData, &savedSchema)
	assert.NoError(t, err)

	// Check if the saved schema has the same properties as the original schema
	assert.Equal(t, s.Name, savedSchema.Name)
	assert.Equal(t, s.Namespace, savedSchema.Namespace)
	assert.Equal(t, s.LabelFieldName, savedSchema.LabelFieldName)
	assert.Equal(t, s.DisableTimestamp, savedSchema.DisableTimestamp)
	assert.Equal(t, s.IsSystemSchema, savedSchema.IsSystemSchema)
	assert.Equal(t, s.IsJunctionSchema, savedSchema.IsJunctionSchema)

	// Check if the saved schema has the filtered fields (non-system fields) only
	assert.Equal(t, 1, len(savedSchema.Fields))
	assert.Equal(t, "field2", savedSchema.Fields[0].Name)
	assert.Equal(t, TypeString, savedSchema.Fields[0].Type)
	assert.Equal(t, "label2", savedSchema.Fields[0].Label)
	assert.False(t, savedSchema.Fields[0].IsSystemField)
	assert.False(t, savedSchema.Fields[0].Filterable)
	assert.False(t, savedSchema.Fields[0].Sortable)
	assert.False(t, savedSchema.Fields[0].Unique)
}

func TestSchemaHasField(t *testing.T) {
	s := &Schema{
		Name:             "user",
		Namespace:        "schema",
		LabelFieldName:   "name",
		DisableTimestamp: false,
		DBColumns:        []string{"column1", "column2"},
		IsSystemSchema:   true,
		IsJunctionSchema: false,
		Fields: []*Field{
			{
				Name:  "field1",
				Type:  TypeString,
				Label: "label1",
			},
			{
				Name:  "field2",
				Type:  TypeString,
				Label: "label2",
			},
		},
	}

	field := &Field{
		Name:  "field1",
		Type:  TypeString,
		Label: "label1",
	}

	exists := s.HasField(field.Name)
	assert.True(t, exists)

	field = &Field{
		Name:  "field3",
		Type:  TypeString,
		Label: "label3",
	}

	exists = s.HasField(field.Name)
	assert.False(t, exists)
}

func TestSchemaValidate(t *testing.T) {
	s := &Schema{
		Name:             "user",
		Namespace:        "schema",
		LabelFieldName:   "name",
		DisableTimestamp: false,
		DBColumns:        []string{"column1", "column2"},
		IsSystemSchema:   true,
		IsJunctionSchema: false,
		Fields: []*Field{
			{
				Name:  "name",
				Type:  TypeString,
				Label: "Name",
			},
			{
				Name:  "field2",
				Type:  TypeString,
				Label: "label2",
			},
		},
	}

	err := s.Validate()
	assert.NoError(t, err)

	// Test missing required fields
	s.Name = ""
	err = s.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "name is required")

	s.Name = "user"
	s.LabelFieldName = ""
	err = s.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "label_field is required")

	s.LabelFieldName = "name"
	s.Namespace = ""
	err = s.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "namespace is required")

	s.Namespace = "schema"
	s.Fields = []*Field{}
	err = s.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "fields is required")

	// Test invalid field names
	s.Fields = []*Field{
		{
			Name:  "id",
			Type:  TypeString,
			Label: "label1",
		},
	}
	err = s.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "field id: field name can't be 'id'")

	// Test missing field name
	s.Fields = []*Field{
		{
			Name:  "",
			Type:  TypeString,
			Label: "label1",
		},
	}
	err = s.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "field : name is required")

	// Test missing field label
	s.Fields = []*Field{
		{
			Name:  "field1",
			Type:  TypeString,
			Label: "",
		},
		{
			Name:  "name",
			Type:  TypeString,
			Label: "Name",
		},
	}
	err = s.Validate()
	assert.NoError(t, err)
	assert.Equal(t, "field1", s.Fields[0].Label)

	// Test invalid field type
	s.Fields = []*Field{
		{
			Name:  "field1",
			Type:  TypeInvalid,
			Label: "label1",
		},
	}
	err = s.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "field field1: invalid field type invalid")

	// Test missing enum values
	s.Fields = []*Field{
		{
			Name:  "field1",
			Type:  TypeEnum,
			Label: "label1",
		},
	}
	err = s.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "field field1: enums values is required")

	// Test missing relation
	s.Fields = []*Field{
		{
			Name:  "field1",
			Type:  TypeRelation,
			Label: "label1",
		},
	}
	err = s.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "field field1: relation is required")

	// Test missing relation schema
	s.Fields = []*Field{
		{
			Name:  "field1",
			Type:  TypeRelation,
			Label: "label1",
			Relation: &Relation{
				TargetSchemaName: "",
				Type:             O2O,
			},
		},
	}
	err = s.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "field field1: relation schema is required")

	// Test missing relation type
	s.Fields = []*Field{
		{
			Name:  "field1",
			Type:  TypeRelation,
			Label: "label1",
			Relation: &Relation{
				TargetSchemaName: "user",
				Type:             RelationInvalid,
			},
		},
	}
	err = s.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "field field1: relation type is required")

	// Test missing m2m relation ref field name
	s.Fields = []*Field{
		{
			Name:  "field1",
			Type:  TypeRelation,
			Label: "label1",
			Relation: &Relation{
				TargetSchemaName: "user",
				Type:             M2M,
				TargetFieldName:  "",
			},
		},
	}
	err = s.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "field field1: m2m relation ref field name is required")

	// Test invalid field type
	s.Fields = []*Field{
		{
			Name:  "field1",
			Type:  TypeInvalid,
			Label: "label1",
		},
	}
	err = s.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "field field1: type is invalid")

	// Test missing label field
	s.LabelFieldName = "nonexistent"
	err = s.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "label field 'nonexistent' is not found")
}

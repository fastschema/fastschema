package schema

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/fastschema/fastschema/entity"
	"github.com/stretchr/testify/assert"
)

func TestSchema(t *testing.T) {
	_, err := NewSchemaFromJSONFile("invalid_file.json")
	assert.Error(t, err)

	s, err := NewSchemaFromJSONFile("../tests/integration/db/data/schemas/user.json")
	assert.NoError(t, err)
	assert.Equal(t, "user", s.Name)

	assert.NoError(t, s.Init(false))
	assert.NoError(t, s.Init(false))
	assert.True(t, s.initialized)
	assert.Equal(t, &Field{
		Name:  entity.FieldID,
		Type:  TypeUint64,
		Label: "ID",
		DB: &FieldDB{
			Attr:      "UNSIGNED",
			Key:       DBPrimaryKey,
			Increment: true,
		},
		Unique:        true,
		Filterable:    true,
		Sortable:      true,
		IsSystemField: true,
		Immutable:     true,
	}, s.Field(entity.FieldID))
	assert.True(t, len(s.dbColumns) > 0)

	assert.NotNil(t, s.Field(entity.FieldCreatedAt))
	assert.NotNil(t, s.Field(entity.FieldUpdatedAt))
	assert.NotNil(t, s.Field(entity.FieldDeletedAt))

	s2, err := NewSchemaFromJSONFile("../tests/integration/db/data/schemas/user.json")
	assert.NoError(t, err)
	assert.Equal(t, "user", s2.Name)

	s2.DisableTimestamp = true
	assert.NoError(t, s2.Init(true))
	f := s2.Field(entity.FieldID)
	assert.Nil(t, f)
}

func TestSchemaClone(t *testing.T) {
	s := &Schema{
		Name:             "user",
		Namespace:        "schema",
		LabelFieldName:   "name",
		DisableTimestamp: false,
		dbColumns:        []string{"column1", "column2"},
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
	assert.Equal(t, s.dbColumns, clone.dbColumns)
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

func TestNewSchemaFromJSON(t *testing.T) {
	// Case 1: Invalid JSON
	jsonData := `invalid`
	schema, err := NewSchemaFromJSON(jsonData)
	assert.Error(t, err)
	assert.Nil(t, schema)

	// Case 2: Success
	jsonData = `{
		"name": "test",
		"namespace": "test_namespace",
		"label_field": "name",
		"disable_timestamp": false,
		"fields": [
			{
				"name": "id",
				"type": "uint64",
				"label": "ID"
			},
			{
				"name": "name",
				"type": "string",
				"label": "Name",
				"unique": true,
				"sortable": true
			},
			{
				"name": "created_at",
				"type": "time",
				"label": "Created At",
				"default": "NOW()"
			}
		]
	}`

	schema, err = NewSchemaFromJSON(jsonData)
	assert.NoError(t, err)
	assert.NoError(t, schema.Init(false))
	assert.Equal(t, "test", schema.Name)
	assert.Equal(t, "test_namespace", schema.Namespace)
	assert.Equal(t, "name", schema.LabelFieldName)
	assert.False(t, schema.DisableTimestamp)
	assert.Equal(t, 5, len(schema.Fields))

	assert.Equal(t, "id", schema.Fields[0].Name)
	assert.Equal(t, TypeUint64, schema.Fields[0].Type)
	assert.Equal(t, "ID", schema.Fields[0].Label)
	// Since the id field is defined in the JSON, it should not be treated as a system field
	assert.False(t, schema.Fields[0].IsSystemField)
	assert.True(t, schema.Fields[0].Unique)
	assert.True(t, schema.Fields[0].Sortable)
	assert.Equal(t, "UNSIGNED", schema.Fields[0].DB.Attr)
	assert.Equal(t, DBPrimaryKey, schema.Fields[0].DB.Key)
	assert.True(t, schema.Fields[0].DB.Increment)

	assert.Equal(t, "name", schema.Fields[1].Name)
	assert.Equal(t, TypeString, schema.Fields[1].Type)
	assert.Equal(t, "Name", schema.Fields[1].Label)
	assert.False(t, schema.Fields[1].IsSystemField)
	assert.True(t, schema.Fields[1].Unique)
	assert.True(t, schema.Fields[1].Sortable)

	assert.Equal(t, "created_at", schema.Fields[2].Name)
	assert.Equal(t, TypeTime, schema.Fields[2].Type)
	assert.Equal(t, "Created At", schema.Fields[2].Label)
	assert.True(t, schema.Fields[2].IsSystemField)
	assert.False(t, schema.Fields[2].Unique)
	assert.False(t, schema.Fields[2].Sortable)
	assert.Equal(t, "CURRENT_TIMESTAMP", schema.Fields[2].Default)
}

func TestSchemaPrimaryFieldProperty(t *testing.T) {
	jsonData := `{
		"name": "notes",
		"namespace": "notes",
		"label_field": "label",
		"primary_field": "ref",
		"fields": [
			{"name":"ref","type":"uuid"},
			{"name":"label","type":"string"}
		]
	}`

	s, err := NewSchemaFromJSON(jsonData)
	assert.NoError(t, err)
	assert.NoError(t, s.Init(false))

	assert.Equal(t, "ref", s.PrimaryKeyName())
	assert.Equal(t, "ref", s.PrimaryFieldName)

	refField := s.Field("ref")
	if assert.NotNil(t, refField) {
		assert.Equal(t, DBPrimaryKey, refField.DB.Key)
	}
}

type structPrimaryTarget struct {
	_     any    `json:"-" fs:"namespace=struct_pk;label_field=title;primary_field=slug"`
	Slug  string `json:"slug" fs:"type=string"`
	Title string `json:"title"`
}

func TestSchemaPrimaryFieldFromStructTag(t *testing.T) {
	s, err := CreateSchema(structPrimaryTarget{})
	assert.NoError(t, err)
	assert.NoError(t, s.Init(false))

	assert.Equal(t, "slug", s.PrimaryKeyName())
	slugField := s.Field("slug")
	assert.NotNil(t, slugField)
}

func TestSaveToFile(t *testing.T) {
	s := &Schema{
		Name:             "user",
		Namespace:        "schema",
		LabelFieldName:   "name",
		DisableTimestamp: false,
		dbColumns:        []string{"column1", "column2"},
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
		dbColumns:        []string{"column1", "column2"},
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
		Name: "field1",
	}

	exists := s.HasField(field.Name)
	assert.True(t, exists)

	field = &Field{
		Name: "field3",
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
		dbColumns:        []string{"column1", "column2"},
		IsSystemSchema:   false,
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
	assert.Contains(t, err.Error(), "label field 'name' is not found")

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

func TestErrFieldNotFound(t *testing.T) {
	err := ErrFieldNotFound("user", "field1")
	assert.Error(t, err)
	assert.Equal(t, "field user.field1 not found", err.Error())
}
func TestNewSchemaFromMap(t *testing.T) {
	// invalid schema map
	data := map[string]any{
		"name": make(chan int),
	}
	schema, err := NewSchemaFromMap(data)
	assert.Error(t, err)
	assert.Nil(t, schema)

	data = map[string]any{
		"name":              "test",
		"namespace":         "test_namespace",
		"label_field":       "name",
		"disable_timestamp": false,
		"fields": []map[string]any{
			{
				"name":  "id",
				"type":  "uint64",
				"label": "ID",
			},
			{
				"name":     "name",
				"type":     "string",
				"label":    "Name",
				"unique":   true,
				"sortable": true,
			},
			{
				"name":     "slug",
				"type":     "string",
				"label":    "Slug",
				"optional": true,
			},
		},
	}

	schema, err = NewSchemaFromMap(data)
	assert.NoError(t, err)
	assert.NoError(t, schema.Init(false))
	assert.Equal(t, "test", schema.Name)
	assert.Equal(t, "test_namespace", schema.Namespace)
	assert.Equal(t, "name", schema.LabelFieldName)
	assert.False(t, schema.DisableTimestamp)
	assert.Equal(t, 6, len(schema.Fields))

	assert.Equal(t, "id", schema.Fields[0].Name)
	assert.Equal(t, TypeUint64, schema.Fields[0].Type)
	assert.Equal(t, "ID", schema.Fields[0].Label)
	// Since the id field is defined in the JSON, it should not be treated as a system field
	assert.False(t, schema.Fields[0].IsSystemField)
	assert.True(t, schema.Fields[0].Unique)
	assert.True(t, schema.Fields[0].Sortable)
	assert.Equal(t, "UNSIGNED", schema.Fields[0].DB.Attr)
	assert.Equal(t, DBPrimaryKey, schema.Fields[0].DB.Key)
	assert.True(t, schema.Fields[0].DB.Increment)

	assert.Equal(t, "name", schema.Fields[1].Name)
	assert.Equal(t, TypeString, schema.Fields[1].Type)
	assert.Equal(t, "Name", schema.Fields[1].Label)
	assert.False(t, schema.Fields[1].IsSystemField)
	assert.True(t, schema.Fields[1].Unique)
	assert.True(t, schema.Fields[1].Sortable)

	assert.Equal(t, "slug", schema.Fields[2].Name)
	assert.Equal(t, TypeString, schema.Fields[2].Type)
	assert.Equal(t, "Slug", schema.Fields[2].Label)
	assert.False(t, schema.Fields[2].IsSystemField)
	assert.True(t, schema.Fields[2].Optional)
	assert.False(t, schema.Fields[2].Unique)
	assert.False(t, schema.Fields[2].Sortable)
}

func TestMergeSchemas(t *testing.T) {
	t.Run("merge namespace override", func(t *testing.T) {
		target := &Schema{
			Name:           "test",
			Namespace:      "original",
			LabelFieldName: "name",
		}
		source := &Schema{
			Name:      "test",
			Namespace: "overridden",
		}
		MergeSchemas(target, source)
		assert.Equal(t, "overridden", target.Namespace)
	})

	t.Run("skip namespace if same", func(t *testing.T) {
		target := &Schema{
			Name:           "test",
			Namespace:      "same",
			LabelFieldName: "name",
		}
		source := &Schema{
			Name:      "test",
			Namespace: "same",
		}
		MergeSchemas(target, source)
		assert.Equal(t, "same", target.Namespace)
	})

	t.Run("merge label field", func(t *testing.T) {
		target := &Schema{
			Name:           "test",
			Namespace:      "ns",
			LabelFieldName: "original",
		}
		source := &Schema{
			LabelFieldName: "overridden",
		}
		MergeSchemas(target, source)
		assert.Equal(t, "overridden", target.LabelFieldName)
	})

	t.Run("merge primary field", func(t *testing.T) {
		target := &Schema{
			Name:             "test",
			Namespace:        "ns",
			LabelFieldName:   "name",
			PrimaryFieldName: "",
		}
		source := &Schema{
			PrimaryFieldName: "custom_id",
		}
		MergeSchemas(target, source)
		assert.Equal(t, "custom_id", target.PrimaryFieldName)
	})

	t.Run("merge disable timestamp", func(t *testing.T) {
		target := &Schema{
			Name:             "test",
			Namespace:        "ns",
			LabelFieldName:   "name",
			DisableTimestamp: false,
		}
		source := &Schema{
			DisableTimestamp: true,
		}
		MergeSchemas(target, source)
		assert.True(t, target.DisableTimestamp)
	})

	t.Run("merge settings", func(t *testing.T) {
		target := &Schema{
			Name:           "test",
			Namespace:      "ns",
			LabelFieldName: "name",
			Settings:       nil,
		}
		newSettings := &SchemaSettings{
			Form: &SchemaFormSettings{ActiveView: "custom"},
		}
		source := &Schema{
			Settings: newSettings,
		}
		MergeSchemas(target, source)
		assert.Equal(t, newSettings, target.Settings)
	})

	t.Run("merge existing fields", func(t *testing.T) {
		target := &Schema{
			Name:           "test",
			Namespace:      "ns",
			LabelFieldName: "name",
			Fields: []*Field{
				{Name: "name", Type: TypeString, Label: "Original Label"},
			},
		}
		source := &Schema{
			Fields: []*Field{
				{Name: "name", Type: TypeString, Label: "New Label", Sortable: true},
			},
		}
		MergeSchemas(target, source)
		assert.Equal(t, "New Label", target.Field("name").Label)
		assert.True(t, target.Field("name").Sortable)
	})

	t.Run("skip merge for system fields", func(t *testing.T) {
		target := &Schema{
			Name:           "test",
			Namespace:      "ns",
			LabelFieldName: "name",
			Fields: []*Field{
				{Name: "id", Type: TypeUint64, Label: "ID", IsSystemField: true},
			},
		}
		source := &Schema{
			Fields: []*Field{
				{Name: "id", Type: TypeUint64, Label: "Custom ID", IsSystemField: true},
			},
		}
		MergeSchemas(target, source)
		// System field should not be merged
		assert.Equal(t, "ID", target.Field("id").Label)
	})

	t.Run("add new non-system fields", func(t *testing.T) {
		target := &Schema{
			Name:           "test",
			Namespace:      "ns",
			LabelFieldName: "name",
			Fields: []*Field{
				{Name: "name", Type: TypeString, Label: "Name"},
			},
		}
		source := &Schema{
			Fields: []*Field{
				{Name: "description", Type: TypeText, Label: "Description"},
			},
		}
		MergeSchemas(target, source)
		assert.NotNil(t, target.Field("description"))
		assert.Equal(t, "Description", target.Field("description").Label)
	})

	t.Run("skip adding system fields from source", func(t *testing.T) {
		target := &Schema{
			Name:           "test",
			Namespace:      "ns",
			LabelFieldName: "name",
			Fields:         []*Field{},
		}
		source := &Schema{
			Fields: []*Field{
				{Name: "created_at", Type: TypeTime, Label: "Created At", IsSystemField: true},
			},
		}
		MergeSchemas(target, source)
		// System field should not be added
		assert.Nil(t, target.Field("created_at"))
	})

	t.Run("merge DB indexes - add new", func(t *testing.T) {
		target := &Schema{
			Name:           "test",
			Namespace:      "ns",
			LabelFieldName: "name",
			DB:             nil,
		}
		source := &Schema{
			DB: &SchemaDB{
				Indexes: []*SchemaDBIndex{
					{Name: "idx_name", Columns: []string{"name"}},
				},
			},
		}
		MergeSchemas(target, source)
		assert.NotNil(t, target.DB)
		assert.Len(t, target.DB.Indexes, 1)
		assert.Equal(t, "idx_name", target.DB.Indexes[0].Name)
	})

	t.Run("merge DB indexes - skip existing", func(t *testing.T) {
		target := &Schema{
			Name:           "test",
			Namespace:      "ns",
			LabelFieldName: "name",
			DB: &SchemaDB{
				Indexes: []*SchemaDBIndex{
					{Name: "idx_name", Columns: []string{"name"}, Unique: false},
				},
			},
		}
		source := &Schema{
			DB: &SchemaDB{
				Indexes: []*SchemaDBIndex{
					{Name: "idx_name", Columns: []string{"name"}, Unique: true}, // Different config
					{Name: "idx_email", Columns: []string{"email"}},
				},
			},
		}
		MergeSchemas(target, source)
		assert.Len(t, target.DB.Indexes, 2)
		// Original index should not be modified
		assert.False(t, target.DB.Indexes[0].Unique)
		// New index should be added
		assert.Equal(t, "idx_email", target.DB.Indexes[1].Name)
	})

	t.Run("merge DB indexes - nil source indexes", func(t *testing.T) {
		target := &Schema{
			Name:           "test",
			Namespace:      "ns",
			LabelFieldName: "name",
			DB: &SchemaDB{
				Indexes: []*SchemaDBIndex{
					{Name: "idx_name", Columns: []string{"name"}},
				},
			},
		}
		source := &Schema{
			DB: &SchemaDB{
				Indexes: nil,
			},
		}
		MergeSchemas(target, source)
		// Target indexes should remain unchanged
		assert.Len(t, target.DB.Indexes, 1)
	})
}

func TestPrimaryKeyName(t *testing.T) {
	t.Run("returns cached primaryField", func(t *testing.T) {
		s := &Schema{
			Name:           "test",
			Namespace:      "ns",
			LabelFieldName: "name",
			primaryField:   "cached_pk",
		}
		assert.Equal(t, "cached_pk", s.PrimaryKeyName())
	})

	t.Run("returns PrimaryFieldName when set", func(t *testing.T) {
		s := &Schema{
			Name:             "test",
			Namespace:        "ns",
			LabelFieldName:   "name",
			PrimaryFieldName: "custom_pk",
		}
		assert.Equal(t, "custom_pk", s.PrimaryKeyName())
	})

	t.Run("returns id when field exists", func(t *testing.T) {
		s := &Schema{
			Name:           "test",
			Namespace:      "ns",
			LabelFieldName: "name",
			Fields: []*Field{
				{Name: entity.FieldID, Type: TypeUint64},
			},
		}
		assert.Equal(t, entity.FieldID, s.PrimaryKeyName())
	})

	t.Run("returns empty when no primary field", func(t *testing.T) {
		s := &Schema{
			Name:           "test",
			Namespace:      "ns",
			LabelFieldName: "name",
			Fields:         []*Field{},
		}
		assert.Equal(t, "", s.PrimaryKeyName())
	})
}

func TestPrimaryField(t *testing.T) {
	t.Run("returns nil when no primary key", func(t *testing.T) {
		s := &Schema{
			Name:           "test",
			Namespace:      "ns",
			LabelFieldName: "name",
			Fields:         []*Field{},
		}
		assert.Nil(t, s.PrimaryField())
	})

	t.Run("returns field when exists", func(t *testing.T) {
		s := &Schema{
			Name:           "test",
			Namespace:      "ns",
			LabelFieldName: "name",
			primaryField:   "custom_id",
			Fields: []*Field{
				{Name: "custom_id", Type: TypeUint64},
			},
		}
		f := s.PrimaryField()
		assert.NotNil(t, f)
		assert.Equal(t, "custom_id", f.Name)
	})
}

func TestEnsurePrimaryFieldEdgeCases(t *testing.T) {
	t.Run("custom primary field not found", func(t *testing.T) {
		s := &Schema{
			Name:             "test",
			Namespace:        "ns",
			LabelFieldName:   "name",
			PrimaryFieldName: "nonexistent",
			Fields: []*Field{
				{Name: "name", Type: TypeString, Label: "Name"},
			},
		}
		err := s.Init(false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "primary field 'nonexistent' is not found")
	})

	t.Run("disable ID column with no candidate", func(t *testing.T) {
		s := &Schema{
			Name:           "test",
			Namespace:      "ns",
			LabelFieldName: "name",
			Fields: []*Field{
				{Name: "name", Type: TypeString},
			},
		}
		// Should succeed - no error when disableIDColumn is true
		err := s.Init(true)
		assert.NoError(t, err)
		assert.Equal(t, "", s.primaryField)
	})
}

func TestApplyPrimaryFieldDefaultsEdgeCases(t *testing.T) {
	t.Run("non-integer type disables increment", func(t *testing.T) {
		s := &Schema{
			Name:             "test",
			Namespace:        "ns",
			LabelFieldName:   "slug",
			PrimaryFieldName: "slug",
			Fields: []*Field{
				{Name: "slug", Type: TypeString, Label: "Slug"},
			},
		}
		err := s.Init(false)
		assert.NoError(t, err)
		slugField := s.Field("slug")
		assert.NotNil(t, slugField)
		assert.False(t, slugField.DB.Increment)
		assert.Equal(t, DBPrimaryKey, slugField.DB.Key)
	})

	t.Run("signed integer primary key", func(t *testing.T) {
		s := &Schema{
			Name:             "test",
			Namespace:        "ns",
			LabelFieldName:   "name",
			PrimaryFieldName: "legacy_id",
			Fields: []*Field{
				{Name: "legacy_id", Type: TypeInt64, Label: "Legacy ID"},
				{Name: "name", Type: TypeString, Label: "Name"},
			},
		}
		err := s.Init(false)
		assert.NoError(t, err)
		pkField := s.Field("legacy_id")
		assert.NotNil(t, pkField)
		// Should not have UNSIGNED for signed int
		assert.NotEqual(t, "UNSIGNED", pkField.DB.Attr)
		// Should have increment for integer
		assert.True(t, pkField.DB.Increment)
	})
}

package schema

import (
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/stretchr/testify/assert"
)

type testcategory struct {
	Name  string      `json:"name"`
	Posts []*testpost `json:"posts" fs.relation:"{'type':'o2m','schema':'testpost','field':'category','owner':true}"`
}

type testpost struct {
	Name       string        `json:"name"`
	CategoryID uint64        `json:"cat_id"`
	Category   *testcategory `json:"category" fs.relation:"{'type':'o2m','schema':'testcategory','field':'posts','fk_columns':{'target_column':'cat_id'}}"`
}

func TestNewBuilderFromSchemasErrorInvalidSchema(t *testing.T) {
	dir := t.TempDir()
	schemas := map[string]*Schema{
		"post": {
			Name:      "post",
			Namespace: "posts",
			Fields: []*Field{
				{
					Name: "name",
					Type: TypeString,
				},
			},
		},
	}

	builder, err := NewBuilderFromSchemas(dir, schemas, testcategory{}, testpost{})
	assert.Nil(t, builder)
	assert.Contains(t, err.Error(), "label_field is required")
}

func TestNewBuilderFromSchemasErrorInvalidSystemSchema(t *testing.T) {
	_, err := NewBuilderFromSchemas(t.TempDir(), nil, (*int)(nil))
	assert.Contains(t, err.Error(), "can not create schema from invalid type *int")
}

func TestNewBuilderFromSchemasErrorDuplicateSchema(t *testing.T) {
	_, err := NewBuilderFromSchemas(t.TempDir(), nil, testcategory{}, testpost{}, testcategory{})
	assert.Contains(t, err.Error(), "testcategory already exists")
}

func TestNewBuilderFromSchemas(t *testing.T) {
	dir := t.TempDir()
	schemas := map[string]*Schema{
		"post": {
			Name:           "post",
			Namespace:      "posts",
			LabelFieldName: "name",
			Fields: []*Field{
				{
					Name: "name",
					Type: TypeString,
				},
			},
		},
	}

	builder := utils.Must(NewBuilderFromSchemas(dir, schemas, testcategory{}, testpost{}))
	assert.Equal(t, dir, builder.dir)
	assert.Equal(t, len(schemas)+2, len(builder.schemas))
	for name, schema := range schemas {
		assert.Equal(t, schema, builder.schemas[name])
	}
}

func TestNewBuilderFromDir(t *testing.T) {
	_, err := NewBuilderFromDir("../tests/invalid", testcategory{}, testpost{})
	assert.Error(t, err)

	tmpDir, err := os.MkdirTemp("../tests/", "testbuilder")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	invalidSchemaJSONFile1 := filepath.Join(tmpDir, "invalid2.json")
	utils.WriteFile(invalidSchemaJSONFile1, "{}")
	_, err = NewBuilderFromDir(tmpDir, testcategory{}, testpost{})
	assert.Error(t, err)

	invalidSchemaJSONFile2 := filepath.Join(tmpDir, "invalid1.json")
	utils.WriteFile(invalidSchemaJSONFile2, "{")
	_, err = NewBuilderFromDir(tmpDir, testcategory{}, testpost{})
	assert.Error(t, err)

	builder, err := NewBuilderFromDir("../tests/data/schemas", testcategory{}, testpost{})
	assert.Nil(t, err)
	assert.NotNil(t, builder)

	schemas := builder.Schemas()
	assert.True(t, len(schemas) > 0)

	newSchema := &Schema{
		Name: "newSchema",
	}
	schemas = append(schemas, newSchema)
	builder.AddSchema(newSchema)
	assert.Equal(t, len(schemas), len(builder.Schemas()))

	userSchema, err := builder.Schema("user")
	assert.Nil(t, err)
	assert.NotNil(t, userSchema)

	_, err = builder.Schema("invalid")
	assert.Error(t, err)
}

func TestBuilderClone(t *testing.T) {
	// Create a new builder
	builder := &Builder{
		dir:       "../tests/data/schemas",
		schemas:   map[string]*Schema{},
		relations: []*Relation{},
	}

	// Add a schema to the builder
	schema := &Schema{
		Name: "user",
	}
	builder.schemas[schema.Name] = schema

	// Add a relation to the builder
	relation := &Relation{
		Type: O2O,
	}
	builder.relations = append(builder.relations, relation)

	// Clone the builder
	clone := builder.Clone()

	// Check if the cloned builder has the same directory
	assert.Equal(t, builder.dir, clone.dir)

	// Check if the cloned builder has the same schemas
	for name, schema := range builder.schemas {
		clonedSchema, ok := clone.schemas[name]
		assert.True(t, ok)
		assert.Equal(t, schema.Name, clonedSchema.Name)
	}

	// Check if the cloned builder has the same relations
	assert.Equal(t, len(builder.relations), len(clone.relations))
	for i := range builder.relations {
		assert.Equal(t, builder.relations[i].Type, clone.relations[i].Type)
	}
}

func TestSaveToDir(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "testsave")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a new builder
	builder := &Builder{
		dir:       "../tests/data/schemas",
		schemas:   map[string]*Schema{},
		relations: []*Relation{},
	}

	// Add a schema to the builder
	schema := &Schema{
		Name: "user",
	}
	builder.schemas[schema.Name] = schema

	// Save the schemas to the temporary directory
	err = builder.SaveToDir(tmpDir)
	assert.Nil(t, err)

	// Check if the schema files are saved correctly
	schemaFile := filepath.Join(tmpDir, "user.json")
	_, err = os.Stat(schemaFile)
	assert.False(t, os.IsNotExist(err))
}

func TestSaveToDirNonExistent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "testsave")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a new builder
	builder := &Builder{
		dir:       "../tests/data/schemas",
		schemas:   map[string]*Schema{},
		relations: []*Relation{},
	}

	// Add a schema to the builder
	schema := &Schema{
		Name: "user",
	}
	builder.schemas[schema.Name] = schema

	// Save the schemas to the temporary directory
	err = builder.SaveToDir(filepath.Join(tmpDir, "nonexistent"))
	assert.Nil(t, err)

	// Check if the schema files are saved correctly
	schemaFile := filepath.Join(tmpDir, "nonexistent", "user.json")
	_, err = os.Stat(schemaFile)
	assert.False(t, os.IsNotExist(err))
}

func TestBuilderDir(t *testing.T) {
	builder := &Builder{}
	expectedDir := "/path/to/directory"

	// Test setting the directory
	builder.Dir(expectedDir)
	assert.Equal(t, expectedDir, builder.dir)

	// Test getting the directory
	actualDir := builder.Dir()
	assert.Equal(t, expectedDir, actualDir)
}

func TestBuilderInitEmptySchemas(t *testing.T) {
	builder := &Builder{
		dir:       "../tests/data/schemas",
		schemas:   nil,
		relations: []*Relation{},
	}

	err := builder.Init()
	assert.Nil(t, err)
	assert.Equal(t, builder.schemas, map[string]*Schema{})
}

func TestBuilderInitCreateRelationsError(t *testing.T) {
	builder := &Builder{
		dir: "../tests/data/schemas",
		schemas: map[string]*Schema{
			"user": {
				Name: "user",
				Fields: []*Field{
					{
						Name: "category",
						Type: TypeRelation,
						Relation: &Relation{
							TargetSchemaName: "invalid",
						},
					},
				},
			},
		},
	}

	err := builder.Init()
	assert.Error(t, err)
}

func TestBuilderInitCreateFkError(t *testing.T) {
	builder := &Builder{
		dir: "../tests/data/schemas",
		relations: []*Relation{
			{
				SchemaName: "invalid",
				BackRef:    &Relation{},
			},
		},
	}

	err := builder.Init()
	assert.Error(t, err)
}

func TestBuilderSchemaFile(t *testing.T) {
	builder := &Builder{
		dir: "/path/to/directory",
	}

	name := "user"
	expectedFile := "/path/to/directory/user.json"
	actualFile := builder.SchemaFile(name)

	assert.Equal(t, expectedFile, actualFile)
}

func TestBuilderAddSchemaNilSchemas(t *testing.T) {
	builder := &Builder{
		dir:     "/path/to/directory",
		schemas: nil,
	}

	schema := &Schema{
		Name: "user",
	}

	builder.AddSchema(schema)
	assert.NotNil(t, builder.schemas)
	assert.Equal(t, 1, len(builder.schemas))
}

func TestReplaceSchema(t *testing.T) {
	builder := &Builder{
		dir:       "../tests/data/schemas",
		schemas:   map[string]*Schema{},
		relations: []*Relation{},
	}

	// Create a schema to replace
	oldSchema := &Schema{
		Name:      "testSchema",
		Namespace: "oldNamespace",
	}
	builder.schemas[oldSchema.Name] = oldSchema

	// Create a new schema
	newSchema := &Schema{
		Name:      "testSchema",
		Namespace: "newNamespace",
	}

	// Replace the schema
	builder.ReplaceSchema(oldSchema.Name, newSchema)

	// Check if the schema is replaced
	schema, ok := builder.schemas[oldSchema.Name]
	assert.True(t, ok)
	assert.Equal(t, newSchema.Namespace, schema.Namespace)
}

func TestBuilderRelations(t *testing.T) {
	builder := &Builder{
		relations: []*Relation{
			{
				Type: O2O,
			},
			{
				Type: O2M,
			},
			{
				Type: M2M,
			},
		},
	}

	relations := builder.Relations()
	assert.Equal(t, 3, len(relations))
	assert.Equal(t, O2O, relations[0].Type)
	assert.Equal(t, O2M, relations[1].Type)
	assert.Equal(t, M2M, relations[2].Type)
}

func TestBuilderCreateRelationsM2MInvalidSchema(t *testing.T) {
	builder := &Builder{
		dir: "../tests/data/schemas",
		relations: []*Relation{
			{
				Type:       M2M,
				SchemaName: "invalid",
			},
		},
	}

	err := builder.CreateRelations()
	assert.Error(t, err)
}

func TestBuilderCreateRelationsBackRefError(t *testing.T) {
	builder := &Builder{
		dir: "../tests/data/schemas",
		relations: []*Relation{
			{
				Type:       O2M,
				SchemaName: "user",
			},
		},
	}

	err := builder.CreateRelations()
	assert.Error(t, err)
}

func TestBuilderCreateRelationsJunctionSchemaError(t *testing.T) {
	builder := &Builder{
		dir: "../tests/data/schemas",
		schemas: map[string]*Schema{
			"user": {
				Name: "user",
			},
		},
		relations: []*Relation{
			{
				Type:       M2M,
				SchemaName: "user",
			},
		},
	}

	err := builder.CreateRelations()
	assert.Error(t, err)
}

func TestCreateM2mJunctionSchemaError(t *testing.T) {
	builder := &Builder{
		dir:       "../tests/data/schemas",
		schemas:   map[string]*Schema{},
		relations: []*Relation{},
	}

	currentSchema := &Schema{
		Name: "user",
	}

	relation := &Relation{
		Type: O2M,
	}

	_, _, err := builder.CreateM2mJunctionSchema(currentSchema, relation)
	assert.Error(t, err)
}

func TestSaveToDirPermissionError(t *testing.T) {
	// make tmp dir read only
	tmpDir, err := os.MkdirTemp("", "testsave")
	assert.Nil(t, err)

	err = os.Chmod(tmpDir, 0400)
	assert.Nil(t, err)

	// Create a new builder
	builder := &Builder{
		dir:       "../tests/data/schemas",
		schemas:   map[string]*Schema{},
		relations: []*Relation{},
	}

	// Add a schema to the builder
	schema := &Schema{
		Name: "user",
	}

	builder.schemas[schema.Name] = schema

	// Save the schemas to the temporary directory
	err = builder.SaveToDir(tmpDir)
	assert.Error(t, err)
}

func TestGetSchemasFromDirError(t *testing.T) {
	type Post struct {
		Name string
	}

	// Case 1: Invalid system schema type
	schemas, err := GetSchemasFromDir(t.TempDir(), "invalid")
	assert.Nil(t, schemas)
	assert.Contains(t, err.Error(), "can not create schema from invalid type string")

	// Case 2: Duplicate schema name
	schemas, err = GetSchemasFromDir(t.TempDir(), Post{}, Post{})
	assert.Nil(t, schemas)
	assert.Contains(t, err.Error(), "system schema post already exists")
}

func TestGetSchemasFromDirExtendsSystemSchemas(t *testing.T) {
	type Post struct {
		Name string
	}

	schemaDir := t.TempDir()
	assert.NoError(t, os.WriteFile(schemaDir+"/post.json", []byte(`{
		"name": "post",
		"fields": [
			{
				"name": "slug",
				"type": "string",
				"label": "Slug"
			}
		]
	}`), 0644))

	schemas, err := GetSchemasFromDir(schemaDir, Post{})
	assert.Nil(t, err)
	assert.NotNil(t, schemas)
	assert.Equal(t, 1, len(schemas))
	assert.NotNil(t, schemas["post"])
}

func TestFKUseExistedField(t *testing.T) {
	sb, err := NewBuilderFromDir(t.TempDir(), testcategory{}, testpost{})
	assert.Nil(t, err)
	postSchema, err := sb.Schema("testpost")
	assert.Nil(t, err)
	assert.NotNil(t, postSchema)
	relation := sb.Relation("testpost.category-testcategory.posts")
	assert.NotNil(t, relation)
	assert.Equal(t, "cat_id", relation.FKColumns.TargetColumn)
}

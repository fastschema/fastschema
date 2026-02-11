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
	Category   *testcategory `json:"category" fs.relation:"{'type':'o2m','schema':'testcategory','field':'posts','source_column':'cat_id'}"`
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
	assert.Contains(t, err.Error(), "Missing 'label_field'")
}

func TestNewBuilderFromSchemasErrorInvalidSystemSchema(t *testing.T) {
	_, err := NewBuilderFromSchemas(t.TempDir(), nil, (*int)(nil))
	assert.Contains(t, err.Error(), "can not create schema from invalid type *int")
}

func TestNewBuilderFromSchemasErrorDuplicateSchema(t *testing.T) {
	_, err := NewBuilderFromSchemas(t.TempDir(), nil, testcategory{}, testpost{}, testcategory{})
	assert.Contains(t, err.Error(), "Duplicate system schema")
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

	builder, err := NewBuilderFromDir("../tests/integration/db/data/schemas", testcategory{}, testpost{})
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
				SourceSchemaName: "invalid",
				BackRef:          &Relation{},
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
				Type:             M2M,
				SourceSchemaName: "invalid",
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
				Type:             O2M,
				SourceSchemaName: "user",
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
				Type:             M2M,
				SourceSchemaName: "user",
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

	sourceSchema := &Schema{Name: "user"}
	relation := &Relation{Type: O2M}
	_, _, err := builder.CreateM2mJunctionSchema(sourceSchema, relation)
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
	assert.Contains(t, err.Error(), "Duplicate system schema")
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
	assert.Equal(t, "cat_id", relation.SourceColumn)
}

// Tests for multi-error collection functions

func TestBuilderErrorsType(t *testing.T) {
	errs := &BuilderErrors{}

	// Test empty errors
	assert.False(t, errs.HasErrors())
	assert.Equal(t, "", errs.Error())

	// Test adding nil error (should be ignored)
	errs.Add(nil)
	assert.False(t, errs.HasErrors())

	// Test adding actual errors
	errs.Add(BuilderSchemaNotFoundError("test1", nil))
	assert.True(t, errs.HasErrors())
	assert.Equal(t, 1, len(errs.Errors))

	errs.Add(BuilderSchemaNotFoundError("test2", nil))
	assert.Equal(t, 2, len(errs.Errors))

	// Test Error() joins messages
	errString := errs.Error()
	assert.Contains(t, errString, "test1")
	assert.Contains(t, errString, "test2")
}

func TestNewBuilderFromSchemasCollectErrors_MultipleSchemaErrors(t *testing.T) {
	dir := t.TempDir()

	// Create schemas with multiple different errors
	schemas := map[string]*Schema{
		"post": {
			Name:      "post",
			Namespace: "posts",
			// Missing label_field
			Fields: []*Field{
				{
					Name: "title",
					Type: TypeString,
				},
			},
		},
		"comment": {
			Name: "comment",
			// Missing namespace
			LabelFieldName: "text",
			Fields: []*Field{
				{
					Name: "text",
					Type: TypeString,
				},
			},
		},
	}

	_, errs := NewBuilderFromSchemasCollectErrors(dir, schemas)

	// Should have errors from both schemas
	assert.True(t, errs.HasErrors())
	assert.GreaterOrEqual(t, len(errs.Errors), 2, "Should collect errors from both schemas")

	// Verify both types of errors are present
	errString := errs.Error()
	assert.Contains(t, errString, "label_field")
	assert.Contains(t, errString, "namespace")
}

func TestNewBuilderFromSchemasCollectErrors_RelationErrors(t *testing.T) {
	dir := t.TempDir()

	// Create schemas with multiple relation errors
	schemas := map[string]*Schema{
		"post": {
			Name:           "post",
			Namespace:      "posts",
			LabelFieldName: "title",
			Fields: []*Field{
				{
					Name: "title",
					Type: TypeString,
				},
				{
					Name: "author",
					Type: TypeRelation,
					Relation: &Relation{
						TargetSchemaName: "nonexistent_user", // Error 1: schema doesn't exist
						TargetFieldName:  "posts",
						Type:             O2M,
					},
				},
				{
					Name: "category",
					Type: TypeRelation,
					Relation: &Relation{
						TargetSchemaName: "nonexistent_category", // Error 2: schema doesn't exist
						TargetFieldName:  "posts",
						Type:             O2M,
					},
				},
			},
		},
	}

	_, errs := NewBuilderFromSchemasCollectErrors(dir, schemas)

	// Should have errors for both missing relation targets
	assert.True(t, errs.HasErrors())
	assert.GreaterOrEqual(t, len(errs.Errors), 2, "Should collect errors from both relations")
}

func TestNewBuilderFromSchemasCollectErrors_ValidSchemas(t *testing.T) {
	dir := t.TempDir()

	// Create valid schemas
	schemas := map[string]*Schema{
		"post": {
			Name:           "post",
			Namespace:      "posts",
			LabelFieldName: "title",
			Fields: []*Field{
				{
					Name: "title",
					Type: TypeString,
				},
			},
		},
	}

	builder, errs := NewBuilderFromSchemasCollectErrors(dir, schemas, testcategory{}, testpost{})

	// Should have no errors
	assert.False(t, errs.HasErrors())
	assert.NotNil(t, builder)
	assert.Equal(t, 3, len(builder.schemas)) // post + testcategory + testpost
}

func TestNewBuilderFromSchemasCollectErrors_DuplicateSystemSchema(t *testing.T) {
	_, errs := NewBuilderFromSchemasCollectErrors(t.TempDir(), nil, testcategory{}, testpost{}, testcategory{})

	// Should collect duplicate schema error
	assert.True(t, errs.HasErrors())
	assert.Contains(t, errs.Error(), "Duplicate system schema")
}

func TestInitCollectErrors_EmptySchemas(t *testing.T) {
	builder := &Builder{
		dir:       t.TempDir(),
		schemas:   nil,
		relations: []*Relation{},
	}

	errs := builder.InitCollectErrors()
	assert.False(t, errs.HasErrors())
	assert.NotNil(t, builder.schemas)
}

func TestCreateRelationsCollectErrors_MultipleErrors(t *testing.T) {
	builder := &Builder{
		dir: t.TempDir(),
		schemas: map[string]*Schema{
			"user": {
				Name: "user",
				Fields: []*Field{
					{
						Name: "role",
						Type: TypeRelation,
						Relation: &Relation{
							TargetSchemaName: "invalid_role", // Error 1
						},
					},
					{
						Name: "profile",
						Type: TypeRelation,
						Relation: &Relation{
							TargetSchemaName: "invalid_profile", // Error 2
						},
					},
				},
			},
		},
	}

	errs := builder.CreateRelationsCollectErrors()

	// Should collect both relation errors
	assert.True(t, errs.HasErrors())
	assert.GreaterOrEqual(t, len(errs.Errors), 2, "Should collect errors from both relation fields")
}

func TestCreateRelationsCollectErrors_BackRefErrors(t *testing.T) {
	builder := &Builder{
		dir: t.TempDir(),
		relations: []*Relation{
			{
				Type:             O2M,
				SourceSchemaName: "post",
				Name:             "post.author-user.posts",
				// Missing BackRef will cause error
			},
			{
				Type:             O2M,
				SourceSchemaName: "comment",
				Name:             "comment.author-user.comments",
				// Missing BackRef will cause error
			},
		},
	}

	errs := builder.CreateRelationsCollectErrors()

	// Should collect both backref errors
	assert.True(t, errs.HasErrors())
	assert.GreaterOrEqual(t, len(errs.Errors), 2, "Should collect errors from both missing backrefs")
}


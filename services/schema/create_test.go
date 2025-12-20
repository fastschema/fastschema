package schemaservice_test

import (
	"bytes"
	"net/http/httptest"
	"testing"

	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
	"github.com/stretchr/testify/assert"
)

func TestSchemaServiceCreate(t *testing.T) {
	testApp, _, server := createSchemaService(t, nil)

	// Case 1: schema already exists
	req := httptest.NewRequest("POST", "/schema", bytes.NewReader([]byte(`{"name": "category"}`)))
	resp := utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 400, resp.StatusCode)
	response := utils.Must(utils.ReadCloserToString(resp.Body))
	assert.Contains(t, response, `schema already exists`)

	// Case 2: schema validation failed
	req = httptest.NewRequest("POST", "/schema", bytes.NewReader([]byte(`{"name": "blog"}`)))
	resp = utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 422, resp.StatusCode)
	response = utils.Must(utils.ReadCloserToString(resp.Body))
	assert.Contains(t, response, `schema validation error`)
	assert.Contains(t, response, `label_field is required`)
	assert.Contains(t, response, `namespace is required`)

	// Case 3: invalid relation schema
	blogSchema := utils.Must(schema.NewSchemaFromYAML(testBlogYAML))
	categoriesField := &schema.Field{
		Type:  schema.TypeRelation,
		Name:  "categories",
		Label: "Categories",
		Relation: &schema.Relation{
			Type:             schema.M2M,
			TargetSchemaName: "cat", // Invalid schema
			TargetFieldName:  "blogs",
			Owner:            true,
			Optional:         false,
		},
	}
	blogSchema.Fields = append(blogSchema.Fields, categoriesField)
	newBlogJSON := schemaToJSON(blogSchema)
	req = httptest.NewRequest("POST", "/schema", bytes.NewReader([]byte(newBlogJSON)))
	resp = utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 400, resp.StatusCode)
	response = utils.Must(utils.ReadCloserToString(resp.Body))
	assert.Contains(t, response, `Invalid field 'blog.categories'. Target schema 'cat' not found`)

	// Case 4: target relation field existed
	blogSchema2 := utils.Must(schema.NewSchemaFromYAML(testBlogYAML))
	categoriesField2 := &schema.Field{
		Type:  schema.TypeRelation,
		Name:  "categories",
		Label: "Categories",
		Relation: &schema.Relation{
			Type:             schema.M2M,
			TargetSchemaName: "category",
			TargetFieldName:  "name", // Invalid: this field already exists
			Owner:            true,
			Optional:         false,
		},
	}
	blogSchema2.Fields = append(blogSchema2.Fields, categoriesField2)
	newBlogJSON = schemaToJSON(blogSchema2)
	req = httptest.NewRequest("POST", "/schema", bytes.NewReader([]byte(newBlogJSON)))
	resp = utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 400, resp.StatusCode)
	response = utils.Must(utils.ReadCloserToString(resp.Body))
	assert.Contains(t, response, `Invalid field 'blog.categories'. Target schema 'category' already has field 'name'`)

	// Case 5: create schema successfully
	blogSchema3 := utils.Must(schema.NewSchemaFromYAML(testBlogYAML))
	categoriesField3 := &schema.Field{
		Type:  schema.TypeRelation,
		Name:  "categories",
		Label: "Categories",
		Relation: &schema.Relation{
			Type:             schema.M2M,
			TargetSchemaName: "category",
			TargetFieldName:  "blogs",
			Owner:            true,
			Optional:         false,
		},
	}
	blogSchema3.Fields = append(blogSchema3.Fields, categoriesField3)
	newBlogJSON = schemaToJSON(blogSchema3)
	req = httptest.NewRequest("POST", "/schema", bytes.NewReader([]byte(newBlogJSON)))
	resp = utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 200, resp.StatusCode)
	response = utils.Must(utils.ReadCloserToString(resp.Body))
	assert.NotEmpty(t, response)

	blogSchemaFromDB := utils.Must(testApp.SchemaBuilder().Schema("blog"))
	blogCategoriesField := blogSchemaFromDB.Field("categories")
	assert.NotNil(t, blogCategoriesField)
	assert.Equal(t, "relation", blogCategoriesField.Type.String())
	assert.Equal(t, schema.M2M, blogCategoriesField.Relation.Type)
	assert.Equal(t, "category", blogCategoriesField.Relation.TargetSchemaName)
	assert.Equal(t, "blogs", blogCategoriesField.Relation.TargetFieldName)

	categorySchema := utils.Must(testApp.SchemaBuilder().Schema("category"))
	categoryBlogsField := categorySchema.Field("blogs")
	assert.NotNil(t, categoryBlogsField)
	assert.Equal(t, "relation", categoryBlogsField.Type.String())
	assert.Equal(t, schema.M2M, categoryBlogsField.Relation.Type)
	assert.Equal(t, "blog", categoryBlogsField.Relation.TargetSchemaName)
	assert.Equal(t, "categories", categoryBlogsField.Relation.TargetFieldName)
}

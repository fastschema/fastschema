package schemaservice_test

import (
	"bytes"
	"net/http/httptest"
	"testing"

	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
	"github.com/stretchr/testify/assert"
)

func TestSchemaServiceDelete(t *testing.T) {
	_, _, server := createSchemaService(t, nil)

	// Case 1: schema not found
	req := httptest.NewRequest("DELETE", "/schema/product", nil)
	resp := utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 404, resp.StatusCode)
	response := utils.Must(utils.ReadCloserToString(resp.Body))
	assert.Contains(t, response, `schema product not found`)

	// Case 2: schema has no relation
	blogSchema := utils.Must(schema.NewSchemaFromYAML(testBlogYAML))
	blogJSON := schemaToJSON(blogSchema)
	req = httptest.NewRequest("POST", "/schema", bytes.NewReader([]byte(blogJSON)))
	resp = utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 200, resp.StatusCode)
	response = utils.Must(utils.ReadCloserToString(resp.Body))
	assert.NotEmpty(t, response)

	req = httptest.NewRequest("DELETE", "/schema/blog", nil)
	resp = utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 200, resp.StatusCode)
	response = utils.Must(utils.ReadCloserToString(resp.Body))
	assert.Contains(t, response, `Schema deleted`)

	// Case 3: schema has relation
	blogSchema = utils.Must(schema.NewSchemaFromYAML(testBlogYAML))
	categoriesField := &schema.Field{
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
	blogSchema.Fields = append(blogSchema.Fields, categoriesField)
	newBlogJSON := schemaToJSON(blogSchema)
	req = httptest.NewRequest("POST", "/schema", bytes.NewReader([]byte(newBlogJSON)))
	resp = utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 200, resp.StatusCode)
	response = utils.Must(utils.ReadCloserToString(resp.Body))
	assert.NotEmpty(t, response)

	req = httptest.NewRequest("DELETE", "/schema/blog", nil)
	resp = utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 200, resp.StatusCode)
	response = utils.Must(utils.ReadCloserToString(resp.Body))
	assert.NotEmpty(t, response)
}

// case: delete schema with relations
func TestSchemaServiceDeleteWithRelationsField(t *testing.T) {
	testApp, _, server := createSchemaService(t, &testSchemaSeviceConfig{
		extraSchemas: map[string]string{
			"blog": testBlogYAML,
			"tag":  testTagYAML,
		},
	})
	// add relation field
	addFieldCategoryToBlog(t, testApp, server)
	req := httptest.NewRequest("DELETE", "/schema/blog", nil)
	resp := utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 200, resp.StatusCode)
	response := utils.Must(utils.ReadCloserToString(resp.Body))
	assert.NotEmpty(t, response)

	// check blogs field was deleted in target schema
	categoryFieldBlogs := testApp.Schema("category").Field("blogs")
	assert.Nil(t, categoryFieldBlogs)
}

package schemaservice_test

import (
	"bytes"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
	"github.com/stretchr/testify/assert"
)

var categorySchemaJSON = `{
	"name": "category",
	"namespace": "categories",
	"label_field": "name",
	"fields": [
		{
			"type": "string",
			"name": "name",
			"label": "Name",
			"unique": true,
			"sortable": true
		}
	]
}`

func TestSchemaServiceUpdateError(t *testing.T) {
	_, _, server := createSchemaService(t, &testSchemaSeviceConfig{
		schemaDir: "/home/phuong/projects/fastschema/fastschema/data/schemas",
		extraSchemas: map[string]string{
			"category": categorySchemaJSON,
		},
	})

	// Case 1: invalid payload
	req := httptest.NewRequest("PUT", "/schema/product", nil)
	resp := utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 500, resp.StatusCode)
	response := utils.Must(utils.ReadCloserToString(resp.Body))
	assert.Contains(t, response, `"error":{`)

	// Case 2: invalid schema
	req = httptest.NewRequest("PUT", "/schema/product", bytes.NewReader([]byte(`{"schema":{}}`)))
	resp = utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 404, resp.StatusCode)
	response = utils.Must(utils.ReadCloserToString(resp.Body))
	assert.Contains(t, response, `schema product not found`)

	// Case 3: update data is empty
	req = httptest.NewRequest("PUT", "/schema/category", bytes.NewReader([]byte(`{}`)))
	resp = utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 400, resp.StatusCode)
	response = utils.Must(utils.ReadCloserToString(resp.Body))
	assert.Contains(t, response, `schema update data is required`)
}

func TestSchemaServiceUpdateSuccess(t *testing.T) {
	newJSON := testBlogJSON
	testApp, _, server := createSchemaService(t, &testSchemaSeviceConfig{
		schemaDir: "/home/phuong/projects/fastschema/fastschema/data/schemas",
		extraSchemas: map[string]string{
			"blog": newJSON,
		},
	})

	// Case 1: add normal field
	newJSON = strings.ReplaceAll(
		newJSON,
		`"fields": [`,
		`"fields": [{"type": "string","name": "description","label": "Description","sortable": true},`,
	)
	req := httptest.NewRequest(
		"PUT", "/schema/blog",
		bytes.NewReader([]byte(
			fmt.Sprintf(`{"schema":%s}`, newJSON),
		)),
	)
	resp := utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 200, resp.StatusCode)
	response := utils.Must(utils.ReadCloserToString(resp.Body))
	assert.NotEmpty(t, response)
	assert.Equal(t, "description", utils.Must(testApp.Schema("blog").Field("description")).Name)

	// Case 2: add relation field
	newJSON = strings.ReplaceAll(
		newJSON,
		`"fields": [`,
		`"fields": [{
			"type": "relation",
			"name": "categories",
			"label": "Categories",
			"relation": {
				"schema": "category",
				"field": "blogs",
				"type": "m2m",
				"owner": false,
				"optional": false
			}
		},`,
	)
	req = httptest.NewRequest(
		"PUT", "/schema/blog",
		bytes.NewReader([]byte(
			fmt.Sprintf(`{"schema":%s}`, newJSON),
		)),
	)
	resp = utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 200, resp.StatusCode)
	response = utils.Must(utils.ReadCloserToString(resp.Body))
	assert.NotEmpty(t, response)

	blogFieldCategories := utils.Must(testApp.Schema("blog").Field("categories"))
	assert.Equal(t, "categories", blogFieldCategories.Name)
	assert.Equal(t, schema.M2M, blogFieldCategories.Relation.Type)
	assert.Equal(t, "category", blogFieldCategories.Relation.TargetSchemaName)
	assert.Equal(t, "blogs", blogFieldCategories.Relation.TargetFieldName)
	assert.False(t, blogFieldCategories.Relation.Owner)
	assert.False(t, blogFieldCategories.Relation.Optional)

	categoryFieldBlogs := utils.Must(testApp.Schema("category").Field("blogs"))
	assert.Equal(t, "blogs", categoryFieldBlogs.Name)
	assert.Equal(t, schema.M2M, categoryFieldBlogs.Relation.Type)
	assert.Equal(t, "blog", categoryFieldBlogs.Relation.TargetSchemaName)
	assert.Equal(t, "categories", categoryFieldBlogs.Relation.TargetFieldName)
	assert.True(t, categoryFieldBlogs.Relation.Owner)
	assert.True(t, categoryFieldBlogs.Relation.Optional)

	// Case 3: update schema name
	newJSON = strings.ReplaceAll(newJSON, `"name": "category"`, `"name": "cat"`)
	req = httptest.NewRequest(
		"PUT", "/schema/category",
		bytes.NewReader([]byte(
			fmt.Sprintf(`{"schema":%s}`, newJSON),
		)),
	)
	resp = utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 200, resp.StatusCode)
	response = utils.Must(utils.ReadCloserToString(resp.Body))
	assert.Contains(t, response, `"name":"cat"`)
	assert.Equal(t, "cat", utils.Must(testApp.SchemaBuilder().Schema("cat")).Name)

	// Case 4: update schema namespace
	newJSON = strings.ReplaceAll(newJSON, `"namespace": "categories"`, `"namespace": "cats"`)
	req = httptest.NewRequest(
		"PUT", "/schema/cat",
		bytes.NewReader([]byte(
			fmt.Sprintf(`{"schema":%s}`, newJSON),
		)),
	)
	resp = utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 200, resp.StatusCode)
	response = utils.Must(utils.ReadCloserToString(resp.Body))
	assert.Contains(t, response, `"namespace":"cats"`)
	assert.Equal(t, "cats", utils.Must(testApp.SchemaBuilder().Schema("cat")).Namespace)
}

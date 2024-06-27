package schemaservice_test

import (
	"bytes"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fastschema/fastschema/pkg/utils"
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
	req = httptest.NewRequest("POST", "/schema", bytes.NewReader([]byte(testBlogJSON)))
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
	newBlogJSON := strings.ReplaceAll(
		testBlogJSON,
		`"fields": [`,
		`"fields": [{
			"type": "relation",
			"name": "categories",
			"label": "Categories",
			"relation": {
				"schema": "category",
				"field": "blogs",
				"type": "m2m",
				"owner": true,
				"optional": false
			}
		},`,
	)
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
			"blog": testBlogJSON,
			"tag":  testTagJSON,
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

package schemaservice_test

import (
	"bytes"
	"net/http/httptest"
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
	req = httptest.NewRequest("POST", "/schema", bytes.NewReader([]byte(`{
		"name": "category",
		"namespace": "categories",
		"label_field": "name",
		"disable_timestamp": false,
		"fields": [
			{
				"type": "string",
				"name": "name",
				"label": "Name",
				"unique": true,
				"sortable": true
			}
		]
	}`)))
	resp = utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 200, resp.StatusCode)
	response = utils.Must(utils.ReadCloserToString(resp.Body))
	assert.NotEmpty(t, response)

	req = httptest.NewRequest("DELETE", "/schema/category", nil)
	resp = utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 200, resp.StatusCode)
	response = utils.Must(utils.ReadCloserToString(resp.Body))
	assert.Contains(t, response, `Schema deleted`)

	// Case 3: schema has relation
	req = httptest.NewRequest("POST", "/schema", bytes.NewReader([]byte(`{
		"name": "category",
		"namespace": "categories",
		"label_field": "name",
		"disable_timestamp": false,
		"fields": [
			{
				"type": "string",
				"name": "name",
				"label": "Name",
				"unique": true,
				"sortable": true
			},
			{
				"type": "relation",
				"name": "blogs",
				"label": "Blogs",
				"relation": {
					"schema": "blog",
					"field": "categories",
					"type": "m2m",
					"owner": true,
					"optional": false
				}
			}
		]
	}`)))
	resp = utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 200, resp.StatusCode)
	response = utils.Must(utils.ReadCloserToString(resp.Body))
	assert.NotEmpty(t, response)

	req = httptest.NewRequest("DELETE", "/schema/category", nil)
	resp = utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 400, resp.StatusCode)
	response = utils.Must(utils.ReadCloserToString(resp.Body))
	assert.Contains(t, response, `schema has relation, can't delete`)
}

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
}`

func TestSchemaServiceUpdateError(t *testing.T) {
	_, _, server := createSchemaService(t, map[string]string{
		"category": categorySchemaJSON,
	})

	// Case 1: invalid payload
	req := httptest.NewRequest("PUT", "/schema/product", nil)
	resp := utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 500, resp.StatusCode)
	response := utils.Must(utils.ReadCloserToString(resp.Body))
	assert.Contains(t, response, `"error":{`)

	// Case 2: invalid schema
	req = httptest.NewRequest("PUT", "/schema/product", bytes.NewReader([]byte(`{}`)))
	resp = utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 404, resp.StatusCode)
	response = utils.Must(utils.ReadCloserToString(resp.Body))
	assert.Contains(t, response, `schema product not found`)

	// Case 3: update data is empty
	req = httptest.NewRequest("PUT", "/schema/blog", bytes.NewReader([]byte(`{}`)))
	resp = utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 400, resp.StatusCode)
	response = utils.Must(utils.ReadCloserToString(resp.Body))
	assert.Contains(t, response, `schema update data is required`)
}

func TestSchemaServiceUpdateSuccess(t *testing.T) {
	newJSON := categorySchemaJSON
	testApp, _, server := createSchemaService(t, map[string]string{"category": newJSON})

	// Case 1: add normal field
	newJSON = strings.ReplaceAll(
		newJSON,
		`"fields": [`,
		`"fields": [{"type": "string","name": "description","label": "Description","sortable": true},`,
	)
	req := httptest.NewRequest(
		"PUT", "/schema/category",
		bytes.NewReader([]byte(
			fmt.Sprintf(`{"schema":%s}`, newJSON),
		)),
	)
	resp := utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 200, resp.StatusCode)
	response := utils.Must(utils.ReadCloserToString(resp.Body))
	assert.NotEmpty(t, response)
	assert.Equal(t, "description", utils.Must(testApp.Schema("category").Field("description")).Name)

	// Case 2: add relation field
	newJSON = strings.ReplaceAll(
		newJSON,
		`"fields": [`,
		`"fields": [{
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
		},`,
	)
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
	assert.NotEmpty(t, response)
	assert.Equal(t, "blogs", utils.Must(testApp.Schema("category").Field("blogs")).Name)
	assert.Equal(t, schema.M2M, utils.Must(testApp.Schema("category").Field("blogs")).Relation.Type)
	assert.Equal(t, "blog", utils.Must(testApp.Schema("category").Field("blogs")).Relation.TargetSchemaName)
	assert.Equal(t, "categories", utils.Must(testApp.Schema("category").Field("blogs")).Relation.TargetFieldName)
	assert.True(t, utils.Must(testApp.Schema("category").Field("blogs")).Relation.Owner)
	assert.False(t, utils.Must(testApp.Schema("category").Field("blogs")).Relation.Optional)

	// Case 2: update schema name
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

	catSchema := utils.Must(testApp.SchemaBuilder().Schema("cat"))
	assert.Equal(t, "cat", catSchema.Name)

	// // Case 2: update schema namespace
	// newJSON = strings.ReplaceAll(newJSON, `"namespace": "categories"`, `"namespace": "cats"`)
	// req = httptest.NewRequest(
	// 	"PUT", "/schema/category",
	// 	bytes.NewReader([]byte(
	// 		fmt.Sprintf(`{"schema":%s}`, newJSON),
	// 	)),
	// )
	// resp = utils.Must(server.Test(req))
	// defer func() { assert.NoError(t, resp.Body.Close()) }()
	// assert.Equal(t, 200, resp.StatusCode)
	// response = utils.Must(utils.ReadCloserToString(resp.Body))
	// assert.Contains(t, response, `"namespace":"cats"`)

	// catSchema = utils.Must(testApp.SchemaBuilder().Schema("cat"))
	// assert.Equal(t, "cat", catSchema.Name)
}

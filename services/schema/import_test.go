package schemaservice_test

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/fastschema/fastschema/pkg/utils"
)

func createTestSchemaFile(t *testing.T, schemaName string, schemaContent string) string {
	tmpFilePath := t.TempDir() + fmt.Sprintf("/%s.yaml", schemaName)

	f, err := os.Create(tmpFilePath)
	assert.NoError(t, err)
	n2, err := f.Write([]byte(schemaContent))
	fmt.Printf("wrote %d bytes\n", n2)
	assert.NoError(t, err)
	defer f.Close()

	return tmpFilePath
}
func createFileBody(t *testing.T, schemaName string, schemaContent string) (*multipart.Writer, *bytes.Buffer) {
	filePath := createTestSchemaFile(t, schemaName, schemaContent)
	body := new(bytes.Buffer)
	mw := multipart.NewWriter(body)
	file, err := os.Open(filePath)
	assert.NoError(t, err)

	w, err := mw.CreateFormFile("field", filePath)
	assert.NoError(t, err)
	_, err = io.Copy(w, file)
	assert.NoError(t, err)
	return mw, body
}

func TestSchemaServiceImport(t *testing.T) {
	_, _, server := createSchemaService(t, nil)
	mw, body := createFileBody(t, "category", `{"name": "category"}`)
	mw.Close()
	// Case 1: schema already exists
	req := httptest.NewRequest("POST", "/schema/import", body)
	req.Header.Add("Content-Type", mw.FormDataContentType())
	resp := utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 400, resp.StatusCode)
	response := utils.Must(utils.ReadCloserToString(resp.Body))
	assert.Contains(t, response, `schema already exists`)

	// Case 2: schema validation failed
	mw, body = createFileBody(t, "blog", `{"name": "blog"}`)
	mw.Close()
	req = httptest.NewRequest("POST", "/schema/import", body)
	req.Header.Add("Content-Type", mw.FormDataContentType())
	resp = utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()

	assert.Equal(t, 500, resp.StatusCode)
	response = utils.Must(utils.ReadCloserToString(resp.Body))
	assert.Contains(t, response, `schema validation error`)
	assert.Contains(t, response, `label_field is required`)
	assert.Contains(t, response, `namespace is required`)

	// Case 3: invalid relation schema
	newBlogJSON := strings.ReplaceAll(
		testBlogJSON,
		`"fields": [`,
		`"fields": [{
			"type": "relation",
			"name": "categories",
			"label": "Categories",
			"relation": {
				"schema": "cat",
				"field": "blogs",
				"type": "m2m",
				"owner": true,
				"optional": false
			}
		},`,
	)

	mw, body = createFileBody(t, "blog", newBlogJSON)
	mw.Close()
	req = httptest.NewRequest("POST", "/schema/import", body)
	req.Header.Add("Content-Type", mw.FormDataContentType())
	resp = utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 500, resp.StatusCode)
	response = utils.Must(utils.ReadCloserToString(resp.Body))
	assert.Contains(t, response, `relation node blog.categories: 'cat' is not found`)

	// Case 4: invalid back ref relation
	newBlogJSON = strings.ReplaceAll(
		testBlogJSON,
		`"fields": [`,
		`"fields": [{
			"type": "relation",
			"name": "categories",
			"label": "Categories",
			"relation": {
				"schema": "category_import",
				"field": "blogs",
				"type": "m2m",
				"owner": true,
				"optional": false
			}
		},`,
	)

	// import 2 schemas blog and category
	categoryPath := createTestSchemaFile(t, "category_import", testCategoryJSONToImport)
	blogPath := createTestSchemaFile(t, "blog", newBlogJSON)
	body = new(bytes.Buffer)
	mw = multipart.NewWriter(body)
	categoryFile, err := os.Open(categoryPath)
	assert.NoError(t, err)
	blogFile, err := os.Open(blogPath)
	assert.NoError(t, err)

	w, err := mw.CreateFormFile("field", categoryPath)

	assert.NoError(t, err)
	_, err = io.Copy(w, categoryFile)
	assert.NoError(t, err)

	w, err = mw.CreateFormFile("field", blogPath)
	assert.NoError(t, err)
	_, err = io.Copy(w, blogFile)
	assert.NoError(t, err)
	mw.Close()

	req = httptest.NewRequest("POST", "/schema/import", body)
	req.Header.Add("Content-Type", mw.FormDataContentType())
	resp = utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 500, resp.StatusCode)
	response = utils.Must(utils.ReadCloserToString(resp.Body))
	assert.NotEmpty(t, response)
	assert.Contains(t, response, `relation for blog.categories is not valid`)

	// Case 5: create schema successfully
	newBlogJSON = strings.ReplaceAll(
		testBlogJSON,
		`"fields": [`,
		`"fields": [{
			"type": "relation",
			"name": "categories",
			"label": "Categories",
			"relation": {
				"schema": "category_import",
				"field": "blogs",
				"type": "m2m",
				"owner": true,
				"optional": false
			}
		},`,
	)
	newCategoryJSON := strings.ReplaceAll(
		testCategoryJSONToImport,
		`"fields": [`,
		`"fields": [{
			"type": "relation",
			"name": "blogs",
			"label": "Blogs",
			"relation": {
				"schema": "blog",
				"field": "categories",
				"type": "m2m",
				"owner": false,
				"optional": false
			}
		},`,
	)

	// import 2 schemas blog and category
	categoryPath = createTestSchemaFile(t, "category_import", newCategoryJSON)
	blogPath = createTestSchemaFile(t, "blog", newBlogJSON)
	body = new(bytes.Buffer)
	mw = multipart.NewWriter(body)
	categoryFile, err = os.Open(categoryPath)
	assert.NoError(t, err)
	blogFile, err = os.Open(blogPath)
	assert.NoError(t, err)

	w, err = mw.CreateFormFile("field", categoryPath)

	assert.NoError(t, err)
	_, err = io.Copy(w, categoryFile)
	assert.NoError(t, err)

	w, err = mw.CreateFormFile("field", blogPath)
	assert.NoError(t, err)
	_, err = io.Copy(w, blogFile)
	assert.NoError(t, err)
	mw.Close()

	req = httptest.NewRequest("POST", "/schema/import", body)
	req.Header.Add("Content-Type", mw.FormDataContentType())
	resp = utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 200, resp.StatusCode)
	response = utils.Must(utils.ReadCloserToString(resp.Body))
	assert.NotEmpty(t, response)
	assert.Contains(t, response, `Schema imported`)
}

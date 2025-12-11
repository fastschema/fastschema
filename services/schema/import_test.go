package schemaservice_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"

	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
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

// schemaToYAML converts a schema object to YAML string using gopkg.in/yaml.v3
func schemaToYAML(s *schema.Schema) string {
	// We need to use yaml directly
	type aliasSchema schema.Schema
	yamlBytes, err := yaml.Marshal((*aliasSchema)(s))
	if err != nil {
		panic(err)
	}
	return string(yamlBytes)
}

// modifyAndConvertToYAML parses JSON schema, applies a modification function, and returns YAML
func modifyAndConvertToYAML(schemaJSON string, modifier func(*schema.Schema)) string {
	var s schema.Schema
	if err := json.Unmarshal([]byte(schemaJSON), &s); err != nil {
		panic(err)
	}
	modifier(&s)
	return schemaToYAML(&s)
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
	newBlogYAML := schemaToYAML(blogSchema)

	mw, body = createFileBody(t, "blog", newBlogYAML)
	mw.Close()
	req = httptest.NewRequest("POST", "/schema/import", body)
	req.Header.Add("Content-Type", mw.FormDataContentType())
	resp = utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 500, resp.StatusCode)
	response = utils.Must(utils.ReadCloserToString(resp.Body))
	assert.Contains(t, response, `relation node blog.categories: 'cat' is not found`)

	// Case 4: invalid back ref relation
	blogSchema = utils.Must(schema.NewSchemaFromYAML(testBlogYAML))
	categoriesField = &schema.Field{
		Type:  schema.TypeRelation,
		Name:  "categories",
		Label: "Categories",
		Relation: &schema.Relation{
			Type:             schema.M2M,
			TargetSchemaName: "category_import",
			TargetFieldName:  "blogs",
			Owner:            true,
			Optional:         false,
		},
	}
	blogSchema.Fields = append(blogSchema.Fields, categoriesField)
	newBlogYAML = schemaToYAML(blogSchema)

	// import 2 schemas blog and category
	categoryPath := createTestSchemaFile(t, "category_import", testCategoryYAMLToImport)
	blogPath := createTestSchemaFile(t, "blog", newBlogYAML)
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
	blogSchema = utils.Must(schema.NewSchemaFromYAML(testBlogYAML))
	categoriesField = &schema.Field{
		Type:  schema.TypeRelation,
		Name:  "categories",
		Label: "Categories",
		Relation: &schema.Relation{
			Type:             schema.M2M,
			TargetSchemaName: "category_import",
			TargetFieldName:  "blogs",
			Owner:            true,
			Optional:         false,
		},
	}
	blogSchema.Fields = append(blogSchema.Fields, categoriesField)
	newBlogYAML = schemaToYAML(blogSchema)

	categorySchema := utils.Must(schema.NewSchemaFromYAML(testCategoryYAMLToImport))
	blogsField := &schema.Field{
		Type:  schema.TypeRelation,
		Name:  "blogs",
		Label: "Blogs",
		Relation: &schema.Relation{
			Type:             schema.M2M,
			TargetSchemaName: "blog",
			TargetFieldName:  "categories",
			Owner:            false,
			Optional:         false,
		},
	}
	categorySchema.Fields = append(categorySchema.Fields, blogsField)
	newCategoryYAML := schemaToYAML(categorySchema)

	// import 2 schemas blog and category
	categoryPath = createTestSchemaFile(t, "category_import", newCategoryYAML)
	blogPath = createTestSchemaFile(t, "blog", newBlogYAML)
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

package schemaservice_test

import (
	"bytes"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fastschema/fastschema/app"
	"github.com/fastschema/fastschema/pkg/restresolver"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
	schemaservice "github.com/fastschema/fastschema/services/schema"
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

func createUpdateTest(t *testing.T) (*testApp, *schemaservice.SchemaService, *restresolver.Server) {
	return createSchemaService(t, &testSchemaSeviceConfig{
		schemaDir: "/home/phuong/projects/fastschema/fastschema/data/schemas",
		extraSchemas: map[string]string{
			"blog": testBlogJSON,
			"tag":  testTagJSON,
		},
	})
}

func addFieldDescriptionToBlog(t *testing.T, testApp *testApp, server *restresolver.Server) string {
	newBlogJSON := strings.ReplaceAll(
		testBlogJSON,
		`"fields": [`,
		fmt.Sprintf(`"fields": [%s,`, testBlogJSONFields["description"]),
	)
	req := httptest.NewRequest(
		"PUT", "/schema/blog",
		bytes.NewReader([]byte(fmt.Sprintf(`{"schema":%s}`, newBlogJSON))),
	)
	resp := utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 200, resp.StatusCode)
	response := utils.Must(utils.ReadCloserToString(resp.Body))
	assert.NotEmpty(t, response)
	assert.Equal(t, "description", utils.Must(testApp.Schema("blog").Field("description")).Name)

	return newBlogJSON
}

func addFieldCategoriesToBlog(t *testing.T, testApp *testApp, server *restresolver.Server) string {
	newJSON := strings.ReplaceAll(
		testBlogJSON,
		`"fields": [`,
		fmt.Sprintf(`"fields": [%s,`, testBlogJSONFields["categories"]),
	)
	req := httptest.NewRequest(
		"PUT", "/schema/blog",
		bytes.NewReader([]byte(fmt.Sprintf(`{"schema":%s}`, newJSON))),
	)
	resp := utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 200, resp.StatusCode)
	response := utils.Must(utils.ReadCloserToString(resp.Body))
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

	return newJSON
}

func addFieldCategoryToBlog(t *testing.T, testApp *testApp, server *restresolver.Server) string {
	newJSON := strings.ReplaceAll(
		testBlogJSON,
		`"fields": [`,
		fmt.Sprintf(`"fields": [%s,`, testBlogJSONFields["category"]),
	)
	req := httptest.NewRequest(
		"PUT", "/schema/blog",
		bytes.NewReader([]byte(fmt.Sprintf(`{"schema":%s}`, newJSON))),
	)
	resp := utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 200, resp.StatusCode)
	response := utils.Must(utils.ReadCloserToString(resp.Body))
	assert.NotEmpty(t, response)

	blogFieldCategory := utils.Must(testApp.Schema("blog").Field("category"))
	assert.Equal(t, "category", blogFieldCategory.Name)
	assert.Equal(t, schema.O2M, blogFieldCategory.Relation.Type)
	assert.Equal(t, "category", blogFieldCategory.Relation.TargetSchemaName)
	assert.Equal(t, "blogs", blogFieldCategory.Relation.TargetFieldName)
	assert.False(t, blogFieldCategory.Relation.Owner)
	assert.False(t, blogFieldCategory.Relation.Optional)

	categoryFieldBlogs := utils.Must(testApp.Schema("category").Field("blogs"))
	assert.Equal(t, "blogs", categoryFieldBlogs.Name)
	assert.Equal(t, schema.O2M, categoryFieldBlogs.Relation.Type)
	assert.Equal(t, "blog", categoryFieldBlogs.Relation.TargetSchemaName)
	assert.Equal(t, "category", categoryFieldBlogs.Relation.TargetFieldName)
	assert.True(t, categoryFieldBlogs.Relation.Owner)
	assert.True(t, categoryFieldBlogs.Relation.Optional)

	return newJSON
}

// Case 1: add relation field with back reference field existed
func TestSchemaServiceUpdateBackRefFieldExisted(t *testing.T) {
	_, _, server := createUpdateTest(t)

	blogFieldCategoriesJSON := testBlogJSONFields["categories"]
	blogFieldCategoriesJSON = strings.ReplaceAll(
		blogFieldCategoriesJSON,
		`"field": "blogs"`,
		`"field": "name"`,
	)
	invalidBlogJSON := strings.ReplaceAll(
		testBlogJSON,
		`"fields": [`,
		fmt.Sprintf(`"fields": [%s,`, blogFieldCategoriesJSON),
	)
	req := httptest.NewRequest(
		"PUT", "/schema/blog",
		bytes.NewReader([]byte(fmt.Sprintf(`{"schema":%s}`, invalidBlogJSON))),
	)
	resp := utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 400, resp.StatusCode)
	response := utils.Must(utils.ReadCloserToString(resp.Body))
	assert.Contains(t, response, `Invalid field 'blog.categories'. Target relation field 'name' already exist in schema 'category'`)
}

// Case 2: add relation field with invalid back reference schema
func TestSchemaServiceUpdateInvalidBackRefSchema(t *testing.T) {
	_, _, server := createUpdateTest(t)

	blogFieldCategoriesJSON := testBlogJSONFields["categories"]
	blogFieldCategoriesJSON = strings.ReplaceAll(
		blogFieldCategoriesJSON,
		`"schema": "category"`,
		`"schema": "invalid"`,
	)
	invalidBlogJSON := strings.ReplaceAll(
		testBlogJSON,
		`"fields": [`,
		fmt.Sprintf(`"fields": [%s,`, blogFieldCategoriesJSON),
	)
	req := httptest.NewRequest(
		"PUT", "/schema/blog",
		bytes.NewReader([]byte(fmt.Sprintf(`{"schema":%s}`, invalidBlogJSON))),
	)
	resp := utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 400, resp.StatusCode)
	response := utils.Must(utils.ReadCloserToString(resp.Body))
	assert.Contains(t, response, `relation target schema 'invalid' not found`)
}

// Case 3: remove relation field success
func TestSchemaServiceUpdateRemoveRelationFieldSuccess(t *testing.T) {
	testApp, _, server := createUpdateTest(t)
	addFieldCategoriesToBlog(t, testApp, server)
	req := httptest.NewRequest(
		"PUT", "/schema/blog",
		bytes.NewReader([]byte(fmt.Sprintf(`{"schema":%s}`, testBlogJSON))),
	)
	resp := utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 200, resp.StatusCode)
	response := utils.Must(utils.ReadCloserToString(resp.Body))
	assert.NotEmpty(t, response)
	fieldCategories, err := testApp.Schema("blog").Field("categories")
	assert.Nil(t, fieldCategories)
	assert.Error(t, err)
}

// Case 4: add normal field
func TestSchemaServiceUpdateAddNormalField(t *testing.T) {
	testApp, _, server := createUpdateTest(t)
	addFieldDescriptionToBlog(t, testApp, server)
}

// Case 5: update schema name that have external and internal relations
func TestSchemaServiceUpdateRenameSchema(t *testing.T) {
	testApp, _, server := createUpdateTest(t)
	newBlogJSON := addFieldCategoriesToBlog(t, testApp, server)
	newBlogJSON = strings.ReplaceAll(newBlogJSON, `"name": "blog"`, `"name": "post"`)
	req := httptest.NewRequest(
		"PUT", "/schema/blog",
		bytes.NewReader([]byte(fmt.Sprintf(`{"schema":%s}`, newBlogJSON))),
	)
	resp := utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 200, resp.StatusCode)
	response := utils.Must(utils.ReadCloserToString(resp.Body))
	assert.Contains(t, response, `"name":"post"`)
	assert.Equal(t, "post", utils.Must(testApp.SchemaBuilder().Schema("post")).Name)
}

// Case 6: update schema namespace
func TestSchemaServiceUpdateRenameNamespace(t *testing.T) {
	testApp, _, server := createUpdateTest(t)

	// rename blog to post
	newBlogJSON := addFieldCategoriesToBlog(t, testApp, server)
	newBlogJSON = strings.ReplaceAll(newBlogJSON, `"name": "blog"`, `"name": "post"`)
	req := httptest.NewRequest(
		"PUT", "/schema/blog",
		bytes.NewReader([]byte(fmt.Sprintf(`{"schema":%s}`, newBlogJSON))),
	)
	resp := utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 200, resp.StatusCode)

	postJSON := strings.ReplaceAll(newBlogJSON, `"namespace": "blogs"`, `"namespace": "posts"`)
	postJSON = strings.ReplaceAll(postJSON, `"schema": "blog"`, `"schema": "post"`)
	req = httptest.NewRequest(
		"PUT", "/schema/post",
		bytes.NewReader([]byte(fmt.Sprintf(`{"schema":%s}`, postJSON))),
	)
	resp = utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 200, resp.StatusCode)
	response := utils.Must(utils.ReadCloserToString(resp.Body))
	assert.Contains(t, response, `"namespace":"posts"`)
	assert.Equal(t, "posts", utils.Must(testApp.SchemaBuilder().Schema("post")).Namespace)
}

// Case 7:add and remove multiple fields
func TestSchemaServiceUpdate(t *testing.T) {
	testApp, _, server := createUpdateTest(t)

	// add the field description and categories
	addFieldDescriptionToBlog(t, testApp, server)
	addFieldCategoriesToBlog(t, testApp, server)

	// Case 8: Remove field description, categories. Add fields note, tags
	blogJSON := strings.ReplaceAll(
		testBlogJSON,
		`"fields": [`,
		fmt.Sprintf(`"fields": [%s, %s,`, testBlogJSONFields["note"], testBlogJSONFields["tags"]),
	)

	req := httptest.NewRequest(
		"PUT", "/schema/blog",
		bytes.NewReader([]byte(fmt.Sprintf(`{"schema":%s}`, blogJSON))),
	)

	resp := utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 200, resp.StatusCode)
	response := utils.Must(utils.ReadCloserToString(resp.Body))
	assert.NotEmpty(t, response)

	fieldDescription, err := testApp.Schema("blog").Field("description")
	assert.Nil(t, fieldDescription)
	assert.Error(t, err)
	fieldCategories, err := testApp.Schema("blog").Field("categories")
	assert.Nil(t, fieldCategories)
	assert.Error(t, err)

	fieldNote := utils.Must(testApp.Schema("blog").Field("note"))
	assert.Equal(t, "note", fieldNote.Name)

	fieldTags := utils.Must(testApp.Schema("blog").Field("tags"))
	assert.Equal(t, "tags", fieldTags.Name)
	assert.Equal(t, schema.M2M, fieldTags.Relation.Type)
	assert.Equal(t, "tag", fieldTags.Relation.TargetSchemaName)
	assert.Equal(t, "blogs", fieldTags.Relation.TargetFieldName)
	assert.False(t, fieldTags.Relation.Owner)
	assert.False(t, fieldTags.Relation.Optional)

	fieldBlogs := utils.Must(testApp.Schema("tag").Field("blogs"))
	assert.Equal(t, "blogs", fieldBlogs.Name)
	assert.Equal(t, schema.M2M, fieldBlogs.Relation.Type)
	assert.Equal(t, "blog", fieldBlogs.Relation.TargetSchemaName)
	assert.Equal(t, "tags", fieldBlogs.Relation.TargetFieldName)
	assert.True(t, fieldBlogs.Relation.Owner)
	assert.True(t, fieldBlogs.Relation.Optional)
}

// Case 8: rename normal field
func TestSchemaServiceUpdateRenameNormalField(t *testing.T) {
	checkMigration := false
	testApp, _, server := createSchemaService(t, &testSchemaSeviceConfig{
		schemaDir: "/home/phuong/projects/fastschema/fastschema/data/schemas",
		extraSchemas: map[string]string{
			"blog": testBlogJSON,
			"tag":  testTagJSON,
		},
		reloadFn: func(migrations *app.Migration) error {
			if !checkMigration {
				return nil
			}
			assert.Len(t, migrations.RenameFields, 1)
			assert.Equal(t, "column", migrations.RenameFields[0].Type)
			assert.Equal(t, "description", migrations.RenameFields[0].From)
			assert.Equal(t, "desc", migrations.RenameFields[0].To)
			assert.Equal(t, "blog", migrations.RenameFields[0].SchemaName)
			assert.Equal(t, "blogs", migrations.RenameFields[0].SchemaNamespace)

			return nil
		},
	})
	newBlogJSON := addFieldDescriptionToBlog(t, testApp, server)

	checkMigration = true
	newBlogJSON = strings.ReplaceAll(
		newBlogJSON,
		`"name": "description"`,
		`"name": "desc"`,
	)
	req := httptest.NewRequest(
		"PUT", "/schema/blog",
		bytes.NewReader([]byte(fmt.Sprintf(`{
			"schema":%s,
			"rename_fields": [{
				"from": "description",
				"to": "desc"
			}]
		}`, newBlogJSON))),
	)
	resp := utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 200, resp.StatusCode)
	response := utils.Must(utils.ReadCloserToString(resp.Body))
	assert.NotEmpty(t, response)
	assert.Equal(t, "desc", utils.Must(testApp.Schema("blog").Field("desc")).Name)
}

// Case 9: rename m2m relation field
func TestSchemaServiceUpdateRenameO2MRelationField(t *testing.T) {
	checkMigration := false
	testApp, _, server := createSchemaService(t, &testSchemaSeviceConfig{
		schemaDir: "/home/phuong/projects/fastschema/fastschema/data/schemas",
		extraSchemas: map[string]string{
			"blog": testBlogJSON,
			"tag":  testTagJSON,
		},
		reloadFn: func(migrations *app.Migration) error {
			if !checkMigration {
				return nil
			}

			assert.Len(t, migrations.RenameFields, 1)
			assert.Equal(t, "column", migrations.RenameFields[0].Type)
			assert.Equal(t, "category_id", migrations.RenameFields[0].From)
			assert.Equal(t, "main_category_id", migrations.RenameFields[0].To)
			assert.Equal(t, "blog", migrations.RenameFields[0].SchemaName)
			assert.Equal(t, "blogs", migrations.RenameFields[0].SchemaNamespace)

			return nil
		},
	})
	newBlogJSON := addFieldCategoryToBlog(t, testApp, server)

	checkMigration = true
	newBlogJSON = strings.ReplaceAll(
		newBlogJSON,
		`"name": "category"`,
		`"name": "main_category"`,
	)
	req := httptest.NewRequest(
		"PUT", "/schema/blog",
		bytes.NewReader([]byte(fmt.Sprintf(`{
			"schema":%s,
			"rename_fields": [{
				"from": "category",
				"to": "main_category"
			}]
		}`, newBlogJSON))),
	)
	resp := utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 200, resp.StatusCode)
	response := utils.Must(utils.ReadCloserToString(resp.Body))
	assert.NotEmpty(t, response)
	assert.Equal(t, "main_category", utils.Must(testApp.Schema("blog").Field("main_category")).Name)
}

// Case 10: rename m2m relation field
func TestSchemaServiceUpdateRenameM2MRelationField(t *testing.T) {
	checkMigration := false
	testApp, _, server := createSchemaService(t, &testSchemaSeviceConfig{
		schemaDir: "/home/phuong/projects/fastschema/fastschema/data/schemas",
		extraSchemas: map[string]string{
			"blog": testBlogJSON,
			"tag":  testTagJSON,
		},
		reloadFn: func(migrations *app.Migration) error {
			if !checkMigration {
				return nil
			}

			assert.Len(t, migrations.RenameFields, 1)
			assert.Equal(t, "column", migrations.RenameFields[0].Type)
			assert.Equal(t, "categories", migrations.RenameFields[0].From)
			assert.Equal(t, "cats", migrations.RenameFields[0].To)
			assert.Equal(t, "blogs_categories", migrations.RenameFields[0].SchemaName)
			assert.Equal(t, "blogs_categories", migrations.RenameFields[0].SchemaNamespace)

			assert.Len(t, migrations.RenameTables, 1)
			assert.Equal(t, "table", migrations.RenameTables[0].Type)
			assert.Equal(t, "blogs_categories", migrations.RenameTables[0].From)
			assert.Equal(t, "blogs_cats", migrations.RenameTables[0].To)
			assert.Equal(t, true, migrations.RenameTables[0].IsJunctionTable)
			assert.Equal(t, "", migrations.RenameTables[0].SchemaName)
			assert.Equal(t, "", migrations.RenameTables[0].SchemaNamespace)

			return nil
		},
	})
	newBlogJSON := addFieldCategoriesToBlog(t, testApp, server)

	checkMigration = true
	newBlogJSON = strings.ReplaceAll(
		newBlogJSON,
		`"name": "categories"`,
		`"name": "cats"`,
	)
	req := httptest.NewRequest(
		"PUT", "/schema/blog",
		bytes.NewReader([]byte(fmt.Sprintf(`{
			"schema":%s,
			"rename_fields": [{
				"from": "categories",
				"to": "cats"
			}]
		}`, newBlogJSON))),
	)
	resp := utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 200, resp.StatusCode)
	response := utils.Must(utils.ReadCloserToString(resp.Body))
	assert.NotEmpty(t, response)
	assert.Equal(t, "cats", utils.Must(testApp.Schema("blog").Field("cats")).Name)
}

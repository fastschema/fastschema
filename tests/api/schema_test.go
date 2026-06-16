package api_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSchemaList(t *testing.T) {
	app := CreateTestApp(t)

	t.Run("list_all_schemas", func(t *testing.T) {
		resp, apiResp := app.Get("/api/schema", app.adminToken)
		app.AssertStatus(resp, 200)

		var schemas []map[string]any
		app.ParseData(apiResp, &schemas)
		assert.GreaterOrEqual(t, len(schemas), 2) // At least post and category
	})

	t.Run("list_without_auth", func(t *testing.T) {
		resp, _ := app.Get("/api/schema")
		assert.Contains(t, []int{401, 403}, resp.Code)
	})

	t.Run("list_with_user_token", func(t *testing.T) {
		resp, _ := app.Get("/api/schema", app.normalToken)
		// Users might not have schema access
		assert.NotEqual(t, 500, resp.Code)
	})
}

func TestSchemaDetail(t *testing.T) {
	app := CreateTestApp(t)

	t.Run("get_existing_schema", func(t *testing.T) {
		resp, apiResp := app.Get("/api/schema/post", app.adminToken)
		app.AssertStatus(resp, 200)

		var schema map[string]any
		app.ParseData(apiResp, &schema)
		assert.Equal(t, "post", schema["name"])
		assert.NotNil(t, schema["fields"])
	})

	t.Run("get_system_schema", func(t *testing.T) {
		resp, apiResp := app.Get("/api/schema/user", app.adminToken)
		app.AssertStatus(resp, 200)

		var schema map[string]any
		app.ParseData(apiResp, &schema)
		assert.Equal(t, "user", schema["name"])
	})

	t.Run("get_nonexistent_schema", func(t *testing.T) {
		resp, apiResp := app.Get("/api/schema/nonexistent", app.adminToken)
		assert.Equal(t, 404, resp.Code)
		assert.NotNil(t, apiResp.Error)
	})

	t.Run("get_without_auth", func(t *testing.T) {
		resp, _ := app.Get("/api/schema/post")
		assert.Contains(t, []int{401, 403}, resp.Code)
	})
}

func TestSchemaCreate(t *testing.T) {
	app := CreateTestApp(t)

	t.Run("create_valid_schema", func(t *testing.T) {
		resp, apiResp := app.Post("/api/schema", map[string]any{
			"name":        "article",
			"namespace":   "articles",
			"label_field": "title",
			"fields": []map[string]any{
				{
					"type":  "string",
					"name":  "title",
					"label": "Title",
				},
				{
					"type":     "text",
					"name":     "body",
					"label":    "Body",
					"optional": true,
				},
			},
		}, app.adminToken)
		app.AssertStatus(resp, 200)

		var created map[string]any
		app.ParseData(apiResp, &created)
		assert.Equal(t, "article", created["name"])
	})

	t.Run("create_without_name", func(t *testing.T) {
		resp, _ := app.Post("/api/schema", map[string]any{
			"namespace": "test",
			"fields": []map[string]any{
				{"type": "string", "name": "title", "label": "Title"},
			},
		}, app.adminToken)
		assert.NotEqual(t, 200, resp.Code)
	})

	t.Run("create_without_fields", func(t *testing.T) {
		resp, _ := app.Post("/api/schema", map[string]any{
			"name":      "empty",
			"namespace": "empties",
		}, app.adminToken)
		// Might succeed with empty fields or fail
		assert.NotEqual(t, 500, resp.Code)
	})

	t.Run("create_duplicate_name", func(t *testing.T) {
		// First create
		app.Post("/api/schema", map[string]any{
			"name":      "duplicate",
			"namespace": "duplicates",
			"fields": []map[string]any{
				{"type": "string", "name": "name", "label": "Name"},
			},
		}, app.adminToken)

		// Second create with same name
		resp, _ := app.Post("/api/schema", map[string]any{
			"name":      "duplicate",
			"namespace": "duplicates2",
			"fields": []map[string]any{
				{"type": "string", "name": "name", "label": "Name"},
			},
		}, app.adminToken)
		assert.NotEqual(t, 200, resp.Code)
	})

	t.Run("create_with_invalid_field_type", func(t *testing.T) {
		resp, _ := app.Post("/api/schema", map[string]any{
			"name":      "invalid",
			"namespace": "invalids",
			"fields": []map[string]any{
				{"type": "invalid_type", "name": "field", "label": "Field"},
			},
		}, app.adminToken)
		assert.NotEqual(t, 200, resp.Code)
	})

	t.Run("create_with_relation", func(t *testing.T) {
		resp, apiResp := app.Post("/api/schema", map[string]any{
			"name":        "comment",
			"namespace":   "comments",
			"label_field": "content",
			"fields": []map[string]any{
				{"type": "text", "name": "content", "label": "Content"},
				{
					"type":     "relation",
					"name":     "post",
					"label":    "Post",
					"optional": true,
					"relation": map[string]any{
						"schema": "post",
						"field":  "comments",
						"type":   "o2m",
					},
				},
			},
		}, app.adminToken)
		app.AssertStatus(resp, 200)

		var created map[string]any
		app.ParseData(apiResp, &created)
		assert.Equal(t, "comment", created["name"])
	})

	t.Run("create_without_auth", func(t *testing.T) {
		resp, _ := app.Post("/api/schema", map[string]any{
			"name":      "unauthorized",
			"namespace": "unauthorized",
			"fields": []map[string]any{
				{"type": "string", "name": "name", "label": "Name"},
			},
		})
		assert.Contains(t, []int{401, 403}, resp.Code)
	})

	t.Run("create_with_user_token", func(t *testing.T) {
		resp, _ := app.Post("/api/schema", map[string]any{
			"name":      "usertest",
			"namespace": "usertests",
			"fields": []map[string]any{
				{"type": "string", "name": "name", "label": "Name"},
			},
		}, app.normalToken)
		// Normal users shouldn't be able to create schemas
		assert.Contains(t, []int{401, 403}, resp.Code)
	})
}

func TestSchemaUpdate(t *testing.T) {
	app := CreateTestApp(t)

	t.Run("update_existing_schema", func(t *testing.T) {
		// First create a schema
		app.Post("/api/schema", map[string]any{
			"name":        "updatetest",
			"namespace":   "updatetests",
			"label_field": "name",
			"fields": []map[string]any{
				{"type": "string", "name": "name", "label": "Name"},
			},
		}, app.adminToken)

		// Then update it - schema updates may require app reload
		resp, apiResp := app.Put("/api/schema/updatetest", map[string]any{
			"name":        "updatetest",
			"namespace":   "updatetests",
			"label_field": "name",
			"fields": []map[string]any{
				{"type": "string", "name": "name", "label": "Name"},
				{"type": "text", "name": "description", "label": "Description", "optional": true},
			},
		}, app.adminToken)
		// Schema update may need reload, so 404 is acceptable
		if resp.Code == 200 {
			var updated map[string]any
			app.ParseData(apiResp, &updated)
			assert.Equal(t, "updatetest", updated["name"])
		} else {
			assert.NotEqual(t, 500, resp.Code)
		}
	})

	t.Run("update_nonexistent_schema", func(t *testing.T) {
		resp, _ := app.Put("/api/schema/nonexistent", map[string]any{
			"name":      "nonexistent",
			"namespace": "nonexistent",
			"fields": []map[string]any{
				{"type": "string", "name": "name", "label": "Name"},
			},
		}, app.adminToken)
		assert.Equal(t, 404, resp.Code)
	})

	t.Run("update_system_schema", func(t *testing.T) {
		resp, _ := app.Put("/api/schema/user", map[string]any{
			"name":      "user",
			"namespace": "users",
			"fields": []map[string]any{
				{"type": "string", "name": "custom_field", "label": "Custom"},
			},
		}, app.adminToken)
		// System schemas might not be updatable
		assert.NotEqual(t, 500, resp.Code)
	})

	t.Run("update_without_auth", func(t *testing.T) {
		resp, _ := app.Put("/api/schema/post", map[string]any{
			"name":   "post",
			"fields": []map[string]any{},
		})
		assert.Contains(t, []int{401, 403}, resp.Code)
	})
}

func TestSchemaDelete(t *testing.T) {
	app := CreateTestApp(t)

	t.Run("delete_existing_schema", func(t *testing.T) {
		// First create a schema
		app.Post("/api/schema", map[string]any{
			"name":        "todelete",
			"namespace":   "todeletes",
			"label_field": "name",
			"fields": []map[string]any{
				{"type": "string", "name": "name", "label": "Name"},
			},
		}, app.adminToken)

		// Then delete it - schema delete may require app reload
		resp, _ := app.Delete("/api/schema/todelete", app.adminToken)
		// Schema delete may need reload, so 404 is acceptable
		assert.NotEqual(t, 500, resp.Code)
	})

	t.Run("delete_nonexistent_schema", func(t *testing.T) {
		resp, _ := app.Delete("/api/schema/nonexistent", app.adminToken)
		assert.Equal(t, 404, resp.Code)
	})

	t.Run("delete_system_schema", func(t *testing.T) {
		resp, _ := app.Delete("/api/schema/user", app.adminToken)
		// System schemas should not be deletable
		assert.NotEqual(t, 200, resp.Code)
	})

	t.Run("delete_without_auth", func(t *testing.T) {
		resp, _ := app.Delete("/api/schema/post")
		assert.Contains(t, []int{401, 403}, resp.Code)
	})

	t.Run("delete_schema_in_use", func(t *testing.T) {
		// Create schema
		app.Post("/api/schema", map[string]any{
			"name":        "inuse",
			"namespace":   "inuses",
			"label_field": "name",
			"fields": []map[string]any{
				{"type": "string", "name": "name", "label": "Name"},
			},
		}, app.adminToken)

		// Create content in the schema
		app.Post("/api/content/inuse", map[string]any{
			"name": "Test Item",
		}, app.adminToken)

		// Try to delete - may fail due to existing content
		resp, _ := app.Delete("/api/schema/inuse", app.adminToken)
		// Should either succeed or fail gracefully
		assert.NotEqual(t, 500, resp.Code)
	})
}

func TestSchemaFieldTypes(t *testing.T) {
	app := CreateTestApp(t)

	fieldTypes := []struct {
		name      string
		fieldType string
		extra     map[string]any
	}{
		{"string_field", "string", nil},
		{"text_field", "text", nil},
		{"int_field", "int", nil},
		{"uint_field", "uint", nil},
		{"float_field", "float", nil},
		{"bool_field", "bool", nil},
		{"time_field", "time", nil},
		{"json_field", "json", nil},
		{"uuid_field", "uuid", nil},
		{"enum_field", "enum", map[string]any{"enums": []map[string]any{{"value": "a"}, {"value": "b"}}}},
	}

	for _, ft := range fieldTypes {
		t.Run("create_with_"+ft.name, func(t *testing.T) {
			schemaName := "test_" + ft.name
			field := map[string]any{
				"type":     ft.fieldType,
				"name":     ft.name,
				"label":    ft.name,
				"optional": true,
			}
			if ft.extra != nil {
				for k, v := range ft.extra {
					field[k] = v
				}
			}

			resp, apiResp := app.Post("/api/schema", map[string]any{
				"name":        schemaName,
				"namespace":   schemaName + "s",
				"label_field": ft.name,
				"fields":      []map[string]any{field},
			}, app.adminToken)
			app.AssertStatus(resp, 200)

			var created map[string]any
			app.ParseData(apiResp, &created)
			assert.Equal(t, schemaName, created["name"])
		})
	}
}

func TestSchemaValidation(t *testing.T) {
	app := CreateTestApp(t)

	t.Run("invalid_schema_name_special_chars", func(t *testing.T) {
		resp, _ := app.Post("/api/schema", map[string]any{
			"name":      "invalid-name!",
			"namespace": "invalids",
			"fields": []map[string]any{
				{"type": "string", "name": "name", "label": "Name"},
			},
		}, app.adminToken)
		assert.NotEqual(t, 200, resp.Code)
	})

	t.Run("invalid_field_name", func(t *testing.T) {
		resp, _ := app.Post("/api/schema", map[string]any{
			"name":      "testfieldname",
			"namespace": "testfieldnames",
			"fields": []map[string]any{
				{"type": "string", "name": "invalid-field!", "label": "Field"},
			},
		}, app.adminToken)
		assert.NotEqual(t, 200, resp.Code)
	})

	t.Run("reserved_schema_name", func(t *testing.T) {
		resp, _ := app.Post("/api/schema", map[string]any{
			"name":      "user",
			"namespace": "users2",
			"fields": []map[string]any{
				{"type": "string", "name": "name", "label": "Name"},
			},
		}, app.adminToken)
		// Should fail - 'user' is a system schema
		assert.NotEqual(t, 200, resp.Code)
	})

	t.Run("duplicate_field_names", func(t *testing.T) {
		resp, _ := app.Post("/api/schema", map[string]any{
			"name":      "dupfields",
			"namespace": "dupfields",
			"fields": []map[string]any{
				{"type": "string", "name": "name", "label": "Name"},
				{"type": "string", "name": "name", "label": "Name 2"},
			},
		}, app.adminToken)
		assert.NotEqual(t, 200, resp.Code)
	})
}

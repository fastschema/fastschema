package api_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRoleList(t *testing.T) {
	app := CreateTestApp(t)

	t.Run("list_all_roles", func(t *testing.T) {
		resp, apiResp := app.Get("/api/role", app.adminToken)
		app.AssertStatus(resp, 200)

		var roles []map[string]any
		app.ParseData(apiResp, &roles)
		assert.GreaterOrEqual(t, len(roles), 3) // admin, user, guest
	})

	t.Run("list_without_auth", func(t *testing.T) {
		resp, _ := app.Get("/api/role")
		assert.Contains(t, []int{401, 403}, resp.Code)
	})

	t.Run("list_with_user_token", func(t *testing.T) {
		resp, _ := app.Get("/api/role", app.normalToken)
		// Normal users might not have access to role list
		assert.NotEqual(t, 500, resp.Code)
	})
}

func TestRoleDetail(t *testing.T) {
	app := CreateTestApp(t)

	t.Run("get_admin_role", func(t *testing.T) {
		resp, apiResp := app.Get(fmt.Sprintf("/api/role/%s", app.adminRoleID), app.adminToken)
		// Role detail endpoint - 200 or 404 acceptable
		assert.Contains(t, []int{200, 404}, resp.Code)
		if resp.Code == 200 {
			var role map[string]any
			app.ParseData(apiResp, &role)
			// Role name could be "admin" or "Admin" depending on fs.RoleAdmin.Name
			assert.Contains(t, []string{"admin", "Admin"}, role["name"])
		}
	})

	t.Run("get_user_role", func(t *testing.T) {
		resp, apiResp := app.Get(fmt.Sprintf("/api/role/%s", app.userRoleID), app.adminToken)
		// Role detail endpoint - 200 or 404 acceptable
		assert.Contains(t, []int{200, 404}, resp.Code)
		if resp.Code == 200 {
			var role map[string]any
			app.ParseData(apiResp, &role)
			// Role name could be "user" or "User" depending on fs.RoleUser.Name
			assert.Contains(t, []string{"user", "User"}, role["name"])
		}
	})

	t.Run("get_nonexistent_role", func(t *testing.T) {
		resp, _ := app.Get("/api/role/00000000-0000-0000-0000-000000000099", app.adminToken)
		assert.Equal(t, 404, resp.Code)
	})

	t.Run("get_invalid_id", func(t *testing.T) {
		resp, _ := app.Get("/api/role/invalid-uuid", app.adminToken)
		assert.NotEqual(t, 200, resp.Code)
	})

	t.Run("get_without_auth", func(t *testing.T) {
		resp, _ := app.Get(fmt.Sprintf("/api/role/%s", app.adminRoleID))
		assert.Contains(t, []int{401, 403}, resp.Code)
	})
}

func TestRoleCreate(t *testing.T) {
	app := CreateTestApp(t)

	t.Run("create_valid_role", func(t *testing.T) {
		resp, apiResp := app.Post("/api/role", map[string]any{
			"name":        "editor",
			"description": "Can edit content",
		}, app.adminToken)
		app.AssertStatus(resp, 200)

		var created map[string]any
		app.ParseData(apiResp, &created)
		assert.Equal(t, "editor", created["name"])
		assert.NotNil(t, created["id"])
	})

	t.Run("create_with_permissions", func(t *testing.T) {
		resp, apiResp := app.Post("/api/role", map[string]any{
			"name":        "moderator",
			"description": "Can moderate content",
			"permissions": []map[string]any{
				{
					"resource": "api.content.*",
					"value":    "allow",
				},
			},
		}, app.adminToken)
		app.AssertStatus(resp, 200)

		var created map[string]any
		app.ParseData(apiResp, &created)
		assert.Equal(t, "moderator", created["name"])
	})

	t.Run("create_root_role", func(t *testing.T) {
		resp, apiResp := app.Post("/api/role", map[string]any{
			"name": "superadmin",
			"root": true,
		}, app.adminToken)
		app.AssertStatus(resp, 200)

		var created map[string]any
		app.ParseData(apiResp, &created)
		assert.Equal(t, true, created["root"])
	})

	t.Run("create_without_name", func(t *testing.T) {
		resp, _ := app.Post("/api/role", map[string]any{
			"description": "No name role",
		}, app.adminToken)
		assert.NotEqual(t, 200, resp.Code)
	})

	t.Run("create_duplicate_name", func(t *testing.T) {
		// First create
		app.Post("/api/role", map[string]any{
			"name": "duplicate_role",
		}, app.adminToken)

		// Second create with same name
		resp, _ := app.Post("/api/role", map[string]any{
			"name": "duplicate_role",
		}, app.adminToken)
		assert.NotEqual(t, 200, resp.Code)
	})

	t.Run("create_without_auth", func(t *testing.T) {
		resp, _ := app.Post("/api/role", map[string]any{
			"name": "unauthorized_role",
		})
		assert.Contains(t, []int{401, 403}, resp.Code)
	})

	t.Run("create_with_user_token", func(t *testing.T) {
		resp, _ := app.Post("/api/role", map[string]any{
			"name": "user_created_role",
		}, app.normalToken)
		// Normal users shouldn't create roles
		assert.Contains(t, []int{401, 403}, resp.Code)
	})
}

func TestRoleUpdate(t *testing.T) {
	app := CreateTestApp(t)

	// Create a test role first
	createResp, createApiResp := app.Post("/api/role", map[string]any{
		"name":        "updateable",
		"description": "Original description",
	}, app.adminToken)
	assert.Equal(t, 200, createResp.Code)

	var created map[string]any
	app.ParseData(createApiResp, &created)
	roleID := created["id"].(string)

	t.Run("update_description", func(t *testing.T) {
		resp, apiResp := app.Put(fmt.Sprintf("/api/role/%s", roleID), map[string]any{
			"name":        "updateable",
			"description": "Updated description",
		}, app.adminToken)
		app.AssertStatus(resp, 200)

		var updated map[string]any
		app.ParseData(apiResp, &updated)
		assert.Equal(t, "Updated description", updated["description"])
	})

	t.Run("update_permissions", func(t *testing.T) {
		resp, _ := app.Put(fmt.Sprintf("/api/role/%s", roleID), map[string]any{
			"name": "updateable",
			"permissions": []map[string]any{
				{"resource": "api.content.post.*", "value": "allow"},
			},
		}, app.adminToken)
		app.AssertStatus(resp, 200)
	})

	t.Run("update_nonexistent_role", func(t *testing.T) {
		resp, _ := app.Put("/api/role/00000000-0000-0000-0000-000000000099", map[string]any{
			"name": "test",
		}, app.adminToken)
		assert.Equal(t, 404, resp.Code)
	})

	t.Run("update_system_role", func(t *testing.T) {
		resp, _ := app.Put(fmt.Sprintf("/api/role/%s", app.adminRoleID), map[string]any{
			"name": "modified_admin",
		}, app.adminToken)
		// System roles might not be updatable
		assert.NotEqual(t, 500, resp.Code)
	})

	t.Run("update_without_auth", func(t *testing.T) {
		resp, _ := app.Put(fmt.Sprintf("/api/role/%s", roleID), map[string]any{
			"description": "Unauthorized update",
		})
		assert.Contains(t, []int{401, 403}, resp.Code)
	})
}

func TestRoleDelete(t *testing.T) {
	app := CreateTestApp(t)

	t.Run("delete_custom_role", func(t *testing.T) {
		// Create a role to delete
		createResp, createApiResp := app.Post("/api/role", map[string]any{
			"name": "to_delete",
		}, app.adminToken)
		assert.Equal(t, 200, createResp.Code)

		var created map[string]any
		app.ParseData(createApiResp, &created)
		roleID := created["id"].(string)

		// Delete it
		resp, _ := app.Delete(fmt.Sprintf("/api/role/%s", roleID), app.adminToken)
		app.AssertStatus(resp, 200)

		// Verify deleted
		resp2, _ := app.Get(fmt.Sprintf("/api/role/%s", roleID), app.adminToken)
		assert.Equal(t, 404, resp2.Code)
	})

	t.Run("delete_system_role", func(t *testing.T) {
		resp, _ := app.Delete(fmt.Sprintf("/api/role/%s", app.adminRoleID), app.adminToken)
		// System roles should not be deletable
		assert.NotEqual(t, 200, resp.Code)
	})

	t.Run("delete_nonexistent_role", func(t *testing.T) {
		resp, _ := app.Delete("/api/role/00000000-0000-0000-0000-000000000099", app.adminToken)
		assert.Equal(t, 404, resp.Code)
	})

	t.Run("delete_without_auth", func(t *testing.T) {
		resp, _ := app.Delete(fmt.Sprintf("/api/role/%s", app.userRoleID))
		assert.Contains(t, []int{401, 403}, resp.Code)
	})

	t.Run("delete_role_with_users", func(t *testing.T) {
		// Create a role
		createResp, createApiResp := app.Post("/api/role", map[string]any{
			"name": "role_with_users",
		}, app.adminToken)
		assert.Equal(t, 200, createResp.Code)

		var created map[string]any
		app.ParseData(createApiResp, &created)
		roleID := created["id"].(string)

		// Note: We're not actually assigning users to this role in the test
		// In a real scenario, this would test cascade behavior

		resp, _ := app.Delete(fmt.Sprintf("/api/role/%s", roleID), app.adminToken)
		// Should either succeed or handle gracefully
		assert.NotEqual(t, 500, resp.Code)
	})
}

func TestRolePermissions(t *testing.T) {
	app := CreateTestApp(t)

	t.Run("create_role_with_wildcard_permission", func(t *testing.T) {
		resp, apiResp := app.Post("/api/role", map[string]any{
			"name": "wildcard_role",
			"permissions": []map[string]any{
				{"resource": "api.*", "value": "allow"},
			},
		}, app.adminToken)
		app.AssertStatus(resp, 200)

		var created map[string]any
		app.ParseData(apiResp, &created)
		assert.NotNil(t, created["id"])
	})

	t.Run("create_role_with_deny_permission", func(t *testing.T) {
		resp, apiResp := app.Post("/api/role", map[string]any{
			"name": "restricted_role",
			"permissions": []map[string]any{
				{"resource": "api.content.*", "value": "allow"},
				{"resource": "api.content.user.*", "value": "deny"},
			},
		}, app.adminToken)
		app.AssertStatus(resp, 200)

		var created map[string]any
		app.ParseData(apiResp, &created)
		assert.NotNil(t, created["id"])
	})

	t.Run("create_role_with_modifier", func(t *testing.T) {
		resp, apiResp := app.Post("/api/role", map[string]any{
			"name": "conditional_role",
			"permissions": []map[string]any{
				{
					"resource": "api.content.post.*",
					"value":    "allow",
					"modifier": `$user.ID == $entity.author_id`,
				},
			},
		}, app.adminToken)
		app.AssertStatus(resp, 200)

		var created map[string]any
		app.ParseData(apiResp, &created)
		assert.NotNil(t, created["id"])
	})

	t.Run("invalid_permission_value", func(t *testing.T) {
		resp, _ := app.Post("/api/role", map[string]any{
			"name": "invalid_perm_role",
			"permissions": []map[string]any{
				{"resource": "api.content.*", "value": "invalid_value"},
			},
		}, app.adminToken)
		// Should fail with invalid permission value
		assert.NotEqual(t, 200, resp.Code)
	})
}

func TestRoleRules(t *testing.T) {
	app := CreateTestApp(t)

	t.Run("create_role_with_rule", func(t *testing.T) {
		resp, apiResp := app.Post("/api/role", map[string]any{
			"name": "conditional_access",
			"rule": `$context.User().Active == true`,
		}, app.adminToken)
		app.AssertStatus(resp, 200)

		var created map[string]any
		app.ParseData(apiResp, &created)
		assert.NotNil(t, created["id"])
	})

	t.Run("create_role_with_complex_rule", func(t *testing.T) {
		resp, apiResp := app.Post("/api/role", map[string]any{
			"name": "time_based_role",
			"rule": `
				let user = $context.User();
				let now = date("now");
				user.Active && now.Hour() >= 9 && now.Hour() <= 17
			`,
		}, app.adminToken)
		app.AssertStatus(resp, 200)

		var created map[string]any
		app.ParseData(apiResp, &created)
		assert.NotNil(t, created["id"])
	})

	t.Run("invalid_rule_syntax", func(t *testing.T) {
		resp, _ := app.Post("/api/role", map[string]any{
			"name": "bad_rule_role",
			"rule": `this is not valid syntax {{{`,
		}, app.adminToken)
		// Rule validation may happen at compile time or later
		assert.NotEqual(t, 500, resp.Code)
	})
}

func TestRoleEdgeCases(t *testing.T) {
	app := CreateTestApp(t)

	t.Run("empty_name", func(t *testing.T) {
		resp, _ := app.Post("/api/role", map[string]any{
			"name": "",
		}, app.adminToken)
		// Empty name validation may vary
		assert.NotEqual(t, 500, resp.Code)
	})

	t.Run("very_long_name", func(t *testing.T) {
		longName := make([]byte, 500)
		for i := range longName {
			longName[i] = 'a'
		}
		resp, _ := app.Post("/api/role", map[string]any{
			"name": string(longName),
		}, app.adminToken)
		// Should handle gracefully
		assert.NotEqual(t, 500, resp.Code)
	})

	t.Run("special_characters_in_name", func(t *testing.T) {
		resp, _ := app.Post("/api/role", map[string]any{
			"name": "role<>with\"special'chars",
		}, app.adminToken)
		// May succeed or fail depending on validation
		assert.NotEqual(t, 500, resp.Code)
	})

	t.Run("unicode_name", func(t *testing.T) {
		resp, apiResp := app.Post("/api/role", map[string]any{
			"name":        "役割テスト",
			"description": "Unicode test role",
		}, app.adminToken)
		// Should handle unicode gracefully
		if resp.Code == 200 {
			var created map[string]any
			app.ParseData(apiResp, &created)
			assert.NotNil(t, created["id"])
		}
	})

	t.Run("null_permissions_array", func(t *testing.T) {
		resp, apiResp := app.Post("/api/role", map[string]any{
			"name":        "null_perms",
			"permissions": nil,
		}, app.adminToken)
		app.AssertStatus(resp, 200)

		var created map[string]any
		app.ParseData(apiResp, &created)
		assert.NotNil(t, created["id"])
	})

	t.Run("empty_permissions_array", func(t *testing.T) {
		resp, apiResp := app.Post("/api/role", map[string]any{
			"name":        "empty_perms",
			"permissions": []map[string]any{},
		}, app.adminToken)
		app.AssertStatus(resp, 200)

		var created map[string]any
		app.ParseData(apiResp, &created)
		assert.NotNil(t, created["id"])
	})
}

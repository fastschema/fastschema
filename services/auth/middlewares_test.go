package authservice_test

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestRoleServiceAuth(t *testing.T) {
	testApp := createTestApp(t)
	server := testApp.restResolver.Server()
	t.Run("Test_ParseUser", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/testuser", nil)
		resp := utils.Must(server.Test(req))
		defer func() { assert.NoError(t, resp.Body.Close()) }()
		assert.Equal(t, 200, resp.StatusCode)
		response := utils.Must(utils.ReadCloserToString(resp.Body))
		assert.Equal(t, `{"data":null}`, response)

		req = httptest.NewRequest("GET", "/api/testuser", nil)
		req.Header.Set("Authorization", "Bearer "+testApp.adminToken)
		resp = utils.Must(server.Test(req))
		defer func() { assert.NoError(t, resp.Body.Close()) }()
		assert.Equal(t, 200, resp.StatusCode)
		response = utils.Must(utils.ReadCloserToString(resp.Body))
		assert.Contains(t, response, `"username":"adminuser"`)

		req = httptest.NewRequest("GET", "/api/testuser", nil)
		req.Header.Set("Authorization", "Bearer "+testApp.normalUserToken)
		resp = utils.Must(server.Test(req))
		defer func() { assert.NoError(t, resp.Body.Close()) }()
		assert.Equal(t, 200, resp.StatusCode)
		response = utils.Must(utils.ReadCloserToString(resp.Body))
		assert.Contains(t, response, `"username":"normaluser"`)
	})

	t.Run("Test_Authorize", func(t *testing.T) {
		// Admin user should have access to any resource without white list or permission set
		req := httptest.NewRequest("GET", "/api/content/blog", nil)
		req.Header.Set("Authorization", "Bearer "+testApp.adminToken)
		resp := utils.Must(server.Test(req))
		defer func() { assert.NoError(t, resp.Body.Close()) }()
		assert.Equal(t, 200, resp.StatusCode, "Admin user should have access to any resource without white list or permission set")
		assert.Equal(t, `{"data":"blog list"}`, utils.Must(utils.ReadCloserToString(resp.Body)))

		// Guest user should have access to white listed resource
		req = httptest.NewRequest("GET", "/api/test", nil)
		resp = utils.Must(server.Test(req))
		defer func() { assert.NoError(t, resp.Body.Close()) }()
		assert.Equal(t, 200, resp.StatusCode, "Guest user should have access to white listed resource")
		assert.Equal(t, `{"data":"test response"}`, utils.Must(utils.ReadCloserToString(resp.Body)))

		// Guest user should not have access to non white listed resource
		req = httptest.NewRequest("GET", "/api/content/blog", nil)
		resp = utils.Must(server.Test(req))
		defer func() { assert.NoError(t, resp.Body.Close()) }()
		assert.Equal(t, 401, resp.StatusCode, "Guest user should not have access to non white listed resource")
		assert.Contains(t, utils.Must(utils.ReadCloserToString(resp.Body)), `Unauthorized`)

		// Inactive user should not have access to any non white listed resource
		req = httptest.NewRequest("GET", "/api/content/blog", nil)
		req.Header.Set("Authorization", "Bearer "+testApp.inactiveUserToken)
		resp = utils.Must(server.Test(req))
		defer func() { assert.NoError(t, resp.Body.Close()) }()
		assert.Equal(t, 403, resp.StatusCode, "Inactive user should not have access to any non white listed resource")
		assert.Contains(t, utils.Must(utils.ReadCloserToString(resp.Body)), `User is inactive`)

		// Active user has these permissions:
		// - content.list: allow
		// - content.detail: deny
		// - content.meta: no permission set
		// - realtime.content.list: allow
		// - realtime.content.update: deny
		// - realtime.content.delete: no permission set
		// Expectation: user should have access to content.list but not content.detail and content.meta
		req = httptest.NewRequest("GET", "/api/content/blog", nil)
		req.Header.Set("Authorization", "Bearer "+testApp.normalUserToken)
		resp = utils.Must(server.Test(req))
		defer func() { assert.NoError(t, resp.Body.Close()) }()
		assert.Equal(t, 200, resp.StatusCode, "User should have access to content.blog.list")
		assert.Equal(t, `{"data":"blog list"}`, utils.Must(utils.ReadCloserToString(resp.Body)))

		req = httptest.NewRequest("GET", "/api/content/1", nil)
		req.Header.Set("Authorization", "Bearer "+testApp.normalUserToken)
		resp = utils.Must(server.Test(req))
		defer func() { assert.NoError(t, resp.Body.Close()) }()
		assert.Equal(t, 403, resp.StatusCode, "User should not have access to content.blog.detail")
		assert.Contains(t, utils.Must(utils.ReadCloserToString(resp.Body)), `Forbidden`)

		req = httptest.NewRequest("GET", "/api/content/blog/meta", nil)
		req.Header.Set("Authorization", "Bearer "+testApp.normalUserToken)
		resp = utils.Must(server.Test(req))
		defer func() { assert.NoError(t, resp.Body.Close()) }()
		assert.Equal(t, 403, resp.StatusCode, "User should not have access to content.blog.meta")
		assert.Contains(t, utils.Must(utils.ReadCloserToString(resp.Body)), `Forbidden`)

		// realtime.content.list: allow
		req = httptest.NewRequest("GET", "/api/realtime/content?schema=blog&event=list", nil)
		req.Header.Set("Authorization", "Bearer "+testApp.normalUserToken)
		req.Header.Set("Connection", "upgrade")
		req.Header.Set("Upgrade", "Websocket")

		resp = utils.Must(server.Test(req))
		defer func() { assert.NoError(t, resp.Body.Close()) }()
		assert.Equal(t, 200, resp.StatusCode, "User should have access to realtime.content.blog.list")
		assert.Equal(t, `{"data":"realtime content"}`, utils.Must(utils.ReadCloserToString(resp.Body)))

		// realtime.content.update: deny
		req = httptest.NewRequest("GET", "/api/realtime/content?schema=blog&event=update", nil)
		req.Header.Set("Authorization", "Bearer "+testApp.normalUserToken)
		req.Header.Set("Connection", "upgrade")
		req.Header.Set("Upgrade", "Websocket")

		resp = utils.Must(server.Test(req))
		defer func() { assert.NoError(t, resp.Body.Close()) }()
		assert.Equal(t, 403, resp.StatusCode, "User should not have access to realtime.content.blog.update")
		assert.Contains(t, utils.Must(utils.ReadCloserToString(resp.Body)), `Forbidden`)

		// realtime.content.delete: no permission set
		req = httptest.NewRequest("GET", "/api/realtime/content?schema=blog&event=delete", nil)
		req.Header.Set("Authorization", "Bearer "+testApp.normalUserToken)
		req.Header.Set("Connection", "upgrade")
		req.Header.Set("Upgrade", "Websocket")

		resp = utils.Must(server.Test(req))
		defer func() { assert.NoError(t, resp.Body.Close()) }()
		assert.Equal(t, 403, resp.StatusCode, "User should not have access to realtime.content.blog.update")
		assert.Contains(t, utils.Must(utils.ReadCloserToString(resp.Body)), `Forbidden`)

		// seniorityUser should have access to api.content.blog.list
		// because it's created_at is not satisfied created_at < date("2023-01-01")
		req = httptest.NewRequest("GET", "/api/content/blog", nil)
		req.Header.Set("Authorization", "Bearer "+testApp.seniorityUserToken)
		resp = utils.Must(server.Test(req))
		defer func() { assert.NoError(t, resp.Body.Close()) }()
		assert.Equal(t, 403, resp.StatusCode, "Non Seniority user should not have access to content.blog.list")

		// Update the created_at of seniorityUser to be less than 2023-01-01
		// seniorityUser should have access to api.content.blog.list
		_, err := testApp.db.Exec(context.Background(), "UPDATE users SET created_at = '2022-01-01' WHERE id = ?", testApp.seniorityUser.ID)
		assert.NoError(t, err)

		req = httptest.NewRequest("GET", "/api/content/blog", nil)
		req.Header.Set("Authorization", "Bearer "+testApp.seniorityUserToken)
		resp = utils.Must(server.Test(req))
		defer func() { assert.NoError(t, resp.Body.Close()) }()
		assert.Equal(t, 200, resp.StatusCode, "Seniority user should have access to content.blog.list")
	})
}

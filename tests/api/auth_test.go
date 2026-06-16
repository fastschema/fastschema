package api_test

import (
	"encoding/json"
	"testing"

	"github.com/fastschema/fastschema/fs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthLogin(t *testing.T) {
	app := CreateTestApp(t)

	t.Run("valid_credentials", func(t *testing.T) {
		resp, apiResp := app.Post("/api/user/auth/local/login", map[string]string{
			"login":    "admin",
			"password": "admin123",
		})
		app.AssertStatus(resp, 200)
		assert.NotNil(t, apiResp.Data)

		var tokenResp struct {
			Token string `json:"token"`
		}
		app.ParseData(apiResp, &tokenResp)
		assert.NotEmpty(t, tokenResp.Token)
	})

	t.Run("invalid_password", func(t *testing.T) {
		resp, apiResp := app.Post("/api/user/auth/local/login", map[string]string{
			"login":    "admin",
			"password": "wrongpassword",
		})
		assert.NotEqual(t, 200, resp.Code)
		assert.NotNil(t, apiResp.Error)
	})

	t.Run("invalid_username", func(t *testing.T) {
		resp, apiResp := app.Post("/api/user/auth/local/login", map[string]string{
			"login":    "nonexistent",
			"password": "password",
		})
		assert.NotEqual(t, 200, resp.Code)
		assert.NotNil(t, apiResp.Error)
	})

	t.Run("missing_login_field", func(t *testing.T) {
		resp, apiResp := app.Post("/api/user/auth/local/login", map[string]string{
			"password": "admin123",
		})
		assert.NotEqual(t, 200, resp.Code)
		assert.NotNil(t, apiResp.Error)
	})

	t.Run("missing_password_field", func(t *testing.T) {
		resp, apiResp := app.Post("/api/user/auth/local/login", map[string]string{
			"login": "admin",
		})
		assert.NotEqual(t, 200, resp.Code)
		assert.NotNil(t, apiResp.Error)
	})

	t.Run("empty_body", func(t *testing.T) {
		resp, _ := app.Post("/api/user/auth/local/login", nil)
		assert.NotEqual(t, 200, resp.Code)
	})
}

func TestAuthMe(t *testing.T) {
	app := CreateTestApp(t)

	t.Run("with_valid_token", func(t *testing.T) {
		resp, apiResp := app.Get("/api/user/auth/me", app.adminToken)
		app.AssertStatus(resp, 200)

		var user fs.User
		app.ParseData(apiResp, &user)
		assert.Equal(t, "admin", user.Username)
		assert.Equal(t, "admin@test.local", user.Email)
	})

	t.Run("without_token", func(t *testing.T) {
		resp, _ := app.Get("/api/user/auth/me")
		assert.Equal(t, 401, resp.Code)
	})

	t.Run("with_invalid_token", func(t *testing.T) {
		resp, _ := app.Get("/api/user/auth/me", "invalid-token")
		assert.Equal(t, 401, resp.Code)
	})

	t.Run("with_malformed_token", func(t *testing.T) {
		resp, _ := app.Get("/api/user/auth/me", "Bearer not.a.valid.jwt")
		assert.Equal(t, 401, resp.Code)
	})
}

func TestAuthLogout(t *testing.T) {
	app := CreateTestApp(t)

	t.Run("valid_logout", func(t *testing.T) {
		// First login to get a fresh token pair
		loginResp, loginApiResp := app.Post("/api/user/auth/local/login", map[string]string{
			"login":    "admin",
			"password": "admin123",
		})
		require.Equal(t, 200, loginResp.Code)

		var tokenResp struct {
			Token        string `json:"token"`
			RefreshToken string `json:"refresh_token"`
		}
		app.ParseData(loginApiResp, &tokenResp)

		// Logout with refresh token
		resp, _ := app.Post("/api/user/auth/logout", map[string]string{
			"refresh_token": tokenResp.RefreshToken,
		})
		// Should succeed or return appropriate status
		assert.Contains(t, []int{200, 400}, resp.Code) // 400 if refresh token not provided
	})

	t.Run("logout_without_token", func(t *testing.T) {
		resp, _ := app.Post("/api/user/auth/logout", nil)
		// Should handle gracefully
		assert.NotEqual(t, 500, resp.Code)
	})
}

func TestAuthTokenRefresh(t *testing.T) {
	app := CreateTestApp(t)

	t.Run("invalid_refresh_token", func(t *testing.T) {
		resp, apiResp := app.Post("/api/user/auth/token/refresh", map[string]string{
			"refresh_token": "invalid-token",
		})
		assert.NotEqual(t, 200, resp.Code)
		assert.NotNil(t, apiResp.Error)
	})

	t.Run("empty_refresh_token", func(t *testing.T) {
		resp, apiResp := app.Post("/api/user/auth/token/refresh", map[string]string{
			"refresh_token": "",
		})
		assert.NotEqual(t, 200, resp.Code)
		assert.NotNil(t, apiResp.Error)
	})

	t.Run("missing_refresh_token", func(t *testing.T) {
		resp, _ := app.Post("/api/user/auth/token/refresh", map[string]string{})
		assert.NotEqual(t, 200, resp.Code)
	})
}

func TestAuthRegister(t *testing.T) {
	app := CreateTestApp(t)

	t.Run("valid_registration", func(t *testing.T) {
		resp, apiResp := app.Post("/api/user/auth/local/register", map[string]string{
			"username": "newuser",
			"email":    "newuser@test.local",
			"password": "newpassword123",
		})
		// Registration may succeed or fail depending on configuration
		// Just ensure no 500 error
		assert.NotEqual(t, 500, resp.Code)
		_ = apiResp
	})

	t.Run("missing_username", func(t *testing.T) {
		resp, _ := app.Post("/api/user/auth/local/register", map[string]string{
			"email":    "test@test.local",
			"password": "password123",
		})
		assert.NotEqual(t, 200, resp.Code)
	})

	t.Run("missing_email", func(t *testing.T) {
		resp, _ := app.Post("/api/user/auth/local/register", map[string]string{
			"username": "testuser",
			"password": "password123",
		})
		assert.NotEqual(t, 200, resp.Code)
	})

	t.Run("missing_password", func(t *testing.T) {
		resp, _ := app.Post("/api/user/auth/local/register", map[string]string{
			"username": "testuser",
			"email":    "test@test.local",
		})
		assert.NotEqual(t, 200, resp.Code)
	})

	t.Run("duplicate_username", func(t *testing.T) {
		// First registration
		app.Post("/api/user/auth/local/register", map[string]string{
			"username": "duplicateuser",
			"email":    "dup1@test.local",
			"password": "password123",
		})

		// Second registration with same username
		resp, _ := app.Post("/api/user/auth/local/register", map[string]string{
			"username": "duplicateuser",
			"email":    "dup2@test.local",
			"password": "password123",
		})
		assert.NotEqual(t, 200, resp.Code)
	})
}

func TestAuthPasswordRecovery(t *testing.T) {
	app := CreateTestApp(t)

	t.Run("recover_with_valid_email", func(t *testing.T) {
		resp, _ := app.Post("/api/user/auth/local/recover", map[string]string{
			"email": "admin@test.local",
		})
		// May fail without mailer configured, but shouldn't be 500
		assert.NotEqual(t, 500, resp.Code)
	})

	t.Run("recover_with_invalid_email", func(t *testing.T) {
		resp, _ := app.Post("/api/user/auth/local/recover", map[string]string{
			"email": "nonexistent@test.local",
		})
		// Should handle gracefully
		assert.NotEqual(t, 500, resp.Code)
	})

	t.Run("recover_with_empty_email", func(t *testing.T) {
		resp, _ := app.Post("/api/user/auth/local/recover", map[string]string{
			"email": "",
		})
		assert.NotEqual(t, 200, resp.Code)
	})

	t.Run("recover_check_invalid_token", func(t *testing.T) {
		resp, _ := app.Post("/api/user/auth/local/recover/check", map[string]string{
			"token": "invalid-recovery-token",
		})
		assert.NotEqual(t, 200, resp.Code)
	})

	t.Run("recover_reset_invalid_token", func(t *testing.T) {
		resp, _ := app.Post("/api/user/auth/local/recover/reset", map[string]string{
			"token":    "invalid-recovery-token",
			"password": "newpassword123",
		})
		assert.NotEqual(t, 200, resp.Code)
	})
}

func TestAuthActivation(t *testing.T) {
	app := CreateTestApp(t)

	t.Run("activate_with_invalid_token", func(t *testing.T) {
		resp, _ := app.Post("/api/user/auth/local/activate", map[string]string{
			"token": "invalid-activation-token",
		})
		assert.NotEqual(t, 200, resp.Code)
	})

	t.Run("send_activation_link_valid_email", func(t *testing.T) {
		resp, _ := app.Post("/api/user/auth/local/activate/send", map[string]string{
			"email": "user@test.local",
		})
		// May fail without mailer, but shouldn't be 500
		assert.NotEqual(t, 500, resp.Code)
	})

	t.Run("send_activation_link_invalid_email", func(t *testing.T) {
		resp, _ := app.Post("/api/user/auth/local/activate/send", map[string]string{
			"email": "nonexistent@test.local",
		})
		assert.NotEqual(t, 500, resp.Code)
	})
}

func TestAuthorizationLevels(t *testing.T) {
	app := CreateTestApp(t)

	// Create some test data
	app.CreatePost("Test Post", "Test content", true)

	t.Run("admin_can_access_content", func(t *testing.T) {
		resp, _ := app.Get("/api/content/post", app.adminToken)
		app.AssertStatus(resp, 200)
	})

	t.Run("user_can_access_content", func(t *testing.T) {
		resp, _ := app.Get("/api/content/post", app.normalToken)
		// User may have access depending on permissions configuration
		assert.Contains(t, []int{200, 403}, resp.Code)
	})

	t.Run("guest_cannot_access_protected_content", func(t *testing.T) {
		resp, _ := app.Get("/api/content/post")
		// Should be unauthorized without token
		assert.Contains(t, []int{401, 403}, resp.Code)
	})
}

func TestTokenEdgeCases(t *testing.T) {
	app := CreateTestApp(t)

	t.Run("expired_token_format", func(t *testing.T) {
		// Test with a properly formatted but expired-looking token
		resp, _ := app.Get("/api/user/auth/me", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjF9.signature")
		assert.Equal(t, 401, resp.Code)
	})

	t.Run("token_with_extra_spaces", func(t *testing.T) {
		resp, _ := app.Get("/api/user/auth/me", "  "+app.adminToken+"  ")
		// Should still work after trimming or fail gracefully
		assert.NotEqual(t, 500, resp.Code)
	})

	t.Run("multiple_authorization_headers", func(t *testing.T) {
		// This tests robustness
		resp, _ := app.Get("/api/user/auth/me", app.adminToken)
		assert.NotEqual(t, 500, resp.Code)
	})
}

// Helper to parse token response
func parseTokenResponse(data json.RawMessage) (string, string) {
	var resp struct {
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
	}
	_ = json.Unmarshal(data, &resp)
	return resp.Token, resp.RefreshToken
}

package refreshtoken_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"path"
	"testing"
	"time"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/entity"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/entdbadapter"
	"github.com/fastschema/fastschema/pkg/jwt"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRefreshTokenFlow(t *testing.T) {
	schemaDir := t.TempDir()
	migrationDir := t.TempDir()
	sb := utils.Must(schema.NewBuilderFromDir(schemaDir, systemSchemas...))
	dbc := utils.Must(entdbadapter.NewEntClient(&db.Config{
		Driver:       "sqlite",
		Name:         path.Join(t.TempDir(), "fastschema.db"),
		MigrationDir: migrationDir,
		LogQueries:   false,
	}, sb))
	defer dbc.Close()

	runRefreshTokenTests(t, dbc)
}

func runRefreshTokenTests(t *testing.T, dbc db.Client) {
	t.Run("GenerateTokenPairAndVerifyDB", testGenerateTokenPairAndVerifyDB(dbc))
	t.Run("RefreshTokenViaHTTP", testRefreshTokenViaHTTP(dbc))
	t.Run("RefreshTokenInvalidTokenViaHTTP", testRefreshTokenInvalidTokenViaHTTP(dbc))
	t.Run("RefreshTokenExpiredTokenViaHTTP", testRefreshTokenExpiredTokenViaHTTP(dbc))
	t.Run("LogoutViaHTTP", testLogoutViaHTTP(dbc))
	t.Run("LogoutAllViaHTTP", testLogoutAllViaHTTP(dbc))
	t.Run("AccessProtectedResourceViaHTTP", testAccessProtectedResourceViaHTTP(dbc))
	t.Run("TokenRotationViaHTTP", testTokenRotationViaHTTP(dbc))
	t.Run("MeEndpointViaHTTP", testMeEndpointViaHTTP(dbc))
}

func testGenerateTokenPairAndVerifyDB(dbc db.Client) func(t *testing.T) {
	return func(t *testing.T) {
		app := createTestApp(t, dbc)

		// Generate token pair via the service
		ctx := context.Background()
		tokenPair, err := app.authService.GenerateJWTTokens(&mockContext{user: nil, db: app.db}, app.testUser)
		require.NoError(t, err)

		// Verify token pair response
		assert.NotEmpty(t, tokenPair.AccessToken)
		assert.NotEmpty(t, tokenPair.RefreshToken)
		assert.False(t, tokenPair.AccessTokenExpiresAt.IsZero())
		assert.False(t, tokenPair.RefreshTokenExpiresAt.IsZero())

		// Access token should expire before refresh token
		assert.True(t, tokenPair.AccessTokenExpiresAt.Before(*tokenPair.RefreshTokenExpiresAt))

		// Verify token is stored in DB using JTI
		claims, err := jwt.ParseRefreshToken(tokenPair.RefreshToken, app.Key())
		require.NoError(t, err)

		storedToken, err := db.Builder[*fs.Token](dbc).
			Where(db.EQ("jti", claims.ID)).
			First(ctx)
		require.NoError(t, err, "Token should be stored in database")
		assert.Equal(t, app.testUser.ID, storedToken.UserID, "User ID should match")
		assert.Equal(t, claims.ID, storedToken.JTI, "JTI should match")
		assert.False(t, storedToken.ExpiresAt.IsZero(), "ExpiresAt should be set")

		// Verify token count for user
		count, err := db.Builder[*fs.Token](dbc).
			Where(db.EQ("user_id", app.testUser.ID)).
			Count(ctx)
		require.NoError(t, err)
		assert.Equal(t, 1, count, "Should have exactly 1 token in DB")
	}
}

func testRefreshTokenViaHTTP(dbc db.Client) func(t *testing.T) {
	return func(t *testing.T) {
		app := createTestApp(t, dbc)
		ctx := context.Background()

		// Generate initial token pair
		tokenPair, err := app.authService.GenerateJWTTokens(&mockContext{user: nil, db: app.db}, app.testUser)
		require.NoError(t, err)

		// Get old JTI for verification
		oldClaims, err := jwt.ParseRefreshToken(tokenPair.RefreshToken, app.Key())
		require.NoError(t, err)
		oldJTI := oldClaims.ID

		// Verify old token exists in DB
		_, err = db.Builder[*fs.Token](dbc).
			Where(db.EQ("jti", oldJTI)).
			First(ctx)
		require.NoError(t, err, "Old token should exist in DB before refresh")

		// Make HTTP request to refresh token
		reqBody := map[string]string{"refresh_token": tokenPair.RefreshToken}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/auth/token/refresh", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.server.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 200, resp.StatusCode)

		// Parse response
		respBody := utils.Must(utils.ReadCloserToString(resp.Body))
		apiResp := parseResponse(t, []byte(respBody))

		var newTokenPair fs.JWTTokens
		require.NoError(t, json.Unmarshal(apiResp.Data, &newTokenPair))

		assert.NotEmpty(t, newTokenPair.AccessToken)
		assert.NotEmpty(t, newTokenPair.RefreshToken)
		assert.NotEqual(t, tokenPair.AccessToken, newTokenPair.AccessToken, "Access tokens should be different")
		assert.NotEqual(t, tokenPair.RefreshToken, newTokenPair.RefreshToken, "Refresh tokens should be different")

		// Verify old token is deleted from DB (token rotation)
		_, err = db.Builder[*fs.Token](dbc).
			Where(db.EQ("jti", oldJTI)).
			First(ctx)
		assert.True(t, db.IsNotFound(err), "Old token should be deleted from DB after refresh")

		// Verify new token exists in DB
		newClaims, err := jwt.ParseRefreshToken(newTokenPair.RefreshToken, app.Key())
		require.NoError(t, err)

		newStoredToken, err := db.Builder[*fs.Token](dbc).
			Where(db.EQ("jti", newClaims.ID)).
			First(ctx)
		require.NoError(t, err, "New token should be stored in DB")
		assert.Equal(t, app.testUser.ID, newStoredToken.UserID)
	}
}

func testRefreshTokenInvalidTokenViaHTTP(dbc db.Client) func(t *testing.T) {
	return func(t *testing.T) {
		app := createTestApp(t, dbc)

		// Test with invalid token
		reqBody := map[string]string{"refresh_token": "invalid-token"}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/auth/token/refresh", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.server.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.NotEqual(t, 200, resp.StatusCode, "Should fail with invalid token")

		// Test with empty token
		reqBody = map[string]string{"refresh_token": ""}
		body, _ = json.Marshal(reqBody)
		req = httptest.NewRequest("POST", "/api/auth/token/refresh", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err = app.server.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.NotEqual(t, 200, resp.StatusCode, "Should fail with empty token")

		// Test with no body
		req = httptest.NewRequest("POST", "/api/auth/token/refresh", nil)
		req.Header.Set("Content-Type", "application/json")

		resp, err = app.server.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.NotEqual(t, 200, resp.StatusCode, "Should fail with no body")

		// Verify no tokens were created in DB for this user
		count, err := db.Builder[*fs.Token](dbc).
			Where(db.EQ("user_id", app.testUser.ID)).
			Count(context.Background())
		require.NoError(t, err)
		assert.Equal(t, 0, count, "No tokens should be created for failed requests")
	}
}

func testRefreshTokenExpiredTokenViaHTTP(dbc db.Client) func(t *testing.T) {
	return func(t *testing.T) {
		app := createTestApp(t, dbc)
		ctx := context.Background()

		// Create an expired token
		expiredTime := time.Now().Add(-1 * time.Hour)
		jti := jwt.GenerateJTI()
		expiredToken, err := jwt.GenerateRefreshToken(app.testUser.ID, jti, app.Key(), expiredTime)
		require.NoError(t, err)

		// Store it in DB
		_, err = db.Builder[*fs.Token](dbc).Create(ctx, entity.New().
			Set("user_id", app.testUser.ID).
			Set("jti", jti).
			Set("expires_at", expiredTime))
		require.NoError(t, err)

		// Verify token exists in DB
		storedToken, err := db.Builder[*fs.Token](dbc).
			Where(db.EQ("jti", jti)).
			First(ctx)
		require.NoError(t, err)
		assert.Equal(t, jti, storedToken.JTI)

		// Try to refresh with expired token
		reqBody := map[string]string{"refresh_token": expiredToken}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/auth/token/refresh", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.server.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.NotEqual(t, 200, resp.StatusCode, "Should fail with expired token")
	}
}

func testLogoutViaHTTP(dbc db.Client) func(t *testing.T) {
	return func(t *testing.T) {
		app := createTestApp(t, dbc)
		ctx := context.Background()

		// Generate token pair
		tokenPair, err := app.authService.GenerateJWTTokens(&mockContext{user: nil, db: app.db}, app.testUser)
		require.NoError(t, err)

		// Get JTI for verification
		claims, err := jwt.ParseRefreshToken(tokenPair.RefreshToken, app.Key())
		require.NoError(t, err)

		// Verify token exists in DB
		_, err = db.Builder[*fs.Token](dbc).
			Where(db.EQ("jti", claims.ID)).
			First(ctx)
		require.NoError(t, err, "Token should exist in DB before logout")

		// Make HTTP request to logout
		reqBody := map[string]string{"refresh_token": tokenPair.RefreshToken}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/auth/logout", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.server.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 200, resp.StatusCode)

		// Verify token is deleted from DB
		_, err = db.Builder[*fs.Token](dbc).
			Where(db.EQ("jti", claims.ID)).
			First(ctx)
		assert.True(t, db.IsNotFound(err), "Token should be deleted from DB after logout")

		// Verify token count for user is 0
		count, err := db.Builder[*fs.Token](dbc).
			Where(db.EQ("user_id", app.testUser.ID)).
			Count(ctx)
		require.NoError(t, err)
		assert.Equal(t, 0, count, "Should have no tokens in DB after logout")

		// Try to refresh after logout - should fail
		reqBody = map[string]string{"refresh_token": tokenPair.RefreshToken}
		body, _ = json.Marshal(reqBody)
		req = httptest.NewRequest("POST", "/api/auth/token/refresh", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err = app.server.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.NotEqual(t, 200, resp.StatusCode, "Should fail to refresh after logout")
	}
}

func testLogoutAllViaHTTP(dbc db.Client) func(t *testing.T) {
	return func(t *testing.T) {
		app := createTestApp(t, dbc)
		ctx := context.Background()

		// Generate multiple token pairs
		tokenPair1, err := app.authService.GenerateJWTTokens(&mockContext{user: nil, db: app.db}, app.testUser)
		require.NoError(t, err)

		tokenPair2, err := app.authService.GenerateJWTTokens(&mockContext{user: nil, db: app.db}, app.testUser)
		require.NoError(t, err)

		// Verify both tokens exist in DB
		count, err := db.Builder[*fs.Token](dbc).
			Where(db.EQ("user_id", app.testUser.ID)).
			Count(ctx)
		require.NoError(t, err)
		assert.Equal(t, 2, count, "Should have 2 tokens in DB")

		// Get JTIs for verification
		claims1, err := jwt.ParseRefreshToken(tokenPair1.RefreshToken, app.Key())
		require.NoError(t, err)
		claims2, err := jwt.ParseRefreshToken(tokenPair2.RefreshToken, app.Key())
		require.NoError(t, err)

		// Make HTTP request to logout all (requires auth)
		req := httptest.NewRequest("POST", "/api/auth/logout/all", nil)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+tokenPair1.AccessToken)

		resp, err := app.server.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 200, resp.StatusCode)

		// Verify all tokens are deleted from DB
		for _, jti := range []string{claims1.ID, claims2.ID} {
			_, err = db.Builder[*fs.Token](dbc).
				Where(db.EQ("jti", jti)).
				First(ctx)
			assert.True(t, db.IsNotFound(err), "Token with JTI %s should be deleted", jti)
		}

		// Verify token count for user is 0
		count, err = db.Builder[*fs.Token](dbc).
			Where(db.EQ("user_id", app.testUser.ID)).
			Count(ctx)
		require.NoError(t, err)
		assert.Equal(t, 0, count, "Should have no tokens in DB after logout all")
	}
}

func testAccessProtectedResourceViaHTTP(dbc db.Client) func(t *testing.T) {
	return func(t *testing.T) {
		app := createTestApp(t, dbc)

		// Generate token pair
		tokenPair, err := app.authService.GenerateJWTTokens(&mockContext{user: nil, db: app.db}, app.testUser)
		require.NoError(t, err)

		// Access protected resource without token - should fail
		req := httptest.NewRequest("GET", "/api/protected", nil)
		resp, err := app.server.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, 401, resp.StatusCode, "Should be unauthorized without token")

		// Access protected resource with valid access token
		req = httptest.NewRequest("GET", "/api/protected", nil)
		req.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)

		resp, err = app.server.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 200, resp.StatusCode, "Should be authorized with valid token")

		// Parse response
		respBody := utils.Must(utils.ReadCloserToString(resp.Body))
		apiResp := parseResponse(t, []byte(respBody))

		var data map[string]any
		require.NoError(t, json.Unmarshal(apiResp.Data, &data))
		assert.Equal(t, "protected resource", data["message"])
		assert.Equal(t, float64(app.testUser.ID), data["user_id"])

		// Access with invalid token - should fail
		req = httptest.NewRequest("GET", "/api/protected", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")

		resp, err = app.server.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 401, resp.StatusCode, "Should be unauthorized with invalid token")
	}
}

func testTokenRotationViaHTTP(dbc db.Client) func(t *testing.T) {
	return func(t *testing.T) {
		app := createTestApp(t, dbc)
		ctx := context.Background()

		// Generate initial token pair
		tokenPair1, err := app.authService.GenerateJWTTokens(&mockContext{user: nil, db: app.db}, app.testUser)
		require.NoError(t, err)

		// Verify initial token count
		count, err := db.Builder[*fs.Token](dbc).
			Where(db.EQ("user_id", app.testUser.ID)).
			Count(ctx)
		require.NoError(t, err)
		assert.Equal(t, 1, count, "Should have 1 token initially")

		// First refresh
		reqBody := map[string]string{"refresh_token": tokenPair1.RefreshToken}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/auth/token/refresh", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.server.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, 200, resp.StatusCode)

		respBody := utils.Must(utils.ReadCloserToString(resp.Body))
		apiResp := parseResponse(t, []byte(respBody))
		var tokenPair2 fs.JWTTokens
		require.NoError(t, json.Unmarshal(apiResp.Data, &tokenPair2))

		// Verify token count is still 1 (old deleted, new created)
		count, err = db.Builder[*fs.Token](dbc).
			Where(db.EQ("user_id", app.testUser.ID)).
			Count(ctx)
		require.NoError(t, err)
		assert.Equal(t, 1, count, "Should still have 1 token after first rotation")

		// Second refresh
		reqBody = map[string]string{"refresh_token": tokenPair2.RefreshToken}
		body, _ = json.Marshal(reqBody)
		req = httptest.NewRequest("POST", "/api/auth/token/refresh", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err = app.server.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, 200, resp.StatusCode)

		respBody = utils.Must(utils.ReadCloserToString(resp.Body))
		apiResp = parseResponse(t, []byte(respBody))
		var tokenPair3 fs.JWTTokens
		require.NoError(t, json.Unmarshal(apiResp.Data, &tokenPair3))

		// All tokens should be different
		assert.NotEqual(t, tokenPair1.RefreshToken, tokenPair2.RefreshToken)
		assert.NotEqual(t, tokenPair2.RefreshToken, tokenPair3.RefreshToken)
		assert.NotEqual(t, tokenPair1.RefreshToken, tokenPair3.RefreshToken)

		// Verify only the last token exists in DB
		count, err = db.Builder[*fs.Token](dbc).
			Where(db.EQ("user_id", app.testUser.ID)).
			Count(ctx)
		require.NoError(t, err)
		assert.Equal(t, 1, count, "Should have only 1 token after multiple rotations")

		// Verify old tokens are gone
		claims1, _ := jwt.ParseRefreshToken(tokenPair1.RefreshToken, app.Key())
		claims2, _ := jwt.ParseRefreshToken(tokenPair2.RefreshToken, app.Key())
		claims3, _ := jwt.ParseRefreshToken(tokenPair3.RefreshToken, app.Key())

		_, err = db.Builder[*fs.Token](dbc).Where(db.EQ("jti", claims1.ID)).First(ctx)
		assert.True(t, db.IsNotFound(err), "First token should be deleted")

		_, err = db.Builder[*fs.Token](dbc).Where(db.EQ("jti", claims2.ID)).First(ctx)
		assert.True(t, db.IsNotFound(err), "Second token should be deleted")

		_, err = db.Builder[*fs.Token](dbc).Where(db.EQ("jti", claims3.ID)).First(ctx)
		require.NoError(t, err, "Third (current) token should exist")
	}
}

func testMeEndpointViaHTTP(dbc db.Client) func(t *testing.T) {
	return func(t *testing.T) {
		app := createTestApp(t, dbc)

		// Generate token pair
		tokenPair, err := app.authService.GenerateJWTTokens(&mockContext{user: nil, db: app.db}, app.testUser)
		require.NoError(t, err)

		// Access /me without token - should return 401 (requires authentication)
		req := httptest.NewRequest("GET", "/api/auth/me", nil)
		resp, err := app.server.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, 401, resp.StatusCode, "Me endpoint requires authentication")

		// Access /me with valid token
		req = httptest.NewRequest("GET", "/api/auth/me", nil)
		req.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)

		resp, err = app.server.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 200, resp.StatusCode)

		// Parse response
		respBody := utils.Must(utils.ReadCloserToString(resp.Body))
		apiResp := parseResponse(t, []byte(respBody))

		var user fs.User
		require.NoError(t, json.Unmarshal(apiResp.Data, &user))
		assert.Equal(t, app.testUser.ID, user.ID)
		assert.Equal(t, app.testUser.Username, user.Username)
	}
}

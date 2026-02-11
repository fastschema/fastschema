package passwordless_otp_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/entity"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/auth"
	"github.com/fastschema/fastschema/pkg/entdbadapter"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	schemaDir    = "../../../tests/integration/passwordless_otp/data/schemas"
	migrationDir = "../../../tests/integration/passwordless_otp/data/migrations"
	sqliteDSN    = "../../../tests/integration/passwordless_otp/data/passwordless_otp.db"
)

func TestOTPPasswordlessLogin(t *testing.T) {
	sb := utils.Must(schema.NewBuilderFromDir(schemaDir, systemSchemas...))
	dbc := utils.Must(entdbadapter.NewEntClient(&db.Config{
		Driver:       "sqlite",
		Name:         sqliteDSN,
		MigrationDir: migrationDir,
		LogQueries:   true,
	}, sb))
	defer dbc.Close()

	runOTPTests(t, dbc)
}

func runOTPTests(t *testing.T, dbc db.Client) {
	// Request OTP tests
	t.Run("OTPRequestSuccess", testOTPRequestSuccess(dbc))
	t.Run("OTPRequestEmailContent", testOTPRequestEmailContent(dbc))
	t.Run("OTPRequestInvalidEmail", testOTPRequestInvalidEmail(dbc))
	t.Run("OTPRequestEmptyEmail", testOTPRequestEmptyEmail(dbc))
	t.Run("OTPRequestUserNotFound", testOTPRequestUserNotFound(dbc))
	t.Run("OTPRequestInactiveUser", testOTPRequestInactiveUser(dbc))
	t.Run("OTPRequestNotEnabled", testOTPRequestNotEnabled(dbc))

	// Verify OTP tests
	t.Run("OTPVerifySuccess", testOTPVerifySuccess(dbc))
	t.Run("OTPVerifyAndAccessProtectedResource", testOTPVerifyAndAccessProtectedResource(dbc))
	t.Run("OTPVerifyInvalidCode", testOTPVerifyInvalidCode(dbc))
	t.Run("OTPVerifyExpired", testOTPVerifyExpired(dbc))
	t.Run("OTPVerifyMaxAttempts", testOTPVerifyMaxAttempts(dbc))
	t.Run("OTPVerifyNoSession", testOTPVerifyNoSession(dbc))
	t.Run("OTPVerifyEmptyOTP", testOTPVerifyEmptyOTP(dbc))
	t.Run("OTPVerifyNotEnabled", testOTPVerifyNotEnabled(dbc))

	// Edge cases
	t.Run("OTPMultipleRequests", testOTPMultipleRequests(dbc))
	t.Run("OTPSessionStorageAndCleanup", testOTPSessionStorageAndCleanup(dbc))
	t.Run("OTPAttemptIncrement", testOTPAttemptIncrement(dbc))
	t.Run("OTPDifferentProviderUser", testOTPDifferentProviderUser(dbc))
	t.Run("OTPFullLoginFlow", testOTPFullLoginFlow(dbc))
	t.Run("OTPCaseSensitivity", testOTPCaseSensitivity(dbc))
}

// ============== Request OTP Tests ==============

func testOTPRequestSuccess(dbc db.Client) func(t *testing.T) {
	return func(t *testing.T) {
		clearOTPSessions(dbc)
		app := createTestApp(t, dbc, &fs.OTPConfig{
			Enabled:     true,
			Length:      6,
			Expiration:  300,
			MaxAttempts: 3,
		})

		// Verify OTP provider and mailer are correctly set up
		require.NotNil(t, app.otpProvider, "OTPProvider should not be nil")
		require.True(t, app.otpProvider.IsEnabled(), "OTP should be enabled")
		require.NotNil(t, app.mailer, "Mailer should not be nil")

		// Request OTP via HTTP
		reqBody := map[string]string{"email": app.testUser.Email}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/auth/otp/request", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.server.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		respBody := utils.Must(utils.ReadCloserToString(resp.Body))
		assert.Equal(t, 200, resp.StatusCode)

		// Parse response
		apiResp := parseResponse(t, []byte(respBody))

		var otpResp auth.OTPResponse
		require.NoError(t, json.Unmarshal(apiResp.Data, &otpResp))

		assert.Equal(t, auth.MSG_OTP_SENT, otpResp.Message)
		assert.Equal(t, 300, otpResp.ExpiresIn)
		assert.NotEmpty(t, otpResp.SessionID, "Session ID should be returned")

		// Verify session ID is a valid UUID
		sessionUUID, err := uuid.Parse(otpResp.SessionID)
		require.NoError(t, err, "Session ID should be a valid UUID")

		// Wait for async email - the email is sent in a goroutine
		for i := 0; i < 20; i++ {
			time.Sleep(50 * time.Millisecond)
			if app.mailer.LastMail() != nil {
				break
			}
		}

		// Check that session was created with the returned session ID
		session, sessionErr := db.Builder[*fs.Session](dbc).
			Where(db.EQ("id", sessionUUID)).
			Where(db.EQ("type", string(fs.SessionTypeOTPLogin))).
			Where(db.EQ("status", string(fs.SessionStatusPendingOTP))).
			First(context.Background())
		require.NoError(t, sessionErr, "OTP session should be created")
		assert.NotEmpty(t, session.OTPHash, "OTP hash should be set")
		assert.Equal(t, 0, session.OTPAttempts)
		assert.Equal(t, app.testUser.ID, session.UserID)

		// Verify email was sent
		require.NotNil(t, app.mailer.LastMail(), "Email should have been sent")
		assert.Equal(t, []string{app.testUser.Email}, app.mailer.LastMail().To)
	}
}

func testOTPRequestEmailContent(dbc db.Client) func(t *testing.T) {
	return func(t *testing.T) {
		clearOTPSessions(dbc)
		app := createTestApp(t, dbc, &fs.OTPConfig{
			Enabled:     true,
			Length:      6,
			Expiration:  300,
			MaxAttempts: 3,
		})

		reqBody := map[string]string{"email": app.testUser.Email}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/auth/otp/request", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.server.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, 200, resp.StatusCode)

		// Wait for async email
		time.Sleep(100 * time.Millisecond)

		// Verify email content
		mail := app.mailer.LastMail()
		require.NotNil(t, mail)

		// Check email subject contains app name and verification code
		assert.Contains(t, mail.Subject, "TestOTPApp")
		assert.Contains(t, mail.Subject, "verification code")

		// Check email body contains verification code info
		assert.Contains(t, mail.Body, "verification code")
		assert.Contains(t, mail.Body, "5 minutes") // Expiration time

		// Check recipient
		assert.Equal(t, []string{app.testUser.Email}, mail.To)
	}
}

func testOTPRequestInvalidEmail(dbc db.Client) func(t *testing.T) {
	return func(t *testing.T) {
		clearOTPSessions(dbc)
		app := createTestApp(t, dbc, &fs.OTPConfig{
			Enabled:     true,
			Length:      6,
			Expiration:  300,
			MaxAttempts: 3,
		})

		// Test with invalid email format
		reqBody := map[string]string{"email": "not-an-email"}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/auth/otp/request", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.server.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 422, resp.StatusCode)

		// Verify no email was sent
		assert.Nil(t, app.mailer.LastMail())

		// Verify no session was created
		count, err := db.Builder[*fs.Session](dbc).
			Where(db.EQ("type", string(fs.SessionTypeOTPLogin))).
			Count(context.Background())
		require.NoError(t, err)
		assert.Equal(t, 0, count)
	}
}

func testOTPRequestEmptyEmail(dbc db.Client) func(t *testing.T) {
	return func(t *testing.T) {
		clearOTPSessions(dbc)
		app := createTestApp(t, dbc, &fs.OTPConfig{
			Enabled:     true,
			Length:      6,
			Expiration:  300,
			MaxAttempts: 3,
		})

		reqBody := map[string]string{"email": ""}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/auth/otp/request", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.server.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 422, resp.StatusCode)
	}
}

func testOTPRequestUserNotFound(dbc db.Client) func(t *testing.T) {
	return func(t *testing.T) {
		clearOTPSessions(dbc)
		app := createTestApp(t, dbc, &fs.OTPConfig{
			Enabled:     true,
			Length:      6,
			Expiration:  300,
			MaxAttempts: 3,
		})

		// Request OTP for non-existent user (should return success to prevent enumeration)
		reqBody := map[string]string{"email": "nonexistent@example.com"}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/auth/otp/request", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.server.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Should return 200 to prevent email enumeration
		assert.Equal(t, 200, resp.StatusCode)

		// Parse response
		respBody := utils.Must(utils.ReadCloserToString(resp.Body))
		apiResp := parseResponse(t, []byte(respBody))

		var otpResp auth.OTPResponse
		require.NoError(t, json.Unmarshal(apiResp.Data, &otpResp))
		assert.Equal(t, auth.MSG_OTP_SENT, otpResp.Message)

		time.Sleep(100 * time.Millisecond)

		// Verify NO email was sent (user doesn't exist)
		assert.Nil(t, app.mailer.LastMail())

		// Verify no session was created
		count, err := db.Builder[*fs.Session](dbc).
			Where(db.EQ("type", string(fs.SessionTypeOTPLogin))).
			Count(context.Background())
		require.NoError(t, err)
		assert.Equal(t, 0, count)
	}
}

func testOTPRequestInactiveUser(dbc db.Client) func(t *testing.T) {
	return func(t *testing.T) {
		clearOTPSessions(dbc)
		app := createTestApp(t, dbc, &fs.OTPConfig{
			Enabled:     true,
			Length:      6,
			Expiration:  300,
			MaxAttempts: 3,
		})

		// Create inactive user
		userModel := utils.Must(dbc.Model("user"))
		inactiveEmail := "inactive" + utils.RandomString(8) + "@example.com"
		_, err := userModel.Create(context.Background(), entity.New().
			Set("username", "inactiveuser"+utils.RandomString(8)).
			Set("email", inactiveEmail).
			Set("password", "testpassword").
			Set("provider", "local").
			Set("provider_id", utils.RandomString(8)).
			Set("active", false))
		require.NoError(t, err)

		reqBody := map[string]string{"email": inactiveEmail}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/auth/otp/request", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.server.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Should return 200 (to prevent enumeration)
		assert.Equal(t, 200, resp.StatusCode)

		time.Sleep(100 * time.Millisecond)

		// No email should be sent for inactive user
		assert.Nil(t, app.mailer.LastMail())
	}
}

func testOTPRequestNotEnabled(dbc db.Client) func(t *testing.T) {
	return func(t *testing.T) {
		clearOTPSessions(dbc)
		// Create app without OTP enabled
		app := createTestApp(t, dbc, nil)

		reqBody := map[string]string{"email": "test@example.com"}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/auth/otp/request", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.server.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Should return 404 (endpoint not registered when OTP disabled)
		assert.Equal(t, 404, resp.StatusCode)
	}
}

// ============== Verify OTP Tests ==============

func testOTPVerifySuccess(dbc db.Client) func(t *testing.T) {
	return func(t *testing.T) {
		clearOTPSessions(dbc)
		app := createTestApp(t, dbc, &fs.OTPConfig{
			Enabled:     true,
			Length:      6,
			Expiration:  300,
			MaxAttempts: 3,
		})

		// Create OTP session with known OTP
		otp := "123456"
		otpHash, _ := auth.HashOTP(otp)
		expiresAt := time.Now().Add(5 * time.Minute)

		sessionID := utils.Must(uuid.NewV7())
		_, err := db.Builder[*fs.Session](dbc).Create(context.Background(), entity.New().
			Set("id", sessionID).
			Set("user_id", app.testUser.ID).
			Set("type", string(fs.SessionTypeOTPLogin)).
			Set("status", string(fs.SessionStatusPendingOTP)).
			Set("otp_hash", otpHash).
			Set("otp_attempts", 0).
			Set("expires_at", expiresAt))
		require.NoError(t, err)

		// Verify OTP via HTTP using session_id
		reqBody := map[string]string{
			"session_id": sessionID.String(),
			"otp":        otp,
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/auth/otp/verify", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.server.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 200, resp.StatusCode)

		// Parse response
		respBody := utils.Must(utils.ReadCloserToString(resp.Body))
		apiResp := parseResponse(t, []byte(respBody))

		var tokens fs.JWTTokens
		require.NoError(t, json.Unmarshal(apiResp.Data, &tokens))

		assert.NotEmpty(t, tokens.AccessToken)
		assert.NotEmpty(t, tokens.RefreshToken)

		// Verify OTP session was deleted
		_, err = db.Builder[*fs.Session](dbc).
			Where(db.EQ("id", sessionID)).
			First(context.Background())
		assert.True(t, db.IsNotFound(err))
	}
}

func testOTPVerifyAndAccessProtectedResource(dbc db.Client) func(t *testing.T) {
	return func(t *testing.T) {
		clearOTPSessions(dbc)
		app := createTestApp(t, dbc, &fs.OTPConfig{
			Enabled:     true,
			Length:      6,
			Expiration:  300,
			MaxAttempts: 3,
		})

		// Create OTP session
		otp := "654321"
		otpHash, _ := auth.HashOTP(otp)
		expiresAt := time.Now().Add(5 * time.Minute)

		sessionID := utils.Must(uuid.NewV7())
		_, err := db.Builder[*fs.Session](dbc).Create(context.Background(), entity.New().
			Set("id", sessionID).
			Set("user_id", app.testUser.ID).
			Set("type", string(fs.SessionTypeOTPLogin)).
			Set("status", string(fs.SessionStatusPendingOTP)).
			Set("otp_hash", otpHash).
			Set("otp_attempts", 0).
			Set("expires_at", expiresAt))
		require.NoError(t, err)

		// Verify OTP using session_id
		reqBody := map[string]string{
			"session_id": sessionID.String(),
			"otp":        otp,
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/auth/otp/verify", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.server.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, 200, resp.StatusCode)

		respBody := utils.Must(utils.ReadCloserToString(resp.Body))
		apiResp := parseResponse(t, []byte(respBody))

		var tokens fs.JWTTokens
		require.NoError(t, json.Unmarshal(apiResp.Data, &tokens))

		// Access protected resource with token
		req = httptest.NewRequest("GET", "/api/protected", nil)
		req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)

		resp, err = app.server.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 200, resp.StatusCode)

		respBody = utils.Must(utils.ReadCloserToString(resp.Body))
		apiResp = parseResponse(t, []byte(respBody))

		var data map[string]any
		require.NoError(t, json.Unmarshal(apiResp.Data, &data))
		assert.Equal(t, "protected resource", data["message"])
		assert.Equal(t, float64(app.testUser.ID), data["user_id"])
	}
}

func testOTPVerifyInvalidCode(dbc db.Client) func(t *testing.T) {
	return func(t *testing.T) {
		clearOTPSessions(dbc)
		app := createTestApp(t, dbc, &fs.OTPConfig{
			Enabled:     true,
			Length:      6,
			Expiration:  300,
			MaxAttempts: 3,
		})

		// Create OTP session with known OTP
		otp := "123456"
		otpHash, _ := auth.HashOTP(otp)
		expiresAt := time.Now().Add(5 * time.Minute)

		sessionID := utils.Must(uuid.NewV7())
		_, err := db.Builder[*fs.Session](dbc).Create(context.Background(), entity.New().
			Set("id", sessionID).
			Set("user_id", app.testUser.ID).
			Set("type", string(fs.SessionTypeOTPLogin)).
			Set("status", string(fs.SessionStatusPendingOTP)).
			Set("otp_hash", otpHash).
			Set("otp_attempts", 0).
			Set("expires_at", expiresAt))
		require.NoError(t, err)

		// Try to verify with wrong OTP
		reqBody := map[string]string{
			"session_id": sessionID.String(),
			"otp":        "000000", // Wrong OTP
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/auth/otp/verify", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.server.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 422, resp.StatusCode) // Invalid OTP returns 422 UnprocessableEntity

		// Verify session attempt was incremented
		session, err := db.Builder[*fs.Session](dbc).
			Where(db.EQ("id", sessionID)).
			First(context.Background())
		require.NoError(t, err)
		assert.Equal(t, 1, session.OTPAttempts)
	}
}

func testOTPVerifyExpired(dbc db.Client) func(t *testing.T) {
	return func(t *testing.T) {
		clearOTPSessions(dbc)
		app := createTestApp(t, dbc, &fs.OTPConfig{
			Enabled:     true,
			Length:      6,
			Expiration:  300,
			MaxAttempts: 3,
		})

		// Create expired OTP session
		otp := "123456"
		otpHash, _ := auth.HashOTP(otp)
		expiredTime := time.Now().Add(-1 * time.Hour)

		sessionID := utils.Must(uuid.NewV7())
		_, err := db.Builder[*fs.Session](dbc).Create(context.Background(), entity.New().
			Set("id", sessionID).
			Set("user_id", app.testUser.ID).
			Set("type", string(fs.SessionTypeOTPLogin)).
			Set("status", string(fs.SessionStatusPendingOTP)).
			Set("otp_hash", otpHash).
			Set("otp_attempts", 0).
			Set("expires_at", expiredTime))
		require.NoError(t, err)

		reqBody := map[string]string{
			"session_id": sessionID.String(),
			"otp":        otp,
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/auth/otp/verify", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.server.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 422, resp.StatusCode) // Expired OTP returns 422 UnprocessableEntity

		// Verify session was marked inactive
		session, err := db.Builder[*fs.Session](dbc).
			Where(db.EQ("id", sessionID)).
			First(context.Background())
		require.NoError(t, err)
		assert.Equal(t, string(fs.SessionStatusInactive), session.Status)
	}
}

func testOTPVerifyMaxAttempts(dbc db.Client) func(t *testing.T) {
	return func(t *testing.T) {
		clearOTPSessions(dbc)
		app := createTestApp(t, dbc, &fs.OTPConfig{
			Enabled:     true,
			Length:      6,
			Expiration:  300,
			MaxAttempts: 3,
		})

		// Create OTP session with max attempts reached
		otp := "123456"
		otpHash, _ := auth.HashOTP(otp)
		expiresAt := time.Now().Add(5 * time.Minute)

		sessionID := utils.Must(uuid.NewV7())
		_, err := db.Builder[*fs.Session](dbc).Create(context.Background(), entity.New().
			Set("id", sessionID).
			Set("user_id", app.testUser.ID).
			Set("type", string(fs.SessionTypeOTPLogin)).
			Set("status", string(fs.SessionStatusPendingOTP)).
			Set("otp_hash", otpHash).
			Set("otp_attempts", 3). // Max attempts already reached
			Set("expires_at", expiresAt))
		require.NoError(t, err)

		reqBody := map[string]string{
			"session_id": sessionID.String(),
			"otp":        otp, // Correct OTP but max attempts reached
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/auth/otp/verify", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.server.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Should fail even with correct OTP
		assert.NotEqual(t, 200, resp.StatusCode)

		// Verify session was marked inactive
		session, err := db.Builder[*fs.Session](dbc).
			Where(db.EQ("id", sessionID)).
			First(context.Background())
		require.NoError(t, err)
		assert.Equal(t, string(fs.SessionStatusInactive), session.Status)
	}
}

func testOTPVerifyNoSession(dbc db.Client) func(t *testing.T) {
	return func(t *testing.T) {
		clearOTPSessions(dbc)
		app := createTestApp(t, dbc, &fs.OTPConfig{
			Enabled:     true,
			Length:      6,
			Expiration:  300,
			MaxAttempts: 3,
		})

		// Try to verify with a random session ID that doesn't exist
		fakeSessionID := utils.Must(uuid.NewV7())
		reqBody := map[string]string{
			"session_id": fakeSessionID.String(),
			"otp":        "123456",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/auth/otp/verify", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.server.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 422, resp.StatusCode) // No session returns 422 UnprocessableEntity (invalid OTP)
	}
}

func testOTPVerifyEmptyOTP(dbc db.Client) func(t *testing.T) {
	return func(t *testing.T) {
		clearOTPSessions(dbc)
		app := createTestApp(t, dbc, &fs.OTPConfig{
			Enabled:     true,
			Length:      6,
			Expiration:  300,
			MaxAttempts: 3,
		})

		sessionID := utils.Must(uuid.NewV7())
		reqBody := map[string]string{
			"session_id": sessionID.String(),
			"otp":        "",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/auth/otp/verify", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.server.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 422, resp.StatusCode)
	}
}

func testOTPVerifyNotEnabled(dbc db.Client) func(t *testing.T) {
	return func(t *testing.T) {
		clearOTPSessions(dbc)
		app := createTestApp(t, dbc, nil)

		sessionID := utils.Must(uuid.NewV7())
		reqBody := map[string]string{
			"session_id": sessionID.String(),
			"otp":        "123456",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/auth/otp/verify", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.server.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Should return 404 (endpoint not registered when OTP disabled)
		assert.Equal(t, 404, resp.StatusCode)
	}
}

// ============== Edge Cases ==============

func testOTPMultipleRequests(dbc db.Client) func(t *testing.T) {
	return func(t *testing.T) {
		clearOTPSessions(dbc)
		app := createTestApp(t, dbc, &fs.OTPConfig{
			Enabled:     true,
			Length:      6,
			Expiration:  300,
			MaxAttempts: 3,
		})

		// First OTP request
		reqBody := map[string]string{"email": app.testUser.Email}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/auth/otp/request", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.server.Test(req)
		require.NoError(t, err)

		respBody := utils.Must(utils.ReadCloserToString(resp.Body))
		resp.Body.Close()
		assert.Equal(t, 200, resp.StatusCode)

		// Parse first session ID
		apiResp := parseResponse(t, []byte(respBody))
		var firstOTPResponse auth.OTPResponse
		require.NoError(t, json.Unmarshal(apiResp.Data, &firstOTPResponse))
		firstSessionID := firstOTPResponse.SessionID

		time.Sleep(50 * time.Millisecond)

		// Second OTP request (should create a new session, not invalidate first)
		body, _ = json.Marshal(reqBody)
		req = httptest.NewRequest("POST", "/api/auth/otp/request", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err = app.server.Test(req)
		require.NoError(t, err)

		respBody = utils.Must(utils.ReadCloserToString(resp.Body))
		resp.Body.Close()
		assert.Equal(t, 200, resp.StatusCode)

		// Parse second session ID
		apiResp = parseResponse(t, []byte(respBody))
		var secondOTPResponse auth.OTPResponse
		require.NoError(t, json.Unmarshal(apiResp.Data, &secondOTPResponse))
		secondSessionID := secondOTPResponse.SessionID

		// Session IDs should be different
		assert.NotEqual(t, firstSessionID, secondSessionID)

		time.Sleep(50 * time.Millisecond)

		// Both sessions should still be pending (multi-device support)
		firstSession, err := db.Builder[*fs.Session](dbc).
			Where(db.EQ("id", utils.Must(uuid.Parse(firstSessionID)))).
			First(context.Background())
		require.NoError(t, err)
		assert.Equal(t, string(fs.SessionStatusPendingOTP), firstSession.Status)

		secondSession, err := db.Builder[*fs.Session](dbc).
			Where(db.EQ("id", utils.Must(uuid.Parse(secondSessionID)))).
			First(context.Background())
		require.NoError(t, err)
		assert.Equal(t, string(fs.SessionStatusPendingOTP), secondSession.Status)

		// Count pending sessions for user
		pendingSessions, err := db.Builder[*fs.Session](dbc).
			Where(db.EQ("user_id", app.testUser.ID)).
			Where(db.EQ("type", string(fs.SessionTypeOTPLogin))).
			Where(db.EQ("status", string(fs.SessionStatusPendingOTP))).
			Get(context.Background())
		require.NoError(t, err)
		assert.Len(t, pendingSessions, 2)
	}
}

func testOTPSessionStorageAndCleanup(dbc db.Client) func(t *testing.T) {
	return func(t *testing.T) {
		clearOTPSessions(dbc)
		app := createTestApp(t, dbc, &fs.OTPConfig{
			Enabled:     true,
			Length:      6,
			Expiration:  300,
			MaxAttempts: 3,
		})

		// Create OTP session
		otp := "111111"
		otpHash, _ := auth.HashOTP(otp)
		expiresAt := time.Now().Add(5 * time.Minute)

		sessionID := utils.Must(uuid.NewV7())
		_, err := db.Builder[*fs.Session](dbc).Create(context.Background(), entity.New().
			Set("id", sessionID).
			Set("user_id", app.testUser.ID).
			Set("type", string(fs.SessionTypeOTPLogin)).
			Set("status", string(fs.SessionStatusPendingOTP)).
			Set("otp_hash", otpHash).
			Set("otp_attempts", 0).
			Set("ip_address", "192.168.1.1").
			Set("device_info", "Test Browser").
			Set("expires_at", expiresAt))
		require.NoError(t, err)

		// Verify session is stored correctly
		session, err := db.Builder[*fs.Session](dbc).
			Where(db.EQ("id", sessionID)).
			First(context.Background())
		require.NoError(t, err)

		assert.Equal(t, app.testUser.ID, session.UserID)
		assert.Equal(t, string(fs.SessionTypeOTPLogin), session.Type)
		assert.Equal(t, string(fs.SessionStatusPendingOTP), session.Status)
		assert.NotEmpty(t, session.OTPHash)
		assert.Equal(t, 0, session.OTPAttempts)
		assert.Equal(t, "192.168.1.1", session.IPAddress)
		assert.Equal(t, "Test Browser", session.DeviceInfo)

		// Verify OTP and delete session
		reqBody := map[string]string{
			"session_id": sessionID.String(),
			"otp":        otp,
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/auth/otp/verify", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.server.Test(req)
		require.NoError(t, err)
		resp.Body.Close()
		assert.Equal(t, 200, resp.StatusCode)

		// Verify session was deleted
		_, err = db.Builder[*fs.Session](dbc).
			Where(db.EQ("id", sessionID)).
			First(context.Background())
		assert.True(t, db.IsNotFound(err))
	}
}

func testOTPAttemptIncrement(dbc db.Client) func(t *testing.T) {
	return func(t *testing.T) {
		clearOTPSessions(dbc)
		app := createTestApp(t, dbc, &fs.OTPConfig{
			Enabled:     true,
			Length:      6,
			Expiration:  300,
			MaxAttempts: 5,
		})

		otp := "999999"
		otpHash, _ := auth.HashOTP(otp)
		expiresAt := time.Now().Add(5 * time.Minute)

		sessionID := utils.Must(uuid.NewV7())
		_, err := db.Builder[*fs.Session](dbc).Create(context.Background(), entity.New().
			Set("id", sessionID).
			Set("user_id", app.testUser.ID).
			Set("type", string(fs.SessionTypeOTPLogin)).
			Set("status", string(fs.SessionStatusPendingOTP)).
			Set("otp_hash", otpHash).
			Set("otp_attempts", 0).
			Set("expires_at", expiresAt))
		require.NoError(t, err)

		// Make multiple wrong attempts
		for i := 0; i < 3; i++ {
			reqBody := map[string]string{
				"session_id": sessionID.String(),
				"otp":        fmt.Sprintf("00000%d", i),
			}
			body, _ := json.Marshal(reqBody)
			req := httptest.NewRequest("POST", "/api/auth/otp/verify", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.server.Test(req)
			require.NoError(t, err)
			resp.Body.Close()
			assert.Equal(t, 422, resp.StatusCode) // Invalid OTP returns 422

			// Verify attempt count
			session, err := db.Builder[*fs.Session](dbc).
				Where(db.EQ("id", sessionID)).
				First(context.Background())
			require.NoError(t, err)
			assert.Equal(t, i+1, session.OTPAttempts)
		}

		// Correct OTP should still work (under max attempts)
		reqBody := map[string]string{
			"session_id": sessionID.String(),
			"otp":        otp,
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/auth/otp/verify", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.server.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, 200, resp.StatusCode)
	}
}

func testOTPDifferentProviderUser(dbc db.Client) func(t *testing.T) {
	return func(t *testing.T) {
		clearOTPSessions(dbc)
		app := createTestApp(t, dbc, &fs.OTPConfig{
			Enabled:     true,
			Length:      6,
			Expiration:  300,
			MaxAttempts: 3,
		})

		// Create a user with Google provider (lowercase email)
		userModel := utils.Must(dbc.Model("user"))
		randomSuffix := utils.RandomString(8)
		googleEmail := strings.ToLower("google" + randomSuffix + "@example.com")
		googleUserID, err := userModel.Create(context.Background(), entity.New().
			Set("username", "googleuser"+randomSuffix).
			Set("email", googleEmail).
			Set("provider", "google").
			Set("provider_id", "google_"+randomSuffix). // Use unique provider_id
			Set("active", true).
			Set("roles", []*entity.Entity{entity.New(2)}))
		require.NoError(t, err)

		// Request OTP for Google user
		reqBody := map[string]string{"email": googleEmail}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/auth/otp/request", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.server.Test(req)
		require.NoError(t, err)

		respBody := utils.Must(utils.ReadCloserToString(resp.Body))
		resp.Body.Close()
		assert.Equal(t, 200, resp.StatusCode)

		// Parse response to get session ID
		apiResp := parseResponse(t, []byte(respBody))
		var otpResponse auth.OTPResponse
		require.NoError(t, json.Unmarshal(apiResp.Data, &otpResponse))

		time.Sleep(100 * time.Millisecond)

		// Verify email was sent
		assert.NotNil(t, app.mailer.LastMail())
		assert.Equal(t, []string{googleEmail}, app.mailer.LastMail().To)

		// Verify session was created for Google user with correct user_id
		sessionID := utils.Must(uuid.Parse(otpResponse.SessionID))
		session, err := db.Builder[*fs.Session](dbc).
			Where(db.EQ("id", sessionID)).
			First(context.Background())
		require.NoError(t, err)
		assert.Equal(t, googleUserID, session.UserID)
	}
}

func testOTPFullLoginFlow(dbc db.Client) func(t *testing.T) {
	return func(t *testing.T) {
		clearOTPSessions(dbc)
		app := createTestApp(t, dbc, &fs.OTPConfig{
			Enabled:     true,
			Length:      6,
			Expiration:  300,
			MaxAttempts: 3,
		})

		// Step 1: Request OTP
		reqBody := map[string]string{"email": app.testUser.Email}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/auth/otp/request", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.server.Test(req)
		require.NoError(t, err)
		respBody := utils.Must(utils.ReadCloserToString(resp.Body))
		resp.Body.Close()
		assert.Equal(t, 200, resp.StatusCode)

		// Parse response to get session ID
		apiResp := parseResponse(t, []byte(respBody))
		var otpResponse auth.OTPResponse
		require.NoError(t, json.Unmarshal(apiResp.Data, &otpResponse))
		sessionID := otpResponse.SessionID

		time.Sleep(100 * time.Millisecond)

		// Step 2: Extract OTP from email
		mail := app.mailer.LastMail()
		require.NotNil(t, mail)

		// In test we need to create a session with known OTP since we can't extract from email
		clearOTPSessions(dbc)
		otp := "555555"
		otpHash, _ := auth.HashOTP(otp)
		expiresAt := time.Now().Add(5 * time.Minute)

		newSessionID := utils.Must(uuid.NewV7())
		_, err = db.Builder[*fs.Session](dbc).Create(context.Background(), entity.New().
			Set("id", newSessionID).
			Set("user_id", app.testUser.ID).
			Set("type", string(fs.SessionTypeOTPLogin)).
			Set("status", string(fs.SessionStatusPendingOTP)).
			Set("otp_hash", otpHash).
			Set("otp_attempts", 0).
			Set("expires_at", expiresAt))
		require.NoError(t, err)
		sessionID = newSessionID.String()

		// Step 3: Verify OTP and get tokens
		verifyBody := map[string]string{
			"session_id": sessionID,
			"otp":        otp,
		}
		body, _ = json.Marshal(verifyBody)
		req = httptest.NewRequest("POST", "/api/auth/otp/verify", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err = app.server.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, 200, resp.StatusCode)

		respBody = utils.Must(utils.ReadCloserToString(resp.Body))
		apiResp = parseResponse(t, []byte(respBody))

		var tokens fs.JWTTokens
		require.NoError(t, json.Unmarshal(apiResp.Data, &tokens))

		// Step 4: Access protected resource
		req = httptest.NewRequest("GET", "/api/protected", nil)
		req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)

		resp, err = app.server.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, 200, resp.StatusCode)

		// Step 5: Access /me endpoint
		req = httptest.NewRequest("GET", "/api/auth/me", nil)
		req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)

		resp, err = app.server.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, 200, resp.StatusCode)

		respBody = utils.Must(utils.ReadCloserToString(resp.Body))
		apiResp = parseResponse(t, []byte(respBody))

		var user fs.User
		require.NoError(t, json.Unmarshal(apiResp.Data, &user))
		assert.Equal(t, app.testUser.ID, user.ID)
		assert.Equal(t, app.testUser.Email, user.Email)

		// Step 6: Logout
		logoutBody := map[string]string{"refresh_token": tokens.RefreshToken}
		body, _ = json.Marshal(logoutBody)
		req = httptest.NewRequest("POST", "/api/auth/logout", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err = app.server.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, 200, resp.StatusCode)
	}
}

func testOTPCaseSensitivity(dbc db.Client) func(t *testing.T) {
	return func(t *testing.T) {
		clearOTPSessions(dbc)
		app := createTestApp(t, dbc, &fs.OTPConfig{
			Enabled:     true,
			Length:      6,
			Expiration:  300,
			MaxAttempts: 3,
		})

		// Request OTP with uppercase email
		upperEmail := strings.ToUpper(app.testUser.Email)
		reqBody := map[string]string{"email": upperEmail}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/auth/otp/request", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.server.Test(req)
		require.NoError(t, err)

		respBody := utils.Must(utils.ReadCloserToString(resp.Body))
		resp.Body.Close()
		assert.Equal(t, 200, resp.StatusCode)

		// Parse response to get session ID
		apiResp := parseResponse(t, []byte(respBody))
		var otpResponse auth.OTPResponse
		require.NoError(t, json.Unmarshal(apiResp.Data, &otpResponse))

		time.Sleep(100 * time.Millisecond)

		// Session should be created for the correct user
		sessionID := utils.Must(uuid.Parse(otpResponse.SessionID))
		session, err := db.Builder[*fs.Session](dbc).
			Where(db.EQ("id", sessionID)).
			Where(db.EQ("type", string(fs.SessionTypeOTPLogin))).
			Where(db.EQ("status", string(fs.SessionStatusPendingOTP))).
			First(context.Background())
		require.NoError(t, err)
		assert.Equal(t, app.testUser.ID, session.UserID)

		// Verify email was sent to the correct address
		assert.NotNil(t, app.mailer.LastMail())
	}
}

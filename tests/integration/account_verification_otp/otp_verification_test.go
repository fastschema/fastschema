package account_verification_otp_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
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
	schemaDir    = "../../../tests/integration/account_verification_otp/data/schemas"
	migrationDir = "../../../tests/integration/account_verification_otp/data/migrations"
	sqliteDSN    = "../../../tests/integration/account_verification_otp/data/account_verification_otp.db"
)

func TestAccountVerificationOTP(t *testing.T) {
	sb := utils.Must(schema.NewBuilderFromDir(schemaDir, systemSchemas...))
	dbc := utils.Must(entdbadapter.NewEntClient(&db.Config{
		Driver:       "sqlite",
		Name:         sqliteDSN,
		MigrationDir: migrationDir,
		LogQueries:   true,
	}, sb))
	defer dbc.Close()

	runAccountVerificationOTPTests(t, dbc)
}

func runAccountVerificationOTPTests(t *testing.T, dbc db.Client) {
	// Activation OTP Request Tests
	t.Run("ActivationOTPRequestSuccess", testActivationOTPRequestSuccess(dbc))
	t.Run("ActivationOTPRequestInvalidEmail", testActivationOTPRequestInvalidEmail(dbc))
	t.Run("ActivationOTPRequestNonExistentUser", testActivationOTPRequestNonExistentUser(dbc))
	t.Run("ActivationOTPRequestActiveUser", testActivationOTPRequestActiveUser(dbc))

	// Activation OTP Verify Tests
	t.Run("ActivationOTPVerifySuccess", testActivationOTPVerifySuccess(dbc))
	t.Run("ActivationOTPVerifyInvalidOTP", testActivationOTPVerifyInvalidOTP(dbc))
	t.Run("ActivationOTPVerifyExpired", testActivationOTPVerifyExpired(dbc))
	t.Run("ActivationOTPVerifyMaxAttempts", testActivationOTPVerifyMaxAttempts(dbc))
	t.Run("ActivationOTPVerifyInvalidSession", testActivationOTPVerifyInvalidSession(dbc))

	// Recovery OTP Request Tests
	t.Run("RecoveryOTPRequestSuccess", testRecoveryOTPRequestSuccess(dbc))
	t.Run("RecoveryOTPRequestNonExistentUser", testRecoveryOTPRequestNonExistentUser(dbc))
	t.Run("RecoveryOTPRequestInactiveUser", testRecoveryOTPRequestInactiveUser(dbc))

	// Recovery OTP Verify (RecoverCheck) Tests
	t.Run("RecoveryOTPVerifySuccess", testRecoveryOTPVerifySuccess(dbc))
	t.Run("RecoveryOTPVerifyInvalidOTP", testRecoveryOTPVerifyInvalidOTP(dbc))
	t.Run("RecoveryOTPVerifyExpired", testRecoveryOTPVerifyExpired(dbc))

	// Password Reset with Session Tests
	t.Run("PasswordResetWithSessionSuccess", testPasswordResetWithSessionSuccess(dbc))
	t.Run("PasswordResetWithUnverifiedSession", testPasswordResetWithUnverifiedSession(dbc))
	t.Run("PasswordResetPasswordMismatch", testPasswordResetPasswordMismatch(dbc))

	// Session Invalidation Tests
	t.Run("OTPSessionInvalidationOnResend", testOTPSessionInvalidationOnResend(dbc))
	t.Run("OTPSessionInvalidationOnRecoveryResend", testOTPSessionInvalidationOnRecoveryResend(dbc))

	// Full Flow Tests
	t.Run("FullActivationOTPFlow", testFullActivationOTPFlow(dbc))
	t.Run("FullRecoveryOTPFlow", testFullRecoveryOTPFlow(dbc))
}

// ============== Activation OTP Request Tests ==============

func testActivationOTPRequestSuccess(dbc db.Client) func(t *testing.T) {
	return func(t *testing.T) {
		clearSessions(dbc)
		app := createTestApp(t, dbc, &fs.OTPConfig{
			Enabled:     true,
			Length:      6,
			Expiration:  300,
			MaxAttempts: 3,
		}, "otp")

		reqBody := map[string]string{"email": app.inactiveUser.Email}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/auth/local/activate/send", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.server.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		respBody := utils.Must(utils.ReadCloserToString(resp.Body))
		assert.Equal(t, 200, resp.StatusCode, "Response: %s", respBody)

		// Parse response
		apiResp := parseResponse(t, []byte(respBody))
		var activation auth.Activation
		require.NoError(t, json.Unmarshal(apiResp.Data, &activation))

		assert.Equal(t, "email", activation.Activation)
		assert.NotEmpty(t, activation.SessionID)
		assert.Equal(t, 300, activation.ExpiresIn)

		// Verify session ID is a valid UUID
		sessionUUID, err := uuid.Parse(activation.SessionID)
		require.NoError(t, err)

		// Wait for async email
		time.Sleep(100 * time.Millisecond)

		// Check session was created
		session, sessionErr := db.Builder[*fs.Session](dbc).
			Where(db.EQ("id", sessionUUID)).
			Where(db.EQ("type", string(fs.SessionTypeActivation))).
			Where(db.EQ("status", string(fs.SessionStatusPendingOTP))).
			First(context.Background())
		require.NoError(t, sessionErr)
		assert.NotEmpty(t, session.OTPHash)
		assert.Equal(t, 0, session.OTPAttempts)
		assert.Equal(t, app.inactiveUser.ID, session.UserID)

		// Verify email was sent
		require.NotNil(t, app.mailer.LastMail())
		assert.Equal(t, []string{app.inactiveUser.Email}, app.mailer.LastMail().To)
		assert.Contains(t, app.mailer.LastMail().Subject, "activation")
	}
}

func testActivationOTPRequestInvalidEmail(dbc db.Client) func(t *testing.T) {
	return func(t *testing.T) {
		clearSessions(dbc)
		app := createTestApp(t, dbc, &fs.OTPConfig{
			Enabled:     true,
			Length:      6,
			Expiration:  300,
			MaxAttempts: 3,
		}, "otp")

		reqBody := map[string]string{"email": "not-an-email"}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/auth/local/activate/send", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.server.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 422, resp.StatusCode)
		assert.Nil(t, app.mailer.LastMail())
	}
}

func testActivationOTPRequestNonExistentUser(dbc db.Client) func(t *testing.T) {
	return func(t *testing.T) {
		clearSessions(dbc)
		app := createTestApp(t, dbc, &fs.OTPConfig{
			Enabled:     true,
			Length:      6,
			Expiration:  300,
			MaxAttempts: 3,
		}, "otp")

		// Non-existent user should still return success (no enumeration)
		reqBody := map[string]string{"email": "nonexistent@example.com"}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/auth/local/activate/send", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.server.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		respBody := utils.Must(utils.ReadCloserToString(resp.Body))
		assert.Equal(t, 200, resp.StatusCode, "Response: %s", respBody)

		time.Sleep(100 * time.Millisecond)

		// No email should be sent
		assert.Nil(t, app.mailer.LastMail())

		// No session should be created
		count, _ := db.Builder[*fs.Session](dbc).
			Where(db.EQ("type", string(fs.SessionTypeActivation))).
			Count(context.Background())
		assert.Equal(t, 0, count)
	}
}

func testActivationOTPRequestActiveUser(dbc db.Client) func(t *testing.T) {
	return func(t *testing.T) {
		clearSessions(dbc)
		app := createTestApp(t, dbc, &fs.OTPConfig{
			Enabled:     true,
			Length:      6,
			Expiration:  300,
			MaxAttempts: 3,
		}, "otp")

		// Active user should still return success (no enumeration)
		reqBody := map[string]string{"email": app.activeUser.Email}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/auth/local/activate/send", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.server.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		respBody := utils.Must(utils.ReadCloserToString(resp.Body))
		assert.Equal(t, 200, resp.StatusCode, "Response: %s", respBody)

		time.Sleep(100 * time.Millisecond)

		// No email should be sent for already active user
		assert.Nil(t, app.mailer.LastMail())
	}
}

// ============== Activation OTP Verify Tests ==============

func testActivationOTPVerifySuccess(dbc db.Client) func(t *testing.T) {
	return func(t *testing.T) {
		clearSessions(dbc)
		app := createTestApp(t, dbc, &fs.OTPConfig{
			Enabled:     true,
			Length:      6,
			Expiration:  300,
			MaxAttempts: 3,
		}, "otp")

		// Create OTP session
		otp := "123456"
		otpHash, _ := auth.HashOTP(otp)
		sessionUUID := utils.Must(uuid.NewV7())

		_, err := db.Builder[*fs.Session](dbc).Create(context.Background(), entity.New().
			Set("id", sessionUUID).
			Set("user_id", app.inactiveUser.ID).
			Set("type", string(fs.SessionTypeActivation)).
			Set("status", string(fs.SessionStatusPendingOTP)).
			Set("otp_hash", otpHash).
			Set("otp_attempts", 0).
			Set("expires_at", time.Now().Add(5*time.Minute)))
		require.NoError(t, err)

		// Verify OTP
		reqBody := map[string]string{
			"session_id": sessionUUID.String(),
			"otp":        otp,
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/auth/local/activate", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.server.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		respBody := utils.Must(utils.ReadCloserToString(resp.Body))
		assert.Equal(t, 200, resp.StatusCode, "Response: %s", respBody)

		// Parse response
		apiResp := parseResponse(t, []byte(respBody))
		var activation auth.Activation
		require.NoError(t, json.Unmarshal(apiResp.Data, &activation))
		assert.Equal(t, "activated", activation.Activation)

		// Verify user is now active
		userModel := utils.Must(dbc.Model("user"))
		user, _ := userModel.Query(db.EQ("id", app.inactiveUser.ID)).First(context.Background())
		assert.True(t, user.Get("active").(bool))

		// For activation, session is deleted on success
		_, err = db.Builder[*fs.Session](dbc).
			Where(db.EQ("id", sessionUUID)).
			First(context.Background())
		assert.True(t, db.IsNotFound(err), "Session should be deleted after successful activation")
	}
}

func testActivationOTPVerifyInvalidOTP(dbc db.Client) func(t *testing.T) {
	return func(t *testing.T) {
		clearSessions(dbc)
		app := createTestApp(t, dbc, &fs.OTPConfig{
			Enabled:     true,
			Length:      6,
			Expiration:  300,
			MaxAttempts: 3,
		}, "otp")

		// Create OTP session
		otp := "123456"
		otpHash, _ := auth.HashOTP(otp)
		sessionUUID := utils.Must(uuid.NewV7())

		_, err := db.Builder[*fs.Session](dbc).Create(context.Background(), entity.New().
			Set("id", sessionUUID).
			Set("user_id", app.inactiveUser.ID).
			Set("type", string(fs.SessionTypeActivation)).
			Set("status", string(fs.SessionStatusPendingOTP)).
			Set("otp_hash", otpHash).
			Set("otp_attempts", 0).
			Set("expires_at", time.Now().Add(5*time.Minute)))
		require.NoError(t, err)

		// Try with wrong OTP
		reqBody := map[string]string{
			"session_id": sessionUUID.String(),
			"otp":        "000000",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/auth/local/activate", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.server.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		respBody := utils.Must(utils.ReadCloserToString(resp.Body))
		assert.Equal(t, 422, resp.StatusCode, "Response: %s", respBody)
		assert.Contains(t, respBody, auth.MSG_OTP_INVALID)

		// Verify attempt was incremented
		session, _ := db.Builder[*fs.Session](dbc).
			Where(db.EQ("id", sessionUUID)).
			First(context.Background())
		assert.Equal(t, 1, session.OTPAttempts)
	}
}

func testActivationOTPVerifyExpired(dbc db.Client) func(t *testing.T) {
	return func(t *testing.T) {
		clearSessions(dbc)
		app := createTestApp(t, dbc, &fs.OTPConfig{
			Enabled:     true,
			Length:      6,
			Expiration:  300,
			MaxAttempts: 3,
		}, "otp")

		// Create expired OTP session
		otp := "123456"
		otpHash, _ := auth.HashOTP(otp)
		sessionUUID := utils.Must(uuid.NewV7())

		_, err := db.Builder[*fs.Session](dbc).Create(context.Background(), entity.New().
			Set("id", sessionUUID).
			Set("user_id", app.inactiveUser.ID).
			Set("type", string(fs.SessionTypeActivation)).
			Set("status", string(fs.SessionStatusPendingOTP)).
			Set("otp_hash", otpHash).
			Set("otp_attempts", 0).
			Set("expires_at", time.Now().Add(-5*time.Minute))) // Expired
		require.NoError(t, err)

		reqBody := map[string]string{
			"session_id": sessionUUID.String(),
			"otp":        otp,
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/auth/local/activate", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.server.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		respBody := utils.Must(utils.ReadCloserToString(resp.Body))
		assert.Equal(t, 422, resp.StatusCode)
		assert.Contains(t, respBody, auth.MSG_OTP_EXPIRED)
	}
}

func testActivationOTPVerifyMaxAttempts(dbc db.Client) func(t *testing.T) {
	return func(t *testing.T) {
		clearSessions(dbc)
		app := createTestApp(t, dbc, &fs.OTPConfig{
			Enabled:     true,
			Length:      6,
			Expiration:  300,
			MaxAttempts: 3,
		}, "otp")

		// Create OTP session with max attempts reached
		otp := "123456"
		otpHash, _ := auth.HashOTP(otp)
		sessionUUID := utils.Must(uuid.NewV7())

		_, err := db.Builder[*fs.Session](dbc).Create(context.Background(), entity.New().
			Set("id", sessionUUID).
			Set("user_id", app.inactiveUser.ID).
			Set("type", string(fs.SessionTypeActivation)).
			Set("status", string(fs.SessionStatusPendingOTP)).
			Set("otp_hash", otpHash).
			Set("otp_attempts", 3). // Max attempts
			Set("expires_at", time.Now().Add(5*time.Minute)))
		require.NoError(t, err)

		reqBody := map[string]string{
			"session_id": sessionUUID.String(),
			"otp":        otp,
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/auth/local/activate", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.server.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		respBody := utils.Must(utils.ReadCloserToString(resp.Body))
		assert.Equal(t, 429, resp.StatusCode)
		assert.Contains(t, respBody, auth.MSG_OTP_MAX_ATTEMPTS)
	}
}

func testActivationOTPVerifyInvalidSession(dbc db.Client) func(t *testing.T) {
	return func(t *testing.T) {
		clearSessions(dbc)
		app := createTestApp(t, dbc, &fs.OTPConfig{
			Enabled:     true,
			Length:      6,
			Expiration:  300,
			MaxAttempts: 3,
		}, "otp")

		reqBody := map[string]string{
			"session_id": "invalid-uuid",
			"otp":        "123456",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/auth/local/activate", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.server.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 400, resp.StatusCode)
	}
}

// ============== Recovery OTP Request Tests ==============

func testRecoveryOTPRequestSuccess(dbc db.Client) func(t *testing.T) {
	return func(t *testing.T) {
		clearSessions(dbc)
		app := createTestApp(t, dbc, &fs.OTPConfig{
			Enabled:     true,
			Length:      6,
			Expiration:  300,
			MaxAttempts: 3,
		}, "otp")

		reqBody := map[string]string{"email": app.activeUser.Email}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/auth/local/recover", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.server.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		respBody := utils.Must(utils.ReadCloserToString(resp.Body))
		assert.Equal(t, 200, resp.StatusCode, "Response: %s", respBody)

		// Parse response
		apiResp := parseResponse(t, []byte(respBody))
		var activation auth.Activation
		require.NoError(t, json.Unmarshal(apiResp.Data, &activation))

		assert.Equal(t, "sent", activation.Activation)
		assert.NotEmpty(t, activation.SessionID)
		assert.Equal(t, 300, activation.ExpiresIn)

		// Wait for async email
		time.Sleep(100 * time.Millisecond)

		// Verify email was sent
		require.NotNil(t, app.mailer.LastMail())
		assert.Equal(t, []string{app.activeUser.Email}, app.mailer.LastMail().To)
		assert.Contains(t, app.mailer.LastMail().Subject, "password reset")
	}
}

func testRecoveryOTPRequestNonExistentUser(dbc db.Client) func(t *testing.T) {
	return func(t *testing.T) {
		clearSessions(dbc)
		app := createTestApp(t, dbc, &fs.OTPConfig{
			Enabled:     true,
			Length:      6,
			Expiration:  300,
			MaxAttempts: 3,
		}, "otp")

		reqBody := map[string]string{"email": "nonexistent@example.com"}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/auth/local/recover", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.server.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		respBody := utils.Must(utils.ReadCloserToString(resp.Body))
		assert.Equal(t, 200, resp.StatusCode, "Response: %s", respBody)

		time.Sleep(100 * time.Millisecond)

		// No email should be sent
		assert.Nil(t, app.mailer.LastMail())
	}
}

func testRecoveryOTPRequestInactiveUser(dbc db.Client) func(t *testing.T) {
	return func(t *testing.T) {
		clearSessions(dbc)
		app := createTestApp(t, dbc, &fs.OTPConfig{
			Enabled:     true,
			Length:      6,
			Expiration:  300,
			MaxAttempts: 3,
		}, "otp")

		reqBody := map[string]string{"email": app.inactiveUser.Email}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/auth/local/recover", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.server.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		respBody := utils.Must(utils.ReadCloserToString(resp.Body))
		assert.Equal(t, 200, resp.StatusCode, "Response: %s", respBody)

		time.Sleep(100 * time.Millisecond)

		// Recovery emails ARE sent for inactive users (they may need to recover their account)
		require.NotNil(t, app.mailer.LastMail())
		assert.Equal(t, []string{app.inactiveUser.Email}, app.mailer.LastMail().To)
	}
}

// ============== Recovery OTP Verify Tests ==============

func testRecoveryOTPVerifySuccess(dbc db.Client) func(t *testing.T) {
	return func(t *testing.T) {
		clearSessions(dbc)
		app := createTestApp(t, dbc, &fs.OTPConfig{
			Enabled:     true,
			Length:      6,
			Expiration:  300,
			MaxAttempts: 3,
		}, "otp")

		// Create recovery OTP session
		otp := "654321"
		otpHash, _ := auth.HashOTP(otp)
		sessionUUID := utils.Must(uuid.NewV7())

		_, err := db.Builder[*fs.Session](dbc).Create(context.Background(), entity.New().
			Set("id", sessionUUID).
			Set("user_id", app.activeUser.ID).
			Set("type", string(fs.SessionTypeRecovery)).
			Set("status", string(fs.SessionStatusPendingOTP)).
			Set("otp_hash", otpHash).
			Set("otp_attempts", 0).
			Set("expires_at", time.Now().Add(5*time.Minute)))
		require.NoError(t, err)

		// Verify OTP via recover/check
		reqBody := map[string]string{
			"session_id": sessionUUID.String(),
			"otp":        otp,
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/auth/local/recover/check", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.server.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		respBody := utils.Must(utils.ReadCloserToString(resp.Body))
		assert.Equal(t, 200, resp.StatusCode, "Response: %s", respBody)

		// Parse response
		apiResp := parseResponse(t, []byte(respBody))
		var activation auth.Activation
		require.NoError(t, json.Unmarshal(apiResp.Data, &activation))
		assert.True(t, activation.Verified)

		// Verify session is now verified
		session, _ := db.Builder[*fs.Session](dbc).
			Where(db.EQ("id", sessionUUID)).
			First(context.Background())
		assert.Equal(t, string(fs.SessionStatusVerified), session.Status)
	}
}

func testRecoveryOTPVerifyInvalidOTP(dbc db.Client) func(t *testing.T) {
	return func(t *testing.T) {
		clearSessions(dbc)
		app := createTestApp(t, dbc, &fs.OTPConfig{
			Enabled:     true,
			Length:      6,
			Expiration:  300,
			MaxAttempts: 3,
		}, "otp")

		otp := "654321"
		otpHash, _ := auth.HashOTP(otp)
		sessionUUID := utils.Must(uuid.NewV7())

		_, err := db.Builder[*fs.Session](dbc).Create(context.Background(), entity.New().
			Set("id", sessionUUID).
			Set("user_id", app.activeUser.ID).
			Set("type", string(fs.SessionTypeRecovery)).
			Set("status", string(fs.SessionStatusPendingOTP)).
			Set("otp_hash", otpHash).
			Set("otp_attempts", 0).
			Set("expires_at", time.Now().Add(5*time.Minute)))
		require.NoError(t, err)

		reqBody := map[string]string{
			"session_id": sessionUUID.String(),
			"otp":        "000000",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/auth/local/recover/check", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.server.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		respBody := utils.Must(utils.ReadCloserToString(resp.Body))
		assert.Equal(t, 422, resp.StatusCode, "Response: %s", respBody)
		assert.Contains(t, respBody, auth.MSG_OTP_INVALID)
	}
}

func testRecoveryOTPVerifyExpired(dbc db.Client) func(t *testing.T) {
	return func(t *testing.T) {
		clearSessions(dbc)
		app := createTestApp(t, dbc, &fs.OTPConfig{
			Enabled:     true,
			Length:      6,
			Expiration:  300,
			MaxAttempts: 3,
		}, "otp")

		otp := "654321"
		otpHash, _ := auth.HashOTP(otp)
		sessionUUID := utils.Must(uuid.NewV7())

		_, err := db.Builder[*fs.Session](dbc).Create(context.Background(), entity.New().
			Set("id", sessionUUID).
			Set("user_id", app.activeUser.ID).
			Set("type", string(fs.SessionTypeRecovery)).
			Set("status", string(fs.SessionStatusPendingOTP)).
			Set("otp_hash", otpHash).
			Set("otp_attempts", 0).
			Set("expires_at", time.Now().Add(-5*time.Minute))) // Expired
		require.NoError(t, err)

		reqBody := map[string]string{
			"session_id": sessionUUID.String(),
			"otp":        otp,
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/auth/local/recover/check", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.server.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		respBody := utils.Must(utils.ReadCloserToString(resp.Body))
		assert.Equal(t, 422, resp.StatusCode)
		assert.Contains(t, respBody, auth.MSG_OTP_EXPIRED)
	}
}

// ============== Password Reset Tests ==============

func testPasswordResetWithSessionSuccess(dbc db.Client) func(t *testing.T) {
	return func(t *testing.T) {
		clearSessions(dbc)
		app := createTestApp(t, dbc, &fs.OTPConfig{
			Enabled:     true,
			Length:      6,
			Expiration:  300,
			MaxAttempts: 3,
		}, "otp")

		// Create verified recovery session
		sessionUUID := utils.Must(uuid.NewV7())
		_, err := db.Builder[*fs.Session](dbc).Create(context.Background(), entity.New().
			Set("id", sessionUUID).
			Set("user_id", app.activeUser.ID).
			Set("type", string(fs.SessionTypeRecovery)).
			Set("status", string(fs.SessionStatusVerified)).
			Set("expires_at", time.Now().Add(5*time.Minute)))
		require.NoError(t, err)

		reqBody := map[string]string{
			"session_id":       sessionUUID.String(),
			"password":         "newpassword123",
			"confirm_password": "newpassword123",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/auth/local/recover/reset", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.server.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		respBody := utils.Must(utils.ReadCloserToString(resp.Body))
		assert.Equal(t, 200, resp.StatusCode, "Response: %s", respBody)

		// Verify session is now inactive
		session, _ := db.Builder[*fs.Session](dbc).
			Where(db.EQ("id", sessionUUID)).
			First(context.Background())
		assert.Equal(t, string(fs.SessionStatusInactive), session.Status)
	}
}

func testPasswordResetWithUnverifiedSession(dbc db.Client) func(t *testing.T) {
	return func(t *testing.T) {
		clearSessions(dbc)
		app := createTestApp(t, dbc, &fs.OTPConfig{
			Enabled:     true,
			Length:      6,
			Expiration:  300,
			MaxAttempts: 3,
		}, "otp")

		// Create unverified (pending_otp) recovery session
		sessionUUID := utils.Must(uuid.NewV7())
		_, err := db.Builder[*fs.Session](dbc).Create(context.Background(), entity.New().
			Set("id", sessionUUID).
			Set("user_id", app.activeUser.ID).
			Set("type", string(fs.SessionTypeRecovery)).
			Set("status", string(fs.SessionStatusPendingOTP)). // Not verified
			Set("otp_hash", "somehash").
			Set("expires_at", time.Now().Add(5*time.Minute)))
		require.NoError(t, err)

		reqBody := map[string]string{
			"session_id":       sessionUUID.String(),
			"password":         "newpassword123",
			"confirm_password": "newpassword123",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/auth/local/recover/reset", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.server.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		respBody := utils.Must(utils.ReadCloserToString(resp.Body))
		assert.Equal(t, 400, resp.StatusCode, "Response: %s", respBody)
		assert.Contains(t, respBody, auth.MSG_OTP_VERIFICATION_REQUIRED)
	}
}

func testPasswordResetPasswordMismatch(dbc db.Client) func(t *testing.T) {
	return func(t *testing.T) {
		clearSessions(dbc)
		app := createTestApp(t, dbc, &fs.OTPConfig{
			Enabled:     true,
			Length:      6,
			Expiration:  300,
			MaxAttempts: 3,
		}, "otp")

		sessionUUID := utils.Must(uuid.NewV7())
		_, err := db.Builder[*fs.Session](dbc).Create(context.Background(), entity.New().
			Set("id", sessionUUID).
			Set("user_id", app.activeUser.ID).
			Set("type", string(fs.SessionTypeRecovery)).
			Set("status", string(fs.SessionStatusVerified)).
			Set("expires_at", time.Now().Add(5*time.Minute)))
		require.NoError(t, err)

		reqBody := map[string]string{
			"session_id":       sessionUUID.String(),
			"password":         "password1",
			"confirm_password": "password2",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/auth/local/recover/reset", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.server.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 422, resp.StatusCode)
	}
}

// ============== Session Invalidation Tests ==============

func testOTPSessionInvalidationOnResend(dbc db.Client) func(t *testing.T) {
	return func(t *testing.T) {
		clearSessions(dbc)
		app := createTestApp(t, dbc, &fs.OTPConfig{
			Enabled:     true,
			Length:      6,
			Expiration:  300,
			MaxAttempts: 3,
		}, "otp")

		// Create existing activation session
		oldSessionUUID := utils.Must(uuid.NewV7())
		_, err := db.Builder[*fs.Session](dbc).Create(context.Background(), entity.New().
			Set("id", oldSessionUUID).
			Set("user_id", app.inactiveUser.ID).
			Set("type", string(fs.SessionTypeActivation)).
			Set("status", string(fs.SessionStatusPendingOTP)).
			Set("otp_hash", "oldhash").
			Set("otp_attempts", 0).
			Set("expires_at", time.Now().Add(5*time.Minute)))
		require.NoError(t, err)

		// Request new OTP
		reqBody := map[string]string{"email": app.inactiveUser.Email}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/auth/local/activate/send", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.server.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, 200, resp.StatusCode)

		// Verify old session is invalidated
		oldSession, _ := db.Builder[*fs.Session](dbc).
			Where(db.EQ("id", oldSessionUUID)).
			First(context.Background())
		assert.Equal(t, string(fs.SessionStatusInactive), oldSession.Status)
	}
}

func testOTPSessionInvalidationOnRecoveryResend(dbc db.Client) func(t *testing.T) {
	return func(t *testing.T) {
		clearSessions(dbc)
		app := createTestApp(t, dbc, &fs.OTPConfig{
			Enabled:     true,
			Length:      6,
			Expiration:  300,
			MaxAttempts: 3,
		}, "otp")

		// Create existing recovery session
		oldSessionUUID := utils.Must(uuid.NewV7())
		_, err := db.Builder[*fs.Session](dbc).Create(context.Background(), entity.New().
			Set("id", oldSessionUUID).
			Set("user_id", app.activeUser.ID).
			Set("type", string(fs.SessionTypeRecovery)).
			Set("status", string(fs.SessionStatusPendingOTP)).
			Set("otp_hash", "oldhash").
			Set("otp_attempts", 0).
			Set("expires_at", time.Now().Add(5*time.Minute)))
		require.NoError(t, err)

		// Request new OTP
		reqBody := map[string]string{"email": app.activeUser.Email}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/auth/local/recover", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.server.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, 200, resp.StatusCode)

		// Verify old session is invalidated
		oldSession, _ := db.Builder[*fs.Session](dbc).
			Where(db.EQ("id", oldSessionUUID)).
			First(context.Background())
		assert.Equal(t, string(fs.SessionStatusInactive), oldSession.Status)
	}
}

// ============== Full Flow Tests ==============

func testFullActivationOTPFlow(dbc db.Client) func(t *testing.T) {
	return func(t *testing.T) {
		clearSessions(dbc)
		app := createTestApp(t, dbc, &fs.OTPConfig{
			Enabled:     true,
			Length:      6,
			Expiration:  300,
			MaxAttempts: 3,
		}, "otp")

		// Step 1: Request activation OTP
		reqBody := map[string]string{"email": app.inactiveUser.Email}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/auth/local/activate/send", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.server.Test(req)
		require.NoError(t, err)
		respBody := utils.Must(utils.ReadCloserToString(resp.Body))
		resp.Body.Close()
		assert.Equal(t, 200, resp.StatusCode, "Response: %s", respBody)

		apiResp := parseResponse(t, []byte(respBody))
		var activation auth.Activation
		require.NoError(t, json.Unmarshal(apiResp.Data, &activation))
		sessionID := activation.SessionID

		// Wait for email
		time.Sleep(100 * time.Millisecond)
		require.NotNil(t, app.mailer.LastMail())

		// For this test, we'll update the session with a known OTP hash
		sessionUUID, _ := uuid.Parse(sessionID)
		knownOTP := "123456"
		otpHash, _ := auth.HashOTP(knownOTP)
		_, _ = db.Builder[*fs.Session](dbc).
			Where(db.EQ("id", sessionUUID)).
			Update(context.Background(), entity.New().Set("otp_hash", otpHash))

		// Step 2: Verify OTP
		verifyBody := map[string]string{
			"session_id": sessionID,
			"otp":        knownOTP,
		}
		body, _ = json.Marshal(verifyBody)
		req = httptest.NewRequest("POST", "/api/auth/local/activate", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err = app.server.Test(req)
		require.NoError(t, err)
		respBody = utils.Must(utils.ReadCloserToString(resp.Body))
		resp.Body.Close()
		assert.Equal(t, 200, resp.StatusCode, "Response: %s", respBody)

		// Verify user is now active
		userModel := utils.Must(dbc.Model("user"))
		user, _ := userModel.Query(db.EQ("id", app.inactiveUser.ID)).First(context.Background())
		assert.True(t, user.Get("active").(bool))

		// For activation, session is deleted on success
		_, err = db.Builder[*fs.Session](dbc).
			Where(db.EQ("id", sessionUUID)).
			First(context.Background())
		assert.True(t, db.IsNotFound(err), "Session should be deleted after successful activation")
	}
}

func testFullRecoveryOTPFlow(dbc db.Client) func(t *testing.T) {
	return func(t *testing.T) {
		clearSessions(dbc)
		app := createTestApp(t, dbc, &fs.OTPConfig{
			Enabled:     true,
			Length:      6,
			Expiration:  300,
			MaxAttempts: 3,
		}, "otp")

		// Step 1: Request recovery OTP
		reqBody := map[string]string{"email": app.activeUser.Email}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/auth/local/recover", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.server.Test(req)
		require.NoError(t, err)
		respBody := utils.Must(utils.ReadCloserToString(resp.Body))
		resp.Body.Close()
		assert.Equal(t, 200, resp.StatusCode, "Response: %s", respBody)

		apiResp := parseResponse(t, []byte(respBody))
		var activation auth.Activation
		require.NoError(t, json.Unmarshal(apiResp.Data, &activation))
		sessionID := activation.SessionID

		// Wait for email
		time.Sleep(100 * time.Millisecond)
		require.NotNil(t, app.mailer.LastMail())

		// Update session with known OTP
		sessionUUID, _ := uuid.Parse(sessionID)
		knownOTP := "654321"
		otpHash, _ := auth.HashOTP(knownOTP)
		_, _ = db.Builder[*fs.Session](dbc).
			Where(db.EQ("id", sessionUUID)).
			Update(context.Background(), entity.New().Set("otp_hash", otpHash))

		// Step 2: Verify OTP
		verifyBody := map[string]string{
			"session_id": sessionID,
			"otp":        knownOTP,
		}
		body, _ = json.Marshal(verifyBody)
		req = httptest.NewRequest("POST", "/api/auth/local/recover/check", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err = app.server.Test(req)
		require.NoError(t, err)
		respBody = utils.Must(utils.ReadCloserToString(resp.Body))
		resp.Body.Close()
		assert.Equal(t, 200, resp.StatusCode, "Response: %s", respBody)

		apiResp = parseResponse(t, []byte(respBody))
		var verifyResp auth.Activation
		require.NoError(t, json.Unmarshal(apiResp.Data, &verifyResp))
		assert.True(t, verifyResp.Verified)

		// Step 3: Reset password
		resetBody := map[string]string{
			"session_id":       sessionID,
			"password":         "newSecurePassword123",
			"confirm_password": "newSecurePassword123",
		}
		body, _ = json.Marshal(resetBody)
		req = httptest.NewRequest("POST", "/api/auth/local/recover/reset", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err = app.server.Test(req)
		require.NoError(t, err)
		respBody = utils.Must(utils.ReadCloserToString(resp.Body))
		resp.Body.Close()
		assert.Equal(t, 200, resp.StatusCode, "Response: %s", respBody)

		// Verify session is inactive
		session, _ := db.Builder[*fs.Session](dbc).
			Where(db.EQ("id", sessionUUID)).
			First(context.Background())
		assert.Equal(t, string(fs.SessionStatusInactive), session.Status)
	}
}

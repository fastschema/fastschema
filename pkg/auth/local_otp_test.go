package auth_test

import (
	"bytes"
	"context"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/entity"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/auth"
	entdbadapter "github.com/fastschema/fastschema/pkg/entdbadapter"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func createOTPLocalAuthProvider(config *testAppConfig) *auth.LocalProvider {
	schemasDir := utils.Must(os.MkdirTemp("", "schemas"))
	migrationsDir := utils.Must(os.MkdirTemp("", "migrations"))
	sb := utils.Must(schema.NewBuilderFromDir(schemasDir, fs.SystemSchemaTypes...))
	config.db = utils.Must(entdbadapter.NewTestClient(migrationsDir, sb))
	redirectURL := "http://localhost:8080/api/auth/local/callback"

	if config.key == "" {
		config.key = testKey
	}

	roleModel := utils.Must(config.db.Model("role"))
	userModel := utils.Must(config.db.Model("user"))
	utils.Must(roleModel.CreateFromJSON(context.Background(), `{"name": "admin"}`))
	utils.Must(roleModel.CreateFromJSON(context.Background(), `{"name": "user"}`))

	if config.createData {
		// inactive user for activation tests
		utils.Must(userModel.CreateFromJSON(context.Background(), `{
			"username": "inactiveuser",
			"password": "inactiveuser",
			"email": "inactive@site.local",
			"provider": "local",
			"active": false,
			"roles": [{"id": 1}]
		}`))
		// active user for recovery tests
		utils.Must(userModel.CreateFromJSON(context.Background(), `{
			"username": "activeuser",
			"password": "activeuser",
			"email": "active@site.local",
			"provider": "local",
			"active": true,
			"roles": [{"id": 1}]
		}`))
	}

	authProvider := utils.Must(auth.NewLocalAuthProvider(fs.Map{
		"activation_method":   config.activation,
		"verification_method": "otp",
	}, redirectURL))
	localAuthProvider := authProvider.(*auth.LocalProvider)
	localAuthProvider.Init(
		func() db.Client { return config.db },
		func() string { return config.key },
		func() string { return "testApp" },
		func() string { return "http://localhost:8080" },
		func(names ...string) fs.Mailer { return config.mailer },
		nil,
		func() *fs.OTPConfig {
			return &fs.OTPConfig{
				Enabled:     true,
				Length:      6,
				Expiration:  300,
				MaxAttempts: 3,
			}
		},
	)

	return localAuthProvider
}

func TestLocalAuthOTPActivationRequest(t *testing.T) {
	mailer := &MockMailer{}
	config := &testAppConfig{activation: "email", createData: true, mailer: mailer}
	provider := createOTPLocalAuthProvider(config)
	server := createServer(t, fs.Post(
		"/user/activate/send",
		provider.SendActivationLink,
		&fs.Meta{Public: true},
	))

	// Case 1: Invalid email
	{
		req := httptest.NewRequest(
			"POST", "/user/activate/send",
			bytes.NewReader([]byte(`{"email": "invalid"}`)),
		)
		resp, _ := server.Test(req)
		defer func() { assert.NoError(t, resp.Body.Close()) }()
		assert.Equal(t, 422, resp.StatusCode)
	}

	// Case 2: Non-existent user (should still return success for security)
	{
		req := httptest.NewRequest(
			"POST", "/user/activate/send",
			bytes.NewReader([]byte(`{"email": "nonexistent@site.local"}`)),
		)
		resp, _ := server.Test(req)
		defer func() { assert.NoError(t, resp.Body.Close()) }()
		assert.Equal(t, 200, resp.StatusCode)
		body := utils.Must(utils.ReadCloserToString(resp.Body))
		assert.Contains(t, body, `"session_id"`)
		assert.Contains(t, body, `"expires_in"`)
	}

	// Case 3: Success - inactive user
	{
		req := httptest.NewRequest(
			"POST", "/user/activate/send",
			bytes.NewReader([]byte(`{"email": "inactive@site.local"}`)),
		)
		resp, _ := server.Test(req)
		defer func() { assert.NoError(t, resp.Body.Close()) }()
		assert.Equal(t, 200, resp.StatusCode)
		body := utils.Must(utils.ReadCloserToString(resp.Body))
		assert.Contains(t, body, `"session_id"`)
		assert.Contains(t, body, `"expires_in"`)
		assert.Contains(t, body, `"activation":"email"`)
	}
}

func TestLocalAuthOTPActivationVerify(t *testing.T) {
	mailer := &MockMailer{}
	config := &testAppConfig{activation: "email", createData: true, mailer: mailer}
	provider := createOTPLocalAuthProvider(config)

	// First, create an OTP session for the inactive user
	sessionModel := utils.Must(config.db.Model("session"))
	otp := "123456"
	otpHash := utils.Must(auth.HashOTP(otp))
	sessionUUID := utils.Must(uuid.NewV7())
	expiresAt := time.Now().Add(5 * time.Minute)

	utils.Must(sessionModel.Create(context.Background(), entity.New().
		Set("id", sessionUUID).
		Set("user_id", uint64(1)). // inactive user
		Set("type", string(fs.SessionTypeActivation)).
		Set("status", string(fs.SessionStatusPendingOTP)).
		Set("otp_hash", otpHash).
		Set("otp_attempts", 0).
		Set("expires_at", expiresAt),
	))

	server := createServer(t, fs.Post(
		"/user/activate",
		provider.Activate,
		&fs.Meta{Public: true},
	))

	// Case 1: Invalid session ID
	{
		req := httptest.NewRequest(
			"POST", "/user/activate",
			bytes.NewReader([]byte(`{"session_id": "invalid", "otp": "123456"}`)),
		)
		resp, _ := server.Test(req)
		defer func() { assert.NoError(t, resp.Body.Close()) }()
		assert.Equal(t, 400, resp.StatusCode)
	}

	// Case 2: Invalid OTP
	{
		req := httptest.NewRequest(
			"POST", "/user/activate",
			bytes.NewReader([]byte(`{"session_id": "`+sessionUUID.String()+`", "otp": "000000"}`)),
		)
		resp, _ := server.Test(req)
		defer func() { assert.NoError(t, resp.Body.Close()) }()
		assert.Equal(t, 422, resp.StatusCode)
		assert.Contains(t, utils.Must(utils.ReadCloserToString(resp.Body)), auth.MSG_OTP_INVALID)
	}

	// Case 3: Valid OTP - successful activation
	{
		// Need to recreate session since previous test incremented attempts
		_, _ = db.Builder[*fs.Session](config.db).
			Where(db.EQ("id", sessionUUID)).
			Update(context.Background(), entity.New().Set("otp_attempts", 0))

		req := httptest.NewRequest(
			"POST", "/user/activate",
			bytes.NewReader([]byte(`{"session_id": "`+sessionUUID.String()+`", "otp": "`+otp+`"}`)),
		)
		resp, _ := server.Test(req)
		defer func() { assert.NoError(t, resp.Body.Close()) }()
		assert.Equal(t, 200, resp.StatusCode)
		assert.Contains(t, utils.Must(utils.ReadCloserToString(resp.Body)), `"activation":"activated"`)
	}
}

func TestLocalAuthOTPRecoveryFlow(t *testing.T) {
	mailer := &MockMailer{}
	config := &testAppConfig{activation: "email", createData: true, mailer: mailer}
	provider := createOTPLocalAuthProvider(config)

	// Test recovery request
	recoverServer := createServer(t, fs.Post(
		"/user/recover",
		provider.Recover,
		&fs.Meta{Public: true},
	))

	// Case 1: Request OTP for recovery
	{
		req := httptest.NewRequest(
			"POST", "/user/recover",
			bytes.NewReader([]byte(`{"email": "active@site.local"}`)),
		)
		resp, _ := recoverServer.Test(req)
		defer func() { assert.NoError(t, resp.Body.Close()) }()
		assert.Equal(t, 200, resp.StatusCode)
		body := utils.Must(utils.ReadCloserToString(resp.Body))
		assert.Contains(t, body, `"session_id"`)
		assert.Contains(t, body, `"activation":"sent"`)
	}

	// Create a session manually for testing verification
	sessionModel := utils.Must(config.db.Model("session"))
	otp := "654321"
	otpHash := utils.Must(auth.HashOTP(otp))
	sessionUUID := utils.Must(uuid.NewV7())
	sessionID := sessionUUID.String()
	expiresAt := time.Now().Add(5 * time.Minute)

	utils.Must(sessionModel.Create(context.Background(), entity.New().
		Set("id", sessionUUID).
		Set("user_id", uint64(2)). // active user
		Set("type", string(fs.SessionTypeRecovery)).
		Set("status", string(fs.SessionStatusPendingOTP)).
		Set("otp_hash", otpHash).
		Set("otp_attempts", 0).
		Set("expires_at", expiresAt),
	))

	// Test recover check with OTP
	recoverCheckServer := createServer(t, fs.Post(
		"/user/recover/check",
		provider.RecoverCheck,
		&fs.Meta{Public: true},
	))

	// Case 2: Verify OTP
	{
		req := httptest.NewRequest(
			"POST", "/user/recover/check",
			bytes.NewReader([]byte(`{"session_id": "`+sessionID+`", "otp": "`+otp+`"}`)),
		)
		resp, _ := recoverCheckServer.Test(req)
		defer func() { assert.NoError(t, resp.Body.Close()) }()
		assert.Equal(t, 200, resp.StatusCode)
		body := utils.Must(utils.ReadCloserToString(resp.Body))
		assert.Contains(t, body, `"verified":true`)
	}

	// Test reset password with session
	resetServer := createServer(t, fs.Post(
		"/user/recover/reset",
		provider.ResetPassword,
		&fs.Meta{Public: true},
	))

	// Case 3: Reset password with verified session
	{
		req := httptest.NewRequest(
			"POST", "/user/recover/reset",
			bytes.NewReader([]byte(`{"session_id": "`+sessionID+`", "password": "newpassword", "confirm_password": "newpassword"}`)),
		)
		resp, _ := resetServer.Test(req)
		defer func() { assert.NoError(t, resp.Body.Close()) }()
		assert.Equal(t, 200, resp.StatusCode)
	}
}

func TestLocalAuthOTPMaxAttempts(t *testing.T) {
	mailer := &MockMailer{}
	config := &testAppConfig{activation: "email", createData: true, mailer: mailer}
	provider := createOTPLocalAuthProvider(config)

	// Create a session with max attempts reached
	sessionModel := utils.Must(config.db.Model("session"))
	otp := "123456"
	otpHash := utils.Must(auth.HashOTP(otp))
	sessionUUID := utils.Must(uuid.NewV7())
	expiresAt := time.Now().Add(5 * time.Minute)

	utils.Must(sessionModel.Create(context.Background(), entity.New().
		Set("id", sessionUUID).
		Set("user_id", uint64(1)).
		Set("type", string(fs.SessionTypeActivation)).
		Set("status", string(fs.SessionStatusPendingOTP)).
		Set("otp_hash", otpHash).
		Set("otp_attempts", 3). // max attempts reached
		Set("expires_at", expiresAt),
	))

	server := createServer(t, fs.Post(
		"/user/activate",
		provider.Activate,
		&fs.Meta{Public: true},
	))

	req := httptest.NewRequest(
		"POST", "/user/activate",
		bytes.NewReader([]byte(`{"session_id": "`+sessionUUID.String()+`", "otp": "`+otp+`"}`)),
	)
	resp, _ := server.Test(req)
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 429, resp.StatusCode)
	assert.Contains(t, utils.Must(utils.ReadCloserToString(resp.Body)), auth.MSG_OTP_MAX_ATTEMPTS)
}

func TestLocalAuthOTPExpired(t *testing.T) {
	mailer := &MockMailer{}
	config := &testAppConfig{activation: "email", createData: true, mailer: mailer}
	provider := createOTPLocalAuthProvider(config)

	// Create an expired session
	sessionModel := utils.Must(config.db.Model("session"))
	otp := "123456"
	otpHash := utils.Must(auth.HashOTP(otp))
	sessionUUID := utils.Must(uuid.NewV7())
	expiresAt := time.Now().Add(-5 * time.Minute) // expired

	utils.Must(sessionModel.Create(context.Background(), entity.New().
		Set("id", sessionUUID).
		Set("user_id", uint64(1)).
		Set("type", string(fs.SessionTypeActivation)).
		Set("status", string(fs.SessionStatusPendingOTP)).
		Set("otp_hash", otpHash).
		Set("otp_attempts", 0).
		Set("expires_at", expiresAt),
	))

	server := createServer(t, fs.Post(
		"/user/activate",
		provider.Activate,
		&fs.Meta{Public: true},
	))

	req := httptest.NewRequest(
		"POST", "/user/activate",
		bytes.NewReader([]byte(`{"session_id": "`+sessionUUID.String()+`", "otp": "`+otp+`"}`)),
	)
	resp, _ := server.Test(req)
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 422, resp.StatusCode)
	assert.Contains(t, utils.Must(utils.ReadCloserToString(resp.Body)), auth.MSG_OTP_EXPIRED)
}

func TestLocalAuthOTPInvalidatePreviousSessions(t *testing.T) {
	mailer := &MockMailer{}
	config := &testAppConfig{activation: "email", createData: true, mailer: mailer}
	provider := createOTPLocalAuthProvider(config)

	// Create an existing session
	sessionModel := utils.Must(config.db.Model("session"))
	otp := "123456"
	otpHash := utils.Must(auth.HashOTP(otp))
	oldSessionUUID := utils.Must(uuid.NewV7())
	expiresAt := time.Now().Add(5 * time.Minute)

	utils.Must(sessionModel.Create(context.Background(), entity.New().
		Set("id", oldSessionUUID).
		Set("user_id", uint64(1)).
		Set("type", string(fs.SessionTypeActivation)).
		Set("status", string(fs.SessionStatusPendingOTP)).
		Set("otp_hash", otpHash).
		Set("otp_attempts", 0).
		Set("expires_at", expiresAt),
	))

	server := createServer(t, fs.Post(
		"/user/activate/send",
		provider.SendActivationLink,
		&fs.Meta{Public: true},
	))

	// Request a new OTP - should invalidate the old session
	req := httptest.NewRequest(
		"POST", "/user/activate/send",
		bytes.NewReader([]byte(`{"email": "inactive@site.local"}`)),
	)
	resp, _ := server.Test(req)
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 200, resp.StatusCode)

	// Verify old session is invalidated
	oldSession, _ := db.Builder[*fs.Session](config.db).
		Where(db.EQ("id", oldSessionUUID)).
		First(context.Background())
	assert.Equal(t, string(fs.SessionStatusInactive), oldSession.Status)
}

func TestConfirmationHelpers(t *testing.T) {
	// Test IsOTPBased
	{
		c := &auth.Confirmation{SessionID: "123", OTP: "456"}
		assert.True(t, c.IsOTPBased())
		assert.False(t, c.IsTokenBased())
	}

	// Test IsTokenBased
	{
		c := &auth.Confirmation{Token: "abc123"}
		assert.True(t, c.IsTokenBased())
		assert.False(t, c.IsOTPBased())
	}

	// Test empty
	{
		c := &auth.Confirmation{}
		assert.False(t, c.IsOTPBased())
		assert.False(t, c.IsTokenBased())
	}
}

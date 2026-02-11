package authservice_test

import (
	"context"
	"net/mail"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/entity"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/logger"
	"github.com/fastschema/fastschema/pkg/auth"
	"github.com/fastschema/fastschema/pkg/entdbadapter"
	rr "github.com/fastschema/fastschema/pkg/restfulresolver"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
	as "github.com/fastschema/fastschema/services/auth"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testOTPApp struct {
	db           db.Client
	config       *fs.Config
	authConfig   *fs.AuthConfig
	resources    *fs.ResourcesManager
	resolver     *rr.RestfulResolver
	mailer       *mockMailer
	otpProvider  *auth.OTPProvider
	authProvider fs.AuthProvider
}

type mockMailer struct {
	mu       sync.Mutex
	lastMail *fs.Mail
	sendErr  error
}

func (m *mockMailer) Send(mail *fs.Mail, froms ...mail.Address) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.lastMail = mail
	return m.sendErr
}

func (m *mockMailer) LastMail() *fs.Mail {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.lastMail
}

func (m *mockMailer) Name() string {
	return "mock"
}

func (m *mockMailer) Driver() string {
	return "mock"
}

func (s testOTPApp) JwtCustomClaimsFunc() fs.JwtCustomClaimsFunc {
	return nil
}

func (s testOTPApp) DB() db.Client {
	return s.db
}

func (s testOTPApp) Key() string {
	return "test-secret-key-32-characters!!"
}

func (s testOTPApp) Config() *fs.Config {
	return s.config
}

func (s testOTPApp) Roles() []*fs.Role {
	return []*fs.Role{fs.RoleAdmin, fs.RoleUser, fs.RoleGuest}
}

func (s testOTPApp) GetAuthProvider(name string) fs.AuthProvider {
	if name == auth.ProviderOTP && s.otpProvider != nil {
		return s.otpProvider
	}
	return s.authProvider
}

func (s testOTPApp) Mailer(names ...string) fs.Mailer {
	return s.mailer
}

func createTestOTPApp(t *testing.T, otpEnabled bool) *testOTPApp {
	t.Helper()
	schemaDir := utils.Must(os.MkdirTemp("", "schemas"))
	migrationDir := utils.Must(os.MkdirTemp("", "migrations"))
	sb := utils.Must(schema.NewBuilderFromDir(schemaDir, fs.SystemSchemaTypes...))
	dbc := utils.Must(entdbadapter.NewTestClient(migrationDir, sb))

	// Create default roles
	roleModel := utils.Must(dbc.Model("role"))
	for _, r := range []*fs.Role{fs.RoleAdmin, fs.RoleUser, fs.RoleGuest} {
		utils.Must(roleModel.Create(context.Background(), entity.New().
			Set("name", r.Name).
			Set("root", r.Root),
		))
	}

	authConfig := &fs.AuthConfig{
		EnableRefreshToken: false,
	}

	if otpEnabled {
		authConfig.OTP = &fs.OTPConfig{
			Enabled:     true,
			Length:      6,
			Expiration:  300,
			MaxAttempts: 3,
		}
	}

	mailer := &mockMailer{}

	app := &testOTPApp{
		db:         dbc,
		authConfig: authConfig,
		mailer:     mailer,
		config: &fs.Config{
			AppName:    "TestApp",
			AppKey:     "test-secret-key-32-characters!!",
			AuthConfig: authConfig,
		},
	}

	// Create and initialize OTP provider if enabled
	if otpEnabled {
		otpProvider, _ := auth.NewOTPAuthProvider(nil, "")
		op := otpProvider.(*auth.OTPProvider)
		op.Init(
			func() db.Client { return dbc },
			func() string { return "TestApp" },
			func(names ...string) fs.Mailer { return mailer },
			func() *fs.OTPConfig { return authConfig.OTP },
		)
		app.otpProvider = op
	}

	t.Cleanup(func() {
		assert.NoError(t, os.RemoveAll(schemaDir))
		assert.NoError(t, os.RemoveAll(migrationDir))
		assert.NoError(t, dbc.Close())
	})

	return app
}

// mockOTPContext is a minimal mock implementation of fs.Context for testing
type mockOTPContext struct {
	user   *fs.User
	locals map[string]any
}

func (m *mockOTPContext) TraceID() string                    { return "test-trace" }
func (m *mockOTPContext) User() *fs.User                     { return m.user }
func (m *mockOTPContext) Value(key any) any                  { return nil }
func (m *mockOTPContext) Logger() logger.Logger              { return logger.CreateMockLogger(false) }
func (m *mockOTPContext) AuthToken() string                  { return "" }
func (m *mockOTPContext) Next() error                        { return nil }
func (m *mockOTPContext) Result(...*fs.Result) *fs.Result    { return nil }
func (m *mockOTPContext) Arg(string, ...string) string       { return "" }
func (m *mockOTPContext) ArgInt(string, ...int) int          { return 0 }
func (m *mockOTPContext) Args() map[string]string            { return nil }
func (m *mockOTPContext) SetArg(key, val string) string      { return "" }
func (m *mockOTPContext) Body() ([]byte, error)              { return nil, nil }
func (m *mockOTPContext) Payload() (*entity.Entity, error)   { return nil, nil }
func (m *mockOTPContext) BodyParser(out any) error           { return nil }
func (m *mockOTPContext) Bind(out any) error                 { return nil }
func (m *mockOTPContext) FormValue(string, ...string) string { return "" }
func (m *mockOTPContext) Resource() *fs.Resource             { return nil }
func (m *mockOTPContext) Redirect(string) error              { return nil }
func (m *mockOTPContext) IP() string                         { return "127.0.0.1" }
func (m *mockOTPContext) Header(string, ...string) string    { return "test-user-agent" }
func (m *mockOTPContext) Local(key string, value ...any) any {
	if m.locals == nil {
		m.locals = make(map[string]any)
	}
	if len(value) > 0 {
		m.locals[key] = value[0]
		return value[0]
	}
	return m.locals[key]
}
func (m *mockOTPContext) Files() ([]*fs.File, error)  { return nil, nil }
func (m *mockOTPContext) WSClient() fs.WSClient       { return nil }
func (m *mockOTPContext) Deadline() (time.Time, bool) { return time.Time{}, false }
func (m *mockOTPContext) Done() <-chan struct{}       { return nil }
func (m *mockOTPContext) Err() error                  { return nil }

// Test OTPProvider through OTPProvider directly

func TestOTPProviderRequestNotEnabled(t *testing.T) {
	app := createTestOTPApp(t, false)

	// OTPProvider is nil when not enabled
	assert.Nil(t, app.otpProvider)
}

func TestOTPProviderRequestInvalidEmail(t *testing.T) {
	app := createTestOTPApp(t, true)
	ctx := &mockOTPContext{}

	// Empty email
	_, err := app.otpProvider.RequestOTP(ctx, &auth.OTPRequest{Email: ""})
	assert.Error(t, err)

	// Invalid email format
	_, err = app.otpProvider.RequestOTP(ctx, &auth.OTPRequest{Email: "invalid-email"})
	assert.Error(t, err)
}

func TestOTPProviderRequestUserNotFound(t *testing.T) {
	app := createTestOTPApp(t, true)
	ctx := &mockOTPContext{}

	// User doesn't exist - should still return success (security)
	resp, err := app.otpProvider.RequestOTP(ctx, &auth.OTPRequest{Email: "nonexistent@example.com"})
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 300, resp.ExpiresIn)

	// No email should be sent
	assert.Nil(t, app.mailer.LastMail())
}

func TestOTPProviderRequestSuccess(t *testing.T) {
	app := createTestOTPApp(t, true)

	// Create a test user
	_, err := db.Builder[*fs.User](app.db).Create(context.Background(), entity.New().
		Set("username", "testuser").
		Set("email", "test@example.com").
		Set("active", true).
		Set("provider", "local"),
	)
	require.NoError(t, err)

	ctx := &mockOTPContext{}
	resp, err := app.otpProvider.RequestOTP(ctx, &auth.OTPRequest{Email: "test@example.com"})
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 300, resp.ExpiresIn)
	assert.NotEmpty(t, resp.SessionID) // Session ID should be returned

	// Wait a bit for async email
	time.Sleep(100 * time.Millisecond)

	// Check that email was sent
	assert.NotNil(t, app.mailer.LastMail())
	assert.Equal(t, []string{"test@example.com"}, app.mailer.LastMail().To)
	assert.Contains(t, app.mailer.LastMail().Subject, "TestApp")

	// Check that OTP session was created using session ID
	sessionID := utils.Must(uuid.Parse(resp.SessionID))
	session, err := db.Builder[*fs.Session](app.db).
		Where(db.EQ("id", sessionID)).
		First(context.Background())
	require.NoError(t, err)
	assert.Equal(t, string(fs.SessionStatusPendingOTP), session.Status)
	assert.NotEmpty(t, session.OTPHash)
}

func TestOTPProviderVerifyInvalidInput(t *testing.T) {
	app := createTestOTPApp(t, true)
	ctx := &mockOTPContext{}

	// Empty session_id
	_, err := app.otpProvider.VerifyOTP(ctx, &auth.OTPVerify{SessionID: "", OTP: "123456"})
	assert.Error(t, err)

	// Empty OTP
	sessionID := utils.Must(uuid.NewV7())
	_, err = app.otpProvider.VerifyOTP(ctx, &auth.OTPVerify{SessionID: sessionID.String(), OTP: ""})
	assert.Error(t, err)
}

func TestOTPProviderVerifyNoSession(t *testing.T) {
	app := createTestOTPApp(t, true)
	ctx := &mockOTPContext{}

	// Use a random session ID that doesn't exist
	sessionID := utils.Must(uuid.NewV7())
	_, err := app.otpProvider.VerifyOTP(ctx, &auth.OTPVerify{SessionID: sessionID.String(), OTP: "123456"})
	assert.Error(t, err)
}

func TestOTPProviderVerifyExpired(t *testing.T) {
	app := createTestOTPApp(t, true)

	// Create a test user
	user, err := db.Builder[*fs.User](app.db).Create(context.Background(), entity.New().
		Set("username", "testuser").
		Set("email", "test@example.com").
		Set("active", true).
		Set("provider", "local"),
	)
	require.NoError(t, err)

	// Create an expired OTP session
	expiredTime := time.Now().Add(-1 * time.Hour)
	otpHash, _ := auth.HashOTP("123456")
	sessionID := utils.Must(uuid.NewV7())
	_, err = db.Builder[*fs.Session](app.db).Create(context.Background(), entity.New().
		Set("id", sessionID).
		Set("user_id", user.ID).
		Set("type", string(fs.SessionTypeOTPLogin)).
		Set("status", string(fs.SessionStatusPendingOTP)).
		Set("otp_hash", otpHash).
		Set("otp_attempts", 0).
		Set("expires_at", expiredTime),
	)
	require.NoError(t, err)

	ctx := &mockOTPContext{}
	_, err = app.otpProvider.VerifyOTP(ctx, &auth.OTPVerify{SessionID: sessionID.String(), OTP: "123456"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expired")
}

func TestOTPProviderVerifyMaxAttempts(t *testing.T) {
	app := createTestOTPApp(t, true)

	// Create a test user
	user, err := db.Builder[*fs.User](app.db).Create(context.Background(), entity.New().
		Set("username", "testuser").
		Set("email", "test@example.com").
		Set("active", true).
		Set("provider", "local"),
	)
	require.NoError(t, err)

	// Create OTP session with max attempts reached
	expiresAt := time.Now().Add(5 * time.Minute)
	otpHash, _ := auth.HashOTP("123456")
	sessionID := utils.Must(uuid.NewV7())
	_, err = db.Builder[*fs.Session](app.db).Create(context.Background(), entity.New().
		Set("id", sessionID).
		Set("user_id", user.ID).
		Set("type", string(fs.SessionTypeOTPLogin)).
		Set("status", string(fs.SessionStatusPendingOTP)).
		Set("otp_hash", otpHash).
		Set("otp_attempts", 3). // Max attempts reached
		Set("expires_at", expiresAt),
	)
	require.NoError(t, err)

	ctx := &mockOTPContext{}
	_, err = app.otpProvider.VerifyOTP(ctx, &auth.OTPVerify{SessionID: sessionID.String(), OTP: "123456"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "attempts")
}

func TestOTPProviderVerifyWrongCode(t *testing.T) {
	app := createTestOTPApp(t, true)

	// Create a test user
	user, err := db.Builder[*fs.User](app.db).Create(context.Background(), entity.New().
		Set("username", "testuser").
		Set("email", "test@example.com").
		Set("active", true).
		Set("provider", "local"),
	)
	require.NoError(t, err)

	// Create valid OTP session
	expiresAt := time.Now().Add(5 * time.Minute)
	otpHash, _ := auth.HashOTP("123456")
	sessionID := utils.Must(uuid.NewV7())
	_, err = db.Builder[*fs.Session](app.db).Create(context.Background(), entity.New().
		Set("id", sessionID).
		Set("user_id", user.ID).
		Set("type", string(fs.SessionTypeOTPLogin)).
		Set("status", string(fs.SessionStatusPendingOTP)).
		Set("otp_hash", otpHash).
		Set("otp_attempts", 0).
		Set("expires_at", expiresAt),
	)
	require.NoError(t, err)

	ctx := &mockOTPContext{}
	_, err = app.otpProvider.VerifyOTP(ctx, &auth.OTPVerify{SessionID: sessionID.String(), OTP: "654321"})
	assert.Error(t, err)

	// Check that attempt counter was incremented
	session, err := db.Builder[*fs.Session](app.db).
		Where(db.EQ("id", sessionID)).
		First(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 1, session.OTPAttempts)
}

func TestOTPProviderVerifySuccess(t *testing.T) {
	app := createTestOTPApp(t, true)

	// Create a test user with roles
	user, err := db.Builder[*fs.User](app.db).Create(context.Background(), entity.New().
		Set("username", "testuser").
		Set("email", "test@example.com").
		Set("active", true).
		Set("provider", "google"). // Can be any provider
		Set("roles", []*entity.Entity{entity.New(fs.RoleUser.ID)}),
	)
	require.NoError(t, err)

	// Create valid OTP session
	expiresAt := time.Now().Add(5 * time.Minute)
	otpHash, _ := auth.HashOTP("123456")
	sessionID := utils.Must(uuid.NewV7())
	_, err = db.Builder[*fs.Session](app.db).Create(context.Background(), entity.New().
		Set("id", sessionID).
		Set("user_id", user.ID).
		Set("type", string(fs.SessionTypeOTPLogin)).
		Set("status", string(fs.SessionStatusPendingOTP)).
		Set("otp_hash", otpHash).
		Set("otp_attempts", 0).
		Set("expires_at", expiresAt),
	)
	require.NoError(t, err)

	ctx := &mockOTPContext{}
	verifiedUser, err := app.otpProvider.VerifyOTP(ctx, &auth.OTPVerify{SessionID: sessionID.String(), OTP: "123456"})
	require.NoError(t, err)
	assert.NotNil(t, verifiedUser)
	assert.Equal(t, "test@example.com", verifiedUser.Email)

	// Check that OTP session was deleted
	_, err = db.Builder[*fs.Session](app.db).
		Where(db.EQ("id", sessionID)).
		First(context.Background())
	assert.Error(t, err)
	assert.True(t, db.IsNotFound(err))
}

func TestOTPProviderVerifyInactiveUser(t *testing.T) {
	app := createTestOTPApp(t, true)

	// Create an inactive user
	user, err := db.Builder[*fs.User](app.db).Create(context.Background(), entity.New().
		Set("username", "testuser").
		Set("email", "test@example.com").
		Set("active", false). // Inactive
		Set("provider", "local"),
	)
	require.NoError(t, err)

	// Create valid OTP session
	expiresAt := time.Now().Add(5 * time.Minute)
	otpHash, _ := auth.HashOTP("123456")
	sessionID := utils.Must(uuid.NewV7())
	_, err = db.Builder[*fs.Session](app.db).Create(context.Background(), entity.New().
		Set("id", sessionID).
		Set("user_id", user.ID).
		Set("type", string(fs.SessionTypeOTPLogin)).
		Set("status", string(fs.SessionStatusPendingOTP)).
		Set("otp_hash", otpHash).
		Set("otp_attempts", 0).
		Set("expires_at", expiresAt),
	)
	require.NoError(t, err)

	ctx := &mockOTPContext{}
	_, err = app.otpProvider.VerifyOTP(ctx, &auth.OTPVerify{SessionID: sessionID.String(), OTP: "123456"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "inactive")
}

// Test AuthService wrapper functions

func TestOTPRequestWrapper(t *testing.T) {
	app := createTestOTPApp(t, true)
	authService := as.New(app)

	// Create a test user
	_, err := db.Builder[*fs.User](app.db).Create(context.Background(), entity.New().
		Set("username", "testuser").
		Set("email", "wrapper@example.com").
		Set("active", true).
		Set("provider", "local"),
	)
	require.NoError(t, err)

	ctx := &mockOTPContext{}
	wrapper := authService.OTPRequestWrapper(app.otpProvider)
	resp, err := wrapper(ctx, &auth.OTPRequest{Email: "wrapper@example.com"})
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 300, resp.ExpiresIn)
}

func TestOTPVerifyWrapper(t *testing.T) {
	app := createTestOTPApp(t, true)
	authService := as.New(app)

	// Create a test user with roles
	user, err := db.Builder[*fs.User](app.db).Create(context.Background(), entity.New().
		Set("username", "testuser").
		Set("email", "wrapper@example.com").
		Set("active", true).
		Set("provider", "local").
		Set("roles", []*entity.Entity{entity.New(fs.RoleUser.ID)}),
	)
	require.NoError(t, err)

	// Create valid OTP session
	expiresAt := time.Now().Add(5 * time.Minute)
	otpHash, _ := auth.HashOTP("123456")
	sessionID := utils.Must(uuid.NewV7())
	_, err = db.Builder[*fs.Session](app.db).Create(context.Background(), entity.New().
		Set("id", sessionID).
		Set("user_id", user.ID).
		Set("type", string(fs.SessionTypeOTPLogin)).
		Set("status", string(fs.SessionStatusPendingOTP)).
		Set("otp_hash", otpHash).
		Set("otp_attempts", 0).
		Set("expires_at", expiresAt),
	)
	require.NoError(t, err)

	ctx := &mockOTPContext{}
	wrapper := authService.OTPVerifyWrapper(app.otpProvider)
	tokens, err := wrapper(ctx, &auth.OTPVerify{SessionID: sessionID.String(), OTP: "123456"})
	require.NoError(t, err)
	assert.NotNil(t, tokens)
	assert.NotEmpty(t, tokens.AccessToken)
}

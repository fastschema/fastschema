package account_verification_otp_test

import (
	"context"
	"encoding/json"
	"net/mail"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/entity"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/logger"
	"github.com/fastschema/fastschema/pkg/auth"
	"github.com/fastschema/fastschema/pkg/errors"
	rr "github.com/fastschema/fastschema/pkg/restfulresolver"
	"github.com/fastschema/fastschema/pkg/utils"
	as "github.com/fastschema/fastschema/services/auth"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// MockMailer captures sent emails for testing (thread-safe)
type MockMailer struct {
	mu        sync.RWMutex
	SentMails []*fs.Mail
	SendErr   error
}

func NewMockMailer() *MockMailer {
	return &MockMailer{
		SentMails: make([]*fs.Mail, 0),
	}
}

func (m *MockMailer) Send(mail *fs.Mail, froms ...mail.Address) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.SendErr != nil {
		return m.SendErr
	}
	m.SentMails = append(m.SentMails, mail)
	return nil
}

func (m *MockMailer) Name() string {
	return "mock"
}

func (m *MockMailer) Driver() string {
	return "mock"
}

func (m *MockMailer) LastMail() *fs.Mail {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if len(m.SentMails) == 0 {
		return nil
	}
	return m.SentMails[len(m.SentMails)-1]
}

func (m *MockMailer) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.SentMails = make([]*fs.Mail, 0)
}

func (m *MockMailer) MailCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.SentMails)
}

// testApp holds all test dependencies
type testApp struct {
	db            db.Client
	config        *fs.Config
	resources     *fs.ResourcesManager
	server        *rr.Server
	authService   *as.AuthService
	localProvider *auth.LocalProvider
	mailer        *MockMailer
	inactiveUser  *fs.User
	activeUser    *fs.User
}

func (s testApp) JwtCustomClaimsFunc() fs.JwtCustomClaimsFunc {
	return nil
}

func (s testApp) DB() db.Client {
	return s.db
}

func (s testApp) Key() string {
	return "test-secret-key-32-characters!!"
}

func (s testApp) Config() *fs.Config {
	return s.config
}

func (s testApp) Roles() []*fs.Role {
	return []*fs.Role{fs.RoleAdmin, fs.RoleUser, fs.RoleGuest}
}

func (s testApp) GetAuthProvider(name string) fs.AuthProvider {
	if name == "local" {
		return s.localProvider
	}
	return nil
}

func (s testApp) Mailer(names ...string) fs.Mailer {
	return s.mailer
}

// mockContext for integration tests
type mockContext struct {
	user    *fs.User
	locals  map[string]any
	db      db.Client
	headers map[string]string
	ip      string
}

func (m *mockContext) TraceID() string                    { return "test-trace" }
func (m *mockContext) User() *fs.User                     { return m.user }
func (m *mockContext) Value(key any) any                  { return nil }
func (m *mockContext) Logger() logger.Logger              { return logger.CreateMockLogger(false) }
func (m *mockContext) AuthToken() string                  { return "" }
func (m *mockContext) Next() error                        { return nil }
func (m *mockContext) Result(...*fs.Result) *fs.Result    { return nil }
func (m *mockContext) Arg(string, ...string) string       { return "" }
func (m *mockContext) ArgInt(string, ...int) int          { return 0 }
func (m *mockContext) Args() map[string]string            { return nil }
func (m *mockContext) SetArg(key, val string) string      { return "" }
func (m *mockContext) Body() ([]byte, error)              { return nil, nil }
func (m *mockContext) Payload() (*entity.Entity, error)   { return nil, nil }
func (m *mockContext) BodyParser(out any) error           { return nil }
func (m *mockContext) Bind(out any) error                 { return nil }
func (m *mockContext) FormValue(string, ...string) string { return "" }
func (m *mockContext) Resource() *fs.Resource             { return nil }
func (m *mockContext) Redirect(string) error              { return nil }
func (m *mockContext) IP() string {
	if m.ip != "" {
		return m.ip
	}
	return "127.0.0.1"
}
func (m *mockContext) Header(key string, val ...string) string {
	if m.headers != nil {
		return m.headers[key]
	}
	return ""
}
func (m *mockContext) Local(key string, value ...any) any {
	if m.locals == nil {
		m.locals = make(map[string]any)
	}
	if len(value) > 0 {
		m.locals[key] = value[0]
		return value[0]
	}
	return m.locals[key]
}
func (m *mockContext) Files() ([]*fs.File, error)  { return nil, nil }
func (m *mockContext) WSClient() fs.WSClient       { return nil }
func (m *mockContext) Deadline() (time.Time, bool) { return time.Time{}, false }
func (m *mockContext) Done() <-chan struct{}       { return nil }
func (m *mockContext) Err() error                  { return nil }

var systemSchemas = []any{
	fs.Role{},
	fs.Permission{},
	fs.User{},
	fs.File{},
	fs.Session{},
}

func createTestApp(t *testing.T, dbc db.Client, otpConfig *fs.OTPConfig, verificationMethod string) *testApp {
	t.Helper()
	roleModel := utils.Must(dbc.Model("role"))
	userModel := utils.Must(dbc.Model("user"))

	// Create default roles if they don't exist
	for _, r := range []*fs.Role{fs.RoleAdmin, fs.RoleUser, fs.RoleGuest} {
		existing, _ := roleModel.Query(db.EQ("name", r.Name)).First(context.Background())
		if existing == nil {
			_, _ = roleModel.Create(context.Background(), entity.New().
				Set("name", r.Name).
				Set("root", r.Root))
		}
	}

	// Get the User role ID
	userRole, err := roleModel.Query(db.EQ("name", "User")).First(context.Background())
	require.NoError(t, err)
	userRoleID := userRole.Get("id")

	// Create inactive user for activation tests
	inactiveUsername := "inactiveuser" + utils.RandomString(8)
	inactiveEmail := strings.ToLower("inactive" + utils.RandomString(8) + "@example.com")
	inactiveUserID, err := userModel.Create(context.Background(), entity.New().
		Set("username", inactiveUsername).
		Set("email", inactiveEmail).
		Set("password", "testpassword").
		Set("provider", "local").
		Set("provider_id", utils.RandomString(8)).
		Set("active", false).
		Set("roles", []*entity.Entity{entity.New().Set("id", userRoleID)}))
	require.NoError(t, err)

	// Create active user for recovery tests
	activeUsername := "activeuser" + utils.RandomString(8)
	activeEmail := strings.ToLower("active" + utils.RandomString(8) + "@example.com")
	activeUserID, err := userModel.Create(context.Background(), entity.New().
		Set("username", activeUsername).
		Set("email", activeEmail).
		Set("password", "testpassword").
		Set("provider", "local").
		Set("provider_id", utils.RandomString(8)).
		Set("active", true).
		Set("roles", []*entity.Entity{entity.New().Set("id", userRoleID)}))
	require.NoError(t, err)

	mailer := NewMockMailer()

	authConfig := &fs.AuthConfig{
		EnableRefreshToken:   true,
		AccessTokenLifetime:  60,
		RefreshTokenLifetime: 3600,
		OTP:                  otpConfig,
	}

	app := &testApp{
		db:     dbc,
		mailer: mailer,
		config: &fs.Config{
			AppName:    "TestVerificationApp",
			AppKey:     "test-secret-key-32-characters!!",
			AuthConfig: authConfig,
		},
		inactiveUser: &fs.User{
			ID:       inactiveUserID.(uuid.UUID),
			Username: inactiveUsername,
			Email:    inactiveEmail,
			Active:   false,
			Roles:    []*fs.Role{fs.RoleUser},
		},
		activeUser: &fs.User{
			ID:       activeUserID.(uuid.UUID),
			Username: activeUsername,
			Email:    activeEmail,
			Active:   true,
			Roles:    []*fs.Role{fs.RoleUser},
		},
	}

	// Create and initialize Local provider with OTP verification
	localProvider, _ := auth.NewLocalAuthProvider(fs.Map{
		"activation_method":   "email",
		"verification_method": verificationMethod,
	}, "http://localhost:8080/callback")
	lp := localProvider.(*auth.LocalProvider)
	lp.Init(
		func() db.Client { return dbc },
		func() string { return "test-secret-key-32-characters!!" },
		func() string { return "TestVerificationApp" },
		func() string { return "http://localhost:8080" },
		func(names ...string) fs.Mailer { return mailer },
		nil,
		func() *fs.OTPConfig { return otpConfig },
	)
	app.localProvider = lp

	app.authService = as.New(app)
	app.resources = fs.NewResourcesManager()
	app.resources.Hooks = func() *fs.Hooks {
		return &fs.Hooks{
			PreResolve: []fs.Middleware{app.authService.Authorize},
		}
	}
	app.resources.Middlewares = append(app.resources.Middlewares, app.authService.ParseUser)

	api := app.resources.Group("api", &fs.Meta{Prefix: "/api"})

	// Auth endpoints
	authGroup := api.Group("auth")
	authGroup.Add(fs.Get("me", app.authService.Me, &fs.Meta{Public: true}))
	authGroup.Add(fs.Post("logout", app.authService.Logout, &fs.Meta{Public: true}))
	authGroup.Add(fs.Post("token/refresh", app.authService.RefreshToken, &fs.Meta{Public: true}))

	// Local auth endpoints
	localGroup := authGroup.Group("local")
	localGroup.Add(
		fs.Post("activate", lp.Activate, &fs.Meta{Public: true}),
		fs.Post("activate/send", lp.SendActivationLink, &fs.Meta{Public: true}),
		fs.Post("recover", lp.Recover, &fs.Meta{Public: true}),
		fs.Post("recover/check", lp.RecoverCheck, &fs.Meta{Public: true}),
		fs.Post("recover/reset", lp.ResetPassword, &fs.Meta{Public: true}),
	)

	// Protected resource for testing
	api.Add(fs.Get("protected", func(c fs.Context, _ any) (any, error) {
		user := c.User()
		if user == nil {
			return nil, errors.Unauthorized("unauthorized")
		}
		return map[string]any{"message": "protected resource", "user_id": user.ID}, nil
	}, &fs.Meta{Public: true}))

	require.NoError(t, app.resources.Init())
	resolver := rr.NewRestfulResolver(&rr.ResolverConfig{
		ResourceManager: app.resources,
		Logger:          logger.CreateMockLogger(false),
	})
	app.server = resolver.Server()

	return app
}

// Helper to parse JSON response
type apiResponse struct {
	Data  json.RawMessage `json:"data"`
	Error string          `json:"error"`
}

func parseResponse(t *testing.T, body []byte) apiResponse {
	t.Helper()
	var resp apiResponse
	require.NoError(t, json.Unmarshal(body, &resp))
	return resp
}

// clearSessions removes all activation and recovery sessions from the database
func clearSessions(dbc db.Client) {
	_, _ = db.Builder[*fs.Session](dbc).
		Where(db.Or(
			db.EQ("type", string(fs.SessionTypeActivation)),
			db.EQ("type", string(fs.SessionTypeRecovery)),
		)).
		Delete(context.Background())
}

package refreshtoken_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/entity"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/logger"
	"github.com/fastschema/fastschema/pkg/errors"
	rr "github.com/fastschema/fastschema/pkg/restfulresolver"
	"github.com/fastschema/fastschema/pkg/utils"
	as "github.com/fastschema/fastschema/services/auth"
	"github.com/stretchr/testify/require"
)

type testApp struct {
	db          db.Client
	config      *fs.Config
	resources   *fs.ResourcesManager
	server      *rr.Server
	authService *as.AuthService
	testUser    *fs.User
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
	return nil
}

// mockContext for integration tests
type mockContext struct {
	user   *fs.User
	locals map[string]any
	db     db.Client
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
func (m *mockContext) IP() string                         { return "127.0.0.1" }
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
	fs.Token{},
}

func createTestApp(t *testing.T, dbc db.Client) *testApp {
	t.Helper()
	roleModel := utils.Must(dbc.Model("role"))
	userModel := utils.Must(dbc.Model("user"))

	for _, r := range []*fs.Role{fs.RoleAdmin, fs.RoleUser, fs.RoleGuest} {
		_, _ = roleModel.Create(context.Background(), entity.New().
			Set("name", r.Name).
			Set("root", r.Root))
	}

	// Create test user with unique credentials
	username := "testuser" + utils.RandomString(8)
	email := "test" + utils.RandomString(8) + "@example.com"
	userID, err := userModel.Create(context.Background(), entity.New().
		Set("username", username).
		Set("email", email).
		Set("password", "testpassword").
		Set("provider", "local").
		Set("provider_id", utils.RandomString(8)).
		Set("active", true).
		Set("roles", []*entity.Entity{entity.New(2)}))
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	authConfig := &fs.AuthConfig{
		EnableRefreshToken:   true,
		AccessTokenLifetime:  60,   // 1 minute for testing
		RefreshTokenLifetime: 3600, // 1 hour for testing
	}

	app := &testApp{
		db: dbc,
		config: &fs.Config{
			AppKey:     "test-secret-key-32-characters!!",
			AuthConfig: authConfig,
		},
		testUser: &fs.User{
			ID:       userID,
			Username: username,
			Email:    email,
			Active:   true,
			Roles:    []*fs.Role{fs.RoleUser},
		},
	}

	app.authService = as.New(app)
	app.resources = fs.NewResourcesManager()
	app.resources.Hooks = func() *fs.Hooks {
		return &fs.Hooks{
			PreResolve: []fs.Middleware{app.authService.Authorize},
		}
	}
	app.resources.Middlewares = append(app.resources.Middlewares, app.authService.ParseUser)

	api := app.resources.Group("api", &fs.Meta{Prefix: "/api"})
	api.Group("auth").
		Add(fs.Get("me", app.authService.Me, &fs.Meta{Public: true})).
		Add(fs.Post("logout", app.authService.Logout, &fs.Meta{Public: true})).
		Add(fs.Post("logout/all", app.authService.LogoutAll, &fs.Meta{Public: true})).
		Add(fs.Post("token/refresh", app.authService.RefreshToken, &fs.Meta{Public: true}))

	// Add a protected resource for testing (public but manually checks user)
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

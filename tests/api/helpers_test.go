package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"testing"
	"time"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/entity"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/logger"
	"github.com/fastschema/fastschema/pkg/auth"
	"github.com/fastschema/fastschema/pkg/entdbadapter"
	"github.com/fastschema/fastschema/pkg/jwt"
	rr "github.com/fastschema/fastschema/pkg/restfulresolver"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
	as "github.com/fastschema/fastschema/services/auth"
	cs "github.com/fastschema/fastschema/services/content"
	rs "github.com/fastschema/fastschema/services/role"
	ss "github.com/fastschema/fastschema/services/schema"
	ts "github.com/fastschema/fastschema/services/tool"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// TestApp holds all components for API testing
type TestApp struct {
	t             *testing.T
	db            db.Client
	schemaBuilder *schema.Builder
	schemaDir     string
	resources     *fs.ResourcesManager
	server        *rr.Server
	authService   *as.AuthService

	adminUser     *fs.User
	normalUser    *fs.User
	guestUser     *fs.User
	adminToken    string
	normalToken   string
	guestToken    string
	adminRoleID   uuid.UUID
	userRoleID    uuid.UUID
	guestRoleID   uuid.UUID
}

// AppLike interface implementations for all services

// Common interfaces
func (a *TestApp) DB() db.Client { return a.db }
func (a *TestApp) Key() string   { return "test-api-secret-key-32-chars!!" }
func (a *TestApp) Config() *fs.Config {
	return &fs.Config{
		AppKey: "test-api-secret-key-32-chars!!",
		AuthConfig: &fs.AuthConfig{
			EnableRefreshToken:   true,
			AccessTokenLifetime:  3600,
			RefreshTokenLifetime: 86400,
		},
	}
}

// AuthService AppLike interface
func (a *TestApp) Roles() []*fs.Role {
	roles, _ := db.Builder[*fs.Role](a.db).Get(context.Background())
	for _, role := range roles {
		_ = role.Compile()
		for _, p := range role.Permissions {
			_ = p.Compile()
		}
	}
	return roles
}
func (a *TestApp) GetAuthProvider(name string) fs.AuthProvider {
	if name == auth.ProviderLocal {
		p, _ := auth.NewLocalAuthProvider(fs.Map{}, "")
		lp := p.(*auth.LocalProvider)
		lp.Init(
			func() db.Client { return a.db },
			func() string { return a.Key() },
			func() string { return "TestApp" },
			func() string { return "http://localhost:8080" },
			func(names ...string) fs.Mailer { return nil },
			nil,
			func() *fs.OTPConfig { return nil },
			nil,
		)
		return lp
	}
	return nil
}
func (a *TestApp) JwtCustomClaimsFunc() fs.JwtCustomClaimsFunc { return nil }
func (a *TestApp) Mailer(names ...string) fs.Mailer           { return nil }

// SchemaService AppLike interface
func (a *TestApp) SchemaBuilder() *schema.Builder { return a.schemaBuilder }
func (a *TestApp) Disk(names ...string) fs.Disk   { return nil }
func (a *TestApp) SystemSchemas() []any           { return fs.SystemSchemaTypes }
func (a *TestApp) Reload(ctx context.Context, changes *db.Changes) error {
	// Simplified reload for testing - just return nil
	return nil
}

// RoleService AppLike interface
func (a *TestApp) UpdateCache(ctx context.Context) error {
	// No-op for tests
	return nil
}

// APIResponse represents a generic API response
type APIResponse struct {
	Data  json.RawMessage `json:"data"`
	Error *APIError       `json:"error,omitempty"`
}

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// PaginatedResponse represents a paginated API response
type PaginatedResponse struct {
	Total       uint              `json:"total"`
	PerPage     uint              `json:"per_page"`
	CurrentPage uint              `json:"current_page"`
	LastPage    uint              `json:"last_page"`
	Items       []json.RawMessage `json:"items"`
}

// CreateTestApp creates a fully configured test app with all services
func CreateTestApp(t *testing.T) *TestApp {
	t.Helper()

	schemaDir := t.TempDir()
	migrationDir := t.TempDir()

	// Create test schemas - simple schemas without user relation issues
	utils.WriteFile(schemaDir+"/post.json", `{
		"name": "post",
		"namespace": "posts",
		"label_field": "title",
		"fields": [
			{"type": "uint64", "name": "id", "label": "ID", "db": {"attr": "UNSIGNED", "key": "PRIMARY", "increment": true}},
			{"type": "string", "name": "title", "label": "Title", "sortable": true},
			{"type": "text", "name": "content", "label": "Content"},
			{"type": "bool", "name": "published", "label": "Published", "optional": true},
			{"type": "relation", "name": "categories", "label": "Categories", "optional": true, "relation": {"schema": "category", "field": "posts", "type": "m2m"}}
		]
	}`)

	utils.WriteFile(schemaDir+"/category.json", `{
		"name": "category",
		"namespace": "categories",
		"label_field": "name",
		"fields": [
			{"type": "uint64", "name": "id", "label": "ID", "db": {"attr": "UNSIGNED", "key": "PRIMARY", "increment": true}},
			{"type": "string", "name": "name", "label": "Name", "unique": true, "sortable": true},
			{"type": "text", "name": "description", "label": "Description", "optional": true},
			{"type": "relation", "name": "posts", "label": "Posts", "optional": true, "relation": {"schema": "post", "field": "categories", "type": "m2m", "owner": true}}
		]
	}`)

	sb, err := schema.NewBuilderFromDir(schemaDir, fs.SystemSchemaTypes...)
	require.NoError(t, err)

	dbc, err := entdbadapter.NewEntClient(&db.Config{
		Driver:       "sqlite",
		Name:         path.Join(t.TempDir(), "test_api.db"),
		MigrationDir: migrationDir,
		LogQueries:   false,
	}, sb)
	require.NoError(t, err)

	t.Cleanup(func() { _ = dbc.Close() })

	app := &TestApp{
		t:             t,
		db:            dbc,
		schemaBuilder: sb,
		schemaDir:     schemaDir,
	}

	// Create roles
	roleModel, _ := dbc.Model("role")
	permModel, _ := dbc.Model("permission")

	adminRoleID, _ := roleModel.Create(context.Background(), entity.New().
		Set("name", fs.RoleAdmin.Name).
		Set("root", true).
		Set("system", true))
	app.adminRoleID = adminRoleID.(uuid.UUID)

	userRoleID, _ := roleModel.Create(context.Background(), entity.New().
		Set("name", fs.RoleUser.Name).
		Set("root", false).
		Set("system", true))
	app.userRoleID = userRoleID.(uuid.UUID)

	guestRoleID, _ := roleModel.Create(context.Background(), entity.New().
		Set("name", fs.RoleGuest.Name).
		Set("root", false).
		Set("system", true))
	app.guestRoleID = guestRoleID.(uuid.UUID)

	// Update global role IDs
	fs.RoleAdmin.ID = app.adminRoleID
	fs.RoleUser.ID = app.userRoleID
	fs.RoleGuest.ID = app.guestRoleID

	// Create permissions for user role
	permModel.Create(context.Background(), entity.New().
		Set("resource", "api.content.*").
		Set("value", fs.PermissionTypeAllow.String()).
		Set("role_id", app.userRoleID))

	// Create users
	userModel, _ := dbc.Model("user")

	adminUserID, _ := userModel.Create(context.Background(), entity.New().
		Set("username", "admin").
		Set("email", "admin@test.local").
		Set("password", "admin123").
		Set("provider", "local").
		Set("active", true).
		Set("roles", []*entity.Entity{entity.New(app.adminRoleID)}))

	normalUserID, _ := userModel.Create(context.Background(), entity.New().
		Set("username", "user").
		Set("email", "user@test.local").
		Set("password", "user123").
		Set("provider", "local").
		Set("active", true).
		Set("roles", []*entity.Entity{entity.New(app.userRoleID)}))

	app.adminUser = &fs.User{
		ID:       adminUserID.(uuid.UUID),
		Username: "admin",
		Email:    "admin@test.local",
		Active:   true,
		Roles:    []*fs.Role{fs.RoleAdmin},
		RoleIDs:  []uuid.UUID{app.adminRoleID},
	}

	app.normalUser = &fs.User{
		ID:       normalUserID.(uuid.UUID),
		Username: "user",
		Email:    "user@test.local",
		Active:   true,
		Roles:    []*fs.Role{fs.RoleUser},
		RoleIDs:  []uuid.UUID{app.userRoleID},
	}

	app.guestUser = &fs.User{
		ID:       uuid.Nil,
		Username: "guest",
		Active:   true,
		Roles:    []*fs.Role{fs.RoleGuest},
		RoleIDs:  []uuid.UUID{app.guestRoleID},
	}

	// Generate tokens
	key := app.Key()
	app.adminToken, _, _ = jwt.GenerateAccessToken(jwt.UserToJwtClaims(app.adminUser), key, time.Time{}, nil)
	app.normalToken, _, _ = jwt.GenerateAccessToken(jwt.UserToJwtClaims(app.normalUser), key, time.Time{}, nil)

	// Create services
	app.authService = as.New(app)
	contentService := cs.New(app)
	schemaService := ss.New(app)
	roleService := rs.New(app)
	toolService := ts.New(app)

	// Create resources
	app.resources = fs.NewResourcesManager()
	app.resources.Hooks = func() *fs.Hooks {
		return &fs.Hooks{
			PreResolve: []fs.Middleware{app.authService.Authorize},
		}
	}
	app.resources.Middlewares = append(app.resources.Middlewares, app.authService.ParseUser)

	api := app.resources.Group("api", &fs.Meta{Prefix: "/api"})

	// Auth endpoints
	authGroup := api.Group("user")
	app.authService.CreateResource(authGroup, map[string]fs.AuthProvider{
		auth.ProviderLocal: app.GetAuthProvider(auth.ProviderLocal),
	})

	// Schema endpoints
	schemaService.CreateResource(api)

	// Content endpoints
	contentService.CreateResource(api)

	// Role endpoints
	roleService.CreateResource(api)

	// Tool endpoints (stats/recent/activity)
	toolService.CreateResource(api)

	require.NoError(t, app.resources.Init())

	resolver := rr.NewRestfulResolver(&rr.ResolverConfig{
		ResourceManager: app.resources,
		Logger:          logger.CreateMockLogger(false),
	})
	app.server = resolver.Server()

	return app
}

// Request helpers

func (a *TestApp) Get(urlPath string, token ...string) (*httptest.ResponseRecorder, *APIResponse) {
	req := httptest.NewRequest("GET", urlPath, nil)
	if len(token) > 0 && token[0] != "" {
		req.Header.Set("Authorization", "Bearer "+token[0])
	}
	return a.doRequest(req)
}

func (a *TestApp) Post(urlPath string, body any, token ...string) (*httptest.ResponseRecorder, *APIResponse) {
	var reader io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		reader = bytes.NewReader(b)
	}
	req := httptest.NewRequest("POST", urlPath, reader)
	req.Header.Set("Content-Type", "application/json")
	if len(token) > 0 && token[0] != "" {
		req.Header.Set("Authorization", "Bearer "+token[0])
	}
	return a.doRequest(req)
}

func (a *TestApp) Put(urlPath string, body any, token ...string) (*httptest.ResponseRecorder, *APIResponse) {
	var reader io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		reader = bytes.NewReader(b)
	}
	req := httptest.NewRequest("PUT", urlPath, reader)
	req.Header.Set("Content-Type", "application/json")
	if len(token) > 0 && token[0] != "" {
		req.Header.Set("Authorization", "Bearer "+token[0])
	}
	return a.doRequest(req)
}

func (a *TestApp) Delete(urlPath string, token ...string) (*httptest.ResponseRecorder, *APIResponse) {
	req := httptest.NewRequest("DELETE", urlPath, nil)
	if len(token) > 0 && token[0] != "" {
		req.Header.Set("Authorization", "Bearer "+token[0])
	}
	return a.doRequest(req)
}

func (a *TestApp) doRequest(req *http.Request) (*httptest.ResponseRecorder, *APIResponse) {
	resp, err := a.server.Test(req)
	require.NoError(a.t, err)

	body, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()

	var apiResp APIResponse
	_ = json.Unmarshal(body, &apiResp)

	return &httptest.ResponseRecorder{Code: resp.StatusCode}, &apiResp
}

// Assertion helpers

func (a *TestApp) AssertStatus(resp *httptest.ResponseRecorder, expected int) {
	a.t.Helper()
	require.Equal(a.t, expected, resp.Code, "unexpected status code")
}

func (a *TestApp) ParseData(resp *APIResponse, v any) {
	a.t.Helper()
	require.NoError(a.t, json.Unmarshal(resp.Data, v))
}

func (a *TestApp) ParsePaginated(resp *APIResponse) *PaginatedResponse {
	a.t.Helper()
	var p PaginatedResponse
	require.NoError(a.t, json.Unmarshal(resp.Data, &p))
	return &p
}

// Data helpers

func (a *TestApp) CreatePost(title, content string, published bool) uint64 {
	a.t.Helper()
	postModel, _ := a.db.Model("post")
	id, err := postModel.Create(context.Background(), entity.New().
		Set("title", title).
		Set("content", content).
		Set("published", published))
	require.NoError(a.t, err)
	return id.(uint64)
}

func (a *TestApp) CreateCategory(name, description string) uint64 {
	a.t.Helper()
	catModel, _ := a.db.Model("category")
	id, err := catModel.Create(context.Background(), entity.New().
		Set("name", name).
		Set("description", description))
	require.NoError(a.t, err)
	return id.(uint64)
}

// Ensure temporary directories are cleaned up
func init() {
	os.Setenv("FASTSCHEMA_TEST", "true")
}

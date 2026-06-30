package cli_login_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
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
	"github.com/fastschema/fastschema/schema"
	as "github.com/fastschema/fastschema/services/auth"
	cs "github.com/fastschema/fastschema/services/content"
	rs "github.com/fastschema/fastschema/services/role"
	ss "github.com/fastschema/fastschema/services/schema"
	ts "github.com/fastschema/fastschema/services/tool"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	// validTestKey is a 32-byte AES-256 key for the auth carrier signing
	validTestKey = "0123456789abcdef0123456789abcdef"
	// testUsername is the seeded local user login
	testUsername = "admin"
	// testPassword is the seeded local user password
	testPassword = "123"
)

// CLITestApp holds all components for CLI integration testing
type CLITestApp struct {
	t             *testing.T
	db            db.Client
	schemaBuilder *schema.Builder
	schemaDir     string
	resources     *fs.ResourcesManager
	server        *rr.Server
	authService   *as.AuthService

	testUser       *fs.User
	testUserID     uuid.UUID
	testToken      string
	cliLoginConfig *fs.CLILoginConfig
}

// Common interfaces
func (a *CLITestApp) DB() db.Client { return a.db }
func (a *CLITestApp) Key() string   { return validTestKey }
func (a *CLITestApp) Config() *fs.Config {
	return &fs.Config{
		AppKey: validTestKey,
		AuthConfig: &fs.AuthConfig{
			EnableRefreshToken:   false,
			AccessTokenLifetime:  3600,
			RefreshTokenLifetime: 86400,
			// Use the configured CLILogin (can be nil to test disabled feature)
			CLILogin: a.cliLoginConfig,
		},
	}
}

// AuthService AppLike interface
func (a *CLITestApp) Roles() []*fs.Role {
	roles, _ := db.Builder[*fs.Role](a.db).Get(context.Background())
	for _, role := range roles {
		_ = role.Compile()
		for _, p := range role.Permissions {
			_ = p.Compile()
		}
	}
	return roles
}

func (a *CLITestApp) GetAuthProvider(name string) fs.AuthProvider {
	if name == auth.ProviderLocal {
		p, _ := auth.NewLocalAuthProvider(fs.Map{}, "")
		lp := p.(*auth.LocalProvider)
		lp.Init(
			func() db.Client { return a.db },
			func() string { return a.Key() },
			func() string { return "CLITestApp" },
			func() string { return "http://localhost:8080" },
			func(names ...string) fs.Mailer { return nil },
			nil,
			func() *fs.OTPConfig { return nil },
			nil,
			nil,
			nil,
		)
		return lp
	}
	return nil
}

func (a *CLITestApp) JwtCustomClaimsFunc() fs.JwtCustomClaimsFunc { return nil }
func (a *CLITestApp) Mailer(names ...string) fs.Mailer            { return nil }

// SchemaService AppLike interface
func (a *CLITestApp) SchemaBuilder() *schema.Builder { return a.schemaBuilder }
func (a *CLITestApp) Disk(names ...string) fs.Disk   { return nil }
func (a *CLITestApp) SystemSchemas() []any           { return fs.SystemSchemaTypes }
func (a *CLITestApp) Reload(ctx context.Context, changes *db.Changes) error {
	return nil
}

// RoleService AppLike interface
func (a *CLITestApp) UpdateCache(ctx context.Context) error {
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

// CreateCLITestApp creates a fully configured test app with CLI login enabled
func CreateCLITestApp(t *testing.T) *CLITestApp {
	t.Helper()

	schemaDir := t.TempDir()
	migrationDir := t.TempDir()

	sb, err := schema.NewBuilderFromDir(schemaDir, fs.SystemSchemaTypes...)
	require.NoError(t, err)

	dbc, err := entdbadapter.NewEntClient(&db.Config{
		Driver:       "sqlite",
		Name:         path.Join(t.TempDir(), "test_cli.db"),
		MigrationDir: migrationDir,
		LogQueries:   false,
	}, sb)
	require.NoError(t, err)

	t.Cleanup(func() { _ = dbc.Close() })

	app := &CLITestApp{
		t:             t,
		db:            dbc,
		schemaBuilder: sb,
		schemaDir:     schemaDir,
		cliLoginConfig: &fs.CLILoginConfig{
			Enabled:              true,
			AllowedRedirectHosts: []string{"app.example.com"},
		},
	}

	// Create roles
	roleModel, _ := dbc.Model("role")

	adminRoleID, _ := roleModel.Create(context.Background(), entity.New().
		Set("name", fs.RoleAdmin.Name).
		Set("root", true).
		Set("system", true))

	userRoleID, _ := roleModel.Create(context.Background(), entity.New().
		Set("name", fs.RoleUser.Name).
		Set("root", false).
		Set("system", true))

	guestRoleID, _ := roleModel.Create(context.Background(), entity.New().
		Set("name", fs.RoleGuest.Name).
		Set("root", false).
		Set("system", true))

	// Update global role IDs
	fs.RoleAdmin.ID = adminRoleID.(uuid.UUID)
	fs.RoleUser.ID = userRoleID.(uuid.UUID)
	fs.RoleGuest.ID = guestRoleID.(uuid.UUID)

	// Create test user (admin/123)
	userModel, _ := dbc.Model("user")

	testUserID, _ := userModel.Create(context.Background(), entity.New().
		Set("username", testUsername).
		Set("email", "admin@test.local").
		Set("password", testPassword).
		Set("provider", "local").
		Set("active", true).
		Set("roles", []*entity.Entity{entity.New(adminRoleID)}))

	app.testUserID = testUserID.(uuid.UUID)
	app.testUser = &fs.User{
		ID:       app.testUserID,
		Username: testUsername,
		Email:    "admin@test.local",
		Active:   true,
		Roles:    []*fs.Role{fs.RoleAdmin},
		RoleIDs:  []uuid.UUID{adminRoleID.(uuid.UUID)},
	}

	// Generate token
	key := app.Key()
	app.testToken, _, _ = jwt.GenerateAccessToken(jwt.UserToJwtClaims(app.testUser), key, time.Time{}, nil)

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

	// Auth endpoints (note: CreateResource adds "auth" group internally)
	userGroup := api.Group("user")
	app.authService.CreateResource(userGroup, map[string]fs.AuthProvider{
		auth.ProviderLocal: app.GetAuthProvider(auth.ProviderLocal),
	})

	// Schema endpoints
	schemaService.CreateResource(api)

	// Content endpoints
	contentService.CreateResource(api)

	// Role endpoints
	roleService.CreateResource(api)

	// Tool endpoints
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

func (a *CLITestApp) Post(urlPath string, body any, token ...string) (*httptest.ResponseRecorder, *APIResponse) {
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

func (a *CLITestApp) Get(urlPath string, token ...string) (*httptest.ResponseRecorder, *APIResponse) {
	req := httptest.NewRequest("GET", urlPath, nil)
	if len(token) > 0 && token[0] != "" {
		req.Header.Set("Authorization", "Bearer "+token[0])
	}
	return a.doRequest(req)
}

func (a *CLITestApp) doRequest(req *http.Request) (*httptest.ResponseRecorder, *APIResponse) {
	resp, err := a.server.Test(req)
	require.NoError(a.t, err)

	body, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()

	var apiResp APIResponse
	_ = json.Unmarshal(body, &apiResp)

	return &httptest.ResponseRecorder{Code: resp.StatusCode}, &apiResp
}

// Helper to generate PKCE challenge from verifier
func generatePKCEChallenge(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

// ===== Test Cases =====

// TestCLILoginHappyPath simulates the full CLI login flow with PKCE
func TestCLILoginHappyPath(t *testing.T) {
	app := CreateCLITestApp(t)

	// Step 1: Initiate CLI login request
	codeVerifier := "this-is-a-test-code-verifier-string-with-sufficient-length"
	codeChallenge := generatePKCEChallenge(codeVerifier)

	initReq := map[string]string{
		"redirect_uri":   "http://127.0.0.1:54321/callback",
		"correlation":    "corr-1",
		"code_challenge": codeChallenge,
	}
	statusResp, apiResp := app.Post("/api/user/auth/cli/initiate", initReq)
	require.Equal(t, http.StatusOK, statusResp.Code, "initiate should succeed")

	// Parse initiate response
	var initResp struct {
		Carrier      string `json:"carrier"`
		AuthorizeURL string `json:"authorize_url"`
	}
	require.NoError(t, json.Unmarshal(apiResp.Data, &initResp))
	assert.NotEmpty(t, initResp.Carrier, "carrier should not be empty")
	assert.True(t, bytes.Contains([]byte(initResp.AuthorizeURL), []byte("/dash/login?cli=")), "authorize_url should start with /dash/login?cli=")

	// Step 2: Local login with credentials
	localLoginReq := map[string]string{
		"login":    testUsername,
		"password": testPassword,
		"carrier":  initResp.Carrier,
	}
	statusResp, apiResp = app.Post("/api/user/auth/cli/login/local", localLoginReq)
	require.Equal(t, http.StatusOK, statusResp.Code, "local login should succeed")

	// Parse local login response
	var localLoginResp struct {
		Redirect string `json:"redirect"`
	}
	require.NoError(t, json.Unmarshal(apiResp.Data, &localLoginResp))
	assert.NotEmpty(t, localLoginResp.Redirect, "redirect should not be empty")

	// Verify redirect starts with loopback and contains code and state
	parsed, err := url.Parse(localLoginResp.Redirect)
	require.NoError(t, err)
	assert.Equal(t, "127.0.0.1:54321", parsed.Host, "redirect should be to loopback")
	assert.NotEmpty(t, parsed.Query().Get("code"), "redirect should contain code")
	assert.Equal(t, "corr-1", parsed.Query().Get("state"), "redirect should contain state/correlation")
	// Verify JWT is NOT in the redirect (it's kept server-side)
	assert.NotContains(t, localLoginResp.Redirect, "eyJ", "redirect should not contain JWT")

	// Step 3: Exchange code for tokens
	code := parsed.Query().Get("code")
	exchangeReq := map[string]string{
		"code":          code,
		"code_verifier": codeVerifier,
	}
	statusResp, apiResp = app.Post("/api/user/auth/exchange", exchangeReq)
	require.Equal(t, http.StatusOK, statusResp.Code, "exchange should succeed")

	// Parse exchange response
	var exchangeResp struct {
		Token   string    `json:"token"`
		Expires time.Time `json:"expires"`
	}
	require.NoError(t, json.Unmarshal(apiResp.Data, &exchangeResp))
	assert.NotEmpty(t, exchangeResp.Token, "token should not be empty")
	assert.Greater(t, exchangeResp.Expires, time.Now(), "expires should be in the future")

	// Step 4: Verify token by calling GET /api/user/auth/me
	statusResp, apiResp = app.Get("/api/user/auth/me", exchangeResp.Token)
	require.Equal(t, http.StatusOK, statusResp.Code, "get me should succeed")

	var meResp struct {
		ID       string `json:"id"`
		Username string `json:"username"`
		Email    string `json:"email"`
	}
	require.NoError(t, json.Unmarshal(apiResp.Data, &meResp))
	assert.Equal(t, testUsername, meResp.Username, "username should match")
	assert.Equal(t, "admin@test.local", meResp.Email, "email should match")
}

// TestCLILoginSingleUseCode tests that a code can only be used once
func TestCLILoginSingleUseCode(t *testing.T) {
	app := CreateCLITestApp(t)

	codeVerifier := "test-verifier-for-single-use"
	codeChallenge := generatePKCEChallenge(codeVerifier)

	// Initiate
	initReq := map[string]string{
		"redirect_uri":   "http://127.0.0.1:54321/callback",
		"code_challenge": codeChallenge,
	}
	statusResp, apiResp := app.Post("/api/user/auth/cli/initiate", initReq)
	require.Equal(t, http.StatusOK, statusResp.Code)

	var initResp struct {
		Carrier string `json:"carrier"`
	}
	json.Unmarshal(apiResp.Data, &initResp)

	// Local login
	localLoginReq := map[string]string{
		"login":    testUsername,
		"password": testPassword,
		"carrier":  initResp.Carrier,
	}
	statusResp, apiResp = app.Post("/api/user/auth/cli/login/local", localLoginReq)
	require.Equal(t, http.StatusOK, statusResp.Code)

	var localLoginResp struct {
		Redirect string `json:"redirect"`
	}
	json.Unmarshal(apiResp.Data, &localLoginResp)

	parsed, _ := url.Parse(localLoginResp.Redirect)
	code := parsed.Query().Get("code")

	// First exchange succeeds
	exchangeReq := map[string]string{
		"code":          code,
		"code_verifier": codeVerifier,
	}
	statusResp, _ = app.Post("/api/user/auth/exchange", exchangeReq)
	require.Equal(t, http.StatusOK, statusResp.Code, "first exchange should succeed")

	// Second exchange with same code fails (single-use)
	statusResp, apiResp = app.Post("/api/user/auth/exchange", exchangeReq)
	require.Equal(t, http.StatusUnauthorized, statusResp.Code, "second exchange should fail (single-use)")
	assert.NotEmpty(t, apiResp.Error, "should have error response")
}

// TestCLILoginPKCEMismatch tests PKCE verification
func TestCLILoginPKCEMismatch(t *testing.T) {
	app := CreateCLITestApp(t)

	codeVerifier := "test-verifier-pkce-mismatch"
	codeChallenge := generatePKCEChallenge(codeVerifier)

	// Initiate
	initReq := map[string]string{
		"redirect_uri":   "http://127.0.0.1:54321/callback",
		"code_challenge": codeChallenge,
	}
	statusResp, apiResp := app.Post("/api/user/auth/cli/initiate", initReq)
	require.Equal(t, http.StatusOK, statusResp.Code)

	var initResp struct {
		Carrier string `json:"carrier"`
	}
	json.Unmarshal(apiResp.Data, &initResp)

	// Local login
	localLoginReq := map[string]string{
		"login":    testUsername,
		"password": testPassword,
		"carrier":  initResp.Carrier,
	}
	statusResp, apiResp = app.Post("/api/user/auth/cli/login/local", localLoginReq)
	require.Equal(t, http.StatusOK, statusResp.Code)

	var localLoginResp struct {
		Redirect string `json:"redirect"`
	}
	json.Unmarshal(apiResp.Data, &localLoginResp)

	parsed, _ := url.Parse(localLoginResp.Redirect)
	code := parsed.Query().Get("code")

	// Exchange with WRONG verifier fails (PKCE mismatch)
	wrongVerifier := "this-is-the-wrong-verifier"
	exchangeReq := map[string]string{
		"code":          code,
		"code_verifier": wrongVerifier,
	}
	statusResp, _ = app.Post("/api/user/auth/exchange", exchangeReq)
	require.Equal(t, http.StatusUnauthorized, statusResp.Code, "exchange with wrong verifier should fail")

	// Second attempt with correct verifier also fails (code already consumed)
	exchangeReq["code_verifier"] = codeVerifier
	statusResp, _ = app.Post("/api/user/auth/exchange", exchangeReq)
	require.Equal(t, http.StatusUnauthorized, statusResp.Code, "code already consumed by failed attempt")
}

// TestCLILoginBadRedirect tests validation of redirect URI
func TestCLILoginBadRedirect(t *testing.T) {
	app := CreateCLITestApp(t)

	// Bad redirect: foreign host not allowlisted
	initReq := map[string]string{
		"redirect_uri":   "https://evil.com/callback",
		"code_challenge": "some-challenge",
	}
	statusResp, _ := app.Post("/api/user/auth/cli/initiate", initReq)
	require.Equal(t, http.StatusBadRequest, statusResp.Code, "foreign host should be rejected")

	// Non-loopback host not in allowlist
	initReq["redirect_uri"] = "http://example.com/callback"
	statusResp, _ = app.Post("/api/user/auth/cli/initiate", initReq)
	require.Equal(t, http.StatusBadRequest, statusResp.Code, "non-loopback non-https should be rejected")

	// Missing code challenge (PKCE required)
	initReq["redirect_uri"] = "http://127.0.0.1:54321/callback"
	initReq["code_challenge"] = ""
	statusResp, _ = app.Post("/api/user/auth/cli/initiate", initReq)
	require.Equal(t, http.StatusBadRequest, statusResp.Code, "missing code_challenge should be rejected")
}

// TestCLILoginDisabledFeature tests behavior when CLI login is disabled
func TestCLILoginDisabledFeature(t *testing.T) {
	// Create app with CLI login disabled
	schemaDir := t.TempDir()
	migrationDir := t.TempDir()

	sb, _ := schema.NewBuilderFromDir(schemaDir, fs.SystemSchemaTypes...)
	dbc, _ := entdbadapter.NewEntClient(&db.Config{
		Driver:       "sqlite",
		Name:         path.Join(t.TempDir(), "test_cli_disabled.db"),
		MigrationDir: migrationDir,
		LogQueries:   false,
	}, sb)
	t.Cleanup(func() { _ = dbc.Close() })

	app := &CLITestApp{
		t:              t,
		db:             dbc,
		schemaBuilder:  sb,
		schemaDir:      schemaDir,
		cliLoginConfig: nil, // Disabled
	}

	// Create roles
	roleModel, _ := dbc.Model("role")
	roleIDMap := make(map[string]uuid.UUID)
	for _, role := range []*fs.Role{fs.RoleAdmin, fs.RoleUser, fs.RoleGuest} {
		id, _ := roleModel.Create(context.Background(), entity.New().
			Set("name", role.Name).
			Set("root", role.Root).
			Set("system", role.System))
		roleIDMap[role.Name] = id.(uuid.UUID)
	}

	fs.RoleAdmin.ID = roleIDMap[fs.RoleAdmin.Name]
	fs.RoleUser.ID = roleIDMap[fs.RoleUser.Name]
	fs.RoleGuest.ID = roleIDMap[fs.RoleGuest.Name]

	// Setup services and resources
	app.authService = as.New(app)
	app.resources = fs.NewResourcesManager()
	app.resources.Hooks = func() *fs.Hooks {
		return &fs.Hooks{
			PreResolve: []fs.Middleware{app.authService.Authorize},
		}
	}
	app.resources.Middlewares = append(app.resources.Middlewares, app.authService.ParseUser)

	api := app.resources.Group("api", &fs.Meta{Prefix: "/api"})
	userGroup := api.Group("user")
	app.authService.CreateResource(userGroup, map[string]fs.AuthProvider{
		auth.ProviderLocal: app.GetAuthProvider(auth.ProviderLocal),
	})

	require.NoError(t, app.resources.Init())

	resolver := rr.NewRestfulResolver(&rr.ResolverConfig{
		ResourceManager: app.resources,
		Logger:          logger.CreateMockLogger(false),
	})
	app.server = resolver.Server()

	// Test: CLI initiate should return 403
	initReq := map[string]string{
		"redirect_uri":   "http://127.0.0.1:54321/callback",
		"code_challenge": "challenge",
	}
	statusResp, _ := app.Post("/api/user/auth/cli/initiate", initReq)
	require.Equal(t, http.StatusForbidden, statusResp.Code, "initiate should be forbidden when CLI login disabled")

	// Test: CLI local login should return 403
	localLoginReq := map[string]string{
		"login":    testUsername,
		"password": testPassword,
		"carrier":  "dummy",
	}
	statusResp, _ = app.Post("/api/user/auth/cli/login/local", localLoginReq)
	require.Equal(t, http.StatusForbidden, statusResp.Code, "local login should be forbidden when CLI login disabled")
}

// TestCLILoginAllowlistedHost tests that allowlisted hosts are accepted
func TestCLILoginAllowlistedHost(t *testing.T) {
	app := CreateCLITestApp(t)

	codeVerifier := "test-verifier-allowlisted"
	codeChallenge := generatePKCEChallenge(codeVerifier)

	// Loopback should always work
	initReq := map[string]string{
		"redirect_uri":   "http://127.0.0.1:54321/callback",
		"code_challenge": codeChallenge,
	}
	statusResp, _ := app.Post("/api/user/auth/cli/initiate", initReq)
	require.Equal(t, http.StatusOK, statusResp.Code, "loopback should be accepted")

	// IPv6 loopback should work
	initReq["redirect_uri"] = "http://[::1]:9/callback"
	statusResp, _ = app.Post("/api/user/auth/cli/initiate", initReq)
	require.Equal(t, http.StatusOK, statusResp.Code, "IPv6 loopback should be accepted")

	// localhost loopback should work
	initReq["redirect_uri"] = "http://localhost:5000/callback"
	statusResp, _ = app.Post("/api/user/auth/cli/initiate", initReq)
	require.Equal(t, http.StatusOK, statusResp.Code, "localhost loopback should be accepted")

	// Allowlisted host should work
	initReq["redirect_uri"] = "https://app.example.com/callback"
	statusResp, _ = app.Post("/api/user/auth/cli/initiate", initReq)
	require.Equal(t, http.StatusOK, statusResp.Code, "allowlisted host should be accepted")
}

func init() {
	os.Setenv("FASTSCHEMA_TEST", "true")
}

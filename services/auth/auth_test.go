package authservice_test

import (
	"context"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/logger"
	"github.com/fastschema/fastschema/pkg/entdbadapter"
	"github.com/fastschema/fastschema/pkg/errors"
	rr "github.com/fastschema/fastschema/pkg/restfulresolver"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
	as "github.com/fastschema/fastschema/services/auth"
	"github.com/stretchr/testify/assert"
)

type testApp struct {
	sb            *schema.Builder
	db            db.Client
	authProviders map[string]fs.AuthProvider
}

func (s testApp) DB() db.Client {
	return s.db
}

func (s testApp) Key() string {
	return "test"
}

func (s testApp) GetAuthProvider(name string) fs.AuthProvider {
	return s.authProviders[name]
}

type testAuthProvider struct{}

func (t testAuthProvider) Name() string {
	return "testauthprovider"
}

func (t testAuthProvider) Login(c fs.Context) (any, error) {
	return c.Redirect("http://auth.example.local?callback=http://localhost:8000/auth/testauthprovider/callback"), nil
}

func (t testAuthProvider) Callback(c fs.Context) (*fs.User, error) {
	shouldError := c.Arg("error")
	if shouldError != "" {
		return nil, errors.InternalServerError("error")
	}

	return &fs.User{
		Provider:         "testauthprovider",
		ProviderID:       "123",
		ProviderUsername: "testuser",
		Username:         "testuser",
		Email:            "testuser@example.local",
	}, nil
}

func createAuthService(t *testing.T) (*as.AuthService, *rr.Server) {
	schemaDir := utils.Must(os.MkdirTemp("", "schemas"))
	migrationDir := utils.Must(os.MkdirTemp("", "migrations"))
	sb := utils.Must(schema.NewBuilderFromDir(schemaDir, fs.SystemSchemaTypes...))
	dbc := utils.Must(entdbadapter.NewTestClient(migrationDir, sb))

	ctx := context.Background()
	assert.NotNil(t, utils.Must(db.Create[*fs.Role](ctx, dbc, map[string]any{"name": "admin"})))
	assert.NotNil(t, utils.Must(db.Create[*fs.Role](ctx, dbc, map[string]any{"name": "user"})))
	assert.NotNil(t, utils.Must(db.Create[*fs.Role](ctx, dbc, map[string]any{"name": "guest"})))

	authService := as.New(&testApp{
		sb: sb,
		db: dbc,
		authProviders: map[string]fs.AuthProvider{
			"testauthprovider": testAuthProvider{},
		},
	})

	resources := fs.NewResourcesManager()
	resources.Group("auth", &fs.Meta{
		Prefix: "/auth/:provider",
		Args: fs.Args{
			"provider": {
				Required:    true,
				Type:        fs.TypeString,
				Description: "The auth provider name. Available providers: testauthprovider",
				Example:     "google",
			},
		},
	}).Add(
		fs.Get("login", authService.Login, &fs.Meta{Public: true}),
		fs.Get("callback", authService.Callback, &fs.Meta{Public: true}),
	)

	assert.NoError(t, resources.Init())
	restResolver := rr.NewRestfulResolver(resources, logger.CreateMockLogger(true))

	return authService, restResolver.Server()
}

func TestNewContentService(t *testing.T) {
	service, server := createAuthService(t)
	assert.NotNil(t, service)
	assert.NotNil(t, server)

	// Case 1: provider not found
	req := httptest.NewRequest("GET", "/auth/invalidprovider/login", nil)
	resp := utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, "404 Not Found", resp.Status)
	assert.Contains(t, utils.Must(utils.ReadCloserToString(resp.Body)), `"message":"invalid provider"`)

	// Case 2: login should redirect to the auth provider
	req = httptest.NewRequest("GET", "/auth/testauthprovider/login", nil)
	resp = utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, "302 Found", resp.Status)
	assert.Equal(t, "http://auth.example.local?callback=http://localhost:8000/auth/testauthprovider/callback", resp.Header.Get("Location"))

	// Case 3: callback error invalid provider
	req = httptest.NewRequest("GET", "/auth/invalidprovider/callback", nil)
	resp = utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, "404 Not Found", resp.Status)
	assert.Contains(t, utils.Must(utils.ReadCloserToString(resp.Body)), `"message":"invalid provider"`)

	// Case 4: callback error due to provider error
	req = httptest.NewRequest("GET", "/auth/testauthprovider/callback?error=1", nil)
	resp = utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, "500 Internal Server Error", resp.Status)

	// Case 5: callback success
	req = httptest.NewRequest("GET", "/auth/testauthprovider/callback", nil)
	resp = utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	response := utils.Must(utils.ReadCloserToString(resp.Body))
	assert.Equal(t, "200 OK", resp.Status)
	assert.Contains(t, response, `"token":`)
	assert.Contains(t, response, `"expires":`)
}

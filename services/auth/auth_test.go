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
	resources     *fs.ResourcesManager
	restResolver  *rr.RestfulResolver
	authService   *as.AuthService

	adminUser         *fs.User
	normalUser        *fs.User
	inactiveUser      *fs.User
	adminToken        string
	normalUserToken   string
	inactiveUserToken string
}

func (s testApp) DB() db.Client {
	return s.db
}

func (s testApp) Key() string {
	return "test"
}

func (s testApp) Roles() []*fs.Role {
	return utils.Must(
		db.Query[*fs.Role](s.db).Select("id", "name", "root", "permissions").Get(context.Background()),
	)
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

func createTestApp(t *testing.T) *testApp {
	schemaDir := utils.Must(os.MkdirTemp("", "schemas"))
	utils.WriteFile(schemaDir+"/blog.json", `{
		"name": "blog",
		"namespace": "blogs",
		"label_field": "name",
		"fields": [
			{
				"type": "string",
				"name": "name",
				"label": "Name",
				"sortable": true
			}
		]
	}`)
	migrationDir := utils.Must(os.MkdirTemp("", "migrations"))
	sb := utils.Must(schema.NewBuilderFromDir(schemaDir, fs.SystemSchemaTypes...))
	dbc := utils.Must(entdbadapter.NewTestClient(migrationDir, sb))

	roleModel := utils.Must(dbc.Model("role"))
	userModel := utils.Must(dbc.Model("user"))
	appRoles := []*fs.Role{fs.RoleAdmin, fs.RoleUser, fs.RoleGuest}

	for _, r := range appRoles {
		utils.Must(roleModel.Create(context.Background(), schema.NewEntity().
			Set("name", r.Name).
			Set("root", r.Root),
		))
	}

	utils.Must(userModel.Create(context.Background(), schema.NewEntity().
		Set("username", "adminuser").
		Set("password", "adminuser").
		Set("roles", []*schema.Entity{schema.NewEntity(1)}),
	))

	utils.Must(userModel.Create(context.Background(), schema.NewEntity().
		Set("username", "normaluser").
		Set("password", "normaluser").
		Set("roles", []*schema.Entity{schema.NewEntity(2)}),
	))

	// There are three resources in this test: content.list, content.detail and content.meta
	// We set role user to have permission to "allow" for content.list but, "deny" for content.detail
	// And no permission set for content.meta
	// We expect that user with role user should have access to content.list but not content.detail and content.meta
	permissionModel := utils.Must(dbc.Model("permission"))
	utils.Must(permissionModel.Create(context.Background(), schema.NewEntity().
		Set("resource", "api.content.blog.list").
		Set("value", fs.PermissionTypeAllow.String()).
		Set("role_id", fs.RoleUser.ID),
	))
	utils.Must(permissionModel.Create(context.Background(), schema.NewEntity().
		Set("resource", "api.content.blog.detail").
		Set("value", fs.PermissionTypeDeny.String()).
		Set("role_id", fs.RoleUser.ID),
	))

	testApp := &testApp{
		sb: sb,
		db: dbc,
		authProviders: map[string]fs.AuthProvider{
			"testauthprovider": testAuthProvider{},
		},
		adminUser: &fs.User{
			ID:       1,
			Username: "adminuser",
			Active:   true,
			Roles:    []*fs.Role{fs.RoleAdmin},
			RoleIDs:  []uint64{1},
		},
		normalUser: &fs.User{
			ID:       2,
			Username: "normaluser",
			Active:   true,
			Roles:    []*fs.Role{fs.RoleUser},
			RoleIDs:  []uint64{2},
		},
		inactiveUser: &fs.User{
			ID:       3,
			Username: "inactiveuser",
			Active:   false,
			Roles:    []*fs.Role{fs.RoleUser},
			RoleIDs:  []uint64{2},
		},
	}

	testApp.adminToken, _, _ = testApp.adminUser.JwtClaim(testApp.Key())
	testApp.normalUserToken, _, _ = testApp.normalUser.JwtClaim(testApp.Key())
	testApp.inactiveUserToken, _, _ = testApp.inactiveUser.JwtClaim(testApp.Key())

	testApp.authService = as.New(testApp)
	testApp.resources = fs.NewResourcesManager()
	testApp.resources.Hooks = func() *fs.Hooks {
		return &fs.Hooks{
			PreResolve: []fs.Middleware{testApp.authService.Authorize},
		}
	}
	testApp.resources.Middlewares = append(testApp.resources.Middlewares, testApp.authService.ParseUser)

	apiGroup := testApp.resources.Group("api", &fs.Meta{Prefix: "/api"})
	apiGroup.Group("auth", &fs.Meta{
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
		fs.Get("login", testApp.authService.Login, &fs.Meta{Public: true}),
		fs.Get("callback", testApp.authService.Callback, &fs.Meta{Public: true}),
	)

	apiGroup.Group("content").
		Add(fs.NewResource("list", func(c fs.Context, _ any) (any, error) {
			return "blog list", nil
		}, &fs.Meta{
			Get: "/:schema",
		})).
		Add(fs.NewResource("detail", func(c fs.Context, _ any) (any, error) {
			return "blog detail", nil
		}, &fs.Meta{
			Get: "/:schema/:id",
		})).
		Add(fs.NewResource("meta", func(c fs.Context, _ any) (any, error) {
			return "blog meta", nil
		}, &fs.Meta{
			Get: "/:schema/meta",
		}))

	apiGroup.
		Add(
			fs.NewResource("testuser", func(c fs.Context, _ any) (any, error) {
				return c.User(), nil
			}, &fs.Meta{Public: true}),
		).
		Add(
			fs.NewResource("test", func(c fs.Context, _ any) (any, error) {
				return "test response", nil
			}, &fs.Meta{Public: true}),
		)

	assert.NoError(t, testApp.resources.Init())
	testApp.restResolver = rr.NewRestfulResolver(&rr.ResolverConfig{
		ResourceManager: testApp.resources,
		Logger:          logger.CreateMockLogger(true),
	})

	return testApp
}

func TestNewContentService(t *testing.T) {
	testApp := createTestApp(t)
	server := testApp.restResolver.Server()
	assert.NotNil(t, testApp)
	assert.NotNil(t, server)

	// Case 1: provider not found
	req := httptest.NewRequest("GET", "/api/auth/invalidprovider/login", nil)
	resp := utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, "404 Not Found", resp.Status)
	assert.Contains(t, utils.Must(utils.ReadCloserToString(resp.Body)), `"message":"invalid provider"`)

	// Case 2: login should redirect to the auth provider
	req = httptest.NewRequest("GET", "/api/auth/testauthprovider/login", nil)
	resp = utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, "302 Found", resp.Status)
	assert.Equal(t, "http://auth.example.local?callback=http://localhost:8000/auth/testauthprovider/callback", resp.Header.Get("Location"))

	// Case 3: callback error invalid provider
	req = httptest.NewRequest("GET", "/api/auth/invalidprovider/callback", nil)
	resp = utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, "404 Not Found", resp.Status)
	assert.Contains(t, utils.Must(utils.ReadCloserToString(resp.Body)), `"message":"invalid provider"`)

	// Case 4: callback error due to provider error
	req = httptest.NewRequest("GET", "/api/auth/testauthprovider/callback?error=1", nil)
	resp = utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, "500 Internal Server Error", resp.Status)

	// Case 5: callback success
	req = httptest.NewRequest("GET", "/api/auth/testauthprovider/callback", nil)
	resp = utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	response := utils.Must(utils.ReadCloserToString(resp.Body))
	assert.Equal(t, "200 OK", resp.Status)
	assert.Contains(t, response, `"token":`)
	assert.Contains(t, response, `"expires":`)
}

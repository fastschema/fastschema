package authservice_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/entity"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/logger"
	"github.com/fastschema/fastschema/pkg/auth"
	"github.com/fastschema/fastschema/pkg/entdbadapter"
	"github.com/fastschema/fastschema/pkg/errors"
	"github.com/fastschema/fastschema/pkg/jwt"
	rr "github.com/fastschema/fastschema/pkg/restfulresolver"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
	as "github.com/fastschema/fastschema/services/auth"
)

type testApp struct {
	sb            *schema.Builder
	db            db.Client
	authProviders map[string]fs.AuthProvider
	resources     *fs.ResourcesManager
	restResolver  *rr.RestfulResolver
	authService   *as.AuthService

	adminUser          *fs.User
	normalUser         *fs.User
	inactiveUser       *fs.User
	seniorityUser      *fs.User
	notFoundUser       *fs.User
	adminToken         string
	normalUserToken    string
	inactiveUserToken  string
	seniorityUserToken string
	notFoundUserToken  string
}

func (s testApp) JwtCustomClaimsFunc() fs.JwtCustomClaimsFunc {
	return nil
}

func (s testApp) DB() db.Client {
	return s.db
}

func (s testApp) Key() string {
	return "test"
}

func (s testApp) Config() *fs.Config {
	return &fs.Config{
		AppKey: "test",
		AuthConfig: &fs.AuthConfig{
			EnableRefreshToken: false,
		},
	}
}

func (s testApp) Roles() []*fs.Role {
	roles := utils.Must(db.Builder[*fs.Role](s.db).Select(
		"id",
		"name",
		"description",
		"root",
		"rule",
		"permissions",
		entity.FieldCreatedAt,
		entity.FieldUpdatedAt,
		entity.FieldDeletedAt,
	).Get(context.Background()))

	for _, role := range roles {
		utils.Must("", role.Compile())

		for _, permission := range role.Permissions {
			utils.Must("", permission.Compile())
		}
	}

	return roles
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

	if c.Arg("niluser") != "" {
		return nil, nil
	}

	return &fs.User{
		Provider:         "testauthprovider",
		ProviderID:       "123",
		ProviderUsername: "testuser",
		Username:         "testuser",
		Email:            "testuser@example.local",
	}, nil
}

func (t testAuthProvider) VerifyIDToken(c fs.Context, p fs.IDToken) (*fs.User, error) {
	return nil, nil
}

func createTestApp(t *testing.T) *testApp {
	schemaDir := utils.Must(os.MkdirTemp("", "schemas"))
	utils.WriteFile(schemaDir+"/blog.yaml", `name: blog
namespace: blogs
label_field: name
fields:
  - name: name
    label: Name
    type: string
    sortable: true
`)
	migrationDir := utils.Must(os.MkdirTemp("", "migrations"))
	sb := utils.Must(schema.NewBuilderFromDir(schemaDir, fs.SystemSchemaTypes...))
	dbc := utils.Must(entdbadapter.NewTestClient(migrationDir, sb))

	roleModel := utils.Must(dbc.Model("role"))
	userModel := utils.Must(dbc.Model("user"))
	seniorityRole := &fs.Role{
		ID:   4,
		Name: "seniority",
		Root: false,
		Rule: `let userId = $context.User().ID;
		let foundUsers = $db.Query($context, "SELECT created_at FROM users WHERE id = ?", userId);
		foundUsers[0].Get("created_at") < date("2023-01-01")
		`,
	}
	appRoles := []*fs.Role{
		fs.RoleAdmin,
		fs.RoleUser,
		fs.RoleGuest,
		seniorityRole,
	}

	for _, r := range appRoles {
		utils.Must(roleModel.Create(context.Background(), entity.New().
			Set("name", r.Name).
			Set("root", r.Root).
			Set("rule", r.Rule),
		))
	}

	utils.Must(userModel.Create(context.Background(), entity.New().
		Set("username", "adminuser").
		Set("password", "adminuser").
		Set("provider", "local").
		Set("active", true).
		Set("roles", []*entity.Entity{entity.New(1)}),
	))

	utils.Must(userModel.Create(context.Background(), entity.New().
		Set("username", "normaluser").
		Set("password", "normaluser").
		Set("provider", "local").
		Set("active", true).
		Set("roles", []*entity.Entity{entity.New(2)}),
	))

	utils.Must(userModel.Create(context.Background(), entity.New().
		Set("username", "inactiveuser").
		Set("password", "inactiveuser").
		Set("provider", "local").
		Set("active", false).
		Set("roles", []*entity.Entity{entity.New(2)}),
	))

	utils.Must(userModel.Create(context.Background(), entity.New().
		Set("username", "seniorityuser").
		Set("password", "seniorityuser").
		Set("provider", "local").
		Set("active", true).
		Set("roles", []*entity.Entity{entity.New(4)}),
	))

	// There are three resources in this test: content.list, content.detail and content.meta
	// We set role user to have permission to "allow" for content.list but, "deny" for content.detail
	// And no permission set for content.meta
	// We expect that user with role user should have access to content.list but not content.detail and content.meta
	permissionModel := utils.Must(dbc.Model("permission"))

	// Role user should have access to api.content.blog.list
	utils.Must(permissionModel.Create(context.Background(), entity.New().
		Set("resource", "api.content.blog.list").
		Set("value", fs.PermissionTypeAllow.String()).
		Set("role_id", fs.RoleUser.ID),
	))

	// Role user should not have access to api.content.blog.detail
	utils.Must(permissionModel.Create(context.Background(), entity.New().
		Set("resource", "api.content.blog.detail").
		Set("value", fs.PermissionTypeDeny.String()).
		Set("role_id", fs.RoleUser.ID),
	))

	// Role seniority should have access to api.content.blog.list
	utils.Must(permissionModel.Create(context.Background(), entity.New().
		Set("resource", "api.content.blog.list").
		Set("value", fs.PermissionTypeAllow.String()).
		Set("role_id", seniorityRole.ID),
	))

	// Realtime permissions for role user
	// Role user should have access to api.realtime.content.blog.list
	utils.Must(permissionModel.Create(context.Background(), entity.New().
		Set("resource", "api.realtime.content.blog.list").
		Set("value", fs.PermissionTypeAllow.String()).
		Set("role_id", fs.RoleUser.ID),
	))

	// Role user should not have access to api.realtime.content.blog.update
	utils.Must(permissionModel.Create(context.Background(), entity.New().
		Set("resource", "api.realtime.content.blog.update").
		Set("value", fs.PermissionTypeDeny.String()).
		Set("role_id", fs.RoleUser.ID),
	))

	localProvider := utils.Must(auth.NewLocalAuthProvider(fs.Map{}, ""))
	testApp := &testApp{
		sb: sb,
		db: dbc,
		authProviders: map[string]fs.AuthProvider{
			"testauthprovider": testAuthProvider{},
			"local":            localProvider,
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
		seniorityUser: &fs.User{
			ID:       4,
			Username: "seniorityuser",
			Active:   true,
			Roles:    []*fs.Role{seniorityRole},
			RoleIDs:  []uint64{4},
		},
		notFoundUser: &fs.User{
			ID:       5,
			Username: "notfounduser",
			Active:   true,
			Roles:    []*fs.Role{fs.RoleUser},
			RoleIDs:  []uint64{2},
		},
	}

	key := testApp.Key()
	testApp.adminToken, _, _ = jwt.GenerateAccessToken(
		jwt.UserToJwtClaims(testApp.adminUser),
		key, time.Time{}, nil,
	)
	testApp.normalUserToken, _, _ = jwt.GenerateAccessToken(
		jwt.UserToJwtClaims(testApp.normalUser),
		key, time.Time{}, nil,
	)
	testApp.inactiveUserToken, _, _ = jwt.GenerateAccessToken(
		jwt.UserToJwtClaims(testApp.inactiveUser),
		key, time.Time{}, nil,
	)
	testApp.seniorityUserToken, _, _ = jwt.GenerateAccessToken(
		jwt.UserToJwtClaims(testApp.seniorityUser),
		key, time.Time{}, nil,
	)
	testApp.notFoundUserToken, _, _ = jwt.GenerateAccessToken(
		jwt.UserToJwtClaims(testApp.notFoundUser),
		key, time.Time{}, nil,
	)

	testApp.authService = as.New(testApp)
	testApp.resources = fs.NewResourcesManager()
	testApp.resources.Hooks = func() *fs.Hooks {
		return &fs.Hooks{
			PreResolve: []fs.Middleware{testApp.authService.Authorize},
		}
	}
	testApp.resources.Middlewares = append(testApp.resources.Middlewares, testApp.authService.ParseUser)

	apiGroup := testApp.resources.Group("api", &fs.Meta{Prefix: "/api"})
	apiGroup.Group("auth").
		Add(fs.Get("me", testApp.authService.Me, &fs.Meta{Public: true})).
		Group(":provider", &fs.Meta{
			Prefix: "/:provider",
			Args: fs.Args{
				"provider": {
					Required:    true,
					Type:        fs.TypeString,
					Description: "The auth provider name. Available providers: testauthprovider",
					Example:     "google",
				},
			},
		}).
		Add(
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
		Group("realtime").
		Add(fs.NewResource("content", func(c fs.Context, _ any) (any, error) {
			return "realtime content", nil
		}, &fs.Meta{Get: "/content"}))

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
		Logger:          logger.CreateMockLogger(false),
	})

	return testApp
}

func TestCreateResource(t *testing.T) {
	testApp := createTestApp(t)
	api := fs.NewResourcesManager().Group("api")
	testApp.authService.CreateResource(api, testApp.authProviders)

	assert.NotNil(t, api.Find("api.auth.me"))
	assert.NotNil(t, api.Find("api.auth.local.login"))
	assert.NotNil(t, api.Find("api.auth.local.register"))
	assert.NotNil(t, api.Find("api.auth.local.activate"))
	assert.NotNil(t, api.Find("api.auth.local.activate/send"))
	assert.NotNil(t, api.Find("api.auth.local.recover"))
	assert.NotNil(t, api.Find("api.auth.local.recover/check"))
	assert.NotNil(t, api.Find("api.auth.local.recover/reset"))
	assert.NotNil(t, api.Find("api.auth.local.recover"))
	assert.NotNil(t, api.Find("api.auth.provider.login"))
	assert.NotNil(t, api.Find("api.auth.provider.callback"))
}

package roleservice_test

import (
	"os"
	"testing"

	"github.com/fastschema/fastschema/app"
	"github.com/fastschema/fastschema/pkg/entdbadapter"
	rr "github.com/fastschema/fastschema/pkg/restresolver"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
	rs "github.com/fastschema/fastschema/services/role"
	"github.com/stretchr/testify/assert"
)

type TestApp struct {
	sb                *schema.Builder
	db                app.DBClient
	resources         *app.ResourcesManager
	adminUser         *app.User
	normalUser        *app.User
	inactiveUser      *app.User
	roleService       *rs.RoleService
	roleModel         app.Model
	server            *rr.Server
	adminToken        string
	normalUserToken   string
	inactiveUserToken string
}

func (s TestApp) DB() app.DBClient {
	return s.db
}

func (s TestApp) Roles() []*app.Role {
	roleEntities := utils.Must(s.roleModel.Query().Select("id", "name", "root", "permissions").Get())
	return app.EntitiesToRoles(roleEntities)
}

func (s TestApp) Key() string {
	return "test"
}

func (s TestApp) UpdateCache() error {
	return nil
}

func (s TestApp) Resources() *app.ResourcesManager {
	return s.resources
}

func createRoleTest() *TestApp {
	schemaDir := utils.Must(os.MkdirTemp("", "schema"))
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
	sb := utils.Must(schema.NewBuilderFromDir(schemaDir))
	db := utils.Must(entdbadapter.NewTestClient(utils.Must(os.MkdirTemp("", "migrations")), sb))
	roleModel := utils.Must(db.Model("role"))
	userModel := utils.Must(db.Model("user"))
	appRoles := []*app.Role{app.RoleAdmin, app.RoleUser, app.RoleGuest}

	for _, r := range appRoles {
		utils.Must(roleModel.Create(schema.NewEntity().
			Set("name", r.Name).
			Set("root", r.Root),
		))
	}

	utils.Must(userModel.Create(schema.NewEntity().
		Set("username", "adminuser").
		Set("password", "adminuser").
		Set("roles", []*schema.Entity{schema.NewEntity(1)}),
	))

	utils.Must(userModel.Create(schema.NewEntity().
		Set("username", "normaluser").
		Set("password", "normaluser").
		Set("roles", []*schema.Entity{schema.NewEntity(2)}),
	))

	// There are three resources in this test: content.list, content.detail and content.meta
	// We set role user to have permission to "allow" for content.list but, "deny" for content.detail
	// And no permission set for content.meta
	// We expect that user with role user should have access to content.list but not content.detail and content.meta
	permissionModel := utils.Must(db.Model("permission"))
	utils.Must(permissionModel.Create(schema.NewEntity().
		Set("resource", "content.blog.list").
		Set("value", app.PermissionTypeAllow.String()).
		Set("role_id", app.RoleUser.ID),
	))
	utils.Must(permissionModel.Create(schema.NewEntity().
		Set("resource", "content.blog.detail").
		Set("value", app.PermissionTypeDeny.String()).
		Set("role_id", app.RoleUser.ID),
	))

	testApp := &TestApp{
		sb:        sb,
		db:        db,
		roleModel: roleModel,
	}

	testApp.adminUser = &app.User{
		ID:       1,
		Username: "adminuser",
		Active:   true,
		Roles:    []*app.Role{app.RoleAdmin},
		RoleIDs:  []uint64{1},
	}
	testApp.normalUser = &app.User{
		ID:       2,
		Username: "normaluser",
		Active:   true,
		Roles:    []*app.Role{app.RoleUser},
		RoleIDs:  []uint64{2},
	}
	testApp.inactiveUser = &app.User{
		ID:       3,
		Username: "inactiveuser",
		Active:   false,
		Roles:    []*app.Role{app.RoleUser},
		RoleIDs:  []uint64{2},
	}

	testApp.adminToken, _, _ = testApp.adminUser.JwtClaim(testApp.Key())
	testApp.normalUserToken, _, _ = testApp.normalUser.JwtClaim(testApp.Key())
	testApp.inactiveUserToken, _, _ = testApp.inactiveUser.JwtClaim(testApp.Key())

	testApp.roleService = rs.New(testApp)
	testApp.resources = app.NewResourcesManager()
	testApp.resources.Hooks = func() *app.Hooks {
		return &app.Hooks{
			PreResolve: []app.Middleware{testApp.roleService.Authorize},
		}
	}
	testApp.resources.Middlewares = append(testApp.resources.Middlewares, testApp.roleService.ParseUser)
	testApp.resources.Group("role").
		Add(app.NewResource("list", testApp.roleService.List, &app.Meta{
			Get: "/",
		})).
		Add(app.NewResource("resources", testApp.roleService.ResourcesList, &app.Meta{
			Get: "/resources",
		})).
		Add(app.NewResource("detail", testApp.roleService.Detail, &app.Meta{
			Get: "/:id",
		})).
		Add(app.NewResource("create", testApp.roleService.Create, &app.Meta{
			Post: "/",
		})).
		Add(app.NewResource("update", testApp.roleService.Update, &app.Meta{
			Put: "/:id",
		})).
		Add(app.NewResource("delete", testApp.roleService.Delete, &app.Meta{
			Delete: "/:id",
		}))

	testApp.resources.Group("content").
		Add(app.NewResource("list", func(c app.Context, _ any) (any, error) {
			return "blog list", nil
		}, &app.Meta{
			Get: "/:schema",
		})).
		Add(app.NewResource("detail", func(c app.Context, _ any) (any, error) {
			return "blog detail", nil
		}, &app.Meta{
			Get: "/:schema/:id",
		})).
		Add(app.NewResource("meta", func(c app.Context, _ any) (any, error) {
			return "blog meta", nil
		}, &app.Meta{
			Get: "/:schema/meta",
		}))

	testApp.resources.
		Add(
			app.NewResource("testuser", func(c app.Context, _ any) (any, error) {
				return c.User(), nil
			}, &app.Meta{Public: true}),
		).
		Add(
			app.NewResource("test", func(c app.Context, _ any) (any, error) {
				return "test response", nil
			}, &app.Meta{Public: true}),
		)

	if err := testApp.resources.Init(); err != nil {
		panic(err)
	}

	testApp.server = rr.NewRestResolver(testApp.resources, app.CreateMockLogger()).Server()

	return testApp
}

func TestNewRoleService(t *testing.T) {
	testApp := createRoleTest()
	assert.NotNil(t, testApp)
	assert.NotNil(t, testApp.roleService)
	assert.NotNil(t, testApp.server)
}

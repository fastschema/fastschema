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
	sb        *schema.Builder
	db        app.DBClient
	resources *app.ResourcesManager
}

func (s TestApp) DB() app.DBClient {
	return s.db
}

func (s TestApp) Roles() []*app.Role {
	roleEntities := utils.Must(roleModel.Query().Select("id", "name", "root", "permissions").Get())
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

var (
	adminUser         *app.User
	normalUser        *app.User
	inactiveUser      *app.User
	testApp           *TestApp
	roleService       *rs.RoleService
	roleModel         app.Model
	server            *rr.Server
	adminToken        string
	normalUserToken   string
	inactiveUserToken string
)

func init() {
	schemaDir := os.TempDir()
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
	db := utils.Must(entdbadapter.NewTestClient(os.TempDir(), sb))
	roleModel = utils.Must(db.Model("role"))
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

	testApp = &TestApp{
		sb: sb,
		db: db,
	}

	adminUser = &app.User{
		ID:       1,
		Username: "adminuser",
		Active:   true,
		Roles:    []*app.Role{app.RoleAdmin},
		RoleIDs:  []uint64{1},
	}
	normalUser = &app.User{
		ID:       2,
		Username: "normaluser",
		Active:   true,
		Roles:    []*app.Role{app.RoleUser},
		RoleIDs:  []uint64{2},
	}
	inactiveUser = &app.User{
		ID:       3,
		Username: "inactiveuser",
		Active:   false,
		Roles:    []*app.Role{app.RoleUser},
		RoleIDs:  []uint64{2},
	}

	adminToken, _, _ = adminUser.JwtClaim(testApp.Key())
	normalUserToken, _, _ = normalUser.JwtClaim(testApp.Key())
	inactiveUserToken, _, _ = inactiveUser.JwtClaim(testApp.Key())

	roleService = rs.New(testApp)
	testApp.resources = app.NewResourcesManager()
	testApp.resources.Middlewares = append(testApp.resources.Middlewares, roleService.ParseUser)
	testApp.resources.BeforeResolveHooks = append(testApp.resources.BeforeResolveHooks, roleService.Authorize)
	testApp.resources.Group("role").
		Add(app.NewResource("list", roleService.List, app.Meta{app.GET: ""})).
		Add(app.NewResource("resources", roleService.ResourcesList, app.Meta{app.GET: "/resources"})).
		Add(app.NewResource("detail", roleService.Detail, app.Meta{app.GET: "/:id"})).
		Add(app.NewResource("create", roleService.Create, app.Meta{app.POST: ""})).
		Add(app.NewResource("update", roleService.Update, app.Meta{app.PUT: "/:id"})).
		Add(app.NewResource("delete", roleService.Delete, app.Meta{app.DELETE: "/:id"}))

	testApp.resources.Group("content").
		Add(app.NewResource("list", func(c app.Context, _ *any) (any, error) {
			return "blog list", nil
		}, app.Meta{app.GET: "/:schema"})).
		Add(app.NewResource("detail", func(c app.Context, _ *any) (any, error) {
			return "blog detail", nil
		}, app.Meta{app.GET: "/:schema/:id"})).
		Add(app.NewResource("meta", func(c app.Context, _ *any) (any, error) {
			return "blog meta", nil
		}, app.Meta{app.GET: "/:schema/meta"}))

	testApp.resources.
		Add(
			app.NewResource("testuser", func(c app.Context, _ *any) (any, error) {
				return c.User(), nil
			}, true),
		).
		Add(
			app.NewResource("test", func(c app.Context, _ *any) (any, error) {
				return "test response", nil
			}, true),
		)

	if err := testApp.resources.Init(); err != nil {
		panic(err)
	}

	server = rr.NewRestResolver(testApp.resources).Init(app.CreateMockLogger(true)).Server()
}

func TestNewRoleService(t *testing.T) {
	assert.NotNil(t, testApp)
	assert.NotNil(t, roleService)
	assert.NotNil(t, server)
}

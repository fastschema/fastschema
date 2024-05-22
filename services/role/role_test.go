package roleservice_test

import (
	"context"
	"os"
	"testing"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/logger"
	"github.com/fastschema/fastschema/pkg/entdbadapter"
	rr "github.com/fastschema/fastschema/pkg/restresolver"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
	rs "github.com/fastschema/fastschema/services/role"
	"github.com/stretchr/testify/assert"
)

type TestApp struct {
	sb                *schema.Builder
	db                db.Client
	resources         *fs.ResourcesManager
	adminUser         *fs.User
	normalUser        *fs.User
	inactiveUser      *fs.User
	roleService       *rs.RoleService
	roleModel         db.Model
	server            *rr.Server
	adminToken        string
	normalUserToken   string
	inactiveUserToken string
}

func (s TestApp) DB() db.Client {
	return s.db
}

func (s TestApp) Roles() []*fs.Role {
	return utils.Must(
		db.Query[*fs.Role](s.db).Select("id", "name", "root", "permissions").Get(context.Background()),
	)
}

func (s TestApp) Key() string {
	return "test"
}

func (s TestApp) UpdateCache(ctx context.Context) error {
	return nil
}

func (s TestApp) Resources() *fs.ResourcesManager {
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
	sb := utils.Must(schema.NewBuilderFromDir(schemaDir, fs.SystemSchemaTypes...))
	db := utils.Must(entdbadapter.NewTestClient(utils.Must(os.MkdirTemp("", "migrations")), sb))
	roleModel := utils.Must(db.Model("role"))
	userModel := utils.Must(db.Model("user"))
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
	permissionModel := utils.Must(db.Model("permission"))
	utils.Must(permissionModel.Create(context.Background(), schema.NewEntity().
		Set("resource", "content.blog.list").
		Set("value", fs.PermissionTypeAllow.String()).
		Set("role_id", fs.RoleUser.ID),
	))
	utils.Must(permissionModel.Create(context.Background(), schema.NewEntity().
		Set("resource", "content.blog.detail").
		Set("value", fs.PermissionTypeDeny.String()).
		Set("role_id", fs.RoleUser.ID),
	))

	testApp := &TestApp{
		sb:        sb,
		db:        db,
		roleModel: roleModel,
	}

	testApp.adminUser = &fs.User{
		ID:       1,
		Username: "adminuser",
		Active:   true,
		Roles:    []*fs.Role{fs.RoleAdmin},
		RoleIDs:  []uint64{1},
	}
	testApp.normalUser = &fs.User{
		ID:       2,
		Username: "normaluser",
		Active:   true,
		Roles:    []*fs.Role{fs.RoleUser},
		RoleIDs:  []uint64{2},
	}
	testApp.inactiveUser = &fs.User{
		ID:       3,
		Username: "inactiveuser",
		Active:   false,
		Roles:    []*fs.Role{fs.RoleUser},
		RoleIDs:  []uint64{2},
	}

	testApp.adminToken, _, _ = testApp.adminUser.JwtClaim(testApp.Key())
	testApp.normalUserToken, _, _ = testApp.normalUser.JwtClaim(testApp.Key())
	testApp.inactiveUserToken, _, _ = testApp.inactiveUser.JwtClaim(testApp.Key())

	testApp.roleService = rs.New(testApp)
	testApp.resources = fs.NewResourcesManager()
	testApp.resources.Hooks = func() *fs.Hooks {
		return &fs.Hooks{
			PreResolve: []fs.Middleware{testApp.roleService.Authorize},
		}
	}
	testApp.resources.Middlewares = append(testApp.resources.Middlewares, testApp.roleService.ParseUser)
	testApp.resources.Group("role").
		Add(fs.NewResource("list", testApp.roleService.List, &fs.Meta{
			Get: "/",
		})).
		Add(fs.NewResource("resources", testApp.roleService.ResourcesList, &fs.Meta{
			Get: "/resources",
		})).
		Add(fs.NewResource("detail", testApp.roleService.Detail, &fs.Meta{
			Get: "/:id",
		})).
		Add(fs.NewResource("create", testApp.roleService.Create, &fs.Meta{
			Post: "/",
		})).
		Add(fs.NewResource("update", testApp.roleService.Update, &fs.Meta{
			Put: "/:id",
		})).
		Add(fs.NewResource("delete", testApp.roleService.Delete, &fs.Meta{
			Delete: "/:id",
		}))

	testApp.resources.Group("content").
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

	testApp.resources.
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

	if err := testApp.resources.Init(); err != nil {
		panic(err)
	}

	testApp.server = rr.NewRestResolver(testApp.resources, logger.CreateMockLogger()).Server()

	return testApp
}

func TestNewRoleService(t *testing.T) {
	testApp := createRoleTest()
	assert.NotNil(t, testApp)
	assert.NotNil(t, testApp.roleService)
	assert.NotNil(t, testApp.server)
}

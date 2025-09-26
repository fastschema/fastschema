package services_test

import (
	"context"
	"testing"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/logger"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
	"github.com/fastschema/fastschema/services"
	authservice "github.com/fastschema/fastschema/services/auth"
	contentservice "github.com/fastschema/fastschema/services/content"
	fileservice "github.com/fastschema/fastschema/services/file"
	realtimeservice "github.com/fastschema/fastschema/services/realtime"
	roleservice "github.com/fastschema/fastschema/services/role"
	schemaservice "github.com/fastschema/fastschema/services/schema"
	toolservice "github.com/fastschema/fastschema/services/tool"
	"github.com/stretchr/testify/assert"
)

type testApp struct {
	services *services.Services
}

func (a *testApp) Key() string                                 { return "" }
func (a *testApp) Name() string                                { return "" }
func (a *testApp) Config() *fs.Config                          { return nil }
func (a *testApp) DB() db.Client                               { return nil }
func (a *testApp) Disk(...string) fs.Disk                      { return nil }
func (a *testApp) Disks() []fs.Disk                            { return nil }
func (a *testApp) GetAuthProvider(string) fs.AuthProvider      { return nil }
func (a *testApp) Reload(context.Context, *db.Migration) error { return nil }
func (a *testApp) Logger() logger.Logger                       { return nil }
func (a *testApp) Roles() []*fs.Role                           { return nil }
func (a *testApp) Resources() *fs.ResourcesManager             { return nil }
func (a *testApp) SchemaBuilder() *schema.Builder              { return nil }
func (a *testApp) SystemSchemas() []any                        { return fs.SystemSchemaTypes }
func (a *testApp) UpdateCache(context.Context) error           { return nil }
func (a *testApp) Mailer(...string) fs.Mailer                  { return nil }
func (a *testApp) Mailers() []fs.Mailer                        { return nil }
func (a *testApp) Hooks() *fs.Hooks                            { return nil }
func (a *testApp) AddMiddlewares(hooks ...fs.Middleware)       {}
func (a *testApp) AddResource(resource *fs.Resource)           {}
func (a *testApp) OnPreResolve(hooks ...fs.Middleware)         {}
func (a *testApp) OnPostResolve(hooks ...fs.Middleware)        {}
func (a *testApp) OnPreDBQuery(hooks ...db.PreDBQuery)         {}
func (a *testApp) OnPostDBQuery(hooks ...db.PostDBQuery)       {}
func (a *testApp) OnPreDBExec(hooks ...db.PreDBExec)           {}
func (a *testApp) OnPostDBExec(hooks ...db.PostDBExec)         {}
func (a *testApp) OnPreDBCreate(hooks ...db.PreDBCreate)       {}
func (a *testApp) OnPostDBCreate(hooks ...db.PostDBCreate)     {}
func (a *testApp) OnPreDBUpdate(hooks ...db.PreDBUpdate)       {}
func (a *testApp) OnPostDBUpdate(hooks ...db.PostDBUpdate)     {}
func (a *testApp) OnPreDBDelete(hooks ...db.PreDBDelete)       {}
func (a *testApp) OnPostDBDelete(hooks ...db.PostDBDelete)     {}
func (a *testApp) Services() *services.Services                { return a.services }
func (a testApp) JwtCustomClaimsFunc() fs.JwtCustomClaimsFunc  { return nil }

func TestNew(t *testing.T) {
	app := &testApp{}
	s := services.New(app)

	assert.NotNil(t, s)
	assert.NotNil(t, s.File())
	assert.NotNil(t, s.Role())
	assert.NotNil(t, s.Schema())
	assert.NotNil(t, s.Content())
	assert.NotNil(t, s.Tool())
	assert.NotNil(t, s.Auth())
	assert.NotNil(t, s.Realtime())
}
func TestGet(t *testing.T) {
	// app does not implement ServicesProvider
	{
		_, err := services.Get[fileservice.FileService]("")
		assert.NotNil(t, err)
	}

	// app implements ServicesProvider
	{
		app := &testApp{}
		app.services = services.New(app)

		auth := utils.Must(services.Get[authservice.AuthService](app))
		assert.NotNil(t, auth)

		content := utils.Must(services.Get[contentservice.ContentService](app))
		assert.NotNil(t, content)

		file := utils.Must(services.Get[fileservice.FileService](app))
		assert.NotNil(t, file)

		realtime := utils.Must(services.Get[realtimeservice.RealtimeService](app))
		assert.NotNil(t, realtime)

		role := utils.Must(services.Get[roleservice.RoleService](app))
		assert.NotNil(t, role)

		schema := utils.Must(services.Get[schemaservice.SchemaService](app))
		assert.NotNil(t, schema)

		tool := utils.Must(services.Get[toolservice.ToolService](app))
		assert.NotNil(t, tool)
	}
}

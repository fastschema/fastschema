package plugins_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/logger"
	"github.com/fastschema/fastschema/pkg/entdbadapter"
	rr "github.com/fastschema/fastschema/pkg/restfulresolver"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
	"github.com/stretchr/testify/assert"
)

type testApp struct {
	config    *fs.Config
	sb        *schema.Builder
	db        db.Client
	resources *fs.ResourcesManager
	server    *rr.Server
	logger    logger.Logger
}

func newTestAppFromConfig(t *testing.T, config *fs.Config) *testApp {
	config.Dir = utils.Must(os.MkdirTemp("", "fastschema"))
	schemasDir := filepath.Join(config.Dir, "schemas")
	migrationsDir := filepath.Join(config.Dir, "migrations")

	assert.NoError(t, os.MkdirAll(schemasDir, 0755))
	assert.NoError(t, os.MkdirAll(migrationsDir, 0755))

	sb := utils.Must(schema.NewBuilderFromDir(schemasDir, fs.SystemSchemaTypes...))
	db := utils.Must(entdbadapter.NewTestClient(migrationsDir, sb, func() *db.Hooks {
		return config.Hooks.DBHooks
	}))

	app := &testApp{
		config:    config,
		sb:        sb,
		db:        db,
		resources: fs.NewResourcesManager(),
		logger:    logger.CreateMockLogger(true),
	}

	return app
}

func (a *testApp) Config() *fs.Config {
	return a.config
}

func (a *testApp) DB() db.Client {
	return a.db
}

func (a *testApp) Resources() *fs.ResourcesManager {
	return a.resources
}

func (a *testApp) Hooks() *fs.Hooks {
	return a.config.Hooks
}

func (a *testApp) Logger() logger.Logger {
	return a.logger
}

// Resolve hooks
func (a *testApp) OnPreResolve(middlewares ...fs.Middleware) {
	a.config.Hooks.PreResolve = append(
		a.config.Hooks.PreResolve,
		middlewares...,
	)
}

func (a *testApp) OnPostResolve(middlewares ...fs.Middleware) {
	a.config.Hooks.PostResolve = append(
		a.config.Hooks.PostResolve,
		middlewares...,
	)
}

// DB Query hooks
func (a *testApp) OnPreDBQuery(hooks ...db.PreDBQuery) {
	a.config.Hooks.DBHooks.PreDBQuery = append(
		a.config.Hooks.DBHooks.PreDBQuery,
		hooks...,
	)
}

func (a *testApp) OnPostDBQuery(hooks ...db.PostDBQuery) {
	a.config.Hooks.DBHooks.PostDBQuery = append(
		a.config.Hooks.DBHooks.PostDBQuery,
		hooks...,
	)
}

// DB Exec hooks
func (a *testApp) OnPreDBExec(hooks ...db.PreDBExec) {
	a.config.Hooks.DBHooks.PreDBExec = append(
		a.config.Hooks.DBHooks.PreDBExec,
		hooks...,
	)
}

func (a *testApp) OnPostDBExec(hooks ...db.PostDBExec) {
	a.config.Hooks.DBHooks.PostDBExec = append(
		a.config.Hooks.DBHooks.PostDBExec,
		hooks...,
	)
}

// DB Create hooks
func (a *testApp) OnPreDBCreate(hooks ...db.PreDBCreate) {
	a.config.Hooks.DBHooks.PreDBCreate = append(
		a.config.Hooks.DBHooks.PreDBCreate,
		hooks...,
	)
}

func (a *testApp) OnPostDBCreate(hooks ...db.PostDBCreate) {
	a.config.Hooks.DBHooks.PostDBCreate = append(
		a.config.Hooks.DBHooks.PostDBCreate,
		hooks...,
	)
}

// DB Update hooks
func (a *testApp) OnPreDBUpdate(hooks ...db.PreDBUpdate) {
	a.config.Hooks.DBHooks.PreDBUpdate = append(
		a.config.Hooks.DBHooks.PreDBUpdate,
		hooks...,
	)
}

func (a *testApp) OnPostDBUpdate(hooks ...db.PostDBUpdate) {
	a.config.Hooks.DBHooks.PostDBUpdate = append(
		a.config.Hooks.DBHooks.PostDBUpdate,
		hooks...,
	)
}

// DB Delete hooks
func (a *testApp) OnPreDBDelete(hooks ...db.PreDBDelete) {
	a.config.Hooks.DBHooks.PreDBDelete = append(
		a.config.Hooks.DBHooks.PreDBDelete,
		hooks...,
	)
}

func (a *testApp) OnPostDBDelete(hooks ...db.PostDBDelete) {
	a.config.Hooks.DBHooks.PostDBDelete = append(
		a.config.Hooks.DBHooks.PostDBDelete,
		hooks...,
	)
}

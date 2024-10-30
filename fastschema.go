package fastschema

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sync"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/logger"
	"github.com/fastschema/fastschema/pkg/openapi"
	rs "github.com/fastschema/fastschema/pkg/restfulresolver"
	"github.com/fastschema/fastschema/schema"
	"github.com/fatih/color"
)

type App struct {
	mu              sync.Mutex
	config          *fs.Config
	cwd             string
	dir             string
	envFile         string
	dataDir         string
	logDir          string
	publicDir       string
	schemasDir      string
	migrationDir    string
	schemaBuilder   *schema.Builder
	restResolver    *rs.RestfulResolver
	resources       *fs.ResourcesManager
	api             *fs.Resource
	roles           []*fs.Role
	disks           []fs.Disk
	defaultDisk     fs.Disk
	caches          []fs.Cache
	defaultCache    fs.Cache
	setupToken      string
	startupMessages []string
	statics         []*fs.StaticFs
	openAPISpec     []byte
	authProviders   map[string]fs.AuthProvider
}

func New(config *fs.Config) (_ *App, err error) {
	a := &App{
		config:        config.Clone(),
		disks:         []fs.Disk{},
		caches:        []fs.Cache{},
		roles:         []*fs.Role{},
		authProviders: map[string]fs.AuthProvider{},
	}

	if a.cwd, err = os.Getwd(); err != nil {
		return nil, err
	}

	if err := a.prepareConfig(); err != nil {
		return nil, err
	}

	if err = a.init(); err != nil {
		return nil, err
	}

	return a, nil
}

func (a *App) AddResource(resource *fs.Resource) {
	a.resources.Add(resource)
}

func (a *App) AddMiddlewares(middlewares ...fs.Middleware) {
	a.resources.Middlewares = append(
		a.resources.Middlewares,
		middlewares...,
	)
}

// Resolve hooks
func (a *App) OnPreResolve(middlewares ...fs.Middleware) {
	a.config.Hooks.PreResolve = append(
		a.config.Hooks.PreResolve,
		middlewares...,
	)
}

func (a *App) OnPostResolve(middlewares ...fs.Middleware) {
	a.config.Hooks.PostResolve = append(
		a.config.Hooks.PostResolve,
		middlewares...,
	)
}

// DB Query hooks
func (a *App) OnPreDBQuery(hooks ...db.PreDBQuery) {
	a.config.Hooks.DBHooks.PreDBQuery = append(
		a.config.Hooks.DBHooks.PreDBQuery,
		hooks...,
	)
}

func (a *App) OnPostDBQuery(hooks ...db.PostDBQuery) {
	a.config.Hooks.DBHooks.PostDBQuery = append(
		a.config.Hooks.DBHooks.PostDBQuery,
		hooks...,
	)
}

// DB Exec hooks
func (a *App) OnPreDBExec(hooks ...db.PreDBExec) {
	a.config.Hooks.DBHooks.PreDBExec = append(
		a.config.Hooks.DBHooks.PreDBExec,
		hooks...,
	)
}

func (a *App) OnPostDBExec(hooks ...db.PostDBExec) {
	a.config.Hooks.DBHooks.PostDBExec = append(
		a.config.Hooks.DBHooks.PostDBExec,
		hooks...,
	)
}

// DB Create hooks
func (a *App) OnPreDBCreate(hooks ...db.PreDBCreate) {
	a.config.Hooks.DBHooks.PreDBCreate = append(
		a.config.Hooks.DBHooks.PreDBCreate,
		hooks...,
	)
}

func (a *App) OnPostDBCreate(hooks ...db.PostDBCreate) {
	a.config.Hooks.DBHooks.PostDBCreate = append(
		a.config.Hooks.DBHooks.PostDBCreate,
		hooks...,
	)
}

// DB Update hooks
func (a *App) OnPreDBUpdate(hooks ...db.PreDBUpdate) {
	a.config.Hooks.DBHooks.PreDBUpdate = append(
		a.config.Hooks.DBHooks.PreDBUpdate,
		hooks...,
	)
}

func (a *App) OnPostDBUpdate(hooks ...db.PostDBUpdate) {
	a.config.Hooks.DBHooks.PostDBUpdate = append(
		a.config.Hooks.DBHooks.PostDBUpdate,
		hooks...,
	)
}

// DB Delete hooks
func (a *App) OnPreDBDelete(hooks ...db.PreDBDelete) {
	a.config.Hooks.DBHooks.PreDBDelete = append(
		a.config.Hooks.DBHooks.PreDBDelete,
		hooks...,
	)
}

func (a *App) OnPostDBDelete(hooks ...db.PostDBDelete) {
	a.config.Hooks.DBHooks.PostDBDelete = append(
		a.config.Hooks.DBHooks.PostDBDelete,
		hooks...,
	)
}

func (a *App) CWD() string {
	return a.cwd
}

func (a *App) Config() *fs.Config {
	return a.config
}

func (a *App) GetAuthProvider(name string) fs.AuthProvider {
	return a.authProviders[name]
}

func (a *App) Key() string {
	return a.config.AppKey
}

func (a *App) Dir() string {
	return a.dir
}

func (a *App) SchemaBuilder() *schema.Builder {
	return a.schemaBuilder
}

func (a *App) DB() db.Client {
	return a.config.DB
}

func (a *App) API() *fs.Resource {
	return a.api
}

func (a *App) Logger() logger.Logger {
	return a.config.Logger
}

func (a *App) Resources() *fs.ResourcesManager {
	return a.resources
}

func (a *App) Roles() (roles []*fs.Role, err error) {
	_, err = a.Cache().Get(context.Background(), "roles", &roles)
	if err != nil {
		return nil, err
	}

	return roles, nil
}

func (a *App) Hooks() *fs.Hooks {
	return a.config.Hooks
}

func (a *App) Disks() []fs.Disk {
	return a.disks
}

func (a *App) Disk(names ...string) fs.Disk {
	if len(names) == 0 {
		return a.defaultDisk
	}

	for _, disk := range a.disks {
		if disk.Name() == names[0] {
			return disk
		}
	}

	return nil
}

func (a *App) Caches() []fs.Cache {
	return a.caches
}

func (a *App) Cache(names ...string) fs.Cache {
	if len(names) == 0 {
		return a.defaultCache
	}

	for _, cache := range a.caches {
		if cache.Name() == names[0] {
			return cache
		}
	}

	return nil
}

func (a *App) Reload(ctx context.Context, migration *db.Migration) (err error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if err = a.createSchemaBuilder(); err != nil {
		return err
	}

	newDB, err := a.DB().Reload(ctx, a.schemaBuilder, migration, a.config.DBConfig.DisableForeignKeys)
	if err != nil {
		return err
	}

	if a.DB() != nil && a.DB().Close() != nil {
		return err
	}

	a.config.DB = newDB

	if _, err := a.CreateOpenAPISpec(true); err != nil {
		return err
	}

	return nil
}

// UpdateCache updates the application cache.
// It fetches all roles from the database and stores them in the cache.
func (a *App) UpdateCache(ctx context.Context, keys ...string) (err error) {
	// if keys is empty, update all cache such as role,...
	if len(keys) == 0 {
		return a.UpdateRoleCache(ctx)
	}

	for _, key := range keys {
		if key == "roles" {
			err = a.UpdateRoleCache(ctx)

			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (a *App) UpdateRoleCache(ctx context.Context) (err error) {
	roles, err := db.Builder[*fs.Role](a.DB()).Select(
		"id",
		"name",
		"description",
		"root",
		"permissions",
		schema.FieldCreatedAt,
		schema.FieldUpdatedAt,
		schema.FieldDeletedAt,
	).Get(ctx)

	if err != nil {
		return err
	}

	return a.Cache().Set(ctx, "roles", roles)
}

// CreateOpenAPISpec generates the openapi spec for the app.
func (a *App) CreateOpenAPISpec(overrides ...bool) ([]byte, error) {
	overrides = append(overrides, false)

	if a.openAPISpec == nil || overrides[0] {
		s, err := openapi.NewSpec(&openapi.OpenAPISpecConfig{
			BaseURL:       a.config.BaseURL,
			Resources:     a.Resources(),
			SchemaBuilder: a.schemaBuilder,
		})

		if err != nil {
			return nil, err
		}

		a.openAPISpec = s.Spec()
	}

	return a.openAPISpec, nil
}

func (a *App) Start() error {
	addr := fmt.Sprintf(":%s", a.config.Port)
	if err := a.resources.Init(); err != nil {
		return err
	}

	if !a.config.HideResourcesInfo {
		a.resources.Print()
	}

	a.restResolver = rs.NewRestfulResolver(&rs.ResolverConfig{
		ResourceManager: a.resources,
		Logger:          a.Logger(),
		StaticFSs:       a.statics,
	})

	fmt.Printf("\n")
	for _, msg := range a.startupMessages {
		color.Green("> %s", msg)
	}
	fmt.Printf("\n")

	return a.restResolver.Start(addr)
}

func (a *App) HTTPAdaptor() (http.HandlerFunc, error) {
	if err := a.resources.Init(); err != nil {
		return nil, err
	}

	if !a.config.HideResourcesInfo {
		a.resources.Print()
	}

	a.restResolver = rs.NewRestfulResolver(&rs.ResolverConfig{
		ResourceManager: a.resources,
		Logger:          a.Logger(),
		StaticFSs:       a.statics,
	})

	fmt.Printf("\n")
	for _, msg := range a.startupMessages {
		color.Green("> %s", msg)
	}
	fmt.Printf("\n")

	return a.restResolver.HTTPAdaptor()
}

func (a *App) Shutdown() error {
	if a.DB() != nil {
		if err := a.DB().Close(); err != nil {
			return err
		}
	}

	if a.restResolver != nil {
		if err := a.restResolver.Shutdown(); err != nil {
			return err
		}
	}

	return nil
}

package fastschema

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/logger"
	"github.com/fastschema/fastschema/pkg/openapi"
	"github.com/fastschema/fastschema/pkg/restfulresolver"
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
	restResolver    *restfulresolver.RestfulResolver
	resources       *fs.ResourcesManager
	api             *fs.Resource
	hooks           *fs.Hooks
	roles           []*fs.Role
	disks           []fs.Disk
	defaultDisk     fs.Disk
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
		roles:         []*fs.Role{},
		authProviders: map[string]fs.AuthProvider{},
		hooks: &fs.Hooks{
			DBHooks: &db.Hooks{},
		},
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
	a.resources.Middlewares = append(a.resources.Middlewares, middlewares...)
}

func (a *App) OnPreResolve(middlewares ...fs.Middleware) {
	a.hooks.PreResolve = append(a.hooks.PreResolve, middlewares...)
}

func (a *App) OnPostResolve(middlewares ...fs.Middleware) {
	a.hooks.PostResolve = append(a.hooks.PostResolve, middlewares...)
}

func (a *App) OnPostDBGet(hooks ...db.PostDBGet) {
	a.hooks.DBHooks.PostDBGet = append(a.hooks.DBHooks.PostDBGet, hooks...)
}

func (a *App) OnPostDBCreate(hooks ...db.PostDBCreate) {
	a.hooks.DBHooks.PostDBCreate = append(a.hooks.DBHooks.PostDBCreate, hooks...)
}

func (a *App) OnPostDBUpdate(hooks ...db.PostDBUpdate) {
	a.hooks.DBHooks.PostDBUpdate = append(a.hooks.DBHooks.PostDBUpdate, hooks...)
}

func (a *App) OnPostDBDelete(hooks ...db.PostDBDelete) {
	a.hooks.DBHooks.PostDBDelete = append(a.hooks.DBHooks.PostDBDelete, hooks...)
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

func (a *App) Roles() []*fs.Role {
	return a.roles
}

func (a *App) Hooks() *fs.Hooks {
	return a.hooks
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
func (a *App) UpdateCache(ctx context.Context) (err error) {
	if a.roles, err = db.Query[*fs.Role](a.DB()).Select(
		"id",
		"name",
		"description",
		"root",
		"permissions",
		schema.FieldCreatedAt,
		schema.FieldUpdatedAt,
		schema.FieldDeletedAt,
	).Get(ctx); err != nil {
		return err
	}

	return nil
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

	a.restResolver = restfulresolver.NewRestfulResolver(&restfulresolver.ResolverConfig{
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

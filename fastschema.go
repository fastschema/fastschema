package fastschema

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/logger"
	"github.com/fastschema/fastschema/pkg/auth"
	"github.com/fastschema/fastschema/pkg/entdbadapter"
	"github.com/fastschema/fastschema/pkg/errors"
	"github.com/fastschema/fastschema/pkg/openapi"
	"github.com/fastschema/fastschema/pkg/rclonefs"
	"github.com/fastschema/fastschema/pkg/restfulresolver"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/pkg/zaplogger"
	"github.com/fastschema/fastschema/schema"
	as "github.com/fastschema/fastschema/services/auth"
	cs "github.com/fastschema/fastschema/services/content"
	ms "github.com/fastschema/fastschema/services/file"
	rs "github.com/fastschema/fastschema/services/role"
	ss "github.com/fastschema/fastschema/services/schema"
	ts "github.com/fastschema/fastschema/services/tool"
	us "github.com/fastschema/fastschema/services/user"
	"github.com/fatih/color"
	"github.com/joho/godotenv"
)

//go:embed all:dash/*
var embedDashStatic embed.FS
var createAuthProvidersFn = map[string]fs.CreateAuthProviderFunc{
	"github": auth.NewGithubAuthProvider,
	"google": auth.NewGoogleAuthProvider,
}

type App struct {
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

func (a *App) CWD() string {
	return a.cwd
}

func (a *App) init() (err error) {
	if err = a.getDefaultDisk(); err != nil {
		return err
	}

	if err = a.getDefaultLogger(); err != nil {
		return err
	}

	if err := a.getAuthProviders(); err != nil {
		return err
	}

	if err = a.createSchemaBuilder(); err != nil {
		return err
	}

	if err = a.createResources(); err != nil {
		return err
	}

	if err = a.getDefaultDBClient(); err != nil {
		return err
	}

	// if a local disk has a public path, then add it to the statics
	for _, disk := range a.disks {
		publicPath := disk.LocalPublicPath()

		if publicPath != "" {
			a.startupMessages = append(
				a.startupMessages,
				fmt.Sprintf("Serving files from disk [%s:%s] at %s", disk.Name(), publicPath, disk.Root()),
			)

			a.statics = append(a.statics, &fs.StaticFs{
				BasePath: publicPath,
				Root:     http.Dir(disk.Root()),
			})
		}
	}

	setupToken, err := a.SetupToken(context.Background())
	if err != nil {
		return err
	}

	if setupToken != "" {
		type setupData struct {
			Token    string `json:"token"`
			Username string `json:"username"`
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		a.api.Add(fs.NewResource("setup", func(c fs.Context, setupData *setupData) (bool, error) {
			if setupToken == "" {
				return false, errors.BadRequest("Setup token is not available")
			}

			if setupData == nil || setupData.Token != setupToken {
				return false, errors.Forbidden("Invalid setup data or token")
			}

			if err := ts.Setup(
				c.Context(),
				a.DB(),
				a.Logger(),
				setupData.Username, setupData.Email, setupData.Password,
			); err != nil {
				return false, err
			}

			if err := a.UpdateCache(c.Context()); err != nil {
				return false, err
			}

			setupToken = ""
			a.setupToken = ""

			return true, nil
		}, &fs.Meta{
			Post:   "/setup",
			Public: true,
		}))

		setupURL := fmt.Sprintf(
			"%s/setup/?token=%s\033[0m",
			a.config.DashURL,
			setupToken,
		)

		a.startupMessages = append(a.startupMessages, fmt.Sprintf(
			"Visit the following URL to setup the app: %s",
			setupURL,
		))
	}

	a.statics = append(a.statics, &fs.StaticFs{
		BasePath:   "/" + a.config.DashBaseName,
		Root:       http.FS(embedDashStatic),
		PathPrefix: "dash",
	})

	return nil
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

func (a *App) Reload(ctx context.Context, migration *db.Migration) (err error) {
	if a.DB() != nil {
		if err = a.DB().Close(); err != nil {
			return err
		}
	}

	if err = a.createSchemaBuilder(); err != nil {
		return err
	}

	newDB, err := a.DB().Reload(ctx, a.schemaBuilder, migration)
	if err != nil {
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

	a.restResolver = restfulresolver.NewRestfulResolver(a.resources, a.Logger(), a.statics...)

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

func (a *App) SetupToken(ctx context.Context) (string, error) {
	// If there is no roles and users, then the app is not setup
	// we need to setup the app.
	// Generate a random token and return it to enable the setup.
	// If there are roles or users, then the app is already setup.
	// Return an empty string to disable the setup.
	needSetup, err := a.needSetup(ctx)
	if err != nil {
		return "", err
	}

	if !needSetup {
		return "", nil
	}

	if a.setupToken == "" {
		a.setupToken = utils.RandomString(32)
	}

	return a.setupToken, nil
}

func (a *App) needSetup(ctx context.Context) (bool, error) {
	// If there is no roles and users, then the app is not setup
	// we need to setup the app.
	var err error
	var userCount int
	var roleCount int
	countOption := &db.CountOption{
		Column: "id",
		Unique: true,
	}

	if userCount, err = db.Query[*fs.User](a.DB()).Count(ctx, countOption); err != nil {
		return false, err
	}

	if roleCount, err = db.Query[*fs.Role](a.DB()).Count(ctx, countOption); err != nil {
		return false, err
	}

	return userCount == 0 && roleCount == 0, nil
}

func (a *App) getDefaultDBClient() (err error) {
	if a.DB() != nil {
		return nil
	}

	if a.config.DBConfig == nil {
		a.config.DBConfig = &db.Config{
			Driver:       utils.Env("DB_DRIVER", "sqlite"),
			Name:         utils.Env("DB_NAME"),
			User:         utils.Env("DB_USER"),
			Pass:         utils.Env("DB_PASS"),
			Host:         utils.Env("DB_HOST", "localhost"),
			Port:         utils.Env("DB_PORT"),
			LogQueries:   utils.Env("DB_LOGGING", "false") == "true",
			Logger:       a.Logger(),
			MigrationDir: a.migrationDir,
			Hooks: func() *db.Hooks {
				return &db.Hooks{
					PostDBGet: a.hooks.DBHooks.PostDBGet,
				}
			},
		}
	}

	if !utils.Contains(db.SupportDrivers, a.config.DBConfig.Driver) {
		return fmt.Errorf("unsupported database driver: %s", a.config.DBConfig.Driver)
	}

	if a.config.DBConfig.MigrationDir == "" {
		a.config.DBConfig.MigrationDir = a.migrationDir
	}

	// If driver is sqlite and the DB_NAME (file path) is not set,
	// Set the DB_NAME to the default sqlite db file path.
	if a.config.DBConfig.Driver == "sqlite" && a.config.DBConfig.Name == "" {
		a.config.DBConfig.Name = path.Join(a.dataDir, "fastschema.db")
		a.startupMessages = append(
			a.startupMessages,
			fmt.Sprintf("Using default sqlite db file: %s", a.config.DBConfig.Name),
		)
	}

	if a.config.DB, err = entdbadapter.NewClient(a.config.DBConfig, a.schemaBuilder); err != nil {
		return err
	}

	if err := a.UpdateCache(context.Background()); err != nil {
		return err
	}

	return nil
}

func (a *App) getDefaultLogger() (err error) {
	if a.config.Logger == nil {
		if a.config.LoggerConfig == nil {
			a.config.LoggerConfig = &logger.Config{
				Development: utils.Env("APP_ENV", "development") == "development",
				LogFile:     path.Join(a.logDir, "app.log"),
			}
		}
		a.config.Logger, err = zaplogger.NewZapLogger(a.config.LoggerConfig)
	}

	return err
}

func (a *App) getDefaultDisk() (err error) {
	if a.config.StorageConfig == nil {
		a.config.StorageConfig = &fs.StorageConfig{}
	}

	defaultDiskName := a.config.StorageConfig.DefaultDisk
	if defaultDiskName == "" {
		defaultDiskName = utils.Env("STORAGE_DEFAULT_DISK", "")
	}

	storageDisksConfig := a.config.StorageConfig.DisksConfig
	if utils.Env("STORAGE_DISKS") != "" && storageDisksConfig == nil {
		if err := json.Unmarshal([]byte(utils.Env("STORAGE_DISKS")), &storageDisksConfig); err != nil {
			return err
		}
	}

	// if threre is no disk config, add a default disk
	if storageDisksConfig == nil {
		if defaultDiskName == "" {
			defaultDiskName = "public"
		}
		storageDisksConfig = []*fs.DiskConfig{{
			Name:       "public",
			Driver:     "local",
			PublicPath: "/files",
			BaseURL:    fmt.Sprintf("%s/files", a.config.BaseURL),
			Root:       a.publicDir,
		}}
	}

	if a.disks, err = rclonefs.NewFromConfig(storageDisksConfig, a.dataDir); err != nil {
		return err
	}

	foundDefaultDisk := false
	for _, disk := range a.disks {
		if disk.Name() == defaultDiskName {
			a.defaultDisk = disk
			foundDefaultDisk = true
			break
		}
	}

	if defaultDiskName != "" && !foundDefaultDisk {
		return fmt.Errorf("default disk [%s] not found", defaultDiskName)
	}

	if a.defaultDisk == nil && len(a.disks) > 0 {
		a.defaultDisk = a.disks[0]
	}

	return nil
}

func (a *App) createSchemaBuilder() (err error) {
	if a.schemaBuilder, err = schema.NewBuilderFromDir(
		a.schemasDir,
		append(fs.SystemSchemaTypes, a.config.SystemSchemas...)...,
	); err != nil {
		return err
	}

	return nil
}

func (a *App) getAuthProviders() (err error) {
	if a.config.AuthConfig == nil {
		return nil
	}

	for name, config := range a.config.AuthConfig.Providers {
		if !utils.Contains(a.config.AuthConfig.EnabledProviders, name) {
			continue
		}

		redirectURL := fmt.Sprintf("%s/%s/auth/%s/callback", a.config.BaseURL, a.config.APIBaseName, name)
		createProviderFn, ok := createAuthProvidersFn[name]
		if !ok {
			return fmt.Errorf("auth provider [%s] is not supported", name)
		}

		provider, err := createProviderFn(config, redirectURL)
		if err != nil {
			return err
		}

		a.authProviders[name] = provider
	}

	return nil
}

func (a *App) createResources() error {
	userService := us.New(a)
	roleService := rs.New(a)
	fileService := ms.New(a)
	schemaService := ss.New(a)
	contentService := cs.New(a)
	toolService := ts.New(a)
	authService := as.New(a)

	a.hooks.DBHooks.PostDBGet = append(a.hooks.DBHooks.PostDBGet, fileService.FileListHook)
	a.hooks.PreResolve = append(a.hooks.PreResolve, roleService.Authorize)

	a.resources = fs.NewResourcesManager()
	a.resources.Middlewares = append(a.resources.Middlewares, roleService.ParseUser)
	a.resources.Hooks = func() *fs.Hooks {
		return a.hooks
	}

	var createArg = func(t fs.ArgType, desc string) fs.Arg {
		return fs.Arg{Type: t, Required: true, Description: desc}
	}

	a.api = a.resources.Group(a.config.APIBaseName)
	a.api.Group("user").
		Add(fs.NewResource("logout", userService.Logout, &fs.Meta{
			Post:   "/logout",
			Public: true,
		})).
		Add(fs.NewResource("me", userService.Me, &fs.Meta{Public: true})).
		Add(fs.NewResource("login", userService.Login, &fs.Meta{
			Post:   "/login",
			Public: true,
		}))

	if len(a.authProviders) > 0 {
		a.api.Group("auth", &fs.Meta{
			Prefix: "/auth/:provider",
			Args: fs.Args{
				"provider": {
					Required:    true,
					Type:        fs.TypeString,
					Description: "The auth provider name",
					Example:     "google",
				},
			},
		}).Add(
			fs.Get("login", authService.Login, &fs.Meta{Public: true}),
			fs.Get("callback", authService.Callback, &fs.Meta{Public: true}),
		)
	}

	a.api.Group("schema").
		Add(fs.NewResource("list", schemaService.List, &fs.Meta{Get: "/"})).
		Add(fs.NewResource("create", schemaService.Create, &fs.Meta{Post: "/"})).
		Add(fs.NewResource("detail", schemaService.Detail, &fs.Meta{
			Get:  "/:name",
			Args: fs.Args{"name": createArg(fs.TypeString, "The schema name")},
		})).
		Add(fs.NewResource("update", schemaService.Update, &fs.Meta{
			Put:  "/:name",
			Args: fs.Args{"name": createArg(fs.TypeString, "The schema name")},
		})).
		Add(fs.NewResource("delete", schemaService.Delete, &fs.Meta{
			Delete: "/:name",
			Args:   fs.Args{"name": createArg(fs.TypeString, "The schema name")},
		}))

	a.api.Group("content", &fs.Meta{
		Prefix: "/content/:schema",
		Args: fs.Args{
			"schema": {
				Required:    true,
				Type:        fs.TypeString,
				Description: "The schema name",
			},
		},
	}).
		Add(fs.NewResource("list", contentService.List, &fs.Meta{Get: "/"})).
		Add(fs.NewResource("detail", contentService.Detail, &fs.Meta{
			Get:  "/:id",
			Args: fs.Args{"id": createArg(fs.TypeUint64, "The content ID")},
		})).
		Add(fs.NewResource("create", contentService.Create, &fs.Meta{Post: "/"})).
		Add(fs.NewResource("update", contentService.Update, &fs.Meta{
			Put:  "/:id",
			Args: fs.Args{"id": createArg(fs.TypeUint64, "The content ID")},
		})).
		Add(fs.NewResource("delete", contentService.Delete, &fs.Meta{
			Delete: "/:id",
			Args:   fs.Args{"id": createArg(fs.TypeUint64, "The content ID")},
		}))

	a.api.Group("role").
		Add(fs.NewResource("list", roleService.List, &fs.Meta{Get: "/"})).
		Add(fs.NewResource("resources", roleService.ResourcesList, &fs.Meta{
			Get: "/resources",
		})).
		Add(fs.NewResource("detail", roleService.Detail, &fs.Meta{
			Get:  "/:id",
			Args: fs.Args{"id": createArg(fs.TypeUint64, "The role ID")},
		})).
		Add(fs.NewResource("create", roleService.Create, &fs.Meta{Post: "/"})).
		Add(fs.NewResource("update", roleService.Update, &fs.Meta{
			Put:  "/:id",
			Args: fs.Args{"id": createArg(fs.TypeUint64, "The role ID")},
		})).
		Add(fs.NewResource("delete", roleService.Delete, &fs.Meta{
			Delete: "/:id",
			Args:   fs.Args{"id": createArg(fs.TypeUint64, "The role ID")},
		}))

	a.api.Group("file").
		Add(fs.NewResource("upload", fileService.Upload, &fs.Meta{Post: "/upload"})).
		Add(fs.NewResource("delete", fileService.Delete, &fs.Meta{Delete: "/"}))

	a.api.Group("tool").
		Add(fs.NewResource("stats", toolService.Stats, &fs.Meta{
			Get:    "/stats",
			Public: true,
		}))

	a.resources.Group("docs").
		Add(fs.NewResource("spec", func(c fs.Context, _ any) (any, error) {
			return a.CreateOpenAPISpec()
		}, &fs.Meta{Get: "/openapi.json"})).
		Add(fs.NewResource("viewer", func(c fs.Context, _ any) (any, error) {
			header := make(http.Header)
			header.Set("Content-Type", "text/html")

			return &fs.HTTPResponse{
				StatusCode: http.StatusOK,
				Header:     header,
				Body: []byte(utils.CreateSwaggerUIPage(
					a.config.BaseURL + "/docs/openapi.json",
				)),
			}, nil
		}, &fs.Meta{Get: "/"}))

	return nil
}

func (a *App) getAppDir() {
	defer func() {
		a.startupMessages = append(a.startupMessages, fmt.Sprintf("Using app directory: %s", a.dir))
	}()

	if a.config.Dir == "" {
		a.dir = a.cwd
		return
	}

	if strings.HasPrefix(a.config.Dir, "/") {
		a.dir = a.config.Dir
		return
	}

	a.dir = path.Join(a.cwd, a.config.Dir)
}

func (a *App) prepareConfig() (err error) {
	a.getAppDir()
	a.dataDir = path.Join(a.dir, "data")
	a.logDir = path.Join(a.dataDir, "logs")
	a.publicDir = path.Join(a.dataDir, "public")
	a.schemasDir = path.Join(a.dataDir, "schemas")
	a.migrationDir = path.Join(a.dataDir, "migrations")
	envFile := path.Join(a.dataDir, ".env")

	if err = utils.MkDirs(
		a.logDir,
		a.publicDir,
		a.schemasDir,
		a.migrationDir,
	); err != nil {
		return err
	}

	if utils.IsFileExists(envFile) {
		a.envFile = envFile
		if err := godotenv.Load(envFile); err != nil {
			return err
		}
	}

	if a.config.AppKey == "" {
		a.config.AppKey = utils.Env("APP_KEY")
	}

	if a.config.Port == "" {
		a.config.Port = utils.Env("APP_PORT", "8000")
	}

	if a.config.BaseURL == "" {
		a.config.BaseURL = utils.Env("APP_BASE_URL")
	}

	if a.config.DashURL == "" {
		a.config.DashURL = utils.Env("APP_DASH_URL")
	}

	if a.config.APIBaseName == "" {
		a.config.APIBaseName = utils.Env("APP_API_BASE_NAME", "api")
	}

	if a.config.DashBaseName == "" {
		a.config.DashBaseName = utils.Env("APP_DASH_BASE_NAME", "dash")
	}

	if a.config.BaseURL == "" {
		a.config.BaseURL = fmt.Sprintf("http://localhost:%s", a.config.Port)
	}

	if a.config.DashURL == "" {
		a.config.DashURL = fmt.Sprintf("%s/%s", a.config.BaseURL, a.config.DashBaseName)
	}

	if a.config.AppKey == "" {
		a.config.AppKey = utils.RandomString(32)
		if err := utils.AppendFile(
			envFile,
			fmt.Sprintf("APP_KEY=%s\n", a.config.AppKey),
		); err != nil {
			return err
		}

		a.startupMessages = append(
			a.startupMessages,
			fmt.Sprintf("APP_KEY is not set. A new key is generated and saved to %s", envFile),
		)
	}

	if a.config.AuthConfig == nil && utils.Env("AUTH") != "" {
		fmt.Println(utils.Env("AUTH"))
		if err := json.Unmarshal([]byte(utils.Env("AUTH")), &a.config.AuthConfig); err != nil {
			return err
		}
	}

	return nil
}

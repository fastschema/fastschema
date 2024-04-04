package fastschema

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/fastschema/fastschema/app"
	"github.com/fastschema/fastschema/cmd"
	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/logger"
	"github.com/fastschema/fastschema/pkg/entdbadapter"
	"github.com/fastschema/fastschema/pkg/errors"
	"github.com/fastschema/fastschema/pkg/rclonefs"
	"github.com/fastschema/fastschema/pkg/restresolver"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/pkg/zaplogger"
	"github.com/fastschema/fastschema/schema"
	cs "github.com/fastschema/fastschema/services/content"
	ms "github.com/fastschema/fastschema/services/media"
	rs "github.com/fastschema/fastschema/services/role"
	ss "github.com/fastschema/fastschema/services/schema"
	ts "github.com/fastschema/fastschema/services/tool"
	us "github.com/fastschema/fastschema/services/user"
	"github.com/joho/godotenv"
)

//go:embed all:dash/*
var embedDashStatic embed.FS

type AppConfig struct {
	Dir          string
	AppKey       string
	Port         string
	DashURL      string
	APIBaseName  string
	DashBaseName string
	Logger       logger.Logger
	DB           db.Client
	Storage      *app.StorageConfig
}

type App struct {
	config        *AppConfig
	dir           string
	schemasDir    string
	migrationDir  string
	db            db.Client
	logger        logger.Logger
	schemaBuilder *schema.Builder
	resources     *app.ResourcesManager
	api           *app.Resource
	hooks         *app.Hooks
	roles         []*app.Role
	disks         []app.Disk
	defaultDisk   app.Disk
	setupToken    string
}

func New(config *AppConfig) (_ *App, err error) {
	appDir, err := getAppDir(config.Dir)
	if err != nil {
		return nil, err
	}

	if err := parseEnvFile(path.Join(appDir, ".env")); err != nil {
		return nil, err
	}

	if config.AppKey == "" {
		config.AppKey = utils.Env("APP_KEY")
	}

	if config.Port == "" {
		config.Port = utils.Env("APP_PORT", "3000")
	}

	if config.DashURL == "" {
		config.DashURL = utils.Env("APP_DASH_URL")
	}

	if config.AppKey == "" {
		return nil, fmt.Errorf("APP_KEY is required. Please check the environment variables")
	}

	if config.APIBaseName == "" {
		config.APIBaseName = utils.Env("APP_API_BASE_NAME", "api")
	}

	if config.DashBaseName == "" {
		config.DashBaseName = utils.Env("APP_DASH_BASE_NAME", "dash")
	}

	a := &App{
		dir:          appDir,
		config:       config,
		schemasDir:   path.Join(appDir, "private/schemas"),
		migrationDir: path.Join(appDir, "private/migrations"),
		logger:       config.Logger,
		db:           config.DB,
		disks:        []app.Disk{},
		roles:        []*app.Role{},
		hooks: &app.Hooks{
			BeforeResolve: []app.ResolveHook{},
			AfterResolve:  []app.ResolveHook{},
			ContentList:   []db.AfterDBContentListHook{},
		},
	}

	if err = a.init(); err != nil {
		return nil, err
	}

	return a, nil
}

func (a *App) init() (err error) {
	if err = utils.MkDirs(
		path.Join(a.dir, "private/logs"),
		path.Join(a.dir, "public"),
		a.schemasDir,
		a.migrationDir,
	); err != nil {
		return err
	}

	if err = a.getDefaultDisk(); err != nil {
		return err
	}

	if err = a.getDefaultLogger(); err != nil {
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

	return nil
}

func (a *App) Key() string {
	return a.config.AppKey
}

func (a *App) Reload(migration *db.Migration) (err error) {
	if a.db != nil {
		if err = a.db.Close(); err != nil {
			return err
		}
	}

	if err = a.createSchemaBuilder(); err != nil {
		return err
	}

	newDB, err := a.db.Reload(a.schemaBuilder, migration)
	if err != nil {
		return err
	}

	a.db = newDB

	return nil
}

func (a *App) Start() {
	addr := fmt.Sprintf(":%s", a.config.Port)
	setupToken, err := a.SetupToken()
	if err != nil {
		log.Fatal(err)
	}

	if setupToken != "" {
		type setupData struct {
			Token    string `json:"token"`
			Username string `json:"username"`
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		a.api.Add(app.NewResource("setup", func(c app.Context, setupData *setupData) (bool, error) {
			if setupToken == "" {
				return false, errors.BadRequest("Setup token is not available")
			}

			if setupData == nil {
				return false, errors.BadRequest("Invalid setup data")
			}

			if setupData.Token != setupToken {
				return false, errors.Unauthorized("Invalid setup token")
			}

			if err := cmd.Setup(a, setupData.Username, setupData.Email, setupData.Password); err != nil {
				return false, err
			}

			if err := a.UpdateCache(); err != nil {
				return false, err
			}

			setupToken = ""
			a.setupToken = ""

			return true, nil
		}, app.Meta{app.POST: "/setup"}, true))
	}

	if err := a.resources.Init(); err != nil {
		log.Fatal(err)
	}

	a.resources.Print()

	if setupToken != "" {
		setupURL := fmt.Sprintf(
			"%s/setup/?token=%s\033[0m",
			a.config.DashURL,
			setupToken,
		)
		fmt.Printf(
			"\n\033[32mYour app is not setup yet. Please visit the following URL to setup the app:\n%s",
			setupURL,
		)
	}

	restResolver := restresolver.NewRestResolver(a.resources)
	log.Fatal(restResolver.Start(addr, a.logger))
}

func (a *App) SchemaBuilder() *schema.Builder {
	return a.schemaBuilder
}

func (a *App) DB() db.Client {
	return a.db
}

func (a *App) Logger() logger.Logger {
	return a.logger
}

func (a *App) Resources() *app.ResourcesManager {
	return a.resources
}

func (a *App) Roles() []*app.Role {
	return a.roles
}

func (a *App) Hooks() *app.Hooks {
	return a.hooks
}

func (a *App) Disk(names ...string) app.Disk {
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

func (a *App) AddResource(resource *app.Resource) {
	a.resources.Add(resource)
}

func (a *App) AddMiddlewares(middlewares ...app.Middleware) {
	a.resources.Middlewares = append(a.resources.Middlewares, middlewares...)
}

func (a *App) OnBeforeResolve(middlewares ...app.Middleware) {
	a.hooks.BeforeResolve = append(a.hooks.BeforeResolve, middlewares...)
}

func (a *App) OnAfterResolve(middlewares ...app.Middleware) {
	a.hooks.AfterResolve = append(a.hooks.AfterResolve, middlewares...)
}

func (a *App) OnAfterDBContentList(hook db.AfterDBContentListHook) {
	a.hooks.ContentList = append(a.hooks.ContentList, hook)
}

func (a *App) GetRolesFromIDs(ids []uint64) []*app.Role {
	result := []*app.Role{}

	for _, role := range a.Roles() {
		for _, id := range ids {
			if role.ID == id {
				result = append(result, role)
			}
		}
	}

	return result
}

func (a *App) GetRoleDetail(roleID uint64) *app.Role {
	for _, role := range a.Roles() {
		if role.ID == roleID {
			return role
		}
	}

	return &app.Role{
		ID:          roleID,
		Permissions: []*app.Permission{},
	}
}

func (a *App) GetRolePermission(roleID uint64, action string) *app.Permission {
	rolePermissions := a.GetRoleDetail(roleID)

	for _, permission := range rolePermissions.Permissions {
		if permission.Resource == action {
			return permission
		}
	}

	return &app.Permission{}
}

func (a *App) UpdateCache() error {
	a.roles = []*app.Role{}
	roleModel, err := a.DB().Model("role")
	if err != nil {
		return err
	}

	roles, err := roleModel.Query().Select(
		"id",
		"name",
		"description",
		"root",
		"permissions",
		schema.FieldCreatedAt,
		schema.FieldUpdatedAt,
		schema.FieldDeletedAt,
	).Get(context.Background())

	if err != nil {
		return err
	}

	for _, r := range roles {
		role := app.EntityToRole(r)
		a.roles = append(a.roles, role)
	}

	return nil
}

func (a *App) SetupToken() (string, error) {
	// If there is no roles and users, then the app is not setup
	// we need to setup the app.
	// Generate a random token and return it to enable the setup.
	// If there are roles or users, then the app is already setup.
	// Return an empty string to disable the setup.
	needSetup, err := a.needSetup()
	if err != nil {
		return "", err
	}

	if a.setupToken == "" && needSetup {
		a.setupToken = utils.RandomString(32)
	}

	return a.setupToken, nil
}

func (a *App) needSetup() (bool, error) {
	// If there is no roles and users, then the app is not setup
	// we need to setup the app.
	var err error
	var userCount int
	var roleCount int
	ctx := context.Background()
	countOption := &db.CountOption{
		Column: "id",
		Unique: true,
	}

	userModel, err := a.DB().Model("user")
	if err != nil {
		return false, err
	}

	roleModel, err := a.DB().Model("role")
	if err != nil {
		return false, err
	}

	if userCount, err = userModel.Query().Count(countOption, ctx); err != nil {
		return false, err
	}

	if roleCount, err = roleModel.Query().Count(countOption, ctx); err != nil {
		return false, err
	}

	return userCount == 0 && roleCount == 0, nil
}

func (a *App) getDefaultDBClient() (err error) {
	if a.db != nil {
		return nil
	}

	dbConfig := &db.DBConfig{
		Driver:       utils.Env("DB_DRIVER"),
		Name:         utils.Env("DB_NAME"),
		User:         utils.Env("DB_USER"),
		Pass:         utils.Env("DB_PASS"),
		Host:         utils.Env("DB_HOST"),
		Port:         utils.Env("DB_PORT"),
		LogQueries:   utils.Env("DB_LOGGING") == "true",
		Logger:       a.logger,
		MigrationDir: a.migrationDir,
		Hooks: &db.Hooks{
			AfterDBContentList: a.hooks.ContentList,
		},
	}

	if dbConfig.Driver == "" {
		return fmt.Errorf("DB_DRIVER is required. Pleas check the environment variables")
	}

	if dbConfig.Name == "" {
		return fmt.Errorf("DB_NAME is required. Pleas check the environment variables")
	}

	if dbConfig.User == "" {
		return fmt.Errorf("DB_USER is required. Pleas check the environment variables")
	}

	if a.db, err = entdbadapter.NewClient(dbConfig, a.schemaBuilder); err != nil {
		return err
	}

	if err := a.UpdateCache(); err != nil {
		return err
	}

	return nil
}

func (a *App) getDefaultLogger() (err error) {
	if a.config.Logger == nil {
		a.logger, err = zaplogger.NewZapLogger(&zaplogger.ZapConfig{
			Development: true,
			LogFile:     path.Join(a.dir, "private/logs/app.log"),
		})
	}

	return err
}

func (a *App) getDefaultDisk() error {
	if a.config.Storage == nil {
		a.config.Storage = &app.StorageConfig{}
	}

	defaultDiskName := a.config.Storage.DefaultDisk
	if defaultDiskName == "" {
		defaultDiskName = utils.Env("STORAGE_DEFAULT_DISK")
	}

	storageDisksConfig := a.config.Storage.DisksConfig
	if utils.Env("STORAGE_DISKS") != "" && storageDisksConfig == nil {
		if err := json.Unmarshal([]byte(utils.Env("STORAGE_DISKS")), &storageDisksConfig); err != nil {
			return err
		}
	}

	a.disks = rclonefs.NewFromConfig(storageDisksConfig, a.dir)

	if defaultDiskName == "" && len(a.disks) > 0 {
		a.defaultDisk = a.disks[0]
		return nil
	}

	for _, disk := range a.disks {
		if disk.Name() == defaultDiskName {
			a.defaultDisk = disk
			break
		}
	}

	return nil
}

func (a *App) createSchemaBuilder() (err error) {
	if a.schemaBuilder, err = schema.NewBuilderFromDir(a.schemasDir); err != nil {
		return err
	}

	return nil
}

func (a *App) createResources() error {
	userService := us.NewUserService(a)
	roleService := rs.NewRoleService(a)
	mediaService := ms.NewMediaService(a)
	schemaService := ss.NewSchemaService(a)
	contentService := cs.NewContentService(a)
	toolService := ts.NewToolService(a)

	a.resources = app.NewResourcesManager()
	a.resources.RegisterStaticResources(&app.StaticResourceConfig{
		Root:       http.FS(embedDashStatic),
		BasePath:   "/" + a.config.DashBaseName,
		PathPrefix: "dash",
	})
	a.resources.Middlewares = append(a.resources.Middlewares, roleService.ParseUser)
	a.resources.BeforeResolveHooks = append(a.resources.BeforeResolveHooks, roleService.Authorize)
	a.resources.BeforeResolveHooks = append(a.resources.BeforeResolveHooks, a.hooks.BeforeResolve...)
	a.resources.AfterResolveHooks = append(a.resources.AfterResolveHooks, a.hooks.AfterResolve...)

	a.OnAfterDBContentList(mediaService.MediaListHook)

	a.api = a.resources.Group(a.config.APIBaseName)
	a.api.Group("user").
		Add(app.NewResource("logout", userService.Logout, app.Meta{app.POST: "/logout"}, true)).
		Add(app.NewResource("me", userService.Me, true)).
		Add(app.NewResource("login", userService.Login, app.Meta{app.POST: "/login"}, true))

	a.api.Group("schema").
		Add(app.NewResource("list", schemaService.List, app.Meta{app.GET: ""})).
		Add(app.NewResource("create", schemaService.Create, app.Meta{app.POST: ""})).
		Add(app.NewResource("detail", schemaService.Detail, app.Meta{app.GET: "/:name"})).
		Add(app.NewResource("update", schemaService.Update, app.Meta{app.PUT: "/:name"})).
		Add(app.NewResource("delete", schemaService.Delete, app.Meta{app.DELETE: "/:name"}))

	a.api.Group("content").
		Add(app.NewResource("list", contentService.List, app.Meta{app.GET: "/:schema"})).
		Add(app.NewResource("detail", contentService.Detail, app.Meta{app.GET: "/:schema/:id"})).
		Add(app.NewResource("create", contentService.Create, app.Meta{app.POST: "/:schema"})).
		Add(app.NewResource("update", contentService.Update, app.Meta{app.PUT: "/:schema/:id"})).
		Add(app.NewResource("delete", contentService.Delete, app.Meta{app.DELETE: "/:schema/:id"}))

	a.api.Group("role").
		Add(app.NewResource("list", roleService.List, app.Meta{app.GET: ""})).
		Add(app.NewResource("resources", roleService.Resources, app.Meta{app.GET: "/resources"})).
		Add(app.NewResource("detail", roleService.Detail, app.Meta{app.GET: "/:id"})).
		Add(app.NewResource("create", roleService.Create, app.Meta{app.POST: ""})).
		Add(app.NewResource("update", roleService.Update, app.Meta{app.PUT: "/:id"})).
		Add(app.NewResource("delete", roleService.Delete, app.Meta{app.DELETE: "/:id"}))

	a.api.Group("media").
		Add(app.NewResource("upload", mediaService.Upload, app.Meta{app.POST: "/upload"})).
		Add(app.NewResource("delete", mediaService.Delete, app.Meta{app.DELETE: ""}))

	a.api.Group("tool").
		Add(app.NewResource("stats", toolService.Stats, app.Meta{app.GET: "/stats"}))

	return nil
}

func getAppDir(dirs ...string) (string, error) {
	cwd, err := os.Getwd()

	if err != nil {
		return "", err
	}

	if len(dirs) == 0 {
		return cwd, nil
	}

	if strings.HasPrefix(dirs[0], "/") {
		return dirs[0], nil
	}

	return path.Join(cwd, dirs[0]), nil
}

func parseEnvFile(envFile string) error {
	if utils.IsFileExists(envFile) {
		if err := godotenv.Load(envFile); err != nil {
			return err
		}
	}

	return nil
}

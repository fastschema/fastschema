package fastschema

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"strings"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/logger"
	"github.com/fastschema/fastschema/pkg/auth"
	"github.com/fastschema/fastschema/pkg/entdbadapter"
	"github.com/fastschema/fastschema/pkg/errors"
	"github.com/fastschema/fastschema/pkg/rclonefs"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/pkg/zaplogger"
	"github.com/fastschema/fastschema/plugins"
	"github.com/fastschema/fastschema/schema"
	ts "github.com/fastschema/fastschema/services/tool"
	"github.com/joho/godotenv"
)

//go:embed all:dash/*
var embedDashStatic embed.FS

func init() {
	fs.RegisterAuthProviderMaker("local", auth.NewLocalAuthProvider)
	fs.RegisterAuthProviderMaker("github", auth.NewGithubAuthProvider)
	fs.RegisterAuthProviderMaker("google", auth.NewGoogleAuthProvider)
}

func (a *App) init() (err error) {
	plugins, err := plugins.NewManager(path.Join(a.dataDir, "plugins"))
	if err != nil {
		return err
	}

	defer func() {
		if err == nil {
			err = plugins.Init(a)
		}
	}()

	if err = plugins.Config(a); err != nil {
		return err
	}

	if err = a.createDisks(); err != nil {
		return err
	}

	if err = a.createLogger(); err != nil {
		return err
	}

	if err := a.createAuthProviders(); err != nil {
		return err
	}

	if err = a.createSchemaBuilder(); err != nil {
		return err
	}

	if err = a.createResources(); err != nil {
		return err
	}

	if err = a.createDBClient(); err != nil {
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

	if err = a.createSetupPage(); err != nil {
		return err
	}

	a.statics = append(a.statics, &fs.StaticFs{
		BasePath:   "/" + a.config.DashBaseName,
		Root:       http.FS(embedDashStatic),
		PathPrefix: "dash",
	})

	return nil
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

	if a.config.Hooks == nil {
		a.config.Hooks = &fs.Hooks{
			DBHooks:     &db.Hooks{},
			PreResolve:  []fs.Middleware{},
			PostResolve: []fs.Middleware{},
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
		if err := json.Unmarshal([]byte(utils.Env("AUTH")), &a.config.AuthConfig); err != nil {
			return err
		}
	}

	return nil
}

func (a *App) createSetupPage() error {
	setupToken, err := a.GetSetupToken(context.Background())
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
				c,
				a.DB(),
				a.Logger(),
				setupData.Username, setupData.Email, setupData.Password,
			); err != nil {
				return false, err
			}

			if err := a.UpdateCache(c); err != nil {
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

	return nil
}

func (a *App) createDisks() (err error) {
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

func (a *App) createLogger() (err error) {
	if a.config.Logger != nil {
		return nil
	}

	if a.config.LoggerConfig == nil {
		a.config.LoggerConfig = &logger.Config{
			Development: utils.Env("APP_ENV", "development") == "development",
			LogFile:     path.Join(a.logDir, "app.log"),
		}
	}

	a.config.Logger, err = zaplogger.NewZapLogger(a.config.LoggerConfig)
	return
}

func (a *App) createAuthProviders() (err error) {
	if a.config.AuthConfig == nil {
		a.config.AuthConfig = &fs.AuthConfig{}
	}

	if a.config.AuthConfig.EnabledProviders == nil {
		a.config.AuthConfig.EnabledProviders = []string{}
	}

	if !utils.Contains(a.config.AuthConfig.EnabledProviders, "local") {
		a.config.AuthConfig.EnabledProviders = append(a.config.AuthConfig.EnabledProviders, "local")
	}

	availableProviders := fs.AuthProviders()
	for _, name := range a.config.AuthConfig.EnabledProviders {
		if _, ok := a.authProviders[name]; ok {
			return fmt.Errorf("auth provider %s is already registered", name)
		}

		if !utils.Contains(availableProviders, name) {
			return fmt.Errorf("auth provider %s is not available", name)
		}

		config := a.config.AuthConfig.Providers[name]
		redirectURL := fmt.Sprintf("%s/%s/auth/%s/callback", a.config.BaseURL, a.config.APIBaseName, name)
		provider, err := fs.CreateAuthProvider(name, config, redirectURL)
		if err != nil {
			return err
		}

		if la, ok := provider.(*auth.LocalAuthProvider); ok {
			la.Init(a.DB, a.Key)
		}

		a.authProviders[name] = provider
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

func (a *App) createDBClient() (err error) {
	if a.DB() != nil {
		return nil
	}

	if a.config.DBConfig == nil {
		a.config.DBConfig = &db.Config{
			Driver:             utils.Env("DB_DRIVER", "sqlite"),
			Name:               utils.Env("DB_NAME"),
			User:               utils.Env("DB_USER"),
			Pass:               utils.Env("DB_PASS"),
			Host:               utils.Env("DB_HOST", "localhost"),
			Port:               utils.Env("DB_PORT"),
			LogQueries:         utils.Env("DB_LOGGING", "false") == "true",
			DisableForeignKeys: utils.Env("DB_DISABLE_FOREIGN_KEYS", "false") == "true",
			Logger:             a.Logger(),
			MigrationDir:       a.migrationDir,
			Hooks: func() *db.Hooks {
				return a.config.Hooks.DBHooks
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

func (a *App) GetSetupToken(ctx context.Context) (string, error) {
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
	countOption := &db.QueryOption{
		Column: "id",
		Unique: true,
	}

	if userCount, err = db.Builder[*fs.User](a.DB()).Count(ctx, countOption); err != nil {
		return false, err
	}

	if roleCount, err = db.Builder[*fs.Role](a.DB()).Count(ctx, countOption); err != nil {
		return false, err
	}

	return userCount == 0 && roleCount == 0, nil
}

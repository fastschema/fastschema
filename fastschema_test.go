package fastschema_test

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"net/http/httptest"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/fastschema/fastschema"
	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/entity"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/logger"
	"github.com/fastschema/fastschema/pkg/entdbadapter"
	"github.com/fastschema/fastschema/pkg/restfulresolver"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
	"github.com/stretchr/testify/assert"
)

func clearEnvs(t *testing.T) {
	envKeys := []string{
		"APP_KEY",
		"APP_PORT",
		"APP_BASE_URL",
		"APP_DASH_URL",
		"APP_API_BASE_NAME",
		"DB_DRIVER",
		"DB_NAME",
		"DB_HOST",
		"DB_PORT",
		"DB_USER",
		"DB_PASS",
		"STORAGE",
		"MAIL",
	}

	for _, key := range envKeys {
		assert.NoError(t, os.Unsetenv(key))
	}
}

// Case 1: Test app custom dir with absolute path
func TestFastSchemaCustomDirAbsolute(t *testing.T) {
	clearEnvs(t)
	config := &fs.Config{
		HideResourcesInfo: true,
		Dir:               t.TempDir(),
	}
	app, err := fastschema.New(config)
	assert.NoError(t, err)
	assert.NotNil(t, app)
	assert.Equal(t, config.Dir, app.Dir())
	envFile := path.Join(config.Dir, "data", ".env")
	assert.FileExists(t, envFile)
	content := string(utils.Must(os.ReadFile(envFile)))
	assert.Contains(t, content, "APP_KEY=")
}

// Case 2: Test app custom dir with relative path
func TestFastSchemaCustomDirRelative(t *testing.T) {
	clearEnvs(t)
	config := &fs.Config{
		HideResourcesInfo: true,
		Dir:               "./",
	}
	app, err := fastschema.New(config)
	assert.NoError(t, err)
	assert.NotNil(t, app)
	assert.Equal(t, path.Join(app.CWD(), config.Dir), app.Dir())
}

// Case 3: Test app custom dir with empty path
func TestFastSchemaCustomDirDefault(t *testing.T) {
	clearEnvs(t)
	config := &fs.Config{
		HideResourcesInfo: true,
	}
	app, err := fastschema.New(config)
	assert.NoError(t, err)
	assert.NotNil(t, app)
	assert.Equal(t, app.CWD(), app.Dir())
}

func TestFastSchemaPrepareConfig(t *testing.T) {
	readonlyDir := "/dev/null"
	config := &fs.Config{
		HideResourcesInfo: true,
		Dir:               readonlyDir,
	}
	_, err := fastschema.New(config)
	assert.Error(t, err)

	clearEnvs(t)
	config = &fs.Config{
		HideResourcesInfo: true,
		Dir:               t.TempDir(),
	}
	envContent := `APP_KEY=testKey
		APP_PORT=8001
		APP_BASE_URL=http://localhost:8001
		APP_DASH_URL=http://localhost:8001/testdash
		APP_API_BASE_NAME=testapi
		APP_DASH_BASE_NAME=testdash`

	dataDir := path.Join(config.Dir, "data")
	assert.NoError(t, os.MkdirAll(dataDir, os.ModePerm))
	assert.NoError(t, utils.WriteFile(path.Join(dataDir, ".env"), envContent))

	app, err := fastschema.New(config)
	assert.NoError(t, err)
	assert.NotNil(t, app)
	assert.Equal(t, config.Dir, app.Dir())
	assert.Equal(t, "testKey", app.Config().AppKey)
	assert.Equal(t, "8001", app.Config().Port)
	assert.Equal(t, "http://localhost:8001", app.Config().BaseURL)
	assert.Equal(t, "http://localhost:8001/testdash", app.Config().DashURL)
	assert.Equal(t, "testapi", app.Config().APIBaseName)
	assert.Equal(t, "testdash", app.Config().DashBaseName)
}

func TestFastschema(t *testing.T) {
	clearEnvs(t)
	config := &fs.Config{
		HideResourcesInfo: true,
		Dir:               t.TempDir(),
	}
	app, err := fastschema.New(config)
	assert.NoError(t, err)
	assert.NotNil(t, app)

	assert.NotNil(t, app.Name())
	assert.NotNil(t, app.Config())
	assert.NotNil(t, app.Logger())
	assert.NotNil(t, app.DB())
	assert.NotNil(t, app.SchemaBuilder())
	assert.NotNil(t, app.Resources())
	assert.NotNil(t, app.Services())
	assert.NotNil(t, app.Roles())
	assert.NotNil(t, app.Hooks())
	assert.NotEmpty(t, app.Key())
	assert.NotEmpty(t, app.Disk())
	assert.NotNil(t, app.Disk("public"))
	assert.Nil(t, app.Disk("invalid"))
}

func TestFastschemaSchemaBuilder(t *testing.T) {
	clearEnvs(t)
	config := &fs.Config{
		HideResourcesInfo: true,
		Dir:               t.TempDir(),
	}
	a, err := fastschema.New(config)
	assert.NoError(t, err)
	assert.NotNil(t, a)
	assert.NotNil(t, a.SchemaBuilder())

	// Create schema builder error
	config = &fs.Config{
		HideResourcesInfo: true,
		Dir:               t.TempDir(),
		SystemSchemas:     []any{"invalid"},
	}
	_, err = fastschema.New(config)
	assert.Error(t, err)
}

func TestFastschemaDisk(t *testing.T) {
	clearEnvs(t)

	// Case 1: Default disk
	config := &fs.Config{
		HideResourcesInfo: true,
		Dir:               t.TempDir(),
	}
	a := utils.Must(fastschema.New(config))
	assert.GreaterOrEqual(t, len(a.Disks()), 1)
	assert.Equal(t, "public", a.Disks()[0].Name())
	assert.Equal(t, path.Join(config.Dir, "data", "public"), a.Disk().Root())

	// Case 2: Invalid disks env
	t.Setenv("STORAGE", "invalid json")
	a, err := fastschema.New(config)
	assert.Error(t, err)
	assert.Nil(t, a)

	// Case 3: Invalid disks config (has no root)
	config = &fs.Config{
		HideResourcesInfo: true,
		Dir:               t.TempDir(),
		StorageConfig: &fs.StorageConfig{
			DefaultDisk: "local_private",
			Disks: []*fs.DiskConfig{{
				Name:       "local_private",
				Driver:     "local",
				PublicPath: "/files",
			}},
		},
	}
	clearEnvs(t)
	_, err = fastschema.New(config)
	assert.Error(t, err)

	// Case 4: Invalid default disk
	config.StorageConfig.DefaultDisk = "invalid"
	config.StorageConfig.Disks[0].Root = "./private"
	_, err = fastschema.New(config)
	assert.Error(t, err)
}

func TestFastschemaMailer(t *testing.T) {
	clearEnvs(t)
	config := &fs.Config{
		HideResourcesInfo: true,
		Dir:               t.TempDir(),
	}
	a, err := fastschema.New(config)
	assert.NoError(t, err)
	assert.NotNil(t, a)
	assert.Nil(t, a.Mailer())

	config = &fs.Config{
		HideResourcesInfo: true,
		Dir:               t.TempDir(),
		MailConfig: &fs.MailConfig{
			SenderMail:        "admin@site.local",
			DefaultClientName: "test",
			Clients: []fs.Map{
				{
					"name":     "test",
					"driver":   "smtp",
					"host":     "site.local",
					"port":     587,
					"username": "test",
					"password": "test",
				},
			},
		},
	}

	a, err = fastschema.New(config)
	assert.NoError(t, err)
	assert.NotNil(t, a)
	assert.NotNil(t, a.Mailer())
	assert.NotNil(t, a.Mailer("test"))
	assert.Nil(t, a.Mailer("invalid"))
	assert.Len(t, a.Mailers(), 1)

	// Invalid sender email
	config.MailConfig.SenderMail = "invalid"
	_, err = fastschema.New(config)
	assert.Error(t, err)

	// Client without host
	config.MailConfig.SenderMail = "admin@site.local"
	config.MailConfig.Clients[0]["host"] = ""
	_, err = fastschema.New(config)
	assert.Error(t, err)

	// Invalid default client
	config.MailConfig.Clients[0]["host"] = "site.local"
	config.MailConfig.DefaultClientName = "invalid"
	_, err = fastschema.New(config)
	assert.Error(t, err)

	// There is no default client config
	config.MailConfig.DefaultClientName = ""
	a, err = fastschema.New(config)
	assert.NoError(t, err)
	assert.NotNil(t, a)
	assert.Equal(t, "test", a.Mailer().Name())

	// User mail config from env
	t.Setenv("MAIL", `{"sender_mail":"accounts@fastschema.com","default_client":"testsmtp","clients":[{"name":"testsmtp","driver":"smtp","host":"site.local","port":2525,"username":"test","password":"test"}]}`)
	a, err = fastschema.New(&fs.Config{
		HideResourcesInfo: true,
	})
	assert.NoError(t, err)
	assert.NotNil(t, a)
	assert.NotNil(t, a.Mailer())

	// Invalid email config from env
	t.Setenv("MAIL", `invalid json`)
	_, err = fastschema.New(&fs.Config{
		HideResourcesInfo: true,
	})
	assert.Error(t, err)
}

func TestFastschemaLogger(t *testing.T) {
	clearEnvs(t)
	config := &fs.Config{
		HideResourcesInfo: true,
		Dir:               t.TempDir(),
	}
	a, err := fastschema.New(config)
	assert.NoError(t, err)
	assert.NotNil(t, a)
	assert.NotNil(t, a.Logger())

	config = &fs.Config{
		HideResourcesInfo: true,
		Dir:               t.TempDir(),
		Logger:            logger.CreateMockLogger(true),
	}
	a, err = fastschema.New(config)
	assert.NoError(t, err)
	mockLogger, ok := a.Logger().(*logger.MockLogger)
	assert.True(t, ok)
	assert.NotNil(t, mockLogger)

	// Logger with read-only file'
	config = &fs.Config{
		HideResourcesInfo: true,
		Dir:               t.TempDir(),
		LoggerConfig: &logger.Config{
			LogFile: "/dev/null/test.log",
		},
	}
	_, err = fastschema.New(config)
	assert.Error(t, err)
}

func TestFastschemaDBClient(t *testing.T) {
	clearEnvs(t)
	config := &fs.Config{
		HideResourcesInfo: true,
		Dir:               t.TempDir(),
	}
	a, err := fastschema.New(config)
	assert.NoError(t, err)
	assert.NotNil(t, a)
	assert.NotNil(t, a.DB())

	sb := utils.Must(schema.NewBuilderFromDir(t.TempDir(), fs.SystemSchemaTypes...))
	db := utils.Must(entdbadapter.NewTestClient(utils.Must(os.MkdirTemp("", "migrations")), sb))
	config = &fs.Config{
		HideResourcesInfo: true,
		Dir:               t.TempDir(),
		DB:                db,
	}
	a, err = fastschema.New(config)
	assert.NoError(t, err)
	assert.NotNil(t, a)
	assert.NotNil(t, a.DB())

	t.Setenv("DB_DRIVER", "invalid")
	config = &fs.Config{
		HideResourcesInfo: true,
		Dir:               t.TempDir(),
	}
	a, err = fastschema.New(config)
	assert.Error(t, err)
	assert.Nil(t, a)

	// error
	clearEnvs(t)
	schemas := utils.Must(schema.GetSchemasFromDir(t.TempDir()))
	delete(schemas, "user")
	delete(schemas, "role")
	delete(schemas, "permission")
	delete(schemas, "file")

	sb = utils.Must(schema.NewBuilderFromSchemas(t.TempDir(), schemas))
	db = utils.Must(entdbadapter.NewTestClient(utils.Must(os.MkdirTemp("", "migrations")), sb))
	config = &fs.Config{
		HideResourcesInfo: true,
		Dir:               t.TempDir(),
		DB:                db,
	}
	_, err = fastschema.New(config)
	assert.Error(t, err)
}

func TestFastschemaReload(t *testing.T) {
	clearEnvs(t)
	config := &fs.Config{
		HideResourcesInfo: true,
		Dir:               t.TempDir(),
	}
	a, err := fastschema.New(config)
	assert.NoError(t, err)
	assert.NotNil(t, a)

	// reload
	assert.NoError(t, a.Reload(context.Background(), nil))

	// Reload error because of invalid migration
	assert.Error(t, a.Reload(context.Background(), &db.Migration{
		RenameTables: []*db.RenameItem{{
			From: "invalid",
			To:   "invalid",
		}},
	}))

	// Reload error because app dir is removed
	assert.NoError(t, os.RemoveAll(a.Dir()))
	assert.Error(t, a.Reload(context.Background(), nil))
}

func TestFastschemaUpdateCache(t *testing.T) {
	clearEnvs(t)
	ctx := context.Background()
	config := &fs.Config{
		HideResourcesInfo: true,
		Dir:               t.TempDir(),
	}
	a, err := fastschema.New(config)
	assert.NoError(t, err)
	assert.NotNil(t, a)

	// reload
	assert.NoError(t, a.UpdateCache(ctx))

	// Error compile role rule
	rs1, err := a.DB().Exec(ctx, "INSERT INTO roles (id, name, rule) VALUES (1, 'testrole', 'invalid')")
	assert.NoError(t, err)
	assert.Error(t, a.UpdateCache(ctx))

	// Error compile permission value
	_, err = a.DB().Exec(ctx, "UPDATE roles SET rule = NULL WHERE id = ?", utils.Must(rs1.LastInsertId()))
	assert.NoError(t, err)
	_, err = a.DB().Exec(ctx, "INSERT INTO permissions (role_id, resource, value) VALUES (?, 'testpermission', 'invalid')", utils.Must(rs1.LastInsertId()))
	assert.NoError(t, err)
	assert.Error(t, a.UpdateCache(ctx))

	// Error db client is closed
	assert.NoError(t, a.DB().Close())
	assert.Error(t, a.UpdateCache(ctx))
}

func TestFastschemaSetup(t *testing.T) {
	clearEnvs(t)
	ctx := context.Background()
	sb := utils.Must(schema.NewBuilderFromDir(t.TempDir(), fs.SystemSchemaTypes...))
	tdb := utils.Must(entdbadapter.NewTestClient(utils.Must(os.MkdirTemp("", "migrations")), sb))
	config := &fs.Config{
		HideResourcesInfo: true,
		Dir:               t.TempDir(),
		DB:                tdb,
	}
	a, err := fastschema.New(config)
	assert.NoError(t, err)
	assert.NotNil(t, a)
	assert.NotEmpty(t, utils.Must(a.GetSetupToken(ctx)))

	// no need to setup
	clearEnvs(t)
	sb = utils.Must(schema.NewBuilderFromDir(t.TempDir(), fs.SystemSchemaTypes...))
	tdb = utils.Must(entdbadapter.NewTestClient(utils.Must(os.MkdirTemp("", "migrations")), sb))
	config = &fs.Config{
		HideResourcesInfo: true,
		Dir:               t.TempDir(),
		DB:                tdb,
	}
	a, err = fastschema.New(config)
	assert.NoError(t, err)

	_, err = db.Create[*fs.Role](ctx, a.DB(), fs.Map{"name": "admin"})
	assert.NoError(t, err)

	setupToken, err := a.GetSetupToken(ctx)
	assert.NoError(t, err)
	assert.Empty(t, setupToken)
}

func TestFastschemaResources(t *testing.T) {
	clearEnvs(t)
	var err error
	sb := utils.Must(schema.NewBuilderFromDir(t.TempDir(), fs.SystemSchemaTypes...))
	entDB := utils.Must(entdbadapter.NewTestClient(
		utils.Must(os.MkdirTemp("", "migrations")),
		sb,
		nil,
	))
	config := &fs.Config{
		HideResourcesInfo: true,
		Dir:               t.TempDir(),
		DB:                entDB,
	}
	a, err := fastschema.New(config)
	assert.NoError(t, err)
	assert.NotNil(t, a)

	a.AddMiddlewares(func(c fs.Context) error {
		restContext, ok := c.(*restfulresolver.Context)
		assert.True(t, ok)

		if restContext.Path() == "/error" {
			return fmt.Errorf("error_from_middleware")
		}

		return c.Next()
	})

	a.OnPreResolve(func(c fs.Context) error {
		assert.NotNil(t, c.Resource())
		return nil
	})

	a.OnPostResolve(func(c fs.Context) error {
		assert.NotNil(t, c.Resource())
		return nil
	})

	a.OnPreDBQuery(func(ctx context.Context, option *db.QueryOption) error {
		assert.NotNil(t, option)
		return nil
	})

	a.OnPostDBQuery(func(
		ctx context.Context,
		query *db.QueryOption,
		entities []*entity.Entity,
	) ([]*entity.Entity, error) {
		if query.Schema.Name == "file" {
			entities = append(entities, entity.New(1))
		}
		return entities, nil
	})

	a.OnPreDBExec(func(ctx context.Context, option *db.QueryOption) error {
		assert.NotNil(t, option)
		return nil
	})

	a.OnPostDBExec(func(ctx context.Context, option *db.QueryOption, result sql.Result) error {
		assert.NotNil(t, option)
		assert.NotNil(t, result)
		return nil
	})

	a.OnPreDBCreate(func(ctx context.Context, schema *schema.Schema, createData *entity.Entity) error {
		assert.NotNil(t, schema)
		assert.NotNil(t, createData)
		return nil
	})

	a.OnPostDBCreate(func(
		ctx context.Context,
		schema *schema.Schema,
		dataCreate *entity.Entity,
		id uint64,
	) error {
		assert.NotNil(t, schema)
		assert.NotNil(t, dataCreate)
		assert.Greater(t, id, uint64(0))
		return nil
	})

	a.OnPreDBUpdate(func(
		ctx context.Context,
		schema *schema.Schema,
		predicates []*db.Predicate,
		updateData *entity.Entity,
	) error {
		assert.NotNil(t, schema)
		assert.NotNil(t, updateData)
		return nil
	})

	a.OnPostDBUpdate(func(
		ctx context.Context,
		schema *schema.Schema,
		predicates []*db.Predicate,
		updateData *entity.Entity,
		originalEntities []*entity.Entity,
		affected int,
	) error {
		assert.NotNil(t, schema)
		assert.NotNil(t, updateData)
		assert.Greater(t, affected, 0)
		return nil
	})

	a.OnPreDBDelete(func(ctx context.Context, schema *schema.Schema, predicates []*db.Predicate) error {
		assert.NotNil(t, schema)
		return nil
	})

	a.OnPostDBDelete(func(
		ctx context.Context,
		schema *schema.Schema,
		predicates []*db.Predicate,
		originalEntities []*entity.Entity,
		affected int,
	) error {
		assert.NotNil(t, schema)
		assert.Greater(t, affected, 0)
		return nil
	})

	hooks := a.Hooks()
	assert.NotNil(t, hooks)
	assert.NotNil(t, hooks.DBHooks)
	// All db hooks expected 2 includes: the default one and the one we added in the test
	assert.Len(t, hooks.DBHooks.PostDBQuery, 2)
	assert.Len(t, hooks.DBHooks.PostDBCreate, 2)
	assert.Len(t, hooks.DBHooks.PostDBUpdate, 2)
	assert.Len(t, hooks.DBHooks.PostDBDelete, 2)

	a.AddResource(fs.NewResource("test", func(c fs.Context, _ any) (any, error) {
		return "test", nil
	}, &fs.Meta{Public: true}))

	a.AddResource(fs.NewResource("error", func(c fs.Context, _ any) (any, error) {
		return "test", nil
	}, &fs.Meta{Public: true}))

	resources := a.Resources()
	assert.NotNil(t, resources)
	assert.True(t, len(a.API().Resources()) > 0)

	assert.NoError(t, resources.Init())
	server := restfulresolver.NewRestfulResolver(&restfulresolver.ResolverConfig{
		ResourceManager: resources,
		Logger:          logger.CreateMockLogger(true),
	}).Server()

	req := httptest.NewRequest("GET", "/test", nil)
	resp := utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 200, resp.StatusCode)
	assert.Contains(t, utils.Must(utils.ReadCloserToString(resp.Body)), `test`)

	req = httptest.NewRequest("GET", "/error", nil)
	resp = utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 500, resp.StatusCode)
	assert.Contains(t, utils.Must(utils.ReadCloserToString(resp.Body)), `error_from_middleware`)

	// Setup empty token
	req = httptest.NewRequest("POST", "/api/setup", bytes.NewReader([]byte(`{"token":""}`)))
	resp = utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 403, resp.StatusCode)
	assert.Contains(t, utils.Must(utils.ReadCloserToString(resp.Body)), `Invalid setup data or token`)

	// Setup invalid token
	req = httptest.NewRequest("POST", "/api/setup", bytes.NewReader([]byte(`{"token":"aaaaa"}`)))
	resp = utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 403, resp.StatusCode)
	assert.Contains(t, utils.Must(utils.ReadCloserToString(resp.Body)), `Invalid setup data or token`)

	// Setup success
	setupToken := utils.Must(a.GetSetupToken(context.Background()))
	req = httptest.NewRequest("POST", "/api/setup", bytes.NewReader([]byte(`{
		"token":"`+setupToken+`",
		"username":"admin",
		"email":"admin@local.ltd",
		"password":"123"
	}`)))
	resp = utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 200, resp.StatusCode)

	// Login
	req = httptest.NewRequest("POST", "/api/auth/local/login", bytes.NewReader([]byte(`{
		"login":"admin",
		"password":"123"
	}`)))
	resp = utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 200, resp.StatusCode)
	response := utils.Must(utils.ReadCloserToString(resp.Body))
	assert.Contains(t, response, `"token":"`)
	token := strings.Split(response, `"token":"`)[1]
	token = strings.Split(token, `"`)[0]

	req = httptest.NewRequest("GET", "/api/content/file", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp = utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 200, resp.StatusCode)
	assert.Contains(t, utils.Must(utils.ReadCloserToString(resp.Body)), `"current_page":1`)

	// Setup not available
	req = httptest.NewRequest("POST", "/api/setup", bytes.NewReader([]byte(`{"token":"aaaaa"}`)))
	resp = utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 400, resp.StatusCode)
	assert.Contains(t, utils.Must(utils.ReadCloserToString(resp.Body)), `Setup token is not available`)

	// Test openapi spec
	req = httptest.NewRequest("GET", "/docs/openapi.json", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp = utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 200, resp.StatusCode)
	assert.Contains(t, utils.Must(utils.ReadCloserToString(resp.Body)), `FastSchema OAS3`)

	// Test swagger ui
	req = httptest.NewRequest("GET", "/docs/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp = utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 200, resp.StatusCode)
	assert.Contains(t, utils.Must(utils.ReadCloserToString(resp.Body)), `/docs/openapi.json`)

	// Test config
	req = httptest.NewRequest("GET", "/api/config", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp = utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 200, resp.StatusCode)
	assert.Contains(t, utils.Must(utils.ReadCloserToString(resp.Body)), `"version":"`)
}

func TestFastschemaStart(t *testing.T) {
	clearEnvs(t)
	sb := utils.Must(schema.NewBuilderFromDir(t.TempDir(), fs.SystemSchemaTypes...))
	db := utils.Must(entdbadapter.NewTestClient(utils.Must(os.MkdirTemp("", "migrations")), sb, func() *db.Hooks {
		return &db.Hooks{}
	}))
	config := &fs.Config{
		HideResourcesInfo: false,
		Dir:               t.TempDir(),
		DB:                db,
		Port:              "8080",
	}
	a, err := fastschema.New(config)
	assert.NoError(t, err)
	assert.NotNil(t, a)

	// Test Start
	go func() {
		time.Sleep(10 * time.Millisecond)
		a2, err := fastschema.New(config)
		assert.NoError(t, err)
		assert.Error(t, a2.Start())
		assert.NoError(t, a.Shutdown())
	}()

	assert.NoError(t, a.Start())
}

func TestFastSchemaCustomConfiguration(t *testing.T) {
	clearEnvs(t)
	config := &fs.Config{
		HideResourcesInfo: false,
		Dir:               t.TempDir(),
		DBConfig: &db.Config{
			Driver: "sqlite",
		},
		StorageConfig: &fs.StorageConfig{
			// DefaultDisk: "local_public",
			Disks: []*fs.DiskConfig{
				{
					Name:       "local_public",
					Driver:     "local",
					Root:       "./public",
					BaseURL:    "http://localhost:8000/files",
					PublicPath: "/files", // This will expose the files in the public path
				},
				{
					Name:   "local_private",
					Driver: "local",
					Root:   "./private",
				},
			},
		},
	}
	app, err := fastschema.New(config)
	assert.NoError(t, err)
	assert.NotNil(t, app)
	assert.NotNil(t, app.DB())
}

func TestFastSchemaGetAuthProvider(t *testing.T) {
	clearEnvs(t)
	// Case 0: Load invalid auth from env
	t.Setenv("AUTH", `invalid json`)
	config := &fs.Config{
		HideResourcesInfo: true,
		Dir:               t.TempDir(),
	}
	app, err := fastschema.New(config)
	assert.Error(t, err)
	assert.Nil(t, app)

	// Case 1: Duplicate provider
	authConfig := &fs.AuthConfig{
		EnabledProviders: []string{"github", "github"},
		Providers: map[string]fs.Map{
			"github": {
				"client_id":     "test",
				"client_secret": "test",
			},
		},
	}
	config = &fs.Config{
		HideResourcesInfo: true,
		Dir:               t.TempDir(),
		AuthConfig:        authConfig,
	}
	_, err = fastschema.New(config)
	assert.Error(t, err)

	// Case 3: Error
	config = &fs.Config{
		HideResourcesInfo: true,
		Dir:               t.TempDir(),
		AuthConfig: &fs.AuthConfig{
			EnabledProviders: []string{"github"},
			Providers: map[string]fs.Map{
				"github": {},
			},
		},
	}
	app, err = fastschema.New(config)
	assert.Error(t, err)
	assert.Nil(t, app)

	config = &fs.Config{
		HideResourcesInfo: true,
		Dir:               t.TempDir(),
		AuthConfig: &fs.AuthConfig{
			EnabledProviders: []string{"github"},
			Providers: map[string]fs.Map{
				"github": {
					"client_id":     "test",
					"client_secret": "test",
				},
				"google": {
					"client_id":     "test",
					"client_secret": "test",
				},
			},
		},
	}
	app, err = fastschema.New(config)
	assert.NoError(t, err)
	assert.NotNil(t, app)

	provider := app.GetAuthProvider("github")
	assert.NotNil(t, provider)
	assert.Equal(t, "github", provider.Name())
}

func TestFastSchemaHTTPAdaptor(t *testing.T) {
	clearEnvs(t)
	config := &fs.Config{
		Dir: t.TempDir(),
	}
	a, err := fastschema.New(config)
	assert.NoError(t, err)
	assert.NotNil(t, a)

	// Test HTTPAdaptor with initialized resources
	handler, err := a.HTTPAdaptor()
	assert.NoError(t, err)
	assert.NotNil(t, handler)
}

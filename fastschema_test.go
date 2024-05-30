package fastschema_test

import (
	"bytes"
	"context"
	"fmt"
	"net/http/httptest"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/fastschema/fastschema"
	"github.com/fastschema/fastschema/db"
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
		"STORAGE_DEFAULT_DISK",
		"STORAGE_DISKS",
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
	clearEnvs(t)
	config := &fs.Config{
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

	assert.NotNil(t, app.Config())
	assert.NotNil(t, app.Logger())
	assert.NotNil(t, app.DB())
	assert.NotNil(t, app.SchemaBuilder())
	assert.NotNil(t, app.Resources())
	assert.NotNil(t, app.Roles())
	assert.NotNil(t, app.Hooks())
	assert.NotEmpty(t, app.Key())
	assert.NotEmpty(t, app.Disk())
	assert.NotNil(t, app.Disk("public"))
	assert.Nil(t, app.Disk("invalid"))
}

func TestFastschemaDisk(t *testing.T) {
	clearEnvs(t)

	// Case 1: Default disk
	config := &fs.Config{
		HideResourcesInfo: true,
		Dir:               t.TempDir(),
	}
	a := utils.Must(fastschema.New(config))
	assert.Len(t, a.Disks(), 1)
	assert.Equal(t, "public", a.Disks()[0].Name())
	assert.Equal(t, path.Join(config.Dir, "data", "public"), a.Disk().Root())

	// Case 2: Invalid disks env
	t.Setenv("STORAGE_DISKS", "invalid json")
	a, err := fastschema.New(config)
	assert.Error(t, err)
	assert.Nil(t, a)

	// Case 3: Invalid default disk
	clearEnvs(t)
	t.Setenv("STORAGE_DEFAULT_DISK", "invalid")
	_, err = fastschema.New(config)
	assert.Error(t, err)

	// Case 4: Invalid disks config (has no root)
	clearEnvs(t)
	_, err = fastschema.New(&fs.Config{
		HideResourcesInfo: true,
		Dir:               t.TempDir(),
		StorageConfig: &fs.StorageConfig{
			DefaultDisk: "local_private",
			DisksConfig: []*fs.DiskConfig{{
				Name:       "local_private",
				Driver:     "local",
				PublicPath: "/files",
			}},
		},
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

	// reload error
	assert.NoError(t, a.Reload(context.Background(), nil))
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
	assert.NotEmpty(t, utils.Must(a.SetupToken(ctx)))

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

	setupToken, err := a.SetupToken(ctx)
	assert.NoError(t, err)
	assert.Empty(t, setupToken)
}

func TestFastschemaResources(t *testing.T) {
	clearEnvs(t)
	var err error
	a := &fastschema.App{}
	sb := utils.Must(schema.NewBuilderFromDir(t.TempDir(), fs.SystemSchemaTypes...))
	entDB := utils.Must(entdbadapter.NewTestClient(utils.Must(os.MkdirTemp("", "migrations")), sb, func() *db.Hooks {
		return a.Hooks().DBHooks
	}))
	config := &fs.Config{
		HideResourcesInfo: true,
		Dir:               t.TempDir(),
		DB:                entDB,
	}
	a, err = fastschema.New(config)
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

	a.OnPostDBGet(func(query *db.QueryOption, entities []*schema.Entity) ([]*schema.Entity, error) {
		if query.Model.Schema().Name == "file" {
			entities = append(entities, schema.NewEntity(1))
		}
		return entities, nil
	})

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
	server := restfulresolver.NewRestfulResolver(resources, logger.CreateMockLogger(false)).Server()

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
	setupToken := utils.Must(a.SetupToken(context.Background()))
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
	req = httptest.NewRequest("POST", "/api/user/login", bytes.NewReader([]byte(`{
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
}

func TestFastschemaStart(t *testing.T) {
	clearEnvs(t)
	sb := utils.Must(schema.NewBuilderFromDir(t.TempDir(), fs.SystemSchemaTypes...))
	db := utils.Must(entdbadapter.NewTestClient(utils.Must(os.MkdirTemp("", "migrations")), sb, func() *db.Hooks {
		return &db.Hooks{}
	}))
	config := &fs.Config{
		HideResourcesInfo: true,
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

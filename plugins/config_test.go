package plugins_test

import (
	"context"
	"fmt"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/entity"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/logger"
	"github.com/fastschema/fastschema/pkg/restfulresolver"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/qjs"
	"github.com/stretchr/testify/assert"
)

const pluginContent = `
const productSchema = {
  "name": "product",
  "namespace": "products",
  "label_field": "name",
  "disable_timestamp": false,
  "fields": [
    {
      "type": "string",
      "name": "name",
      "label": "Name",
      "optional": true,
      "sortable": true
    },
    {
      "type": "string",
      "name": "description",
      "label": "Description",
      "optional": true,
      "sortable": true,
      "renderer": {
        "class": "textarea"
      }
    }
  ]
};

const PreDBQuery = (context) => {$logEvent('PreDBQuery')};
const PostDBQuery = (context) => {$logEvent('PostDBQuery')};

const PreDBExec = (context) => {$logEvent('PreDBExec')};
const PostDBExec = (context) => {$logEvent('PostDBExec')};

const PreDBCreate = (context) => {$logEvent('PreDBCreate')};
const PostDBCreate = (context) => {$logEvent('PostDBCreate')};

const PreDBUpdate = (context) => {$logEvent('PreDBUpdate')};
const PostDBUpdate = (context) => {$logEvent('PostDBUpdate')};

const PreDBDelete = (context) => {$logEvent('PreDBDelete')};
const PostDBDelete = (context) => {$logEvent('PostDBDelete')};

const PreResolve = (context) => {$logEvent('PreResolve')};
const PostResolve = (context) => {$logEvent('PostResolve')};

export const Config = config => {
	config.Set({ port: '9000' });
	config.AddSchemas(productSchema);

  config.OnPreDBQuery(PreDBQuery);
  config.OnPostDBQuery(PostDBQuery);

  config.OnPreDBExec(PreDBExec);
  config.OnPostDBExec(PostDBExec);

  config.OnPreDBCreate(PreDBCreate);
  config.OnPostDBCreate(PostDBCreate);

  config.OnPreDBUpdate(PreDBUpdate);
  config.OnPostDBUpdate(PostDBUpdate);

  config.OnPreDBDelete(PreDBDelete);
  config.OnPostDBDelete(PostDBDelete);

  config.OnPreResolve(PreResolve);
  config.OnPostResolve(PostResolve);
};

export default {
	Config,
	PreDBQuery,
	PostDBQuery,
	PreDBExec,
	PostDBExec,
	PreDBCreate,
	PostDBCreate,
	PreDBUpdate,
	PostDBUpdate,
	PreDBDelete,
	PostDBDelete,
	PreResolve,
	PostResolve,
};
`

func TestConfigSet(t *testing.T) {
	// Config Set with invalid port type
	_, plugin, _ := createPlugin(t, `export const Config = config => {
		config.Set({
			port: false, // port must be a string
		});
	}
	export default { Config };
	`, nil)
	assert.Error(t, plugin.Config())

	// Config Set with valid values
	testLogFile := filepath.Join(os.TempDir(), "fastschema_test.log")
	defer os.Remove(testLogFile)
	app, plugin, _ := createPlugin(t, fmt.Sprintf(`export const Config = config => {
		config.Set({
			port: '9000',
			app_name: 'FastSchema Test',
			logger_config: {
				development: true,
				log_file: '%s',
				caller_skip: 1,
				disable_console: true
			},
			db_config: {
				driver: 'mysql',
				name: 'testdb',
				host: 'localhost',
				port: '3306',
				user: 'test',
				pass: 'test',
				log_queries: true,
				migration_dir: 'migrations',
				ignore_migration: true,
				disable_foreign_keys: true,
				use_soft_deletes: true
			},
			storage_config: {
				default_disk: 'testdisk',
				disks: [
					{
						name: 'testdisk',
						driver: 's3',
						root: 'storage/app',
						base_url: 'http://localhost/storage',
						public_path: 'public/storage',
						provider: 'aws',
						endpoint: 'http://localhost:9000',
						region: 'us-east-1',
						bucket: 'test-bucket',
						access_key_id: 'test',
						secret_access_key: 'test',
						acl: 'public-read',
					},
				],
			},
		});
	}
	export default { Config };
	`, testLogFile), nil)
	assert.NoError(t, plugin.Config())
	appConfig := app.Config()

	// Common config
	assert.Equal(t, "9000", appConfig.Port)
	assert.Equal(t, "FastSchema Test", appConfig.AppName)

	// Logger config
	assert.NotNil(t, appConfig.LoggerConfig)
	assert.Equal(t, true, appConfig.LoggerConfig.Development)
	assert.Equal(t, testLogFile, appConfig.LoggerConfig.LogFile)
	assert.Equal(t, 1, appConfig.LoggerConfig.CallerSkip)
	assert.Equal(t, true, appConfig.LoggerConfig.DisableConsole)

	// DB config
	assert.NotNil(t, appConfig.DBConfig)
	assert.Equal(t, "mysql", appConfig.DBConfig.Driver)
	assert.Equal(t, "testdb", appConfig.DBConfig.Name)
	assert.Equal(t, "localhost", appConfig.DBConfig.Host)
	assert.Equal(t, "3306", appConfig.DBConfig.Port)
	assert.Equal(t, "test", appConfig.DBConfig.User)
	assert.Equal(t, "test", appConfig.DBConfig.Pass)
	assert.Equal(t, true, appConfig.DBConfig.LogQueries)
	assert.Equal(t, "migrations", appConfig.DBConfig.MigrationDir)
	assert.Equal(t, true, appConfig.DBConfig.IgnoreMigration)
	assert.Equal(t, true, appConfig.DBConfig.DisableForeignKeys)
	assert.Equal(t, true, appConfig.DBConfig.UseSoftDeletes)

	// Storage config
	assert.NotNil(t, appConfig.StorageConfig)
	assert.Equal(t, "testdisk", appConfig.StorageConfig.DefaultDisk)
	assert.Len(t, appConfig.StorageConfig.Disks, 1)
	disk := appConfig.StorageConfig.Disks[0]
	assert.Equal(t, "testdisk", disk.Name)
	assert.Equal(t, "s3", disk.Driver)
	assert.Equal(t, "storage/app", disk.Root)
	assert.Equal(t, "http://localhost/storage", disk.BaseURL)
	assert.Equal(t, "public/storage", disk.PublicPath)
	assert.Equal(t, "aws", disk.Provider)
	assert.Equal(t, "http://localhost:9000", disk.Endpoint)
	assert.Equal(t, "us-east-1", disk.Region)
	assert.Equal(t, "test-bucket", disk.Bucket)
	assert.Equal(t, "test", disk.AccessKeyID)
	assert.Equal(t, "test", disk.SecretAccessKey)
	assert.Equal(t, "public-read", disk.ACL)
}

func TestConfig(t *testing.T) {
	events := []string{}
	runtimeSetup := func(rt *qjs.Runtime, inPool bool) error {
		fmt.Println("Setting up runtime")
		rt.Context().SetFunc("$logEvent", func(this *qjs.This) (*qjs.Value, error) {
			event := this.Args()[0].String()
			events = append(events, event)
			return nil, nil
		})
		return nil
	}
	app, plugin, _ := createPlugin(t, pluginContent, runtimeSetup)
	assert.Equal(t, "test-plugin", plugin.Name())
	assert.NoError(t, plugin.Config())

	// Assert created system schemas
	assert.Len(t, app.Config().SystemSchemas, 1)
	// Assert updated app port
	assert.Equal(t, "9000", app.Config().Port)

	type testHooksLength struct {
		name   string
		expect int
		actual int
	}

	hooks := app.Hooks()
	tests := []testHooksLength{
		{"PreResolve", 2, len(hooks.PreResolve)}, // including the default one: authorize
		{"PostResolve", 1, len(hooks.PostResolve)},
		{"PreDBQuery", 1, len(hooks.DBHooks.PreDBQuery)},
		{"PostDBQuery", 2, len(hooks.DBHooks.PostDBQuery)}, // including the default one: FileListHook
		{"PreDBExec", 1, len(hooks.DBHooks.PreDBExec)},
		{"PostDBExec", 1, len(hooks.DBHooks.PostDBExec)},
		{"PreDBCreate", 1, len(hooks.DBHooks.PreDBCreate)},
		{"PostDBCreate", 2, len(hooks.DBHooks.PostDBCreate)}, // including the default one: realtime.ContentCreateHook
		{"PreDBUpdate", 1, len(hooks.DBHooks.PreDBUpdate)},
		{"PostDBUpdate", 2, len(hooks.DBHooks.PostDBUpdate)}, // including the default one: realtime.ContentUpdateHook
		{"PreDBDelete", 1, len(hooks.DBHooks.PreDBDelete)},
		{"PostDBDelete", 2, len(hooks.DBHooks.PostDBDelete)}, // including the default one: realtime.ContentDeleteHook
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expect, test.actual)
		})
	}

	// Trigger some hooks and assert the events
	tx := utils.Must(app.DB().Tx(context.Background()))
	defer tx.Rollback() // always rollback

	// Create a dummy role to trigger PreDBCreate and PostDBCreate
	utils.Must(db.Builder[*fs.Role](tx).Create(t.Context(), entity.New().Set("name", "test-role")))
	assert.Contains(t, events, "PreDBCreate")
	assert.Contains(t, events, "PostDBCreate")

	// Create also trigger query hooks
	assert.Contains(t, events, "PreDBQuery")
	assert.Contains(t, events, "PostDBQuery")

	// Update a dummy role to trigger PreDBUpdate and PostDBUpdate
	utils.Must(db.Builder[*fs.Role](tx).
		Where(db.EQ("name", "test-role")).
		Update(t.Context(), entity.New().Set("name", "test-role-updated")),
	)
	assert.Contains(t, events, "PreDBUpdate")
	assert.Contains(t, events, "PostDBUpdate")

	// Delete a dummy role to trigger PreDBDelete and PostDBDelete
	utils.Must(db.Builder[*fs.Role](tx).
		Where(db.EQ("name", "test-role-updated")).
		Delete(t.Context()),
	)
	assert.Contains(t, events, "PreDBDelete")
	assert.Contains(t, events, "PostDBDelete")

	// Exec to trigger PreDBExec and PostDBExec
	utils.Must(tx.Exec(t.Context(), "UPDATE roles SET name = name"))
	assert.Contains(t, events, "PreDBExec")
	assert.Contains(t, events, "PostDBExec")

	// Call to /test to trigger PreResolve and PostResolve
	app.AddResource(fs.NewResource("test", func(c fs.Context, _ any) (any, error) {
		return "test", nil
	}, &fs.Meta{Public: true}))
	resources := app.Resources()
	server := restfulresolver.NewRestfulResolver(&restfulresolver.ResolverConfig{
		ResourceManager: resources,
		Logger:          logger.CreateMockLogger(true),
	}).Server()

	req := httptest.NewRequest("GET", "/test", nil)
	resp := utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 200, resp.StatusCode)
	assert.Contains(t, utils.Must(utils.ReadCloserToString(resp.Body)), `test`)
	assert.Contains(t, events, "PreResolve")
	assert.Contains(t, events, "PostResolve")
}

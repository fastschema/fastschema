package plugins_test

import (
	"context"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/dop251/goja_nodejs/console"
	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/logger"
	rr "github.com/fastschema/fastschema/pkg/restfulresolver"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/plugins"
	"github.com/stretchr/testify/assert"
)

const pluginValid = `
const schemaProduct = {
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
}

const onPreResolve = ctx => console.log('hook run: onPreResolve');
const onPostResolve = ctx => console.log('hook run: onPostResolve');

const onPreDBQuery = ctx => console.log('hook run: onPreDBQuery');
const onPostDBQuery = ctx => console.log('hook run: onPostDBQuery');

const onPreDBCreate = ctx => console.log('hook run: onPreDBCreate');
const onPostDBCreate = ctx => console.log('hook run: onPostDBCreate');

const onPreDBUpdate = ctx => console.log('hook run: onPreDBUpdate');
const onPostDBUpdate = ctx => console.log('hook run: onPostDBUpdate');

const onPreDBExec = ctx => console.log('hook run: onPreDBExec');
const onPostDBExec = ctx => console.log('hook run: onPostDBExec');

const onPreDBDelete = ctx => console.log('hook run: onPreDBDelete');
const onPostDBDelete = ctx => console.log('hook run: onPostDBDelete');

const Config = config => {
	config.AddSchemas(schemaProduct);
	config.port = 9000;

	config.OnPreResolve(onPreResolve);
  config.OnPostResolve(onPostResolve);

  config.OnPreDBQuery(onPreDBQuery);
  config.OnPostDBQuery(onPostDBQuery);

  config.OnPreDBExec(onPreDBExec);
  config.OnPostDBExec(onPostDBExec);

  config.OnPreDBCreate(onPreDBCreate);
  config.OnPostDBCreate(onPostDBCreate);

  config.OnPreDBUpdate(onPreDBUpdate);
  config.OnPostDBUpdate(onPostDBUpdate);

  config.OnPreDBDelete(onPreDBDelete);
  config.OnPostDBDelete(onPostDBDelete);
}

const hello = ctx => {
  const result = $db().Query(ctx, "SELECT 'query db in plugin'");
  console.log(result);
  return 'world';
}

const Init = plugin => {
  console.log('init plugin');
  const result = $db().Query($context(), 'SELECT 1');
  console.log(result);

	$logger().Info('hello from plugin');

  plugin.resources.Add(hello, { public: true });
}


`

var pluginNoConfigFunction = `
console.log('no config function');
`

func createPlugin(t *testing.T, dir, name, content string) string {
	pluginDir := filepath.Join(dir, name)
	pluginFile := filepath.Join(pluginDir, "plugin.js")
	assert.NoError(t, os.Mkdir(pluginDir, 0755))
	assert.NoError(t, os.WriteFile(pluginFile, []byte(content), 0644))
	return pluginFile
}

func TestPlugin(t *testing.T) {
	// Override console writer
	stdOutLines := []string{}
	stdErrorLines := []string{}
	plugins.Require.RegisterNativeModule(
		console.ModuleName,
		console.RequireWithPrinter(console.StdPrinter{
			StdoutPrint: func(s string) {
				stdOutLines = append(stdOutLines, s)
			},
			StderrPrint: func(s string) {
				stdErrorLines = append(stdErrorLines, s)
			},
		}),
	)

	app := &testApp{
		config: &fs.Config{
			Dir: utils.Must(os.MkdirTemp("", "fsapp")),
			Hooks: &fs.Hooks{
				DBHooks: &db.Hooks{},
			},
		},
	}
	pluginsDir := filepath.Join(app.config.Dir, "data", "plugins")
	assert.NoError(t, os.MkdirAll(pluginsDir, 0755))

	// Create a new plugin
	// Plugin with no config/init function
	plugin, err := plugins.NewPlugin(createPlugin(t, pluginsDir, "noconfig", pluginNoConfigFunction))
	assert.NoError(t, err)
	assert.NoError(t, plugin.Config(app))
	app = newTestAppFromConfig(t, app.config)
	assert.NoError(t, plugin.Init(app))

	// Plugin with config function
	plugin, err = plugins.NewPlugin(createPlugin(t, pluginsDir, "hello", pluginValid))
	assert.NoError(t, err)
	assert.NotNil(t, plugin)
	assert.Equal(t, "hello", plugin.Name())

	// Call the Config method of the plugin
	err = plugin.Config(app)
	assert.NoError(t, err)
	assert.Len(t, app.config.SystemSchemas, 1)

	app = newTestAppFromConfig(t, app.config)
	err = plugin.Init(app)
	assert.NoError(t, err)

	app.resources.Add(fs.Get("ping", func(ctx fs.Context, _ any) (string, error) {
		return "pong", nil
	}, &fs.Meta{Public: true}))

	app.resources.Hooks = func() *fs.Hooks {
		return app.config.Hooks
	}
	app.server = rr.NewRestfulResolver(&rr.ResolverConfig{
		ResourceManager: app.resources,
		Logger:          logger.CreateMockLogger(true),
	}).Server()

	// Test resolver hooks
	req := httptest.NewRequest("GET", "/ping", nil)
	resp := utils.Must(app.server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 200, resp.StatusCode)
	response := utils.Must(utils.ReadCloserToString(resp.Body))
	assert.Equal(t, `{"data":"pong"}`, response)

	// Test call resource registered by the plugin
	req = httptest.NewRequest("GET", "/hello", nil)
	resp = utils.Must(app.server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 200, resp.StatusCode)
	response = utils.Must(utils.ReadCloserToString(resp.Body))
	assert.Equal(t, `{"data":"world"}`, response)
	assert.True(t, utils.Contains(stdOutLines, `{"'query db in plugin'":"query db in plugin"}`))

	assert.True(t, utils.Contains(stdOutLines, "hook run: onPreResolve"))
	assert.True(t, utils.Contains(stdOutLines, "hook run: onPostResolve"))

	ctx := context.Background()

	// Test db create hooks
	createdRole, err := db.Create[*fs.Role](ctx, app.db, map[string]any{
		"name": "admin",
	})

	assert.NoError(t, err)
	assert.NotNil(t, createdRole)
	assert.True(t, utils.Contains(stdOutLines, "hook run: onPreDBCreate"))
	assert.True(t, utils.Contains(stdOutLines, "hook run: onPostDBCreate"))

	assert.True(t, utils.Contains(stdOutLines, "hook run: onPreDBQuery"))
	assert.True(t, utils.Contains(stdOutLines, "hook run: onPostDBQuery"))

	// Test db update hooks
	updatedRole, err := db.Update[*fs.Role](ctx, app.db, map[string]any{
		"name": "admin updated",
	}, []*db.Predicate{
		db.EQ("id", createdRole.ID),
	})
	assert.NoError(t, err)
	assert.Equal(t, "admin updated", updatedRole[0].Name)

	assert.True(t, utils.Contains(stdOutLines, "hook run: onPreDBUpdate"))
	assert.True(t, utils.Contains(stdOutLines, "hook run: onPostDBUpdate"))

	// Test db exec hooks
	result, err := db.Exec(ctx, app.db, "UPDATE roles SET name = ? WHERE id = ?", "admin", createdRole.ID)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), utils.Must(result.RowsAffected()))
	assert.True(t, utils.Contains(stdOutLines, "hook run: onPreDBExec"))
	assert.True(t, utils.Contains(stdOutLines, "hook run: onPostDBExec"))

	// Test db delete hooks
	deleted, err := db.Delete[*fs.Role](ctx, app.db, []*db.Predicate{
		db.EQ("id", createdRole.ID),
	})
	assert.NoError(t, err)
	assert.Equal(t, 1, deleted)
	assert.True(t, utils.Contains(stdOutLines, "hook run: onPreDBDelete"))
	assert.True(t, utils.Contains(stdOutLines, "hook run: onPostDBDelete"))

	// Call the init method of the plugin
	err = plugin.Init(app)
	assert.NoError(t, err)
	assert.True(t, utils.Contains(stdOutLines, "init plugin"))
}

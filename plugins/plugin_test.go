package plugins_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/fastschema/fastschema"
	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/plugins"
	"github.com/fastschema/qjs"
	"github.com/stretchr/testify/assert"
)

func createApp() *fastschema.App {
	return utils.Must(fastschema.New(&fs.Config{
		Dir: utils.Must(os.MkdirTemp("", "fastschema")),
		DBConfig: &db.Config{
			Driver: "sqlite",
			Name:   ":memory:",
		},
	}))
}

func createPlugin(
	t *testing.T,
	content string,
	runtimeSetup func(rt *qjs.Runtime, inPool bool) error,
) (*fastschema.App, *plugins.Plugin, error) {
	app := createApp()
	pluginDir := filepath.Join(app.Dir(), "data", "plugins", "test-plugin")
	pluginFile := filepath.Join(pluginDir, "plugin.js")
	assert.NoError(t, os.MkdirAll(pluginDir, 0755))
	utils.Must[any](nil, os.WriteFile(
		pluginFile,
		[]byte(content),
		0644,
	))

	plugin, err := plugins.NewPlugin(app, "data/plugins/test-plugin", runtimeSetup, nil)
	return app, plugin, err
}

func createManager(
	t *testing.T,
	content string,
	runtimeSetup func(rt *qjs.Runtime, inPool bool) error,
) (*fastschema.App, *plugins.Manager, error) {
	app := createApp()
	assert.NotNil(t, app)
	pluginDir := filepath.Join(app.Dir(), "data", "plugins", "test-plugin")
	pluginFile := filepath.Join(pluginDir, "plugin.js")
	assert.NoError(t, os.MkdirAll(pluginDir, 0755))

	utils.Must[any](nil, os.WriteFile(
		pluginFile,
		[]byte(content),
		0644,
	))

	manager, err := plugins.NewManager(app, filepath.Dir(pluginDir), runtimeSetup)
	return app, manager, err
}

func TestManager(t *testing.T) {
	// Create manager error due to empty dir
	_, err := plugins.NewManager(nil, "", nil)
	assert.Error(t, err)

	// Create manager error due to glob pattern error
	_, err = plugins.NewManager(nil, "[", nil)
	assert.Error(t, err)

	// Create manager with invalid plugin
	_, _, err = createManager(t, `invalid js`, nil)
	assert.Error(t, err)

	// Manager Config error due to invalid config
	_, manager, err := createManager(t, `export const Config = config => {
		config.Set({
			port: false, // port must be a string
		});
	}
	export default { Config };
	`, nil)
	assert.NoError(t, err)
	assert.Error(t, manager.Config())

	// Manager Init error due to invalid init
	_, manager, err = createManager(t, `export const Init = () => {
		throw new Error('Init error');
	}
	export default { Init };
	`, nil)
	assert.NoError(t, err)
	assert.NoError(t, manager.Config())
	assert.Error(t, manager.Init())

	// Create manager successfully
	_, manager, err = createManager(t, `export const Init = () => {
		console.log($context() !== null);
		console.log($db() !== null);
		$logger().Info('Plugin loaded');
		console.log('Plugin initialized');
	}
	export default { Init };
	`, nil)
	assert.NoError(t, err)
	assert.NoError(t, manager.Config())
	assert.NoError(t, manager.Init())
	plugin, ok := manager.Get("test-plugin")
	assert.True(t, ok)
	assert.NotNil(t, plugin)
}

func TestPlugin(t *testing.T) {
	// Create plugin error: invalid js
	_, _, err := createPlugin(t, `invalid js`, nil)
	assert.Error(t, err)

	// Create plugin error: eval error
	_, _, err = createPlugin(t, `throw new Error('Eval error');`, nil)
	assert.Error(t, err)

	// Create plugin error: runtime setup error
	_, _, err = createPlugin(t, `export default {};`, func(rt *qjs.Runtime, inPool bool) error {
		return os.ErrInvalid
	})
	assert.Error(t, err)

	// Plugin Config is not a function
	_, plugin, _ := createPlugin(t, `
	const Config = "not a function";
	export default { Config };
	`, nil)
	assert.Error(t, plugin.Config())

	// Plugin Init is not a function
	_, plugin, _ = createPlugin(t, `
	const Init = "not a function";
	export default { Init };
	`, nil)
	assert.NoError(t, plugin.Config())
	assert.Error(t, plugin.Init())

	// Plugin does not export Config or Init
	_, plugin, _ = createPlugin(t, `export default {};`, nil)
	assert.NoError(t, plugin.Config())
	assert.NoError(t, plugin.Init())

	// Callback is not a function
	_, plugin, _ = createPlugin(t, `
	const Config = config => {
		config.OnPreDBQuery("not a function");
	}
	const Init = () => {
		$db().Query('SELECT 1');
	};
	export default { Config, Init };
	`, nil)

	assert.Error(t, plugin.Config())
	assert.Error(t, plugin.Init())

	// Callback function does not have a name
	_, plugin, _ = createPlugin(t, `
	const Config = config => {
		config.OnPreDBQuery(function() {});
	}
	const Init = () => {
		$db().Query('SELECT 1');
	};
	export default { Config, Init };
	`, nil)

	assert.Error(t, plugin.Config())
	assert.Error(t, plugin.Init())

	// Callback function name not in default exports
	_, plugin, _ = createPlugin(t, `
	const Config = config => {
		config.OnPreDBQuery(function unknownCallback() {});
	}
	const Init = () => {
		$db().Query('SELECT 1');
	};
	export default { Config, Init };
	`, nil)

	assert.Error(t, plugin.Config())
	assert.Error(t, plugin.Init())

	// Runtime setup function error
	_, plugin, _ = createPlugin(t, `
	const Config = config => {
		config.OnPreDBQuery(PreDBQuery);
	}
	const Init = () => {
		$db().Query($context(), 'SELECT 1');
	};
	const PreDBQuery = (context) => {
		$logger().Info('PreDBQuery called');
	};
	export default { Config, Init, PreDBQuery };
	`, func(rt *qjs.Runtime, inPool bool) error {
		if inPool {
			return os.ErrInvalid
		}
		return nil
	})
	assert.NoError(t, plugin.Config())
	assert.Error(t, plugin.Init())

	// Create plugin error due to qjs initialization error
	func() {
		app := createApp()
		_, err := plugins.NewPlugin(app, "data/plugins/test-plugin", nil, []byte("invalid wasm"))
		assert.Error(t, err)
	}()
}

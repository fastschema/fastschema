package plugins_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/fastschema/fastschema"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/plugins"
	"github.com/stretchr/testify/assert"
)

func TestManager(t *testing.T) {
	// Empty plugins dir
	manager, err := plugins.NewManager("")
	assert.NotNil(t, err)
	assert.Nil(t, manager)

	// With plugins dir
	// Create sample plugin
	pluginsDir := utils.Must(os.MkdirTemp("", "plugins"))
	defer os.RemoveAll(pluginsDir)
	samplePluginDir := filepath.Join(pluginsDir, "sample")
	assert.NoError(t, os.Mkdir(samplePluginDir, 0755))

	// With invalid plugin
	invalidPlugin := []byte(`function a() }`)
	assert.NoError(t, os.WriteFile(filepath.Join(samplePluginDir, "plugin.js"), invalidPlugin, 0644))
	manager, err = plugins.NewManager(pluginsDir)
	assert.NotNil(t, err)
	assert.Nil(t, manager)

	// With valid plugin
	validPlugin := []byte(`
	function Config() {
		console.log("Plugin Config called");
	}

	function Init() {
		console.log("Plugin Init called");
	}`)

	assert.NoError(t, os.WriteFile(filepath.Join(samplePluginDir, "plugin.js"), validPlugin, 0644))
	manager, err = plugins.NewManager(pluginsDir)
	assert.Nil(t, err)
	assert.NotNil(t, manager)

	sampleFsApp := utils.Must(fastschema.New(&fs.Config{
		Dir: utils.Must(os.MkdirTemp("", "fastschema")),
	}))
	assert.NoError(t, manager.Config(sampleFsApp))
	assert.NoError(t, manager.Init(sampleFsApp))
}

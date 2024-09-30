package plugins

import (
	"errors"
	"fmt"
	"path"
	"path/filepath"

	"github.com/fastschema/fastschema/fs"
)

type Manager struct {
	pluginsDir string
	plugins    map[string]*Plugin
}

func NewManager(pluginsDir string) (*Manager, error) {
	if pluginsDir == "" {
		return nil, errors.New("plugins: plugin dir is required")
	}

	manager := &Manager{
		pluginsDir: pluginsDir,
		plugins:    map[string]*Plugin{},
	}

	return manager.load()
}

func (m *Manager) load() (*Manager, error) {
	pluginFiles, err := filepath.Glob(path.Join(m.pluginsDir, "**/plugin.js"))
	if err != nil {
		return nil, err
	}

	for _, pluginFile := range pluginFiles {
		pluginName := filepath.Base(filepath.Dir(pluginFile))
		fmt.Printf("[Plugin] Loading plugin: %s\n", pluginName)
		plugin, err := NewPlugin(pluginFile)
		if err != nil {
			fmt.Printf("[Plugin] Failed to load plugin: %s", pluginName)
			return nil, err
		}

		m.plugins[plugin.name] = plugin
	}

	return m, nil
}

// Config allows plugins to update the app configuration
//
//	It is called before the app is initialized
func (m *Manager) Config(app fs.App) error {
	for _, plugin := range m.plugins {
		if err := plugin.Config(app); err != nil {
			return err
		}
	}
	return nil
}

// Init initializes the plugin
//
//	It is called after the app is initialized and after the Config method is called
func (m *Manager) Init(app fs.App) error {
	for _, plugin := range m.plugins {
		if err := plugin.Init(app); err != nil {
			return err
		}
	}

	return nil
}

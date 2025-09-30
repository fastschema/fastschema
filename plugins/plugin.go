package plugins

import (
	"context"
	"errors"
	"fmt"
	"log"
	"path"
	"path/filepath"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/logger"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/qjs"
)

type AppLike interface {
	fs.Hookable
	DB() db.Client
	Resources() *fs.ResourcesManager
	Config() *fs.Config
	Logger() logger.Logger
	Dir() string
}

type Plugin struct {
	app              AppLike
	appConfig        *AppConfig
	name             string
	file             string
	dir              string
	bytecode         []byte
	runtime          *qjs.Runtime
	pool             *qjs.Pool
	hooks            fs.Hooks
	exportedNames    []string
	runtimeSetupFunc func(rt *qjs.Runtime, inPool bool) error
}

type Manager struct {
	app              AppLike
	pluginsDir       string
	plugins          map[string]*Plugin
	runtimeSetupFunc func(rt *qjs.Runtime, inPool bool) error
}

func NewManager(
	app AppLike,
	dir string,
	runtimeSetupFunc func(rt *qjs.Runtime, inPool bool) error,
) (*Manager, error) {
	if dir == "" {
		return nil, errors.New("plugins: plugin dir is required")
	}

	return (&Manager{
		app,
		dir,
		map[string]*Plugin{},
		runtimeSetupFunc,
	}).load()
}

func (m *Manager) load() (*Manager, error) {
	pluginDirs, err := filepath.Glob(path.Join(m.pluginsDir, "**"))
	if err != nil {
		return nil, err
	}

	for _, pluginDir := range pluginDirs {
		pluginName := filepath.Base(pluginDir)
		pluginDir = path.Join("data", "plugins", pluginDir[len(m.pluginsDir):])

		log.Printf("[Plugin] Loading plugin: %s\n", pluginName)
		plugin, err := NewPlugin(m.app, pluginDir, m.runtimeSetupFunc, nil)
		if err != nil {
			log.Printf("[Plugin] Failed to load plugin: %s\n", pluginName)
			return nil, err
		}

		m.plugins[plugin.name] = plugin
	}

	return m, nil
}

// Get returns the plugin by name
func (m *Manager) Get(name string) (*Plugin, bool) {
	plugin, ok := m.plugins[name]
	return plugin, ok
}

// Config allows plugins to update the app configuration before initialization
func (m *Manager) Config() error {
	for _, plugin := range m.plugins {
		if err := plugin.Config(); err != nil {
			return err
		}
	}

	return nil
}

// Init initializes the plugin after the app and Config have been initialized
func (m *Manager) Init() error {
	for _, plugin := range m.plugins {
		if err := plugin.Init(); err != nil {
			return err
		}
	}

	return nil
}

// NewPlugin creates a new plugin instance
//   - app: the fastschema app instance
//   - dir: the relative path to the plugin directory
func NewPlugin(
	app AppLike,
	pluginDir string,
	runtimeSetupFunc func(rt *qjs.Runtime, inPool bool) error,
	qjsWasmBytes []byte,
) (p *Plugin, err error) {
	name := filepath.Base(pluginDir)
	pluginFile := filepath.Join(pluginDir, "plugin.js")
	p = &Plugin{
		app:              app,
		name:             name,
		file:             pluginFile,
		dir:              pluginDir,
		hooks:            fs.Hooks{},
		runtimeSetupFunc: runtimeSetupFunc,
	}

	if p.runtime, err = qjs.New(&qjs.Option{
		CWD:              app.Dir(),
		QuickJSWasmBytes: qjsWasmBytes,
	}); err != nil {
		return nil, err
	}

	// Compile plugin file to bytecode
	if p.bytecode, err = p.runtime.Compile(pluginFile, qjs.TypeModule()); err != nil {
		return nil, err
	}

	// Create pool
	p.pool = qjs.NewPool(10, nil, func(rt *qjs.Runtime) error {
		return p.EvalPluginFile(rt, true)
	})

	// Eval to get exported names
	if err := p.EvalPluginFile(p.runtime, false); err != nil {
		return nil, err
	}

	defaultExports := p.runtime.Context().Global().GetPropertyStr("defaultExports")
	for _, prop := range defaultExports.GetOwnProperties() {
		p.exportedNames = append(p.exportedNames, prop.String())
	}

	return p, nil
}

func (p *Plugin) Name() string {
	return p.name
}

func (p *Plugin) EvalPluginFile(rt *qjs.Runtime, inPool bool) (err error) {
	result, err := rt.Eval(p.file, qjs.Bytecode(p.bytecode), qjs.TypeModule())
	if err != nil {
		return err
	}

	if p.runtimeSetupFunc != nil {
		if err := p.runtimeSetupFunc(rt, inPool); err != nil {
			return fmt.Errorf("plugin %s runtime setup error: %w", p.name, err)
		}
	}

	rt.Context().SetFunc("$context", func(this *qjs.This) (*qjs.Value, error) {
		return rt.Context().NewProxyValue(context.Background()), nil
	})

	rt.Context().SetFunc("$db", func(this *qjs.This) (*qjs.Value, error) {
		return qjs.ToJSValue(rt.Context(), NewDB(p.app.DB))
	})

	rt.Context().SetFunc("$logger", func(this *qjs.This) (*qjs.Value, error) {
		return qjs.ToJSValue(this.Context(), p.app.Logger())
	})
	rt.Context().Global().SetPropertyStr("defaultExports", result)
	return nil
}

func (p *Plugin) WithJSFuncName(v *qjs.Value, cb func(jsFuncName string)) error {
	if v == nil || !v.IsFunction() {
		return fmt.Errorf("JS callback is nil or not a function: %s", v)
	}

	jsFuncName := v.GetPropertyStr("name").String()
	if jsFuncName == "" {
		return fmt.Errorf("JS callback name is empty")
	}

	if !utils.Contains(p.exportedNames, jsFuncName) {
		return fmt.Errorf("%s/plugin.js default export does not contain: '%s'", p.name, jsFuncName)
	}

	cb(jsFuncName)
	return nil
}

func (p *Plugin) InvokeJsFunc(jsFuncName string, args ...any) (*qjs.Value, error) {
	rt, err := p.pool.Get()
	if err != nil {
		return nil, err
	}
	defer p.pool.Put(rt)

	defaultExports := rt.Context().Global().GetPropertyStr("defaultExports")
	return defaultExports.Invoke(jsFuncName, args...)
}

func (p *Plugin) Config() (err error) {
	defaultExports := p.runtime.Context().Global().GetPropertyStr("defaultExports")
	jsConfig := defaultExports.GetPropertyStr("Config")
	if jsConfig.IsUndefined() {
		return nil
	}

	if !jsConfig.IsFunction() {
		return fmt.Errorf("Plugin export 'Config' is not a function")
	}

	p.appConfig = NewAppConfig(p, p.app, nil)
	_, err = defaultExports.Invoke("Config", p.appConfig)
	return err
}

func (p *Plugin) Init() (err error) {
	defaultExports := p.runtime.Context().Global().GetPropertyStr("defaultExports")
	jsConfig := defaultExports.GetPropertyStr("Init")
	if jsConfig.IsUndefined() {
		return nil
	}

	if !jsConfig.IsFunction() {
		return fmt.Errorf("Plugin export 'Init' is not a function")
	}

	_, err = defaultExports.Invoke("Init", map[string]any{
		"resources": NewResource(p.app.Resources().Resource, p),
	})
	return err
}

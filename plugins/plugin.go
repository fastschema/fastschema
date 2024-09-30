package plugins

import (
	"context"
	"log"
	"path/filepath"
	"strings"

	"github.com/dop251/goja_nodejs/console"
	gojarequire "github.com/dop251/goja_nodejs/require"
	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/logger"
)

var Require = new(gojarequire.Registry)

func init() {
	printer := console.StdPrinter{
		StdoutPrint: func(s string) {
			log.Print(s)
		},
		StderrPrint: func(s string) {
			log.Print(s)
		},
	}
	Require.RegisterNativeModule(console.ModuleName, console.RequireWithPrinter(printer))
}

type GetVMSet = func() map[string]any

type AppLike interface {
	fs.Hookable
	DB() db.Client
	Resources() *fs.ResourcesManager
	Config() *fs.Config
	Logger() logger.Logger
}

type Plugin struct {
	name    string
	file    string
	program *Program
	hooks   fs.Hooks
}

func (p *Plugin) Name() string {
	return p.name
}

func NewPlugin(file string) (*Plugin, error) {
	name := filepath.Base(filepath.Dir(file))
	program, _, err := CreateGoJaProgram(file, nil)
	if err != nil {
		return nil, err
	}

	return &Plugin{
		name: name,
		file: file,
		program: NewProgram(
			program,
			"plugin."+name,
		),
		hooks: fs.Hooks{},
	}, nil
}

func (p *Plugin) Config(app AppLike) (err error) {
	if _, err = p.program.CallFunc("Config", nil, NewConfigActions(
		app,
		p.program,
		nil,
	)); err != nil && strings.Contains(err.Error(), "is not found") {
		err = nil
	}

	return
}

func (p *Plugin) Init(app AppLike) (err error) {
	set := map[string]any{
		"$context": context.Background,
		"$logger":  app.Logger,
		"$db": func() *DB {
			return NewDB(app.DB())
		},
	}

	if _, err = p.program.CallFunc("Init", set, map[string]any{
		"resources": NewResource(
			app.Resources().Resource,
			p.program,
			set,
		),
	}); err != nil && strings.Contains(err.Error(), "is not found") {
		err = nil
	}

	return
}

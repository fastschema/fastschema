package plugins

import (
	"context"
	"database/sql"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/entity"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/schema"
	"github.com/fastschema/qjs"
	"github.com/mitchellh/mapstructure"
)

type AppConfig struct {
	config *fs.Config
	plugin *Plugin
	app    AppLike
}

func NewAppConfig(plugin *Plugin, app AppLike, set map[string]any) *AppConfig {
	return &AppConfig{
		config: app.Config(),
		plugin: plugin,
		app:    app,
	}
}

func (c *AppConfig) Set(config map[string]any) error {
	decoder, _ := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName: "json",
		Result:  c.config,
	})

	if err := decoder.Decode(config); err != nil {
		return err
	}

	return nil
}

func (c *AppConfig) AddSchemas(schemas ...*schema.Schema) {
	for _, s := range schemas {
		c.config.SystemSchemas = append(c.config.SystemSchemas, s)
	}
}

func (p *AppConfig) OnPreDBQuery(value *qjs.Value) (err error) {
	return p.plugin.WithJSFuncName(value, func(jsFuncName string) {
		p.app.OnPreDBQuery(func(c context.Context, option *db.QueryOption) (err error) {
			_, err = p.plugin.InvokeJsFunc(jsFuncName, c, option)
			return
		})
	})
}

func (p *AppConfig) OnPostDBQuery(value *qjs.Value) error {
	return p.plugin.WithJSFuncName(value, func(jsFuncName string) {
		p.app.OnPostDBQuery(func(
			c context.Context,
			option *db.QueryOption,
			entities []*entity.Entity,
		) (_ []*entity.Entity, err error) {
			_, err = p.plugin.InvokeJsFunc(jsFuncName, c, option, entities)
			return entities, err
		})
	})
}

func (p *AppConfig) OnPreDBExec(value *qjs.Value) error {
	return p.plugin.WithJSFuncName(value, func(jsFuncName string) {
		p.app.OnPreDBExec(func(
			c context.Context,
			option *db.QueryOption,
		) (err error) {
			_, err = p.plugin.InvokeJsFunc(jsFuncName, c, option)
			return err
		})
	})
}

func (p *AppConfig) OnPostDBExec(value *qjs.Value) error {
	return p.plugin.WithJSFuncName(value, func(jsFuncName string) {
		p.app.OnPostDBExec(func(
			c context.Context,
			option *db.QueryOption,
			result sql.Result,
		) (err error) {
			_, err = p.plugin.InvokeJsFunc(jsFuncName, c, option, result)
			return err
		})
	})
}

func (p *AppConfig) OnPreDBCreate(value *qjs.Value) error {
	return p.plugin.WithJSFuncName(value, func(jsFuncName string) {
		p.app.OnPreDBCreate(func(
			c context.Context,
			schema *schema.Schema,
			createData *entity.Entity,
		) (err error) {
			_, err = p.plugin.InvokeJsFunc(jsFuncName, c, schema, createData)
			return err
		})
	})
}

func (p *AppConfig) OnPostDBCreate(value *qjs.Value) error {
	return p.plugin.WithJSFuncName(value, func(jsFuncName string) {
		p.app.OnPostDBCreate(func(
			c context.Context,
			schema *schema.Schema,
			createData *entity.Entity,
			createdID uint64,
		) (err error) {
			_, err = p.plugin.InvokeJsFunc(
				jsFuncName,
				c,
				schema,
				createData,
				createdID,
			)
			return err
		})
	})
}

func (p *AppConfig) OnPreDBUpdate(value *qjs.Value) error {
	return p.plugin.WithJSFuncName(value, func(jsFuncName string) {
		p.app.OnPreDBUpdate(func(
			c context.Context,
			schema *schema.Schema,
			predicates *[]*db.Predicate,
			updateData *entity.Entity,
		) (err error) {
			_, err = p.plugin.InvokeJsFunc(
				jsFuncName,
				c,
				schema,
				predicates,
				updateData,
			)
			return err
		})
	})
}

func (p *AppConfig) OnPostDBUpdate(value *qjs.Value) error {
	return p.plugin.WithJSFuncName(value, func(jsFuncName string) {
		p.app.OnPostDBUpdate(func(
			c context.Context,
			schema *schema.Schema,
			predicates *[]*db.Predicate,
			updateData *entity.Entity,
			originalEntities []*entity.Entity,
			affected int,
		) (err error) {
			_, err = p.plugin.InvokeJsFunc(
				jsFuncName,
				c,
				schema,
				predicates,
				updateData,
				originalEntities,
				affected,
			)
			return err
		})
	})
}

func (p *AppConfig) OnPreDBDelete(value *qjs.Value) error {
	return p.plugin.WithJSFuncName(value, func(jsFuncName string) {
		p.app.OnPreDBDelete(func(
			c context.Context,
			schema *schema.Schema,
			predicates *[]*db.Predicate,
		) (err error) {
			_, err = p.plugin.InvokeJsFunc(jsFuncName, c, schema, predicates)
			return err
		})
	})
}

func (p *AppConfig) OnPostDBDelete(value *qjs.Value) error {
	return p.plugin.WithJSFuncName(value, func(jsFuncName string) {
		p.app.OnPostDBDelete(func(
			c context.Context,
			schema *schema.Schema,
			predicates *[]*db.Predicate,
			originalEntities []*entity.Entity,
			affected int,
		) (err error) {
			_, err = p.plugin.InvokeJsFunc(
				jsFuncName,
				c,
				schema,
				predicates,
				originalEntities,
				affected,
			)
			return err
		})
	})
}

func (p *AppConfig) OnPreResolve(value *qjs.Value) (err error) {
	return p.plugin.WithJSFuncName(value, func(jsFuncName string) {
		p.app.OnPreResolve(func(c fs.Context) error {
			_, err := p.plugin.InvokeJsFunc(jsFuncName, c)
			return err
		})
	})
}

func (p *AppConfig) OnPostResolve(value *qjs.Value) (err error) {
	return p.plugin.WithJSFuncName(value, func(jsFuncName string) {
		p.app.OnPostResolve(func(c fs.Context) error {
			_, err := p.plugin.InvokeJsFunc(jsFuncName, c)
			return err
		})
	})
}

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

func (ac *AppConfig) Set(config map[string]any) error {
	decoder, _ := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName: "json",
		Result:  ac.config,
	})

	if err := decoder.Decode(config); err != nil {
		return err
	}

	return nil
}

func (ac *AppConfig) AddSchemas(schemas ...*schema.Schema) {
	for _, s := range schemas {
		ac.config.SystemSchemas = append(ac.config.SystemSchemas, s)
	}
}

func (ac *AppConfig) OnPreDBQuery(value *qjs.Value) (err error) {
	return ac.plugin.WithJSFuncName(value, func(jsFuncName string) {
		ac.app.OnPreDBQuery(func(c context.Context, option *db.QueryOption) (err error) {
			_, err = ac.plugin.InvokeJsFunc(jsFuncName, c, option)
			return
		})
	})
}

func (ac *AppConfig) OnPostDBQuery(value *qjs.Value) error {
	return ac.plugin.WithJSFuncName(value, func(jsFuncName string) {
		ac.app.OnPostDBQuery(func(
			c context.Context,
			option *db.QueryOption,
			entities []*entity.Entity,
		) (_ []*entity.Entity, err error) {
			_, err = ac.plugin.InvokeJsFunc(jsFuncName, c, option, entities)
			return entities, err
		})
	})
}

func (ac *AppConfig) OnPreDBExec(value *qjs.Value) error {
	return ac.plugin.WithJSFuncName(value, func(jsFuncName string) {
		ac.app.OnPreDBExec(func(
			c context.Context,
			option *db.QueryOption,
		) (err error) {
			_, err = ac.plugin.InvokeJsFunc(jsFuncName, c, option)
			return err
		})
	})
}

func (ac *AppConfig) OnPostDBExec(value *qjs.Value) error {
	return ac.plugin.WithJSFuncName(value, func(jsFuncName string) {
		ac.app.OnPostDBExec(func(
			c context.Context,
			option *db.QueryOption,
			result sql.Result,
		) (err error) {
			_, err = ac.plugin.InvokeJsFunc(jsFuncName, c, option, result)
			return err
		})
	})
}

func (ac *AppConfig) OnPreDBCreate(value *qjs.Value) error {
	return ac.plugin.WithJSFuncName(value, func(jsFuncName string) {
		ac.app.OnPreDBCreate(func(
			c context.Context,
			schema *schema.Schema,
			createData *entity.Entity,
		) (err error) {
			_, err = ac.plugin.InvokeJsFunc(jsFuncName, c, schema, createData)
			return err
		})
	})
}

func (ac *AppConfig) OnPostDBCreate(value *qjs.Value) error {
	return ac.plugin.WithJSFuncName(value, func(jsFuncName string) {
		ac.app.OnPostDBCreate(func(
			c context.Context,
			schema *schema.Schema,
			createData *entity.Entity,
			createdID uint64,
		) (err error) {
			_, err = ac.plugin.InvokeJsFunc(
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

func (ac *AppConfig) OnPreDBUpdate(value *qjs.Value) error {
	return ac.plugin.WithJSFuncName(value, func(jsFuncName string) {
		ac.app.OnPreDBUpdate(func(
			c context.Context,
			schema *schema.Schema,
			predicates *[]*db.Predicate,
			updateData *entity.Entity,
		) (err error) {
			_, err = ac.plugin.InvokeJsFunc(
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

func (ac *AppConfig) OnPostDBUpdate(value *qjs.Value) error {
	return ac.plugin.WithJSFuncName(value, func(jsFuncName string) {
		ac.app.OnPostDBUpdate(func(
			c context.Context,
			schema *schema.Schema,
			predicates *[]*db.Predicate,
			updateData *entity.Entity,
			originalEntities []*entity.Entity,
			affected int,
		) (err error) {
			_, err = ac.plugin.InvokeJsFunc(
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

func (ac *AppConfig) OnPreDBDelete(value *qjs.Value) error {
	return ac.plugin.WithJSFuncName(value, func(jsFuncName string) {
		ac.app.OnPreDBDelete(func(
			c context.Context,
			schema *schema.Schema,
			predicates *[]*db.Predicate,
		) (err error) {
			_, err = ac.plugin.InvokeJsFunc(jsFuncName, c, schema, predicates)
			return err
		})
	})
}

func (ac *AppConfig) OnPostDBDelete(value *qjs.Value) error {
	return ac.plugin.WithJSFuncName(value, func(jsFuncName string) {
		ac.app.OnPostDBDelete(func(
			c context.Context,
			schema *schema.Schema,
			predicates *[]*db.Predicate,
			originalEntities []*entity.Entity,
			affected int,
		) (err error) {
			_, err = ac.plugin.InvokeJsFunc(
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

func (ac *AppConfig) OnPreResolve(value *qjs.Value) (err error) {
	return ac.plugin.WithJSFuncName(value, func(jsFuncName string) {
		ac.app.OnPreResolve(func(c fs.Context) error {
			_, err := ac.plugin.InvokeJsFunc(jsFuncName, c)
			return err
		})
	})
}

func (ac *AppConfig) OnPostResolve(value *qjs.Value) (err error) {
	return ac.plugin.WithJSFuncName(value, func(jsFuncName string) {
		ac.app.OnPostResolve(func(c fs.Context) error {
			_, err := ac.plugin.InvokeJsFunc(jsFuncName, c)
			return err
		})
	})
}

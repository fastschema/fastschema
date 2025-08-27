package plugins

import (
	"context"
	"database/sql"

	"github.com/dop251/goja"
	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/entity"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/schema"
)

type ConfigActions struct {
	*fs.Config `json:"config"`

	app     AppLike
	program *Program
	set     map[string]any
}

func NewConfigActions(app AppLike, program *Program, set map[string]any) *ConfigActions {
	return &ConfigActions{
		Config:  app.Config(),
		app:     app,
		program: program,
		set:     set,
	}
}

func (p *ConfigActions) AddSchemas(schemas ...map[string]any) error {
	newSchemas := []any{}
	for _, s := range schemas {
		ss, err := schema.NewSchemaFromMap(s)
		if err != nil {
			return err
		}

		newSchemas = append(newSchemas, ss)
	}

	p.SystemSchemas = append(p.SystemSchemas, newSchemas...)

	return nil
}

func (p *ConfigActions) createResolveHook(hook goja.Value) (fs.ResolveHook, error) {
	fnName, err := p.program.VerifyJsFunc(hook)
	if err != nil {
		return nil, err
	}

	return func(c fs.Context) (err error) {
		_, err = p.program.CallFunc(fnName, p.set, c)
		return
	}, nil
}

func (p *ConfigActions) OnPreResolve(hooks ...goja.Value) (err error) {
	for _, hook := range hooks {
		hookFn, err := p.createResolveHook(hook)
		if err != nil {
			return err
		}

		p.app.OnPreResolve(hookFn)
	}

	return nil
}

func (p *ConfigActions) OnPostResolve(hooks ...goja.Value) (err error) {
	for _, hook := range hooks {
		hookFn, err := p.createResolveHook(hook)
		if err != nil {
			return err
		}

		p.app.OnPostResolve(hookFn)
	}

	return nil
}

func (p *ConfigActions) OnPreDBQuery(hooks ...goja.Value) (err error) {
	for _, hook := range hooks {
		if err := p.program.WithFuncName(hook, func(fnName string) {
			p.app.OnPreDBQuery(func(c context.Context, option *db.QueryOption) (err error) {
				_, err = p.program.CallFunc(fnName, p.set, c, option)
				return
			})
		}); err != nil {
			return err
		}
	}

	return nil
}

func (p *ConfigActions) OnPostDBQuery(hooks ...goja.Value) error {
	for _, hook := range hooks {
		if err := p.program.WithFuncName(hook, func(fnName string) {
			p.app.OnPostDBQuery(func(
				c context.Context,
				option *db.QueryOption,
				entities []*entity.Entity,
			) (_ []*entity.Entity, err error) {
				_, err = p.program.CallFunc(fnName, p.set, c, option, entities)
				return entities, err
			})
		}); err != nil {
			return err
		}
	}

	return nil
}

func (p *ConfigActions) OnPreDBExec(hooks ...goja.Value) error {
	for _, hook := range hooks {
		if err := p.program.WithFuncName(hook, func(fnName string) {
			p.app.OnPreDBExec(func(c context.Context, option *db.QueryOption) (err error) {
				_, err = p.program.CallFunc(fnName, p.set, c, option)
				return
			})
		}); err != nil {
			return err
		}
	}

	return nil
}

func (p *ConfigActions) OnPostDBExec(hooks ...goja.Value) error {
	for _, hook := range hooks {
		if err := p.program.WithFuncName(hook, func(fnName string) {
			p.app.OnPostDBExec(func(
				c context.Context,
				option *db.QueryOption,
				result sql.Result,
			) (err error) {
				_, err = p.program.CallFunc(fnName, p.set, c, option, result)
				return
			})
		}); err != nil {
			return err
		}
	}

	return nil
}

func (p *ConfigActions) OnPreDBCreate(hooks ...goja.Value) error {
	for _, hook := range hooks {
		if err := p.program.WithFuncName(hook, func(fnName string) {
			p.app.OnPreDBCreate(func(c context.Context, schema *schema.Schema, createData *entity.Entity) (err error) {
				_, err = p.program.CallFunc(fnName, p.set, c, schema, createData)
				return
			})
		}); err != nil {
			return err
		}
	}

	return nil
}

func (p *ConfigActions) OnPostDBCreate(hooks ...goja.Value) error {
	for _, hook := range hooks {
		if err := p.program.WithFuncName(hook, func(fnName string) {
			p.app.OnPostDBCreate(func(
				c context.Context,
				schema *schema.Schema,
				createData *entity.Entity,
				createdID uint64,
			) (err error) {
				_, err = p.program.CallFunc(fnName, p.set, c, schema, createData, createdID)
				return
			})
		}); err != nil {
			return err
		}
	}

	return nil
}

func (p *ConfigActions) OnPreDBUpdate(hooks ...goja.Value) error {
	for _, hook := range hooks {
		if err := p.program.WithFuncName(hook, func(fnName string) {
			p.app.OnPreDBUpdate(func(
				c context.Context,
				schema *schema.Schema,
				predicates *[]*db.Predicate,
				updateData *entity.Entity,
			) (err error) {
				_, err = p.program.CallFunc(fnName, p.set, c, schema, predicates, updateData)
				return
			})
		}); err != nil {
			return err
		}
	}

	return nil
}

func (p *ConfigActions) OnPostDBUpdate(hooks ...goja.Value) error {
	for _, hook := range hooks {
		if err := p.program.WithFuncName(hook, func(fnName string) {
			p.app.OnPostDBUpdate(func(
				c context.Context,
				schema *schema.Schema,
				predicates *[]*db.Predicate,
				updateData *entity.Entity,
				originalEntities []*entity.Entity,
				affected int,
			) (err error) {
				_, err = p.program.CallFunc(fnName, p.set, c, schema, predicates, updateData, originalEntities, affected)
				return
			})
		}); err != nil {
			return err
		}
	}

	return nil
}

func (p *ConfigActions) OnPreDBDelete(hooks ...goja.Value) error {
	for _, hook := range hooks {
		if err := p.program.WithFuncName(hook, func(fnName string) {
			p.app.OnPreDBDelete(func(
				c context.Context,
				schema *schema.Schema,
				predicates *[]*db.Predicate,
			) (err error) {
				_, err = p.program.CallFunc(fnName, p.set, c, schema, predicates)
				return
			})
		}); err != nil {
			return err
		}
	}

	return nil
}

func (p *ConfigActions) OnPostDBDelete(hooks ...goja.Value) error {
	for _, hook := range hooks {
		if err := p.program.WithFuncName(hook, func(fnName string) {
			p.app.OnPostDBDelete(func(
				c context.Context,
				schema *schema.Schema,
				predicates *[]*db.Predicate,
				originalEntities []*entity.Entity,
				affected int,
			) (err error) {
				_, err = p.program.CallFunc(fnName, p.set, c, schema, predicates, originalEntities, affected)
				return
			})
		}); err != nil {
			return err
		}
	}

	return nil
}

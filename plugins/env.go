package plugins

import (
	"context"
	"database/sql"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/entity"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/schema"
	"github.com/fastschema/qjs"
)

// Database access helper for plugins

type DB struct {
	getDB func() db.Client
}

type Builder struct {
	builder *db.QueryBuilder[*entity.Entity]
	client  db.Client
	schema  *schema.Schema
}

func NewDB(db func() db.Client) *DB {
	return &DB{getDB: db}
}

func (d *DB) Close() error {
	return d.getDB().Close()
}

func (d *DB) Query(ctx context.Context, query string, args []any) ([]*entity.Entity, error) {
	return db.Query[*entity.Entity](ctx, d.getDB(), query, args...)
}

func (d *DB) Exec(ctx context.Context, query string, args []any) (sql.Result, error) {
	return db.Exec(ctx, d.getDB(), query, args...)
}

func (d *DB) Create(ctx context.Context, schemaName string, value map[string]any) (*entity.Entity, error) {
	return db.Create[*entity.Entity](ctx, d.getDB(), value, schemaName)
}

func (d *DB) Builder(schemaName string) (*Builder, error) {
	client := d.getDB()
	builder := db.Builder[*entity.Entity](client, schemaName)
	s, err := client.SchemaBuilder().Schema(schemaName)
	if err != nil {
		return nil, err
	}

	return &Builder{builder, client, s}, nil
}

func (b *Builder) Create(ctx context.Context, value map[string]any) (*entity.Entity, error) {
	return b.builder.Create(ctx, value)
}

func (b *Builder) Where(predicates ...map[string]any) (*Builder, error) {
	sb := b.client.SchemaBuilder()
	preds := []*db.Predicate{}
	for _, p := range predicates {
		pred, err := db.CreatePredicatesFromFilterMap(sb, b.schema, p)
		if err != nil {
			return nil, err
		}

		preds = append(preds, pred...)
	}

	b.builder.Where(preds...)
	return b, nil
}

func (b *Builder) Limit(limit uint) *Builder {
	b.builder.Limit(limit)
	return b
}

func (b *Builder) Offset(offset uint) *Builder {
	b.builder.Offset(offset)
	return b
}

func (b *Builder) Select(fields []string) *Builder {
	b.builder.Select(fields...)
	return b
}

func (b *Builder) Count(ctx context.Context) (int, error) {
	return b.builder.Count(ctx)
}

func (b *Builder) Get(ctx context.Context) ([]*entity.Entity, error) {
	return b.builder.Get(ctx)
}

func (b *Builder) First(ctx context.Context) (*entity.Entity, error) {
	return b.builder.First(ctx)
}

func (b *Builder) Only(ctx context.Context) (*entity.Entity, error) {
	return b.builder.Only(ctx)
}

func (b *Builder) Update(ctx context.Context, value map[string]any) ([]*entity.Entity, error) {
	return b.builder.Update(ctx, value)
}

func (b *Builder) Delete(ctx context.Context) (int, error) {
	return b.builder.Delete(ctx)
}

func (d *DB) Tx(ctx context.Context) (*DB, error) {
	tx, err := d.getDB().Tx(ctx)
	if err != nil {
		return nil, err
	}

	return &DB{getDB: func() db.Client { return tx }}, nil
}

func (d *DB) Commit() error {
	return d.getDB().Commit()
}

func (d *DB) Rollback() error {
	return d.getDB().Rollback()
}

// Resources access helper for plugins
type Resource struct {
	resource *fs.Resource
	plugin   *Plugin
}

func NewResource(resource *fs.Resource, plugin *Plugin) *Resource {
	return &Resource{resource, plugin}
}

func (r *Resource) Find(id string) *Resource {
	resource := r.resource.Find(id)
	if resource == nil {
		return nil
	}

	return &Resource{resource, r.plugin}
}

func (r *Resource) Group(name string, metas ...*fs.Meta) *Resource {
	return &Resource{r.resource.Group(name, metas...), r.plugin}
}

func (r *Resource) Add(value *qjs.Value, metas ...*fs.Meta) (*Resource, error) {
	return r, r.plugin.WithJSFuncName(value, func(jsFuncName string) {
		r.resource.Add(fs.NewResource(jsFuncName, func(c fs.Context, _ any) (_ any, err error) {
			result, err := r.plugin.InvokeJsFunc(jsFuncName, c)
			if err != nil {
				return nil, err
			}

			if result.IsPromise() {
				result, err = result.Await()
				if err != nil {
					return nil, err
				}
			}

			return qjs.JsValueToGo[any](result)
		}, metas...))
	})
}

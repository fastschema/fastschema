package plugins

import (
	"context"
	"database/sql"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/entity"
	"github.com/fastschema/fastschema/schema"
)

type DB struct {
	db db.Client
}

func NewDB(db db.Client) *DB {
	return &DB{db: db}
}

func (d *DB) Builder(schemaName string) (*Builder, error) {
	builder := db.Builder[*entity.Entity](d.db, schemaName)
	s, err := d.db.SchemaBuilder().Schema(schemaName)
	if err != nil {
		return nil, err
	}

	return &Builder{
		QueryBuilder: builder,
		client:       d.db,
		schema:       s,
	}, nil
}

type Builder struct {
	*db.QueryBuilder[*entity.Entity]

	client db.Client
	schema *schema.Schema
}

func (b *Builder) Where(predicates ...map[string]any) (*Builder, error) {
	preds := []*db.Predicate{}
	for _, p := range predicates {
		pred, err := db.CreatePredicatesFromFilterMap(
			b.client.SchemaBuilder(),
			b.schema,
			p,
		)
		if err != nil {
			return nil, err
		}

		preds = append(preds, pred...)
	}

	b.QueryBuilder.Where(preds...)

	return b, nil
}

func (d *DB) Query(ctx context.Context, query string, args ...any) ([]*entity.Entity, error) {
	return db.Query[*entity.Entity](ctx, d.db, query, args...)
}

func (d *DB) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return db.Exec(ctx, d.db, query, args...)
}

func (d *DB) Tx(ctx context.Context) (*DB, error) {
	tx, err := d.db.Tx(ctx)
	if err != nil {
		return nil, err
	}

	return &DB{db: tx}, nil
}

func (d *DB) Commit() error {
	return d.db.Commit()
}

func (d *DB) Rollback() error {
	return d.db.Rollback()
}

package expr

import (
	"context"
	"database/sql"

	"github.com/fastschema/fastschema/entity"
)

type DBLike interface {
	Exec(ctx context.Context, query string, args ...any) (sql.Result, error)
	Query(ctx context.Context, query string, args ...any) ([]*entity.Entity, error)
}

type DB struct {
	db func() DBLike
}

func NewDB(db func() DBLike) *DB {
	return &DB{db: db}
}

func (d *DB) Query(ctx context.Context, query string, args ...any) ([]*entity.Entity, error) {
	return d.db().Query(ctx, query, args...)
}

func (d *DB) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return d.db().Exec(ctx, query, args...)
}

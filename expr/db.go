package expr

import (
	"context"
	"database/sql"

	"github.com/fastschema/fastschema/entity"
)

// DBLike defines the interface for database operations accessible within expressions.
type DBLike interface {
	Exec(ctx context.Context, query string, args ...any) (sql.Result, error)
	Query(ctx context.Context, query string, args ...any) ([]*entity.Entity, error)
}

// DB wraps a DBLike provider to support lazy execution or context-aware DB retrieval.
type DB struct {
	db func() DBLike
}

// NewDB creates a new DB wrapper.
func NewDB(db func() DBLike) *DB {
	return &DB{db: db}
}

// Query executes a query that returns rows, typically SELECT.
func (d *DB) Query(ctx context.Context, query string, args ...any) ([]*entity.Entity, error) {
	return d.db().Query(ctx, query, args...)
}

// Exec executes a query without returning result rows, typically INSERT/UPDATE/DELETE.
func (d *DB) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return d.db().Exec(ctx, query, args...)
}

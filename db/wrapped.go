package db

import (
	"context"
	"database/sql"

	"github.com/fastschema/fastschema/entity"
	"github.com/fastschema/fastschema/schema"
)

// WrappedClient implements the Client interface with a wrapped client.
type WrappedClient struct {
	client func() Client
}

func NewWrappedClient(client func() Client) *WrappedClient {
	return &WrappedClient{client: client}
}

var _ Client = (*WrappedClient)(nil)

func (n *WrappedClient) Dialect() string {
	return n.client().Dialect()
}

func (n *WrappedClient) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return n.client().Exec(ctx, query, args...)
}

func (n *WrappedClient) Query(ctx context.Context, query string, args ...any) ([]*entity.Entity, error) {
	return n.client().Query(ctx, query, args...)
}

func (n *WrappedClient) Rollback() error {
	return n.client().Rollback()
}

func (n *WrappedClient) Commit() error {
	return n.client().Commit()
}

func (n *WrappedClient) Tx(ctx context.Context) (Client, error) {
	return n.client().Tx(ctx)
}

func (n *WrappedClient) IsTx() bool {
	return n.client().IsTx()
}

func (n *WrappedClient) Model(model any) (Model, error) {
	return n.client().Model(model)
}

func (n *WrappedClient) Close() error {
	return n.client().Close()
}

func (n *WrappedClient) SchemaBuilder() *schema.Builder {
	return n.client().SchemaBuilder()
}

func (n *WrappedClient) Reload(
	ctx context.Context,
	newSchemaBuilder *schema.Builder,
	migration *Migration,
	disableForeignKeys bool,
	enableMigrations ...bool,
) (Client, error) {
	return n.client().Reload(ctx, newSchemaBuilder, migration, disableForeignKeys, enableMigrations...)
}

func (n *WrappedClient) DB() *sql.DB {
	return n.client().DB()
}

func (n *WrappedClient) Config() *Config {
	return n.client().Config()
}

func (n *WrappedClient) Hooks() *Hooks {
	return n.client().Hooks()
}

// NoopClient implements the Client interface with no-op methods.
type NoopClient struct{}

var _ Client = (*NoopClient)(nil)

func (n *NoopClient) Dialect() string {
	return ""
}

func (n *NoopClient) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return nil, nil
}

func (n *NoopClient) Query(ctx context.Context, query string, args ...any) ([]*entity.Entity, error) {
	return nil, nil
}

func (n *NoopClient) Rollback() error {
	return nil
}

func (n *NoopClient) Commit() error {
	return nil
}

func (n *NoopClient) Tx(ctx context.Context) (Client, error) {
	return nil, nil
}

func (n *NoopClient) IsTx() bool {
	return false
}

func (n *NoopClient) Model(model any) (Model, error) {
	return nil, nil
}

func (n *NoopClient) Close() error {
	return nil
}

func (n *NoopClient) SchemaBuilder() *schema.Builder {
	return nil
}

func (n *NoopClient) Reload(
	ctx context.Context,
	newSchemaBuilder *schema.Builder,
	migration *Migration,
	disableForeignKeys bool,
	enableMigrations ...bool,
) (Client, error) {
	return nil, nil
}

func (n *NoopClient) DB() *sql.DB {
	return nil
}

func (n *NoopClient) Config() *Config {
	return nil
}

func (n *NoopClient) Hooks() *Hooks {
	return nil
}

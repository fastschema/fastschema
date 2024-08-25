package db

import (
	"context"
	"database/sql"
	"database/sql/driver"

	_ "github.com/DATA-DOG/go-sqlmock"
	"github.com/fastschema/fastschema/logger"
	"github.com/fastschema/fastschema/schema"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "modernc.org/sqlite"
)

// SupportDrivers returns list of supported drivers.
var SupportDrivers = []string{"mysql", "pgx", "sqlite"}

type PostDBGet = func(
	query *QueryOption,
	entities []*schema.Entity,
) ([]*schema.Entity, error)

type PostDBCreate = func(
	schema *schema.Schema,
	id uint64,
	dataCreate *schema.Entity,
) error

type PostDBUpdate = func(
	schema *schema.Schema,
	predicates []*Predicate,
	updateData *schema.Entity,
	originalEntities []*schema.Entity,
	affected int,
) error

type PostDBDelete = func(
	schema *schema.Schema,
	predicates []*Predicate,
	originalEntities []*schema.Entity,
	affected int,
) error

type PreDBGet = func(
	query *QueryOption,
) error

type PreDBCreate = func(
	schema *schema.Schema,
	dataCreate *schema.Entity,
) error

type PreDBUpdate = func(
	schema *schema.Schema,
	predicates []*Predicate,
) error

type PreDBDelete = func(
	schema *schema.Schema,
	predicates []*Predicate,
) error

type Hooks struct {
	PostDBGet    []PostDBGet
	PostDBCreate []PostDBCreate
	PostDBUpdate []PostDBUpdate
	PostDBDelete []PostDBDelete
	PreDBGet     []PreDBGet
	PreDBCreate  []PreDBCreate
	PreDBUpdate  []PreDBUpdate
	PreDBDelete  []PreDBDelete
}

type RenameItem struct {
	Type            string `json:"type"` // "column" or "table"
	From            string `json:"from"`
	To              string `json:"to"`
	IsJunctionTable bool   `json:"is_junction_table,omitempty"` // use in rename table: If the table is a junction table
	SchemaName      string `json:"schema,omitempty"`            // use in rename column: The schema name of the column
	SchemaNamespace string `json:"schema_namespace,omitempty"`  // use in rename column: The schema name of the column
}

type Migration struct {
	Dir          string
	RenameTables []*RenameItem
	RenameFields []*RenameItem
}

type Config struct {
	Driver             string
	Name               string
	Host               string
	Port               string
	User               string
	Pass               string
	Logger             logger.Logger
	LogQueries         bool
	MigrationDir       string
	IgnoreMigration    bool
	DisableForeignKeys bool
	Hooks              func() *Hooks
}

func (c *Config) Clone() *Config {
	return &Config{
		Driver:             c.Driver,
		Name:               c.Name,
		Host:               c.Host,
		Port:               c.Port,
		User:               c.User,
		Pass:               c.Pass,
		Logger:             c.Logger,
		LogQueries:         c.LogQueries,
		MigrationDir:       c.MigrationDir,
		IgnoreMigration:    c.IgnoreMigration,
		DisableForeignKeys: c.DisableForeignKeys,
		Hooks:              c.Hooks,
	}
}

type Client interface {
	Dialect() string
	// Exec executes a query that does not return records. For example, in SQL, INSERT or UPDATE.
	// It return a sql.Result and an error if any.
	Exec(ctx context.Context, query string, args any) (sql.Result, error)
	// Query executes a query that returns rows, typically a SELECT in SQL.
	// It return a slice of *schema.Entity and an error if any.
	Query(ctx context.Context, query string, args any) ([]*schema.Entity, error)
	Rollback() error
	Commit() error
	CreateDBModel(s *schema.Schema, rs ...*schema.Relation) Model
	Tx(ctx context.Context) (Client, error)
	IsTx() bool
	Model(name string, types ...any) (Model, error)
	Close() error
	SchemaBuilder() *schema.Builder
	Reload(ctx context.Context, newSchemaBuilder *schema.Builder, migration *Migration, disableForeignKeys bool) (Client, error)
	DB() *sql.DB
	Config() *Config
	Hooks() *Hooks
}

type Model interface {
	Query(predicates ...*Predicate) Querier
	Mutation() Mutator
	Schema() *schema.Schema
	CreateFromJSON(ctx context.Context, json string) (id uint64, err error)
	Create(ctx context.Context, e *schema.Entity) (id uint64, err error)
	SetClient(client Client) Model
	Clone() Model
}

type QueryOption struct {
	Limit      uint         `json:"limit"`
	Offset     uint         `json:"offset"`
	Columns    []string     `json:"columns"`
	Order      []string     `json:"order"`
	Predicates []*Predicate `json:"predicates"`
	Model      Model        `json:"-"`
}

type Querier interface {
	Where(predicates ...*Predicate) Querier
	Limit(limit uint) Querier
	Offset(offset uint) Querier
	Select(columns ...string) Querier
	Order(order ...string) Querier
	Count(ctx context.Context, options *CountOption) (int, error)
	Get(ctx context.Context) ([]*schema.Entity, error)
	First(ctx context.Context) (*schema.Entity, error)
	Only(ctx context.Context) (*schema.Entity, error)
	Options() *QueryOption
}

type Mutator interface {
	Where(predicates ...*Predicate) Mutator
	GetRelationEntityIDs(fieldName string, fieldValue any) ([]driver.Value, error)
	Create(ctx context.Context, e *schema.Entity) (id uint64, err error)
	Update(ctx context.Context, e *schema.Entity) (affected int, err error)
	Delete(ctx context.Context) (affected int, err error)
}

package app

import (
	"context"
	"database/sql"
	"database/sql/driver"

	_ "github.com/DATA-DOG/go-sqlmock"
	"github.com/fastschema/fastschema/schema"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "modernc.org/sqlite"
)

type DBClient interface {
	Dialect() string
	Exec(ctx context.Context, query string, args any, bindValue any) error
	Rollback() error
	Commit() error
	CreateDBModel(s *schema.Schema, rs ...*schema.Relation) Model
	Tx(ctx context.Context) (DBClient, error)
	IsTx() bool
	Model(name string) (Model, error)
	Close() error
	SchemaBuilder() *schema.Builder
	Reload(newSchemaBuilder *schema.Builder, migration *Migration) (DBClient, error)
	DB() *sql.DB
	Config() *DBConfig
	Hooks() *Hooks
}

type Model interface {
	Query(predicates ...*Predicate) Query
	Mutation(skipTxs ...bool) Mutation
	Schema() *schema.Schema
	CreateFromJSON(json string) (id uint64, err error)
	Create(e *schema.Entity) (id uint64, err error)
	SetClient(client DBClient) Model
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

type Query interface {
	Where(predicates ...*Predicate) Query
	Limit(limit uint) Query
	Offset(offset uint) Query
	Select(columns ...string) Query
	Order(order ...string) Query
	Count(options *CountOption, ctxs ...context.Context) (int, error)
	Get(ctxs ...context.Context) ([]*schema.Entity, error)
	First(ctxs ...context.Context) (*schema.Entity, error)
	Only(ctxs ...context.Context) (*schema.Entity, error)
	Options() *QueryOption
}

type Mutation interface {
	Where(predicates ...*Predicate) Mutation
	GetRelationEntityIDs(fieldName string, fieldValue any) ([]driver.Value, error)
	Create(e *schema.Entity) (id uint64, err error)
	Update(e *schema.Entity) (affected int, err error)
	Delete() (affected int, err error)
}

type AfterDBContentListHook = func(query *QueryOption, entities []*schema.Entity) ([]*schema.Entity, error)

type DBConfig struct {
	Driver          string
	Name            string
	Host            string
	Port            string
	User            string
	Pass            string
	Logger          Logger
	LogQueries      bool
	MigrationDir    string
	IgnoreMigration bool
	Hooks           *Hooks
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

func (c *DBConfig) Clone() *DBConfig {
	return &DBConfig{
		Driver:       c.Driver,
		Name:         c.Name,
		Host:         c.Host,
		Port:         c.Port,
		User:         c.User,
		Pass:         c.Pass,
		Logger:       c.Logger,
		LogQueries:   c.LogQueries,
		MigrationDir: c.MigrationDir,
	}
}

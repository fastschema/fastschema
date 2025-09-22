package db

import (
	"context"
	"database/sql"
	"database/sql/driver"

	_ "github.com/DATA-DOG/go-sqlmock"
	"github.com/fastschema/fastschema/entity"
	"github.com/fastschema/fastschema/logger"
	"github.com/fastschema/fastschema/schema"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"

	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
)

// SupportDrivers returns list of supported drivers.
var SupportDrivers = []string{"mysql", "pgx", "sqlite"}

// PreDBQuery is a hook that is called before a database query is executed.
type PreDBQuery = func(
	ctx context.Context,
	option *QueryOption,
) error

// PostDBQuery is a hook that is called after a database query is executed.
type PostDBQuery = func(
	ctx context.Context,
	option *QueryOption,
	entities []*entity.Entity,
) ([]*entity.Entity, error)

// PreDBExec is a hook that is called before a database exec is executed.
type PreDBExec = func(
	ctx context.Context,
	option *QueryOption,
) error

// PostDBExec is a hook that is called after a database exec is executed.
type PostDBExec = func(
	ctx context.Context,
	option *QueryOption,
	result sql.Result,
) error

// PreDBCreate is a hook that is called before a database create is executed.
type PreDBCreate = func(
	ctx context.Context,
	schema *schema.Schema,
	createData *entity.Entity,
) error

// PostDBCreate is a hook that is called after a database create is executed.
type PostDBCreate = func(
	ctx context.Context,
	schema *schema.Schema,
	dataCreate *entity.Entity,
	id uint64,
) error

// PreDBUpdate is a hook that is called before a database update is executed.
type PreDBUpdate = func(
	ctx context.Context,
	schema *schema.Schema,
	predicates *[]*Predicate,
	updateData *entity.Entity,
) error

// PostDBUpdate is a hook that is called after a database update is executed.
type PostDBUpdate = func(
	ctx context.Context,
	schema *schema.Schema,
	predicates *[]*Predicate,
	updateData *entity.Entity,
	originalEntities []*entity.Entity,
	affected int,
) error

// PreDBDelete is a hook that is called before a database delete is executed.
type PreDBDelete = func(
	ctx context.Context,
	schema *schema.Schema,
	predicates *[]*Predicate,
) error

// PostDBDelete is a hook that is called after a database delete is executed.
type PostDBDelete = func(
	ctx context.Context,
	schema *schema.Schema,
	predicates *[]*Predicate,
	originalEntities []*entity.Entity,
	affected int,
) error

type Hooks struct {
	PreDBQuery   []PreDBQuery
	PostDBQuery  []PostDBQuery
	PreDBExec    []PreDBExec
	PostDBExec   []PostDBExec
	PreDBCreate  []PreDBCreate
	PostDBCreate []PostDBCreate
	PreDBUpdate  []PreDBUpdate
	PostDBUpdate []PostDBUpdate
	PreDBDelete  []PreDBDelete
	PostDBDelete []PostDBDelete
}

func (h *Hooks) Clone() *Hooks {
	return &Hooks{
		PostDBQuery:  append([]PostDBQuery{}, h.PostDBQuery...),
		PostDBCreate: append([]PostDBCreate{}, h.PostDBCreate...),
		PostDBUpdate: append([]PostDBUpdate{}, h.PostDBUpdate...),
		PostDBDelete: append([]PostDBDelete{}, h.PostDBDelete...),
		PreDBQuery:   append([]PreDBQuery{}, h.PreDBQuery...),
		PreDBCreate:  append([]PreDBCreate{}, h.PreDBCreate...),
		PreDBUpdate:  append([]PreDBUpdate{}, h.PreDBUpdate...),
		PreDBDelete:  append([]PreDBDelete{}, h.PreDBDelete...),
	}
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
	Driver             string        `json:"driver"`
	Name               string        `json:"name"`
	Host               string        `json:"host"`
	Port               string        `json:"port"`
	User               string        `json:"user"`
	Pass               string        `json:"pass"`
	Logger             logger.Logger `json:"-"`
	LogQueries         bool          `json:"log_queries"`
	MigrationDir       string        `json:"migration_dir"`
	IgnoreMigration    bool          `json:"ignore_migration"`
	DisableForeignKeys bool          `json:"disable_foreign_keys"`
	UseSoftDeletes     bool          `json:"use_soft_deletes"`
	Hooks              func() *Hooks `json:"-"`
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
	Exec(ctx context.Context, query string, args ...any) (sql.Result, error)
	// Query executes a query that returns rows, typically a SELECT in SQL.
	// It return a slice of *entity.Entity and an error if any.
	Query(ctx context.Context, query string, args ...any) ([]*entity.Entity, error)
	Rollback() error
	Commit() error
	Tx(ctx context.Context) (Client, error)
	IsTx() bool

	// Model return the model from given name.
	//
	//	Support finding model from name or types
	//	If the input model is a string, it will use the name to find the model
	//	Others, it will use the types of the input to find the model
	//
	//	entities: The entities that the model will be created from
	//  entities can be one of the following types:
	//		- []*entity.Entity: The first entity will be used to create the model
	//		- [][]byte: The first byte slice will be used to create the model by unmarshalling it
	Model(model any) (Model, error)
	Close() error
	SchemaBuilder() *schema.Builder
	Reload(
		ctx context.Context,
		newSchemaBuilder *schema.Builder,
		migration *Migration,
		disableForeignKeys bool,
		enableMigrations ...bool,
	) (Client, error)
	DB() *sql.DB
	Config() *Config
	Hooks() *Hooks
}

type Model interface {
	Query(predicates ...*Predicate) Querier
	Mutation() Mutator
	Schema() *schema.Schema
	CreateFromJSON(ctx context.Context, json string) (id uint64, err error)
	Create(ctx context.Context, e *entity.Entity) (id uint64, err error)
	SetClient(client Client) Model
	Clone() Model
}

// QueryOption is a struct that contains query options
//
//	Column and Unique are used for count query.
type QueryOption struct {
	Schema     *schema.Schema `json:"schema"`
	Limit      uint           `json:"limit"`
	Offset     uint           `json:"offset"`
	Columns    *[]string      `json:"columns"`
	Order      []string       `json:"order"`
	Predicates *[]*Predicate  `json:"predicates"`
	Query      string         `json:"query"`
	Args       any            `json:"args"`
	// For count query
	Column string `json:"column"`
	Unique bool   `json:"unique"`
}

type Querier interface {
	// Find with soft-deleted records.
	WithTrashed() Querier
	// Find only soft-deleted records.
	OnlyTrashed() Querier
	Where(predicates ...*Predicate) Querier
	Limit(limit uint) Querier
	Offset(offset uint) Querier
	Select(columns ...string) Querier
	Order(order ...string) Querier
	Count(ctx context.Context, options ...*QueryOption) (int, error)
	Get(ctx context.Context) ([]*entity.Entity, error)
	First(ctx context.Context) (*entity.Entity, error)
	Only(ctx context.Context) (*entity.Entity, error)
	Options() *QueryOption
}

type Mutator interface {
	Where(predicates ...*Predicate) Mutator
	GetRelationEntityIDs(fieldName string, fieldValue any) ([]driver.Value, error)
	Create(ctx context.Context, e *entity.Entity) (id uint64, err error)
	Update(ctx context.Context, e *entity.Entity) (affected int, err error)
	Delete(ctx context.Context) (affected int, err error)
}

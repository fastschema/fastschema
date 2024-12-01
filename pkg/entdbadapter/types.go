package entdbadapter

import (
	"context"
	"database/sql"
	"database/sql/driver"

	"entgo.io/ent/dialect"
	entSchema "entgo.io/ent/dialect/sql/schema"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/entity"
	"github.com/fastschema/fastschema/schema"
)

type EntAdapter interface {
	NewEdgeSpec(
		r *schema.Relation,
		nodeIDs []driver.Value,
	) (*sqlgraph.EdgeSpec, error)
	NewEdgeStepOption(r *schema.Relation) (sqlgraph.StepOption, error)
	Reload(
		ctx context.Context,
		newSchemaBuilder *schema.Builder,
		migration *db.Migration,
		disableForeignKeys bool,
		enableMigrations ...bool,
	) (_ db.Client, err error)
	Driver() dialect.Driver
	Close() error
	Commit() error
	Rollback() error
	Config() *db.Config
	DB() *sql.DB
	Dialect() string
	Exec(
		ctx context.Context,
		query string,
		args ...any,
	) (sql.Result, error)
	Query(
		ctx context.Context,
		query string,
		args ...any,
	) ([]*entity.Entity, error)
	Hooks() *db.Hooks
	IsTx() bool
	Model(model any) (db.Model, error)
	SchemaBuilder() *schema.Builder
	Tx(ctx context.Context) (db.Client, error)
	Migrate(
		ctx context.Context,
		migration *db.Migration,
		disableForeignKeys bool,
		appendEntTables ...*entSchema.Table,
	) (err error)
	SetSQLDB(db *sql.DB)
	SetDriver(driver dialect.Driver)
}

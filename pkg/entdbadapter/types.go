package entdbadapter

import (
	"context"
	"database/sql"
	"database/sql/driver"

	"entgo.io/ent/dialect"
	entSchema "entgo.io/ent/dialect/sql/schema"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/schema"
)

type EntAdapter interface {
	db.Client

	Driver() dialect.Driver
	SetSQLDB(db *sql.DB)
	SetDriver(driver dialect.Driver)
	NewEdgeStepOption(r *schema.Relation) (sqlgraph.StepOption, error)
	NewEdgeSpec(
		r *schema.Relation,
		nodeIDs []driver.Value,
	) (*sqlgraph.EdgeSpec, error)
	Migrate(
		ctx context.Context,
		changes *db.Changes,
		disableForeignKeys bool,
		appendEntTables ...*entSchema.Table,
	) (err error)
}

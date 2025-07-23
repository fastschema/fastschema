package entdbadapter

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"time"

	atlasMigrate "ariga.io/atlas/sql/migrate"
	"ariga.io/atlas/sql/mysql"
	"ariga.io/atlas/sql/postgres"
	atlasSchema "ariga.io/atlas/sql/schema"
	"ariga.io/atlas/sql/sqlite"
	"entgo.io/ent/dialect"
	dialectsql "entgo.io/ent/dialect/sql"
	entSchema "entgo.io/ent/dialect/sql/schema"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/entity"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
)

// RelMaps map the relation type to the ent relation type
var RelMaps = map[schema.RelationType]sqlgraph.Rel{
	schema.O2O: sqlgraph.O2O,
	schema.O2M: sqlgraph.O2M,
	schema.M2M: sqlgraph.M2M,
}

var dialectMap = map[string]string{
	"sqlmock":  dialect.MySQL,
	"mysql":    dialect.MySQL,
	"pgx":      dialect.Postgres,
	"postgres": dialect.Postgres,
	"sqlite":   dialect.SQLite,
}

// entFieldTypesMapper map the field type to the ent field type
var entFieldTypesMapper = map[schema.FieldType]field.Type{
	schema.TypeString:  field.TypeString,
	schema.TypeText:    field.TypeString,
	schema.TypeEnum:    field.TypeEnum,
	schema.TypeInt:     field.TypeInt,
	schema.TypeBool:    field.TypeBool,
	schema.TypeTime:    field.TypeTime,
	schema.TypeJSON:    field.TypeJSON,
	schema.TypeUUID:    field.TypeUUID,
	schema.TypeBytes:   field.TypeBytes,
	schema.TypeInt8:    field.TypeInt8,
	schema.TypeInt16:   field.TypeInt16,
	schema.TypeInt32:   field.TypeInt32,
	schema.TypeInt64:   field.TypeInt64,
	schema.TypeUint8:   field.TypeUint8,
	schema.TypeUint16:  field.TypeUint16,
	schema.TypeUint32:  field.TypeUint32,
	schema.TypeUint:    field.TypeUint,
	schema.TypeUint64:  field.TypeUint64,
	schema.TypeFloat32: field.TypeFloat32,
	schema.TypeFloat64: field.TypeFloat64,
}

// createEntColumn convert a field to ent column
func createEntColumn(f *schema.Field) *entSchema.Column {
	entColumn := &entSchema.Column{
		Name: f.Name,
		Type: entFieldTypesMapper[f.Type],
	}

	if f.Type == schema.TypeText {
		entColumn.Size = 2147483647
	}

	if f.DB != nil {
		entColumn.Increment = f.DB.Increment
		entColumn.Attr = f.DB.Attr
		entColumn.Key = f.DB.Key
		entColumn.Collation = f.DB.Collation
	}

	if f.Size > 0 {
		entColumn.Size = f.Size
	}

	entColumn.Unique = f.Unique
	entColumn.Default = f.Default
	entColumn.Nullable = f.Optional

	if f.Enums != nil {
		entColumn.Enums = utils.Map(f.Enums, func(e *schema.FieldEnum) string {
			return e.Value
		})
	}

	if f.Type == schema.TypeTime {
		entColumn.SchemaType = map[string]string{"mysql": "datetime"}
	}

	return entColumn
}

// CreateDBDSN create a DSN string for the database connection
func CreateDBDSN(config *db.Config) string {
	dsn := ""

	if config.Driver == "mysql" {
		dsn = fmt.Sprintf(
			"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&collation=utf8mb4_unicode_ci&parseTime=true&multiStatements=true",
			config.User,
			config.Pass,
			config.Host,
			config.Port,
			config.Name,
		)
	}

	if config.Driver == "pgx" {
		dsn = fmt.Sprintf(
			"host=%s port=%s user=%s dbname=%s password=%s sslmode=disable",
			config.Host,
			config.Port,
			config.User,
			config.Name,
			config.Pass,
		)
	}

	if config.Driver == "sqlite" {
		if config.Name == ":memory:" {
			return ":memory:?cache=shared&_fk=1&_pragma=foreign_keys(1)"
		}

		dsn = fmt.Sprintf(
			"file:%s?cache=shared&_fk=1&_pragma=foreign_keys(1)",
			config.Name,
		)
	}

	return dsn
}

func CreateDebugFN(config *db.Config) func(ctx context.Context, args ...any) {
	return func(ctx context.Context, args ...any) {
		msg := fmt.Sprintf("%v", args)
		args = []any{msg}
		traceable, ok := ctx.(fs.Traceable)
		if ok {
			if traceID := traceable.TraceID(); traceID != "" {
				args = append(args, map[string]any{
					fs.TraceID: traceID,
				})
			}
		} else {
			if traceID := ctx.Value(fs.ContextKeyTraceID); traceID != nil {
				args = append(args, map[string]any{
					fs.TraceID: traceID,
				})
			}
		}

		if config.Logger != nil {
			// config.Logger.Debug(args...)
			config.Logger.WithContext(nil, 5).Debug(args...)
		} else {
			fmt.Println(args...)
		}
	}
}

func GetEntDialect(config *db.Config) (string, error) {
	entDialect, ok := dialectMap[config.Driver]
	if !ok {
		return "", fmt.Errorf("unsupported driver: %v", config.Driver)
	}

	return entDialect, nil
}

func createRenameColumnsHook(
	renameTables []*db.RenameItem,
	renameColumns []*db.RenameItem,
) entSchema.DiffHook {
	return func(next entSchema.Differ) entSchema.Differ {
		return entSchema.DiffFunc(func(current, desired *atlasSchema.Schema) ([]atlasSchema.Change, error) {
			changes, err := next.Diff(current, desired)
			if err != nil {
				return nil, err
			}

			// Skip renaming table for now because ent will automatically filter out the renaming table changes.
			// If the change is add new table, check if the new table is renamed from another table
			// if yes, remove the add table changes
			atlasChanges := atlasSchema.Changes(changes)
			for _, c := range changes {
				addTable, ok := c.(*atlasSchema.AddTable)
				if !ok {
					continue
				}

				// check if the new table is renamed from another table
				matchedRenameTables := utils.Filter(renameTables, func(rt *db.RenameItem) bool {
					return addTable.T.Name == rt.To
				})

				// if the table is not in the renameTables, do nothing
				if len(matchedRenameTables) == 0 {
					continue
				}

				// if the table is in the renameTables, remove the add table changes
				indexAddTable := atlasChanges.IndexAddTable(addTable.T.Name)
				atlasChanges.RemoveIndex(indexAddTable)
				changes = []atlasSchema.Change(atlasChanges)
			}

			// Check changes for rename columns.
			for _, c := range changes {
				modifyTable, ok := c.(*atlasSchema.ModifyTable)
				if !ok {
					continue
				}

				// check if changes is add new column
				changes := atlasSchema.Changes(modifyTable.Changes)
				for _, change := range changes {
					if addColumn, ok := change.(*atlasSchema.AddColumn); ok {
						// check if the new column is renamed from another column
						matchedRenameFields := utils.Filter(renameColumns, func(rf *db.RenameItem) bool {
							return addColumn.C.Name == rf.To
						})

						// if the new column is not in the renameColumns, do nothing
						if len(matchedRenameFields) == 0 {
							continue
						}

						// if the new column is in the renameColumns, add a new renaming change and remove the add column changes
						renameField := matchedRenameFields[0]
						currentTable, ok := current.Table(renameField.SchemaNamespace)
						if !ok {
							return nil, fmt.Errorf("rename_column: table %s not found", renameField.SchemaNamespace)
						}

						fromColumn, ok := currentTable.Column(renameField.From)
						if !ok {
							return nil, fmt.Errorf("rename_column: column %s.%s not found", renameField.SchemaNamespace, renameField.From)
						}

						// Append a new renaming change.
						changes = append(changes, &atlasSchema.RenameColumn{
							From: fromColumn,
							To:   addColumn.C,
						})

						// Remove the add changes.
						indexAddColumn := changes.IndexAddColumn(addColumn.C.Name)
						changes.RemoveIndex(indexAddColumn)
						modifyTable.Changes = changes
					}
				}
			}

			return changes, nil
		})
	}
}

func getPlanForRenameTables(
	ctx context.Context,
	migrateDriver atlasMigrate.Driver,
	renameTables []*db.RenameItem,
) (*atlasMigrate.Plan, error) {
	if len(renameTables) == 0 {
		return nil, nil
	}

	allTables := []string{}
	for _, c := range renameTables {
		if !utils.Contains(allTables, c.From) {
			allTables = append(allTables, c.From)
		}
	}

	inspectedSchema, err := migrateDriver.InspectSchema(ctx, "", &atlasSchema.InspectOptions{
		Tables: allTables,
	})

	if err != nil {
		return nil, err
	}

	changes := []atlasSchema.Change{}

	for _, c := range renameTables {
		table, ok := inspectedSchema.Table(c.From)
		if !ok {
			return nil, fmt.Errorf("table %s not found", c.From)
		}

		changes = append(changes, &atlasSchema.RenameTable{
			From: table,
			To:   cloneMigrateTableWithNewName(table, c.To),
		})
	}

	return migrateDriver.PlanChanges(
		ctx,
		"simulate_changes",
		changes,
	)
}

func getAtlasMigrateDriver(dialect string, db *sql.DB) (atlasMigrate.Driver, error) {
	switch dialect {
	case "mysql":
		return mysql.Open(db)
	case "pgx", "postgres":
		return postgres.Open(db)
	case "sqlite", "sqlite3":
		return sqlite.Open(db)
	}

	return nil, fmt.Errorf("unsupported dialect: %v", dialect)
}

func cloneMigrateTableWithNewName(t *atlasSchema.Table, name string) *atlasSchema.Table {
	return &atlasSchema.Table{
		Name:        name,
		Schema:      t.Schema,
		Columns:     t.Columns,
		Indexes:     t.Indexes,
		PrimaryKey:  t.PrimaryKey,
		ForeignKeys: t.ForeignKeys,
		Attrs:       t.Attrs,
	}
}

func NOW(dialect string) any {
	switch dialect {
	case "mysql":
		return dialectsql.Expr("NOW()")
	case "pgx", "postgres":
		return dialectsql.Expr("now()")
	case "sqlite", "sqlite3":
		return dialectsql.Expr("datetime('now')")
	}

	return time.Now().Format("2006-01-02 15:04:05")
}

func isDateTimeColumn(scanType reflect.Type, databaseTypeName string) bool {
	isSQLTime := databaseTypeName == "DATETIME"
	isStructTime := scanType != nil &&
		scanType.Kind() == reflect.Struct &&
		scanType.String() == "time.Time"
	return isStructTime || isSQLTime
}

func driverExec(
	driver dialect.Driver,
	ctx context.Context,
	query string,
	args any,
) (sql.Result, error) {
	var result = new(sql.Result)
	if err := driver.Exec(ctx, query, args, result); err != nil {
		return nil, err
	}

	return *result, nil
}

func driverQuery(
	driver dialect.Driver,
	ctx context.Context,
	query string,
	args any,
) ([]*entity.Entity, error) {
	var rows = &dialectsql.Rows{}
	if err := driver.Query(ctx, query, args, rows); err != nil {
		return nil, err
	}

	columns, err := getRowsColumns(rows)
	if err != nil {
		return nil, err
	}

	entities := []*entity.Entity{}
	for rows.Next() {
		values := rawRowsScanValues(columns)
		if err := rows.Scan(values...); err != nil {
			return nil, err
		}

		e := entity.New()
		for i, column := range columns {
			if v, err := columnAssignValue(
				columns[i].Name,
				columns[i].FieldType,
				values[i],
				e,
			); err != nil {
				return nil, fmt.Errorf("columnAssignValue for column '%v': %w", column, err)
			} else {
				e.Set(column.Name, v)
			}
		}

		entities = append(entities, e)
	}

	return entities, nil
}

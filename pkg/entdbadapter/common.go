package entdbadapter

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"reflect"
	"strings"
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
	"github.com/google/uuid"
	_ "github.com/ncruces/go-sqlite3/vfs/memdb"
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
	"sqlite3":  dialect.SQLite,
}

var goSqlDriverNameMap = map[string]string{
	"mysql":    "mysql",
	"pgx":      "pgx",
	"postgres": "pgx",
	"sqlite":   "sqlite3",
	"sqlite3":  "sqlite3",
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
		entColumn.Key = f.DB.Key.String()
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
		sslMode := config.SSLMode
		if sslMode == "" {
			sslMode = "prefer"
		}
		dsn = fmt.Sprintf(
			"host=%s port=%s user=%s dbname=%s password=%s sslmode=%s",
			config.Host,
			config.Port,
			config.User,
			config.Name,
			config.Pass,
			sslMode,
		)
	}

	if config.Driver == "sqlite" {
		if after, ok := strings.CutPrefix(config.Name, ":memory:"); ok {
			name := after
			return fmt.Sprintf("file:/fastschema_%s.db?vfs=memdb&_fk=1&_pragma=foreign_keys(1)", name)
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
	isNullTime := scanType != nil &&
		scanType.Kind() == reflect.Struct &&
		scanType.String() == "sql.NullTime"
	return isStructTime || isSQLTime || isNullTime
}

func normalizeIDValue(field *schema.Field, value any) (driver.Value, error) {
	if field == nil || value == nil {
		return value, nil
	}

	switch field.Type {
	case schema.TypeUint, schema.TypeUint8, schema.TypeUint16, schema.TypeUint32, schema.TypeUint64:
		converted, err := utils.AnyToUint[uint64](value)
		if err != nil {
			return nil, fmt.Errorf("convert %s to unsigned integer: %w", field.Name, err)
		}
		return converted, nil
	case schema.TypeInt, schema.TypeInt8, schema.TypeInt16, schema.TypeInt32, schema.TypeInt64:
		converted, err := utils.AnyToInt[int64](value)
		if err != nil {
			return nil, fmt.Errorf("convert %s to integer: %w", field.Name, err)
		}
		return converted, nil
	case schema.TypeUUID:
		return normalizeUUIDValue(value)
	default:
		return value, nil
	}
}

// normalizeUUIDValue converts various UUID representations to uuid.UUID
func normalizeUUIDValue(value any) (driver.Value, error) {
	if value == nil {
		return nil, nil
	}

	switch v := value.(type) {
	case uuid.UUID:
		return v, nil
	case *uuid.UUID:
		if v == nil {
			return nil, nil
		}
		return *v, nil
	case string:
		parsed, err := uuid.Parse(v)
		if err != nil {
			return nil, fmt.Errorf("parse UUID from string: %w", err)
		}
		return parsed, nil
	case []byte:
		if len(v) == 16 {
			parsed, err := uuid.FromBytes(v)
			if err != nil {
				return nil, fmt.Errorf("parse UUID from bytes: %w", err)
			}
			return parsed, nil
		}
		// Try parsing as string
		parsed, err := uuid.Parse(string(v))
		if err != nil {
			return nil, fmt.Errorf("parse UUID from bytes string: %w", err)
		}
		return parsed, nil
	case [16]byte:
		return uuid.UUID(v), nil
	default:
		return nil, fmt.Errorf("unsupported UUID value type: %T", value)
	}
}

func getRelationTargetField(builder *schema.Builder, relation *schema.Relation) (*schema.Field, error) {
	if builder == nil {
		return nil, fmt.Errorf("schema builder is not initialized")
	}

	if relation == nil {
		return nil, fmt.Errorf("relation is not defined")
	}

	targetSchema, err := builder.Schema(relation.TargetSchemaName)
	if err != nil {
		return nil, err
	}

	targetColumn := relation.TargetColumn
	if targetColumn == "" || relation.Type.IsM2M() {
		targetColumn = targetSchema.PrimaryKeyName()
		if targetColumn == "" {
			targetColumn = entity.FieldID
		}
	}

	targetField := targetSchema.Field(targetColumn)
	if targetField == nil {
		return nil, schema.ErrFieldNotFound(targetSchema.Name, targetColumn)
	}

	return targetField, nil
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

func isZeroValue(value any) bool {
	if value == nil {
		return true
	}

	if v, ok := value.(uuid.UUID); ok {
		return v == uuid.Nil
	}

	if v, ok := value.(*uuid.UUID); ok {
		return v == nil || *v == uuid.Nil
	}

	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.Bool:
		return !rv.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return rv.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return rv.Float() == 0
	case reflect.String:
		return rv.Len() == 0
	case reflect.Slice, reflect.Map:
		return rv.Len() == 0
	case reflect.Interface, reflect.Pointer:
		return rv.IsNil()
	case reflect.Array:
		zero := reflect.Zero(rv.Type()).Interface()
		return reflect.DeepEqual(value, zero)
	default:
		zero := reflect.Zero(rv.Type()).Interface()
		return reflect.DeepEqual(value, zero)
	}
}

func valueKey(value any) string {
	if value == nil {
		return "<nil>"
	}

	switch v := value.(type) {
	case []byte:
		return fmt.Sprintf("%T:%x", value, v)
	default:
		return fmt.Sprintf("%T:%v", value, value)
	}
}

func fieldTypeError(fieldType string, fieldValue any) error {
	return fmt.Errorf("expected value of type '%s', got '%T'", fieldType, fieldValue)
}

func invalidFKError(edgeSchemaName, fkColumn string, id any, err error) error {
	return fmt.Errorf(
		`invalid FK value %s.%s for node id=%v: %w`,
		edgeSchemaName, fkColumn, id, err,
	)
}

func noFKNodeError(schemaName, edgeSchemaName, fkColumn string, id, fk any) error {
	return fmt.Errorf(
		`no FK node (%s) found for (%s=%v).%s=%v`,
		schemaName, edgeSchemaName, id, fkColumn, fk,
	)
}

func invalidEntityArrayError(schemaName, fieldName string, edgeValues any) error {
	return fmt.Errorf(
		`edge values %s.%s=%v (%T) is not []*entity.Entity`,
		schemaName, fieldName, edgeValues, edgeValues,
	)
}

func collectEntityIDs(
	schemaName string,
	primaryField *schema.Field,
	entities []*entity.Entity,
) ([]driver.Value, map[string]*entity.Entity, error) {
	ids := make([]driver.Value, 0, len(entities))
	byKey := make(map[string]*entity.Entity, len(entities))
	for _, node := range entities {
		var idValue any
		if primaryField != nil {
			idValue = node.Get(primaryField.Name)
		}
		if isZeroValue(idValue) {
			idValue = node.ID()
		}
		if isZeroValue(idValue) {
			return nil, nil, fmt.Errorf("entity %s has invalid id", schemaName)
		}
		normalized, err := normalizeIDValue(primaryField, idValue)
		if err != nil {
			return nil, nil, fmt.Errorf("entity %s has invalid id: %w", schemaName, err)
		}
		key := valueKey(normalized)
		ids = append(ids, normalized)
		byKey[key] = node
	}
	return ids, byKey, nil
}

// collectParentRefs collects reference values from parent entities for edge loading.
// Returns: slice of unique ref values, map of ref value -> parent entities.
// - refField: field definition to get value from (uses field.Name to get value)
// - schemaName: parent schema name for error messages
// - skipNullFK: if true, skip entities with null FK values (for non-owner side)
func collectParentRefs(
	entities []*entity.Entity,
	refColumn string,
	refField *schema.Field,
	schemaName string,
	skipNullFK bool,
) ([]any, map[string][]*entity.Entity, error) {
	refs := make([]any, 0, len(entities))
	parentMap := make(map[string][]*entity.Entity)

	for _, ent := range entities {
		refValue := ent.Get(refColumn)

		if isZeroValue(refValue) {
			if skipNullFK {
				continue // FK is null, no edge to load
			}
			return nil, nil, invalidFKError(schemaName, refColumn, ent.ID(), fmt.Errorf("empty reference value"))
		}

		normalized, err := normalizeIDValue(refField, refValue)
		if err != nil {
			return nil, nil, err
		}

		key := valueKey(normalized)
		if _, exists := parentMap[key]; !exists {
			refs = append(refs, normalized)
		}
		parentMap[key] = append(parentMap[key], ent)
	}

	return refs, parentMap, nil
}

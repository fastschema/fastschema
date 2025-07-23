package entdbadapter

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"

	"entgo.io/ent/dialect"
	dialectSql "entgo.io/ent/dialect/sql"
	entSchema "entgo.io/ent/dialect/sql/schema"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/schema"
)

// NewClient creates a new ent client
func NewClient(config *db.Config, schemaBuilder *schema.Builder) (_ db.Client, err error) {
	return NewEntClient(config, schemaBuilder)
}

// NewTestClient creates a new ent client with sqlite memory
func NewTestClient(
	migrationDir string,
	schemaBuilder *schema.Builder,
	hookFns ...func() *db.Hooks,
) (_ db.Client, err error) {
	hookFns = append(hookFns, func() *db.Hooks {
		return &db.Hooks{}
	})
	return NewClient(&db.Config{
		Driver:       "sqlite",
		Name:         ":memory:",
		MigrationDir: migrationDir,
		LogQueries:   false,
		Hooks:        hookFns[0],
	}, schemaBuilder)
}

// NewEntClient creates a new ent client
func NewEntClient(
	config *db.Config,
	schemaBuilder *schema.Builder,
	useEntDrivers ...*dialectSql.Driver,
) (_ db.Client, err error) {
	var (
		db            *sql.DB
		entDialect    string
		sqlDriver     *dialectSql.Driver
		dialectDriver dialect.Driver
	)

	if schemaBuilder == nil {
		return nil, errors.New("schema builder is required")
	}

	if db, err = sql.Open(config.Driver, CreateDBDSN(config)); err != nil {
		return nil, err
	}

	if len(useEntDrivers) > 0 && useEntDrivers[0] != nil {
		sqlDriver = useEntDrivers[0]
	}

	if sqlDriver == nil {
		if entDialect, err = GetEntDialect(config); err != nil {
			return nil, fmt.Errorf("unsupported driver: %v", config.Driver)
		}
		sqlDriver = dialectSql.OpenDB(entDialect, db)
	}

	dialectDriver = sqlDriver
	adapter, err := NewDBAdapter(config, schemaBuilder)
	if err != nil {
		return nil, err
	}

	entAdapter, ok := adapter.(EntAdapter)
	if !ok {
		return nil, errors.New("invalid adapter")
	}

	if config.LogQueries {
		dialectDriver = dialect.DebugWithContext(sqlDriver, CreateDebugFN(config))
	}

	entAdapter.SetSQLDB(db)
	entAdapter.SetDriver(dialectDriver)
	if config.Driver == "sqlmock" {
		return adapter, nil
	}

	if !config.IgnoreMigration {
		if err = entAdapter.Migrate(context.Background(), nil, config.DisableForeignKeys); err != nil {
			return nil, err
		}
	}

	return adapter, nil
}

func (d *Adapter) Reload(
	ctx context.Context,
	newSchemaBuilder *schema.Builder,
	migration *db.Migration,
	disableForeignKeys bool,
	enableMigrations ...bool,
) (_ db.Client, err error) {
	enableMigrations = append(enableMigrations, true)
	renamedEntTables := make([]*entSchema.Table, 0)
	newConfig := d.config.Clone()
	newConfig.IgnoreMigration = !enableMigrations[0]
	newAdapter, err := NewClient(newConfig, newSchemaBuilder)
	if err != nil {
		return nil, err
	}

	// When a table is renamed, the table with old name will not exist in the schema builder.
	// Ent won't know about the old table, so any operations on it will fail.
	// Append the old ent table to the tables list to help ent know about it.

	// When rename a m2m field, there are 2 updates that need to be done:
	// 1. Rename the m2m field in the junction table
	// 2. Rename junction table
	// (1) Can be done via Ent DiffHook.
	// (2) Can be done via Ent ApplyHook.
	// The order is (1) then (2): rename the m2m field first, then rename the junction table.

	// (1) To help Ent know about the junction table field was renamed,
	// we will create a new junction ent table with the new field name.
	// At that time, the new junction table will has it's new name.
	// So we need to rename the new junction table to the old name
	// and add the old junction table to the tables list to help Ent know about it
	// Ent will then know about the old junction table and be able to rename it's columns.

	if migration != nil && len(migration.RenameTables) > 0 {
		for _, renameTable := range migration.RenameTables {
			if !renameTable.IsJunctionTable {
				continue
			}

			oldJunctionModel, err := d.Model(renameTable.From)
			if err != nil {
				return nil, err
			}

			newJunctionModel, err := newAdapter.Model(renameTable.To)
			if err != nil {
				return nil, err
			}

			oldEntJunctionModel, ok := oldJunctionModel.(*Model)
			if !ok {
				return nil, errors.New("model is not an ent model")
			}

			newEntJunctionModel, ok := newJunctionModel.(*Model)
			if !ok {
				return nil, errors.New("model is not an ent model")
			}

			newJunctionTable := newEntJunctionModel.GetEntTable()
			renamedEntTables = append(renamedEntTables, &entSchema.Table{
				// Override the new junction table name with the old name
				// to help Ent know about the old junction table columns changes
				Name:        oldEntJunctionModel.GetEntTable().Name,
				Columns:     newJunctionTable.Columns,
				PrimaryKey:  newJunctionTable.PrimaryKey,
				ForeignKeys: newJunctionTable.ForeignKeys,
				Indexes:     newJunctionTable.Indexes,
				Annotation:  newJunctionTable.Annotation,
			})
		}
	}

	if err := d.Close(); err != nil {
		return nil, err
	}

	newEntAdapter, ok := newAdapter.(EntAdapter)
	if !ok {
		return nil, fmt.Errorf("invalid adapter, want EntAdapter, got %T", newAdapter)
	}

	if err = newEntAdapter.Migrate(ctx, migration, disableForeignKeys, renamedEntTables...); err != nil {
		return nil, err
	}

	return newAdapter, nil
}

func NewDBAdapter(
	config *db.Config,
	schemaBuilder *schema.Builder,
) (db.Client, error) {
	a := &Adapter{
		driver:        nil,
		sqldb:         nil,
		config:        config,
		migrationDir:  config.MigrationDir,
		schemaBuilder: schemaBuilder,
		models:        make([]*Model, 0),
		tables:        make([]*entSchema.Table, 0),
		edgeSpec:      make(map[string]sqlgraph.EdgeSpec),
		typesModels:   make(map[reflect.Type]*Model),
	}

	if err := a.init(); err != nil {
		return nil, err
	}

	for _, s := range schemaBuilder.Schemas() {
		if s == nil || s.SystemSchema == nil {
			continue
		}

		typeModel, err := a.model(s.Name)
		if err != nil {
			return nil, err
		}

		a.typesModels[s.RType] = typeModel
	}

	return a, nil
}

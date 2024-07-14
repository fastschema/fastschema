package entdbadapter

import (
	"context"
	"fmt"

	atlasMigrate "ariga.io/atlas/sql/migrate"
	"ariga.io/atlas/sql/sqltool"
	entDialect "entgo.io/ent/dialect"
	entSchema "entgo.io/ent/dialect/sql/schema"
	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/pkg/utils"
)

func (d *Adapter) Migrate(
	ctx context.Context,
	migration *db.Migration,
	disableForeignKeys bool,
	appendEntTables ...*entSchema.Table,
) (err error) {
	tables := d.tables
	migration = utils.If(migration == nil, &db.Migration{}, migration)
	renameTables := migration.RenameTables
	renameFields := migration.RenameFields
	migrationDir, err := atlasMigrate.NewLocalDir(d.migrationDir)
	if err != nil {
		return err
	}

	migrateDriver, err := getAtlasMigrateDriver(d.driver.Dialect(), d.sqldb)
	if err != nil {
		return err
	}

	// When a table is renamed, it will not exist in the schema builder.
	// Ent won't know about the old table, so any operations on it will fail.
	// Append the ent table of the old table to the tables list to help ent know about it.
	// Ent will then be able to perform the following operations on the old table:
	// - Rename the old table columns if needed:
	// 		when a m2m relation is renamed, both junction table name and columns are renamed.
	// - Rename the old table name.

	tables = append(tables, appendEntTables...)
	renameTablesPlan, err := getPlanForRenameTables(ctx, migrateDriver, renameTables)
	if err != nil {
		return err
	}

	if err := atlasMigrate.Validate(migrationDir); err != nil {
		return fmt.Errorf("validating migration directory: %w", err)
	}

	applyHook := entSchema.WithApplyHook(func(next entSchema.Applier) entSchema.Applier {
		return entSchema.ApplyFunc(func(ctx context.Context, conn entDialect.ExecQuerier, plan *atlasMigrate.Plan) error {
			defer func() {
				if len(plan.Changes) > 0 {
					atlasMigrate.NewPlanner(nil, migrationDir, []atlasMigrate.PlannerOption{
						atlasMigrate.WithFormatter(sqltool.GolangMigrateFormatter),
						atlasMigrate.PlanWithChecksum(true),
					}...).WritePlan(plan)
				}
			}()

			if renameTablesPlan != nil {
				plan.Changes = append(plan.Changes, renameTablesPlan.Changes...)
			}

			return next.Apply(ctx, conn, plan)
		})
	})
	migrateOptions := []entSchema.MigrateOption{
		entSchema.WithDir(migrationDir),                    // provide migration directory
		entSchema.WithMigrationMode(entSchema.ModeInspect), // provide migration mode
		entSchema.WithDialect(d.driver.Dialect()),          // Ent dialect to use
		entSchema.WithFormatter(atlasMigrate.DefaultFormatter),
		entSchema.WithDropIndex(true),
		entSchema.WithForeignKeys(!disableForeignKeys),
		applyHook,
		entSchema.WithDiffHook(
			createRenameColumnsHook(renameTables, renameFields),
		),
	}

	migrate, err := entSchema.NewMigrate(d.driver, migrateOptions...)
	if err != nil {
		return err
	}

	if err = migrate.Create(ctx, tables...); err != nil {
		return err
	}

	return nil
}

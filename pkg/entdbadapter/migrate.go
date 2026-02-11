package entdbadapter

import (
	"context"
	"fmt"
	"strings"

	atlasMigrate "ariga.io/atlas/sql/migrate"
	"ariga.io/atlas/sql/sqltool"
	entDialect "entgo.io/ent/dialect"
	entSchema "entgo.io/ent/dialect/sql/schema"
	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/pkg/utils"
)

func (d *Adapter) Migrate(
	ctx context.Context,
	changes *db.Changes,
	disableForeignKeys bool,
	appendEntTables ...*entSchema.Table,
) (err error) {
	tables := d.tables
	changes = utils.If(changes == nil, &db.Changes{}, changes)
	renameTables := changes.RenameTables
	renameFields := changes.RenameFields
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

	// if err := atlasMigrate.Validate(migrationDir); err != nil {
	// 	return fmt.Errorf("validating migration directory: %w", err)
	// }

	applyHook := entSchema.WithApplyHook(func(next entSchema.Applier) entSchema.Applier {
		return entSchema.ApplyFunc(func(ctx context.Context, conn entDialect.ExecQuerier, plan *atlasMigrate.Plan) error {
			defer func() {
				if len(plan.Changes) > 0 {
					if err := atlasMigrate.NewPlanner(nil, migrationDir, []atlasMigrate.PlannerOption{
						atlasMigrate.WithFormatter(sqltool.GolangMigrateFormatter),
						atlasMigrate.PlanWithChecksum(false),
					}...).WritePlan(plan); err != nil {
						panic(fmt.Errorf("writing migration plan: %w", err))
					}
				}
			}()

			if renameTablesPlan != nil {
				plan.Changes = append(plan.Changes, renameTablesPlan.Changes...)
			}

			return next.Apply(ctx, conn, plan)
		})
	})

	migrate, err := d.newEntMigrate(
		migrationDir,
		entSchema.WithForeignKeys(!disableForeignKeys),
		applyHook,
		entSchema.WithDiffHook(
			createRenameColumnsHook(renameTables, renameFields),
		),
	)
	if err != nil {
		return err
	}

	if err = migrate.Create(ctx, tables...); err != nil {
		return err
	}

	return nil
}

// newEntMigrate creates a new ent migrate instance with common options
func (d *Adapter) newEntMigrate(dir atlasMigrate.Dir, opts ...entSchema.MigrateOption) (*entSchema.Atlas, error) {
	migrateOptions := []entSchema.MigrateOption{
		entSchema.WithDir(dir),
		entSchema.WithMigrationMode(entSchema.ModeInspect),
		entSchema.WithDialect(d.driver.Dialect()),
		entSchema.WithFormatter(sqltool.GolangMigrateFormatter),
		entSchema.WithDropIndex(true),
		entSchema.WithForeignKeys(true),
	}
	migrateOptions = append(migrateOptions, opts...)

	return entSchema.NewMigrate(d.driver, migrateOptions...)
}

// GenerateMigrationFiles compares current database state with schema and generates migration SQL.
// This method uses ent's Atlas integration to generate migration files in the migration directory.
func (d *Adapter) GenerateMigrationFiles(ctx context.Context, name string) error {
	if name == "" {
		name = "changes"
	}

	// Create a local migration directory
	migrationDir, err := atlasMigrate.NewLocalDir(d.migrationDir)
	if err != nil {
		return fmt.Errorf("failed to create migration dir: %w", err)
	}

	migrate, err := d.newEntMigrate(migrationDir)
	if err != nil {
		return fmt.Errorf("failed to create migrate: %w", err)
	}

	// Use NamedDiff to generate migration files without applying them
	if err = migrate.NamedDiff(ctx, name, d.tables...); err != nil {
		// ErrNoPlan means no changes detected
		if strings.Contains(err.Error(), "no plan") {
			return nil
		}
		return fmt.Errorf("failed to generate diff: %w", err)
	}

	return nil
}

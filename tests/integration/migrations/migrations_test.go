package migrations_test

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/entdbadapter"
	"github.com/fastschema/fastschema/schema"
	toolservice "github.com/fastschema/fastschema/services/tool"
	h "github.com/fastschema/fastschema/tests/integration/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var systemSchemas = []any{
	fs.User{},
	fs.Role{},
	fs.Permission{},
	fs.File{},
	fs.Session{},
	fs.Migration{},
}

// Test context
func ctx() context.Context {
	return context.Background()
}

// createTestSchemaBuilder creates a schema builder for testing
func createTestSchemaBuilder(t *testing.T, schemaDir string) *schema.Builder {
	t.Helper()
	sb, err := schema.NewBuilderFromDir(schemaDir, systemSchemas...)
	require.NoError(t, err)
	return sb
}

// createSQLiteClient creates a SQLite client for testing
func createSQLiteClient(t *testing.T, dbPath, migrationDir string, sb *schema.Builder) db.Client {
	t.Helper()
	h.RemoveAllMigrationFiles(migrationDir)

	client, err := entdbadapter.NewEntClient(&db.Config{
		Driver:       "sqlite",
		Name:         dbPath,
		MigrationDir: migrationDir,
		LogQueries:   false,
	}, sb)
	require.NoError(t, err)

	// Clear auto-generated migration files after client creation
	// This ensures tests start with a clean migration directory
	h.RemoveAllMigrationFiles(migrationDir)

	t.Cleanup(func() { _ = client.Close() })
	return client
}

func testMigrationFlow(t *testing.T, client db.Client, migrationDir string, upSQL, downSQL, verifySQL string) {
	h.RemoveAllMigrationFiles(migrationDir)
	mf := &fs.MigrationFile{
		Version: toolservice.GenerateVersion(),
		Name:    "test_table",
		UpSQL:   upSQL,
		DownSQL: downSQL,
	}
	require.NoError(t, toolservice.WriteMigrationFiles(migrationDir, mf))

	applied, err := toolservice.MigrationUp(ctx(), client, 0)
	require.NoError(t, err)
	assert.Len(t, applied, 1)

	// Verify
	_, err = client.Exec(ctx(), verifySQL)
	require.NoError(t, err)

	// Rollback
	rolledBack, err := toolservice.MigrationDown(ctx(), client, 1)
	require.NoError(t, err)
	assert.Len(t, rolledBack, 1)
}

// TestMigrationCreate tests creating empty migration files
func TestMigrationCreate(t *testing.T) {
	migrationDir := t.TempDir()
	client := createSQLiteClient(t, path.Join(t.TempDir(), "test.db"), migrationDir, createTestSchemaBuilder(t, t.TempDir()))

	// Create empty migration
	mf, err := toolservice.MigrationNew(ctx(), client, "add_custom_data")
	require.NoError(t, err)
	require.NotNil(t, mf)

	// Verify files were created
	assert.FileExists(t, mf.UpFile)
	assert.FileExists(t, mf.DownFile)

	// Verify content is placeholder
	upContent, err := os.ReadFile(mf.UpFile)
	require.NoError(t, err)
	assert.Contains(t, string(upContent), "-- Write your UP migration SQL here")

	downContent, err := os.ReadFile(mf.DownFile)
	require.NoError(t, err)
	assert.Contains(t, string(downContent), "-- Write your DOWN migration SQL here")
}

// TestMigrationStatus tests migration status reporting
func TestMigrationStatus(t *testing.T) {
	schemaDir := t.TempDir()
	migrationDir := t.TempDir()
	sb := createTestSchemaBuilder(t, schemaDir)
	client := createSQLiteClient(t, path.Join(t.TempDir(), "test.db"), migrationDir, sb)

	// Initially no migrations
	applied, pending, err := toolservice.MigrationStatus(ctx(), client)
	require.NoError(t, err)
	assert.Empty(t, applied)
	assert.Empty(t, pending)

	// Create migration files with explicit versions
	for i, name := range []string{"first", "second"} {
		mf := &fs.MigrationFile{
			Version: fmt.Sprintf("2025121210000%d", i),
			Name:    name,
			UpSQL:   "-- " + name,
			DownSQL: "-- " + name,
		}
		require.NoError(t, toolservice.WriteMigrationFiles(migrationDir, mf))
	}

	// Should show 2 pending
	applied, pending, err = toolservice.MigrationStatus(ctx(), client)
	require.NoError(t, err)
	assert.Empty(t, applied)
	assert.Len(t, pending, 2)
}

// TestMigrationUpDown tests applying and rolling back migrations
func TestMigrationUpDown(t *testing.T) {
	schemaDir := t.TempDir()
	migrationDir := t.TempDir()
	sb := createTestSchemaBuilder(t, schemaDir)
	client := createSQLiteClient(t, path.Join(t.TempDir(), "test.db"), migrationDir, sb)

	// Create migration with real SQL
	mf := &fs.MigrationFile{
		Version: toolservice.GenerateVersion(),
		Name:    "create_test_table",
		UpSQL:   "CREATE TABLE migration_test (id INTEGER PRIMARY KEY, name TEXT);",
		DownSQL: "DROP TABLE migration_test;",
	}
	require.NoError(t, toolservice.WriteMigrationFiles(migrationDir, mf))

	// Apply migration
	applied, err := toolservice.MigrationUp(ctx(), client, 0)
	require.NoError(t, err)
	require.Len(t, applied, 1)
	assert.Equal(t, "create_test_table", applied[0].Name)

	// Verify table exists
	_, err = client.Exec(ctx(), "INSERT INTO migration_test (name) VALUES ('test')")
	require.NoError(t, err)

	// Check status
	appliedStatus, pendingStatus, err := toolservice.MigrationStatus(ctx(), client)
	require.NoError(t, err)
	assert.Len(t, appliedStatus, 1)
	assert.Empty(t, pendingStatus)

	// Rollback migration
	rolledBack, err := toolservice.MigrationDown(ctx(), client, 1)
	require.NoError(t, err)
	require.Len(t, rolledBack, 1)

	// Verify table dropped
	_, err = client.Exec(ctx(), "INSERT INTO migration_test (name) VALUES ('test')")
	assert.Error(t, err, "Table should not exist after rollback")

	// Check status again
	appliedStatus, pendingStatus, err = toolservice.MigrationStatus(ctx(), client)
	require.NoError(t, err)
	assert.Empty(t, appliedStatus)
	assert.Len(t, pendingStatus, 1)
}

// TestMigrationUpPartial tests applying specific number of migrations
func TestMigrationUpPartial(t *testing.T) {
	schemaDir := t.TempDir()
	migrationDir := t.TempDir()
	sb := createTestSchemaBuilder(t, schemaDir)
	client := createSQLiteClient(t, path.Join(t.TempDir(), "test.db"), migrationDir, sb)

	// Create 3 migrations with explicit unique versions
	for i := 1; i <= 3; i++ {
		mf := &fs.MigrationFile{
			Version: fmt.Sprintf("2025121212000%d", i),
			Name:    fmt.Sprintf("mig%d", i),
			UpSQL:   fmt.Sprintf("-- migration %d", i),
			DownSQL: "-- rollback",
		}
		require.NoError(t, toolservice.WriteMigrationFiles(migrationDir, mf))
	}

	// Apply only 2
	applied, err := toolservice.MigrationUp(ctx(), client, 2)
	require.NoError(t, err)
	assert.Len(t, applied, 2)

	// Check status
	appliedStatus, pendingStatus, err := toolservice.MigrationStatus(ctx(), client)
	require.NoError(t, err)
	assert.Len(t, appliedStatus, 2)
	assert.Len(t, pendingStatus, 1)
}

// TestMigrationWithMySQL tests migrations with MySQL
func TestMigrationWithMySQL(t *testing.T) {
	for _, cfg := range h.MysqlConfigs {
		t.Run(cfg.Name, func(t *testing.T) {
			schemaDir := t.TempDir()
			migrationDir := filepath.Join(t.TempDir(), "mysql_migrations")
			os.MkdirAll(migrationDir, 0755)

			sb := createTestSchemaBuilder(t, schemaDir)
			client := h.NewMySQLClient(t, cfg, sb, migrationDir).C

			testMigrationFlow(t, client, migrationDir,
				"CREATE TABLE mysql_test (id INT AUTO_INCREMENT PRIMARY KEY, name VARCHAR(255));",
				"DROP TABLE IF EXISTS mysql_test;",
				"INSERT INTO mysql_test (name) VALUES ('test')",
			)
		})
	}
}

// TestMigrationWithPostgres tests migrations with PostgreSQL
func TestMigrationWithPostgres(t *testing.T) {
	for _, config := range h.PostgresConfigs {
		t.Run(config.Name, func(t *testing.T) {
			schemaDir := t.TempDir()
			migrationDir := filepath.Join(t.TempDir(), "postgres_migrations")
			os.MkdirAll(migrationDir, 0755)

			sb := createTestSchemaBuilder(t, schemaDir)
			client := h.NewPostgresClient(t, config, sb, migrationDir).C

			testMigrationFlow(t, client, migrationDir,
				"CREATE TABLE postgres_test (id SERIAL PRIMARY KEY, name VARCHAR(255));",
				"DROP TABLE IF EXISTS postgres_test;",
				"INSERT INTO postgres_test (name) VALUES ('test')",
			)
		})
	}
}

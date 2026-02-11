package helpers

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"entgo.io/ent/dialect"
	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/pkg/entdbadapter"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
)

// DBConfig describes a database target used in integration tests.
type DBConfig struct {
	Name string
	Port int
}

// DBClient wraps a database client alongside the target name for test labelling.
type DBClient struct {
	Name string
	C    db.Client
}

// Ctx returns a background context for integration tests.
func Ctx() context.Context { return context.Background() }

func IsMySQLFamily(name string) bool {
	return strings.HasPrefix(name, "mysql") || strings.HasPrefix(name, "mariadb")
}

func ClearDBData(client db.Client, tables ...string) {
	sqls := []string{}

	if client.Dialect() == dialect.MySQL {
		sqls = append(sqls, "SET FOREIGN_KEY_CHECKS=0")
	}

	if client.Dialect() == dialect.SQLite {
		sqls = append(sqls, "PRAGMA foreign_keys = OFF;")
	}

	if client.Dialect() == dialect.MySQL {
		sqls = append(sqls, strings.Join(utils.Map(tables, func(table string) string {
			return fmt.Sprintf("TRUNCATE TABLE `%s`", table)
		}), ";"))
	}

	if client.Dialect() == dialect.SQLite {
		sqls = append(sqls, strings.Join(utils.Map(tables, func(table string) string {
			return fmt.Sprintf(
				"DELETE FROM %s; DELETE FROM SQLITE_SEQUENCE WHERE name='%s'",
				table,
				table,
			)
		}), ";"))
	}

	if client.Dialect() == dialect.Postgres {
		sqls = append(sqls, fmt.Sprintf(
			"TRUNCATE TABLE %s CASCADE",
			strings.Join(tables, ", "),
		))
		sqls = append(sqls, utils.Map(tables, func(table string) string {
			return fmt.Sprintf(
				"ALTER SEQUENCE IF EXISTS %s_id_seq RESTART WITH 1",
				table,
			)
		})...)
	}

	if client.Dialect() == dialect.MySQL {
		sqls = append(sqls, "SET FOREIGN_KEY_CHECKS=1")
	}

	if client.Dialect() == dialect.SQLite {
		sqls = append(sqls, "PRAGMA foreign_keys = ON;")
	}

	sqls = utils.Filter(sqls, func(sql string) bool {
		return strings.TrimSpace(sql) != ""
	})

	if _, err := client.Exec(
		Ctx(),
		strings.Join(sqls, "; "),
	); err != nil {
		panic(err)
	}
	fmt.Printf("\n")
}

// RemoveAllMigrationFiles deletes generated migration files and atlas sums in dir.
func RemoveAllMigrationFiles(dir string) {
	files := utils.Must(filepath.Glob(path.Join(dir, "*.sql")))
	for _, file := range files {
		if err := os.RemoveAll(file); err != nil {
			panic(err)
		}
	}
	atlasFile := path.Join(dir, "atlas.sum")
	if _, err := os.Stat(atlasFile); err == nil {
		if err := os.Remove(atlasFile); err != nil {
			panic(err)
		}
	}
}

// EnsureMigrationDir returns <base>/<name> after creating it.
func EnsureMigrationDir(base, name string) string {
	dir := path.Join(base, name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		panic(err)
	}
	return dir
}

// NewMySQLClient creates a MySQL-compatible client for integration tests.
func NewMySQLClient(t *testing.T, cfg DBConfig, sb *schema.Builder, migrationDir string) DBClient {
	t.Helper()
	RemoveAllMigrationFiles(migrationDir)
	client := utils.Must(entdbadapter.NewEntClient(&db.Config{
		Driver:       "mysql",
		Name:         "fastschema",
		User:         "root",
		Pass:         "123",
		Host:         "127.0.0.1",
		Port:         strconv.Itoa(cfg.Port),
		MigrationDir: migrationDir,
		LogQueries:   false,
	}, sb))
	t.Cleanup(func() { _ = client.Close() })
	return DBClient{Name: cfg.Name, C: client}
}

// NewPostgresClient creates a PostgreSQL client for integration tests.
func NewPostgresClient(t *testing.T, cfg DBConfig, sb *schema.Builder, migrationDir string) DBClient {
	t.Helper()
	RemoveAllMigrationFiles(migrationDir)
	client := utils.Must(entdbadapter.NewEntClient(&db.Config{
		Driver:       "pgx",
		Name:         "fastschema",
		User:         "postgres",
		Pass:         "123",
		Host:         "localhost",
		Port:         strconv.Itoa(cfg.Port),
		MigrationDir: migrationDir,
		LogQueries:   false,
	}, sb))
	t.Cleanup(func() { _ = client.Close() })
	return DBClient{Name: cfg.Name, C: client}
}

// NewSQLiteClient creates a SQLite client backed by the provided database path.
func NewSQLiteClient(t *testing.T, name, dbPath, migrationDir string, sb *schema.Builder) DBClient {
	t.Helper()
	RemoveAllMigrationFiles(migrationDir)
	client := utils.Must(entdbadapter.NewEntClient(&db.Config{
		Driver:       "sqlite",
		Name:         dbPath,
		MigrationDir: migrationDir,
		LogQueries:   false,
	}, sb))
	t.Cleanup(func() { _ = client.Close() })
	return DBClient{Name: name, C: client}
}

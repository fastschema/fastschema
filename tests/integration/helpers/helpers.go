package helpers

import (
	"context"
	"database/sql"
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
	u "github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	defaultDBName = "fastschema"
	mysqlUser     = "root"
	mysqlPass     = "123"
	mysqlHost     = "127.0.0.1"
	postgresUser  = "postgres"
	postgresPass  = "123"
	postgresHost  = "localhost"
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

func IDUint64(t *testing.T, value any) uint64 {
	t.Helper()
	id, err := u.AnyToUint[uint64](value)
	require.NoError(t, err)
	return id
}

// AssertUint64ID asserts that the value is a valid uint64 ID (for non-system schemas).
func AssertUint64ID(t *testing.T, value any) {
	t.Helper()
	id, err := utils.AnyToInt[int64](value)
	assert.NoError(t, err)
	assert.Greater(t, id, int64(0))
}

// AssertID asserts that the value is a valid ID (either UUID or uint64).
func AssertID(t *testing.T, value any) {
	t.Helper()
	// Check if it's a UUID
	if uuidVal, ok := value.(uuid.UUID); ok {
		assert.NotEqual(t, uuid.Nil, uuidVal)
		return
	}
	// Check if it's a string that can be parsed as UUID
	if strVal, ok := value.(string); ok {
		if _, err := uuid.Parse(strVal); err == nil {
			return
		}
	}
	// Otherwise, check if it's a uint64
	id, err := utils.AnyToInt[int64](value)
	assert.NoError(t, err)
	assert.Greater(t, id, int64(0))
}

// ToJSONID converts an ID to its JSON representation.
// For UUIDs, returns a quoted string like `"abc-def"`.
// For integers, returns the number like `123`.
func ToJSONID(id any) string {
	if uuidVal, ok := id.(uuid.UUID); ok {
		return `"` + uuidVal.String() + `"`
	}
	if strVal, ok := id.(string); ok {
		if _, err := uuid.Parse(strVal); err == nil {
			return `"` + strVal + `"`
		}
	}
	return fmt.Sprintf("%v", id)
}

// IsZeroID checks if an ID is zero/nil.
// For UUIDs, returns true if it's uuid.Nil.
// For integers, returns true if it's 0.
func IsZeroID(id any) bool {
	if id == nil {
		return true
	}
	if uuidVal, ok := id.(uuid.UUID); ok {
		return uuidVal == uuid.Nil
	}
	if strVal, ok := id.(string); ok {
		if parsed, err := uuid.Parse(strVal); err == nil {
			return parsed == uuid.Nil
		}
		return strVal == "" || strVal == "0"
	}
	// Check for integer types
	if intVal, err := utils.AnyToInt[int64](id); err == nil {
		return intVal == 0
	}
	return false
}

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
	dbName := deriveDatabaseName(defaultDBName, migrationDir)
	resetMySQLDatabase(t, cfg, dbName)
	client := utils.Must(entdbadapter.NewEntClient(&db.Config{
		Driver:       "mysql",
		Name:         dbName,
		User:         mysqlUser,
		Pass:         mysqlPass,
		Host:         mysqlHost,
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
	dbName := deriveDatabaseName(defaultDBName, migrationDir)
	resetPostgresDatabase(t, cfg, dbName)
	client := utils.Must(entdbadapter.NewEntClient(&db.Config{
		Driver:       "pgx",
		Name:         dbName,
		User:         postgresUser,
		Pass:         postgresPass,
		Host:         postgresHost,
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
		LogQueries:   true,
	}, sb))
	t.Cleanup(func() { _ = client.Close() })
	return DBClient{Name: name, C: client}
}

func deriveDatabaseName(defaultName, migrationDir string) string {
	cleaned := filepath.Clean(migrationDir)
	parent := filepath.Dir(cleaned)
	grandParent := filepath.Dir(parent)
	suffix := strings.ToLower(filepath.Base(grandParent))
	sanitized := sanitizeIdentifier(suffix)
	if sanitized == "" || sanitized == defaultName {
		return defaultName
	}
	return fmt.Sprintf("%s_%s", defaultName, sanitized)
}

func sanitizeIdentifier(value string) string {
	var b strings.Builder
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		default:
			b.WriteRune('_')
		}
	}
	return strings.Trim(b.String(), "_")
}

func resetMySQLDatabase(t *testing.T, cfg DBConfig, dbName string) {
	t.Helper()
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/?multiStatements=true&parseTime=true", mysqlUser, mysqlPass, mysqlHost, cfg.Port)
	sqlDB, err := sql.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}
	defer sqlDB.Close()
	if _, err := sqlDB.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", quoteMySQLIdent(dbName))); err != nil {
		panic(err)
	}
	if _, err := sqlDB.Exec(fmt.Sprintf("CREATE DATABASE %s CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci", quoteMySQLIdent(dbName))); err != nil {
		panic(err)
	}
}

func resetPostgresDatabase(t *testing.T, cfg DBConfig, dbName string) {
	t.Helper()
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=postgres sslmode=disable", postgresHost, cfg.Port, postgresUser, postgresPass))
	if err != nil {
		panic(err)
	}
	defer conn.Close(ctx)
	if _, err := conn.Exec(ctx, fmt.Sprintf("DROP DATABASE IF EXISTS %s", quotePostgresIdent(dbName))); err != nil {
		panic(err)
	}
	if _, err := conn.Exec(ctx, fmt.Sprintf("CREATE DATABASE %s", quotePostgresIdent(dbName))); err != nil {
		panic(err)
	}
}

func quoteMySQLIdent(name string) string {
	return fmt.Sprintf("`%s`", strings.ReplaceAll(name, "`", "``"))
}

func quotePostgresIdent(name string) string {
	return fmt.Sprintf("\"%s\"", strings.ReplaceAll(name, "\"", "\"\""))
}

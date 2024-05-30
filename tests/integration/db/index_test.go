package db

import (
	"context"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/entdbadapter"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
)

type dbConfig struct {
	name string
	port int
}

type dbClient struct {
	name   string
	client db.Client
}

func Ctx() context.Context {
	return context.Background()
}

func removeAllMigrationFiles(migrationDir string) {
	// Remove all migration files in the directory with prefix .sql
	files := utils.Must(filepath.Glob(path.Join(migrationDir, "*.sql")))

	for _, file := range files {
		if err := os.RemoveAll(file); err != nil {
			panic(err)
		}
	}

	atlasFile := path.Join(migrationDir, "atlas.sum")
	if _, err := os.Stat(atlasFile); err == nil {
		if err := os.Remove(atlasFile); err != nil {
			panic(err)
		}
	}
}

func runTests(t *testing.T, clients []dbClient) {
	for _, client := range clients {
		for _, test := range tests {
			t.Run(client.name+"/"+test.name, func(t *testing.T) {
				test.fn(t, client.client)
			})
		}
	}
}

var systemSchemas = []any{
	fs.Role{},
	fs.Permission{},
	fs.User{},
	fs.File{},
}

var tests = []struct {
	name string
	fn   func(t *testing.T, client db.Client)
}{
	{"DBQueryNode", DBQueryNode},
	{"DBCountNode", DBCountNode},
	{"DBCreateNode", DBCreateNode},
	{"DBCreateNodeEdges", DBCreateNodeEdges},
	{"DBUpdateNodes", DBUpdateNodes},
	{"DBDeleteNodes", DBDeleteNodes},
}

func TestMysql(t *testing.T) {
	runTests(t, utils.Map([]dbConfig{
		{"mysql56", 33061},
		{"mysql57", 33062},
		{"mysql8", 33063},
		{"mariadb", 33064},
		{"mariadb102", 33065},
		{"mariadb103", 33066},
	}, func(sc dbConfig) dbClient {
		sb := utils.Must(schema.NewBuilderFromDir("../../../tests/data/schemas", systemSchemas...))
		removeAllMigrationFiles("../../../tests/data/migrations")
		client := utils.Must(entdbadapter.NewEntClient(&db.Config{
			Driver:       "mysql",
			Name:         "fastschema",
			User:         "root",
			Pass:         "123",
			Host:         "127.0.0.1",
			Port:         strconv.Itoa(sc.port),
			MigrationDir: "../../../tests/data/migrations",
			LogQueries:   false,
		}, sb))

		return dbClient{
			name:   sc.name,
			client: client,
		}
	}))
}

func TestPostgres(t *testing.T) {
	runTests(t, utils.Map([]dbConfig{
		{"postgres10", 54321},
		{"postgres11", 54322},
		{"postgres12", 54323},
		{"postgres13", 54324},
		{"postgres14", 54325},
		{"postgres15", 54326},
	}, func(sc dbConfig) dbClient {
		sb := utils.Must(schema.NewBuilderFromDir("../../../tests/data/schemas", systemSchemas...))
		removeAllMigrationFiles("../../../tests/data/migrations")
		client := utils.Must(entdbadapter.NewEntClient(&db.Config{
			Driver:       "pgx",
			Name:         "fastschema",
			User:         "postgres",
			Pass:         "123",
			Host:         "localhost",
			Port:         strconv.Itoa(sc.port),
			MigrationDir: "../../../tests/data/migrations",
			LogQueries:   false,
		}, sb))

		return dbClient{
			name:   sc.name,
			client: client,
		}
	}))
}

func TestSQLite(t *testing.T) {
	sb := utils.Must(schema.NewBuilderFromDir("../../../tests/data/schemas", systemSchemas...))
	removeAllMigrationFiles("../../../tests/data/migrations")
	client := utils.Must(entdbadapter.NewEntClient(&db.Config{
		Driver:       "sqlite",
		Name:         path.Join(t.TempDir(), "fastschema"),
		MigrationDir: "../../../tests/data/migrations",
		LogQueries:   false,
	}, sb))

	runTests(t, []dbClient{{
		name:   "sqlite",
		client: client,
	}})
}

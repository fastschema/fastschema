package db

import (
	"testing"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
	h "github.com/fastschema/fastschema/tests/integration/helpers"
)

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

const (
	schemaDir    = "../../../tests/integration/db/data/schemas"
	migrationDir = "../../../tests/integration/db/data/migrations"
	sqliteDSN    = "../../../tests/integration/db/data/db_test.db"
)

func runTests(t *testing.T, clients []h.DBClient) {
	for _, client := range clients {
		for _, test := range tests {
			t.Run(client.Name+"/"+test.name, func(t *testing.T) {
				test.fn(t, client.C)
			})
		}
	}
}

func TestMysql(t *testing.T) {
	runTests(t, utils.Map(h.MysqlConfigs, func(sc h.DBConfig) h.DBClient {
		sb := utils.Must(schema.NewBuilderFromDir(schemaDir, systemSchemas...))
		return h.NewMySQLClient(t, sc, sb, migrationDir)
	}))
}

func TestPostgres(t *testing.T) {
	runTests(t, utils.Map(h.PostgresConfigs, func(sc h.DBConfig) h.DBClient {
		sb := utils.Must(schema.NewBuilderFromDir(schemaDir, systemSchemas...))
		return h.NewPostgresClient(t, sc, sb, migrationDir)
	}))
}

func TestSQLite(t *testing.T) {
	sb := utils.Must(schema.NewBuilderFromDir(schemaDir, systemSchemas...))
	client := h.NewSQLiteClient(t, "sqlite", sqliteDSN, migrationDir, sb)
	runTests(t, []h.DBClient{client})
}

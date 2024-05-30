package entdbadapter

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"entgo.io/ent/dialect"
	dialectSql "entgo.io/ent/dialect/sql"
	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
	"github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEntClient(t *testing.T) {
	_, err := NewEntClient(&db.Config{
		Driver: "invalid",
	}, &schema.Builder{})
	assert.Equal(t, `sql: unknown driver "invalid" (forgotten import?)`, err.Error())

	_, err = NewEntClient(&db.Config{
		Driver: "mysql",
	}, nil)
	assert.Equal(t, `schema builder is required`, err.Error())

	sql.Register("mysql2", &mysql.MySQLDriver{})

	_, err = NewEntClient(&db.Config{
		Driver: "mysql2",
	}, &schema.Builder{})
	assert.Equal(t, `unsupported driver: mysql2`, err.Error())

	sb := createSchemaBuilder()
	dbClient, err := NewMockExpectClient(func(d *sql.DB) db.Client {
		driver := utils.Must(NewEntClient(&db.Config{
			Driver: "sqlmock",
		}, sb, dialectSql.OpenDB(dialect.MySQL, d)))
		return driver
	}, nil, func(m sqlmock.Sqlmock) {
		m.ExpectBegin()
		m.ExpectQuery("SELECT 1").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	}, false)
	require.NoError(t, err)
	client := dbClient.(EntAdapter)
	assert.NotNil(t, client)

	tx, err := client.Tx(context.Background())
	require.NoError(t, err)
	assert.NotNil(t, tx)

	assert.NotNil(t, client.SchemaBuilder())

	assert.Equal(t, dialect.MySQL, client.Driver().Dialect())
	assert.Equal(t, false, client.IsTx())
	assert.Equal(t, nil, client.Rollback())
	assert.Equal(t, nil, client.Commit())
	_, err = client.Query(context.Background(), "SELECT 1", []any{})
	assert.Equal(t, nil, err)
}

func TestNewClient(t *testing.T) {
	config := &db.Config{
		Driver: "sqlmock",
	}

	sb := createSchemaBuilder()

	client, err := NewClient(config, sb)
	require.NoError(t, err)
	assert.NotNil(t, client)
}

func TestNewTestClient(t *testing.T) {
	schemaBuilder := &schema.Builder{}
	client, err := NewTestClient(utils.Must(os.MkdirTemp("", "migrations")), schemaBuilder)
	assert.NoError(t, err)
	assert.NotNil(t, client)
}

func TestAdapterReloadErrorTableNotFound(t *testing.T) {
	// Create a new schema builder
	newSchemaBuilder := createSchemaBuilder()

	// Create a mock migration
	migration := &db.Migration{
		RenameTables: []*db.RenameItem{
			{
				Type:            "table",
				From:            "old_table",
				To:              "new_table",
				IsJunctionTable: true,
			},
		},
	}

	// Create a mock adapter
	adapter := &Adapter{
		config: &db.Config{
			Driver: "sqlmock",
		},
	}

	// Call the Reload function
	_, err := adapter.Reload(context.Background(), newSchemaBuilder, migration)
	assert.Error(t, err)
}

func TestAdapterReloadError(t *testing.T) {
	// Create a new schema builder
	newSchemaBuilder := createSchemaBuilder()

	// Create a mock migration
	migration := &db.Migration{
		RenameTables: []*db.RenameItem{
			{
				Type:            "table",
				From:            "user",
				To:              "user",
				IsJunctionTable: true,
			},
		},
	}

	// Create a mock adapter
	adapter := createMockAdapter(t)
	// Call the Reload function
	_, err := adapter.Reload(context.Background(), newSchemaBuilder, migration)
	require.Error(t, err)
}

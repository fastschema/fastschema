package entdbadapter

import (
	"context"
	"database/sql"
	"testing"

	"entgo.io/ent/dialect"
	dialectSql "entgo.io/ent/dialect/sql"
	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/fastschema/fastschema/app"
	"github.com/fastschema/fastschema/pkg/testutils"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEntClient(t *testing.T) {
	_, err := NewEntClient(&app.DBConfig{
		Driver: "invalid",
	}, nil)
	assert.Equal(t, `sql: unknown driver "invalid" (forgotten import?)`, err.Error())

	sql.Register("mysql2", &mysql.MySQLDriver{})

	_, err = NewEntClient(&app.DBConfig{
		Driver: "mysql2",
	}, nil)
	assert.Equal(t, `unsupported driver: mysql2`, err.Error())

	sb := createSchemaBuilder()
	dbClient, err := testutils.NewMockClient(func(d *sql.DB) app.DBClient {
		driver := utils.Must(NewEntClient(&app.DBConfig{
			Driver: "sqlmock",
		}, sb, dialectSql.OpenDB(dialect.MySQL, d)))
		return driver
	}, nil, func(m sqlmock.Sqlmock) {
		m.ExpectBegin()
		m.ExpectExec("SELECT 1").WillReturnResult(sqlmock.NewResult(1, 1))
	}, false)
	require.NoError(t, err)
	client := dbClient.(*Adapter)
	assert.NotNil(t, client)

	tx, err := client.Tx(context.Background())
	require.NoError(t, err)
	assert.NotNil(t, tx)

	assert.NotNil(t, client.SchemaBuilder())

	assert.Equal(t, dialect.MySQL, client.Driver().Dialect())
	assert.Equal(t, false, client.IsTx())
	assert.Equal(t, nil, client.Rollback())
	assert.Equal(t, nil, client.Commit())
	assert.Equal(t, nil, client.Exec(context.Background(), "SELECT 1", []any{}, nil))
}
func TestNewClient(t *testing.T) {
	config := &app.DBConfig{
		Driver: "sqlmock",
	}

	sb := createSchemaBuilder()

	client, err := NewClient(config, sb)
	require.NoError(t, err)
	assert.NotNil(t, client)
}

func TestAdapterReloadErrorTableNotFound(t *testing.T) {
	// Create a new schema builder
	newSchemaBuilder := createSchemaBuilder()

	// Create a mock migration
	migration := &app.Migration{
		RenameTables: []*app.RenameItem{
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
		config: &app.DBConfig{
			Driver: "sqlmock",
		},
	}

	// Call the Reload function
	_, err := adapter.Reload(newSchemaBuilder, migration)
	assert.Error(t, err)
}

func TestAdapterReloadError(t *testing.T) {
	// Create a new schema builder
	newSchemaBuilder := createSchemaBuilder()

	// Create a mock migration
	migration := &app.Migration{
		RenameTables: []*app.RenameItem{
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
	_, err := adapter.Reload(newSchemaBuilder, migration)
	require.Error(t, err)
}

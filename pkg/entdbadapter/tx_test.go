package entdbadapter

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"os"
	"testing"

	"entgo.io/ent/dialect"
	dialectSql "entgo.io/ent/dialect/sql"
	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
	"github.com/stretchr/testify/assert"
)

func createTestSchemaBuilder(t *testing.T) *schema.Builder {
	sb := &schema.Builder{}

	groupSchema := &schema.Schema{}
	assert.Nil(t, json.Unmarshal([]byte(groupSchemaJSON), groupSchema))
	assert.NoError(t, groupSchema.Init(false))

	carSchema := &schema.Schema{}
	assert.Nil(t, json.Unmarshal([]byte(carSchemaJSON), carSchema))
	assert.NoError(t, carSchema.Init(false))

	sb.AddSchema(groupSchema)
	sb.AddSchema(carSchema)
	assert.NoError(t, sb.Init())

	return sb
}

func createTx(t *testing.T, client db.Client, sb *schema.Builder) *Tx {
	tx := utils.Must(NewTx(context.Background(), client))
	assert.Equal(t, sb, tx.SchemaBuilder())

	return tx
}

func TestNewTx(t *testing.T) {
	sb := createTestSchemaBuilder(t)
	client := utils.Must(NewTestClient(
		utils.Must(os.MkdirTemp("", "migrations")),
		sb,
	))
	tx := createTx(t, client, sb)
	assert.NotNil(t, utils.Must(tx.Model("car")))

	userModel, err := tx.Model("user")
	assert.Nil(t, userModel)
	assert.Error(t, err)
	assert.NotNil(t, tx.Driver())
	assert.NotNil(t, tx.Config())
	assert.NotNil(t, tx.DB())
	assert.NotNil(t, tx.Dialect())
	assert.Nil(t, tx.Migrate(context.Background(), nil, false))
	assert.Equal(t, true, tx.IsTx())
	assert.Equal(t, tx, utils.Must(tx.Tx(context.Background())))
	tx.SetSQLDB(nil)
	tx.SetDriver(nil)
	_, err = tx.Reload(
		context.Background(),
		tx.SchemaBuilder(),
		nil,
		false,
	)
	assert.NoError(t, err)

	txDriver := tx.driver.(*TxDriver)
	assert.NotNil(t, txDriver.ID())
	assert.NotNil(t, txDriver.Dialect())
	assert.NoError(t, txDriver.Commit())
	assert.NoError(t, txDriver.Rollback())
	assert.NoError(t, txDriver.Close())
	assert.Equal(t, txDriver, utils.Must(txDriver.Tx(context.Background())))

	// Invalid edge
	_, err = tx.NewEdgeSpec(&schema.Relation{}, []driver.Value{1})
	assert.Error(t, err)

	// Valid edge
	edge, err := tx.NewEdgeSpec(tx.SchemaBuilder().Relations()[0], []driver.Value{1})
	assert.NoError(t, err)
	assert.NotNil(t, edge)

	// Invalid edge step
	_, err = tx.NewEdgeStepOption(&schema.Relation{})
	assert.Error(t, err)

	// Valid edge step
	step, err := tx.NewEdgeStepOption(tx.SchemaBuilder().Relations()[0])
	assert.NoError(t, err)
	assert.NotNil(t, step)
}

func TestTxCommit(t *testing.T) {
	sb := createTestSchemaBuilder(t)
	mdb, mock, err := sqlmock.New()
	assert.NoError(t, err)
	client := utils.Must(NewEntClient(&db.Config{
		Driver: "sqlmock",
	}, sb, dialectSql.OpenDB(dialect.MySQL, mdb)))

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT 1").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectExec("SELECT 2").WillReturnResult(sqlmock.NewResult(2, 1))
	mock.ExpectCommit()

	tx := createTx(t, client, sb)
	_, err1 := tx.Query(context.Background(), "SELECT 1")
	_, err2 := tx.Exec(context.Background(), "SELECT 2")
	assert.Nil(t, err1)
	assert.Nil(t, err2)
	assert.NoError(t, tx.Commit())
}

func TestTxRollback(t *testing.T) {
	sb := createTestSchemaBuilder(t)
	mdb, mock, err := sqlmock.New()
	assert.NoError(t, err)
	client := utils.Must(NewEntClient(&db.Config{
		Driver: "sqlmock",
	}, sb, dialectSql.OpenDB(dialect.MySQL, mdb)))

	mock.ExpectBegin()
	mock.ExpectExec("SELECT 1").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("SELECT 2").WillReturnResult(sqlmock.NewResult(2, 1))
	mock.ExpectRollback()

	tx := createTx(t, client, sb)
	_, err1 := tx.Exec(context.Background(), "SELECT 1")
	_, err2 := tx.Exec(context.Background(), "SELECT 2")
	assert.Nil(t, err1)
	assert.Nil(t, err2)
	assert.NoError(t, tx.Rollback())
}

func TestTxClose(t *testing.T) {
	sb := createTestSchemaBuilder(t)
	mdb, mock, err := sqlmock.New()
	assert.NoError(t, err)
	client := utils.Must(NewEntClient(&db.Config{
		Driver: "sqlmock",
	}, sb, dialectSql.OpenDB(dialect.MySQL, mdb)))

	mock.ExpectBegin()
	mock.ExpectExec("SELECT 1").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("SELECT 2").WillReturnResult(sqlmock.NewResult(2, 1))
	mock.ExpectClose()

	tx := createTx(t, client, sb)
	_, err1 := tx.Exec(context.Background(), "SELECT 1")
	_, err2 := tx.Exec(context.Background(), "SELECT 2")
	assert.Nil(t, err1)
	assert.Nil(t, err2)
	assert.NoError(t, tx.Close())
}

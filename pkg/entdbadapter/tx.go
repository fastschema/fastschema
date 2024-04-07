package entdbadapter

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"reflect"

	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/fastschema/fastschema/app"
	"github.com/fastschema/fastschema/schema"
)

// Tx hold the transaction and the schema manager.
type Tx struct {
	ctx    context.Context
	driver dialect.Driver
	client app.DBClient
	config *app.DBConfig
}

// NewTx creates a new transaction.
func NewTx(ctx context.Context, client app.DBClient) (*Tx, error) {
	entAdapter := client.(*Adapter)
	driver := entAdapter.Driver()
	tx, err := driver.Tx(ctx)
	if err != nil {
		return nil, err
	}

	txd := &Tx{
		ctx:    ctx,
		driver: &TxDriver{driver: driver, dialectTx: tx},
		client: client,
		config: client.Config(),
	}

	return txd, nil
}

func (tx *Tx) NewEdgeSpec(r *schema.Relation, nodeIDs []driver.Value) (*sqlgraph.EdgeSpec, error) {
	entAdapter, ok := tx.client.(*Adapter)
	if !ok {
		return nil, fmt.Errorf("client is not an ent adapter")
	}

	return entAdapter.NewEdgeSpec(r, nodeIDs)
}

func (tx *Tx) NewEdgeStepOption(r *schema.Relation) (sqlgraph.StepOption, error) {
	entAdapter, ok := tx.client.(*Adapter)
	if !ok {
		return nil, fmt.Errorf("client is not an ent adapter")
	}
	return entAdapter.NewEdgeStepOption(r)
}

func (tx *Tx) Config() *app.DBConfig {
	return tx.config
}

func (tx *Tx) Hooks() *app.Hooks {
	return tx.client.Hooks()
}

func (tx *Tx) DB() *sql.DB {
	return tx.client.DB()
}

// Reload reloads the schema.
func (tx *Tx) Reload(newSchemaBuilder *schema.Builder, migration *app.Migration) (app.DBClient, error) {
	return tx.client.Reload(newSchemaBuilder, migration)
}

// SchemaBuilder returns the schema builder.
func (tx *Tx) SchemaBuilder() *schema.Builder {
	return tx.client.SchemaBuilder()
}

// Model returns the model by name.
func (tx *Tx) Model(name string) (app.Model, error) {
	m, err := tx.client.Model(name)
	if err != nil {
		return nil, err
	}

	return m.Clone().SetClient(tx), nil
}

// Dialect returns the dialect name.
func (tx *Tx) Dialect() string {
	return tx.driver.Dialect()
}

// Driver returns the underlying driver.
func (tx *Tx) Driver() dialect.Driver {
	return tx.driver
}

// CreateDBModel creates a new model from the schema.
func (tx *Tx) CreateDBModel(s *schema.Schema, relations ...*schema.Relation) app.Model {
	return tx.client.CreateDBModel(s, relations...)
}

// Exec executes a query.
func (tx *Tx) Exec(ctx context.Context, query string, args any, bindValue any) error {
	return tx.driver.Exec(ctx, query, args, bindValue)
}

// Close closes the transaction.
func (tx *Tx) Close() error {
	return tx.driver.Close()
}

// Rollback rollbacks the transaction.
func (tx *Tx) Rollback() error {
	txDriver := tx.driver.(*TxDriver)
	return txDriver.dialectTx.Rollback()
}

// Commit commits the transaction.
func (tx *Tx) Commit() error {
	txDriver := tx.driver.(*TxDriver)
	return txDriver.dialectTx.Commit()
}

// IsTx returns true if the client is a transaction.
func (tx *Tx) IsTx() bool {
	return true
}

// Tx returns the transaction.
func (tx *Tx) Tx(ctx context.Context) (t app.DBClient, err error) {
	return tx, nil
}

// TxDriver is the driver for transaction.
type TxDriver struct {
	driver    dialect.Driver // the driver we started the transaction from.
	dialectTx dialect.Tx     // tx is the underlying transaction.
}

func (tx *TxDriver) Close() error                           { return nil }
func (tx *TxDriver) Commit() error                          { return nil }
func (tx *TxDriver) Rollback() error                        { return nil }
func (tx *TxDriver) Dialect() string                        { return tx.driver.Dialect() }
func (tx *TxDriver) Tx(context.Context) (dialect.Tx, error) { return tx, nil }

// ID returns the transaction id.
func (tx *TxDriver) ID() string {
	debugTx, _ := tx.dialectTx.(*dialect.DebugTx)
	debugTxValue := reflect.ValueOf(*debugTx)
	return debugTxValue.FieldByName("id").String()
}

// Exec calls tx.Exec.
func (tx *TxDriver) Exec(ctx context.Context, query string, args, v any) error {
	return tx.dialectTx.Exec(ctx, query, args, v)
}

// Query calls tx.Query.
func (tx *TxDriver) Query(ctx context.Context, query string, args, v any) error {
	return tx.dialectTx.Query(ctx, query, args, v)
}

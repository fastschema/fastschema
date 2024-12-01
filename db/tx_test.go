package db_test

import (
	"testing"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/fs"
	"github.com/stretchr/testify/assert"
)

func TestWithTxSuccess(t *testing.T) {
	client, ctx := prepareTest()
	err := db.WithTx(client, ctx, func(tx db.Client) error {
		_, err := db.Create[TestCategory](ctx, tx, fs.Map{
			"name": "category 1",
		})
		assert.NoError(t, err)

		_, err = db.Create[TestCategory](ctx, tx, fs.Map{
			"name": "category 2",
		})
		assert.NoError(t, err)

		return nil
	})
	assert.NoError(t, err)
}

func TestWithTxCreateTxError(t *testing.T) {
	client, ctx := prepareTest()
	client.Close()
	err := db.WithTx(client, ctx, func(tx db.Client) error {
		return nil
	})
	assert.Error(t, err)
}

func TestWithTxCommitTwice(t *testing.T) {
	client, ctx := prepareTest()
	err := db.WithTx(client, ctx, func(tx db.Client) error {
		_, err := db.Create[TestCategory](ctx, tx, fs.Map{
			"name": "category 1",
		})
		assert.NoError(t, err)

		_, err = db.Create[TestCategory](ctx, tx, fs.Map{
			"name": "category 2",
		})
		assert.NoError(t, err)

		return tx.Commit()
	})
	assert.Error(t, err)
}

func TestWithTxReturnError(t *testing.T) {
	client, ctx := prepareTest()
	err := db.WithTx(client, ctx, func(tx db.Client) error {
		_, err := db.Create[TestCategory](ctx, tx, fs.Map{
			"name": "category 1",
		})
		assert.NoError(t, err)

		_, err = db.Create[TestCategory](ctx, tx, fs.Map{
			"name": "category 2",
		})
		assert.NoError(t, err)

		return assert.AnError
	})
	assert.Error(t, err)
}

func TestWithTxReturnErrorRollbackTwice(t *testing.T) {
	client, ctx := prepareTest()
	err := db.WithTx(client, ctx, func(tx db.Client) error {
		_, err := db.Create[TestCategory](ctx, tx, fs.Map{
			"name": "category 1",
		})
		assert.NoError(t, err)

		_, err = db.Create[TestCategory](ctx, tx, fs.Map{
			"name": "category 2",
		})
		assert.NoError(t, err)

		assert.NoError(t, tx.Rollback())

		return assert.AnError
	})
	assert.Error(t, err)
}

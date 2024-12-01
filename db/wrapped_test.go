package db_test

// Tests for wrapped.go

import (
	"context"
	"testing"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestWrappedClient(t *testing.T) {
	client, ctx := prepareTest()

	wrapped := db.NewWrappedClient(func() db.Client {
		return client
	})

	assert.Equal(t, client.Dialect(), wrapped.Dialect())
	assert.Equal(t, client.IsTx(), wrapped.IsTx())
	assert.Equal(t, client.SchemaBuilder(), wrapped.SchemaBuilder())
	assert.Equal(t, client.DB(), wrapped.DB())
	assert.Equal(t, client.Hooks(), wrapped.Hooks())
	assert.Equal(t, client.Config(), wrapped.Config())
	assert.Equal(t, utils.Must(client.Model("category")), utils.Must(wrapped.Model("category")))
	assert.NotNil(t, utils.Must(wrapped.Exec(ctx, "SELECT 1")))
	assert.NotNil(t, utils.Must(wrapped.Query(ctx, "SELECT 1")))
	assert.NotNil(t, utils.Must(wrapped.Tx(ctx)))
	assert.NotNil(t, utils.Must(wrapped.Reload(ctx, wrapped.SchemaBuilder(), nil, false)))
	assert.Nil(t, wrapped.Rollback())
	assert.Nil(t, wrapped.Commit())
	assert.Nil(t, wrapped.Close())
}

func TestNoopClient(t *testing.T) {
	ctx := context.Background()
	client := &db.NoopClient{}

	assert.Equal(t, "", client.Dialect())
	assert.Equal(t, false, client.IsTx())
	assert.Nil(t, client.SchemaBuilder())
	assert.Nil(t, client.DB())
	assert.Nil(t, client.Hooks())
	assert.Nil(t, client.Config())
	assert.Nil(t, client.Rollback())
	assert.Nil(t, client.Commit())
	assert.Nil(t, client.Close())
	assert.Nil(t, utils.Must(client.Tx(ctx)))
	assert.Nil(t, utils.Must(client.Reload(ctx, nil, nil, false)))
	assert.Nil(t, utils.Must(client.Model("category")))
	assert.Nil(t, utils.Must(client.Exec(ctx, "SELECT 1")))
	assert.Nil(t, utils.Must(client.Query(ctx, "SELECT 1")))
}

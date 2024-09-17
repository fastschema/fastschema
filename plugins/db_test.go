package plugins_test

import (
	"context"
	"os"
	"testing"

	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/entdbadapter"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/plugins"
	"github.com/fastschema/fastschema/schema"
	"github.com/stretchr/testify/assert"
)

func creatPluginsDB() *plugins.DB {
	migrationDir := utils.Must(os.MkdirTemp("", "migration"))
	sb := utils.Must(schema.NewBuilderFromDir(
		utils.Must(os.MkdirTemp("", "schemas")),
		fs.SystemSchemaTypes...,
	))
	dbc := utils.Must(entdbadapter.NewTestClient(
		migrationDir,
		sb,
	))

	return plugins.NewDB(dbc)
}

func TestBuilder(t *testing.T) {
	pluginsDB := creatPluginsDB()
	_, err := pluginsDB.Builder("invalid_schema")
	assert.Error(t, err)

	builder, err := pluginsDB.Builder("user")
	assert.NoError(t, err)
	assert.NotNil(t, builder)

	_, err = builder.Where(map[string]interface{}{
		"invalid_field": 1,
	})
	assert.Error(t, err)

	_, err = builder.Where(map[string]interface{}{
		"id": 1,
	})
	assert.NoError(t, err)
}

func TestDBOperations(t *testing.T) {
	ctx := context.Background()
	pluginsDB := creatPluginsDB()
	builder := utils.Must(pluginsDB.Builder("role"))

	// Create role using builder
	testRole, err := builder.Create(ctx, map[string]interface{}{
		"name": "testrole",
	})
	assert.NoError(t, err)
	assert.NotNil(t, testRole)

	// Query role using raw query
	roles, err := pluginsDB.Query(ctx, "SELECT * FROM roles WHERE name = ?", "testrole")
	assert.NoError(t, err)
	assert.Len(t, roles, 1)

	// Update role using builder
	builder, err = builder.Where(map[string]interface{}{
		"id": testRole.ID(),
	})
	assert.NoError(t, err)
	updated, err := builder.Update(ctx, map[string]interface{}{
		"name": "updatedrole",
	})
	assert.NoError(t, err)
	assert.NotNil(t, updated)

	// Query role using builder
	newRole, err := utils.Must(utils.Must(pluginsDB.Builder("role")).Where(map[string]interface{}{
		"id": testRole.ID(),
	})).First(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "updatedrole", newRole.Get("name"))

	// Delete role using exec
	result, err := pluginsDB.Exec(ctx, "DELETE FROM roles WHERE id = ?", testRole.ID())
	assert.NoError(t, err)
	assert.Equal(t, int64(1), utils.Must(result.RowsAffected()))

	// Transaction rollback
	tx := utils.Must(pluginsDB.Tx(ctx))

	result, err = tx.Exec(ctx, "INSERT INTO roles (name) VALUES (?)", "testrole")
	assert.NoError(t, err)
	lastInsertID := utils.Must(result.LastInsertId())
	assert.Greater(t, lastInsertID, int64(0))

	assert.NoError(t, tx.Rollback())

	roles, err = pluginsDB.Query(ctx, "SELECT * FROM roles WHERE name = ?", "testrole")
	assert.NoError(t, err)
	assert.Len(t, roles, 0)

	// Transaction commit
	tx = utils.Must(pluginsDB.Tx(ctx))

	result, err = tx.Exec(ctx, "INSERT INTO roles (name) VALUES (?)", "testrole")
	assert.NoError(t, err)
	lastInsertID = utils.Must(result.LastInsertId())
	assert.Greater(t, lastInsertID, int64(0))

	assert.NoError(t, tx.Commit())

	roles, err = pluginsDB.Query(ctx, "SELECT * FROM roles WHERE name = ?", "testrole")
	assert.NoError(t, err)
	assert.Len(t, roles, 1)
}

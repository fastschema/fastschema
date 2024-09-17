package db_test

import (
	"fmt"
	"testing"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
	"github.com/stretchr/testify/assert"
)

func TestRawQuery(t *testing.T) {
	client, ctx := prepareTest()

	for i := 1; i <= 5; i++ {
		_, err := db.Create[TestCategory](ctx, client, fs.Map{
			"name": fmt.Sprintf("category %d", i),
		})
		assert.NoError(t, err)
	}

	// Case 1: Invalid query
	_, err := db.Query[TestCategory](ctx, client, "SELECT * FROM categories ?", 0)
	assert.Error(t, err)

	// Case 2: Success scan to struct
	rows, err := db.Query[TestCategory](ctx, client, "SELECT * FROM categories WHERE id = ?", 5)
	assert.NoError(t, err)
	assert.Len(t, rows, 1)
	assert.Equal(t, uint64(5), rows[0].ID)
	assert.Equal(t, "category 5", rows[0].Name)

	// Case 3: Success scan to entity
	entities, err := db.Query[*schema.Entity](ctx, client, "SELECT * FROM categories WHERE id > ?", 1)
	assert.NoError(t, err)
	assert.Len(t, entities, 4)
}

func TestExec(t *testing.T) {
	client, ctx := prepareTest()

	for i := 1; i <= 5; i++ {
		_, err := db.Create[TestCategory](ctx, client, fs.Map{
			"name": fmt.Sprintf("category %d", i),
		})
		assert.NoError(t, err)
	}

	// Case 1: Invalid query
	_, err := db.Exec(ctx, client, "UPDATE categories SET WHERE id = ?", 0)
	assert.Error(t, err)

	// Case 2: Success
	result, err := db.Exec(ctx, client, "UPDATE categories SET name = ? WHERE id > ?", "new name", 1)
	assert.NoError(t, err)
	assert.Equal(t, int64(4), utils.Must(result.RowsAffected()))
}

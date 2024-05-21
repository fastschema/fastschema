package db_test

import (
	"testing"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/schema"
	"github.com/stretchr/testify/assert"
)

func TestMutation(t *testing.T) {
	client, ctx := prepareTest()

	// Case 1: Create with invalid model
	_, err := db.Create[testPost](ctx, client, schema.NewEntity().Set("title", "tutorial"))
	assert.Contains(t, err.Error(), "model testPost not found")

	// Case 2: Create error
	_, err = db.Create[TestCategory](ctx, client, schema.NewEntity().Set("slug", "tutorial"))
	assert.Contains(t, err.Error(), "column category.slug not found")

	// Case 3: Create success
	jsonCategory, err := db.Create[TestCategory](ctx, client, schema.NewEntity().Set("name", "Tutorial"))
	assert.NoError(t, err)
	assert.Equal(t, "Tutorial", jsonCategory.Name)

	// Case 4: CreateFromJSON error
	_, err = db.CreateFromJSON[TestCategory](ctx, client, `{`)
	assert.Contains(t, err.Error(), "JSON error")

	// Case 5: CreateFromJSON success
	jsonCategory, err = db.CreateFromJSON[TestCategory](ctx, client, `{"name": "Tutorial 2"}`)
	assert.NoError(t, err)
	assert.Equal(t, "Tutorial 2", jsonCategory.Name)

	// Case 6: Invalid model
	_, err = db.Update[testPost](ctx, client, schema.NewEntity().Set("title", "tutorial"))
	assert.Contains(t, err.Error(), "model testPost not found")

	// Case 7: Update error
	_, err = db.Update[TestCategory](ctx, client, schema.NewEntity().Set("slug", "tutorial"))
	assert.Contains(t, err.Error(), "column category.slug not found")

	// Case 8: Update success
	updatedCategories, err := db.Update[TestCategory](
		ctx,
		client,
		schema.NewEntity().Set("name", "Tutorial updated"),
		db.EQ("id", jsonCategory.ID),
	)
	assert.NoError(t, err)
	assert.Equal(t, "Tutorial updated", updatedCategories[0].Name)

	// Case 9: Delete error invalid model
	_, err = db.Delete[testPost](ctx, client)
	assert.Contains(t, err.Error(), "model testPost not found")

	// Case 10: Delete error
	_, err = db.Delete[TestCategory](ctx, client, db.EQ("slug", "tutorial"))
	assert.Contains(t, err.Error(), "no such column")

	// Case 11: Delete success
	affected, err := db.Delete[TestCategory](ctx, client, db.EQ("id", jsonCategory.ID))
	assert.NoError(t, err)
	assert.Equal(t, 1, affected)
}

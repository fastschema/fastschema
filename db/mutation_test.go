package db_test

import (
	"testing"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/entity"
	"github.com/fastschema/fastschema/fs"
	"github.com/stretchr/testify/assert"
)

func TestMutation(t *testing.T) {
	client, ctx := prepareTest()

	// Case 1: Create with invalid model
	_, err := db.Create[testPost](ctx, client, fs.Map{"title": "tutorial"})
	assert.Contains(t, err.Error(), "model testPost not found")

	// Case 2: Create error invalid column
	_, err = db.Create[TestCategory](ctx, client, fs.Map{"slug": "tutorial"})
	assert.Contains(t, err.Error(), "column category.slug not found")

	// Case 3: Create error invalid type
	_, err = db.Create[TestCategory](ctx, client, "invalid")
	assert.Contains(t, err.Error(), "cannot create entity")

	// Case 4: Create success
	jsonCategory, err := db.Create[TestCategory](ctx, client, fs.Map{"name": "Tutorial"})
	assert.NoError(t, err)
	assert.Equal(t, "Tutorial", jsonCategory.Name)

	// Case 5: CreateFromJSON error
	_, err = db.CreateFromJSON[TestCategory](ctx, client, `{`)
	assert.Contains(t, err.Error(), "JSON error")

	// Case 6: CreateFromJSON success
	jsonCategory, err = db.CreateFromJSON[TestCategory](ctx, client, `{"name": "Tutorial 2"}`)
	assert.NoError(t, err)
	assert.Equal(t, "Tutorial 2", jsonCategory.Name)

	// Case 7: Invalid model
	_, err = db.Update[testPost](ctx, client, fs.Map{"title": "tutorial"}, nil)
	assert.Contains(t, err.Error(), "model testPost not found")

	// Case 8: Create with entity.Entity: Invalid schema name
	_, err = db.Create[*entity.Entity](ctx, client, fs.Map{"name": "Tutorial"}, "invalid")
	assert.ErrorContains(t, err, "model invalid not found")

	// Case 9: Create with entity.Entity: no schema name
	_, err = db.Create[*entity.Entity](ctx, client, fs.Map{"name": "Tutorial"})
	assert.ErrorContains(t, err, "schema name is required for type entity.Entity")

	// Case 10: Create with entity.Entity: Success
	cat, err := db.Create[*entity.Entity](ctx, client, fs.Map{"name": "Tutorial"}, "category")
	assert.NoError(t, err)
	assert.Equal(t, "Tutorial", cat.Get("name"))

	// Case 11: Update error
	_, err = db.Update[TestCategory](ctx, client, fs.Map{"slug": "tutorial"}, nil)
	assert.Contains(t, err.Error(), "column category.slug not found")

	// Case 12: Update success
	updatedCategories, err := db.Update[TestCategory](
		ctx,
		client,
		fs.Map{"name": "Tutorial updated"},
		[]*db.Predicate{db.EQ("id", jsonCategory.ID)},
	)
	assert.NoError(t, err)
	assert.Equal(t, "Tutorial updated", updatedCategories[0].Name)

	// Case 13: Update with entity.Entity: Invalid data
	_, err = db.Update[entity.Entity](ctx, client, "invalid", nil)
	assert.ErrorContains(t, err, "cannot create entity")

	// Case 14: Update with entity.Entity: Success
	updatedCats, err := db.Update[*entity.Entity](
		ctx,
		client,
		fs.Map{"name": "Tutorial updated 2"},
		[]*db.Predicate{db.EQ("id", jsonCategory.ID)},
		"category",
	)
	assert.NoError(t, err)
	assert.Equal(t, "Tutorial updated 2", updatedCats[0].Get("name"))

	// Case 15: Delete error invalid model
	_, err = db.Delete[testPost](ctx, client, nil)
	assert.Contains(t, err.Error(), "model testPost not found")

	// Case 16: Delete error
	_, err = db.Delete[TestCategory](ctx, client, []*db.Predicate{db.EQ("slug", "tutorial")})
	assert.Contains(t, err.Error(), "no such column")

	// Case 17: Delete success
	affected, err := db.Delete[TestCategory](ctx, client, []*db.Predicate{db.EQ("id", jsonCategory.ID)})
	assert.NoError(t, err)
	assert.Equal(t, 1, affected)
}

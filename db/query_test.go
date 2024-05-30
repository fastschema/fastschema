package db_test

import (
	"fmt"
	"testing"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/schema"
	"github.com/stretchr/testify/assert"
)

func TestQuery(t *testing.T) {
	client, ctx := prepareTest()

	for i := 1; i <= 5; i++ {
		_, err := db.Create[TestCategory](ctx, client, fs.Map{
			"name": fmt.Sprintf("category %d", i),
		})
		assert.NoError(t, err)
	}

	// Case 1: Get invalid model.
	_, err := db.Query[testPost](client).
		Where(db.EQ("id", 1)).
		Get(ctx)
	assert.Error(t, err)

	// Case 2: Query invalid column.
	_, err = db.Query[TestCategory](client).
		Where(db.EQ("invalid_column", 1)).
		Get(ctx)
	assert.Error(t, err)

	// Case 3: Query success.
	categories, err := db.Query[TestCategory](client).
		Where(db.GTE("id", 3)).
		Order("-id").
		Limit(1).
		Offset(1).
		Select("id", "name").
		Get(ctx)

	assert.NoError(t, err)
	assert.Len(t, categories, 1)
	assert.Equal(t, uint64(4), categories[0].ID)
	assert.Equal(t, "category 4", categories[0].Name)

	// Case 4: Count invalid model.
	_, err = db.Query[testPost](client).Count(ctx, nil)
	assert.Error(t, err)

	// Case 5: Count invalid filter column.
	_, err = db.Query[TestCategory](client).
		Where(db.EQ("invalid_column", 1)).
		Count(ctx, nil)
	assert.Error(t, err)

	// Case 6: Count success.
	count, err := db.Query[TestCategory](client).
		Where(db.GTE("id", 3)).
		Count(ctx, nil)
	assert.NoError(t, err)
	assert.Equal(t, 3, count)

	// Case 7: First with invalid model.
	_, err = db.Query[testPost](client).
		Where(db.EQ("id", 1)).
		First(ctx)
	assert.Error(t, err)

	// Case 8: First not found.
	_, err = db.Query[TestCategory](client).
		Where(db.EQ("id", 100)).
		First(ctx)
	assert.Error(t, err)

	// Case 9: First success.
	category, err := db.Query[TestCategory](client).
		Where(db.GTE("id", 2)).
		First(ctx)
	assert.NoError(t, err)
	assert.Equal(t, uint64(2), category.ID)

	// Case 10: Only invalid model.
	_, err = db.Query[testPost](client).
		Where(db.EQ("id", 1)).
		Only(ctx)
	assert.Error(t, err)

	// Case 11: Only not found.
	_, err = db.Query[TestCategory](client).
		Where(db.EQ("id", 100)).
		Only(ctx)
	assert.Error(t, err)

	// Case 12: Only query returns more than one result.
	_, err = db.Query[TestCategory](client).
		Where(db.GTE("id", 3)).
		Only(ctx)
	assert.Error(t, err)

	// Case 13: Only success.
	category, err = db.Query[TestCategory](client).
		Where(db.EQ("id", 2)).
		Only(ctx)
	assert.NoError(t, err)

	// Case 14: Query with schema.Entity: Invalid schema name
	_, err = db.Query[*schema.Entity](client, "invalid").Get(ctx)
	assert.ErrorContains(t, err, "model invalid not found")

	// Case 15: Query with schema.Entity: No schema name
	_, err = db.Query[*schema.Entity](client).Get(ctx)
	assert.ErrorContains(t, err, "schema name is required for type schema.Entity")

	// Case 16: Query with schema.Entity: Success
	cats, err := db.Query[*schema.Entity](client, "category").Get(ctx)
	assert.NoError(t, err)
	assert.Len(t, cats, 5)
	assert.Equal(t, "category 1", cats[0].Get("name"))
}

type TestData struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func TestBindStruct(t *testing.T) {
	// Case 1: Invalid target type.
	var intVar int
	assert.Error(t, db.BindStruct(nil, intVar))

	// Case 2: Invalid data type.
	assert.Error(t, db.BindStruct(1, &TestData{}))

	// Case 3: Bind success with map.
	data := map[string]interface{}{
		"id":   1,
		"name": "John Doe",
	}
	target := &TestData{}
	assert.NoError(t, db.BindStruct(data, &target))

	expected := &TestData{ID: 1, Name: "John Doe"}
	assert.Equal(t, expected, target)

	// Case 4: Bind success with schema.Entity.
	entityData := schema.NewEntity().Set("field1", "value1").Set("field2", "value2")
	var entityTarget map[string]interface{}
	assert.NoError(t, db.BindStruct(entityData, &entityTarget))

	expectedEntity := map[string]interface{}{
		"field1": "value1",
		"field2": "value2",
	}

	assert.Equal(t, expectedEntity, entityTarget)
}

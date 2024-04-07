package entdbadapter

import (
	"database/sql/driver"
	"testing"

	"github.com/fastschema/fastschema/app"
	"github.com/fastschema/fastschema/schema"
	"github.com/stretchr/testify/assert"
)

func TestMutation(t *testing.T) {
	mutation := &Mutation{
		ctx:    nil,
		skipTx: false,
		client: nil,
		model: &Model{
			name: "test",
		},
		predicates: []*app.Predicate{},
	}

	mutation.Where(app.LT("id", 1))

	assert.Equal(t, 1, len(mutation.predicates))
	assert.Equal(t, []*app.Predicate{app.LT("id", 1)}, mutation.predicates)

	_, err := mutation.GetRelationEntityIDs("test", 1)
	assert.Equal(t, "relation value for test.test is invalid", err.Error())

	entity := schema.NewEntity(1)
	relationEntityIDs, err := mutation.GetRelationEntityIDs("test", entity)
	assert.Nil(t, err)
	assert.Equal(t, []driver.Value{uint64(1)}, relationEntityIDs)

	relationEntityIDs, err = mutation.GetRelationEntityIDs("test", []*schema.Entity{
		schema.NewEntity(1),
		schema.NewEntity(2),
	})
	assert.Nil(t, err)
	assert.Equal(t, []driver.Value{uint64(1), uint64(2)}, relationEntityIDs)
}

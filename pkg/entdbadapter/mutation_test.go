package entdbadapter

import (
	"database/sql/driver"
	"testing"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/entity"
	"github.com/stretchr/testify/assert"
)

func TestMutation(t *testing.T) {
	mutation := &Mutation{
		client: nil,
		model: &Model{
			name: "test",
		},
		predicates: &[]*db.Predicate{},
	}

	mutation.Where(db.LT("id", 1))

	assert.Equal(t, 1, len(*mutation.predicates))
	assert.Equal(t, &[]*db.Predicate{db.LT("id", 1)}, mutation.predicates)

	_, err := mutation.GetRelationEntityIDs("test", 1)
	assert.Equal(t, "relation value for test.test is invalid", err.Error())

	e := entity.New(1)
	relationEntityIDs, err := mutation.GetRelationEntityIDs("test", e)
	assert.Nil(t, err)
	assert.Equal(t, []driver.Value{uint64(1)}, relationEntityIDs)

	relationEntityIDs, err = mutation.GetRelationEntityIDs("test", []*entity.Entity{
		entity.New(1),
		entity.New(2),
	})
	assert.Nil(t, err)
	assert.Equal(t, []driver.Value{uint64(1), uint64(2)}, relationEntityIDs)
}

func TestMutationGetRelationEntityIDsNil(t *testing.T) {
	var expected []driver.Value
	mutation := &Mutation{}
	value, err := mutation.GetRelationEntityIDs("test", nil)
	assert.NoError(t, err)
	assert.Equal(t, expected, value)

	mutation2 := &Mutation{}
	value2, err := mutation2.GetRelationEntityIDs("test", entity.New())
	assert.NoError(t, err)
	assert.Equal(t, []driver.Value{}, value2)

	mutation3 := &Mutation{
		model: &Model{
			name: "test",
		},
	}
	_, err = mutation3.GetRelationEntityIDs("test", entity.New(0))
	assert.Error(t, err)
}

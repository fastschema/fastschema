package entdbadapter

import (
	"database/sql/driver"
	"testing"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/entity"
	"github.com/fastschema/fastschema/schema"
	"github.com/stretchr/testify/assert"
)

func TestMutation(t *testing.T) {
	testSchema := &schema.Schema{
		Name: "test",
		Fields: []*schema.Field{
			{
				Name: "test",
				Type: schema.TypeRelation,
				Relation: &schema.Relation{
					Type:             schema.O2M,
					TargetSchemaName: "other",
					TargetFieldName:  "tests",
				},
			},
		},
	}

	mutation := &Mutation{
		client: nil,
		model: &Model{
			name:   "test",
			schema: testSchema,
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

func TestMutationGetRelationEntityIDsTargetColumn(t *testing.T) {
	relationSchema := &schema.Schema{
		Name: "source",
		Fields: []*schema.Field{
			{
				Name: "ref",
				Type: schema.TypeRelation,
				Relation: &schema.Relation{
					Type:             schema.O2M,
					TargetSchemaName: "target",
					TargetFieldName:  "sources",
					TargetColumn:     "legacy_id",
				},
			},
		},
	}

	mutation := &Mutation{
		model: &Model{
			name:   "source",
			schema: relationSchema,
		},
	}

	relationEntity := entity.New()
	relationEntity.Set("legacy_id", uint64(42))
	values, err := mutation.GetRelationEntityIDs("ref", relationEntity)
	assert.NoError(t, err)
	assert.Equal(t, []driver.Value{uint64(42)}, values)

	missingValue := entity.New(1)
	_, err = mutation.GetRelationEntityIDs("ref", missingValue)
	assert.EqualError(t, err, "relation entity for source.ref target column 'legacy_id' is invalid, value=0, err=cannot get uint64 value from entity: legacy_id")

	fullEntity := entity.New(99)
	fullEntity.Set("legacy_id", uint64(77))
	values, err = mutation.GetRelationEntityIDs("ref", fullEntity)
	assert.NoError(t, err)
	assert.Equal(t, []driver.Value{uint64(77)}, values)
}

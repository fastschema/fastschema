package schema

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRelation(t *testing.T) {
	relation := &Relation{
		Type:             O2M,
		TargetSchemaName: "user",
		TargetFieldName:  "id",
		Owner:            false,
		SchemaName:       "post",
		FieldName:        "owner_id",
		Optional:         false,
	}

	field := &Field{
		Name:     "owner",
		Type:     TypeRelation,
		Relation: relation,
	}

	schema := &Schema{
		Name:           "post",
		Namespace:      "posts",
		LabelFieldName: "name",
		Fields: []*Field{
			{
				Name: "name",
				Type: TypeString,
			},
			field,
		},
	}
	assert.NoError(t, schema.Init(false))

	relation.Init(schema, schema, field)
	assert.Equal(t, field.Optional, relation.Optional)
	assert.Equal(t, field.Name, relation.FieldName)
	assert.Equal(t, "post.owner-user.id", relation.Name)
	assert.Equal(t, "owner_id", relation.GetTargetFKColumn())
	// assert.Equal(t, map[string][]string{"owner": {"owner_id"}}, schema.RelationsFKColumns)
	assert.Equal(t, "user.id-post.owner", relation.GetBackRefName())
	assert.Equal(t, false, relation.IsSameType())
	assert.Equal(t, true, relation.HasFKs())
	assert.NoError(t, relation.CreateFKFields())

	assert.Equal(t, "relation node post.owner: 'user' is not found", NewRelationNodeError(schema, field).Error())
	assert.Equal(t, "backref relation for post.owner is not valid: 'user.id', please check the 'field' property in the 'user.id' relation definition", NewRelationBackRefError(relation).Error())
}

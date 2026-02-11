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
		SourceSchemaName: "post",
		SourceFieldName:  "owner_id",
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
	assert.Equal(t, field.Name, relation.SourceFieldName)
	assert.Equal(t, "post.owner-user.id", relation.Name)
	assert.Equal(t, "owner_id", relation.SourceColumn)
	assert.Equal(t, NoAction, relation.OnDelete)
	assert.Equal(t, NoAction, relation.OnDeleteOption())
	assert.Equal(t, NoAction, relation.OnUpdate)
	assert.Equal(t, NoAction, relation.OnUpdateOption())
	assert.Equal(t, "user.id-post.owner", relation.GetBackRefName())
	assert.Equal(t, false, relation.IsSameType())
	assert.Equal(t, true, relation.HasFKs())
	_, err := relation.CreateFKField()
	assert.NoError(t, err)

	assert.Equal(t, "relation node post.owner: 'user' is not found", NewRelationNodeError(schema, field).Error())
	assert.Equal(t, "backref relation for post.owner is not valid: 'user.id', please check the 'field' property in the 'user.id' relation definition", NewRelationBackRefError(relation).Error())
}

func TestRelationClone(t *testing.T) {
	var r *Relation
	assert.Nil(t, r.Clone())

	relation := &Relation{
		Type:             O2M,
		TargetSchemaName: "user",
		TargetFieldName:  "id",
		Owner:            false,
		SourceSchemaName: "post",
		SourceFieldName:  "owner_id",
		Optional:         false,
		OnDelete:         Cascade,
		OnUpdate:         Restrict,
	}

	clone := relation.Clone()

	assert.Equal(t, relation.TargetSchemaName, clone.TargetSchemaName)
	assert.Equal(t, relation.TargetFieldName, clone.TargetFieldName)
	assert.Equal(t, relation.Type, clone.Type)
	assert.Equal(t, relation.Owner, clone.Owner)
	assert.Equal(t, relation.Optional, clone.Optional)
	assert.Equal(t, relation.OnDelete, clone.OnDelete)
	assert.Equal(t, relation.OnUpdate, clone.OnUpdate)
}

func TestRelationOnDeleteDefaults(t *testing.T) {
	relation := &Relation{
		Type:             O2M,
		TargetSchemaName: "user",
		TargetFieldName:  "owner",
		Owner:            false,
	}

	field := &Field{
		Name:     "owner",
		Type:     TypeRelation,
		Optional: true,
		Relation: relation,
	}

	schema := &Schema{
		Name:           "pet",
		Namespace:      "pets",
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
	assert.Equal(t, SetNull, relation.OnDelete)
	assert.Equal(t, SetNull, relation.OnDeleteOption())

	// owner side should not report any option
	ownerRelation := &Relation{Type: O2M, Owner: true}
	assert.Equal(t, ReferenceOptionTypeInvalid, ownerRelation.OnDeleteOption())
}

func TestRelationOnUpdateDefaults(t *testing.T) {
	relation := &Relation{
		Type:             O2M,
		TargetSchemaName: "user",
		TargetFieldName:  "owner",
		Owner:            false,
	}

	field := &Field{
		Name:     "owner",
		Type:     TypeRelation,
		Relation: relation,
	}

	schema := &Schema{
		Name:           "pet",
		Namespace:      "pets",
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
	assert.Equal(t, NoAction, relation.OnUpdate)
	assert.Equal(t, NoAction, relation.OnUpdateOption())

	ownerRelation := &Relation{Type: O2M, Owner: true}
	assert.Equal(t, ReferenceOptionTypeInvalid, ownerRelation.OnUpdateOption())
}

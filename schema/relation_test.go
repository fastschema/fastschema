package schema

import (
	"testing"

	"github.com/fastschema/fastschema/entity"
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

	targetSchema := &Schema{
		Name:           "user",
		Namespace:      "users",
		LabelFieldName: "name",
		Fields: []*Field{
			{Name: entity.FieldID, Type: TypeUint64},
			{Name: "name", Type: TypeString},
		},
	}
	assert.NoError(t, targetSchema.Init(false))

	relation.Init(schema, targetSchema, field)
	assert.Equal(t, field.Optional, relation.Optional)
	assert.Equal(t, field.Name, relation.SourceFieldName)
	assert.Equal(t, "post.owner-user.id", relation.Name)
	assert.Equal(t, "owner_id", relation.SourceColumn)
	assert.Equal(t, entity.FieldID, relation.TargetColumn)
	assert.Equal(t, NoAction, relation.OnDelete)
	assert.Equal(t, NoAction, relation.OnDeleteOption())
	assert.Equal(t, NoAction, relation.OnUpdate)
	assert.Equal(t, NoAction, relation.OnUpdateOption())
	assert.Equal(t, "user.id-post.owner", relation.GetBackRefName())
	assert.Equal(t, false, relation.IsSameType())
	assert.Equal(t, true, relation.HasFKs())
	idField := schema.Field(entity.FieldID)
	assert.NotNil(t, idField)
	fkField, err := relation.CreateFKField(idField)
	assert.NoError(t, err)
	assert.Equal(t, relation.SourceColumn, fkField.Name)
	assert.Equal(t, idField.Type, fkField.Type)

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

func TestRelationTargetColumnFollowsPrimaryField(t *testing.T) {
	customer := &Schema{
		Name:             "customer",
		Namespace:        "customers",
		LabelFieldName:   "name",
		PrimaryFieldName: "slug",
		Fields: []*Field{
			{Name: "name", Type: TypeString},
			{Name: "slug", Type: TypeString},
		},
	}
	assert.NoError(t, customer.Init(false))

	order := &Schema{
		Name:           "order",
		Namespace:      "orders",
		LabelFieldName: "reference",
		Fields: []*Field{
			{Name: "reference", Type: TypeString},
		},
	}

	relation := &Relation{
		Type:             O2M,
		TargetSchemaName: customer.Name,
		TargetFieldName:  "orders",
		Owner:            false,
	}
	relationField := &Field{
		Name:     "customer",
		Type:     TypeRelation,
		Relation: relation,
	}
	order.Fields = append(order.Fields, relationField)
	assert.NoError(t, order.Init(false))

	relation.Init(order, customer, relationField)
	assert.Equal(t, "slug", relation.TargetColumn)
	assert.Equal(t, "customer_slug", relation.SourceColumn)

	// Ensure FK field matches target schema type
	fkField, err := relation.CreateFKField(customer.Field("slug"))
	assert.NoError(t, err)
	assert.Equal(t, TypeString, fkField.Type)
}

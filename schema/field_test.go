package schema

import (
	"fmt"
	"testing"

	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestField(t *testing.T) {
	field := &Field{}
	field.Init()
	assert.NotNil(t, field.DB)

	uint64Field := CreateUint64Field("id")
	assert.Equal(t, TypeUint64, uint64Field.Type)
	assert.Equal(t, "id", uint64Field.Name)
	assert.Equal(t, "id", uint64Field.Label)
	assert.Equal(t, false, uint64Field.Unique)
	assert.Equal(t, false, uint64Field.Optional)
	assert.Equal(t, "UNSIGNED", uint64Field.DB.Attr)
}

func TestFieldInitTypeMedia(t *testing.T) {
	schemaName := "testSchema"
	field := &Field{
		Type: TypeMedia,
		Name: "mediaField",
	}

	field.Init(schemaName)

	assert.NotNil(t, field.DB)
	assert.Equal(t, TypeMedia, field.Type)
	assert.Equal(t, false, field.DB.Increment)

	assert.NotNil(t, field.Relation)
	assert.Equal(t, utils.If(field.IsMultiple, M2M, O2M), field.Relation.Type)
	assert.Equal(t, false, field.Relation.Owner)
	assert.Equal(t, "media", field.Relation.TargetSchemaName)
	assert.Equal(t, fmt.Sprintf("%s_%s", schemaName, field.Name), field.Relation.TargetFieldName)
	assert.Nil(t, field.Relation.BackRef)
}

func TestValidValue(t *testing.T) {
	type args struct {
		fieldType FieldType
		value     any
		expect    bool
		enums     []*FieldEnum
	}

	tests := []args{
		{
			fieldType: TypeBool,
			value:     nil,
			expect:    true,
		},
		{
			fieldType: TypeBool,
			value:     false,
			expect:    true,
		},
		{
			fieldType: TypeBool,
			value:     "false",
			expect:    false,
		},
		{
			fieldType: TypeTime,
			value:     "2020-01-01 00:00:00",
			expect:    true,
		},
		{
			fieldType: TypeTime,
			value:     "2020-01-01",
			expect:    false,
		},
		{
			fieldType: TypeJSON,
			value:     `{"name":"test"}`,
			expect:    true,
		},
		{
			fieldType: TypeJSON,
			value:     1,
			expect:    false,
		},
		{
			fieldType: TypeJSON,
			value:     `{"name":"}`,
			expect:    false,
		},
		{
			fieldType: TypeUUID,
			value:     "123e4567-e89b-12d3-a456-426614174000",
			expect:    true,
		},
		{
			fieldType: TypeUUID,
			value:     1,
			expect:    false,
		},
		{
			fieldType: TypeBytes,
			value:     []byte("test"),
			expect:    true,
		},
		{
			fieldType: TypeBytes,
			value:     "test",
			expect:    false,
		},
		{
			fieldType: TypeEnum,
			value:     1,
			expect:    false,
		},
		{
			fieldType: TypeEnum,
			value:     1,
			enums: []*FieldEnum{
				{
					Value: "1",
					Label: "1",
				},
			},
			expect: false,
		},
		{
			fieldType: TypeEnum,
			value:     "1",
			enums: []*FieldEnum{
				{
					Value: "1",
					Label: "1",
				},
			},
			expect: true,
		},
		{
			fieldType: TypeInt,
			value:     "one",
			expect:    false,
		},
		{
			fieldType: TypeInt,
			value:     1,
			expect:    true,
		},
		{
			fieldType: TypeUint,
			value:     -1,
			expect:    false,
		},
		{
			fieldType: TypeUint,
			value:     1,
			expect:    true,
		},
		{
			fieldType: TypeFloat32,
			value:     "-1",
			expect:    false,
		},
		{
			fieldType: TypeFloat32,
			value:     1.5,
			expect:    true,
		},
		{
			fieldType: TypeInvalid,
			value:     1,
			expect:    false,
		},
		{
			fieldType: TypeInt,
			value:     []any{1, 2},
			expect:    true,
		},
		{
			fieldType: TypeInt,
			value:     []any{"a", "b"},
			expect:    false,
		},
	}

	for _, test := range tests {
		field := &Field{
			Type:  test.fieldType,
			Enums: test.enums,
		}
		assert.Equal(t, test.expect, field.IsValidValue(test.value))
	}
}
func TestFieldClone(t *testing.T) {
	field := &Field{
		Type:  TypeInt,
		Name:  "age",
		Label: "Age",
		Renderer: &FieldRenderer{
			Class: "number",
		},
		Size:          10,
		IsMultiple:    false,
		Unique:        true,
		Optional:      false,
		Default:       0,
		Sortable:      true,
		Filterable:    true,
		IsSystemField: false,
		Relation: &Relation{
			SchemaName:       "user",
			TargetSchemaName: "address",
			TargetFieldName:  "address_id",
		},
		DB: &FieldDB{
			Attr:      "UNSIGNED",
			Collation: "utf8mb4_unicode_ci",
			Increment: true,
			Key:       "PRI",
		},
		Enums: []*FieldEnum{
			{
				Value: "1",
				Label: "One",
			},
			{
				Value: "2",
				Label: "Two",
			},
		},
	}

	clonedField := field.Clone()

	// Check if the cloned field has the same values as the original field
	assert.Equal(t, field.Type, clonedField.Type)
	assert.Equal(t, field.Name, clonedField.Name)
	assert.Equal(t, field.Label, clonedField.Label)
	assert.Equal(t, field.Renderer, clonedField.Renderer)
	assert.Equal(t, field.Size, clonedField.Size)
	assert.Equal(t, field.IsMultiple, clonedField.IsMultiple)
	assert.Equal(t, field.Unique, clonedField.Unique)
	assert.Equal(t, field.Optional, clonedField.Optional)
	assert.Equal(t, field.Default, clonedField.Default)
	assert.Equal(t, field.Sortable, clonedField.Sortable)
	assert.Equal(t, field.Filterable, clonedField.Filterable)
	assert.Equal(t, field.IsSystemField, clonedField.IsSystemField)

	// Check if the cloned field's relation is a separate instance
	assert.Equal(t, field.Relation.Type, clonedField.Relation.Type)
	assert.Equal(t, field.Relation.TargetSchemaName, clonedField.Relation.TargetSchemaName)
	assert.Equal(t, field.Relation.TargetFieldName, clonedField.Relation.TargetFieldName)

	// Check if the cloned field's DB is a separate instance
	assert.Equal(t, field.DB.Attr, clonedField.DB.Attr)
	assert.Equal(t, field.DB.Collation, clonedField.DB.Collation)
	assert.Equal(t, field.DB.Increment, clonedField.DB.Increment)
	assert.Equal(t, field.DB.Key, clonedField.DB.Key)

	// Check if the cloned field's enums are separate instances
	assert.Len(t, clonedField.Enums, len(field.Enums))
	for i := range field.Enums {
		assert.Equal(t, field.Enums[i].Value, clonedField.Enums[i].Value)
		assert.Equal(t, field.Enums[i].Label, clonedField.Enums[i].Label)
	}
}

package schema

import (
	"testing"

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

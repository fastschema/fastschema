package schema

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestField(t *testing.T) {
	field := &Field{}
	assert.NoError(t, field.Init())
	assert.NotNil(t, field.DB)

	uint64Field := CreateUint64Field("id")
	assert.Equal(t, TypeUint64, uint64Field.Type)
	assert.Equal(t, "id", uint64Field.Name)
	assert.Equal(t, "id", uint64Field.Label)
	assert.Equal(t, false, uint64Field.Unique)
	assert.Equal(t, false, uint64Field.Optional)
	assert.Equal(t, "UNSIGNED", uint64Field.DB.Attr)
}

func TestFieldInitTypeFile(t *testing.T) {
	// File field without schema name
	assert.Error(t, (&Field{
		Type: TypeFile,
		Name: "fileField",
	}).Init())

	schemaName := "testSchema"
	field := &Field{
		Type: TypeFile,
		Name: "fileField",
	}

	assert.NoError(t, field.Init(schemaName))

	assert.NotNil(t, field.DB)
	assert.Equal(t, TypeFile, field.Type)
	assert.Equal(t, false, field.DB.Increment)

	assert.NotNil(t, field.Relation)
	assert.Equal(t, utils.If(field.IsMultiple, M2M, O2M), field.Relation.Type)
	assert.Equal(t, false, field.Relation.Owner)
	assert.Equal(t, "file", field.Relation.TargetSchemaName)
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
		Immutable:     true,
		Setter:        "5",
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
	assert.Equal(t, field.Immutable, clonedField.Immutable)
	assert.Equal(t, field.Setter, clonedField.Setter)
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

func TestErrInvalidFieldValue(t *testing.T) {
	fieldName := "testField"
	value := "invalidValue"
	err := errors.New("test error")

	expectedError := fmt.Sprintf("invalid field value: %s=%#v - %s", fieldName, value, err.Error())
	result := ErrInvalidFieldValue(fieldName, value, err)
	assert.EqualError(t, result, expectedError)

	// Test with no error input
	expectedError = fmt.Sprintf("invalid field value: %s=%#v", fieldName, value)
	result = ErrInvalidFieldValue(fieldName, value)
	assert.EqualError(t, result, expectedError)
}

func TestStringToFieldValue(t *testing.T) {
	type testInfo struct {
		Field  *Field
		Input  string
		Expect any
		Error  string
	}
	var createTest = func(fieldType FieldType, input string, expect any, errs ...string) *testInfo {
		ti := &testInfo{
			Field:  &Field{Name: fieldType.String(), Type: fieldType},
			Input:  input,
			Expect: expect,
		}

		if len(errs) > 0 {
			ti.Error = errs[0]
		}

		return ti
	}

	tests := []*testInfo{
		createTest(TypeBool, "true", true),
		createTest(TypeBool, "false", false),
		createTest(TypeBool, "invalid", nil, "invalid field value: bool"),
		createTest(TypeInt, "1", 1),
		createTest(TypeInt, "invalid", nil, "invalid field value: int"),
		createTest(TypeInt8, "1", int8(1)),
		createTest(TypeInt8, "invalid", nil, "invalid field value: int8"),
		createTest(TypeInt16, "1", int16(1)),
		createTest(TypeInt16, "invalid", nil, "invalid field value: int16"),
		createTest(TypeInt32, "1", int32(1)),
		createTest(TypeInt32, "invalid", nil, "invalid field value: int32"),
		createTest(TypeInt64, "1", int64(1)),
		createTest(TypeInt64, "invalid", nil, "invalid field value: int64"),
		createTest(TypeUint, "1", uint(1)),
		createTest(TypeUint, "invalid", nil, "invalid field value: uint"),
		createTest(TypeUint8, "1", uint8(1)),
		createTest(TypeUint8, "invalid", nil, "invalid field value: uint8"),
		createTest(TypeUint16, "1", uint16(1)),
		createTest(TypeUint16, "invalid", nil, "invalid field value: uint16"),
		createTest(TypeUint32, "1", uint32(1)),
		createTest(TypeUint32, "invalid", nil, "invalid field value: uint32"),
		createTest(TypeUint64, "1", uint64(1)),
		createTest(TypeUint64, "invalid", nil, "invalid field value: uint64"),
		createTest(TypeFloat32, "1.5", float32(1.5)),
		createTest(TypeFloat32, "invalid", nil, "invalid field value: float32"),
		createTest(TypeFloat64, "1.5", float64(1.5)),
		createTest(TypeFloat64, "invalid", nil, "invalid field value: float64"),
		createTest(TypeTime, "NOW()", "NOW()"),
		createTest(TypeTime, "2024-05-19T16:45:01Z", time.Date(2024, 5, 19, 16, 45, 1, 0, time.UTC)),
		createTest(TypeTime, "2024-05-19T16:45:01-07:00", time.Date(2024, 5, 19, 16, 45, 1, 0, time.FixedZone("", -7*60*60))),
		createTest(TypeTime, "invalid", nil, "invalid field value: time"),
		createTest(TypeString, `string`, "string"),
	}

	t.Run("StringToFieldValue", func(t *testing.T) {
		for _, test := range tests {
			t.Logf("Test: %s = %s", test.Field.Type, test.Input)
			value, err := StringToFieldValue[any](test.Field, test.Input)
			if test.Error != "" {
				assert.Nil(t, value)
				assert.Contains(t, err.Error(), test.Error)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.Expect, value)
			}
		}
	})

	// Test invalid generic type
	t.Run("StringToFieldValueInvalidType", func(t *testing.T) {
		_, err := StringToFieldValue[int](tests[0].Field, tests[0].Input)
		assert.Contains(t, err.Error(), "can't convert")
	})
}

func TestMergeFields(t *testing.T) {
	f1 := &Field{
		Type:          TypeInt,
		Name:          "age",
		Label:         "Age",
		IsMultiple:    false,
		Size:          10,
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

	f2 := &Field{
		Type:          TypeString,
		Name:          "name",
		Label:         "Name",
		IsMultiple:    true,
		Size:          20,
		Unique:        false,
		Optional:      true,
		Default:       "John",
		Sortable:      false,
		Filterable:    false,
		IsSystemField: true,
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
				Value: "3",
				Label: "Three",
			},
			{
				Value: "4",
				Label: "Four",
			},
		},
		Renderer: &FieldRenderer{
			Class: "text",
			Settings: map[string]any{
				"rows": 5,
			},
		},
	}

	expectedField := &Field{
		Type:          TypeString,
		Name:          "name",
		Label:         "Name",
		IsMultiple:    true,
		Size:          20,
		Unique:        false,
		Optional:      true,
		Default:       "John",
		Sortable:      false,
		Filterable:    false,
		IsSystemField: true,
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
				Value: "3",
				Label: "Three",
			},
			{
				Value: "4",
				Label: "Four",
			},
		},
		Renderer: &FieldRenderer{
			Class: "text",
			Settings: map[string]any{
				"rows": 5,
			},
		},
	}

	MergeFields(f1, f2)

	assert.Equal(t, expectedField.Type, f1.Type)
	assert.Equal(t, expectedField.Name, f1.Name)
	assert.Equal(t, expectedField.Label, f1.Label)
	assert.Equal(t, expectedField.IsMultiple, f1.IsMultiple)
	assert.Equal(t, expectedField.Size, f1.Size)
	assert.Equal(t, expectedField.Unique, f1.Unique)
	assert.Equal(t, expectedField.Optional, f1.Optional)
	assert.Equal(t, expectedField.Default, f1.Default)
	assert.Equal(t, expectedField.Sortable, f1.Sortable)
	assert.Equal(t, expectedField.Filterable, f1.Filterable)
	assert.Equal(t, expectedField.IsSystemField, f1.IsSystemField)
	assert.Equal(t, expectedField.Renderer, f1.Renderer)

	assert.Equal(t, expectedField.Relation.Type, f1.Relation.Type)
	assert.Equal(t, expectedField.Relation.TargetSchemaName, f1.Relation.TargetSchemaName)
	assert.Equal(t, expectedField.Relation.TargetFieldName, f1.Relation.TargetFieldName)

	assert.Equal(t, expectedField.DB.Attr, f1.DB.Attr)
	assert.Equal(t, expectedField.DB.Collation, f1.DB.Collation)
	assert.Equal(t, expectedField.DB.Increment, f1.DB.Increment)
	assert.Equal(t, expectedField.DB.Key, f1.DB.Key)

	assert.Len(t, f1.Enums, len(expectedField.Enums))
	for i := range expectedField.Enums {
		assert.Equal(t, expectedField.Enums[i].Value, f1.Enums[i].Value)
		assert.Equal(t, expectedField.Enums[i].Label, f1.Enums[i].Label)
	}
}

func TestFieldGetterSetter(t *testing.T) {
	createField := func() *Field {
		return &Field{
			Type:  TypeInt,
			Name:  "age",
			Label: "Age",
		}
	}

	// Invalid getter
	{
		field := createField()
		field.Getter = "invalid"
		assert.Error(t, field.Init())
	}

	// Invalid setter
	{
		field := createField()
		field.Setter = "invalid"
		assert.Error(t, field.Init())
	}

	// Getter and setter
	{
		field := createField()
		field.Setter = "5"
		field.Getter = "$args.Value * 5"
		assert.NoError(t, field.Init())
	}
}

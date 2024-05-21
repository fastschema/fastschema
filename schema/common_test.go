package schema

import (
	"reflect"
	"testing"

	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestCommon(t *testing.T) {
	assert.Equal(t, "id", FieldID)

	assert.Equal(t, "bool", TypeBool.String())
	assert.Equal(t, "invalid", endFieldTypes.String())

	assert.Equal(t, []byte(`"bool"`), utils.Must(TypeBool.MarshalJSON()))

	var fieldType FieldType
	assert.NoError(t, fieldType.UnmarshalJSON([]byte(`"bool"`)))
	assert.Equal(t, TypeBool, fieldType)

	assert.Error(t, fieldType.UnmarshalJSON([]byte(`"bool`)))
	assert.Error(t, fieldType.UnmarshalJSON([]byte(`"invalidFieldType"`)))
	assert.Equal(t, TypeInvalid, fieldType)

	var relationType RelationType
	assert.NoError(t, relationType.UnmarshalJSON([]byte(`"o2o"`)))
	assert.Equal(t, true, relationType.IsO2O())
	assert.NoError(t, relationType.UnmarshalJSON([]byte(`"o2m"`)))
	assert.Equal(t, true, relationType.IsO2M())
	assert.NoError(t, relationType.UnmarshalJSON([]byte(`"m2m"`)))
	assert.Equal(t, true, relationType.IsM2M())
	assert.Equal(t, "m2m", relationType.String())
	assert.Equal(t, true, relationType.Valid())
	assert.Equal(t, "invalid", endRelationTypes.String())

	assert.Equal(t, []byte(`"m2m"`), utils.Must(M2M.MarshalJSON()))
	assert.NoError(t, relationType.UnmarshalJSON([]byte(`"m2m"`)))
	assert.Equal(t, M2M, relationType)
	assert.Error(t, relationType.UnmarshalJSON([]byte(`"m2m`)))
}

func TestFieldEnumClone(t *testing.T) {
	// Create a sample FieldEnum instance
	field := &FieldEnum{
		Value: "value",
		Label: "label",
	}

	// Clone the FieldEnum instance
	clone := field.Clone()

	// Verify that the cloned instance has the same values as the original
	assert.Equal(t, field.Value, clone.Value)
	assert.Equal(t, field.Label, clone.Label)
}

func TestFieldDBClone(t *testing.T) {
	var f *FieldDB
	assert.Nil(t, f.Clone())

	// Create a sample FieldDB instance
	field := &FieldDB{
		Attr:      "attr",
		Collation: "collation",
		Increment: true,
		Key:       "key",
	}

	// Clone the FieldDB instance
	clone := field.Clone()

	// Verify that the cloned instance has the same values as the original
	assert.Equal(t, field.Attr, clone.Attr)
	assert.Equal(t, field.Collation, clone.Collation)
	assert.Equal(t, field.Increment, clone.Increment)
	assert.Equal(t, field.Key, clone.Key)
}

func TestFieldTypeStructType(t *testing.T) {
	tests := []struct {
		name      string
		fieldType FieldType
		expected  reflect.Type
	}{
		{
			name:      "String FieldType",
			fieldType: TypeString,
			expected:  reflect.TypeOf(""),
		},
		{
			name:      "Bool FieldType",
			fieldType: TypeBool,
			expected:  reflect.TypeOf(true),
		},
		{
			name:      "Invalid FieldType",
			fieldType: endFieldTypes + 1,
			expected:  reflect.TypeOf(nil),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.fieldType.StructType()
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestFieldRendererClone(t *testing.T) {
	// Case 1: nil FieldRenderer
	var fr *FieldRenderer
	assert.Nil(t, fr.Clone())

	// Case 2: FieldRenderer with settings
	field := &FieldRenderer{
		Class: "class",
		Settings: map[string]any{
			"key": "value",
		},
	}
	assert.NotNil(t, field.Clone())
}

func TestFieldTypeFromReflectType(t *testing.T) {
	// Case 1: Nil type
	ft := FieldTypeFromReflectType(nil)
	assert.Equal(t, TypeInvalid, ft)

	// Case 2: Invalid type
	ft = FieldTypeFromReflectType(reflect.TypeOf(nil))
	assert.Equal(t, TypeInvalid, ft)

	// Case 3: Valid type
	ft = FieldTypeFromReflectType(reflect.TypeOf(""))
	assert.Equal(t, TypeString, ft)
}

func TestFieldTypeFromString(t *testing.T) {
	// Case 1: Invalid string
	ft := FieldTypeFromString("invalidtype")
	assert.Equal(t, TypeInvalid, ft)

	// Case 2: Valid string
	ft = FieldTypeFromString("string")
	assert.Equal(t, TypeString, ft)
}

func TestFieldTypeIsAtomic(t *testing.T) {
	tests := []struct {
		name      string
		fieldType FieldType
		expected  bool
	}{
		{
			name:      "Time",
			fieldType: TypeTime,
			expected:  true,
		},
		{
			name:      "Atomic type",
			fieldType: TypeBool,
			expected:  true,
		},
		{
			name:      "Non-atomic type",
			fieldType: TypeRelation,
			expected:  false,
		},
		{
			name:      "Invalid type",
			fieldType: TypeInvalid,
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.fieldType.IsAtomic()
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestRelationTypeFromString(t *testing.T) {
	// Case 1: Invalid string
	rt := RelationTypeFromString("invalidtype")
	assert.Equal(t, RelationInvalid, rt)

	// Case 2: Valid string
	rt = RelationTypeFromString("o2o")
	assert.Equal(t, O2O, rt)
}

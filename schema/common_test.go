package schema

import (
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

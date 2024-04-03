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

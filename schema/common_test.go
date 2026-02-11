package schema

import (
	"reflect"
	"testing"

	"github.com/fastschema/fastschema/entity"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestCommon(t *testing.T) {
	assert.Equal(t, "id", entity.FieldID)

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
		Key:       DBUniqueKey,
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

func TestReferenceOptionType(t *testing.T) {
	tests := []struct {
		name    string
		opt     ReferenceOptionType
		str     string
		valid   bool
		jsonVal string
	}{
		{"NoAction", NoAction, "NO ACTION", true, `"NO ACTION"`},
		{"Restrict", Restrict, "RESTRICT", true, `"RESTRICT"`},
		{"Cascade", Cascade, "CASCADE", true, `"CASCADE"`},
		{"SetNull", SetNull, "SET NULL", true, `"SET NULL"`},
		{"SetDefault", SetDefault, "SET DEFAULT", true, `"SET DEFAULT"`},
		{"Invalid", ReferenceOptionTypeInvalid, "INVALID", false, `"INVALID"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test String()
			assert.Equal(t, tt.str, tt.opt.String())

			// Test Valid()
			assert.Equal(t, tt.valid, tt.opt.Valid())

			// Test MarshalJSON()
			jsonBytes, err := tt.opt.MarshalJSON()
			assert.NoError(t, err)
			assert.Equal(t, tt.jsonVal, string(jsonBytes))
		})
	}

	// Test String() with out-of-bounds value
	assert.Equal(t, "INVALID", endReferenceOptionTypes.String())

	// Test ReferenceOptionTypeFromString
	assert.Equal(t, NoAction, ReferenceOptionTypeFromString("NO ACTION"))
	assert.Equal(t, Restrict, ReferenceOptionTypeFromString("RESTRICT"))
	assert.Equal(t, Cascade, ReferenceOptionTypeFromString("CASCADE"))
	assert.Equal(t, SetNull, ReferenceOptionTypeFromString("SET NULL"))
	assert.Equal(t, SetDefault, ReferenceOptionTypeFromString("SET DEFAULT"))
	assert.Equal(t, ReferenceOptionTypeInvalid, ReferenceOptionTypeFromString("invalid"))
	assert.Equal(t, ReferenceOptionTypeInvalid, ReferenceOptionTypeFromString(""))

	// Test UnmarshalJSON
	var opt ReferenceOptionType
	assert.NoError(t, opt.UnmarshalJSON([]byte(`"CASCADE"`)))
	assert.Equal(t, Cascade, opt)

	// Invalid JSON
	assert.Error(t, opt.UnmarshalJSON([]byte(`"CASCADE`)))

	// Unknown option (sets to zero value)
	assert.NoError(t, opt.UnmarshalJSON([]byte(`"UNKNOWN"`)))
	assert.Equal(t, ReferenceOptionTypeInvalid, opt)
}

func TestDBKeyType(t *testing.T) {
	tests := []struct {
		name    string
		key     DBKeyType
		str     string
		valid   bool
		jsonVal string
	}{
		{"Empty", DBEmptyKey, "", true, `""`},
		{"Primary", DBPrimaryKey, "PRI", true, `"PRI"`},
		{"Unique", DBUniqueKey, "UNI", true, `"UNI"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test String()
			assert.Equal(t, tt.str, tt.key.String())

			// Test Valid()
			assert.Equal(t, tt.valid, tt.key.Valid())

			// Test MarshalJSON()
			jsonBytes, err := tt.key.MarshalJSON()
			assert.NoError(t, err)
			assert.Equal(t, tt.jsonVal, string(jsonBytes))
		})
	}

	// Test String() with out-of-bounds value
	assert.Equal(t, "", endDBKeyTypes.String())

	// Test DBKeyTypeFromString
	assert.Equal(t, DBEmptyKey, DBKeyTypeFromString(""))
	assert.Equal(t, DBPrimaryKey, DBKeyTypeFromString("PRI"))
	assert.Equal(t, DBUniqueKey, DBKeyTypeFromString("UNI"))
	assert.Equal(t, DBEmptyKey, DBKeyTypeFromString("invalid"))
	assert.Equal(t, DBEmptyKey, DBKeyTypeFromString("UNKNOWN"))

	// Test Valid() boundary - out of range should be false
	outOfRange := DBKeyType(100)
	assert.False(t, outOfRange.Valid())

	// Test UnmarshalJSON
	var key DBKeyType
	assert.NoError(t, key.UnmarshalJSON([]byte(`"PRI"`)))
	assert.Equal(t, DBPrimaryKey, key)

	assert.NoError(t, key.UnmarshalJSON([]byte(`"UNI"`)))
	assert.Equal(t, DBUniqueKey, key)

	assert.NoError(t, key.UnmarshalJSON([]byte(`""`)))
	assert.Equal(t, DBEmptyKey, key)

	// Invalid JSON
	assert.Error(t, key.UnmarshalJSON([]byte(`"PRI`)))

	// Unknown key (sets to zero value which is DBEmptyKey)
	assert.NoError(t, key.UnmarshalJSON([]byte(`"UNKNOWN"`)))
	assert.Equal(t, DBEmptyKey, key)
}

func TestFieldTypeIsInteger(t *testing.T) {
	integerTypes := []FieldType{
		TypeInt, TypeInt8, TypeInt16, TypeInt32, TypeInt64,
		TypeUint, TypeUint8, TypeUint16, TypeUint32, TypeUint64,
	}
	for _, ft := range integerTypes {
		assert.True(t, ft.IsInteger(), "expected %s to be integer", ft)
	}

	nonIntegerTypes := []FieldType{
		TypeBool, TypeString, TypeText, TypeTime, TypeJSON,
		TypeUUID, TypeBytes, TypeEnum, TypeFloat32, TypeFloat64,
		TypeRelation, TypeFile, TypeInvalid,
	}
	for _, ft := range nonIntegerTypes {
		assert.False(t, ft.IsInteger(), "expected %s to not be integer", ft)
	}
}

func TestFieldTypeIsUnsignedInteger(t *testing.T) {
	unsignedTypes := []FieldType{
		TypeUint, TypeUint8, TypeUint16, TypeUint32, TypeUint64,
	}
	for _, ft := range unsignedTypes {
		assert.True(t, ft.IsUnsignedInteger(), "expected %s to be unsigned integer", ft)
	}

	signedOrOtherTypes := []FieldType{
		TypeInt, TypeInt8, TypeInt16, TypeInt32, TypeInt64,
		TypeBool, TypeString, TypeText, TypeTime, TypeJSON,
		TypeUUID, TypeBytes, TypeEnum, TypeFloat32, TypeFloat64,
		TypeRelation, TypeFile, TypeInvalid,
	}
	for _, ft := range signedOrOtherTypes {
		assert.False(t, ft.IsUnsignedInteger(), "expected %s to not be unsigned integer", ft)
	}
}

func TestFieldTypeIsRelationType(t *testing.T) {
	assert.True(t, TypeRelation.IsRelationType())
	assert.True(t, TypeFile.IsRelationType())
	assert.False(t, TypeString.IsRelationType())
	assert.False(t, TypeInt.IsRelationType())
}

func TestFieldTypeValid(t *testing.T) {
	// Valid types
	validTypes := []FieldType{
		TypeBool, TypeTime, TypeJSON, TypeUUID, TypeBytes, TypeEnum,
		TypeString, TypeText, TypeInt, TypeInt8, TypeInt16, TypeInt32,
		TypeInt64, TypeUint, TypeUint8, TypeUint16, TypeUint32, TypeUint64,
		TypeFloat32, TypeFloat64, TypeRelation, TypeFile,
	}
	for _, ft := range validTypes {
		assert.True(t, ft.Valid(), "expected %s to be valid", ft)
	}

	// Invalid types
	assert.False(t, TypeInvalid.Valid())
	assert.False(t, endFieldTypes.Valid())
	assert.False(t, FieldType(-1).Valid())
}

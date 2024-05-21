package openapi_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/openapi"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
	"github.com/ogen-go/ogen"
	"github.com/stretchr/testify/assert"
)

type testStruct struct {
	Name string
}

func TestGetElemSchema(t *testing.T) {
	oas := &openapi.OpenAPISpec{}
	rt := reflect.TypeOf([]string{})
	elemSchema := oas.GetElemSchema(rt)
	expectedSchema := ogen.NewSchema()
	expectedSchema.Type = "string"
	assert.Equal(t, expectedSchema, elemSchema)
}

func TestCreateSliceSchema(t *testing.T) {
	oas := utils.Must(openapi.NewSpec(&openapi.OpenAPISpecConfig{
		Resources: fs.NewResourcesManager(),
	}))

	// Case 1: slice of string
	dtType := reflect.TypeOf([]string{})
	sliceSchema := oas.CreateSliceSchema(dtType)
	expectedSchema := ogen.NewSchema().SetType("string").AsArray()
	assert.Equal(t, expectedSchema, sliceSchema)

	// Case 2: slice of any type
	dtType = reflect.TypeOf([]any{})
	sliceSchema = oas.CreateSliceSchema(dtType)
	expectedSchema = ogen.NewSchema().AsArray()
	assert.Equal(t, expectedSchema, sliceSchema)

	// Case 3: slice of struct
	dtType = reflect.TypeOf([]testStruct{})
	sliceSchema = oas.CreateSliceSchema(dtType)
	expectedSchema = openapi.RefSchema("testStruct").AsArray()
	assert.Equal(t, expectedSchema, sliceSchema)

	// Case 4: array of struct
	twoItems := uint64(2)
	dtType = reflect.TypeOf([2]testStruct{})
	sliceSchema = oas.CreateSliceSchema(dtType)
	expectedSchema = openapi.RefSchema("testStruct").AsArray()
	expectedSchema.MaxItems = &twoItems
	assert.Equal(t, expectedSchema, sliceSchema)
}

func TestCreateMapSchema(t *testing.T) {
	// Case 1: map of any type
	oas := utils.Must(openapi.NewSpec(&openapi.OpenAPISpecConfig{
		Resources: fs.NewResourcesManager(),
	}))
	dtType := reflect.TypeOf(map[string]any{})
	mapSchema := oas.CreateMapSchema(dtType)
	expectedSchema := ogen.NewSchema()
	expectedSchema.SetType("object")
	expectedSchema.AdditionalProperties = &ogen.AdditionalProperties{
		Schema: *ogen.NewSchema(),
	}
	assert.Equal(t, expectedSchema, mapSchema)

	// Case 3: map of struct
	dtType = reflect.TypeOf(map[string]testStruct{})
	mapSchema = oas.CreateMapSchema(dtType)
	expectedSchema = ogen.NewSchema()
	expectedSchema.SetType("object")
	expectedSchema.AdditionalProperties = &ogen.AdditionalProperties{
		Schema: *openapi.RefSchema("testStruct"),
	}
	assert.Equal(t, expectedSchema, mapSchema)
}

func TestCreateStructSchema(t *testing.T) {
	// Case 1: struct with no name and fields
	oas := utils.Must(openapi.NewSpec(&openapi.OpenAPISpecConfig{
		Resources: fs.NewResourcesManager(),
	}))
	dtType := reflect.TypeOf(struct{}{})
	structSchema := oas.CreateStructSchema(dtType)

	expectedSchema := openapi.RefSchema("Schema001")
	assert.Equal(t, expectedSchema, structSchema)

	// Case 2: schema.Entity
	dtType = reflect.TypeOf(schema.Entity{})
	structSchema = oas.CreateStructSchema(dtType)
	expectedSchema = openapi.RefSchema("Schema.Entity")
	assert.Equal(t, expectedSchema, structSchema)
}

func TestTypeToOgenSchema(t *testing.T) {
	oas := utils.Must(openapi.NewSpec(&openapi.OpenAPISpecConfig{
		Resources: fs.NewResourcesManager(),
	}))

	// Test case 1: Type is already an ogen.Schema object
	schema := ogen.NewSchema()
	result := oas.TypeToOgenSchema(schema)
	assert.Equal(t, schema, result)

	// Test case 2: Type is not dereferenceable
	result = oas.TypeToOgenSchema(nil)
	expectedSchema := ogen.NewSchema()
	assert.Equal(t, expectedSchema, result)

	// Test case 3: Primitive types
	result = oas.TypeToOgenSchema(int(0))
	expectedSchema = ogen.NewSchema().SetType("integer")
	assert.Equal(t, expectedSchema, result)

	// Test case 4: time.Time type
	result = oas.TypeToOgenSchema(time.Time{})
	expectedSchema = ogen.DateTime()
	assert.Equal(t, expectedSchema, result)

	// Test case 5: []uint8 type
	result = oas.TypeToOgenSchema([]uint8{})
	expectedSchema = ogen.Bytes()
	assert.Equal(t, expectedSchema, result)

	// Test case 6: Slice type
	result = oas.TypeToOgenSchema([]string{})
	expectedSchema = oas.CreateSliceSchema(reflect.TypeOf([]string{}))
	assert.Equal(t, expectedSchema, result)

	// Test case 7: Map type
	result = oas.TypeToOgenSchema(map[string]interface{}{})
	expectedSchema = oas.CreateMapSchema(reflect.TypeOf(map[string]interface{}{}))
	assert.Equal(t, expectedSchema, result)

	// Test case 8: Struct type
	result = oas.TypeToOgenSchema(testStruct{})
	expectedSchema = oas.CreateStructSchema(reflect.TypeOf(testStruct{}))
	assert.Equal(t, expectedSchema, result)
}

func TestResolveSchemaReferences(t *testing.T) {
	oas := utils.Must(openapi.NewSpec(&openapi.OpenAPISpecConfig{
		Resources: fs.NewResourcesManager(),
	}))

	oas.ResolveSchemaReferences()

	type testStruct2 struct {
		Name string
		Test struct {
			private     string
			NotExported string `json:"-"`
			Meta        any
			NestField   string
			NestedStrct testStruct
			Updated     time.Time
		}
	}

	// Test case 1: struct with nested struct
	dtType := reflect.TypeOf(testStruct2{})
	structSchema := oas.CreateStructSchema(dtType)
	assert.NotNil(t, structSchema)
	oas.ResolveSchemaReferences()

	testStructSchema := openapi.RefSchema("testStruct")
	assert.NotNil(t, testStructSchema)

	struct2Test := oas.Schema("testStruct2.Test")
	assert.NotNil(t, struct2Test)

	expectedStruct2Test := ogen.NewSchema()
	expectedStruct2Test.SetType("object")
	expectedStruct2Test.Properties = []ogen.Property{
		{
			Name:   "Meta",
			Schema: ogen.NewSchema(),
		},
		{
			Name:   "NestField",
			Schema: ogen.String(),
		},
		{
			Name:   "NestedStrct",
			Schema: testStructSchema,
		},
		{
			Name:   "Updated",
			Schema: ogen.DateTime(),
		},
	}

	assert.Equal(t, expectedStruct2Test, struct2Test)
}

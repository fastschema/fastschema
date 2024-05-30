package openapi_test

import (
	"reflect"
	"testing"

	"github.com/fastschema/fastschema/pkg/openapi"
	"github.com/ogen-go/ogen"
	"github.com/stretchr/testify/assert"
)

func TestPrimitiveToOgenTypeMaps(t *testing.T) {
	expectedMaps := map[reflect.Kind]func() *ogen.Schema{
		reflect.Bool:  ogen.Bool,
		reflect.Int:   ogen.Int,
		reflect.Int8:  ogen.Int32,
		reflect.Int16: ogen.Int32,
		reflect.Int32: ogen.Int32,
		reflect.Int64: ogen.Int64,
		reflect.Uint: func() *ogen.Schema {
			return ogen.Int().SetMinimum(&openapi.UintMinValue)
		},
		reflect.Uint8: func() *ogen.Schema {
			return ogen.Int().SetMinimum(&openapi.UintMinValue)
		},
		reflect.Uint16: func() *ogen.Schema {
			return ogen.Int().SetMinimum(&openapi.UintMinValue)
		},
		reflect.Uint32: func() *ogen.Schema {
			return ogen.Int32().SetMinimum(&openapi.UintMinValue)
		},
		reflect.Uint64: func() *ogen.Schema {
			return ogen.Int64().SetMinimum(&openapi.UintMinValue)
		},
		reflect.Uintptr: func() *ogen.Schema {
			return ogen.Int64().SetMinimum(&openapi.UintMinValue)
		},
		reflect.Float32:    ogen.Float,
		reflect.Float64:    ogen.Double,
		reflect.Complex64:  ogen.Float,
		reflect.Complex128: ogen.Float,
		reflect.String:     ogen.String,
	}

	for kind, expectedFunc := range expectedMaps {
		actualFunc := openapi.PrimitiveToOgenTypeMaps[kind]
		assert.Equal(t, expectedFunc(), actualFunc())
	}
}

func TestCreateInfoObject(t *testing.T) {
	info := openapi.CreateInfoObject()

	assert.Equal(t, "FastSchema", info.Title)
	assert.Equal(t, "FastSchema OAS3", info.Description)
	assert.Equal(t, "https://fastschema.com/terms", info.TermsOfService)

	assert.NotNil(t, info.Contact)
	assert.Equal(t, "FastSchema Team", info.Contact.Name)
	assert.Equal(t, "https://fastschema.com", info.Contact.URL)
	assert.Equal(t, "contact@fastschema.com", info.Contact.Email)

	assert.NotNil(t, info.License)
	assert.Equal(t, "MIT", info.License.Name)
	assert.Equal(t, "https://opensource.org/licenses/MIT", info.License.URL)

	assert.Equal(t, "0.0.1", info.Version)
}

func TestCreateParamsObject(t *testing.T) {
	params := openapi.CreateParamsObject()

	assert.NotNil(t, params["authBearerHeader"])
	assert.Equal(t, "Authorization", params["authBearerHeader"].Name)
	assert.Equal(t, "header", params["authBearerHeader"].In)
	assert.Equal(t, "Authorization bearer token", params["authBearerHeader"].Description)
	assert.Equal(t, "string", params["authBearerHeader"].Schema.Type)

	example := params["authBearerHeader"].Examples["example"]
	assert.NotNil(t, example)
	assert.Equal(t, "Bearer token example", example.Summary)
	assert.Equal(t, ogen.ExampleValue([]byte(`"Bearer token"`)), example.Value)
}

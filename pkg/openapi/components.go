package openapi

import (
	"reflect"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/gen/ir"
)

const SchemaIDOnlyName = "Schema.IDOnly"
const SchemaCreate = "Create"

var IDOnlySchema = ogen.NewSchema().AddRequiredProperties(
	PrimitiveToOgenTypeMaps[reflect.Uint64]().ToProperty("id"),
)

var UintMinValue int64 = 0

// PrimitiveToOgenTypeMaps is a map that maps Go primitive types to their corresponding Ogen schema types.
var PrimitiveToOgenTypeMaps = map[reflect.Kind]func() *ogen.Schema{
	reflect.Bool:  ogen.Bool,
	reflect.Int:   ogen.Int,
	reflect.Int8:  ogen.Int32,
	reflect.Int16: ogen.Int32,
	reflect.Int32: ogen.Int32,
	reflect.Int64: ogen.Int64,
	reflect.Uint: func() *ogen.Schema {
		return ogen.Int().SetMinimum(&UintMinValue)
	},
	reflect.Uint8: func() *ogen.Schema {
		return ogen.Int().SetMinimum(&UintMinValue)
	},
	reflect.Uint16: func() *ogen.Schema {
		return ogen.Int().SetMinimum(&UintMinValue)
	},
	reflect.Uint32: func() *ogen.Schema {
		return ogen.Int32().SetMinimum(&UintMinValue)
	},
	reflect.Uint64: func() *ogen.Schema {
		return ogen.Int64().SetMinimum(&UintMinValue)
	},
	reflect.Uintptr: func() *ogen.Schema {
		return ogen.Int64().SetMinimum(&UintMinValue)
	},
	reflect.Float32:    ogen.Float,
	reflect.Float64:    ogen.Double,
	reflect.Complex64:  ogen.Float,
	reflect.Complex128: ogen.Float,
	reflect.String:     ogen.String,
}

var unAuthorizedSchema = &ogen.Schema{
	Type:        "object",
	Description: "Unauthorized Error Response",
	Properties: []ogen.Property{
		{
			Name: "code",
			Schema: &ogen.Schema{
				Type:    "integer",
				Format:  "int32",
				Default: []byte("401"),
				// Example: []byte("401"),
			},
		},
		{
			Name: "message",
			Schema: &ogen.Schema{
				Type:    "string",
				Default: []byte(`"Unauthorized"`),
				// Example: []byte(`"Unauthorized"`),
			},
		},
	},
}
var unAuthorizedResponse = &ogen.Response{
	Description: "Unauthorized",
	Content: map[string]ogen.Media{
		ir.EncodingJSON.String(): {
			Schema: ogen.NewSchema().AddRequiredProperties(unAuthorizedSchema.ToProperty("error")),
			Examples: map[string]*ogen.Example{
				"example": {
					Summary: "Unauthorized example",
					Value:   []byte(`{"error": {"code": 401, "message": "Unauthorized"}}`),
				},
			},
		},
	},
}

func CreateInfoObject() ogen.Info {
	return ogen.Info{
		Title:          "FastSchema",
		Description:    "FastSchema OAS3",
		TermsOfService: "https://fastschema.com/terms",
		Contact: &ogen.Contact{
			Name:  "FastSchema Team",
			URL:   "https://fastschema.com",
			Email: "contact@fastschema.com",
		},
		License: &ogen.License{
			Name: "MIT",
			URL:  "https://opensource.org/licenses/MIT",
		},
		Version: "0.0.1",
	}
}

func CreateParamsObject() map[string]*ogen.Parameter {
	return map[string]*ogen.Parameter{
		// Auth bearer header parameter
		"authBearerHeader": {
			Name:        "Authorization",
			In:          "header",
			Description: "Authorization bearer token",
			Schema:      &ogen.Schema{Type: "string"},
			// Example:     ogen.ExampleValue([]byte(`"Bearer token"`)),
			Examples: map[string]*ogen.Example{
				"example": {
					Summary: "Bearer token example",
					Value:   ogen.ExampleValue([]byte(`"Bearer token"`)),
				},
			},
		},
	}
}

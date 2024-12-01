package openapi

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/ogen-go/ogen"
)

type TypeToOgenConfig struct {
	skipStruct bool
	structName string
}

// TypeToOgenSchema converts a given type to an ogen.Schema object based on the OpenAPI specification.
//
//	It accepts a type `t` and optional configuration parameters `configs`.
//	If the type `t` is already an ogen.Schema object, it is returned as is.
//	If the type `t` is not dereferenceable, an empty ogen.Schema object is returned.
//	For primitive types, the corresponding ogen.Schema object is created and returned.
//	For the `time.Time` type, an ogen.Schema object representing a DateTime is returned.
//	For the `[]uint8` type, an ogen.Schema object representing Bytes is returned.
//	For slice or array types, the corresponding ogen.Schema object is created and returned.
//	For map types, the corresponding ogen.Schema object is created and returned.
//	For struct types, the corresponding ogen.Schema object is created and returned, unless the `skipStruct` configuration is set to true.
//	If none of the above conditions are met, nil is returned.
func (oas *OpenAPISpec) TypeToOgenSchema(t any, configs ...*TypeToOgenConfig) (structSchema *ogen.Schema) {
	ogenSchema, ok := t.(*ogen.Schema)
	if ok {
		return ogenSchema
	}

	if !utils.Dereferenceable(t) {
		return ogen.NewSchema()
	}

	configs = append(configs, &TypeToOgenConfig{})
	dtType := utils.GetDereferencedType(t)
	dtKind := dtType.Kind()

	// Create schema for primitive types
	primitiveSchema, ok := PrimitiveToOgenTypeMaps[dtKind]
	if ok {
		return primitiveSchema()
	}

	// Create schema for time.Time type
	if dtType.String() == "time.Time" {
		return ogen.DateTime()
	}

	// Create schema for ogen.Bytes type
	if dtType.String() == "[]uint8" {
		return ogen.Bytes()
	}

	// Create schema for slice/array type
	if dtKind == reflect.Slice || dtKind == reflect.Array {
		return oas.CreateSliceSchema(dtType)
	}

	// Create schema for map type
	if dtKind == reflect.Map {
		return oas.CreateMapSchema(dtType)
	}

	// Create schema for struct type
	if !configs[0].skipStruct && dtKind == reflect.Struct {
		return oas.CreateStructSchema(dtType, configs[0].structName)
	}

	// Return schema of any type
	return ogen.NewSchema()
}

// Create schema for struct type
//
//	Struct may contain nested struct, so we need to use the reference schema.
//	The reference schemas are stored in OpenAPISpec.referenceSchemas and will be created later.
func (oas *OpenAPISpec) CreateStructSchema(dtType reflect.Type, fieldNames ...string) *ogen.Schema {
	fieldNames = append(fieldNames, "")
	schemaName := dtType.Name()
	if dtType.String() == "entity.Entity" {
		schemaName = "Entity.Entity"
	}

	if schemaName == "" {
		schemaName = fieldNames[0]
	}

	if schemaName == "" {
		schemaName = fmt.Sprintf("Schema%03d", oas.GetSchemaIndex())
	}

	oas.referenceSchemas[dtType.String()] = ReferenceSchemaType{
		Name: schemaName,
		Type: dtType,
	}

	return RefSchema(schemaName)
}

// CreateMapSchema creates a schema for a map data type in the OpenAPI specification.
//
//	If the data type is map[string]interface{}, return a schema for map of any type.
//	Otherwise, it retrieves the element schema for the data type using the GetElemSchema method.
//	If the element schema is not found, it a schema for map of any type.
func (oas *OpenAPISpec) CreateMapSchema(dtType reflect.Type) *ogen.Schema {
	mapSchema := ogen.NewSchema()
	mapSchema.SetType("object")
	if dtType.String() == "map[string]interface {}" {
		mapSchema.AdditionalProperties = &ogen.AdditionalProperties{
			Schema: *ogen.NewSchema(),
		}

		return mapSchema
	}

	elemSchema := oas.GetElemSchema(dtType)
	mapSchema.AdditionalProperties = &ogen.AdditionalProperties{
		Schema: *elemSchema,
	}

	return mapSchema
}

// CreateSliceSchema creates a schema for a slice type in the OpenAPI specification.
//
//	If the input type is []interface{}, it creates a schema for slice of any type.
//	Otherwise, it creates a schema for a slice with a specific element type.
//	The function sets the maximum number of items in the schema if the input type is an array.
//	If the element schema cannot be retrieved, it returns nil.
func (oas *OpenAPISpec) CreateSliceSchema(dtType reflect.Type) *ogen.Schema {
	if dtType.String() == "[]interface {}" {
		sliceSchema := ogen.NewSchema().AsArray()
		return sliceSchema
	}

	elemSchema := oas.GetElemSchema(dtType)
	dtKind := dtType.Kind()
	sliceSchema := elemSchema.AsArray()
	if dtKind == reflect.Array {
		arrayLength := uint64(dtType.Len())
		sliceSchema.MaxItems = &arrayLength
	}

	return sliceSchema
}

// GetElemSchema returns the OpenAPI schema for the element type slice or map.
func (oas *OpenAPISpec) GetElemSchema(rt reflect.Type) *ogen.Schema {
	elemType := rt.Elem()
	elemValue := utils.CreateZeroValue(elemType)
	return oas.TypeToOgenSchema(elemValue)
}

// ResolveSchemaReferences resolves the schema references in the OpenAPISpec.
//
//	Schema may reference other schemas or itself.
//	When creating schema, if a field is a struct/map, we create a reference schema for it.
//	These references need to be created so openapi can resolve them.
//	While creating reference schema, there may be new reference schemas created.
//	We need to resolve these new reference schemas as well.
func (oas *OpenAPISpec) ResolveSchemaReferences() {
	hasNewReferenceSchemas := false
	for _, refSchema := range oas.referenceSchemas {
		sType := refSchema.Type
		// Skip if this is not a struct
		if sType.Kind() != reflect.Struct {
			continue
		}

		// Skip if schema is already defined
		if _, ok := oas.ogenSpec.Components.Schemas[refSchema.Name]; ok {
			continue
		}

		structSchema := ogen.NewSchema()

		for i := 0; i < sType.NumField(); i++ {
			field := sType.Field(i)
			fieldName := strings.Split(field.Tag.Get("json"), ",")[0]
			fieldName = utils.If(fieldName == "", field.Name, fieldName)

			// if if field is not exported or fieldName is "-", skip it
			if field.PkgPath != "" || fieldName == "-" {
				continue
			}

			// create zero value of field.Type
			zeroedField := utils.CreateZeroValue(field.Type)
			fieldType := utils.GetDereferencedType(zeroedField)
			if fieldType == nil {
				fieldSchema := ogen.NewSchema()
				structSchema.AddOptionalProperties(fieldSchema.ToProperty(fieldName))
				continue
			}

			// time.Time is a special case, it is a struct but we want to treat it as ogen.DateTime
			if fieldType.String() == "time.Time" {
				structSchema.AddOptionalProperties(ogen.DateTime().ToProperty(fieldName))
				continue
			}

			// If field is a struct, create reference schema for it
			if fieldType.Kind() == reflect.Struct {
				hasNewReferenceSchemas = true
				fieldSchema := oas.CreateStructSchema(fieldType, refSchema.Name+"."+field.Name)
				structSchema.AddOptionalProperties(fieldSchema.ToProperty(fieldName))
				continue
			}

			// Create schema for other types
			fieldSchema := oas.TypeToOgenSchema(zeroedField, &TypeToOgenConfig{skipStruct: true})
			structSchema.AddOptionalProperties(fieldSchema.ToProperty(fieldName))

			continue
		}

		oas.ogenSpec.Components.Schemas[refSchema.Name] = structSchema
	}

	if hasNewReferenceSchemas {
		oas.ResolveSchemaReferences()
		return
	}
}

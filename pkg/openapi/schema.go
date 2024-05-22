package openapi

import (
	"reflect"

	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
	"github.com/ogen-go/ogen"
)

// GetSchemaRefName returns the reference name for a schema with a custom type.
//
//	If no types are provided, it returns the reference name in the format "Schema.<s capitalized>".
//	If types are provided, it returns the reference name in the format "Schema.<s capitalized>.<types[0]>".
func GetSchemaRefName(s string, types ...string) string {
	if len(types) == 0 {
		return "Schema." + utils.Capitalize(s)
	}

	return "Schema." + utils.Capitalize(s) + "." + types[0]
}

// ContentListResponseSchema generates the OpenAPI schema for the content list response.
func ContentListResponseSchema(s *schema.Schema) *ogen.Schema {
	responseSchema := ogen.NewSchema()
	responseSchema.AddRequiredProperties(PrimitiveToOgenTypeMaps[reflect.Uint]().ToProperty("total"))
	responseSchema.AddRequiredProperties(PrimitiveToOgenTypeMaps[reflect.Uint]().ToProperty("per_page"))
	responseSchema.AddRequiredProperties(PrimitiveToOgenTypeMaps[reflect.Uint]().ToProperty("current_page"))
	responseSchema.AddRequiredProperties(PrimitiveToOgenTypeMaps[reflect.Uint]().ToProperty("last_page"))

	itemSchema := ContentDetailSchema(s)
	responseSchema.AddRequiredProperties(itemSchema.AsArray().ToProperty("items"))

	return responseSchema
}

// ContentDetailSchema returns the OpenAPI schema for the content detail.
func ContentDetailSchema(s *schema.Schema) *ogen.Schema {
	return RefSchema(GetSchemaRefName(s.Name))
}

// ContentCreateSchema creates an OpenAPI schema for creating content.
func ContentCreateSchema(s *schema.Schema) *ogen.Schema {
	return RefSchema(GetSchemaRefName(s.Name, SchemaCreate))
}

// SchemasToOGenSchemas converts the schemas defined in the OpenAPISpec to OGenSchemas.
//
//	It iterates over the schema builder's schemas and calls schemaToOGenSchema for each schema.
func (oas *OpenAPISpec) SchemasToOGenSchemas() {
	if oas.config.SchemaBuilder == nil {
		return
	}

	for _, s := range oas.config.SchemaBuilder.Schemas() {
		oas.SchemaToOGenSchema(s)
	}
}

// SchemaToOGenSchema converts a schema.Schema object to an ogen.Schema object and adds it to the OpenAPISpec.
func (oas *OpenAPISpec) SchemaToOGenSchema(s *schema.Schema) {
	schemaName := GetSchemaRefName(s.Name)
	schemaCreateName := GetSchemaRefName(s.Name, SchemaCreate)

	if _, ok := oas.ogenSpec.Components.Schemas[schemaName]; ok {
		return
	}

	ogenSchema := ogen.NewSchema()
	ogenCreateSchema := ogen.NewSchema()
	IDOnlySchema := RefSchema(SchemaIDOnlyName)

	for _, field := range s.Fields {
		// Non-relation field
		if !field.Type.IsRelationType() {
			zeroedField := utils.CreateZeroValue(field.Type.StructType())
			fieldSchema := oas.TypeToOgenSchema(zeroedField)
			ogenSchema.AddOptionalProperties(fieldSchema.ToProperty(field.Name))
			ogenCreateSchema.AddOptionalProperties(fieldSchema.ToProperty(field.Name))
			continue
		}

		// if the rel field is array, we need to create an array schema
		rel := field.Relation
		isArrayField := rel.Type.IsM2M() || (rel.Type.IsO2M() && rel.Owner)
		relSchema := RefSchema(GetSchemaRefName(rel.TargetSchemaName))

		// if not array, we need to create a reference schema
		if !isArrayField {
			ogenSchema.AddOptionalProperties(relSchema.ToProperty(field.Name))
			ogenCreateSchema.AddOptionalProperties(IDOnlySchema.ToProperty(field.Name))
		} else {
			ogenSchema.AddOptionalProperties(relSchema.AsArray().ToProperty(field.Name))
			ogenCreateSchema.AddOptionalProperties(IDOnlySchema.AsArray().ToProperty(field.Name))
		}
	}

	oas.ogenSpec.Components.Schemas[schemaName] = ogenSchema
	oas.ogenSpec.Components.Schemas[schemaCreateName] = ogenCreateSchema
}

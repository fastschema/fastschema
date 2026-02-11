package schema

import (
	"fmt"
	"strings"
)

// Error code constants for categorization
const (
	// Schema-level errors (SCH-xxx)
	ErrCodeSchemaNameRequired       = "SCH-001"
	ErrCodeSchemaLabelFieldRequired = "SCH-002"
	ErrCodeSchemaNamespaceRequired  = "SCH-003"
	ErrCodeSchemaLabelFieldNotFound = "SCH-004"
	ErrCodeSchemaPrimaryFieldMissing = "SCH-005"

	// Field-level errors (FLD-xxx)
	ErrCodeFieldNameRequired       = "FLD-001"
	ErrCodeFieldInvalidType        = "FLD-002"
	ErrCodeFieldEnumRequired       = "FLD-003"
	ErrCodeFieldRelationRequired   = "FLD-004"
	ErrCodeFieldRelationSchema     = "FLD-005"
	ErrCodeFieldRelationType       = "FLD-006"
	ErrCodeFieldRelationField      = "FLD-007"
	ErrCodeFieldTypeInvalid        = "FLD-008"
	ErrCodeFieldInvalidTypeParse   = "FLD-009"

	// Relation errors (REL-xxx)
	ErrCodeRelationTargetNotFound = "REL-001"
	ErrCodeRelationBackRefMissing = "REL-002"
	ErrCodeRelationFKFieldNotFound = "REL-003"
	ErrCodeRelationFKFieldClone    = "REL-004"

	// Builder errors (BLD-xxx)
	ErrCodeBuilderDuplicateSchema    = "BLD-001"
	ErrCodeBuilderSchemaNotFound     = "BLD-002"
	ErrCodeBuilderNotM2MRelation     = "BLD-003"
	ErrCodeBuilderMissingPrimaryKey  = "BLD-004"
	ErrCodeBuilderJunctionFieldFailed = "BLD-005"
)

// ValidFieldTypes returns all valid field type names for error messages
func ValidFieldTypes() string {
	return "string, text, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, bool, time, json, uuid, bytes, enum, relation, file"
}

// ValidFieldTypesShort returns commonly used field types for concise error messages
func ValidFieldTypesShort() string {
	return "string, text, int, uint, uint64, float, float64, bool, time, json, enum, relation, file"
}

// ValidRelationTypes returns valid relation types for error messages
func ValidRelationTypes() string {
	return "o2o (one-to-one), o2m (one-to-many), m2m (many-to-many)"
}

// SchemaError provides detailed, actionable error messages for AI correction
type SchemaError struct {
	Code     string `json:"code"`
	Location string `json:"location"`
	Message  string `json:"message"`
	Fix      string `json:"fix"`
	Example  string `json:"example,omitempty"`
}

// Error implements the error interface with a structured format
func (e *SchemaError) Error() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "[%s] %s: %s", e.Code, e.Location, e.Message)
	if e.Fix != "" {
		fmt.Fprintf(&sb, ". FIX: %s", e.Fix)
	}
	if e.Example != "" {
		fmt.Fprintf(&sb, ". EXAMPLE: %s", e.Example)
	}
	return sb.String()
}

// NewSchemaError creates a new SchemaError with the given parameters
func NewSchemaError(code, location, message, fix string, example ...string) *SchemaError {
	e := &SchemaError{
		Code:     code,
		Location: location,
		Message:  message,
		Fix:      fix,
	}
	if len(example) > 0 && example[0] != "" {
		e.Example = example[0]
	}
	return e
}

// Unwrap returns nil (SchemaError does not wrap other errors)
func (e *SchemaError) Unwrap() error {
	return nil
}

// Is allows error comparison by code
func (e *SchemaError) Is(target error) bool {
	if t, ok := target.(*SchemaError); ok {
		return e.Code == t.Code
	}
	return false
}

// Helper functions for common errors

// SchemaNameRequiredError creates an error for missing schema name
func SchemaNameRequiredError() *SchemaError {
	return NewSchemaError(
		ErrCodeSchemaNameRequired,
		"Schema",
		"Missing 'name' property",
		`Add "name": "<unique_snake_case_name>" to your schema. The name must be unique across all schemas and use snake_case`,
		`"name": "blog_post"`,
	)
}

// SchemaLabelFieldRequiredError creates an error for missing label_field
func SchemaLabelFieldRequiredError(schemaName string) *SchemaError {
	return NewSchemaError(
		ErrCodeSchemaLabelFieldRequired,
		fmt.Sprintf("Schema '%s'", schemaName),
		"Missing 'label_field' property",
		`Add "label_field": "<field_name>" where the field is a string/text field that best represents this record`,
		`"label_field": "title"`,
	)
}

// SchemaNamespaceRequiredError creates an error for missing namespace
func SchemaNamespaceRequiredError(schemaName string) *SchemaError {
	return NewSchemaError(
		ErrCodeSchemaNamespaceRequired,
		fmt.Sprintf("Schema '%s'", schemaName),
		"Missing 'namespace' property",
		`Add "namespace": "<plural_name>" (typically the plural form of the schema name)`,
		`"namespace": "posts"`,
	)
}

// SchemaLabelFieldNotFoundError creates an error for label_field not found in schema
func SchemaLabelFieldNotFoundError(schemaName, labelField string, availableFields []string) *SchemaError {
	fix := `Set label_field to one of the existing string/text fields in the schema. The label_field MUST match an existing field name exactly`
	if len(availableFields) > 0 {
		fix = fmt.Sprintf(`Set label_field to one of these existing string/text fields: [%s]`, strings.Join(availableFields, ", "))
	}
	return NewSchemaError(
		ErrCodeSchemaLabelFieldNotFound,
		fmt.Sprintf("Schema '%s'", schemaName),
		fmt.Sprintf("label_field '%s' not found", labelField),
		fix,
		fmt.Sprintf(`"label_field": "%s"`, func() string {
			if len(availableFields) > 0 {
				return availableFields[0]
			}
			return "name"
		}()),
	)
}

// SchemaLabelFieldSystemSchemaError creates an error for invalid label_field on system schema
func SchemaLabelFieldSystemSchemaError(schemaName, labelField string) *SchemaError {
	labelFieldMap := map[string]string{
		"user": "username",
		"role": "name",
		"file": "name",
	}
	expectedLabelField := labelFieldMap[schemaName]
	return NewSchemaError(
		ErrCodeSchemaLabelFieldNotFound,
		fmt.Sprintf("System schema '%s'", schemaName),
		fmt.Sprintf("label_field '%s' is invalid for system schema", labelField),
		fmt.Sprintf(`System schemas have FIXED label_field values: user='username', role='name', file='name'. You MUST use label_field='%s' for %s`, expectedLabelField, schemaName),
		fmt.Sprintf(`"label_field": "%s"`, expectedLabelField),
	)
}

// FieldNameRequiredError creates an error for missing field name
func FieldNameRequiredError(schemaName string, fieldIndex int) *SchemaError {
	return NewSchemaError(
		ErrCodeFieldNameRequired,
		fmt.Sprintf("Schema '%s', field at index %d", schemaName, fieldIndex),
		"Field is missing 'name' property",
		`Add "name": "<field_name>" using snake_case`,
		`"name": "created_at"`,
	)
}

// FieldInvalidTypeError creates an error for invalid field type
func FieldInvalidTypeError(schemaName, fieldName, invalidType string) *SchemaError {
	return NewSchemaError(
		ErrCodeFieldInvalidType,
		fmt.Sprintf("Schema '%s', field '%s'", schemaName, fieldName),
		fmt.Sprintf("Invalid type '%s'", invalidType),
		fmt.Sprintf("Use one of the valid types: %s", ValidFieldTypesShort()),
		`"type": "string"`,
	)
}

// FieldEnumRequiredError creates an error for enum field without enums array
func FieldEnumRequiredError(schemaName, fieldName string) *SchemaError {
	return NewSchemaError(
		ErrCodeFieldEnumRequired,
		fmt.Sprintf("Schema '%s', field '%s'", schemaName, fieldName),
		"Enum type requires 'enums' array",
		`Add an 'enums' array with value-label pairs. Each enum must have 'value' (stored in DB) and 'label' (displayed to user)`,
		`"enums": [{"value": "draft", "label": "Draft"}, {"value": "published", "label": "Published"}]`,
	)
}

// FieldRelationRequiredError creates an error for relation field without relation object
func FieldRelationRequiredError(schemaName, fieldName string) *SchemaError {
	return NewSchemaError(
		ErrCodeFieldRelationRequired,
		fmt.Sprintf("Schema '%s', field '%s'", schemaName, fieldName),
		"Relation type requires 'relation' object",
		`Add a 'relation' object with 'type', 'schema', and 'field' properties. Both sides of a relation MUST reference each other`,
		`"relation": {"type": "o2m", "schema": "category", "field": "posts"}`,
	)
}

// FieldRelationSchemaRequiredError creates an error for missing relation.schema
func FieldRelationSchemaRequiredError(schemaName, fieldName string) *SchemaError {
	return NewSchemaError(
		ErrCodeFieldRelationSchema,
		fmt.Sprintf("Schema '%s', field '%s'", schemaName, fieldName),
		"relation.schema is required",
		`Specify the target schema name in the relation object`,
		`"relation": {"type": "o2m", "schema": "user", "field": "posts"}`,
	)
}

// FieldRelationTypeRequiredError creates an error for missing/invalid relation.type
func FieldRelationTypeRequiredError(schemaName, fieldName string) *SchemaError {
	return NewSchemaError(
		ErrCodeFieldRelationType,
		fmt.Sprintf("Schema '%s', field '%s'", schemaName, fieldName),
		"relation.type is required",
		fmt.Sprintf("Add relation.type with one of: %s", ValidRelationTypes()),
		`"relation": {"type": "o2m", "schema": "user", "field": "posts"}`,
	)
}

// FieldRelationFieldRequiredError creates an error for missing relation.field
func FieldRelationFieldRequiredError(schemaName, fieldName string) *SchemaError {
	return NewSchemaError(
		ErrCodeFieldRelationField,
		fmt.Sprintf("Schema '%s', field '%s'", schemaName, fieldName),
		"relation.field is required",
		`Specify the back-reference field name in the target schema. The target schema MUST have a relation field with this name pointing back to this schema`,
		`"relation": {"type": "o2m", "schema": "user", "field": "posts"}`,
	)
}

// FieldTypeInvalidError creates an error for invalid/missing field type
func FieldTypeInvalidError(schemaName, fieldName string) *SchemaError {
	return NewSchemaError(
		ErrCodeFieldTypeInvalid,
		fmt.Sprintf("Schema '%s', field '%s'", schemaName, fieldName),
		"Type is invalid or missing",
		fmt.Sprintf("Add a valid 'type' property. Valid types: %s", ValidFieldTypesShort()),
		`"type": "string"`,
	)
}

// FieldTypeParseError creates an error for JSON type parsing failure
func FieldTypeParseError(invalidType string) *SchemaError {
	return NewSchemaError(
		ErrCodeFieldInvalidTypeParse,
		"Field type",
		fmt.Sprintf("Invalid type '%s' in JSON", invalidType),
		fmt.Sprintf("Use one of the valid types: %s", ValidFieldTypes()),
		`"type": "string"`,
	)
}

// RelationTargetNotFoundError creates an error for relation referencing non-existent schema
func RelationTargetNotFoundError(schemaName, fieldName, targetSchemaName string) *SchemaError {
	fix := fmt.Sprintf(`Ensure the target schema '%s' is defined in your schemas array`, targetSchemaName)
	example := fmt.Sprintf(`{"name": "%s", "namespace": "%ss", "label_field": "name", "fields": [...]}`, targetSchemaName, targetSchemaName)

	// Special handling for system schemas
	if targetSchemaName == "user" || targetSchemaName == "role" || targetSchemaName == "file" {
		fix = fmt.Sprintf(`System schema '%s' is pre-defined. To extend it, add a schema with name='%s' and only define additional fields needed for the back-reference`, targetSchemaName, targetSchemaName)
		if targetSchemaName == "user" {
			example = `{"name": "user", "namespace": "users", "label_field": "username", "fields": [{"name": "posts", "type": "relation", "relation": {...}}]}`
		} else {
			example = fmt.Sprintf(`{"name": "%s", "namespace": "%ss", "label_field": "name", "fields": [...]}`, targetSchemaName, targetSchemaName)
		}
	}

	return NewSchemaError(
		ErrCodeRelationTargetNotFound,
		fmt.Sprintf("Schema '%s', field '%s'", schemaName, fieldName),
		fmt.Sprintf("Target schema '%s' does not exist", targetSchemaName),
		fix,
		example,
	)
}

// RelationBackRefError creates an error for missing/invalid back-reference
func RelationBackRefError(sourceSchema, sourceField, targetSchema, targetField string, relationType RelationType) *SchemaError {
	var example string
	switch relationType {
	case O2O:
		example = fmt.Sprintf(`In %s.json: {"name": "%s", "type": "relation", "relation": {"type": "o2o", "schema": "%s", "field": "%s"}}`,
			targetSchema, targetField, sourceSchema, sourceField)
	case O2M:
		example = fmt.Sprintf(`In %s.json: {"name": "%s", "type": "relation", "relation": {"type": "o2m", "schema": "%s", "field": "%s", "owner": true}}`,
			targetSchema, targetField, sourceSchema, sourceField)
	case M2M:
		example = fmt.Sprintf(`In %s.json: {"name": "%s", "type": "relation", "relation": {"type": "m2m", "schema": "%s", "field": "%s"}}`,
			targetSchema, targetField, sourceSchema, sourceField)
	default:
		example = fmt.Sprintf(`In %s.json: {"name": "%s", "type": "relation", "relation": {"type": "...", "schema": "%s", "field": "%s"}}`,
			targetSchema, targetField, sourceSchema, sourceField)
	}

	return NewSchemaError(
		ErrCodeRelationBackRefMissing,
		fmt.Sprintf("Relation '%s.%s' -> '%s.%s'", sourceSchema, sourceField, targetSchema, targetField),
		fmt.Sprintf("Back-reference field '%s' not found in schema '%s' or does not reference back correctly", targetField, targetSchema),
		fmt.Sprintf(`Add a relation field named '%s' in schema '%s' that points back to '%s.%s'. Both sides of a relation MUST reference each other`, targetField, targetSchema, sourceSchema, sourceField),
		example,
	)
}

// BuilderDuplicateSchemaError creates an error for duplicate system schema
func BuilderDuplicateSchemaError(schemaName string) *SchemaError {
	return NewSchemaError(
		ErrCodeBuilderDuplicateSchema,
		fmt.Sprintf("Schema '%s'", schemaName),
		"Duplicate system schema",
		`System schemas can only be defined once. To extend a system schema, use the same name and add only new fields. Do not redefine existing system fields`,
		"",
	)
}

// BuilderSchemaNotFoundError creates an error for schema not found during build
func BuilderSchemaNotFoundError(schemaName string, availableSchemas []string) *SchemaError {
	fix := fmt.Sprintf(`Ensure the schema '%s' is defined in your schemas array. Check for typos in the schema name`, schemaName)
	if len(availableSchemas) > 0 {
		fix += fmt.Sprintf(`. Available schemas: [%s]`, strings.Join(availableSchemas, ", "))
	}
	return NewSchemaError(
		ErrCodeBuilderSchemaNotFound,
		"Builder",
		fmt.Sprintf("Schema '%s' not found", schemaName),
		fix,
		"",
	)
}

// BuilderMissingPrimaryKeyError creates an error for schema missing primary key
func BuilderMissingPrimaryKeyError(schemaName string) *SchemaError {
	return NewSchemaError(
		ErrCodeBuilderMissingPrimaryKey,
		fmt.Sprintf("Schema '%s'", schemaName),
		"Missing primary key field (id)",
		`Add an 'id' field of type uint64 to your schema, OR specify a custom primary_field at the schema level`,
		`{"name": "id", "type": "uint64"}`,
	)
}

// BuilderNotM2MRelationError creates an error when trying to create junction for non-M2M relation
func BuilderNotM2MRelationError(fieldName string) *SchemaError {
	return NewSchemaError(
		ErrCodeBuilderNotM2MRelation,
		fmt.Sprintf("Field '%s'", fieldName),
		"Cannot create junction table for non-M2M relation",
		`Only relations with "type": "m2m" create junction tables. For o2o/o2m relations, one side holds the foreign key`,
		`"relation": {"type": "m2m", "schema": "tag", "field": "posts"}`,
	)
}

// BuilderJunctionFieldError creates an error for failed junction field creation
func BuilderJunctionFieldError(junctionTable string) *SchemaError {
	return NewSchemaError(
		ErrCodeBuilderJunctionFieldFailed,
		fmt.Sprintf("Junction table '%s'", junctionTable),
		"Failed to create junction field",
		`This is an internal error. Ensure both schemas in the M2M relation have valid primary key fields`,
		"",
	)
}

package schema

import (
	"errors"
	"fmt"
	"strings"
)

// Error code constants — dotted lowercase, hierarchical for filtering.
const (
	// Schema-level
	CodeSchemaNameRequired           = "schema.name.required"
	CodeSchemaLabelFieldRequired     = "schema.label_field.required"
	CodeSchemaNamespaceRequired      = "schema.namespace.required"
	CodeSchemaLabelFieldNotFound     = "schema.label_field.not_found"
	CodeSchemaLabelFieldSystemSchema = "schema.label_field.system_schema"
	CodeSchemaPrimaryFieldNotFound   = "schema.primary_field.not_found"
	CodeSchemaPrimaryFieldRequired   = "schema.primary_field.required"
	CodeSchemaIOReadError            = "schema.io.read_error"
	CodeSchemaInitUnknown            = "schema.init.unknown"

	// Field-level
	CodeFieldNameRequired       = "field.name.required"
	CodeFieldTypeInvalid        = "field.type.invalid"
	CodeFieldTypeMissing        = "field.type.missing"
	CodeFieldTypeParseError     = "field.type.parse_error"
	CodeFieldEnumRequired       = "field.enum.required"
	CodeFieldRelationRequired   = "field.relation.required"
	CodeFieldRelationSchemaReq  = "field.relation.schema.required"
	CodeFieldRelationTypeReq    = "field.relation.type.required"
	CodeFieldRelationFieldReq   = "field.relation.field.required"
	CodeFieldNotFound           = "field.not_found"
	CodeFieldFileSchemaRequired = "field.file.schema.required"
	CodeFieldSetterCompileError = "field.setter.compile_error"
	CodeFieldGetterCompileError = "field.getter.compile_error"

	// Relation
	CodeRelationTargetNotFound   = "relation.target.not_found"
	CodeRelationBackRefMissing   = "relation.back_ref.missing"
	CodeRelationFKTargetNotFound = "relation.fk.target.not_found"
	CodeRelationFKCloneFailed    = "relation.fk.clone_failed"
	CodeRelationConfigMissing    = "relation.config.missing"

	// Builder
	CodeBuilderSchemaDuplicate         = "builder.schema.duplicate"
	CodeBuilderSchemaNotFound          = "builder.schema.not_found"
	CodeBuilderSchemaPrimaryKeyMissing = "builder.schema.primary_key.missing"
	CodeBuilderRelationNotM2M          = "builder.relation.not_m2m"
	CodeBuilderJunctionFieldFailed     = "builder.junction_field.create_failed"
)

// FieldError is a per-field error within a schema validation batch.
// Schema context is supplied by the enclosing SchemaErrors wrapper.
type FieldError struct {
	Code    string `json:"code"`
	Field   string `json:"field,omitempty"`
	Index   *int   `json:"index,omitempty"`
	Message string `json:"message"`
	Cause   error  `json:"-"`
}

func (e *FieldError) Error() string {
	var b strings.Builder
	fmt.Fprintf(&b, "[%s] ", e.Code)
	switch {
	case e.Field != "":
		fmt.Fprintf(&b, "field '%s': ", e.Field)
	case e.Index != nil:
		fmt.Fprintf(&b, "field at index %d: ", *e.Index)
	}
	b.WriteString(e.Message)
	return b.String()
}

func (e *FieldError) Unwrap() error { return e.Cause }

func (e *FieldError) Is(target error) bool {
	var t *FieldError
	if errors.As(target, &t) {
		return e.Code == t.Code
	}
	return false
}

// SchemaError is an error with explicit schema context.
// Used for cross-schema errors (relation back-ref), builder-level errors,
// and standalone errors escaping a per-schema batch.
type SchemaError struct {
	Code    string `json:"code"`
	Schema  string `json:"schema,omitempty"`
	Field   string `json:"field,omitempty"`
	Index   *int   `json:"index,omitempty"`
	Target  string `json:"target,omitempty"`
	Message string `json:"message"`
	Cause   error  `json:"-"`
}

func (e *SchemaError) Error() string {
	var b strings.Builder
	fmt.Fprintf(&b, "[%s]", e.Code)
	if e.Schema != "" {
		fmt.Fprintf(&b, " schema '%s'", e.Schema)
		if e.Field != "" {
			fmt.Fprintf(&b, ", field '%s'", e.Field)
		}
		if e.Target != "" {
			fmt.Fprintf(&b, " -> '%s'", e.Target)
		}
	} else if e.Field != "" {
		fmt.Fprintf(&b, " field '%s'", e.Field)
	}
	b.WriteString(": ")
	b.WriteString(e.Message)
	return b.String()
}

func (e *SchemaError) Unwrap() error { return e.Cause }

func (e *SchemaError) Is(target error) bool {
	var t *SchemaError
	if errors.As(target, &t) {
		return e.Code == t.Code
	}
	return false
}

// SchemaErrors collects FieldError instances from a single-schema validation batch.
type SchemaErrors struct {
	Schema      string        `json:"schema"`
	FieldErrors []*FieldError `json:"field_errors,omitempty"`
}

func (e *SchemaErrors) Error() string {
	var b strings.Builder
	fmt.Fprintf(&b, "schema validation error: [%s]", e.Schema)
	for _, fe := range e.FieldErrors {
		b.WriteString("\n  ")
		b.WriteString(fe.Error())
	}
	return b.String()
}

func (e *SchemaErrors) Unwrap() []error {
	out := make([]error, len(e.FieldErrors))
	for i, fe := range e.FieldErrors {
		out[i] = fe
	}
	return out
}

func (e *SchemaErrors) HasErrors() bool { return len(e.FieldErrors) > 0 }

func (e *SchemaErrors) HasCode(code string) bool {
	for _, fe := range e.FieldErrors {
		if fe.Code == code {
			return true
		}
	}
	return false
}

func (e *SchemaErrors) ByCode(code string) []*FieldError {
	var out []*FieldError
	for _, fe := range e.FieldErrors {
		if fe.Code == code {
			out = append(out, fe)
		}
	}
	return out
}

// ToSchemaErrors lifts FieldErrors into []*SchemaError with Schema filled in,
// for aggregation into BuilderErrors at the builder level.
func (e *SchemaErrors) ToSchemaErrors() []*SchemaError {
	out := make([]*SchemaError, len(e.FieldErrors))
	for i, fe := range e.FieldErrors {
		var idx *int
		if fe.Index != nil {
			v := *fe.Index
			idx = &v
		}
		out[i] = &SchemaError{
			Code:    fe.Code,
			Schema:  e.Schema,
			Field:   fe.Field,
			Index:   idx,
			Message: fe.Message,
			Cause:   fe.Cause,
		}
	}
	return out
}

func (e *SchemaErrors) Add(fe *FieldError) {
	if fe != nil {
		e.FieldErrors = append(e.FieldErrors, fe)
	}
}

// BuilderErrors aggregates errors across multiple schemas at build time.
type BuilderErrors struct {
	Errors []*SchemaError `json:"errors"`
}

func (e *BuilderErrors) Error() string {
	if len(e.Errors) == 0 {
		return ""
	}
	msgs := make([]string, len(e.Errors))
	for i, se := range e.Errors {
		msgs[i] = se.Error()
	}
	return strings.Join(msgs, "\n")
}

func (e *BuilderErrors) Unwrap() []error {
	out := make([]error, len(e.Errors))
	for i, se := range e.Errors {
		out[i] = se
	}
	return out
}

func (e *BuilderErrors) Add(se *SchemaError) {
	if se != nil {
		e.Errors = append(e.Errors, se)
	}
}

// AddBatch lifts a SchemaErrors batch into the builder collection.
func (e *BuilderErrors) AddBatch(batch *SchemaErrors) {
	if batch == nil {
		return
	}
	e.Errors = append(e.Errors, batch.ToSchemaErrors()...)
}

// AddAny accepts any error and routes by type. Useful for transitional
// call sites that produce mixed typed/untyped errors (e.g., schema.Init
// may return *SchemaErrors, *FieldError, *SchemaError, or plain error).
func (e *BuilderErrors) AddAny(err error) {
	if err == nil {
		return
	}
	var batch *SchemaErrors
	if errors.As(err, &batch) {
		e.AddBatch(batch)
		return
	}
	var se *SchemaError
	if errors.As(err, &se) {
		e.Errors = append(e.Errors, se)
		return
	}
	var fe *FieldError
	if errors.As(err, &fe) {
		e.Errors = append(e.Errors, &SchemaError{
			Code:    fe.Code,
			Field:   fe.Field,
			Index:   fe.Index,
			Message: fe.Message,
			Cause:   fe.Cause,
		})
		return
	}
	e.Errors = append(e.Errors, &SchemaError{
		Code:    "schema.unknown_error",
		Message: err.Error(),
		Cause:   err,
	})
}

func (e *BuilderErrors) HasErrors() bool { return len(e.Errors) > 0 }

func (e *BuilderErrors) HasCode(code string) bool {
	for _, se := range e.Errors {
		if se.Code == code {
			return true
		}
	}
	return false
}

func (e *BuilderErrors) BySchema(name string) []*SchemaError {
	var out []*SchemaError
	for _, se := range e.Errors {
		if se.Schema == name {
			out = append(out, se)
		}
	}
	return out
}

// ----- Helper constructors -----
//
// Convention:
//   - Helpers used inside Schema.Validate() return *FieldError; schema context
//     is supplied by the enclosing SchemaErrors wrapper at aggregation time.
//   - Helpers used outside a batch (cross-schema, builder, FK, IO, file-field
//     init) return *SchemaError with explicit schema context.

func SchemaNameRequired() *FieldError {
	return &FieldError{
		Code:    CodeSchemaNameRequired,
		Message: "schema name is required",
	}
}

func SchemaLabelFieldRequired() *FieldError {
	return &FieldError{
		Code:    CodeSchemaLabelFieldRequired,
		Message: "label_field is required",
	}
}

func SchemaNamespaceRequired() *FieldError {
	return &FieldError{
		Code:    CodeSchemaNamespaceRequired,
		Message: "namespace is required",
	}
}

func SchemaLabelFieldNotFound(labelField string, available []string) *FieldError {
	msg := fmt.Sprintf("label_field '%s' is not a string/text field", labelField)
	if len(available) > 0 {
		msg += fmt.Sprintf("; available: %s", strings.Join(available, ", "))
	}
	return &FieldError{
		Code:    CodeSchemaLabelFieldNotFound,
		Message: msg,
	}
}

func SchemaLabelFieldSystemSchema(schemaName, labelField string) *FieldError {
	expected := map[string]string{"user": "username", "role": "name", "file": "name"}[schemaName]
	return &FieldError{
		Code:    CodeSchemaLabelFieldSystemSchema,
		Message: fmt.Sprintf("label_field '%s' invalid for system schema '%s'; must be '%s'", labelField, schemaName, expected),
	}
}

func SchemaPrimaryFieldNotFound(schemaName, primaryField string) *SchemaError {
	return &SchemaError{
		Code:    CodeSchemaPrimaryFieldNotFound,
		Schema:  schemaName,
		Message: fmt.Sprintf("primary_field '%s' is not a field in this schema", primaryField),
	}
}

func SchemaPrimaryFieldRequired(schemaName string) *SchemaError {
	return &SchemaError{
		Code:    CodeSchemaPrimaryFieldRequired,
		Schema:  schemaName,
		Message: "primary key field is required (id field or primary_field)",
	}
}

func FieldNameRequired(index int) *FieldError {
	idx := index
	return &FieldError{
		Code:    CodeFieldNameRequired,
		Index:   &idx,
		Message: fmt.Sprintf("field at index %d is missing 'name'", index),
	}
}

func FieldInvalidType(fieldName, invalidType string) *FieldError {
	return &FieldError{
		Code:    CodeFieldTypeInvalid,
		Field:   fieldName,
		Message: fmt.Sprintf("field type '%s' is not recognized", invalidType),
	}
}

func FieldTypeMissing(fieldName string) *FieldError {
	return &FieldError{
		Code:    CodeFieldTypeMissing,
		Field:   fieldName,
		Message: "field type is missing",
	}
}

func FieldTypeParseError(invalidType string) *SchemaError {
	return &SchemaError{
		Code:    CodeFieldTypeParseError,
		Message: fmt.Sprintf("field type '%s' is not a valid type identifier", invalidType),
	}
}

func FieldEnumRequired(fieldName string) *FieldError {
	return &FieldError{
		Code:    CodeFieldEnumRequired,
		Field:   fieldName,
		Message: "enum field requires 'enums' array with value-label pairs",
	}
}

func FieldRelationRequired(fieldName string) *FieldError {
	return &FieldError{
		Code:    CodeFieldRelationRequired,
		Field:   fieldName,
		Message: "relation field requires 'relation' object",
	}
}

func FieldRelationSchemaRequired(fieldName string) *FieldError {
	return &FieldError{
		Code:    CodeFieldRelationSchemaReq,
		Field:   fieldName,
		Message: "relation.schema is required",
	}
}

func FieldRelationTypeRequired(fieldName string) *FieldError {
	return &FieldError{
		Code:    CodeFieldRelationTypeReq,
		Field:   fieldName,
		Message: "relation.type is required (o2o, o2m, or m2m)",
	}
}

func FieldRelationFieldRequired(fieldName string) *FieldError {
	return &FieldError{
		Code:    CodeFieldRelationFieldReq,
		Field:   fieldName,
		Message: "relation.field is required",
	}
}

func FieldNotFound(schemaName, fieldName string) *SchemaError {
	return &SchemaError{
		Code:    CodeFieldNotFound,
		Schema:  schemaName,
		Field:   fieldName,
		Message: fmt.Sprintf("field '%s' is not defined in schema '%s'", fieldName, schemaName),
	}
}

func FieldFileSchemaRequired() *FieldError {
	return &FieldError{
		Code:    CodeFieldFileSchemaRequired,
		Message: "file field requires owning schema name",
	}
}

func FieldSetterCompileError(fieldName string, cause error) *FieldError {
	return &FieldError{
		Code:    CodeFieldSetterCompileError,
		Field:   fieldName,
		Message: fmt.Sprintf("setter expression failed to compile: %v", cause),
		Cause:   cause,
	}
}

func FieldGetterCompileError(fieldName string, cause error) *FieldError {
	return &FieldError{
		Code:    CodeFieldGetterCompileError,
		Field:   fieldName,
		Message: fmt.Sprintf("getter expression failed to compile: %v", cause),
		Cause:   cause,
	}
}

// ----- Cross-schema helpers (return *SchemaError) -----

func RelationTargetNotFound(sourceSchema, sourceField, targetSchema string) *SchemaError {
	return &SchemaError{
		Code:    CodeRelationTargetNotFound,
		Schema:  sourceSchema,
		Field:   sourceField,
		Target:  targetSchema,
		Message: fmt.Sprintf("target schema '%s' is not defined", targetSchema),
	}
}

func RelationBackRefMissing(sourceSchema, sourceField, targetSchema, targetField string, relationType RelationType) *SchemaError {
	return &SchemaError{
		Code:    CodeRelationBackRefMissing,
		Schema:  sourceSchema,
		Field:   sourceField,
		Target:  fmt.Sprintf("%s.%s", targetSchema, targetField),
		Message: fmt.Sprintf("back-reference field '%s' not found in target schema '%s' (relation type %s)", targetField, targetSchema, relationType),
	}
}

func RelationFKTargetNotFound(sourceSchema, sourceField, targetField string) *SchemaError {
	return &SchemaError{
		Code:    CodeRelationFKTargetNotFound,
		Schema:  sourceSchema,
		Field:   sourceField,
		Message: fmt.Sprintf("foreign key target field '%s' not found in target schema", targetField),
	}
}

func RelationFKCloneFailed(sourceSchema, sourceField string) *SchemaError {
	return &SchemaError{
		Code:    CodeRelationFKCloneFailed,
		Schema:  sourceSchema,
		Field:   sourceField,
		Message: "foreign key field clone failed",
	}
}

func RelationConfigMissing(schemaName, fieldName string) *SchemaError {
	return &SchemaError{
		Code:    CodeRelationConfigMissing,
		Schema:  schemaName,
		Field:   fieldName,
		Message: fmt.Sprintf("relation field '%s' is missing 'relation' configuration", fieldName),
	}
}

// ----- Builder-level helpers (return *SchemaError) -----

func BuilderSchemaDuplicate(schemaName string) *SchemaError {
	return &SchemaError{
		Code:    CodeBuilderSchemaDuplicate,
		Schema:  schemaName,
		Message: fmt.Sprintf("system schema '%s' is defined more than once", schemaName),
	}
}

func BuilderSchemaNotFound(schemaName string, available []string) *SchemaError {
	msg := fmt.Sprintf("schema '%s' not found", schemaName)
	if len(available) > 0 {
		msg += fmt.Sprintf("; available: %s", strings.Join(available, ", "))
	}
	return &SchemaError{
		Code:    CodeBuilderSchemaNotFound,
		Schema:  schemaName,
		Message: msg,
	}
}

func BuilderSchemaPrimaryKeyMissing(schemaName string) *SchemaError {
	return &SchemaError{
		Code:    CodeBuilderSchemaPrimaryKeyMissing,
		Schema:  schemaName,
		Message: fmt.Sprintf("schema '%s' is missing primary key field", schemaName),
	}
}

func BuilderRelationNotM2M(schemaName, fieldName string) *SchemaError {
	return &SchemaError{
		Code:    CodeBuilderRelationNotM2M,
		Schema:  schemaName,
		Field:   fieldName,
		Message: "cannot create junction table; relation is not m2m",
	}
}

func BuilderJunctionFieldFailed(junctionTable string) *SchemaError {
	return &SchemaError{
		Code:    CodeBuilderJunctionFieldFailed,
		Schema:  junctionTable,
		Message: fmt.Sprintf("junction field creation failed for '%s'", junctionTable),
	}
}

func SchemaIOReadError(path string, cause error) *SchemaError {
	return &SchemaError{
		Code:    CodeSchemaIOReadError,
		Message: fmt.Sprintf("schema directory '%s' read failed: %v", path, cause),
		Cause:   cause,
	}
}

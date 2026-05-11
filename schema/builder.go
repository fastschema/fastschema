package schema

import (
	"errors"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/fastschema/fastschema/entity"
	"github.com/fastschema/fastschema/pkg/utils"
)

// Builder holds the schema of the database.
type Builder struct {
	dir       string
	schemas   map[string]*Schema
	relations []*Relation
}

func GetSchemasFromDir(dir string, systemSchemaTypes ...any) (map[string]*Schema, error) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil, err
	}

	jsonFiles, err := filepath.Glob(path.Join(dir, "*.json"))
	if err != nil {
		return nil, err
	}

	schemas := make(map[string]*Schema)

	// Create system schemas
	for _, systemSchema := range systemSchemaTypes {
		systemSchema, err := CreateSchema(systemSchema)
		if err != nil {
			return nil, err
		}

		// Prevent duplicate system schemas
		if _, ok := schemas[systemSchema.Name]; ok {
			return nil, BuilderSchemaDuplicate(systemSchema.Name)
		}

		schemas[systemSchema.Name] = systemSchema
	}

	for _, jsonFile := range jsonFiles {
		schema, err := NewSchemaFromJSONFile(jsonFile)

		if err != nil {
			return nil, err
		}

		if existingSchema, ok := schemas[schema.Name]; ok {
			// Merge the schema from JSON file into the existing schema
			// This allows user customizations to override system schema properties
			MergeSchemas(existingSchema, schema)
		} else {
			schemas[schema.Name] = schema
		}
	}

	return schemas, nil
}

// NewBuilderFromSchemas creates a new schema builder from a map of schemas,
// collecting all validation errors instead of stopping at the first one.
// Returns (*Builder, error) — when error is non-nil, the underlying typed
// value is *BuilderErrors and the builder may be incomplete. Use errors.As
// to extract the typed error for rich access (HasCode, Errors slice).
func NewBuilderFromSchemas(dir string, schemas map[string]*Schema, systemSchemaTypes ...any) (*Builder, error) {
	b := &Builder{dir: dir, schemas: map[string]*Schema{}}
	errs := &BuilderErrors{}

	// Create system schemas - collect errors instead of returning early
	for _, systemSchema := range systemSchemaTypes {
		systemSchema, err := CreateSchema(systemSchema)
		if err != nil {
			errs.AddAny(err)
			continue
		}

		if err := systemSchema.Init(false); err != nil {
			errs.AddAny(err)
			continue
		}

		// Prevent duplicate system schemas
		if _, ok := b.schemas[systemSchema.Name]; ok {
			errs.Add(BuilderSchemaDuplicate(systemSchema.Name))
			continue
		}

		b.schemas[systemSchema.Name] = systemSchema
	}

	// Process user schemas - collect errors instead of returning early
	for _, schema := range schemas {
		if existingSchema, ok := b.schemas[schema.Name]; ok {
			// Merge the schema into the existing schema first
			MergeSchemas(existingSchema, schema)
			// Re-init the merged schema to ensure all fields are properly initialized
			if err := existingSchema.Init(false); err != nil {
				errs.AddAny(err)
			}
		} else {
			if err := schema.Init(false); err != nil {
				errs.AddAny(err)
				continue
			}
			b.schemas[schema.Name] = schema
		}
	}

	// Initialize relations and FKs - collect errors via merged Builder.Init
	if initErr := b.Init(); initErr != nil {
		var be *BuilderErrors
		if errors.As(initErr, &be) {
			errs.Errors = append(errs.Errors, be.Errors...)
		} else {
			errs.AddAny(initErr)
		}
	}

	if errs.HasErrors() {
		return b, errs
	}
	return b, nil
}

// NewBuilderFromDir creates a new schema builder from a directory.
func NewBuilderFromDir(dir string, systemSchemaTypes ...any) (*Builder, error) {
	schemas, err := GetSchemasFromDir(dir, systemSchemaTypes...)
	if err != nil {
		return nil, err
	}

	return NewBuilderFromSchemas(dir, schemas)
}

// Clone clones the builder.
func (b *Builder) Clone() *Builder {
	clone := &Builder{
		dir:       b.dir,
		schemas:   map[string]*Schema{},
		relations: []*Relation{},
	}

	for _, schema := range b.schemas {
		clone.schemas[schema.Name] = schema.Clone()
	}

	for _, relation := range b.relations {
		clone.relations = append(clone.relations, relation.Clone())
	}

	return clone
}

// SaveToDir saves all the schemas to a directory.
func (b *Builder) SaveToDir(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err = os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	for _, schema := range b.schemas {
		schemaFile := path.Join(dir, schema.Name+".json")
		if err := schema.SaveToFile(schemaFile); err != nil {
			return err
		}
	}

	return nil
}

// Dir returns the directory of the builder.
// If dirs is not empty, it will set the dir to the first element of dirs.
func (b *Builder) Dir(dirs ...string) string {
	if len(dirs) > 0 {
		b.dir = dirs[0]
	}

	return b.dir
}

// Init initializes the builder, collecting all errors from relation and
// FK creation. FKs only run when relations have no errors (FKs depend on
// resolved relations). Returns *BuilderErrors via the error interface
// when any stage fails; nil otherwise.
func (b *Builder) Init() error {
	errs := &BuilderErrors{}

	if b.schemas == nil {
		b.schemas = map[string]*Schema{}
	}

	relErr := b.CreateRelations()
	if relErr != nil {
		var be *BuilderErrors
		if errors.As(relErr, &be) {
			errs.Errors = append(errs.Errors, be.Errors...)
		} else {
			errs.AddAny(relErr)
		}
	}

	// Only proceed to FKs if no relation errors (FKs depend on relations being valid)
	if !errs.HasErrors() {
		if fkErr := b.CreateFKs(); fkErr != nil {
			var be *BuilderErrors
			if errors.As(fkErr, &be) {
				errs.Errors = append(errs.Errors, be.Errors...)
			} else {
				errs.AddAny(fkErr)
			}
		}
	}

	if errs.HasErrors() {
		return errs
	}
	return nil
}

// SchemaFile returns the json file path of a schema
func (b *Builder) SchemaFile(name string) string {
	return path.Join(b.dir, name+".json")
}

// Schemas returns all schemas
func (b *Builder) Schemas() []*Schema {
	schemas := make([]*Schema, 0)
	for _, schema := range b.schemas {
		schemas = append(schemas, schema)
	}

	return schemas
}

// AddSchema adds a new schema
func (b *Builder) AddSchema(schema *Schema) {
	if b.schemas == nil {
		b.schemas = map[string]*Schema{}
	}

	b.schemas[schema.Name] = schema
}

// ReplaceSchema replaces a schema
func (b *Builder) ReplaceSchema(name string, schema *Schema) {
	b.schemas[name] = schema
}

// Relations returns all relations
func (b *Builder) Relations() []*Relation {
	return b.relations
}

// CreateRelations creates all relations between nodes, collecting all
// errors instead of stopping at the first one. Nil-relation safe — emits
// RelationConfigMissing when a relation-typed field has no Relation
// configured. Returns *BuilderErrors via the error interface when any
// failure occurs; nil otherwise.
func (b *Builder) CreateRelations() error {
	errs := &BuilderErrors{}

	for _, s := range b.schemas {
		for _, field := range s.Fields {
			if !field.Type.IsRelationType() {
				continue
			}

			// Check for nil relation before accessing its properties
			if field.Relation == nil {
				errs.Add(RelationConfigMissing(s.Name, field.Name))
				continue
			}

			relationSchema, err := b.Schema(field.Relation.TargetSchemaName)
			if err != nil {
				errs.AddAny(NewRelationNodeError(s, field))
				continue
			}

			b.relations = append(b.relations, field.Relation.Init(s, relationSchema, field))
		}
	}

	for _, r := range b.relations {
		if r.Type == M2M {
			sourceSchema, err := b.Schema(r.SourceSchemaName)
			if err != nil {
				errs.AddAny(err)
				continue
			}

			junctionSchema, exists, err := b.CreateM2mJunctionSchema(sourceSchema, r)
			if err != nil {
				errs.AddAny(err)
				continue
			}

			r.JunctionSchema = junctionSchema

			if !exists {
				b.AddSchema(junctionSchema)
			}
		}

		if r.BackRef == nil {
			r.BackRef = b.Relation(r.GetBackRefName())
			if r.BackRef == nil {
				errs.AddAny(NewRelationBackRefError(r))
			}
		}
	}

	if errs.HasErrors() {
		return errs
	}
	return nil
}

// CreateFKs creates all foreign keys for relations, collecting all
// errors instead of stopping at the first one. Returns *BuilderErrors
// via the error interface when any failure occurs; nil otherwise.
func (b *Builder) CreateFKs() error {
	errs := &BuilderErrors{}

	for _, relation := range b.relations {
		schema, err := b.Schema(relation.SourceSchemaName)
		if err != nil {
			errs.AddAny(err)
			continue
		}

		// O2O and O2M relations
		if relation.Type.IsO2O() || relation.Type.IsO2M() {
			targetField, err := b.relationTargetField(relation)
			if err != nil {
				errs.AddAny(err)
				continue
			}

			fkField, err := relation.CreateFKField(targetField)
			if err != nil {
				errs.AddAny(err)
				continue
			}

			if fkField != nil {
				foundFKField := schema.Field(fkField.Name)
				if foundFKField != nil {
					MergeFields(foundFKField, fkField)
				} else {
					foundFKField = fkField
					schema.Fields = utils.SliceInsertBeforeElement(
						schema.Fields,
						foundFKField,
						func(f *Field) bool {
							return f.Name == entity.FieldCreatedAt
						},
					)
					schema.dbColumns = utils.SliceInsertBeforeElement(
						schema.dbColumns,
						relation.SourceColumn,
						func(c string) bool {
							return c == entity.FieldCreatedAt
						},
					)
				}

				relation.FKFields = []*Field{foundFKField}
			}
		}
	}

	if errs.HasErrors() {
		return errs
	}
	return nil
}

// Schema returns a node by it's name
func (b *Builder) Schema(name string) (*Schema, error) {
	for _, schema := range b.schemas {
		if schema.Name == name {
			return schema, nil
		}
	}

	// Collect available schema names for the error message
	var availableSchemas []string
	for schemaName := range b.schemas {
		availableSchemas = append(availableSchemas, schemaName)
	}
	return nil, BuilderSchemaNotFound(name, availableSchemas)
}

// Relation returns a relation by it's name
func (b *Builder) Relation(name string) *Relation {
	for _, relation := range b.relations {
		if relation.Name == name {
			return relation
		}
	}

	return nil
}

func (b *Builder) relationTargetField(r *Relation) (*Field, error) {
	targetSchema, err := b.Schema(r.TargetSchemaName)
	if err != nil {
		return nil, err
	}

	targetColumnName := r.TargetColumn
	if targetColumnName == "" {
		targetColumnName = targetSchema.PrimaryKeyName()
		if targetColumnName == "" {
			targetColumnName = entity.FieldID
		}
	}
	targetField := targetSchema.Field(targetColumnName)
	if targetField == nil {
		return nil, ErrFieldNotFound(targetSchema.Name, targetColumnName)
	}

	return targetField, nil
}

func (b *Builder) CreateM2mJunctionSchema(sourceSchema *Schema, r *Relation) (*Schema, bool, error) {
	if r == nil || !r.Type.IsM2M() {
		return nil, false, BuilderRelationNotM2M(r.SourceSchemaName, r.Name)
	}

	targetSchema, err := b.Schema(r.TargetSchemaName)
	if err != nil {
		return nil, false, err
	}

	// first/second FK names default to field identifiers but allow overrides
	defaultFirstFKName := utils.If(r.IsBidi(), r.SourceSchemaName, r.SourceFieldName)
	defaultSecondFKName := r.TargetFieldName

	sourceColumn := utils.If(r.SourceColumn == "", defaultFirstFKName, r.SourceColumn)
	targetColumn := utils.If(r.TargetColumn == "", defaultSecondFKName, r.TargetColumn)
	fKColumnNames := []string{sourceColumn, targetColumn}
	r.RelationSchemas = []*Schema{targetSchema, sourceSchema}
	r.SourceColumn = sourceColumn
	r.TargetColumn = targetColumn

	tableNameParts := []string{sourceColumn, targetColumn}
	sort.Strings(tableNameParts)
	if r.JunctionTable == "" {
		r.JunctionTable = strings.Join(tableNameParts, "_")
	}

	// If the junction schema already exists, skip creating it
	junctionSchema, _ := b.Schema(r.JunctionTable)
	if junctionSchema != nil {
		return junctionSchema, true, nil
	}

	junctionSchema = &Schema{
		Name:             r.JunctionTable,
		Namespace:        r.JunctionTable,
		LabelFieldName:   fKColumnNames[0],
		IsJunctionSchema: true,
		IsSystemSchema:   true,
	}

	targetIDField := targetSchema.PrimaryField()
	if targetIDField == nil {
		return nil, false, BuilderSchemaPrimaryKeyMissing(targetSchema.Name)
	}

	sourceIDField := sourceSchema.PrimaryField()
	if sourceIDField == nil {
		return nil, false, BuilderSchemaPrimaryKeyMissing(sourceSchema.Name)
	}

	firstFKField := cloneReferenceField(targetIDField, fKColumnNames[0])
	secondFKField := cloneReferenceField(sourceIDField, fKColumnNames[1])
	for _, fkField := range []*Field{firstFKField, secondFKField} {
		if fkField == nil {
			return nil, false, BuilderJunctionFieldFailed(r.JunctionTable)
		}
		fkField.IsSystemField = true
		fkField.Immutable = true
		fkField.Optional = false
		fkField.Unique = false
		fkField.DB.Increment = false
		fkField.DB.Key = DBEmptyKey
	}

	junctionSchema.Fields = []*Field{firstFKField, secondFKField}

	if err := junctionSchema.Init(true); err != nil {
		return nil, false, err
	}

	return junctionSchema, false, nil
}

// NewBuilderFromRelations builds a Builder from schemas containing only
// relation-type fields and validates the cross-schema relation topology.
// Use case: relation-graph-only validators (UI relation editor, build/CI
// checks, schema-diff tooling) where full schema definitions aren't
// available because primitives have been caller-stripped.
//
// For each input schema:
//   - A synthetic UUID primary field is prepended if neither a primary
//     field nor an "id" field already exists (so junction/back-ref
//     resolution has a target to point at).
//   - LabelFieldName defaults to entity.FieldID if empty.
//   - Only relation-type fields are initialized; primitives are skipped.
//
// The returned error carries primarily relation.* and builder.* codes
// inside a *BuilderErrors typed value (extract via errors.As). Schema-level
// (schema.*) and most per-field (field.*) errors are NOT detected. Exception:
// if a relation-type field has a malformed setter or getter expression,
// Field.Init surfaces field.setter.compile_error or field.getter.compile_error
// since expression compilation is unconditional. Callers needing full
// per-schema validation should use NewBuilderFromSchemas with full schemas,
// or Schema.Init per schema.
//
// Trade-off: cannot detect FK column-name conflicts with primitive fields
// or PK type mismatches across relations. CreateFKs is intentionally
// skipped because FK column generation needs primitive field types that
// have been stripped from the input.
//
// NOTE: this function mutates the input *Schema values (synthetic id field
// prepend, LabelFieldName default, s.initialized = true). Callers that
// reuse the same *Schema pointers after this call should be aware of that.
func NewBuilderFromRelations(schemas map[string]*Schema) (*Builder, error) {
	b := &Builder{schemas: map[string]*Schema{}}
	errs := &BuilderErrors{}

	for _, s := range schemas {
		if s == nil {
			continue
		}

		// Inject synthetic PK if no primary field exists.
		// defaultIDField() returns the project's standard UUID PK
		// (matches system-schema convention since commit 95b55b4).
		if s.PrimaryField() == nil {
			s.Fields = append([]*Field{defaultIDField()}, s.Fields...)
		}

		// Default LabelFieldName to id so downstream code that relies on it
		// (e.g. relation/junction setup) has a valid reference.
		if s.LabelFieldName == "" {
			s.LabelFieldName = entity.FieldID
		}

		// Initialize relation fields only. Primitive fields are caller-
		// stripped; calling Field.Init on them would either no-op or fail
		// depending on type — we skip them entirely to keep behavior crisp.
		for _, f := range s.Fields {
			if !f.Type.IsRelationType() {
				continue
			}
			if err := f.Init(s.Name); err != nil {
				errs.AddAny(err)
			}
		}

		// Mark initialized so downstream lookups (Builder.Schema) don't
		// attempt re-initialization that would invoke ensurePrimaryField
		// and primitive Field.Init.
		s.initialized = true

		if _, dup := b.schemas[s.Name]; dup {
			errs.Add(BuilderSchemaDuplicate(s.Name))
			continue
		}
		b.schemas[s.Name] = s
	}

	// Single source of truth for cross-schema relation validation.
	if relErr := b.CreateRelations(); relErr != nil {
		var be *BuilderErrors
		if errors.As(relErr, &be) {
			errs.Errors = append(errs.Errors, be.Errors...)
		} else {
			errs.AddAny(relErr)
		}
	}

	// CreateFKs INTENTIONALLY SKIPPED — needs primitive types.

	if errs.HasErrors() {
		return b, errs
	}
	return b, nil
}

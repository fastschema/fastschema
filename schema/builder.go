package schema

import (
	"fmt"
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
			return nil, fmt.Errorf("system schema %s already exists", systemSchema.Name)
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

// NewBuilderFromSchemas creates a new schema from a map of schemas.
func NewBuilderFromSchemas(dir string, schemas map[string]*Schema, systemSchemaTypes ...any) (*Builder, error) {
	b := &Builder{dir: dir, schemas: map[string]*Schema{}}

	// Create system schemas
	for _, systemSchema := range systemSchemaTypes {
		systemSchema, err := CreateSchema(systemSchema)
		if err != nil {
			return nil, err
		}

		if err := systemSchema.Init(false); err != nil {
			return nil, err
		}

		// Prevent duplicate system schemas
		if _, ok := b.schemas[systemSchema.Name]; ok {
			return nil, fmt.Errorf("system schema %s already exists", systemSchema.Name)
		}

		b.schemas[systemSchema.Name] = systemSchema
	}

	for _, schema := range schemas {
		if err := schema.Init(false); err != nil {
			return nil, err
		}

		b.schemas[schema.Name] = schema
	}

	return b, b.Init()
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

// Init initializes the builder.
func (b *Builder) Init() (err error) {
	if b.schemas == nil {
		b.schemas = map[string]*Schema{}
	}

	if err = b.CreateRelations(); err != nil {
		return err
	}

	if err = b.CreateFKs(); err != nil {
		return err
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

// CreateRelations creates all relations between nodes
func (b *Builder) CreateRelations() (err error) {
	for _, s := range b.schemas {
		for _, field := range s.Fields {
			if !field.Type.IsRelationType() {
				continue
			}

			relationSchema, err := b.Schema(field.Relation.TargetSchemaName)
			if err != nil {
				return NewRelationNodeError(s, field)
			}

			b.relations = append(b.relations, field.Relation.Init(s, relationSchema, field))
		}
	}

	for _, r := range b.relations {
		if r.Type == M2M {
			sourceSchema, err := b.Schema(r.SourceSchemaName)
			if err != nil {
				return err
			}

			junctionSchema, exists, err := b.CreateM2mJunctionSchema(sourceSchema, r)
			if err != nil {
				return err
			}

			r.JunctionSchema = junctionSchema

			if !exists {
				b.AddSchema(junctionSchema)
			}
		}

		if r.BackRef == nil {
			r.BackRef = b.Relation(r.GetBackRefName())
			if r.BackRef == nil {
				return NewRelationBackRefError(r)
			}
		}
	}

	return nil
}

// CreateFKs creates all foreign keys for relations
func (b *Builder) CreateFKs() error {
	for _, relation := range b.relations {
		schema, err := b.Schema(relation.SourceSchemaName)
		if err != nil {
			return err
		}

		// O2O and O2M relations
		if relation.Type.IsO2O() || relation.Type.IsO2M() {
			targetField, err := b.relationTargetField(relation)
			if err != nil {
				return err
			}

			fkField, err := relation.CreateFKField(targetField)
			if err != nil {
				return err
			}

			// if relation.FKFields != nil {
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

	return nil
}

// Schema returns a node by it's name
func (b *Builder) Schema(name string) (*Schema, error) {
	for _, schema := range b.schemas {
		if schema.Name == name {
			return schema, nil
		}
	}

	return nil, fmt.Errorf("schema %s not found", name)
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
		return nil, false, fmt.Errorf("field %s is not a m2m relation", r.Name)
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
		return nil, false, fmt.Errorf("schema %s is missing id field", targetSchema.Name)
	}

	sourceIDField := sourceSchema.PrimaryField()
	if sourceIDField == nil {
		return nil, false, fmt.Errorf("schema %s is missing id field", sourceSchema.Name)
	}

	firstFKField := cloneReferenceField(targetIDField, fKColumnNames[0])
	secondFKField := cloneReferenceField(sourceIDField, fKColumnNames[1])
	for _, fkField := range []*Field{firstFKField, secondFKField} {
		if fkField == nil {
			return nil, false, fmt.Errorf("failed to create junction field for %s", r.JunctionTable)
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

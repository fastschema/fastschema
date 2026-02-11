package schema

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/fastschema/fastschema/entity"
	"github.com/fastschema/fastschema/pkg/utils"
)

// Schema holds the node data.
type Schema struct {
	*SystemSchema `json:"-"`

	initialized bool
	dbColumns   []string `json:"-"`

	Name             string          `json:"name"`
	Namespace        string          `json:"namespace"`
	LabelFieldName   string          `json:"label_field"`
	PrimaryFieldName string          `json:"primary_field,omitempty"`
	DisableTimestamp bool            `json:"disable_timestamp"`
	Fields           []*Field        `json:"fields"`
	IsSystemSchema   bool            `json:"is_system_schema,omitempty"`
	IsJunctionSchema bool            `json:"is_junction_schema,omitempty"`
	DB               *SchemaDB       `json:"db,omitempty"`
	Settings         *SchemaSettings `json:"settings,omitempty"`
	primaryField     string          `json:"-"`
}

// NewSchemaFromJSON creates a new node from a json string.
func NewSchemaFromJSON(jsonData string) (*Schema, error) {
	n := &Schema{}
	if err := json.Unmarshal([]byte(jsonData), &n); err != nil {
		return nil, err
	}

	return n, nil
}

// NewSchemaFromJSONFile creates a new node from a json file.
func NewSchemaFromJSONFile(jsonFile string) (*Schema, error) {
	s := &Schema{}
	jsonData, err := os.ReadFile(jsonFile)

	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal(jsonData, &s); err != nil {
		return nil, err
	}

	return s, nil
}

func NewSchemaFromMap(data map[string]any) (*Schema, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	return NewSchemaFromJSON(string(jsonData))
}

// Init initializes the node.
func (s *Schema) Init(disableIDColumn bool) error {
	if s.initialized {
		return nil
	}

	defer func() {
		s.initialized = true
	}()

	if err := s.Validate(); err != nil {
		return err
	}

	if err := s.ensurePrimaryField(disableIDColumn); err != nil {
		return err
	}

	for _, f := range s.Fields {
		if err := f.Init(s.Name); err != nil {
			return err
		}
		if !f.Type.IsRelationType() {
			s.dbColumns = append(s.dbColumns, f.Name)
		}
	}

	if !s.DisableTimestamp {
		timeFields := [][4]string{
			{entity.FieldCreatedAt, "Created At", "false", "NOW()"},
			{entity.FieldUpdatedAt, "Updated At", "true"},
			{entity.FieldDeletedAt, "Deleted At", "true"},
		}

		for _, timeField := range timeFields {
			tsField := &Field{
				IsSystemField: true,
				Immutable:     true,
				Type:          TypeTime,
				Name:          timeField[0],
				Label:         timeField[1],
				Optional:      timeField[2] == "true",
				Filterable:    false,
				Sortable:      false,
			}

			if timeField[3] == "NOW()" {
				tsField.Default = "CURRENT_TIMESTAMP"
			}

			existedTimeField := s.Field(timeField[0])
			if existedTimeField != nil {
				MergeFields(existedTimeField, tsField)
			} else {
				s.dbColumns = append(s.dbColumns, timeField[0])
				s.Fields = append(s.Fields, tsField)
				if err := tsField.Init(); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// Clone returns a copy of the schema.
func (s *Schema) Clone() *Schema {
	clone := &Schema{
		Name:             s.Name,
		Namespace:        s.Namespace,
		LabelFieldName:   s.LabelFieldName,
		PrimaryFieldName: s.PrimaryFieldName,
		DisableTimestamp: s.DisableTimestamp,
		dbColumns:        s.dbColumns,
		IsSystemSchema:   s.IsSystemSchema,
		IsJunctionSchema: s.IsJunctionSchema,
		Settings:         s.Settings,
		primaryField:     s.primaryField,
	}

	for _, f := range s.Fields {
		clone.Fields = append(clone.Fields, f.Clone())
	}

	return clone
}

// MergeSchemas merges the source schema into the target schema.
// This is used to merge user customizations from JSON files into system schemas.
// - Fields that exist in both schemas will be merged (source overrides target for non-system fields)
// - Fields that only exist in source will be added to target (if not system fields)
// - Schema-level properties from source will override target if they are set
// - DB indexes from source will be merged with target
func MergeSchemas(target, source *Schema) {
	// Merge schema-level properties (only if explicitly set in source)
	if source.Namespace != "" && source.Namespace != target.Namespace {
		target.Namespace = source.Namespace
	}
	if source.LabelFieldName != "" {
		target.LabelFieldName = source.LabelFieldName
	}
	if source.PrimaryFieldName != "" {
		target.PrimaryFieldName = source.PrimaryFieldName
	}
	if source.DisableTimestamp {
		target.DisableTimestamp = source.DisableTimestamp
	}
	if source.Settings != nil {
		target.Settings = source.Settings
	}

	// Merge fields
	for _, sourceField := range source.Fields {
		existingField := target.Field(sourceField.Name)
		if existingField != nil {
			// Field exists in target - merge the properties
			// Only merge if source field is not a system field (user customization)
			if !sourceField.IsSystemField {
				MergeFields(existingField, sourceField)
			}
		} else {
			// Field doesn't exist in target - add it (only non-system fields from JSON)
			if !sourceField.IsSystemField {
				target.Fields = append(target.Fields, sourceField)
			}
		}
	}

	// Merge DB indexes
	if source.DB != nil && source.DB.Indexes != nil {
		if target.DB == nil {
			target.DB = &SchemaDB{
				Indexes: []*SchemaDBIndex{},
			}
		}

		// Add indexes from source that don't exist in target
		for _, sourceIndex := range source.DB.Indexes {
			indexExists := false
			for _, targetIndex := range target.DB.Indexes {
				if targetIndex.Name == sourceIndex.Name {
					indexExists = true
					break
				}
			}
			if !indexExists {
				target.DB.Indexes = append(target.DB.Indexes, sourceIndex)
			}
		}
	}
}

// SaveToFile saves the schema to a file.
func (s *Schema) SaveToFile(filename string) error {
	filteredSchema := s.Clone()
	filteredSchema.Fields = utils.Filter(filteredSchema.Fields, func(field *Field) bool {
		return !field.IsSystemField
	})

	fileData, err := json.MarshalIndent(filteredSchema, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, fileData, 0600)
}

// HasField checks if the schema has a field.
func (s *Schema) HasField(fieldName string) bool {
	existedFields := utils.Filter(s.Fields, func(f *Field) bool {
		return f.Name == fieldName
	})

	return len(existedFields) > 0
}

// Field return field by it's name
func (s *Schema) Field(name string) *Field {
	return utils.Find(s.Fields, func(f *Field) bool {
		return f.Name == name
	})
}

// PrimaryField returns the primary key field definition.
func (s *Schema) PrimaryField() *Field {
	primaryName := s.PrimaryKeyName()
	if primaryName == "" {
		return nil
	}

	return s.Field(primaryName)
}

// PrimaryKeyName returns the resolved primary key field name.
func (s *Schema) PrimaryKeyName() string {
	if s.primaryField != "" {
		return s.primaryField
	}

	if s.PrimaryFieldName != "" {
		return s.PrimaryFieldName
	}

	if s.Field(entity.FieldID) != nil {
		return entity.FieldID
	}

	return ""
}

// Validate inspects the fields of the schema for validation errors.
func (s *Schema) Validate() error {
	var schemaErrors []string
	if s.Name == "" {
		schemaErrors = append(schemaErrors, "name is required")
	}
	if s.LabelFieldName == "" {
		schemaErrors = append(schemaErrors, "label_field is required")
	}

	if s.Namespace == "" {
		schemaErrors = append(schemaErrors, "namespace is required")
	}

	// if len(s.Fields) == 0 {
	// 	schemaErrors = append(schemaErrors, "fields is required")
	// }

	hasLabelField := false

	for _, field := range s.Fields {
		if s.LabelFieldName == field.Name {
			hasLabelField = true
		}

		if field.Name == "" {
			schemaErrors = append(schemaErrors, fmt.Sprintf("field %s: name is required", field.Name))
		}

		if field.Label == "" {
			field.Label = field.Name
		}

		if !field.Type.IsRelationType() && field.Type != TypeEnum {
			if !field.Type.Valid() {
				schemaErrors = append(schemaErrors, fmt.Sprintf("field %s: invalid field type %s", field.Name, field.Type))
			}
		}

		if field.Type == TypeEnum && len(field.Enums) == 0 {
			schemaErrors = append(schemaErrors, fmt.Sprintf("field %s: enums values is required", field.Name))
		}

		if field.Type.IsRelationType() {
			relation := field.Relation
			if relation == nil {
				schemaErrors = append(schemaErrors, fmt.Sprintf("field %s: relation is required", field.Name))
				break
			}

			if relation.TargetSchemaName == "" {
				schemaErrors = append(schemaErrors, fmt.Sprintf("field %s: relation schema is required", field.Name))
			}

			if relation.Type == RelationInvalid {
				schemaErrors = append(schemaErrors, fmt.Sprintf("field %s: relation type is required", field.Name))
			}

			if relation.Type == M2M && relation.TargetFieldName == "" {
				schemaErrors = append(schemaErrors, fmt.Sprintf("field %s: m2m relation ref field name is required", field.Name))
			}
		}

		if field.Type == TypeInvalid {
			schemaErrors = append(schemaErrors, fmt.Sprintf("field %s: type is invalid", field.Name))
		}
	}

	// If schema is system schema, skip checking label field
	if !s.IsSystemSchema && s.LabelFieldName != "" && !hasLabelField {
		schemaErrors = append(schemaErrors, fmt.Sprintf("label field '%s' is not found", s.LabelFieldName))
	}

	if len(schemaErrors) > 0 {
		return fmt.Errorf("schema validation error: [%s] %s", s.Name, strings.Join(schemaErrors, "\n "))
	}

	return nil
}

func (s *Schema) ensurePrimaryField(disableIDColumn bool) error {
	var candidate *Field
	var autoCreated bool
	userDefined := s.PrimaryFieldName != ""

	if s.PrimaryFieldName != "" && s.PrimaryFieldName != entity.FieldID {
		candidate = s.Field(s.PrimaryFieldName)
		if candidate == nil {
			return fmt.Errorf("schema %s: primary field '%s' is not found", s.Name, s.PrimaryFieldName)
		}
	}

	if candidate == nil && !disableIDColumn {
		candidate = s.Field(entity.FieldID)
	}

	if candidate == nil && !disableIDColumn {
		candidate = defaultIDField()
		autoCreated = true
		s.Fields = append([]*Field{candidate}, s.Fields...)
	}

	if candidate == nil {
		s.primaryField = ""
		if disableIDColumn {
			return nil
		}

		return fmt.Errorf("schema %s: primary key field is required", s.Name)
	}

	applyPrimaryFieldDefaults(candidate, autoCreated)

	s.primaryField = candidate.Name
	if candidate.Name != entity.FieldID || userDefined {
		s.PrimaryFieldName = candidate.Name
	} else if !userDefined {
		s.PrimaryFieldName = ""
	}

	return nil
}

func ErrFieldNotFound(schemaName, fieldName string) error {
	return fmt.Errorf("field %s.%s not found", schemaName, fieldName)
}

func defaultIDField() *Field {
	idField := &Field{
		Name:  entity.FieldID,
		Type:  TypeUint64,
		Label: "ID",
		DB: &FieldDB{
			Attr:      "UNSIGNED",
			Key:       DBPrimaryKey,
			Increment: true,
		},
		IsSystemField: true,
	}

	applyPrimaryFieldDefaults(idField, true)

	return idField
}

func applyPrimaryFieldDefaults(field *Field, autoCreated bool) {
	if field == nil {
		return
	}

	if field.Type == TypeInvalid {
		field.Type = TypeUint64
	}

	dbProvided := field.DB != nil
	if !dbProvided {
		field.DB = &FieldDB{}
	}

	if field.Label == "" {
		field.Label = "ID"
	}

	if field.DB.Key == DBEmptyKey {
		field.DB.Key = DBPrimaryKey
	}

	if field.DB.Attr == "" && field.Type.IsUnsignedInteger() {
		field.DB.Attr = "UNSIGNED"
	}

	if !field.Type.IsInteger() {
		field.DB.Increment = false
	} else if autoCreated || !dbProvided {
		field.DB.Increment = true
	}

	field.IsSystemField = field.IsSystemField || autoCreated

	field.Immutable = true
	field.Unique = true
	field.Filterable = true
	field.Sortable = true
	field.Optional = false
}

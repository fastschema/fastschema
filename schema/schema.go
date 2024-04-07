package schema

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/fastschema/fastschema/pkg/utils"
)

type SchemaDBIndex struct {
	Name    string   `json:"name,omitempty"`
	Unique  bool     `json:"unique,omitempty"`
	Columns []string `json:"columns,omitempty"`
}

type SchemaDB struct {
	Indexes []*SchemaDBIndex `json:"indexes,omitempty"`
}

// Schema holds the node data.
type Schema struct {
	initialized      bool
	Name             string    `json:"name"`
	Namespace        string    `json:"namespace"`
	LabelFieldName   string    `json:"label_field"`
	DisableTimestamp bool      `json:"disable_timestamp"`
	Fields           []*Field  `json:"fields"`
	IsSystemSchema   bool      `json:"is_system_schema,omitempty"`
	IsJunctionSchema bool      `json:"is_junction_schema,omitempty"`
	DB               *SchemaDB `json:"db,omitempty"`
	// RelationsFKColumns map[string][]string `json:"-"`
	DBColumns []string `json:"-"`
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

	// s.RelationsFKColumns = map[string][]string{}

	if !disableIDColumn {
		s.Fields = append([]*Field{{
			Name:          FieldID,
			Type:          TypeUint64,
			IsSystemField: true,
			Label:         "ID",
			DB: &FieldDB{
				Attr:      "UNSIGNED",
				Key:       "UNI",
				Increment: true,
			},
			Unique:     true,
			Filterable: true,
			Sortable:   true,
		}}, s.Fields...)
	}

	for _, f := range s.Fields {
		f.Init(s.Name)
		if !f.Type.IsRelationType() {
			s.DBColumns = append(s.DBColumns, f.Name)
		}
	}

	if !s.DisableTimestamp {
		timeFields := [][4]string{
			{FieldCreatedAt, "Created At", "false", "CURRENT_TIMESTAMP"},
			{FieldUpdatedAt, "Updated At", "true"},
			{FieldDeletedAt, "Deleted At", "true"},
		}

		for _, timeField := range timeFields {
			tsField := &Field{
				IsSystemField: true,
				Type:          TypeTime,
				Name:          timeField[0],
				Label:         timeField[1],
				Optional:      timeField[2] == "true",
				Filterable:    false,
				Sortable:      false,
			}

			if timeField[3] != "" {
				tsField.Default = timeField[3]
			}

			s.DBColumns = append(s.DBColumns, timeField[0])
			s.Fields = append(s.Fields, tsField)
			tsField.Init()
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
		DisableTimestamp: s.DisableTimestamp,
		DBColumns:        s.DBColumns,
		IsSystemSchema:   s.IsSystemSchema,
		IsJunctionSchema: s.IsJunctionSchema,
	}

	for _, f := range s.Fields {
		clone.Fields = append(clone.Fields, f.Clone())
	}

	return clone
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
func (s *Schema) HasField(field *Field) bool {
	existedFields := utils.Filter(s.Fields, func(oldField *Field) bool {
		return oldField.Name == field.Name
	})

	return len(existedFields) > 0
}

// Field return field by it's name
func (s *Schema) Field(name string) (*Field, error) {
	for _, f := range s.Fields {
		if f.Name == name {
			return f, nil
		}
	}

	return nil, fmt.Errorf("field %s.%s not found", s.Name, name)
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

	if len(s.Fields) == 0 {
		schemaErrors = append(schemaErrors, "fields is required")
	}

	hasLabelField := false

	for _, field := range s.Fields {
		if s.LabelFieldName == field.Name {
			hasLabelField = true
		}
		if strings.ToLower(field.Name) == FieldID {
			schemaErrors = append(schemaErrors, fmt.Sprintf("field %s: field name can't be 'id'", field.Name))
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

	if s.LabelFieldName != "" && !hasLabelField {
		schemaErrors = append(schemaErrors, fmt.Sprintf("label field '%s' is not found", s.LabelFieldName))
	}

	if len(schemaErrors) > 0 {
		return fmt.Errorf("schema validation error: [%s] %s", s.Name, strings.Join(schemaErrors, "\n "))
	}

	return nil
}

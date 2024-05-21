package schema

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/jinzhu/inflection"
)

// While converting struct to schema, we need to convert the struct fields to schema fields.
//
//	We will need to know which struct field maps to which schema field.
//	This is done by making pairs of struct field and schema field.
type FieldPair struct {
	StructField string
	SchemaField string
}

type SystemSchema struct {
	Instance   any
	RType      reflect.Type
	FieldPairs []*FieldPair
}

// CustomizableSchema allows the struct to customize the schema.
//
//	Method `Schema` should return the customized schema.
//	It supports the following types of customization:
//		- Customize schema information:
//			- Name
//			- Namespace
//			- LabelFieldName
//			- DisableTimestamp
//			- IsJunctionSchema
//			- DB
//		- Customize all field information, except `IsSystemField`:
//			- Adding `Fields` ([]*Field) property to the returned schema to customize the fields.
//			- Each field item contains data to customize schema field (partially or fully).
//			- Matching will be done based on the field name.
//			- Matched field item will be used to customize the corresponding schema field.
//			- Non-matched field item will be ignored.
type CustomizableSchema interface {
	Schema() *Schema
}

// CreateSchema creates a schema from the given struct.
//
//	Schema information:
//		- Will be created from the struct type: name, namespace.
//		- Customize through CustomizableSchema interface: Supports all schema properties/fields.
//		- Customize through specific field "_ any `\"fs:name=schema_name, name\"`": Does not support customizing fields.
//	Fields information:
//		- The struct fields will be converted to schema fields.
//		- The struct field tags will be used to customize the schema fields.
//		- The struct field  conversion will follow the following rules:
//		- If the field is not exported, it will be ignored.
//		- If the field is a primitive/time type, it will be mapped arcorrdingly reflectTypesToFieldType.
//		- If the field is a struct or slice of struct, it must has a field tag define it as a relation.
//		- If the field is a complex type of primitives, it must has a field tag to define the type as json.
//		- Enums field must be string with struct tag: fs.enums="[{'value': 'v1', 'label': 'L1'}, {'value': 'v2', 'label': 'L2'}]".
func CreateSchema(t any) (*Schema, error) {
	tType := utils.GetDereferencedType(t)
	if tType == nil || tType.Kind() != reflect.Struct {
		return nil, fmt.Errorf("can not create schema from invalid type %T", t)
	}

	schemaName := utils.ToSnakeCase(tType.Name())
	schemaNamespace := inflection.Plural(schemaName)
	schema := &Schema{
		SystemSchema: &SystemSchema{
			Instance:   t,
			RType:      tType,
			FieldPairs: []*FieldPair{},
		},
		// Non-customizable fields
		IsSystemSchema: true,
		Fields:         []*Field{},

		// Customizable through method Schema()
		Name:             schemaName,
		Namespace:        schemaNamespace,
		LabelFieldName:   "",
		DisableTimestamp: false,
		IsJunctionSchema: false,
		DB:               nil,
	}

	if _, err := schema.Customize(); err != nil {
		return nil, err
	}

	return schema, nil
}

func (s *Schema) Customize() (_ *Schema, err error) {
	if err := s.CreateFields(); err != nil {
		return nil, err
	}

	// Customize using specific field _
	if sf, ok := s.RType.FieldByName("_"); ok {
		// Customize common properties
		fieldProps := utils.ParseStructFieldTag(sf, "fs")
		for key, value := range fieldProps {
			switch key {
			case "name":
				s.Name = value
			case "namespace":
				s.Namespace = value
			case "label_field":
				s.LabelFieldName = value
			case "disable_timestamp":
				s.DisableTimestamp = true
			case "is_junction_schema":
				s.IsJunctionSchema = true
			}
		}

		// Customize db property
		dbTag := sf.Tag.Get("fs.db")
		if dbTag != "" {
			if s.DB, err = utils.ParseHJSON[*SchemaDB]([]byte(dbTag)); err != nil {
				return nil, fmt.Errorf("%s._=%s: invalid db format: %w", s.Name, dbTag, err)
			}
		}
	}

	// Customize using the CustomizableSchema interface
	dtZero := utils.CreateZeroValue(s.RType)
	if customizableSchema, ok := dtZero.(CustomizableSchema); ok {
		customizedSchema := customizableSchema.Schema()

		if customizedSchema.Name != "" {
			s.Name = customizedSchema.Name
		}

		if customizedSchema.Namespace != "" {
			s.Namespace = customizedSchema.Namespace
		}

		if customizedSchema.LabelFieldName != "" {
			s.LabelFieldName = customizedSchema.LabelFieldName
		}

		if customizedSchema.DisableTimestamp {
			s.DisableTimestamp = customizedSchema.DisableTimestamp
		}

		if customizedSchema.IsJunctionSchema {
			s.IsJunctionSchema = customizedSchema.IsJunctionSchema
		}

		if customizedSchema.DB != nil {
			s.DB = customizedSchema.DB
		}

		// loop through the customizedSchema.Fields and update the matched schema field
		for _, customizedField := range customizedSchema.Fields {
			matchedSchemaField := s.Field(customizedField.Name)

			if matchedSchemaField == nil {
				return nil, fmt.Errorf(
					"%s: customized field %s not found in struct %T",
					s.Name, customizedField.Name, s.Instance,
				)
			}

			MergeFields(matchedSchemaField, customizedField)
			matchedSchemaField.IsSystemField = true
		}
	}

	if err := s.Validate(); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Schema) CreateFields() error {
	stringFields := []string{}

	for i := 0; i < s.RType.NumField(); i++ {
		sf := s.RType.Field(i)
		field, err := s.CreateField(sf)
		if err != nil {
			return err
		}

		if field == nil {
			continue
		}

		s.Fields = append(s.Fields, field)
		s.FieldPairs = append(s.FieldPairs, &FieldPair{
			StructField: sf.Name,
			SchemaField: field.Name,
		})
		if field.Type == TypeString {
			stringFields = append(stringFields, field.Name)
		}
	}

	// Set the label field to the first string field if not set
	if s.LabelFieldName == "" && len(stringFields) > 0 {
		s.LabelFieldName = stringFields[0]
	}

	return nil
}

func (s *Schema) CreateField(sf reflect.StructField) (*Field, error) {
	fieldName := utils.GetStructFieldName(sf)
	if fieldName == "" {
		return nil, nil
	}

	ignoreByJSONTag := strings.Split(sf.Tag.Get("json"), ",")[0] == "-"
	ignoreByFSTag := strings.Split(sf.Tag.Get("fs"), ",")[0] == "-"
	if ignoreByJSONTag || ignoreByFSTag {
		return nil, nil
	}

	// if field struct tag fs has key label, then set the schema label field
	// if field struct tag fs has key type, override the field type
	field := &Field{
		IsSystemField: true,
		Name:          fieldName,
		Label:         utils.Title(fieldName),
		Type:          FieldTypeFromReflectType(sf.Type),
	}

	if err := s.ExtendFieldByTag(sf, field); err != nil {
		return nil, err
	}

	// If the field type is invalid, ignore the field
	if !field.Type.Valid() {
		return nil, nil
	}

	if s.Field(field.Name) != nil {
		return nil, fmt.Errorf("field %s.%s already exists", s.Name, field.Name)
	}

	return field, nil
}

// ExtendFieldByTag extends the given field by parsing the struct field tag and updating the field properties accordingly.
//
//	Common properties format:
//	- E.g: `fs:"type=string,name=custom_name,label=Custom Label,size=10,multiple,unique,optional,sortable,filterable,default=10"`
//	- Supported field properties:
//		- type: Tag fs="type=string".
//		- name: Use json tag to customize the field name, e.g. `json:"custom_name"`.
//		- label: Tag fs="label=Custom Label".
//		- size: Tag fs="size=10".
//		- multiple: Tag fs="multiple".
//		- unique: Tag fs="unique".
//		- optional: Tag fs="optional".
//		- sortable: Tag fs="sortable".
//		- filterable: Tag fs="filterable".
//		- default: Tag fs="default=10", if field is time, use RFC3339 format.
//	Complex properties format:
//	- E.g: `fs.enums="[{'value': 'v1', 'label': 'L1'}, {'value': 'v2', 'label': 'L2'}]"`
//	- E.g: `fs.relation="{'type': 'o2m', 'schema': 'post', 'field': 'categories', 'owner': true}"`
//		- enums: Only for string fields. Tag fs.enums=hjson -> []*FieldEnum
//		- relation: Tag fs.relation=hjson -> *Relation
//		- renderer: Tag fs.renderer=hjson -> *Field.Renderer
//		- db: Tag fs.db=hjson -> *FieldDB
//
// To extend other field properties, you can implement the CustomizableSchema interface.
func (s *Schema) ExtendFieldByTag(sf reflect.StructField, field *Field) error {
	// Customize field name using json tag
	jsonTag := strings.Split(sf.Tag.Get("json"), ",")
	if len(jsonTag) > 0 && jsonTag[0] != "" {
		field.Name = jsonTag[0]
	}

	// Customize common properties
	fieldProps := utils.ParseStructFieldTag(sf, "fs")
	for key, value := range fieldProps {
		switch key {
		case "type":
			fieldType := FieldTypeFromString(value)
			if !fieldType.Valid() {
				return fmt.Errorf("%s.%s.%s=%s: invalid field type", s.Name, field.Name, key, value)
			}
			field.Type = fieldType
		case "label":
			field.Label = value
		case "size":
			size, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return fmt.Errorf("%s.%s.%s=%s: invalid field size", s.Name, field.Name, key, value)
			}
			field.Size = size
		case "multiple":
			field.IsMultiple = true
		case "unique":
			field.Unique = true
		case "optional":
			field.Optional = true
		case "sortable":
			field.Sortable = true
		case "filterable":
			field.Filterable = true
		case "default":
			// only set the default value if the field type is primitive type
			if field.Type.IsAtomic() {
				value, err := StringToFieldValue[any](field, value)
				if err != nil {
					return fmt.Errorf("%s %w", s.Name, err)
				}

				field.Default = value
			}

		// label_field is a special field that is used to display the label of the schema content.
		// if the label field name is available, then set the schema label field.
		case "label_field":
			s.LabelFieldName = field.Name
		}
	}

	// Customize enum property
	enumsTag := sf.Tag.Get("fs.enums")
	if enumsTag != "" {
		fieldEnums, err := utils.ParseHJSON[[]*FieldEnum]([]byte(enumsTag))
		if err != nil {
			return fmt.Errorf("%s.%s=%s: invalid enums format: %w", s.Name, field.Name, enumsTag, err)
		}

		field.Type = TypeEnum
		field.Enums = fieldEnums
	}

	// Customize relation property
	relationTag := sf.Tag.Get("fs.relation")
	if relationTag != "" {
		relation, err := utils.ParseHJSON[*Relation]([]byte(relationTag))
		if err != nil {
			return fmt.Errorf("%s.%s=%s: invalid relation format: %w", s.Name, field.Name, relationTag, err)
		}

		field.Type = TypeRelation
		field.Relation = relation
	}

	// Customize renderer property
	rendererTag := sf.Tag.Get("fs.renderer")
	if rendererTag != "" {
		renderer, err := utils.ParseHJSON[*FieldRenderer]([]byte(rendererTag))
		if err != nil {
			return fmt.Errorf("%s.%s=%s: invalid renderer format: %w", s.Name, field.Name, rendererTag, err)
		}

		field.Renderer = renderer
	}

	// Customize DB property
	dbTag := sf.Tag.Get("fs.db")
	if dbTag != "" {
		db, err := utils.ParseHJSON[*FieldDB]([]byte(dbTag))
		if err != nil {
			return fmt.Errorf("%s.%s=%s: invalid db format: %w", s.Name, field.Name, dbTag, err)
		}

		field.DB = db
	}

	return nil
}

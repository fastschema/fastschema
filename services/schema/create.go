package schemaservice

import (
	"fmt"

	"github.com/fastschema/fastschema/app"
	"github.com/fastschema/fastschema/pkg/errors"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func (ss *SchemaService) Create(c app.Context, newSchemaData *schema.Schema) (*schema.Schema, error) {
	schemaFile := fmt.Sprintf("%s/%s.json", ss.app.SchemaBuilder().Dir(), newSchemaData.Name)
	updateSchemas := map[string]*schema.Schema{}

	if utils.IsFileExists(schemaFile) {
		return nil, errors.BadRequest("schema already exists")
	}

	if err := newSchemaData.Validate(); err != nil {
		return nil, errors.UnprocessableEntity(err.Error())
	}

	// add the back reference field to the related schema
	for _, field := range newSchemaData.Fields {
		if !field.Type.IsRelationType() {
			continue
		}

		field.Init(newSchemaData.Name)
		relation := field.Relation
		targetSchema, err := ss.app.SchemaBuilder().Schema(relation.TargetSchemaName)
		// targetSchema, ok := su.updateSchemas[relation.TargetSchemaName]
		if err != nil {
			return nil, errors.BadRequest(
				"Invalid field '%s.%s'. Target schema '%s' not found",
				newSchemaData.Name,
				field.Name,
				relation.TargetSchemaName,
			)
		}

		// check if target schema has the back reference field
		isTargetRelationOwner := !relation.Owner
		targetRelationField := &schema.Field{
			Type:     schema.TypeRelation,
			Name:     relation.TargetFieldName,
			Label:    cases.Title(language.Und, cases.NoLower).String(relation.TargetFieldName),
			Optional: isTargetRelationOwner,
			Sortable: false,
			Relation: &schema.Relation{
				TargetSchemaName: newSchemaData.Name,
				TargetFieldName:  field.Name,
				Type:             relation.Type,
				Owner:            isTargetRelationOwner,
				Optional:         isTargetRelationOwner,
			},
		}
		if targetSchema.HasField(targetRelationField.Name) {
			return nil, errors.BadRequest(
				"Invalid field '%s.%s'. Target schema '%s' already has field '%s'",
				newSchemaData.Name,
				field.Name,
				relation.TargetSchemaName,
				relation.TargetFieldName,
			)
		}

		// add the back reference field to the related schema in the new schema builder
		targetSchema.Fields = append(targetSchema.Fields, targetRelationField)
		updateSchemas[targetSchema.Name] = targetSchema
	}

	if err := newSchemaData.SaveToFile(schemaFile); err != nil {
		return nil, errors.InternalServerError("could not save schema")
	}

	// update the related schemas
	for _, schema := range updateSchemas {
		if err := schema.SaveToFile(fmt.Sprintf("%s/%s.json", ss.app.SchemaBuilder().Dir(), schema.Name)); err != nil {
			return nil, errors.InternalServerError("could not save schema")
		}
	}

	if err := ss.app.Reload(nil); err != nil {
		c.Logger().Errorf("could not reload app: %s", err.Error())
		return nil, errors.InternalServerError("could not reload app: %s", err.Error())
	}

	return newSchemaData, nil
}

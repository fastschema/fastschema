package schemaservice

import (
	"fmt"
	"io"
	"os"

	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/errors"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
	// "golang.org/x/text/cases"
	// "golang.org/x/text/language"
)

func (ss *SchemaService) Upload(c fs.Context, _ any) (*schema.Schema, error) {

	// upload to tmp dir
	files, err := c.Files()
	
	// upload files to schema/tmp dir
	tmpDir, err := os.MkdirTemp("", "tmp_schemas")
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if _, err := ss.Disk().Put(c.Context(), file); err != nil {
	}

	// schemaBuilder, error := schema.NewBuilderFromDir(ss.app.SchemaBuilder().Dir())
	// get all schemas from directory
	schemas, err := schema.GetSchemasFromDir("")
	if err != nil {
		return nil, err
	}
	// init mapSchemas by schema name
	// mapSchemas := map[string]*schema.Schema{}
	// for _, schema := range schemas {
	// 	mapSchemas[schema.Name] = schema
	// }
	// validate one by one schema with the following rules
	for _, sc := range schemas {
		currentSchemaFile := fmt.Sprintf("%s/%s.json", ss.app.SchemaBuilder().Dir(), sc.Name)
		if utils.IsFileExists(currentSchemaFile) {
			return nil, errors.BadRequest("schema already exists in current system")
		}
	}

	_, error := schema.NewBuilderFromDir(ss.app.SchemaBuilder().Dir())
	if err != nil {
		return nil, error
	}
	// if err := newSchemaData.SaveToFile(schemaFile); err != nil {
	// 	return nil, errors.InternalServerError("could not save schema")
	// }

	if err := ss.app.Reload(c.Context(), nil); err != nil {
		c.Logger().Errorf("could not reload app: %s", err.Error())
		return nil, errors.InternalServerError("could not reload app: %s", err.Error())
	}

	return nil, nil
}


// schemaBuilder, error := schema.NewBuilderFromDir(ss.app.SchemaBuilder().Dir())
// 	// get all schemas from directory
// 	schemas, err := schema.GetSchemasFromDir("")
// 	if err != nil {
// 		return nil, err
// 	}
// 	// init mapSchemas by schema name
// 	mapSchemas := map[string]*schema.Schema{}
// 	for _, schema := range schemas {
// 		mapSchemas[schema.Name] = schema
// 	}
// 	// validate one by one schema with the following rules
// 	for _, sc := range schemas {
// 		schemaFile := fmt.Sprintf("%s/%s.json", ss.app.SchemaBuilder().Dir(), sc.Name)
// 		if utils.IsFileExists(schemaFile) {
// 			return nil, errors.BadRequest("schema already exists in current system")
// 		}

// 		for _, field := range sc.Fields {
// 			if !field.Type.IsRelationType() {
// 				continue
// 			}

// 			field.Init(sc.Name)
// 			relation := field.Relation

// 			// check if target schema exists
// 			// skip check if the relation type is bi-directional
// 			if relation.TargetSchemaName == sc.Name {
// 				continue
// 			}

// 			targetSchema := mapSchemas[relation.TargetSchemaName]
// 			// targetSchema, ok := su.updateSchemas[relation.TargetSchemaName]
// 			if targetSchema == nil {
// 				return nil, errors.BadRequest(
// 					"Invalid field '%s.%s'. Target schema '%s' not found",
// 					sc.Name,
// 					field.Name,
// 					relation.TargetSchemaName,
// 				)
// 			}

// 			// check if target schema has the back reference field
// 			isTargetRelationOwner := !relation.Owner
// 			targetRelationField := &schema.Field{
// 				Type:     schema.TypeRelation,
// 				Name:     relation.TargetFieldName,
// 				Label:    cases.Title(language.Und, cases.NoLower).String(relation.TargetFieldName),
// 				Optional: isTargetRelationOwner,
// 				Sortable: false,
// 				Relation: &schema.Relation{
// 					TargetSchemaName: sc.Name,
// 					TargetFieldName:  field.Name,
// 					Type:             relation.Type,
// 					Owner:            isTargetRelationOwner,
// 					Optional:         isTargetRelationOwner,
// 				},
// 			}
// 			if !targetSchema.HasField(targetRelationField.Name) {
// 				return nil, errors.BadRequest(
// 					"Invalid field '%s.%s'. Target schema '%s' has no field '%s'",
// 					sc.Name,
// 					field.Name,
// 					relation.TargetSchemaName,
// 					relation.TargetFieldName,
// 				)
// 			}

// 			// add the back reference field to the related schema in the new schema builder
// 			// targetSchema.Fields = append(targetSchema.Fields, targetRelationField)
// 			// updateSchemas[targetSchema.Name] = targetSchema

// 			if err := sc.Validate(); err != nil {
// 				return nil, errors.UnprocessableEntity(err.Error())
// 			}

// 		}
// 	}

// 	// if err := newSchemaData.SaveToFile(schemaFile); err != nil {
// 	// 	return nil, errors.InternalServerError("could not save schema")
// 	// }

// 	if err := ss.app.Reload(c.Context(), nil); err != nil {
// 		c.Logger().Errorf("could not reload app: %s", err.Error())
// 		return nil, errors.InternalServerError("could not reload app: %s", err.Error())
// 	}

// 	return nil, nil
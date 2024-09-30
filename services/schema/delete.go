package schemaservice

import (
	"os"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/errors"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
)

func (ss *SchemaService) Delete(c fs.Context, _ any) (fs.Map, error) {
	schemaName := c.Arg("name")
	currentSchema, err := ss.app.SchemaBuilder().Schema(schemaName)
	if err != nil {
		return nil, errors.NotFound(err.Error())
	}

	hasRelation := false
	// remove relation fields if the field type is relation
	updateFields := utils.Filter(currentSchema.Fields, func(field *schema.Field) bool {
		if field.Type.IsRelationType() && field.Relation.TargetSchemaName != schemaName {
			hasRelation = true
			return false
		}
		return true
	})

	// update the schema with the new fields without the relation fields
	if hasRelation {
		updateData := &SchemaUpdateData{
			Data: &schema.Schema{
				Name:           currentSchema.Name,
				Fields:         updateFields,
				Namespace:      currentSchema.Namespace,
				LabelFieldName: currentSchema.LabelFieldName,
			},
			RenameFields: []*db.RenameItem{},
			RenameTables: []*db.RenameItem{},
		}

		su := &SchemaUpdate{
			updateData:           updateData,
			currentSchemaBuilder: ss.app.SchemaBuilder(),
			updateSchemas:        map[string]*schema.Schema{},
			currentSchema:        currentSchema,
		}

		if err := su.update(); err != nil {
			return nil, errors.InternalServerError(err.Error())
		}
	}

	// delete the schema file
	schemaFile := ss.app.SchemaBuilder().SchemaFile(schemaName)
	if err := os.Remove(schemaFile); err != nil {
		return nil, errors.InternalServerError(err.Error())
	}

	if err := ss.app.Reload(c, nil); err != nil {
		return nil, errors.InternalServerError(err.Error())
	}

	return fs.Map{"message": "Schema deleted"}, nil
}

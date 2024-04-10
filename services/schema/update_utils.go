package schemaservice

import (
	"fmt"
	"os"
	"path"
	"time"

	"github.com/fastschema/fastschema/app"
	"github.com/fastschema/fastschema/pkg/errors"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func (su *SchemaUpdate) prepare() (err error) {
	oldSchemasDir := su.DB().SchemaBuilder().Dir()

	// Create the new schema update dir
	if su.updateDir, err = createSchemaUpdateDir(oldSchemasDir, true); err != nil {
		return err
	}

	if su.updateData.RenameTables == nil {
		su.updateData.RenameTables = []*app.RenameItem{}
	}

	// add the type and schema information to the rename fields
	su.updateData.RenameFields = utils.Map(su.updateData.RenameFields, func(rf *app.RenameItem) *app.RenameItem {
		rf.Type = "column"
		rf.SchemaName = su.oldSchema.Name
		rf.SchemaNamespace = su.oldSchema.Namespace
		return rf
	})

	if err := su.updateBasicData(oldSchemasDir); err != nil {
		return err
	}

	if err := su.updateRelationsData(); err != nil {
		return err
	}

	if err = su.renameDir(oldSchemasDir); err != nil {
		return err
	}

	su.updateData.RenameFields = utils.Filter(su.updateData.RenameFields, func(f *app.RenameItem) bool {
		return f.From != f.To
	})

	return err
}

// updateBasicData update the data that only exist in the schema json files
func (su *SchemaUpdate) updateBasicData(oldSchemasDir string) (err error) {
	if su.updateSchemas, err = schema.GetSchemasFromDir(oldSchemasDir); err != nil {
		return err
	}

	// overwrite the current schema with the new schema
	su.updateSchemas[su.updateData.Schema.Name] = su.updateData.Schema

	if err = su.applyRenameSchema(); err != nil {
		return err
	}

	if err = su.applyRenameSchemaNamespace(); err != nil {
		return err
	}

	if err = su.updateRelatedSchemas(); err != nil {
		return err
	}

	for _, schema := range su.updateSchemas {
		schemaFile := path.Join(su.updateDir, fmt.Sprintf("%s.json", schema.Name))
		if err = schema.SaveToFile(schemaFile); err != nil {
			return err
		}
	}

	su.newSchemaBuilder, err = schema.NewBuilderFromDir(su.updateDir)
	if err != nil {
		return err
	}

	if su.newSchema, err = su.newSchemaBuilder.Schema(su.updateData.Schema.Name); err != nil {
		return err
	}

	return nil
}

// applyRenameSchema check if the schema name is renamed,
// if so, rename the schema name in the relation fields.
func (su *SchemaUpdate) applyRenameSchema() (err error) {
	newSchemaName := ""
	if su.oldSchema.Name != su.updateData.Schema.Name {
		newSchemaName = su.updateData.Schema.Name
	}

	if newSchemaName == "" {
		return nil
	}

	// loop the current schema fields to rename the bidi relation fields
	for i, field := range su.updateData.Schema.Fields {
		if !field.Type.IsRelationType() {
			continue
		}

		// check if the relation target schema is renamed
		if field.Relation.TargetSchemaName == su.oldSchema.Name {
			su.updateData.Schema.Fields[i].Relation.TargetSchemaName = newSchemaName
		}
	}

	// loop through all the schemas to rename the relation fields
	for _, s := range su.updateSchemas {
		for i, field := range s.Fields {
			if !field.Type.IsRelationType() {
				continue
			}

			// check if the relation target schema is renamed
			if field.Relation.TargetSchemaName == su.oldSchema.Name {
				s.Fields[i].Relation.TargetSchemaName = newSchemaName
			}
		}
	}

	return nil
}

// applyRenameSchemaNamespace check if the schema namespace is renamed,
// if so, rename the schema namespace in the relation fields.
func (su *SchemaUpdate) applyRenameSchemaNamespace() (err error) {
	newSchemaNamespace := ""
	if su.oldSchema.Namespace != su.updateData.Schema.Namespace {
		newSchemaNamespace = su.updateData.Schema.Namespace
	}

	if newSchemaNamespace == "" {
		return nil
	}

	su.updateData.RenameTables = append(su.updateData.RenameTables, &app.RenameItem{
		Type: "table",
		From: su.oldSchema.Namespace,
		To:   newSchemaNamespace,
	})

	return nil
}

// updateRelationsData update the data that related to the relation fields
func (su *SchemaUpdate) updateRelationsData() (err error) {
	// add the back reference field to the related schema
	for _, field := range su.updateData.Schema.Fields {
		if !field.Type.IsRelationType() {
			continue
		}

		// if field exist in su.oldSchema,
		// then the field is not added, no need to add the back reference field.
		// otherwise, add the back reference field
		if su.oldSchema.HasField(field) {
			continue
		}

		if err = su.renameRelations(field); err != nil {
			return err
		}
	}

	return nil
}

func (su *SchemaUpdate) updateRelatedSchemas() error {
	// remove the back reference field from the related schema
	for _, field := range su.oldSchema.Fields {
		if !field.Type.IsRelationType() {
			continue
		}

		// if field exist in su.updateData.Schema,
		// then the field is not removed, no need to remove the back reference field.
		// otherwise, remove the back reference field
		if su.updateData.Schema.HasField(field) {
			continue
		}

		// find the related schema from the new schema builder
		relation := field.Relation
		targetSchema, ok := su.updateSchemas[relation.TargetSchemaName]
		if !ok {
			return fmt.Errorf("schema '%s' not found", relation.TargetSchemaName)
		}

		// remove the back reference field from the related schema in the new schema builder
		targetSchema.Fields = utils.Filter(targetSchema.Fields, func(f *schema.Field) bool {
			return f.Name != relation.TargetFieldName
		})
	}

	// add the back reference field to the related schema
	for _, field := range su.updateData.Schema.Fields {
		if !field.Type.IsRelationType() {
			continue
		}

		field.Init(su.updateData.Schema.Name)

		// if field exist in su.oldSchema,
		// then the field is not added, no need to add the back reference field.
		// otherwise, add the back reference field
		if su.oldSchema.HasField(field) {
			continue
		}

		relation := field.Relation
		targetSchema, ok := su.updateSchemas[relation.TargetSchemaName]
		if !ok {
			return fmt.Errorf("schema '%s' not found", relation.TargetSchemaName)
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
				TargetSchemaName: su.oldSchema.Name,
				TargetFieldName:  field.Name,
				Type:             relation.Type,
				Owner:            isTargetRelationOwner,
				Optional:         isTargetRelationOwner,
			},
		}
		if targetSchema.HasField(targetRelationField) {
			return errors.New(
				fmt.Sprintf(
					"Invalid field '%s.%s'. Target relation field '%s' already exist in schema '%s'",
					su.oldSchema.Name,
					field.Name,
					targetRelationField.Name,
					targetSchema.Name,
				),
			)
		}

		// add the back reference field to the related schema in the new schema builder
		targetSchema.Fields = append(targetSchema.Fields, targetRelationField)
	}

	return nil
}

// renameRelations check for the newly added relation fields,
// to find out if the relation field was renamed, if so,
// add the rename columns for the foreign key columns
func (su *SchemaUpdate) renameRelations(newField *schema.Field) (err error) {
	// filter from RenameFields to find the original field name (old field name)
	matchedFields := utils.Filter(su.updateData.RenameFields, func(f *app.RenameItem) bool {
		return f.To == newField.Name
	})

	// if the field is not renamed, then we don't need to rename the foreign key column
	if len(matchedFields) == 0 {
		return nil
	}

	originalField, err := su.oldSchema.Field(matchedFields[0].From)
	if err != nil {
		return err
	}

	defer func() {
		// RenameFields will be used to rename the db table columns,
		// remove the current relation field from the RenameFields because these fields are not exist in table,
		// we use the FK columns instead of these fields.
		su.updateData.RenameFields = utils.Filter(su.updateData.RenameFields, func(rf *app.RenameItem) bool {
			// with m2m relations, we have added a rename for two columns of the junction table,
			// the two columns of the junction table have the same name with the two fields of the two relation schemas.
			// so there will be two rename fields with the same From and To:
			// - 1st: The rename that was sent from the client,
			// 				this is the renaming for the relation field of the current editting schema,
			// 				this rename will have the SchemaName equal to the original schema name and must be removed.
			// - 2nd: The rename that was added for the junction table columns,
			// 				this rename will have the SchemaName equal to the junction table name and must be kept.
			// we have to remove the 1st rename because the relation field is not exist in the current editting schema table,
			isFiltered := (rf.SchemaName == su.oldSchema.Name && // check the 1st rename
				rf.From == originalField.Name && rf.To == newField.Name) // check From and To
			return !isFiltered
		})
	}()

	relationSchema, err := su.newSchemaBuilder.Schema(newField.Relation.TargetSchemaName)
	if err != nil {
		return err
	}

	newFieldRelation := newField.Relation.Clone()
	newFieldRelation.Init(su.newSchema, relationSchema, newField)
	originalFieldRelation := originalField.Relation

	// process the m2m relation
	if newFieldRelation.Type == schema.M2M {
		originalJunctionSchema := originalFieldRelation.JunctionSchema
		newJunctionSchema, _, err := su.newSchemaBuilder.CreateM2mJunctionSchema(su.oldSchema, newFieldRelation)
		if err != nil {
			return err
		}

		// Ent do not perform the rename table operation, so we have to do it manually.
		// - Rename the junction table columns: Add the rename columns to the RenameFields so DiffHook can rename the columns.
		// - Rename the junction table: Add the rename table to the RenameTables and manually rename the table in ApplyHook.
		su.updateData.RenameTables = append(su.updateData.RenameTables, &app.RenameItem{
			Type:            "table",
			From:            originalJunctionSchema.Namespace,
			To:              newJunctionSchema.Namespace,
			IsJunctionTable: true,
		})

		su.updateData.RenameFields = append(su.updateData.RenameFields, &app.RenameItem{
			Type:            "column",
			From:            originalFieldRelation.FKColumns.CurrentColumn,
			To:              newFieldRelation.FKColumns.CurrentColumn,
			SchemaName:      originalJunctionSchema.Name,      // newJunctionSchema.Name
			SchemaNamespace: originalJunctionSchema.Namespace, // newJunctionSchema.Namespace
		})

		su.updateData.RenameFields = append(su.updateData.RenameFields, &app.RenameItem{
			Type:            "column",
			From:            originalFieldRelation.FKColumns.TargetColumn,
			To:              newFieldRelation.FKColumns.TargetColumn,
			SchemaName:      originalJunctionSchema.Name,      // newJunctionSchema.Name
			SchemaNamespace: originalJunctionSchema.Namespace, // newJunctionSchema.Namespace
		})

		return nil
	}

	// Other relation types
	// if the fields don't has foreign key, then we don't need to rename the foreign key column
	if !originalFieldRelation.HasFKs() || !newFieldRelation.HasFKs() {
		return nil
	}

	su.updateData.RenameFields = append(su.updateData.RenameFields, &app.RenameItem{
		Type:            "column",
		From:            originalFieldRelation.GetTargetFKColumn(),
		To:              newFieldRelation.GetTargetFKColumn(),
		SchemaName:      su.oldSchema.Name,
		SchemaNamespace: su.oldSchema.Namespace,
	})

	return nil
}

func (su *SchemaUpdate) renameDir(oldSchemasDir string) error {
	backupDir := su.updateDir + "_backup"

	if err := os.Rename(oldSchemasDir, backupDir); err != nil {
		return err
	}

	if err := os.Rename(su.updateDir, oldSchemasDir); err != nil {
		return err
	}

	if err := os.Rename(backupDir, su.updateDir); err != nil {
		return err
	}

	return nil
}

func createSchemaUpdateDir(currentSchemaDir string, create bool) (string, error) {
	parentDir := path.Join(path.Dir(currentSchemaDir), "backup")
	now := time.Now()
	schemaDirName := path.Base(currentSchemaDir)
	backupDirName := fmt.Sprintf("%s_%s", schemaDirName, now.Format("20060102150405"))
	updateDir := path.Join(parentDir, backupDirName)

	if create {
		if err := os.MkdirAll(updateDir, 0755); err != nil {
			return "", err
		}
	}

	return updateDir, nil
}

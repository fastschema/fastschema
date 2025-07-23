package schemaservice

import (
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/errors"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
	"github.com/otiai10/copy"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type SchemaUpdateData struct {
	Data         *schema.Schema   `json:"schema"`
	RenameFields []*db.RenameItem `json:"rename_fields"`
	RenameTables []*db.RenameItem `json:"rename_tables"`
}

type SchemaUpdate struct {
	currentSchemaBuilder *schema.Builder           // the current schema builder
	currentSchema        *schema.Schema            // the current schema to be updated. Get from the current schema builder
	updateData           *SchemaUpdateData         // data from request, including the new schema data and rename fields
	newSchemaBuilderDir  string                    // Create a new directory to store the new schema files
	newSchemaBuilder     *schema.Builder           // the new schema builder used to store the new schema files
	updateSchemas        map[string]*schema.Schema // schemas that need to be updated, includes: current schema, related schemas
	systemSchemas        []any                     // inherited system schemas, like User, Role, etc.
}

func (ss *SchemaService) Update(
	c fs.Context,
	updateData *SchemaUpdateData,
) (_ *schema.Schema, err error) {
	currentSchemaBuilderDir := ss.app.SchemaBuilder().Dir()
	su := &SchemaUpdate{
		updateData:           updateData,
		currentSchemaBuilder: ss.app.SchemaBuilder(),
		updateSchemas:        map[string]*schema.Schema{},
		systemSchemas:        ss.app.SystemSchemas(),
	}

	if su.currentSchema, err = su.currentSchemaBuilder.Schema(c.Arg("name")); err != nil {
		return nil, errors.NotFound(err.Error())
	}

	if su.updateData.Data == nil {
		return nil, errors.BadRequest("schema update data is required")
	}

	if err := su.update(); err != nil {
		return nil, err
	}

	if err = ss.app.Reload(c, &db.Migration{
		RenameTables: su.updateData.RenameTables,
		RenameFields: su.updateData.RenameFields,
	}); err != nil {
		// rollback
		// remove the current schema dir that contains the new schema files
		if e := os.RemoveAll(currentSchemaBuilderDir); e != nil {
			return nil, errors.InternalServerError(e.Error())
		}

		// rename the backup dir to the current schema dir
		// newSchemaBuilderDir is now holding the original schemas
		if e := os.Rename(su.newSchemaBuilderDir, currentSchemaBuilderDir); e != nil {
			return nil, errors.InternalServerError(e.Error())
		}

		if e := ss.app.Reload(c, nil); e != nil {
			return nil, errors.InternalServerError(e.Error())
		}

		return nil, errors.InternalServerError(err.Error())
	}

	return ss.app.SchemaBuilder().Schema(updateData.Data.Name)
}

func (su *SchemaUpdate) update() (err error) {
	if err := su.createNewSchemaDir(); err != nil {
		return errors.InternalServerError(err.Error())
	}

	// add the type and schema information to the rename fields
	su.updateData.RenameFields = utils.Map(su.updateData.RenameFields, func(rf *db.RenameItem) *db.RenameItem {
		rf.Type = "column"
		rf.SchemaName = su.currentSchema.Name
		rf.SchemaNamespace = su.currentSchema.Namespace
		return rf
	})

	if err := su.updateBasicData(); err != nil {
		return err
	}

	if err := su.updateRelatedSchemasBackRefs(); err != nil {
		return err
	}

	// Write the updated schema files to the new schema directory
	for _, s := range su.updateSchemas {
		schemaFile := path.Join(su.newSchemaBuilderDir, s.Name+".json")
		if err = s.SaveToFile(schemaFile); err != nil {
			return err
		}
	}

	su.newSchemaBuilder, err = schema.NewBuilderFromDir(su.newSchemaBuilderDir, su.systemSchemas...)
	if err != nil {
		return err
	}

	// Check rename relation fields
	for _, f := range su.updateData.Data.Fields {
		// if field exist in su.currentSchema,
		// then the current relation field is not a new relation field, no need to check for field rename.
		// otherwise, add the back reference field
		if !f.Type.IsRelationType() || su.currentSchema.HasField(f.Name) {
			continue
		}

		if err := su.applyRenameRelationField(f); err != nil {
			return err
		}
	}

	su.updateData.RenameFields = utils.Filter(su.updateData.RenameFields, func(rf *db.RenameItem) bool {
		return rf.From != rf.To
	})

	return su.renameDir()
}

// if the target schema is not existed in updateSchemas, add it.
func (su *SchemaUpdate) setUpdateRelationSchema(relation *schema.Relation) error {
	if _, ok := su.updateSchemas[relation.TargetSchemaName]; !ok {
		targetSchema, err := su.currentSchemaBuilder.Schema(relation.TargetSchemaName)
		if err != nil {
			return errors.BadRequest("relation target schema '%s' not found", relation.TargetSchemaName)
		}

		su.updateSchemas[relation.TargetSchemaName] = targetSchema.Clone() // clone the schema to avoid changing the original schema
	}

	return nil
}

// Add the back reference fields to the relation target schemas.
func (su *SchemaUpdate) applyAddNewRelationFields() error {
	// loop through all the fields in the new schema data
	// and check if the relation field is added
	for _, f := range su.updateData.Data.Fields {
		// if field exist in su.currentSchema,
		// then the current relation field is not a new relation field, no need to add the back reference field.
		// otherwise, add the back reference field
		if !f.Type.IsRelationType() || su.currentSchema.HasField(f.Name) {
			continue
		}

		// new relation field may be a file field
		// need to init the field so the relation will be initialized
		if err := f.Init(su.updateData.Data.Name); err != nil {
			return errors.InternalServerError("could not initialize field")
		}
		if err := su.setUpdateRelationSchema(f.Relation); err != nil {
			return err
		}

		// check if target schema already has the back reference field
		if su.updateSchemas[f.Relation.TargetSchemaName].HasField(f.Relation.TargetFieldName) {
			return errors.BadRequest(fmt.Sprintf(
				"Invalid field '%s.%s'. Target relation field '%s' already exist in schema '%s'",
				su.currentSchema.Name,
				f.Name,
				f.Relation.TargetFieldName,
				su.updateSchemas[f.Relation.TargetSchemaName].Name,
			))
		}

		// add the back reference field to the related schema in the new schema builder
		// if the f.Relation.Owner == false, then the back reference field will be the owner and be optional
		// otherwise, the back reference will be optional if the f.Relation.Optional == true
		backRefOptional := f.Relation.Optional
		if !f.Relation.Owner {
			backRefOptional = true
		}

		su.updateSchemas[f.Relation.TargetSchemaName].Fields = append(
			su.updateSchemas[f.Relation.TargetSchemaName].Fields,
			&schema.Field{
				Type:     schema.TypeRelation,
				Name:     f.Relation.TargetFieldName,
				Label:    cases.Title(language.Und, cases.NoLower).String(f.Relation.TargetFieldName),
				Optional: backRefOptional,
				Sortable: false,
				Relation: &schema.Relation{
					TargetSchemaName: su.currentSchema.Name,
					TargetFieldName:  f.Name,
					Type:             f.Relation.Type,
					Owner:            !f.Relation.Owner,
					Optional:         backRefOptional,
				},
			},
		)
	}

	return nil
}

func (su *SchemaUpdate) applyRenameRelationField(newField *schema.Field) (err error) {
	// filter from RenameFields to find the original field name (old field name)
	matchedFields := utils.Filter(su.updateData.RenameFields, func(f *db.RenameItem) bool {
		return f.To == newField.Name
	})

	// if the field is not renamed, then we don't need to rename the foreign key column
	if len(matchedFields) == 0 {
		return nil
	}

	originalField := su.currentSchema.Field(matchedFields[0].From)
	if originalField == nil {
		return schema.ErrFieldNotFound(su.currentSchema.Name, matchedFields[0].From)
	}

	defer func() {
		// RenameFields will be used to rename the db table columns,
		// remove the current relation field from the RenameFields because these fields are not exist in table,
		// we use the FK columns instead of these fields.
		su.updateData.RenameFields = utils.Filter(su.updateData.RenameFields, func(rf *db.RenameItem) bool {
			// with m2m relations, we have added a rename for two columns of the junction table,
			// the two columns of the junction table have the same name with the two fields of the two relation schemas.
			// so there will be two rename fields with the same From and To:
			// - 1st: The rename that was sent from the client,
			// 				this is the renaming for the relation field of the current editting schema,
			// 				this rename will have the SchemaName equal to the original schema name and must be removed.
			// - 2nd: The rename that was added for the junction table columns,
			// 				this rename will have the SchemaName equal to the junction table name and must be kept.
			// we have to remove the 1st rename because the relation field is not exist in the current editting schema table,
			isFiltered := (rf.SchemaName == su.currentSchema.Name && // check the 1st rename
				rf.From == originalField.Name && rf.To == newField.Name) // check From and To
			return !isFiltered
		})
	}()

	relationSchema, err := su.newSchemaBuilder.Schema(newField.Relation.TargetSchemaName)
	if err != nil {
		return err
	}

	newFieldRelation := newField.Relation.Clone()
	newFieldRelation.Init(su.currentSchema, relationSchema, newField)
	originalFieldRelation := originalField.Relation

	// process the m2m relation
	if newFieldRelation.Type == schema.M2M {
		originalJunctionSchema := originalFieldRelation.JunctionSchema
		newJunctionSchema, _, err := su.newSchemaBuilder.CreateM2mJunctionSchema(su.currentSchema, newFieldRelation)
		if err != nil {
			return err
		}

		// Ent do not perform the rename table operation, so we have to do it manually.
		// - Rename the junction table columns: Add the rename columns to the RenameFields so DiffHook can rename the columns.
		// - Rename the junction table: Add the rename table to the RenameTables and manually rename the table in ApplyHook.
		su.updateData.RenameTables = append(su.updateData.RenameTables, &db.RenameItem{
			Type:            "table",
			From:            originalJunctionSchema.Namespace,
			To:              newJunctionSchema.Namespace,
			IsJunctionTable: true,
		})

		su.updateData.RenameFields = append(su.updateData.RenameFields, &db.RenameItem{
			Type:            "column",
			From:            originalFieldRelation.FKColumns.CurrentColumn,
			To:              newFieldRelation.FKColumns.CurrentColumn,
			SchemaName:      originalJunctionSchema.Name,      // newJunctionSchema.Name
			SchemaNamespace: originalJunctionSchema.Namespace, // newJunctionSchema.Namespace
		})

		su.updateData.RenameFields = append(su.updateData.RenameFields, &db.RenameItem{
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

	su.updateData.RenameFields = append(su.updateData.RenameFields, &db.RenameItem{
		Type:            "column",
		From:            originalFieldRelation.GetTargetFKColumn(),
		To:              newFieldRelation.GetTargetFKColumn(),
		SchemaName:      su.currentSchema.Name,
		SchemaNamespace: su.currentSchema.Namespace,
	})

	return nil
}

// Remove the back reference field from the relation target schema.
func (su *SchemaUpdate) applyRemoveRelationFields() error {
	// loop through all the fields in the current schema
	// and check if the relation field is removed
	for _, f := range su.currentSchema.Fields {
		// if relation field exist in su.updateData.Data,
		// then the current relation field is not removed, no need to remove the back reference field.
		// otherwise, remove the back reference field
		if !f.Type.IsRelationType() || su.updateData.Data.HasField(f.Name) {
			continue
		}

		if err := su.setUpdateRelationSchema(f.Relation); err != nil {
			return err
		}

		// remove the back reference field from the relation target schema
		su.updateSchemas[f.Relation.TargetSchemaName].Fields = utils.Filter(
			su.updateSchemas[f.Relation.TargetSchemaName].Fields,
			func(targetField *schema.Field) bool {
				return targetField.Name != f.Relation.TargetFieldName
			},
		)
	}

	return nil
}

// When updating relation fields, there are two cases:
//   - Add new relation fields: need to add the back reference fields to the relation target schemas.
//   - Remove relation fields: need to remove the back reference fields from the relation target schemas.
func (su *SchemaUpdate) updateRelatedSchemasBackRefs() (err error) {
	if err := su.applyRemoveRelationFields(); err != nil {
		return err
	}

	if err := su.applyAddNewRelationFields(); err != nil {
		return err
	}

	return nil
}

// updateBasicData update the data that only exist in the schema json files.
func (su *SchemaUpdate) updateBasicData() (err error) {
	// overwrite the current schema with the new schema
	su.updateSchemas[su.updateData.Data.Name] = su.updateData.Data

	su.applyRenameSchemaNamespace()

	if err = su.applyRenameSchema(); err != nil {
		return err
	}

	return nil
}

// applyRenameSchema check if the schema name is renamed,
// if so, rename the schema name in the relation fields.
func (su *SchemaUpdate) applyRenameSchema() (err error) {
	newSchemaName := ""
	if su.currentSchema.Name != su.updateData.Data.Name {
		newSchemaName = su.updateData.Data.Name
		currentSchemaFile := path.Join(su.newSchemaBuilderDir, su.currentSchema.Name+".json")

		// if the name is changed:
		// 	- remove the current schema from the update schemas
		//	- delete the current schema json file
		delete(su.updateSchemas, su.currentSchema.Name)
		if err = os.Remove(currentSchemaFile); err != nil {
			return err
		}
	}

	if newSchemaName == "" {
		return nil
	}

	// loop the current schema fields to rename the bidi relation fields
	for i, f := range su.updateData.Data.Fields {
		if !f.Type.IsRelationType() {
			continue
		}

		// if TargetSchemaName is the current schema name, then it is bidi relation
		// rename the TargetSchemaName to the new schema name
		if f.Relation.TargetSchemaName == su.currentSchema.Name {
			su.updateData.Data.Fields[i].Relation.TargetSchemaName = newSchemaName
		}
	}

	// loop through all the schemas and check:
	// if schema contains fields that have relation to the current schema,
	// then rename the relation target schema name to the new schema name
	// skip checking current schema
	for _, s := range su.currentSchemaBuilder.Schemas() {
		if s.Name == su.currentSchema.Name {
			continue
		}

		for i, field := range s.Fields {
			if !field.Type.IsRelationType() {
				continue
			}

			// check if the relation target schema should be renamed
			if field.Relation.TargetSchemaName == su.currentSchema.Name {
				// if the schema is not existed in the update schemas, add it
				if _, ok := su.updateSchemas[s.Name]; !ok {
					su.updateSchemas[s.Name] = s.Clone() // clone the schema to avoid changing the original schema
				}

				// rename the relation target schema name
				su.updateSchemas[s.Name].Fields[i].Relation.TargetSchemaName = newSchemaName
			}
		}
	}

	return nil
}

// applyRenameSchemaNamespace check if the schema namespace is renamed,
// if so, rename the schema namespace in the relation fields.
func (su *SchemaUpdate) applyRenameSchemaNamespace() {
	newSchemaNamespace := ""
	if su.currentSchema.Namespace != su.updateData.Data.Namespace {
		newSchemaNamespace = su.updateData.Data.Namespace
	}

	if newSchemaNamespace == "" {
		return
	}

	su.updateData.RenameTables = append(su.updateData.RenameTables, &db.RenameItem{
		Type: "table",
		From: su.currentSchema.Namespace,
		To:   newSchemaNamespace,
	})
}

// copy the current schemas dir to a new backup dir.
// rename the new schema dir to the current schema dir to apply the new changes.
// rename the backup dir to the new schema dir to keep the backup.
// the new schema dir is now holding the backed up schemas (the original schemas).
func (su *SchemaUpdate) renameDir() error {
	currentSchemaDir := su.currentSchemaBuilder.Dir()
	backupDir := su.newSchemaBuilderDir + "_backup"

	if err := os.Rename(currentSchemaDir, backupDir); err != nil {
		return err
	}

	if err := os.Rename(su.newSchemaBuilderDir, currentSchemaDir); err != nil {
		return err
	}

	if err := os.Rename(backupDir, su.newSchemaBuilderDir); err != nil {
		return err
	}

	return nil
}

func (su *SchemaUpdate) createNewSchemaDir() error {
	schemasDir := su.currentSchemaBuilder.Dir()
	parentDir := path.Join(path.Dir(schemasDir), "backup")
	now := time.Now()
	schemaDirName := path.Base(schemasDir)
	backupDirName := fmt.Sprintf("%s_%s", schemaDirName, now.Format("2006_01_02_150405_.000000"))
	backupDirName = strings.ReplaceAll(backupDirName, ".", "")
	su.newSchemaBuilderDir = path.Join(parentDir, backupDirName)

	if err := os.MkdirAll(su.newSchemaBuilderDir, 0755); err != nil {
		return err
	}

	// Copy all files from the current schema directory to the new directory
	if err := copy.Copy(schemasDir, su.newSchemaBuilderDir); err != nil {
		return err
	}

	return nil
}

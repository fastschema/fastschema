package schemaservice

import (
	"os"

	"github.com/fastschema/fastschema/app"
	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/pkg/errors"
	"github.com/fastschema/fastschema/schema"
)

type SchemaUpdateData struct {
	Schema       *schema.Schema   `json:"schema"`
	RenameFields []*db.RenameItem `json:"rename_fields"`
	RenameTables []*db.RenameItem `json:"rename_tables"`
}

type SchemaUpdate struct {
	DB               func() db.Client
	updateDir        string
	updateSchemas    map[string]*schema.Schema
	newSchemaBuilder *schema.Builder
	oldSchema        *schema.Schema    // the original schema that get from schema builder
	newSchema        *schema.Schema    // the new schema that get from schema builder
	updateData       *SchemaUpdateData // data from request, including the new schema data and rename fields
}

func (ss *SchemaService) Update(c app.Context, updateData *SchemaUpdateData) (_ *schema.Schema, err error) {
	schemaName := c.Arg("name")
	oldSchemaDir := ss.app.SchemaBuilder().Dir()
	su := &SchemaUpdate{
		DB:               ss.app.DB,
		updateSchemas:    map[string]*schema.Schema{},
		newSchemaBuilder: nil,
		oldSchema:        nil,
		updateData:       updateData,
	}

	if err = su.prepare(schemaName); err != nil {
		return nil, errors.BadRequest(err.Error())
	}

	if err = ss.app.Reload(
		&db.Migration{
			RenameTables: su.updateData.RenameTables,
			RenameFields: su.updateData.RenameFields,
		},
	); err != nil {
		// rollback schema files
		if e := os.RemoveAll(oldSchemaDir); e != nil {
			return nil, errors.InternalServerError(e.Error())
		}

		if e := os.Rename(su.updateDir, oldSchemaDir); e != nil {
			return nil, errors.InternalServerError(e.Error())
		}

		if e := ss.app.Reload(nil); e != nil {
			return nil, errors.InternalServerError(e.Error())
		}

		return nil, errors.InternalServerError(err.Error())
	}

	return su.updateData.Schema, nil
}

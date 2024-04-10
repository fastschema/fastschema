package schemaservice

import (
	"os"

	"github.com/fastschema/fastschema/app"
	"github.com/fastschema/fastschema/pkg/errors"
	"github.com/fastschema/fastschema/schema"
)

type SchemaUpdateData struct {
	Schema       *schema.Schema    `json:"schema"`
	RenameFields []*app.RenameItem `json:"rename_fields"`
	RenameTables []*app.RenameItem `json:"rename_tables"`
}

type SchemaUpdate struct {
	DB               func() app.DBClient
	updateDir        string
	updateSchemas    map[string]*schema.Schema
	newSchemaBuilder *schema.Builder
	oldSchema        *schema.Schema    // the original schema that get from schema builder
	newSchema        *schema.Schema    // the new schema that get from schema builder
	updateData       *SchemaUpdateData // data from request, including the new schema data and rename fields
}

func (ss *SchemaService) Update(c app.Context, updateData *SchemaUpdateData) (_ *schema.Schema, err error) {
	oldSchemaDir := ss.app.SchemaBuilder().Dir()
	su := &SchemaUpdate{
		DB:               ss.app.DB,
		updateSchemas:    map[string]*schema.Schema{},
		newSchemaBuilder: nil,
		oldSchema:        nil,
		updateData:       updateData,
	}

	if su.oldSchema, err = su.DB().SchemaBuilder().Schema(c.Arg("name")); err != nil {
		return nil, errors.NotFound(err.Error())
	}

	if su.updateData.Schema == nil {
		return nil, errors.BadRequest("schema update data is required")
	}

	if err = su.prepare(); err != nil {
		return nil, errors.BadRequest(err.Error())
	}

	// catJSON := string(utils.Must(os.ReadFile(oldSchemaDir + "/category.json")))
	// blogJSON := string(utils.Must(os.ReadFile(oldSchemaDir + "/blog.json")))

	// fmt.Println(catJSON)
	// fmt.Println(blogJSON)

	if err = ss.app.Reload(
		&app.Migration{
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

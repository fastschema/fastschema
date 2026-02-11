package contentservice

import (
	"strings"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/entity"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/errors"
	"github.com/fastschema/fastschema/pkg/utils"
)

func (cs *ContentService) Detail(c fs.Context, _ any) (*entity.Entity, error) {
	schemaName := c.Arg("schema")
	model, err := cs.DB().Model(schemaName)
	if err != nil {
		return nil, errors.BadRequest(err.Error())
	}

	idValue, err := parseIDArg(model.Schema(), c.Arg("id"))
	if err != nil {
		return nil, errors.NotFound(err.Error())
	}

	columns := []string{}
	if fields := c.Arg("select", ""); fields != "" {
		columns = strings.Split(fields, ",")
	}

	// Parse relation options for controlling relation field loading
	relationOptions, err := db.ParseRelationOptions(c.Arg("select_options", ""))
	if err != nil {
		return nil, errors.BadRequest(err.Error())
	}

	query := model.
		Query(db.EQ(model.Schema().PrimaryKeyName(), idValue)).
		Select(columns...)

	// Apply relation options if provided
	if relationOptions != nil {
		query = query.WithRelationOptions(relationOptions)
	}

	entity, err := query.First(c)
	if err != nil {
		e := utils.If(db.IsNotFound(err), errors.NotFound, errors.InternalServerError)
		return nil, e(err.Error())
	}

	if schemaName == "user" {
		entity.Delete("password")
	}

	return entity, nil
}

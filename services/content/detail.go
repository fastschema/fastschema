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
	id := c.ArgInt("id")
	schemaName := c.Arg("schema")
	model, err := cs.DB().Model(schemaName)
	if err != nil {
		return nil, errors.BadRequest(err.Error())
	}

	columns := []string{}
	if fields := c.Arg("select", ""); fields != "" {
		columns = strings.Split(fields, ",")
	}

	entity, err := model.Query(db.EQ("id", id)).
		Select(columns...).
		First(c)
	if err != nil {
		e := utils.If(db.IsNotFound(err), errors.NotFound, errors.InternalServerError)
		return nil, e(err.Error())
	}

	if schemaName == "user" {
		entity.Delete("password")
	}

	return entity, nil
}

package contentservice

import (
	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/errors"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
)

func (cs *ContentService) Delete(c fs.Context, _ any) (any, error) {
	model, err := cs.DB().Model(c.Arg("schema"))
	if err != nil {
		return nil, errors.BadRequest(err.Error())
	}

	id := c.ArgInt("id")

	_, err = model.Query(db.EQ("id", id)).Only(c.Context())

	if err != nil {
		e := utils.If(db.IsNotFound(err), errors.NotFound, errors.InternalServerError)
		return nil, e(err.Error())
	}

	if _, err := model.Mutation().Where(db.EQ("id", id)).Delete(c.Context()); err != nil {
		return nil, errors.BadRequest(err.Error())
	}

	return schema.NewEntity(uint64(id)), nil
}

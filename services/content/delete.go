package contentservice

import (
	"github.com/fastschema/fastschema/app"
	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/pkg/errors"
)

func (cs *ContentService) Delete(c app.Context, _ *any) (any, error) {
	model, err := cs.app.DB().Model(c.Arg("schema"))
	if err != nil {
		return nil, errors.BadRequest(err.Error())
	}

	id := c.ArgInt("id")

	_, err = model.Query(db.EQ("id", id)).Only()
	if err != nil {
		return nil, errors.BadRequest(err.Error())
	}

	mutation, err := model.Mutation()
	if err != nil {
		return nil, errors.BadRequest(err.Error())
	}

	if _, err := mutation.Where(db.EQ("id", id)).Delete(); err != nil {
		return nil, errors.BadRequest(err.Error())
	}

	return nil, nil
}

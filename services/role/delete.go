package roleservice

import (
	"github.com/fastschema/fastschema/app"
	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/pkg/errors"
)

func (rs *RoleService) Delete(c app.Context, _ *any) (any, error) {
	id := c.ArgInt("id")

	if id <= 3 {
		return nil, errors.BadRequest("Can't delete default roles")
	}

	model, err := rs.app.DB().Model("role")
	if err != nil {
		return nil, errors.InternalServerError(err.Error())
	}

	mutation, err := model.Mutation()
	if err != nil {
		return nil, errors.InternalServerError(err.Error())
	}

	if _, err := mutation.Where(db.EQ("id", id)).Delete(); err != nil {
		return nil, errors.InternalServerError(err.Error())
	}

	return nil, nil
}

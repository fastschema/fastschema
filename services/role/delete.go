package roleservice

import (
	"github.com/fastschema/fastschema/app"
	"github.com/fastschema/fastschema/pkg/errors"
	"github.com/fastschema/fastschema/pkg/utils"
)

func (rs *RoleService) Delete(c app.Context, _ *any) (any, error) {
	id := c.ArgInt("id")

	if id <= 3 {
		return nil, errors.BadRequest("Can't delete default roles")
	}

	model, err := rs.DB().Model("role")
	if err != nil {
		return nil, errors.InternalServerError(err.Error())
	}

	if _, err := model.Query(app.EQ("id", id)).First(c.Context()); err != nil {
		e := utils.If(app.IsNotFound(err), errors.NotFound, errors.InternalServerError)
		return nil, e(err.Error())
	}

	if _, err := model.Mutation().Where(app.EQ("id", id)).Delete(); err != nil {
		return nil, errors.InternalServerError(err.Error())
	}

	return nil, nil
}

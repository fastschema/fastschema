package roleservice

import (
	"github.com/fastschema/fastschema/app"
	"github.com/fastschema/fastschema/pkg/errors"
	"github.com/fastschema/fastschema/schema"
)

func (rs *RoleService) List(c app.Context, _ any) ([]*schema.Entity, error) {
	model, err := rs.DB().Model("role")
	if err != nil {
		return nil, errors.BadRequest(err.Error())
	}

	roles, err := model.Query().Get(c.Context())
	if err != nil {
		return nil, errors.InternalServerError(err.Error())
	}

	return roles, nil
}

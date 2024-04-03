package roleservice

import (
	"github.com/fastschema/fastschema/app"
	"github.com/fastschema/fastschema/pkg/errors"
)

func (rs *RoleService) Create(c app.Context, _ *any) (*app.Role, error) {
	entity, err := c.Entity()
	if err != nil {
		return nil, errors.BadRequest(err.Error())
	}

	model, err := rs.app.DB().Model("role")
	if err != nil {
		return nil, errors.InternalServerError(err.Error())
	}

	id, err := model.Create(entity)
	if err != nil {
		return nil, errors.BadRequest(err.Error())
	}

	if err := entity.SetID(id); err != nil {
		return nil, errors.BadRequest(err.Error())
	}

	return app.EntityToRole(entity), nil
}

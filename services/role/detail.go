package roleservice

import (
	"github.com/fastschema/fastschema/app"
	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/pkg/errors"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
)

func (rs *RoleService) Detail(c app.Context, _ *any) (*schema.Entity, error) {
	roleID := c.ArgInt("id")
	model, err := rs.app.DB().Model("role")
	if err != nil {
		return nil, errors.BadRequest(err.Error())
	}

	role, err := model.Query().
		Where(db.EQ("id", roleID)).
		Select("permissions", "users.id", "users.username").
		First(c.Context())
	if err != nil {
		e := utils.If(db.IsNotFound(err), errors.NotFound, errors.InternalServerError)
		return nil, e(err.Error())
	}

	return role, nil
}

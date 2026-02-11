package roleservice

import (
	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/errors"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/google/uuid"
)

func (rs *RoleService) Detail(c fs.Context, _ any) (*fs.Role, error) {
	roleID, err := uuid.Parse(c.Arg("id"))
	if err != nil {
		return nil, errors.BadRequest("Invalid role ID")
	}
	role, err := db.Builder[*fs.Role](rs.DB()).Where(db.EQ("id", roleID)).
		Select("permissions").
		First(c)
	if err != nil {
		e := utils.If(db.IsNotFound(err), errors.NotFound, errors.InternalServerError)
		return nil, e(err.Error())
	}

	return role, nil
}

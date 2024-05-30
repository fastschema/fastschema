package roleservice

import (
	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/errors"
	"github.com/fastschema/fastschema/pkg/utils"
)

func (rs *RoleService) Detail(c fs.Context, _ any) (*fs.Role, error) {
	roleID := c.ArgInt("id")
	role, err := db.Query[*fs.Role](rs.DB()).Where(db.EQ("id", roleID)).First(c.Context())
	if err != nil {
		e := utils.If(db.IsNotFound(err), errors.NotFound, errors.InternalServerError)
		return nil, e(err.Error())
	}

	return role, nil
}

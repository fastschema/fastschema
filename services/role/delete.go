package roleservice

import (
	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/errors"
	"github.com/fastschema/fastschema/pkg/utils"
)

func (rs *RoleService) Delete(c fs.Context, _ any) (any, error) {
	id := c.ArgInt("id")

	if id <= 3 {
		return nil, errors.BadRequest("Can't delete default roles")
	}

	where := db.EQ("id", id)
	if _, err := db.Query[*fs.Role](rs.DB()).Where(where).First(c.Context()); err != nil {
		e := utils.If(db.IsNotFound(err), errors.NotFound, errors.InternalServerError)
		return nil, e(err.Error())
	}

	return db.Delete[*fs.Role](c.Context(), rs.DB(), where)
}

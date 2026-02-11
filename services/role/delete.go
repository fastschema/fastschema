package roleservice

import (
	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/errors"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/google/uuid"
)

func (rs *RoleService) Delete(c fs.Context, _ any) (any, error) {
	id, err := uuid.Parse(c.Arg("id"))
	if err != nil {
		return nil, errors.BadRequest("Invalid role ID")
	}

	// Fetch the role first to check if it's a system role
	role, err := db.Builder[*fs.Role](rs.DB()).Where(db.EQ("id", id)).First(c)
	if err != nil {
		e := utils.If(db.IsNotFound(err), errors.NotFound, errors.InternalServerError)
		return nil, e(err.Error())
	}

	if role.System {
		return nil, errors.BadRequest("Can't delete default roles")
	}

	conditions := []*db.Predicate{
		db.EQ("id", id),
	}

	return db.Delete[*fs.Role](c, rs.DB(), conditions)
}

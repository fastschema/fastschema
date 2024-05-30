package roleservice

import (
	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/errors"
)

func (rs *RoleService) Create(c fs.Context, _ *fs.Role) (*fs.Role, error) {
	entity, err := c.Entity()
	if err != nil {
		return nil, errors.BadRequest(err.Error())
	}

	return db.Create[*fs.Role](c.Context(), rs.DB(), entity)
}

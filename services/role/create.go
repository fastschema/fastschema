package roleservice

import (
	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/errors"
)

func (rs *RoleService) Create(c fs.Context, _ any) (*fs.Role, error) {
	entity, err := c.Payload()
	if err != nil {
		return nil, errors.BadRequest(err.Error())
	}

	return db.Create[*fs.Role](c, rs.DB(), entity)
}

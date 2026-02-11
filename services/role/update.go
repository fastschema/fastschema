package roleservice

import (
	"context"
	"fmt"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/entity"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/errors"
	"github.com/fastschema/fastschema/pkg/utils"
)

func (rs *RoleService) Update(c fs.Context, _ any) (_ *fs.Role, err error) {
	id := c.ArgInt("id")
	updateRoleData, err := c.Payload()
	if err != nil {
		return nil, errors.BadRequest(err.Error())
	}

	return rs.UpdateRole(c, uint64(id), updateRoleData)
}

func (rs *RoleService) UpdateRole(
	c context.Context,
	id uint64,
	updateRoleData *entity.Entity,
) (_ *fs.Role, err error) {
	tx, err := rs.DB().Tx(c)
	if err != nil {
		return nil, errors.InternalServerError(err.Error())
	}

	defer func() {
		if err != nil {
			err = fmt.Errorf("role update error: %w, rollback error: %w", err, tx.Rollback())
			return
		}

		if err = tx.Commit(); err != nil {
			err = fmt.Errorf("commit error: %w, rollback error: %w", err, tx.Rollback())
			err = errors.InternalServerError(err.Error())
			return
		}

		err = rs.UpdateCache(c)
	}()

	if err := updateRoleData.SetID(id); err != nil {
		return nil, errors.BadRequest(err.Error())
	}
	existingRole, err := db.Builder[*fs.Role](tx).
		Where(db.EQ("id", id)).
		Select("permissions").
		First(c)
	if err != nil {
		e := utils.If(db.IsNotFound(err), errors.NotFound, errors.InternalServerError)
		return nil, e(err.Error())
	}

	if err := updateRolePermissions(
		c,
		existingRole,
		updateRoleData,
		tx,
	); err != nil {
		return nil, err
	}

	checkRole := &fs.Role{Rule: updateRoleData.GetString("rule", "")}
	if err := checkRole.Compile(); err != nil {
		return nil, errors.InternalServerError(err.Error())
	}

	if _, err := db.Update[*fs.Role](c, tx, updateRoleData, []*db.Predicate{db.EQ("id", id)}); err != nil {
		return nil, errors.InternalServerError(err.Error())
	}

	return db.Builder[*fs.Role](tx).
		Where(db.EQ("id", id)).
		Select("permissions").
		First(c)
}

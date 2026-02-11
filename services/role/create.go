package roleservice

import (
	"fmt"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/entity"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/errors"
	"github.com/fastschema/fastschema/pkg/utils"
)

func (rs *RoleService) Create(c fs.Context, _ any) (*fs.Role, error) {
	payload, err := c.Payload()
	if err != nil {
		return nil, errors.BadRequest(err.Error())
	}

	rolePermissions := []*entity.Entity{}
	var ok bool
	if rp := payload.Get("permissions"); rp != nil {
		if rolePermissions, ok = rp.([]*entity.Entity); !ok {
			return nil, errors.BadRequest("permissions must be an array of entities")
		}
	}

	tx, err := rs.DB().Tx(c)
	if err != nil {
		return nil, errors.InternalServerError(err.Error())
	}

	createRoleData := payload.Delete("permissions")
	createdRole, err := db.Create[*fs.Role](c, tx, createRoleData)
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

	updateRoleData := entity.New(createdRole.ID).Set("permissions", rolePermissions)
	existingRole, err := db.Builder[*fs.Role](tx).
		Where(db.EQ("id", createdRole.ID)).
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

	if _, err := db.Update[*fs.Role](c, tx, updateRoleData, []*db.Predicate{db.EQ("id", createdRole.ID)}); err != nil {
		return nil, errors.InternalServerError(err.Error())
	}

	return db.Builder[*fs.Role](tx).
		Where(db.EQ("id", createdRole.ID)).
		Select("permissions").
		First(c)
}

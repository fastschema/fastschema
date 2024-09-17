package roleservice

import (
	"context"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/errors"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
)

func (rs *RoleService) Update(c fs.Context, _ any) (_ *fs.Role, err error) {
	tx, err := rs.DB().Tx(c)
	if err != nil {
		return nil, errors.InternalServerError(err.Error())
	}

	defer func() {
		if err != nil {
			rollback(tx, c)
			return
		}

		if err := tx.Commit(); err != nil {
			rollback(tx, c)
			err = errors.InternalServerError(err.Error())
			return
		}

		if err := rs.UpdateCache(c); err != nil {
			c.Logger().Error(err.Error())
		}
	}()

	id := c.ArgInt("id")
	updateRoleData, err := c.Payload()
	if err != nil {
		return nil, errors.BadRequest(err.Error())
	}

	updateRoleData.SetID(id)
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

	if _, err := db.Update[*fs.Role](c, tx, updateRoleData, []*db.Predicate{db.EQ("id", id)}); err != nil {
		return nil, errors.InternalServerError(err.Error())
	}

	return db.Builder[*fs.Role](tx).
		Where(db.EQ("id", id)).
		Select("permissions").
		First(c)
}

func updateRolePermissions(
	ctx context.Context,
	existingRole *fs.Role,
	updateRoleData *schema.Entity,
	tx db.Client,
) error {
	currentPermissions := []string{}
	for _, permission := range existingRole.Permissions {
		currentPermissions = append(currentPermissions, permission.Resource)
	}

	added, removed, err := getPermissionsUpdate(currentPermissions, updateRoleData)
	if err != nil {
		return err
	}

	for _, permissionName := range added {
		permissionEntity := schema.NewEntity().
			Set("resource", permissionName).
			Set("value", "allow").
			Set("role_id", existingRole.ID)

		if _, err := db.Create[*fs.Permission](ctx, tx, permissionEntity); err != nil {
			return errors.InternalServerError(err.Error())
		}
	}

	for _, permissionName := range removed {
		if _, err := db.Delete[*fs.Permission](ctx, tx, []*db.Predicate{db.And(
			db.EQ("role_id", existingRole.ID),
			db.EQ("resource", permissionName),
		)}); err != nil {
			return errors.InternalServerError(err.Error())
		}
	}

	return nil
}

func getPermissionsUpdate(currentRolePermissions []string, updateRoleData *schema.Entity) ([]string, []string, error) {
	permissionValue := updateRoleData.Get("permissions", []string{})
	updateRoleData.Delete("permissions")
	addedPermissions := []string{}
	removedPermissions := []string{}
	updatePermissions, _ := permissionValue.([]string)
	updatePermissionsAny, _ := permissionValue.([]any)

	if len(updatePermissions) == 0 && len(updatePermissionsAny) > 0 {
		for _, permission := range updatePermissionsAny {
			permissionName, ok := permission.(string)
			if !ok {
				return nil, nil, errors.BadRequest("permission must be a string")
			}

			updatePermissions = append(updatePermissions, permissionName)
		}
	}

	for _, permission := range updatePermissions {
		if !utils.Contains(currentRolePermissions, permission) {
			addedPermissions = append(addedPermissions, permission)
		}
	}

	for _, permission := range currentRolePermissions {
		if !utils.Contains(updatePermissions, permission) {
			removedPermissions = append(removedPermissions, permission)
		}
	}

	return addedPermissions, removedPermissions, nil
}

func rollback(tx db.Client, c fs.Context) {
	if err := tx.Rollback(); err != nil {
		c.Logger().Error(err.Error())
	}
}

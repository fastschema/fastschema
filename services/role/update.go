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

	updateRole := &fs.Role{Rule: updateRoleData.GetString("rule", "")}
	if err := updateRole.Compile(); err != nil {
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

func updateRolePermissions(
	ctx context.Context,
	existingRole *fs.Role,
	updateRoleData *entity.Entity,
	tx db.Client,
) error {
	currentPermissionsNames := utils.Map(existingRole.Permissions, func(p *fs.Permission) string {
		return p.Resource
	})

	updated, added, removed, err := getPermissionsUpdate(currentPermissionsNames, updateRoleData)
	if err != nil {
		return err
	}

	// Update permissions
	for _, permission := range updated {
		permissionEntity := entity.New().Set("value", permission.Value)
		if _, err := db.Update[*fs.Permission](ctx, tx, permissionEntity, []*db.Predicate{
			db.EQ("role_id", existingRole.ID),
			db.EQ("resource", permission.Resource),
		}); err != nil {
			return errors.InternalServerError(err.Error())
		}
	}

	// Create new permissions
	for _, permission := range added {
		permissionEntity := entity.New().
			Set("resource", permission.Resource).
			Set("value", permission.Value).
			Set("role_id", existingRole.ID)

		if _, err := db.Create[*fs.Permission](ctx, tx, permissionEntity); err != nil {
			return errors.InternalServerError(err.Error())
		}
	}

	// Remove permissions
	removedPermissionsNames := utils.Map(removed, func(p string) any {
		return p
	})
	if _, err := db.Delete[*fs.Permission](ctx, tx, []*db.Predicate{db.And(
		db.EQ("role_id", existingRole.ID),
		db.In("resource", removedPermissionsNames),
	)}); err != nil {
		return errors.InternalServerError(err.Error())
	}

	return nil
}

func getPermissionsUpdate(
	currentRolePermissions []string,
	updateRoleData *entity.Entity,
) ([]*fs.Permission, []*fs.Permission, []string, error) {
	permissionsListData, exists := updateRoleData.Data().Get("permissions")
	if !exists {
		return nil, nil, nil, nil
	}

	updateRoleData.Delete("permissions")
	permissionsList, _ := permissionsListData.([]*fs.Permission)
	permissionsListEntities, _ := permissionsListData.([]*entity.Entity)

	updatePermissions := []*fs.Permission{}
	addedPermissions := []*fs.Permission{}
	removedPermissions := []string{}

	if len(permissionsList) == 0 && len(permissionsListEntities) > 0 {
		for _, permission := range permissionsListEntities {
			value := permission.GetString("value", "allow")
			value = utils.If(value == "", "allow", value)
			permissionsList = append(permissionsList, &fs.Permission{
				Resource: permission.GetString("resource"),
				Value:    value,
				Modifier: permission.GetString("modifier", ""),
			})
		}
	}

	permissionsNames := utils.Map(permissionsList, func(p *fs.Permission) string {
		return p.Resource
	})

	for _, permission := range permissionsList {
		if err := permission.Compile(); err != nil {
			return nil, nil, nil, errors.InternalServerError(
				fmt.Sprintf(
					"error compiling permission rule for %s: %s",
					permission.Resource,
					err.Error(),
				),
			)
		}

		if !utils.Contains(currentRolePermissions, permission.Resource) {
			addedPermissions = append(addedPermissions, permission)
		} else {
			updatePermissions = append(updatePermissions, permission)
		}
	}

	for _, permission := range currentRolePermissions {
		if !utils.Contains(permissionsNames, permission) {
			removedPermissions = append(removedPermissions, permission)
		}
	}

	return updatePermissions, addedPermissions, removedPermissions, nil
}

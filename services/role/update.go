package roleservice

import (
	"github.com/fastschema/fastschema/app"
	"github.com/fastschema/fastschema/pkg/errors"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
)

func (rs *RoleService) Update(c app.Context, _ any) (_ *app.Role, err error) {
	tx, err := rs.DB().Tx(c.Context())
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

		if err := rs.UpdateCache(); err != nil {
			c.Logger().Error(err.Error())
		}
	}()

	id := c.ArgInt("id")
	updateRoleData, err := c.Entity()
	if err != nil {
		return nil, errors.BadRequest(err.Error())
	}

	updateRoleData.SetID(id)
	roleModel, err := tx.Model("role")
	if err != nil {
		return nil, errors.InternalServerError(err.Error())
	}

	existingRole, err := roleModel.Query().
		Where(app.EQ("id", id)).
		Select("permissions").
		First()
	if err != nil {
		e := utils.If(app.IsNotFound(err), errors.NotFound, errors.InternalServerError)
		return nil, e(err.Error())
	}

	if err := updateRolePermissions(
		app.EntityToRole(existingRole),
		updateRoleData,
		tx,
	); err != nil {
		return nil, err
	}

	if _, err := roleModel.Mutation().Where(app.EQ("id", id)).Update(updateRoleData); err != nil {
		return nil, errors.InternalServerError(err.Error())
	}

	return app.EntityToRole(updateRoleData), nil
}

func updateRolePermissions(existingRole *app.Role, updateRoleData *schema.Entity, tx app.DBClient) error {
	currentPermissions := []string{}
	for _, permission := range existingRole.Permissions {
		currentPermissions = append(currentPermissions, permission.Resource)
	}

	added, removed, err := getPermissionsUpdate(currentPermissions, updateRoleData)
	if err != nil {
		return err
	}

	permissionModel, err := tx.Model("permission")
	if err != nil {
		return errors.InternalServerError(err.Error())
	}

	for _, permissionName := range added {
		permissionEntity := schema.NewEntity().
			Set("resource", permissionName).
			Set("value", "allow").
			Set("role_id", existingRole.ID)

		if _, err := permissionModel.Create(permissionEntity); err != nil {
			return errors.InternalServerError(err.Error())
		}
	}

	for _, permissionName := range removed {
		if _, err := permissionModel.Mutation().Where(app.And(
			app.EQ("role_id", existingRole.ID),
			app.EQ("resource", permissionName),
		)).Delete(); err != nil {
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

func rollback(tx app.DBClient, c app.Context) {
	if err := tx.Rollback(); err != nil {
		c.Logger().Error(err.Error())
	}
}

package toolservice

import (
	"context"
	"fmt"

	"github.com/fastschema/fastschema/app"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
)

func CreateRole(db app.DBClient, roleData *app.Role) (uint64, error) {
	roleModel, err := db.Model("role")
	if err != nil {
		return 0, err
	}

	role := schema.NewEntity().
		Set("name", roleData.Name).
		Set("description", roleData.Description).
		Set("root", roleData.Root)

	return roleModel.Create(role)
}

func Setup(
	dbClient app.DBClient,
	logger app.Logger,
	username, email, password string,
) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}
	}()

	tx := utils.Must(dbClient.Tx(context.Background()))

	defer func() {
		if err != nil {
			if err := tx.Rollback(); err != nil {
				logger.Error("rollback error: %v", err)
			}
		}
	}()

	userModel := utils.Must(tx.Model("user"))
	adminUser, err := userModel.Query(app.EQ("username", username)).First()
	if err != nil && !app.IsNotFound(err) {
		return err
	}

	if adminUser != nil {
		return fmt.Errorf("user %s already exists", username)
	}

	adminRoleID := utils.Must(CreateRole(tx, app.RoleAdmin))
	utils.Must(CreateRole(tx, app.RoleUser))
	utils.Must(CreateRole(tx, app.RoleGuest))
	adminPassword, err := utils.GenerateHash(password)
	if err != nil {
		return err
	}

	_, err = userModel.Create(schema.NewEntityFromMap(map[string]any{
		"username": username,
		"email":    email,
		"password": adminPassword,
		"active":   true,
		"roles": []*schema.Entity{
			schema.NewEntity(adminRoleID),
		},
	}))

	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	logger.Info("Setup root user successfully")

	return nil
}

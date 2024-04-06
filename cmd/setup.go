package cmd

import (
	"context"
	"fmt"

	"github.com/fastschema/fastschema/app"
	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/logger"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
)

func createRole(db db.Client, roleData *app.Role) (uint64, error) {
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
	dbClient db.Client,
	logger logger.Logger,
	username, email, password string,
) error {
	tx := utils.Must(dbClient.Tx(context.Background()))
	userModel := utils.Must(tx.Model("user"))
	adminUser, err := userModel.Query(db.EQ("username", username)).First()
	if err != nil && !db.IsNotFound(err) {
		return err
	}

	if adminUser != nil {
		return fmt.Errorf("user %s already exists", username)
	}

	adminRoleID := utils.Must(createRole(tx, app.RoleAdmin))
	utils.Must(createRole(tx, app.RoleUser))
	utils.Must(createRole(tx, app.RoleGuest))
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

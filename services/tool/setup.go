package toolservice

import (
	"context"
	"fmt"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/logger"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
)

func Setup(
	ctx context.Context,
	dbClient db.Client,
	logger logger.Logger,
	username, email, password string,
) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}
	}()

	tx := utils.Must(dbClient.Tx(ctx))

	defer func() {
		if err != nil {
			logger.Error(err)
			if err := tx.Rollback(); err != nil {
				logger.Error("rollback error: %v", err)
			}
		}
	}()

	adminUser, err := db.Query[*fs.User](tx).Where(db.EQ("username", username)).First(ctx)
	if err != nil {
		if db.IsNotFound(err) {
			err = nil
		} else {
			return err
		}
	}

	if adminUser != nil {
		return fmt.Errorf("user %s already exists", username)
	}

	adminRole := utils.Must(CreateRole(ctx, tx, fs.RoleAdmin))
	utils.Must(CreateRole(ctx, tx, fs.RoleUser))
	utils.Must(CreateRole(ctx, tx, fs.RoleGuest))
	adminPassword, err := utils.GenerateHash(password)
	if err != nil {
		return err
	}

	utils.Must(db.Create[*fs.User](ctx, tx, fs.Map{
		"username": username,
		"email":    email,
		"password": adminPassword,
		"active":   true,
		"roles":    []*schema.Entity{schema.NewEntity(adminRole.ID)},
	}))

	if err := tx.Commit(); err != nil {
		return err
	}

	logger.Info("Setup root user successfully")

	return nil
}

func CreateRole(ctx context.Context, dbc db.Client, roleData *fs.Role) (*fs.Role, error) {
	return db.Create[*fs.Role](ctx, dbc, schema.NewEntity().
		Set("name", roleData.Name).
		Set("description", roleData.Description).
		Set("root", roleData.Root))
}

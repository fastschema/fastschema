package toolservice

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/entity"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/logger"
	"github.com/fastschema/fastschema/pkg/utils"
)

func Setup(
	ctx context.Context,
	dbClient db.Client,
	logger logger.Logger,
	username, email, password string,
) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v\n%s", r, string(debug.Stack()))
		}
	}()

	tx := utils.Must(dbClient.Tx(ctx))

	defer func() {
		if err != nil {
			logger.Error(err)
			if err := tx.Rollback(); err != nil {
				logger.Errorf("rollback error: %w", err)
			}
		}
	}()

	adminUser, err := db.Builder[*fs.User](tx).Where(db.EQ("username", username)).First(ctx)
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

	utils.Must(db.Create[*fs.User](ctx, tx, fs.Map{
		"username": username,
		"email":    email,
		"password": password,
		"provider": "local",
		"active":   true,
		"roles":    []*entity.Entity{entity.New(adminRole.ID)},
	}))

	if err := tx.Commit(); err != nil {
		return err
	}

	logger.Info("Setup root user successfully")

	return nil
}

func CreateRole(ctx context.Context, dbc db.Client, roleData *fs.Role) (*fs.Role, error) {
	return db.Create[*fs.Role](ctx, dbc, entity.New().
		Set("name", roleData.Name).
		Set("description", roleData.Description).
		Set("root", roleData.Root))
}

func ResetAdminPassword(ctx context.Context,
	dbClient db.Client,
	password string,
	id int,
) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v\n%s", r, string(debug.Stack()))
		}
	}()

	tx := utils.Must(dbClient.Tx(ctx))

	defer func() {
		if err != nil {
			fmt.Println(err)
			if err := tx.Rollback(); err != nil {
				_ = fmt.Errorf("rollback error: %w", err)
			}
		}
	}()

	if password == "" {
		return errors.New("password cannot be empty")
	}

	admin, err := db.Builder[*fs.User](tx).
		Where(db.EQ("id", id)).
		Select("id", "roles").
		First(ctx)
	if err != nil {
		if db.IsNotFound(err) {
			err = nil
		} else {
			return err
		}
	}

	if admin == nil {
		return errors.New("cannot find admin user. Please setup the app first")
	}

	if !admin.IsRoot() {
		return errors.New("user is not an admin")
	}

	utils.Must(db.Update[*fs.User](ctx, tx, fs.Map{
		"password": password,
	}, []*db.Predicate{db.EQ("id", id)}))

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("%s", err.Error())
	}

	fmt.Println("Reset admin password successfully")

	return nil
}

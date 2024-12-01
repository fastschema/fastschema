package toolservice_test

import (
	"context"
	"os"
	"testing"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/entity"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/logger"
	"github.com/fastschema/fastschema/pkg/entdbadapter"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
	toolservice "github.com/fastschema/fastschema/services/tool"
	"github.com/stretchr/testify/assert"
)

func TestCreateRoleError(t *testing.T) {
	sb := &schema.Builder{}
	db := utils.Must(entdbadapter.NewTestClient(utils.Must(os.MkdirTemp("", "migrations")), sb))
	_, err := toolservice.CreateRole(context.Background(), db, &fs.Role{
		Name:        "admin",
		Description: "admin",
		Root:        true,
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), `model Role not found`)
}

func TestSetup(t *testing.T) {
	logger := logger.CreateMockLogger(true)
	// Case 1: Error when model user not found
	sb := &schema.Builder{}
	entDB := utils.Must(entdbadapter.NewTestClient(utils.Must(os.MkdirTemp("", "migrations")), sb))
	err := toolservice.Setup(
		context.Background(),
		entDB,
		logger,
		"admin",
		"admin@local.ltd",
		"123",
	)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), `model User not found`)

	sb = utils.Must(schema.NewBuilderFromDir(t.TempDir(), fs.SystemSchemaTypes...))
	entDB = utils.Must(entdbadapter.NewTestClient(utils.Must(os.MkdirTemp("", "migrations")), sb))

	// Case 2: Success
	assert.NoError(t, toolservice.Setup(
		context.Background(),
		entDB,
		logger,
		"admin",
		"admin@local.ltd",
		"123",
	))

	roleModel := utils.Must(entDB.Model("role"))
	roleCount := utils.Must(roleModel.Query().Count(context.Background(), &db.QueryOption{
		Unique: true,
		Column: "id",
	}))

	assert.Equal(t, 3, roleCount)
	userModel := utils.Must(entDB.Model("user"))
	ctx := context.WithValue(context.Background(), "keeppassword", "true")
	adminUser := utils.Must(userModel.Query(db.EQ("username", "admin")).First(ctx))
	assert.Equal(t, "admin", adminUser.Get("username"))
	assert.Equal(t, "admin@local.ltd", adminUser.Get("email"))

	checkPassword := utils.CheckHash("123", adminUser.GetString("password"))
	assert.NoError(t, checkPassword)

	// Case 4: User already exists
	err = toolservice.Setup(
		context.Background(),
		entDB,
		logger,
		"admin",
		"admin@local.ltd",
		"123",
	)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), `user admin already exists`)
}

func TestResetAdminPasswordSuccess(t *testing.T) {
	logger := logger.CreateMockLogger(true)
	sb := utils.Must(schema.NewBuilderFromDir(t.TempDir(), fs.SystemSchemaTypes...))
	entDB := utils.Must(entdbadapter.NewTestClient(utils.Must(os.MkdirTemp("", "migrations")), sb))

	// Setup initial admin user
	err := toolservice.Setup(
		context.Background(),
		entDB,
		logger,
		"admin",
		"admin@local.ltd",
		"123",
	)
	assert.NoError(t, err)

	// Reset password
	err = toolservice.ResetAdminPassword(
		context.Background(),
		entDB,
		"newpassword",
		1,
	)
	assert.NoError(t, err)

	// Verify password change
	userModel := utils.Must(entDB.Model("user"))
	ctx := context.WithValue(context.Background(), "keeppassword", "true")
	adminUser := utils.Must(userModel.Query(db.EQ("id", 1)).First(ctx))
	checkPassword := utils.CheckHash("newpassword", adminUser.GetString("password"))
	assert.NoError(t, checkPassword)
}

func TestResetAdminPasswordEmptyPassword(t *testing.T) {
	logger := logger.CreateMockLogger(true)
	sb := utils.Must(schema.NewBuilderFromDir(t.TempDir(), fs.SystemSchemaTypes...))
	entDB := utils.Must(entdbadapter.NewTestClient(utils.Must(os.MkdirTemp("", "migrations")), sb))

	// Setup initial admin user
	err := toolservice.Setup(
		context.Background(),
		entDB,
		logger,
		"admin",
		"admin@local.ltd",
		"123",
	)
	assert.NoError(t, err)

	// Attempt to reset password with empty password
	err = toolservice.ResetAdminPassword(
		context.Background(),
		entDB,
		"",
		1,
	)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "password cannot be empty")
}

func TestResetAdminPasswordUserNotFound(t *testing.T) {
	logger.CreateMockLogger(true)
	sb := utils.Must(schema.NewBuilderFromDir(t.TempDir(), fs.SystemSchemaTypes...))
	entDB := utils.Must(entdbadapter.NewTestClient(utils.Must(os.MkdirTemp("", "migrations")), sb))

	// Attempt to reset password for non-existent user
	err := toolservice.ResetAdminPassword(
		context.Background(),
		entDB,
		"newpassword",
		999,
	)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot find admin user. Please setup the app first")
}

func TestResetAdminPasswordUserNotAdmin(t *testing.T) {
	logger.CreateMockLogger(true)
	sb := utils.Must(schema.NewBuilderFromDir(t.TempDir(), fs.SystemSchemaTypes...))
	entDB := utils.Must(entdbadapter.NewTestClient(utils.Must(os.MkdirTemp("", "migrations")), sb))

	// Setup initial non-admin user
	utils.Must(db.Create[*fs.User](context.Background(), entDB, fs.Map{
		"username": "user",
		"email":    "user@local.ltd",
		"provider": "local",
		"password": utils.Must(utils.GenerateHash("123")),
		"active":   true,
		"roles":    []*entity.Entity{},
	}))

	// Attempt to reset password for non-admin user
	err := toolservice.ResetAdminPassword(
		context.Background(),
		entDB,
		"newpassword",
		1,
	)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user is not an admin")
}

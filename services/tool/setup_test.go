package toolservice_test

import (
	"context"
	"os"
	"testing"

	"github.com/fastschema/fastschema/db"
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

	// Case 2: Invalid password
	err = toolservice.Setup(
		context.Background(),
		entDB,
		logger,
		"admin",
		"admin@local.ltd",
		"",
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), `hash: input cannot be empty`)

	// Case 3: Success
	assert.NoError(t, toolservice.Setup(
		context.Background(),
		entDB,
		logger,
		"admin",
		"admin@local.ltd",
		"123",
	))

	roleModel := utils.Must(entDB.Model("role"))
	roleCount := utils.Must(roleModel.Query().Count(context.Background(), &db.CountOption{
		Unique: true,
		Column: "id",
	}))

	assert.Equal(t, 3, roleCount)
	userModel := utils.Must(entDB.Model("user"))
	adminUser := utils.Must(userModel.Query(db.EQ("username", "admin")).First(context.Background()))
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

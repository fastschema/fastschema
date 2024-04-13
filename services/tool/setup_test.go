package toolservice_test

import (
	"testing"

	"github.com/fastschema/fastschema/app"
	"github.com/fastschema/fastschema/pkg/entdbadapter"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
	toolservice "github.com/fastschema/fastschema/services/tool"
	"github.com/stretchr/testify/assert"
)

func TestCreateRoleError(t *testing.T) {
	sb := &schema.Builder{}
	db := utils.Must(entdbadapter.NewTestClient(t.TempDir(), sb))
	_, err := toolservice.CreateRole(db, &app.Role{
		Name:        "admin",
		Description: "admin",
		Root:        true,
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), `model role not found`)
}

func TestSetup(t *testing.T) {
	logger := app.CreateMockLogger(true)
	// Case 1: Error when model user not found
	sb := &schema.Builder{}
	db := utils.Must(entdbadapter.NewTestClient(t.TempDir(), sb))
	err := toolservice.Setup(
		db,
		logger,
		"admin",
		"admin@local.ltd",
		"123",
	)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), `model user not found`)

	sb = utils.Must(schema.NewBuilderFromDir(t.TempDir()))
	db = utils.Must(entdbadapter.NewTestClient(t.TempDir(), sb))

	// Case 2: Invalid password
	err = toolservice.Setup(
		db,
		logger,
		"admin",
		"admin@local.ltd",
		"",
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), `hash: input cannot be empty`)

	// Case 3: Success
	assert.NoError(t, toolservice.Setup(
		db,
		logger,
		"admin",
		"admin@local.ltd",
		"123",
	))

	roleModel := utils.Must(db.Model("role"))
	roleCount := utils.Must(roleModel.Query().Count(&app.CountOption{
		Unique: true,
		Column: "id",
	}))

	assert.Equal(t, 3, roleCount)
	userModel := utils.Must(db.Model("user"))
	adminUser := utils.Must(userModel.Query(app.EQ("username", "admin")).First())
	assert.Equal(t, "admin", adminUser.Get("username"))
	assert.Equal(t, "admin@local.ltd", adminUser.Get("email"))

	checkPassword := utils.CheckHash("123", adminUser.GetString("password"))
	assert.NoError(t, checkPassword)

	// Case 4: User already exists
	err = toolservice.Setup(
		db,
		logger,
		"admin",
		"admin@local.ltd",
		"123",
	)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), `user admin already exists`)
}

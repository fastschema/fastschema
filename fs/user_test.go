package fs_test

import (
	"testing"

	"github.com/fastschema/fastschema/fs"
	"github.com/stretchr/testify/assert"
)

func TestUserIsRoot(t *testing.T) {
	// Test case 1: User is nil
	var u *fs.User
	isRoot := u.IsRoot()
	assert.False(t, isRoot, "Expected IsRoot to return false for nil user")

	// Test case 2: User has no roles
	u = &fs.User{}
	isRoot = u.IsRoot()
	assert.False(t, isRoot, "Expected IsRoot to return false for user with no roles")

	// Test case 3: User has roles but none are root
	u = &fs.User{
		Roles: []*fs.Role{
			{Root: false},
			{Root: false},
		},
	}
	isRoot = u.IsRoot()
	assert.False(t, isRoot, "Expected IsRoot to return false for user with non-root roles")

	// Test case 4: User has a root role
	u = &fs.User{
		Roles: []*fs.Role{
			{Root: false},
			{Root: true},
			{Root: false},
		},
	}
	isRoot = u.IsRoot()
	assert.True(t, isRoot, "Expected IsRoot to return true for user with a root role")
}

func TestUserSchema(t *testing.T) {
	u := fs.User{}
	schema := u.Schema()

	assert.NotNil(t, schema, "Expected Schema to return a non-nil schema")
	assert.NotNil(t, schema.DB, "Expected Schema.DB to be non-nil")
	assert.NotNil(t, schema.DB.Indexes, "Expected Schema.DB.Indexes to be non-nil")
	assert.Len(t, schema.DB.Indexes, 3, "Expected 3 database indexes")

	// Verify index names
	indexNames := make([]string, len(schema.DB.Indexes))
	for i, idx := range schema.DB.Indexes {
		indexNames[i] = idx.Name
	}

	assert.Contains(t, indexNames, "idx_user_provider_provider_id")
	assert.Contains(t, indexNames, "idx_user_username")
	assert.Contains(t, indexNames, "idx_user_email")

	// Verify each index is unique
	for _, idx := range schema.DB.Indexes {
		assert.True(t, idx.Unique, "Expected index %s to be unique", idx.Name)
	}
}

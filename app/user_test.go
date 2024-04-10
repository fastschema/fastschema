package app_test

import (
	"testing"
	"time"

	"github.com/fastschema/fastschema/app"
	"github.com/stretchr/testify/assert"
)

func TestUserIsRoot(t *testing.T) {
	// Test case 1: User is nil
	var u *app.User
	isRoot := u.IsRoot()
	assert.False(t, isRoot, "Expected IsRoot to return false for nil user")

	// Test case 2: User has no roles
	u = &app.User{}
	isRoot = u.IsRoot()
	assert.False(t, isRoot, "Expected IsRoot to return false for user with no roles")

	// Test case 3: User has roles but none are root
	u = &app.User{
		Roles: []*app.Role{
			{Root: false},
			{Root: false},
		},
	}
	isRoot = u.IsRoot()
	assert.False(t, isRoot, "Expected IsRoot to return false for user with non-root roles")

	// Test case 4: User has a root role
	u = &app.User{
		Roles: []*app.Role{
			{Root: false},
			{Root: true},
			{Root: false},
		},
	}
	isRoot = u.IsRoot()
	assert.True(t, isRoot, "Expected IsRoot to return true for user with a root role")
}

func TestUserJwtClaim(t *testing.T) {
	// Test case 1: User has no roles
	u := &app.User{}
	key := "secret"
	token, _, err := u.JwtClaim(key)
	assert.NoError(t, err, "Expected no error")
	assert.NotEmpty(t, token, "Expected token to be non-empty")

	// Test case 2: User has roles
	u = &app.User{
		Roles: []*app.Role{
			{ID: 1},
			{ID: 2},
		},
	}
	token, _, err = u.JwtClaim(key)
	assert.NoError(t, err, "Expected no error")
	assert.NotEmpty(t, token, "Expected token to be non-empty")

	// Test case 3: With expiration
	exp := time.Now().Add(1 * time.Hour)
	token, expTime, err := u.JwtClaim(key, exp)
	assert.NoError(t, err, "Expected no error")
	assert.NotEmpty(t, token, "Expected token to be non-empty")
	assert.Equal(t, exp, expTime, "Expected expiration time to match input")
}

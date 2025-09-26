package fs_test

import (
	"testing"
	"time"

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

func TestUserJwtClaim(t *testing.T) {
	// Test case 1: User has no roles
	u := &fs.User{}
	key := "secret"
	token, _, err := u.JwtClaim(nil, &fs.UserJwtConfig{Key: key})
	assert.NoError(t, err, "Expected no error")
	assert.NotEmpty(t, token, "Expected token to be non-empty")

	// Test case 2: User has roles
	u = &fs.User{
		Roles: []*fs.Role{
			{ID: 1},
			{ID: 2},
		},
	}
	token, _, err = u.JwtClaim(nil, &fs.UserJwtConfig{Key: key})
	assert.NoError(t, err, "Expected no error")
	assert.NotEmpty(t, token, "Expected token to be non-empty")

	// Test case 3: With expiration
	exp := time.Now().Add(1 * time.Hour)
	config := &fs.UserJwtConfig{Key: key, ExpiresAt: exp}
	token, expTime, err := u.JwtClaim(nil, config)
	assert.NoError(t, err, "Expected no error")
	assert.NotEmpty(t, token, "Expected token to be non-empty")
	assert.Equal(t, exp, expTime, "Expected expiration time to match input")
}

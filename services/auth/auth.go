package authservice

import (
	"strings"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/fs"
)

type AppLike interface {
	DB() db.Client
	Key() string
	GetAuthProvider(string) fs.AuthProvider
	Roles() []*fs.Role
}

type AuthService struct {
	DB              func() db.Client
	AppKey          func() string
	GetAuthProvider func(string) fs.AuthProvider
	Roles           func() []*fs.Role
}

func New(app AppLike) *AuthService {
	return &AuthService{
		DB:              app.DB,
		AppKey:          app.Key,
		GetAuthProvider: app.GetAuthProvider,
		Roles:           app.Roles,
	}
}

func (as *AuthService) GetRolesFromIDs(ids []uint64) []*fs.Role {
	result := []*fs.Role{}

	for _, role := range as.Roles() {
		for _, id := range ids {
			if role.ID == id {
				result = append(result, role)
			}
		}
	}

	return result
}

func (as *AuthService) GetPermission(roleID uint64, resource string) *fs.Permission {
	matchedRole := &fs.Role{
		ID:          roleID,
		Permissions: []*fs.Permission{},
	}

	for _, role := range as.Roles() {
		if role.ID == roleID {
			matchedRole = role
		}
	}

	for _, permission := range matchedRole.Permissions {
		allowWildcard := strings.HasSuffix(permission.Resource, ".*") &&
			strings.HasPrefix(resource, permission.Resource[:len(permission.Resource)-2])
		allowExact := permission.Resource == resource

		if allowWildcard || allowExact {
			return permission
		}
	}

	return &fs.Permission{}
}

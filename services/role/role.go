package roleservice

import (
	"context"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/fs"
)

type AppLike interface {
	DB() db.Client
	Roles() []*fs.Role
	Key() string
	UpdateCache(ctx context.Context) error
	Resources() *fs.ResourcesManager
}

type RoleService struct {
	DB          func() db.Client
	Roles       func() []*fs.Role
	AppKey      func() string
	UpdateCache func(context.Context) error
	Resources   func() *fs.ResourcesManager
}

func New(app AppLike) *RoleService {
	return &RoleService{
		DB:          app.DB,
		Roles:       app.Roles,
		AppKey:      app.Key,
		UpdateCache: app.UpdateCache,
		Resources:   app.Resources,
	}
}

func (rs *RoleService) GetRolesFromIDs(ids []uint64) []*fs.Role {
	result := []*fs.Role{}

	for _, role := range rs.Roles() {
		for _, id := range ids {
			if role.ID == id {
				result = append(result, role)
			}
		}
	}

	return result
}

func (rs *RoleService) GetPermission(roleID uint64, action string) *fs.Permission {
	matchedRole := &fs.Role{
		ID:          roleID,
		Permissions: []*fs.Permission{},
	}

	for _, role := range rs.Roles() {
		if role.ID == roleID {
			matchedRole = role
		}
	}

	for _, permission := range matchedRole.Permissions {
		if permission.Resource == action {
			return permission
		}
	}

	return &fs.Permission{}
}

package roleservice

import (
	"github.com/fastschema/fastschema/app"
)

type AppLike interface {
	DB() app.DBClient
	Roles() []*app.Role
	Key() string
	UpdateCache() error
	Resources() *app.ResourcesManager
}

type RoleService struct {
	DB          func() app.DBClient
	Roles       func() []*app.Role
	AppKey      func() string
	UpdateCache func() error
	Resources   func() *app.ResourcesManager
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

func (rs *RoleService) GetRolesFromIDs(ids []uint64) []*app.Role {
	result := []*app.Role{}

	for _, role := range rs.Roles() {
		for _, id := range ids {
			if role.ID == id {
				result = append(result, role)
			}
		}
	}

	return result
}

func (rs *RoleService) GetPermission(roleID uint64, action string) *app.Permission {
	matchedRole := &app.Role{
		ID:          roleID,
		Permissions: []*app.Permission{},
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

	return &app.Permission{}
}

package app

import (
	"time"

	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
)

// EntityToRole converts an entity to a role
func EntityToRole(e *schema.Entity) *Role {
	role := &Role{
		ID:          e.ID(),
		Root:        false,
		Users:       []*User{},
		Permissions: []*Permission{},
	}

	if name, ok := e.Get("name").(string); ok {
		role.Name = name
	}

	if description, ok := e.Get("description").(string); ok {
		role.Description = description
	}

	if root, ok := e.Get("root").(bool); ok {
		role.Root = root
	}
	if createdAt, ok := e.Get(schema.FieldCreatedAt).(*time.Time); ok {
		role.CreatedAt = createdAt
	}
	if updatedAt, ok := e.Get(schema.FieldUpdatedAt).(*time.Time); ok {
		role.UpdatedAt = updatedAt
	}
	if deletedAt, ok := e.Get(schema.FieldDeletedAt).(*time.Time); ok {
		role.DeletedAt = deletedAt
	}

	permissions, ok := e.Get("permissions").([]*schema.Entity)
	if ok {
		for _, p := range permissions {
			role.Permissions = append(role.Permissions, &Permission{
				Resource: p.Get("resource").(string),
				Value:    p.Get("value").(string),
			})
		}
	}

	return role
}

// EntitiesToRoles converts entities to roles
func EntitiesToRoles(entities []*schema.Entity) []*Role {
	roles := make([]*Role, 0, len(entities))
	for _, e := range entities {
		roles = append(roles, EntityToRole(e))
	}
	return roles
}

// EntityToUser converts an entity to a user
func EntityToUser(e *schema.Entity) *User {
	if e == nil {
		return nil
	}

	user := &User{
		ID:               e.ID(),
		Username:         e.GetString("username"),
		Email:            e.GetString("email"),
		Password:         e.GetString("password"),
		Provider:         e.GetString("provider"),
		ProviderID:       e.GetString("provider_id"),
		ProviderUsername: e.GetString("provider_username"),
		Roles:            []*Role{},
	}

	if roles, ok := e.Get("roles").([]*schema.Entity); ok {
		for _, r := range roles {
			user.Roles = append(user.Roles, EntityToRole(r))
		}
	}

	user.RoleIDs = utils.Map(user.Roles, func(role *Role) uint64 {
		return role.ID
	})

	if active, ok := e.Get("active").(bool); ok {
		user.Active = active
	}
	if createdAt, ok := e.Get(schema.FieldCreatedAt).(*time.Time); ok {
		user.CreatedAt = createdAt
	}
	if updatedAt, ok := e.Get(schema.FieldUpdatedAt).(*time.Time); ok {
		user.UpdatedAt = updatedAt
	}
	if deletedAt, ok := e.Get(schema.FieldDeletedAt).(*time.Time); ok {
		user.DeletedAt = deletedAt
	}

	return user
}

// EntityToFile converts an entity to a file
func EntityToFile(e *schema.Entity, disks ...Disk) *File {
	if e == nil {
		return nil
	}

	file := &File{
		ID:     e.ID(),
		Disk:   e.GetString("disk"),
		Name:   e.GetString("name"),
		Path:   e.GetString("path"),
		Type:   e.GetString("type"),
		Size:   utils.Must(e.GetUint64("size", false)),
		UserID: utils.Must(e.GetUint64("user_id", true)),
	}

	if len(disks) > 0 && disks[0] != nil {
		file.URL = disks[0].URL(file.Path)
	}

	if createdAt, ok := e.Get(schema.FieldCreatedAt).(*time.Time); ok {
		file.CreatedAt = createdAt
	}
	if updatedAt, ok := e.Get(schema.FieldUpdatedAt).(*time.Time); ok {
		file.UpdatedAt = updatedAt
	}
	if deletedAt, ok := e.Get(schema.FieldDeletedAt).(*time.Time); ok {
		file.DeletedAt = deletedAt
	}

	return file
}

// EntitiesToFiles converts entities to files
func EntitiesToFiles(entities []*schema.Entity, disks ...Disk) []*File {
	files := make([]*File, 0, len(entities))
	for _, e := range entities {
		files = append(files, EntityToFile(e, disks...))
	}
	return files
}

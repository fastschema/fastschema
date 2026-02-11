package fs

import (
	"fmt"
)

// RolePermissionSettingsConfig holds role/permission overrides from environment variable
type RolePermissionSettingsConfig struct {
	Roles []*RoleConfig `json:"roles"`
}

// RoleConfig defines permissions for a role
type RoleConfig struct {
	Name        string        `json:"name"`
	Permissions []*Permission `json:"permissions"`
}

// Validate checks the configuration for errors
func (c *RolePermissionSettingsConfig) Validate() error {
	if c == nil {
		return nil
	}

	for i, role := range c.Roles {
		if role.Name == "" {
			return fmt.Errorf("role at index %d has empty name", i)
		}

		for j, perm := range role.Permissions {
			if perm.Resource == "" {
				return fmt.Errorf("permission at index %d for role '%s' has empty resource", j, role.Name)
			}
			if perm.Value == "" {
				return fmt.Errorf("permission at index %d for role '%s' has empty value", j, role.Name)
			}
		}
	}

	return nil
}

// Clone creates a deep copy of the configuration
func (c *RolePermissionSettingsConfig) Clone() *RolePermissionSettingsConfig {
	if c == nil {
		return nil
	}

	clone := &RolePermissionSettingsConfig{
		Roles: make([]*RoleConfig, len(c.Roles)),
	}

	for i, role := range c.Roles {
		clonedRole := &RoleConfig{
			Name:        role.Name,
			Permissions: make([]*Permission, len(role.Permissions)),
		}
		for j, perm := range role.Permissions {
			clonedRole.Permissions[j] = &Permission{
				Resource: perm.Resource,
				Value:    perm.Value,
				Modifier: perm.Modifier,
			}
		}
		clone.Roles[i] = clonedRole
	}

	return clone
}

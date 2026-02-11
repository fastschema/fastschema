package roleservice

import (
	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/fs"
)

// Export exports the current role permissions from the database as a configuration
// that can be used with the ROLE_PERMISSION_SETTINGS environment variable.
func (rs *RoleService) Export(c fs.Context, _ any) (*fs.RolePermissionSettingsConfig, error) {
	roles, err := db.Builder[*fs.Role](rs.DB()).Select(
		"id",
		"name",
		"permissions",
	).Get(c)
	if err != nil {
		return nil, err
	}

	config := &fs.RolePermissionSettingsConfig{
		Roles: []*fs.RoleConfig{},
	}

	for _, role := range roles {
		if len(role.Permissions) == 0 {
			continue
		}

		// Only include resource, value, and modifier fields for export
		exportPerms := make([]*fs.Permission, len(role.Permissions))
		for i, perm := range role.Permissions {
			exportPerms[i] = &fs.Permission{
				Resource: perm.Resource,
				Value:    perm.Value,
				Modifier: perm.Modifier,
			}
		}

		roleConfig := &fs.RoleConfig{
			Name:        role.Name,
			Permissions: exportPerms,
		}
		config.Roles = append(config.Roles, roleConfig)
	}

	return config, nil
}

package fs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRolePermissionSettingsConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  *RolePermissionSettingsConfig
		wantErr bool
		errMsg  string
	}{
		{
			name:    "nil config is valid",
			config:  nil,
			wantErr: false,
		},
		{
			name: "valid config with permissions",
			config: &RolePermissionSettingsConfig{
				Roles: []*RoleConfig{
					{
						Name: "User",
						Permissions: []*Permission{
							{Resource: "content.blog.list", Value: "allow"},
							{Resource: "content.blog.create", Value: "deny"},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid config with modifier",
			config: &RolePermissionSettingsConfig{
				Roles: []*RoleConfig{
					{
						Name: "User",
						Permissions: []*Permission{
							{
								Resource: "content.blog.update",
								Value:    "$context.User().ID == $args.AuthorID",
								Modifier: "let _ = $context.SetArg('filter', '{}')",
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "empty role name",
			config: &RolePermissionSettingsConfig{
				Roles: []*RoleConfig{
					{
						Name: "",
						Permissions: []*Permission{
							{Resource: "content.blog.list", Value: "allow"},
						},
					},
				},
			},
			wantErr: true,
			errMsg:  "role at index 0 has empty name",
		},
		{
			name: "empty permission resource",
			config: &RolePermissionSettingsConfig{
				Roles: []*RoleConfig{
					{
						Name: "User",
						Permissions: []*Permission{
							{Resource: "", Value: "allow"},
						},
					},
				},
			},
			wantErr: true,
			errMsg:  "permission at index 0 for role 'User' has empty resource",
		},
		{
			name: "empty permission value",
			config: &RolePermissionSettingsConfig{
				Roles: []*RoleConfig{
					{
						Name: "User",
						Permissions: []*Permission{
							{Resource: "content.blog.list", Value: ""},
						},
					},
				},
			},
			wantErr: true,
			errMsg:  "permission at index 0 for role 'User' has empty value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRolePermissionSettingsConfigClone(t *testing.T) {
	t.Run("clone nil config", func(t *testing.T) {
		var config *RolePermissionSettingsConfig
		clone := config.Clone()
		assert.Nil(t, clone)
	})

	t.Run("clone valid config", func(t *testing.T) {
		config := &RolePermissionSettingsConfig{
			Roles: []*RoleConfig{
				{
					Name: "User",
					Permissions: []*Permission{
						{Resource: "content.blog.list", Value: "allow", Modifier: "test"},
					},
				},
			},
		}

		clone := config.Clone()

		// Verify clone is not nil
		assert.NotNil(t, clone)

		// Verify clone has same values
		assert.Len(t, clone.Roles, 1)
		assert.Equal(t, "User", clone.Roles[0].Name)
		assert.Len(t, clone.Roles[0].Permissions, 1)
		assert.Equal(t, "content.blog.list", clone.Roles[0].Permissions[0].Resource)
		assert.Equal(t, "allow", clone.Roles[0].Permissions[0].Value)
		assert.Equal(t, "test", clone.Roles[0].Permissions[0].Modifier)

		// Verify clone is independent (modifying clone doesn't affect original)
		clone.Roles[0].Name = "Modified"
		assert.Equal(t, "User", config.Roles[0].Name)
	})
}

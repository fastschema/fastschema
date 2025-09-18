package auth_test

import (
	"testing"

	"github.com/fastschema/fastschema/entity"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/auth"
	"github.com/stretchr/testify/assert"
)

func TestRegisterEntity(t *testing.T) {
	tests := []struct {
		name             string
		register         auth.Register
		activationMethod string
		provider         string
		expectedEntity   *entity.Entity
	}{
		{
			name: "auto activation",
			register: auth.Register{
				Username:        "testuser",
				FirstName:       "first",
				LastName:        "last",
				Email:           "testuser@site.local",
				Password:        "password123",
				ConfirmPassword: "password123",
			},
			activationMethod: "auto",
			provider:         "testprovider",
			expectedEntity: entity.New().
				Set("email", "testuser@site.local").
				Set("password", "password123").
				Set("active", true).
				Set("provider", "testprovider").
				Set("roles", []*entity.Entity{
					entity.New(fs.RoleUser.ID),
				}).
				Set("username", "testuser").
				Set("first_name", "first").
				Set("last_name", "last"),
		},
		{
			name: "manual activation",
			register: auth.Register{
				Username:        "testuser",
				FirstName:       "first",
				LastName:        "last",
				Email:           "testuser@site.local",
				Password:        "password123",
				ConfirmPassword: "password123",
			},
			activationMethod: "manual",
			provider:         "testprovider",
			expectedEntity: entity.New().
				Set("email", "testuser@site.local").
				Set("password", "password123").
				Set("active", false).
				Set("provider", "testprovider").
				Set("roles", []*entity.Entity{
					entity.New(fs.RoleUser.ID),
				}).
				Set("username", "testuser").
				Set("first_name", "first").
				Set("last_name", "last"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entity := tt.register.Entity(tt.activationMethod, tt.provider)
			expectedJson, err := tt.expectedEntity.ToJSON()
			assert.NoError(t, err)
			entityJson, err := entity.ToJSON()
			assert.NoError(t, err)
			assert.Equal(t, expectedJson, entityJson)
		})
	}
}

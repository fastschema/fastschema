package app_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/fastschema/fastschema/app"
	"github.com/stretchr/testify/assert"
)

func TestPermissionTypeString(t *testing.T) {
	tests := []struct {
		name     string
		pt       app.PermissionType
		expected string
	}{
		{
			name:     "Invalid",
			pt:       app.PermissionTypeInvalid,
			expected: "invalid",
		},
		{
			name:     "Allow",
			pt:       app.PermissionTypeAllow,
			expected: "allow",
		},
		{
			name:     "Deny",
			pt:       app.PermissionTypeDeny,
			expected: "deny",
		},
		{
			name:     "Unknown",
			pt:       app.PermissionType(99),
			expected: "invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.pt.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPermissionTypeMarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		pt       app.PermissionType
		expected string
	}{
		{
			name:     "Invalid",
			pt:       app.PermissionTypeInvalid,
			expected: `"invalid"`,
		},
		{
			name:     "Allow",
			pt:       app.PermissionTypeAllow,
			expected: `"allow"`,
		},
		{
			name:     "Deny",
			pt:       app.PermissionTypeDeny,
			expected: `"deny"`,
		},
		{
			name:     "Unknown",
			pt:       app.PermissionType(99),
			expected: `"invalid"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := json.Marshal(tt.pt)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, string(result))
		})
	}
}

func TestPermissionTypeUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected app.PermissionType
	}{
		{
			name:     "Invalid",
			input:    `"invalid"`,
			expected: app.PermissionTypeInvalid,
		},
		{
			name:     "Allow",
			input:    `"allow"`,
			expected: app.PermissionTypeAllow,
		},
		{
			name:     "Deny",
			input:    `"deny"`,
			expected: app.PermissionTypeDeny,
		},
		{
			name:     "Unknown",
			input:    `"unknown"`,
			expected: app.PermissionTypeInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result app.PermissionType
			err := json.Unmarshal([]byte(tt.input), &result)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPermissionTypeUnmarshalJSONError(t *testing.T) {
	var result *app.PermissionType
	err := result.UnmarshalJSON([]byte(`"invalid`))
	assert.Error(t, err)
}

func TestPermissionTypeValid(t *testing.T) {
	tests := []struct {
		name     string
		pt       app.PermissionType
		expected bool
	}{
		{
			name:     "Invalid",
			pt:       app.PermissionTypeInvalid,
			expected: false,
		},
		{
			name:     "Allow",
			pt:       app.PermissionTypeAllow,
			expected: true,
		},
		{
			name:     "Deny",
			pt:       app.PermissionTypeDeny,
			expected: true,
		},
		{
			name:     "Unknown",
			pt:       app.PermissionType(99),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.pt.Valid()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPermissionTypeValues(t *testing.T) {
	expected := []string{"allow", "deny"}
	result := app.PermissionTypeValues()
	assert.Equal(t, expected, result)
}

func TestGetPermissionTypeFromName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected app.PermissionType
	}{
		{
			name:     "Invalid",
			input:    "invalid",
			expected: app.PermissionTypeInvalid,
		},
		{
			name:     "Allow",
			input:    "allow",
			expected: app.PermissionTypeAllow,
		},
		{
			name:     "Deny",
			input:    "deny",
			expected: app.PermissionTypeDeny,
		},
		{
			name:     "Unknown",
			input:    "unknown",
			expected: app.PermissionTypeInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := app.GetPermissionTypeFromName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPermission(t *testing.T) {
	createdAt := time.Now()
	updatedAt := time.Now()
	deletedAt := time.Now()

	p := app.Permission{
		ID:        1,
		RoleID:    2,
		Resource:  "resource",
		Value:     "value",
		Role:      &app.Role{Name: "role"},
		CreatedAt: &createdAt,
		UpdatedAt: &updatedAt,
		DeletedAt: &deletedAt,
	}

	assert.Equal(t, 1, p.ID)
	assert.Equal(t, 2, p.RoleID)
	assert.Equal(t, "resource", p.Resource)
	assert.Equal(t, "value", p.Value)
	assert.Equal(t, &app.Role{Name: "role"}, p.Role)
	assert.Equal(t, &createdAt, p.CreatedAt)
	assert.Equal(t, &updatedAt, p.UpdatedAt)
	assert.Equal(t, &deletedAt, p.DeletedAt)
}

// package app

// import (
// 	"testing"
// )

// func TestGetPermissionTypeFromName(t *testing.T) {
// 	tests := []struct {
// 		name     string
// 		expected PermissionType
// 	}{
// 		{"read", ReadPermission},
// 		{"write", WritePermission},
// 		{"execute", ExecutePermission},
// 		{"admin", AdminPermission},
// 		{"invalid", InvalidPermission},
// 	}

// 	for _, tt := range tests {
// 		actual := GetPermissionTypeFromName(tt.name)
// 		if actual != tt.expected {
// 			t.Errorf("GetPermissionTypeFromName(%s) returned incorrect permission type, got: %v, want: %v", tt.name, actual, tt.expected)
// 		}
// 	}
// }

package fs_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/fastschema/fastschema/fs"
	"github.com/stretchr/testify/assert"
)

func TestPermissionTypeString(t *testing.T) {
	tests := []struct {
		name     string
		pt       fs.PermissionType
		expected string
	}{
		{
			name:     "Invalid",
			pt:       fs.PermissionTypeInvalid,
			expected: "invalid",
		},
		{
			name:     "Allow",
			pt:       fs.PermissionTypeAllow,
			expected: "allow",
		},
		{
			name:     "Deny",
			pt:       fs.PermissionTypeDeny,
			expected: "deny",
		},
		{
			name:     "Unknown",
			pt:       fs.PermissionType(99),
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
		pt       fs.PermissionType
		expected string
	}{
		{
			name:     "Invalid",
			pt:       fs.PermissionTypeInvalid,
			expected: `"invalid"`,
		},
		{
			name:     "Allow",
			pt:       fs.PermissionTypeAllow,
			expected: `"allow"`,
		},
		{
			name:     "Deny",
			pt:       fs.PermissionTypeDeny,
			expected: `"deny"`,
		},
		{
			name:     "Unknown",
			pt:       fs.PermissionType(99),
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
		expected fs.PermissionType
	}{
		{
			name:     "Invalid",
			input:    `"invalid"`,
			expected: fs.PermissionTypeInvalid,
		},
		{
			name:     "Allow",
			input:    `"allow"`,
			expected: fs.PermissionTypeAllow,
		},
		{
			name:     "Deny",
			input:    `"deny"`,
			expected: fs.PermissionTypeDeny,
		},
		{
			name:     "Unknown",
			input:    `"unknown"`,
			expected: fs.PermissionTypeInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result fs.PermissionType
			err := json.Unmarshal([]byte(tt.input), &result)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPermissionTypeUnmarshalJSONError(t *testing.T) {
	var result *fs.PermissionType
	err := result.UnmarshalJSON([]byte(`"invalid`))
	assert.Error(t, err)
}

func TestPermissionTypeValid(t *testing.T) {
	tests := []struct {
		name     string
		pt       fs.PermissionType
		expected bool
	}{
		{
			name:     "Invalid",
			pt:       fs.PermissionTypeInvalid,
			expected: false,
		},
		{
			name:     "Allow",
			pt:       fs.PermissionTypeAllow,
			expected: true,
		},
		{
			name:     "Deny",
			pt:       fs.PermissionTypeDeny,
			expected: true,
		},
		{
			name:     "Unknown",
			pt:       fs.PermissionType(99),
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
	result := fs.PermissionTypeValues()
	assert.Equal(t, expected, result)
}

func TestGetPermissionTypeFromName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected fs.PermissionType
	}{
		{
			name:     "Invalid",
			input:    "invalid",
			expected: fs.PermissionTypeInvalid,
		},
		{
			name:     "Allow",
			input:    "allow",
			expected: fs.PermissionTypeAllow,
		},
		{
			name:     "Deny",
			input:    "deny",
			expected: fs.PermissionTypeDeny,
		},
		{
			name:     "Unknown",
			input:    "unknown",
			expected: fs.PermissionTypeInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fs.GetPermissionTypeFromName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPermission(t *testing.T) {
	createdAt := time.Now()
	updatedAt := time.Now()
	deletedAt := time.Now()

	p := fs.Permission{
		ID:        1,
		RoleID:    2,
		Resource:  "resource",
		Value:     "value",
		Role:      &fs.Role{Name: "role"},
		CreatedAt: &createdAt,
		UpdatedAt: &updatedAt,
		DeletedAt: &deletedAt,
	}

	assert.Equal(t, 1, p.ID)
	assert.Equal(t, 2, p.RoleID)
	assert.Equal(t, "resource", p.Resource)
	assert.Equal(t, "value", p.Value)
	assert.Equal(t, &fs.Role{Name: "role"}, p.Role)
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

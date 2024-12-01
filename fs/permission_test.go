package fs_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/fastschema/fastschema/expr"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/utils"
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

func TestPermissionIsAllowedDenied(t *testing.T) {
	tests := []struct {
		name          string
		permission    fs.Permission
		expectAllowed bool
		expectDenied  bool
	}{
		{
			name: "Allowed",
			permission: fs.Permission{
				Value: fs.PermissionTypeAllow.String(),
			},
			expectAllowed: true,
			expectDenied:  false,
		},
		{
			name: "Denied",
			permission: fs.Permission{
				Value: fs.PermissionTypeDeny.String(),
			},
			expectAllowed: false,
			expectDenied:  true,
		},
		{
			name: "Empty",
			permission: fs.Permission{
				Value: "",
			},
			expectAllowed: false,
			expectDenied:  true,
		},
		{
			name: "Invalid",
			permission: fs.Permission{
				Value: "invalid",
			},
			expectAllowed: false,
			expectDenied:  false,
		},
	}

	for i := range tests {
		tt := &tests[i]
		t.Run(tt.name, func(t *testing.T) {
			allowed := tt.permission.IsAllowed()
			denied := tt.permission.IsDenied()
			assert.Equal(t, tt.expectAllowed, allowed)
			assert.Equal(t, tt.expectDenied, denied)
		})
	}
}
func TestPermissionCompile(t *testing.T) {
	tests := []struct {
		name        string
		permission  fs.Permission
		expectError bool
	}{
		{
			name: "EmptyValue",
			permission: fs.Permission{
				Value: "",
			},
			expectError: false,
		},
		{
			name: "AllowValue",
			permission: fs.Permission{
				Value: "allow",
			},
			expectError: false,
		},
		{
			name: "DenyValue",
			permission: fs.Permission{
				Value: "deny",
			},
			expectError: false,
		},
		{
			name: "ValidExpression",
			permission: fs.Permission{
				Value: "1 > 0",
			},
			expectError: false,
		},
		{
			name: "InvalidExpression",
			permission: fs.Permission{
				Value: "1 > ",
			},
			expectError: true,
		},
		{
			name: "InvalidResult",
			permission: fs.Permission{
				Value: "'invalid'",
			},
			expectError: true,
		},
	}

	for i := range tests {
		tt := &tests[i]
		t.Run(tt.name, func(t *testing.T) {
			err := tt.permission.Compile()
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCheck(t *testing.T) {
	ctx := context.Background()

	// Permission is allowed
	{
		p := &fs.Permission{Value: "allow"}
		err := p.Check(ctx, expr.Config{})
		assert.NoError(t, err)
	}

	// Permission is denied
	{
		err := (&fs.Permission{Value: "deny"}).Check(ctx, expr.Config{})
		assert.Error(t, err)
		err = (&fs.Permission{Value: ""}).Check(ctx, expr.Config{})
		assert.Error(t, err)
	}

	// Permission is not compiled
	{
		err := (&fs.Permission{Value: "true"}).Check(ctx, expr.Config{})
		assert.Error(t, err)
	}

	// Permission run error
	{
		program := "$context.Get(\"invalid\") > 0"
		compiled := utils.Must(expr.Compile[*fs.Permission, bool](program))
		p := &fs.Permission{Value: program, RuleProgram: compiled}
		assert.Error(t, p.Check(ctx, expr.Config{}))
	}

	// Permission value error
	{
		ctx = context.WithValue(ctx, "check", "nonbool")
		program := "$context.Value(\"check\")"
		compiled := utils.Must(expr.Compile[*fs.Permission, bool](program))
		p := &fs.Permission{Value: program, RuleProgram: compiled}
		assert.Error(t, p.Check(ctx, expr.Config{}))
	}

	// Permission is denied
	{
		program := "1 < 0"
		compiled := utils.Must(expr.Compile[*fs.Permission, bool](program))
		p := &fs.Permission{Value: program, RuleProgram: compiled}
		assert.Error(t, p.Check(ctx, expr.Config{}))
	}

	// Permission is allowed
	{
		program := "1 > 0"
		compiled := utils.Must(expr.Compile[*fs.Permission, bool](program))
		p := &fs.Permission{Value: program, RuleProgram: compiled}
		assert.NoError(t, p.Check(ctx, expr.Config{}))
	}
}

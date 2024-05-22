package fs_test

import (
	"fmt"
	"testing"

	"github.com/fastschema/fastschema/fs"
	"github.com/stretchr/testify/assert"
)

func TestArgsClone(t *testing.T) {
	a := fs.Args{
		"key1": {Type: fs.TypeString, Required: true},
	}

	clone := a.Clone()

	// Check if the clone is equal to the original
	assert.Equal(t, a, clone)

	// Check if the clone is a different instance
	assert.NotSame(t, &a, &clone)

	// Modify the clone and check if it doesn't affect the original
	clone["key1"] = fs.Arg{Type: fs.TypeInt, Required: false}
	assert.NotEqual(t, a, clone)
}

func TestArgTypeCommon(t *testing.T) {
	tests := []struct {
		name     string
		argType  fs.ArgType
		expected string
	}{
		{
			name:     "Valid ArgType",
			argType:  fs.TypeString,
			expected: "string",
		},
		{
			name:     "Invalid ArgType",
			argType:  fs.TypeInvalid,
			expected: "invalid",
		},
		{
			name:     "Invalid ArgType 2",
			argType:  fs.TypeFloat64 + 1,
			expected: "invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.argType.Common()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestArgTypeString(t *testing.T) {
	tests := []struct {
		name     string
		argType  fs.ArgType
		expected string
	}{
		{
			name:     "Valid ArgType",
			argType:  fs.TypeString,
			expected: "string",
		},
		{
			name:     "Invalid ArgType",
			argType:  fs.TypeInvalid,
			expected: "invalid",
		},
		{
			name:     "Invalid ArgType 2",
			argType:  fs.TypeFloat64 + 1,
			expected: "invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.argType.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestArgTypeValid(t *testing.T) {
	tests := []struct {
		name     string
		argType  fs.ArgType
		expected bool
	}{
		{
			name:     "Valid ArgType",
			argType:  fs.TypeString,
			expected: true,
		},
		{
			name:     "Invalid ArgType",
			argType:  fs.TypeInvalid,
			expected: false,
		},
		{
			name:     "Invalid ArgType 2",
			argType:  fs.TypeFloat64 + 1,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.argType.Valid()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestArgTypeMarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		argType  fs.ArgType
		expected string
	}{
		{
			name:     "Valid ArgType",
			argType:  fs.TypeString,
			expected: `"string"`,
		},
		{
			name:     "Invalid ArgType",
			argType:  fs.TypeInvalid,
			expected: `"invalid"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.argType.MarshalJSON()
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, string(result))
		})
	}
}

func TestArgTypeUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name         string
		json         string
		expectedType fs.ArgType
		expectedErr  error
	}{
		{
			name:         "Valid JSON",
			json:         `"string"`,
			expectedType: fs.TypeString,
			expectedErr:  nil,
		},
		{
			name:         "Invalid JSON",
			json:         `"unknown"`,
			expectedType: fs.TypeInvalid,
			expectedErr:  fmt.Errorf("invalid arg type %q", "unknown"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var argType fs.ArgType
			err := argType.UnmarshalJSON([]byte(tt.json))

			assert.Equal(t, tt.expectedType, argType)
			assert.Equal(t, tt.expectedErr, err)
		})
	}
}

func TestArgTypeUnmarshalJSONError(t *testing.T) {
	var argType fs.ArgType
	err := argType.UnmarshalJSON([]byte("invalid json"))

	assert.Error(t, err)
}

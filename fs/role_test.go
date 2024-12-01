package fs_test

import (
	"context"
	"testing"

	"github.com/fastschema/fastschema/expr"
	"github.com/fastschema/fastschema/fs"
	"github.com/stretchr/testify/assert"
)

func TestRoleCompile(t *testing.T) {
	tests := []struct {
		name    string
		role    *fs.Role
		wantErr bool
	}{
		{
			name: "Empty rule",
			role: &fs.Role{
				Rule: "",
			},
			wantErr: false,
		},
		{
			name: "Valid rule",
			role: &fs.Role{
				Rule: "true",
			},
			wantErr: false,
		},
		{
			name: "Invalid rule",
			role: &fs.Role{
				Rule: "invalid rule",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.role.Compile()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.role.Rule != "" {
					assert.NotNil(t, tt.role.RuleProgram)
				}
			}
		})
	}
}
func TestRoleCheck(t *testing.T) {
	tests := []struct {
		name    string
		role    *fs.Role
		context context.Context
		config  expr.Config
		wantErr bool
	}{
		{
			name: "No rule program",
			role: &fs.Role{
				RuleProgram: nil,
			},
			context: nil,
			config:  expr.Config{},
			wantErr: false,
		},
		{
			name: "Rule program run error",
			role: func() *fs.Role {
				r := &fs.Role{
					Rule: "$context.Value('invalid') > 1",
				}
				err := r.Compile()
				assert.NoError(t, err)
				return r
			}(),
			context: context.Background(),
			config:  expr.Config{},
			wantErr: true,
		},
		{
			name: "Rule program return non-bool",
			role: func() *fs.Role {
				r := &fs.Role{
					Rule: "$context.Value('stringkey')",
				}
				err := r.Compile()
				assert.NoError(t, err)
				return r
			}(),
			context: context.WithValue(context.Background(), "stringkey", "stringval"),
			config:  expr.Config{},
			wantErr: true,
		},
		{
			name: "Rule program run true",
			role: func() *fs.Role {
				r := &fs.Role{
					Rule: "true",
				}
				err := r.Compile()
				assert.NoError(t, err)
				return r
			}(),
			context: nil,
			config:  expr.Config{},
			wantErr: false,
		},
		{
			name: "Rule program run false",
			role: func() *fs.Role {
				r := &fs.Role{
					Rule: "false",
				}
				err := r.Compile()
				assert.NoError(t, err)
				return r
			}(),
			context: nil,
			config:  expr.Config{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("Testing %s", tt.name)
			err := tt.role.Check(tt.context, tt.config)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

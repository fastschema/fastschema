package fs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/fastschema/fastschema/expr"
	"github.com/fastschema/fastschema/pkg/errors"
)

// Permission is a struct that contains the permission data
type Permission struct {
	mu          sync.Mutex
	ID          int                              `json:"id,omitempty"`
	RoleID      int                              `json:"role_id,omitempty"`
	Resource    string                           `json:"resource,omitempty"`
	Value       string                           `json:"value,omitempty"`
	Modifier    string                           `json:"modifier,omitempty" fs:"type=json;optional"`
	Role        *Role                            `json:"role,omitempty" fs.relation:"{'type':'o2m','schema':'role','field':'permissions','owner':false,'fk_columns':{'target_column':'role_id'}}"`
	CreatedAt   *time.Time                       `json:"created_at,omitempty"`
	UpdatedAt   *time.Time                       `json:"updated_at,omitempty"`
	DeletedAt   *time.Time                       `json:"deleted_at,omitempty"`
	RuleProgram *expr.Program[*Permission, bool] `json:"-"` // The compiled rule program
}

func (p *Permission) IsAllowed() bool {
	return p.Value == PermissionTypeAllow.String()
}

func (p *Permission) IsDenied() bool {
	return p.Value == "" || p.Value == PermissionTypeDeny.String()
}

func (p *Permission) Compile() (err error) {
	if p.Value == "" || p.Value == "allow" || p.Value == "deny" {
		return nil
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if p.RuleProgram, err = expr.Compile[*Permission, bool](p.Value); err != nil {
		return err
	}

	return nil
}

func (p *Permission) Check(c context.Context, config expr.Config) error {
	if p.IsAllowed() {
		return nil
	}

	if p.IsDenied() {
		return errors.Unauthorized("permission denied")
	}

	if p.RuleProgram == nil {
		return errors.InternalServerError("permission rule is not compiled for permission: %v", p)
	}

	result, err := p.RuleProgram.Run(c, p, config)
	if err != nil {
		return fmt.Errorf("error running permission rule: %w", err)
	}

	check, err := result.Value()
	if err != nil {
		return fmt.Errorf("error getting permission rule value: %w", err)
	}

	if !check {
		return errors.Unauthorized("permission denied")
	}

	return nil
}

// PermissionType is an enum that represents the type of a permission
type PermissionType uint

const (
	PermissionTypeInvalid PermissionType = iota
	PermissionTypeAllow
	PermissionTypeDeny
	endPermissionTypes
)

var (
	// PermissionTypeToStrings is a map that contains the string representation of the enum values
	PermissionTypeToStrings = [...]string{
		PermissionTypeInvalid: "invalid",
		PermissionTypeAllow:   "allow",
		PermissionTypeDeny:    "deny",
	}

	stringToPermissionTypees = map[string]PermissionType{
		"invalid": PermissionTypeInvalid,
		"allow":   PermissionTypeAllow,
		"deny":    PermissionTypeDeny,
	}
)

// GetPermissionTypeFromName returns the type from a string.
func GetPermissionTypeFromName(name string) PermissionType {
	return stringToPermissionTypees[name]
}

// PermissionTypeValues returns all possible values of the enum.
func PermissionTypeValues() []string {
	return PermissionTypeToStrings[1:]
}

// String returns the string representation of a type.
func (p PermissionType) String() string {
	if p < endPermissionTypes {
		return PermissionTypeToStrings[p]
	}
	return PermissionTypeToStrings[PermissionTypeInvalid]
}

// Valid reports if the given type if known type.
func (p PermissionType) Valid() bool {
	return p > PermissionTypeInvalid && p < endPermissionTypes
}

// MarshalJSON marshal an enum value to the quoted json string value
func (p PermissionType) MarshalJSON() ([]byte, error) {
	if !p.Valid() {
		p = PermissionTypeInvalid
	}

	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(PermissionTypeToStrings[p])
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

// UnmarshalJSON unmashals a quoted json string to the enum value
func (p *PermissionType) UnmarshalJSON(b []byte) error {
	var j string
	if err := json.Unmarshal(b, &j); err != nil {
		return err
	}
	*p = stringToPermissionTypees[j] // If the string can't be found, it will be set to the zero value: 'invalid'
	return nil
}

package app

import (
	"bytes"
	"encoding/json"
	"time"
)

type Permission struct {
	ID       int    `json:"id,omitempty"`
	RoleID   int    `json:"role_id,omitempty"`
	Resource string `json:"resource,omitempty"`
	Value    string `json:"value,omitempty"`
	Role     *Role  `json:"role,omitempty"`

	CreatedAt *time.Time `json:"created_at,omitempty"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
	DeletedAt *time.Time `json:"Deleted_at,omitempty"`
}

type PermissionType uint

const (
	PermissionTypeInvalid PermissionType = iota
	PermissionTypeAllow
	PermissionTypeDeny
	endPermissionTypes
)

var (
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

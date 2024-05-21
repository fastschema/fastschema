package fs

import (
	"time"
)

// RoleAdmin is the admin role
var RoleAdmin = &Role{
	ID:   1,
	Name: "Admin",
	Root: true,
}

// RoleUser is the user role
var RoleUser = &Role{
	ID:   2,
	Name: "User",
	Root: false,
}

// RoleGuest is the guest role
var RoleGuest = &Role{
	ID:   3,
	Name: "Guest",
	Root: false,
}

// GuestUser is the guest user
var GuestUser = &User{
	ID:       0,
	Username: "",
	Roles:    []*Role{RoleGuest},
}

// Role is a struct that contains the role data
type Role struct {
	_           any           `json:"-" fs:"label_field=name"`
	ID          uint64        `json:"id,omitempty"`
	Name        string        `json:"name,omitempty"`
	Description string        `json:"description,omitempty" fs:"optional"`
	Root        bool          `json:"root,omitempty" fs:"optional"`
	Users       []*User       `json:"users,omitempty" fs:"type=relation" fs.relation:"{'type':'m2m','schema':'user','field':'roles','owner':true}"`
	Permissions []*Permission `json:"permissions,omitempty" fs:"type=relation" fs.relation:"{'type':'o2m','schema':'permission','field':'role','owner':true}"`

	CreatedAt *time.Time `json:"created_at,omitempty" fs:"default=NOW()"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

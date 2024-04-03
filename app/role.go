package app

import (
	"time"
)

var RoleAdmin = &Role{
	ID:   1,
	Name: "Admin",
	Root: true,
}

var RoleUser = &Role{
	ID:   2,
	Name: "User",
	Root: false,
}

var RoleGuest = &Role{
	ID:   3,
	Name: "Guest",
	Root: false,
}

var GuestUser = &User{
	ID:       0,
	Username: "",
	Roles:    []*Role{RoleGuest},
}

type Role struct {
	ID          uint64        `json:"id,omitempty"`
	Name        string        `json:"name,omitempty"`
	Description string        `json:"description,omitempty"`
	Root        bool          `json:"root,omitempty"`
	Users       []*User       `json:"users,omitempty"`
	Permissions []*Permission `json:"permissions,omitempty"`

	CreatedAt *time.Time `json:"created_at,omitempty"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

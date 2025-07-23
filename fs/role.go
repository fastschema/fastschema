package fs

import (
	"context"
	"fmt"
	"time"

	"github.com/fastschema/fastschema/expr"
	"github.com/fastschema/fastschema/pkg/errors"
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

// Role is a struct that contains the role data
type Role struct {
	_           any           `json:"-" fs:"label_field=name"`
	ID          uint64        `json:"id,omitempty"`
	Name        string        `json:"name,omitempty"`
	Description string        `json:"description,omitempty" fs:"optional"`
	Root        bool          `json:"root,omitempty" fs:"optional"`
	Users       []*User       `json:"users,omitempty" fs.relation:"{'type':'m2m','schema':'user','field':'roles','owner':true}"`
	Permissions []*Permission `json:"permissions,omitempty" fs.relation:"{'type':'o2m','schema':'permission','field':'role','owner':true}"`

	Rule        string                     `json:"rule" fs:"optional"`
	RuleProgram *expr.Program[*Role, bool] `json:"-"` // The compiled rule program

	CreatedAt *time.Time `json:"created_at,omitempty" fs:"default=NOW()"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

func (r *Role) Compile() (err error) {
	if r.Rule == "" {
		return nil
	}

	r.RuleProgram, err = expr.Compile[*Role, bool](r.Rule)
	return err
}

func (r *Role) Check(c context.Context, config expr.Config) error {
	if r.RuleProgram == nil {
		return nil
	}

	result, err := r.RuleProgram.Run(c, r, config)
	if err != nil {
		return fmt.Errorf("error running role rule: %w", err)
	}

	check, err := result.Value()
	if err != nil {
		return fmt.Errorf("error getting role rule value: %w", err)
	}

	if !check {
		return errors.New("role rule returned false")
	}

	return nil
}

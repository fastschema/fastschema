package fs

import (
	"time"

	"github.com/fastschema/fastschema/schema"
)

// User is a struct that contains user data
type User struct {
	_         any    `json:"-" fs:"label_field=username"`
	ID        uint64 `json:"id,omitempty"`
	Username  string `json:"username,omitempty" fs:"optional;sortable"`
	Email     string `json:"email,omitempty" fs:"optional;sortable"`
	FirstName string `json:"first_name,omitempty" fs:"optional;sortable"`
	LastName  string `json:"last_name,omitempty" fs:"optional;sortable"`
	Password  string `json:"password,omitempty" fs:"optional" fs.setter:"$args.Exist && $args.Value != '' ? $hash($args.Value) : $undefined" fs.getter:"$context.Value('keeppassword') == 'true' ? $args.Value : $undefined"`

	Active               bool   `json:"active,omitempty" fs:"optional;sortable"`
	Provider             string `json:"provider,omitempty" fs:"optional"`
	ProviderID           string `json:"provider_id,omitempty" fs:"optional"`
	ProviderUsername     string `json:"provider_username,omitempty" fs:"optional"`
	ProviderProfileImage string `json:"provider_profile_image,omitempty" fs:"optional"`

	RoleIDs []uint64 `json:"role_ids,omitempty"`
	Roles   []*Role  `json:"roles,omitempty" fs.relation:"{'type':'m2m','schema':'role','field':'users','owner':false}"`
	Files   []*File  `json:"files,omitempty" fs.relation:"{'type':'o2m','schema':'file','field':'user','owner':true}"`

	CreatedAt *time.Time `json:"created_at,omitempty"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

func (u User) Schema() *schema.Schema {
	return &schema.Schema{
		Fields: []*schema.Field{},
		DB: &schema.SchemaDB{
			Indexes: []*schema.SchemaDBIndex{
				// unique index on provider + provider_id
				{
					Name:    "idx_user_provider_provider_id",
					Unique:  true,
					Columns: []string{"provider", "provider_id"},
				},
				// unique index on username
				{
					Name:    "idx_user_username",
					Unique:  true,
					Columns: []string{"username"},
				},
				// unique index on email
				{
					Name:    "idx_user_email",
					Unique:  true,
					Columns: []string{"email"},
				},
			},
		},
	}
}

func (u *User) IsRoot() bool {
	if u == nil {
		return false
	}

	if u.Roles == nil {
		return false
	}

	for _, role := range u.Roles {
		if role.Root {
			return true
		}
	}

	return false
}

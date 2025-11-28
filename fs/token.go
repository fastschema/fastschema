package fs

import (
	"time"

	"github.com/fastschema/fastschema/schema"
)

// Token is the schema for storing refresh tokens
type Token struct {
	_         any        `json:"-" fs:"label_field=jti"`
	ID        uint64     `json:"id,omitempty"`
	UserID    uint64     `json:"user_id,omitempty"`
	JTI       string     `json:"jti,omitempty" fs:"unique"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

func (t Token) Schema() *schema.Schema {
	return &schema.Schema{
		Fields: []*schema.Field{},
		DB: &schema.SchemaDB{
			Indexes: []*schema.SchemaDBIndex{
				// Index on jti for fast lookups
				{
					Name:    "idx_token_jti",
					Unique:  true,
					Columns: []string{"jti"},
				},
				// Index on user_id for finding all tokens for a user
				{
					Name:    "idx_token_user_id",
					Columns: []string{"user_id"},
				},
				// Index on expires_at for cleanup of expired tokens
				{
					Name:    "idx_token_expires_at",
					Columns: []string{"expires_at"},
				},
			},
		},
	}
}

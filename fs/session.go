package fs

import (
	"time"

	"github.com/fastschema/fastschema/schema"
)

// SessionStatus represents the status of a session
type SessionStatus string

const (
	SessionStatusActive   SessionStatus = "active"
	SessionStatusInactive SessionStatus = "inactive"
	SessionStatusRevoked  SessionStatus = "revoked"
)

// Session is the schema for storing user sessions (refresh tokens)
type Session struct {
	_              any        `json:"-" fs:"label_field=id"`
	ID             uint64     `json:"id,omitempty"`
	UserID         uint64     `json:"user_id,omitempty"`
	DeviceInfo     string     `json:"device_info,omitempty" fs:"size=512;optional"`
	IPAddress      string     `json:"ip_address,omitempty" fs:"optional"`
	LastActivityAt *time.Time `json:"last_activity_at,omitempty" fs:"optional"`
	Status         string     `json:"status,omitempty" fs:"size=20;optional"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`
	CreatedAt      *time.Time `json:"created_at,omitempty"`
	DeletedAt      *time.Time `json:"deleted_at,omitempty"`
}

func (s Session) Schema() *schema.Schema {
	return &schema.Schema{
		Fields: []*schema.Field{},
		DB: &schema.SchemaDB{
			Indexes: []*schema.SchemaDBIndex{
				// Index on user_id for finding all sessions for a user
				{
					Name:    "idx_session_user_id",
					Columns: []string{"user_id"},
				},
				// Index on expires_at for cleanup of expired sessions
				{
					Name:    "idx_session_expires_at",
					Columns: []string{"expires_at"},
				},
				// Index on status for filtering active/inactive sessions
				{
					Name:    "idx_session_status",
					Columns: []string{"status"},
				},
				// Composite index for common queries
				{
					Name:    "idx_session_user_status",
					Columns: []string{"user_id", "status"},
				},
			},
		},
	}
}

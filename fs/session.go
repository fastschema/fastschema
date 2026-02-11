package fs

import (
	"time"

	"github.com/fastschema/fastschema/schema"
	"github.com/google/uuid"
)

// SessionStatus represents the status of a session
type SessionStatus string

const (
	SessionStatusActive     SessionStatus = "active"
	SessionStatusInactive   SessionStatus = "inactive"
	SessionStatusRevoked    SessionStatus = "revoked"
	SessionStatusPendingOTP SessionStatus = "pending_otp"
	SessionStatusVerified   SessionStatus = "verified" // OTP verified, awaiting action (e.g., password reset)
)

// SessionType represents the type of session
type SessionType string

const (
	SessionTypeRefreshToken SessionType = "refresh_token"
	SessionTypeOTPLogin     SessionType = "otp_login"
	SessionTypeActivation   SessionType = "activation" // Account activation OTP
	SessionTypeRecovery     SessionType = "recovery"   // Password recovery OTP
	// SessionTypeOTP2FA
)

// Session is the schema for storing user sessions (refresh tokens and OTP sessions)
type Session struct {
	_              any        `json:"-" fs:"label_field=id"`
	ID             uuid.UUID  `json:"id" fs:"type=uuid;filterable;sortable"`
	UserID         uuid.UUID  `json:"user_id,omitempty" fs:"type=uuid"`
	DeviceInfo     string     `json:"device_info,omitempty" fs:"size=512;optional"`
	IPAddress      string     `json:"ip_address,omitempty" fs:"optional"`
	LastActivityAt *time.Time `json:"last_activity_at,omitempty" fs:"optional"`
	Status         string     `json:"status,omitempty" fs:"size=20;optional"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`
	CreatedAt      *time.Time `json:"created_at,omitempty"`
	DeletedAt      *time.Time `json:"deleted_at,omitempty"`

	// OTP-specific fields for passwordless login (and future 2FA)
	Type        string `json:"type,omitempty" fs:"size=20;optional"`      // Session type: refresh_token, otp_login, otp_2fa
	OTPHash     string `json:"otp_hash,omitempty" fs:"size=255;optional"` // Hashed OTP code for security
	OTPAttempts int    `json:"otp_attempts,omitempty" fs:"optional"`      // Number of verification attempts
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
				// Index on type for filtering session types
				{
					Name:    "idx_session_type",
					Columns: []string{"type"},
				},
				// Composite index for OTP session queries
				{
					Name:    "idx_session_type_status",
					Columns: []string{"type", "status"},
				},
			},
		},
	}
}

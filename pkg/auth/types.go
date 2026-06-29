package auth

import (
	"github.com/fastschema/fastschema/entity"
	"github.com/google/uuid"
)

type LoginData struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type Register struct {
	Username        string `json:"username"`
	Email           string `json:"email"`
	FirstName       string `json:"first_name"`
	LastName        string `json:"last_name"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirm_password"`
}

// Entity builds the user entity for registration.
// roleID must be the real DB-generated role ID for the "User" role; caller resolves it before calling.
func (d *Register) Entity(activationMethod, provider string, roleID uuid.UUID) *entity.Entity {
	e := entity.New().
		Set("email", d.Email).
		Set("password", d.Password).
		Set("active", activationMethod == "auto").
		Set("provider", provider).
		Set("roles", []*entity.Entity{
			entity.New(roleID),
		})

	if d.Username != "" {
		e.Set("username", d.Username)
	}

	if d.FirstName != "" {
		e.Set("first_name", d.FirstName)
	}

	if d.LastName != "" {
		e.Set("last_name", d.LastName)
	}

	return e
}

type Recovery struct {
	Email string `json:"email"`
}

// ChangeEmailRequest initiates an authenticated email change. Requires the
// current password (re-authentication) before a pending change is created.
type ChangeEmailRequest struct {
	NewEmail        string `json:"new_email"`
	CurrentPassword string `json:"current_password"`
}

// ConfirmEmailChange completes a pending email change via the single-use,
// time-limited token sent to the new address.
type ConfirmEmailChange struct {
	Token string `json:"token"`
}

// Confirmation supports both link-based (token) and OTP-based (session_id + otp) verification
type Confirmation struct {
	Token     string `json:"token,omitempty"`      // For link-based verification
	SessionID string `json:"session_id,omitempty"` // For OTP-based verification
	OTP       string `json:"otp,omitempty"`        // For OTP-based verification
}

// IsOTPBased returns true if this is an OTP-based confirmation
func (c *Confirmation) IsOTPBased() bool {
	return c.SessionID != "" && c.OTP != ""
}

// IsTokenBased returns true if this is a token-based confirmation
func (c *Confirmation) IsTokenBased() bool {
	return c.Token != ""
}

// SendActivation is used to request activation link or OTP
type SendActivation struct {
	Token string `json:"token,omitempty"` // For resend link (existing flow)
	Email string `json:"email,omitempty"` // For OTP request (new flow)
}

// ResetPassword supports both link-based (token) and OTP-based (session_id) password reset
type ResetPassword struct {
	Token           string `json:"token,omitempty"`      // For link-based verification
	SessionID       string `json:"session_id,omitempty"` // For OTP-based verification
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirm_password"`
}

// EmailChangeResponse is the generic response for email-change endpoints.
type EmailChangeResponse struct {
	Message string `json:"message"`
}

// Activation represents the activation status response
type Activation struct {
	Activation string `json:"activation"`           // auto, manual, email, activated
	SessionID  string `json:"session_id,omitempty"` // Only for OTP flow
	ExpiresIn  int    `json:"expires_in,omitempty"` // Only for OTP flow (seconds)
	Verified   bool   `json:"verified,omitempty"`   // For recover/check OTP response
}

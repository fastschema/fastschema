package auth

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/auth/credentials/idtoken"
	"github.com/fastschema/fastschema/pkg/errors"
)

// GoogleUser captures common OIDC claims.
type GoogleUser struct {
	Sub           string    `json:"sub"` // stable Google user id
	Email         string    `json:"email"`
	EmailVerified bool      `json:"email_verified"`
	Name          string    `json:"name"`
	Picture       string    `json:"picture"`
	Issuer        string    `json:"iss"`
	Audience      string    `json:"aud"`
	ExpiresAt     time.Time `json:"exp"`
	IssuedAt      time.Time `json:"iat"`
}

// VerifyGoogleIDToken validates a Google ID token and returns parsed claims.
// audience must be a Web client ID from Google Cloud console.
func VerifyGoogleIDToken(ctx context.Context, rawToken, audience string) (*GoogleUser, error) {
	if rawToken == "" {
		return nil, errors.New("missing ID token")
	}

	if audience == "" {
		return nil, errors.New("missing audience (client ID)")
	}

	// Validate signature, expiry, issuer, and (if provided) aud
	payload, err := idtoken.Validate(ctx, rawToken, audience)
	if err != nil {
		return nil, fmt.Errorf("id token validation failed: %w", err)
	}

	// Map standard fields
	u := &GoogleUser{
		Sub:       payload.Subject,
		Issuer:    payload.Issuer,
		Audience:  payload.Audience,
		ExpiresAt: time.Unix(payload.Expires, 0),
		IssuedAt:  time.Unix(payload.IssuedAt, 0),
	}

	// Pull common profile/email claims from payload.Claims
	if c := payload.Claims; c != nil {
		if v, ok := c["email"].(string); ok {
			u.Email = v
		}

		if v, ok := c["email_verified"].(bool); ok {
			u.EmailVerified = v
		}

		if v, ok := c["name"].(string); ok {
			u.Name = v
		}

		if v, ok := c["picture"].(string); ok {
			u.Picture = v
		}
	}

	return u, nil
}

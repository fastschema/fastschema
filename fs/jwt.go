package fs

import (
	"time"

	jwt "github.com/golang-jwt/jwt/v4"
)

// UserJwtClaims represents the claims in a user JWT
type UserJwtClaims struct {
	jwt.RegisteredClaims

	User *User `json:"user"`
}

// JwtCustomClaimsFunc is a function that allows customization of JWT claims
type JwtCustomClaimsFunc func(Context, *UserJwtClaims) (jwt.Claims, error)

// JWTTokens represents both access and refresh tokens
type JWTTokens struct {
	AccessToken           string    `json:"token"`
	AccessTokenExpiresAt  time.Time `json:"expires"`
	RefreshToken          string    `json:"refresh_token,omitempty"`
	RefreshTokenExpiresAt time.Time `json:"refresh_token_expires,omitempty"`
}

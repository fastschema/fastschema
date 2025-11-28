package jwt

import (
	"time"

	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/errors"
	"github.com/fastschema/fastschema/pkg/utils"
	jwt "github.com/golang-jwt/jwt/v4"
)

// UserClaims represents the minimal user data needed for JWT generation
type UserClaims struct {
	ID                   uint64   `json:"id,omitempty"`
	Username             string   `json:"username,omitempty"`
	Email                string   `json:"email,omitempty"`
	FirstName            string   `json:"first_name,omitempty"`
	LastName             string   `json:"last_name,omitempty"`
	Active               bool     `json:"active,omitempty"`
	Provider             string   `json:"provider,omitempty"`
	ProviderProfileImage string   `json:"provider_profile_image,omitempty"`
	RoleIDs              []uint64 `json:"role_ids,omitempty"`
}

// AccessTokenClaims represents the claims in an access token JWT
type AccessTokenClaims struct {
	jwt.RegisteredClaims

	User *UserClaims `json:"user"`
}

// RefreshTokenClaims represents the claims in a refresh token JWT
type RefreshTokenClaims struct {
	jwt.RegisteredClaims

	UserID    uint64 `json:"uid"`
	SessionID uint64 `json:"sid"` // Session ID from database
}

// CustomClaimsFunc is a function that allows customization of JWT claims
type CustomClaimsFunc func(claims *AccessTokenClaims) (jwt.Claims, error)

// Config holds configuration for token generation
type Config struct {
	Key                    string
	AccessTokenExpiration  time.Duration
	RefreshTokenExpiration time.Duration
	CustomClaimsFunc       CustomClaimsFunc
}

// DefaultConfig returns default token configuration
func DefaultConfig() *Config {
	return &Config{
		AccessTokenExpiration:  15 * time.Minute,   // Short-lived access token
		RefreshTokenExpiration: 7 * 24 * time.Hour, // 7 days refresh token
	}
}

// JWTTokens represents both access and refresh tokens
type JWTTokens struct {
	AccessToken           string     `json:"token"`
	AccessTokenExpiresAt  time.Time  `json:"expires"`
	RefreshToken          string     `json:"refresh_token,omitempty"`
	RefreshTokenExpiresAt *time.Time `json:"refresh_token_expires,omitempty"`
}

// GenerateAccessToken generates a JWT access token for a user
func GenerateAccessToken(
	user *UserClaims,
	key string,
	expiresAt time.Time,
	customClaimsFunc CustomClaimsFunc,
) (string, time.Time, error) {
	if key == "" {
		return "", time.Time{}, errors.InternalServerError("jwt: missing secret key")
	}

	exp := expiresAt
	if expiresAt.IsZero() {
		exp = time.Now().Add(time.Hour * 24 * 7) // Default to 7 days
	}

	jwtClaims := &AccessTokenClaims{
		User: user,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    utils.Env("APP_NAME"),
			ExpiresAt: &jwt.NumericDate{Time: exp},
		},
	}

	var claims jwt.Claims = jwtClaims
	if customClaimsFunc != nil {
		customClaims, err := customClaimsFunc(jwtClaims)
		if err != nil {
			return "", time.Time{}, err
		}

		if customClaims != nil {
			claims = customClaims
		}
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedString, err := jwtToken.SignedString([]byte(key))

	return signedString, exp, err
}

// GenerateRefreshToken generates a new refresh token JWT with the given session ID
func GenerateRefreshToken(
	userID uint64,
	sessionID uint64,
	key string,
	expiresAt time.Time,
) (string, error) {
	if key == "" {
		return "", errors.InternalServerError("jwt: missing secret key")
	}

	claims := &RefreshTokenClaims{
		UserID:    userID,
		SessionID: sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    utils.Env("APP_NAME"),
			ExpiresAt: &jwt.NumericDate{Time: expiresAt},
			IssuedAt:  &jwt.NumericDate{Time: time.Now()},
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(key))
}

// ParseAccessToken parses and validates an access token
func ParseAccessToken(tokenString, key string) (*AccessTokenClaims, error) {
	if tokenString == "" {
		return nil, errors.BadRequest("access token is required")
	}

	token, err := jwt.ParseWithClaims(
		tokenString,
		&AccessTokenClaims{},
		func(token *jwt.Token) (any, error) {
			return []byte(key), nil
		},
	)
	if err != nil {
		return nil, errors.Unauthorized("invalid access token")
	}

	claims, ok := token.Claims.(*AccessTokenClaims)
	if !ok || !token.Valid {
		return nil, errors.Unauthorized("invalid access token")
	}

	return claims, nil
}

// ParseRefreshToken parses and validates a refresh token
func ParseRefreshToken(tokenString, key string) (*RefreshTokenClaims, error) {
	if tokenString == "" {
		return nil, errors.BadRequest("refresh token is required")
	}

	token, err := jwt.ParseWithClaims(
		tokenString,
		&RefreshTokenClaims{},
		func(token *jwt.Token) (any, error) {
			return []byte(key), nil
		},
	)
	if err != nil {
		return nil, errors.Unauthorized("invalid refresh token")
	}

	claims, ok := token.Claims.(*RefreshTokenClaims)
	if !ok || !token.Valid {
		return nil, errors.Unauthorized("invalid refresh token")
	}

	return claims, nil
}

// UserToJwtClaims converts an fs.User to jwt.UserClaims
func UserToJwtClaims(user *fs.User) *UserClaims {
	roleIDs := make([]uint64, 0)
	for _, role := range user.Roles {
		if role.ID > 0 {
			roleIDs = append(roleIDs, role.ID)
		}
	}

	return &UserClaims{
		ID:                   user.ID,
		Username:             user.Username,
		Email:                user.Email,
		FirstName:            user.FirstName,
		LastName:             user.LastName,
		Active:               user.Active,
		Provider:             user.Provider,
		ProviderProfileImage: user.ProviderProfileImage,
		RoleIDs:              roleIDs,
	}
}

// JwtClaimsToUser converts jwt.UserClaims to fs.User
func JwtClaimsToUser(claims *UserClaims) *fs.User {
	if claims == nil {
		return nil
	}
	return &fs.User{
		ID:                   claims.ID,
		Username:             claims.Username,
		Email:                claims.Email,
		FirstName:            claims.FirstName,
		LastName:             claims.LastName,
		Active:               claims.Active,
		Provider:             claims.Provider,
		ProviderProfileImage: claims.ProviderProfileImage,
		RoleIDs:              claims.RoleIDs,
	}
}

// WrapCustomClaimsFunc wraps fs.JwtCustomClaimsFunc to jwt.CustomClaimsFunc
func WrapCustomClaimsFunc(c fs.Context, fn func() fs.JwtCustomClaimsFunc) CustomClaimsFunc {
	if fn == nil || fn() == nil {
		return nil
	}
	fsFunc := fn()
	return func(claims *AccessTokenClaims) (jwt.Claims, error) {
		fsClaims := &fs.UserJwtClaims{
			User: JwtClaimsToUser(claims.User),
		}
		return fsFunc(c, fsClaims)
	}
}

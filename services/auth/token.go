package authservice

import (
	"time"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/entity"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/auth"
	"github.com/fastschema/fastschema/pkg/errors"
	"github.com/fastschema/fastschema/pkg/jwt"
)

// RefreshTokenRequest represents the request to refresh a token
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// LocalLoginWrapper wraps the local login to support token generation
func (as *AuthService) LocalLoginWrapper(
	localAuthProvider *auth.LocalProvider,
) func(c fs.Context, payload *auth.LoginData) (*fs.JWTTokens, error) {
	return func(c fs.Context, payload *auth.LoginData) (*fs.JWTTokens, error) {
		user, err := localAuthProvider.LocalLogin(c, payload)
		if err != nil {
			return nil, err
		}
		return as.GenerateJWTTokens(c, user)
	}
}

// GenerateJWTTokens generates tokens for a user.
// If refresh token is enabled, it generates both access and refresh tokens.
// Otherwise, it generates only an access token.
func (as *AuthService) GenerateJWTTokens(c fs.Context, user *fs.User) (*fs.JWTTokens, error) {
	now := time.Now()
	accessExpiration := as.GetAccessTokenExpiration()
	accessExpiresAt := now.Add(accessExpiration)

	// Generate access token
	userClaims := jwt.UserToJwtClaims(user)
	customClaimsFunc := jwt.WrapCustomClaimsFunc(c, as.JwtCustomClaimsFunc)

	accessToken, _, err := jwt.GenerateAccessToken(
		userClaims,
		as.AppKey(),
		accessExpiresAt,
		customClaimsFunc,
	)
	if err != nil {
		return nil, err
	}

	// If refresh token is not enabled, return only access token
	if !as.IsRefreshTokenEnabled() {
		return &fs.JWTTokens{
			AccessToken:          accessToken,
			AccessTokenExpiresAt: accessExpiresAt,
		}, nil
	}

	// Generate refresh token with JTI
	refreshExpiration := as.GetRefreshTokenExpiration()
	refreshExpiresAt := now.Add(refreshExpiration)

	jti := jwt.GenerateJTI()
	refreshToken, err := jwt.GenerateRefreshToken(user.ID, jti, as.AppKey(), refreshExpiresAt)
	if err != nil {
		return nil, err
	}

	// Store JTI in database
	if _, err = db.Builder[*fs.Token](as.DB()).Create(c, entity.New().
		Set("user_id", user.ID).
		Set("jti", jti).
		Set("expires_at", refreshExpiresAt),
	); err != nil {
		c.Logger().Errorf("failed to store refresh token: %v", err)
		return nil, errors.InternalServerError("failed to create session")
	}

	return &fs.JWTTokens{
		AccessToken:           accessToken,
		AccessTokenExpiresAt:  accessExpiresAt,
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: &refreshExpiresAt,
	}, nil
}

// RefreshToken handles the token refresh endpoint
func (as *AuthService) RefreshToken(c fs.Context, req *RefreshTokenRequest) (*fs.JWTTokens, error) {
	if req == nil || req.RefreshToken == "" {
		return nil, errors.BadRequest("refresh token is required")
	}

	// Parse and validate the refresh token
	claims, err := jwt.ParseRefreshToken(req.RefreshToken, as.AppKey())
	if err != nil {
		return nil, err
	}

	// Check if token is expired (this is also checked in ParseRefreshToken, but double-check here)
	if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
		return nil, errors.Unauthorized("refresh token expired")
	}

	// Check if the JTI exists in database
	storedToken, err := db.Builder[*fs.Token](as.DB()).
		Where(db.EQ("jti", claims.ID)).
		First(c)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, errors.Unauthorized("invalid refresh token")
		}
		c.Logger().Errorf("failed to lookup refresh token: %v", err)
		return nil, errors.InternalServerError("failed to validate token")
	}

	// Check if the stored token is expired
	if storedToken.ExpiresAt != nil && storedToken.ExpiresAt.Before(time.Now()) {
		// Clean up expired token
		_, _ = db.Builder[*fs.Token](as.DB()).
			Where(db.EQ("id", storedToken.ID)).
			Delete(c)
		return nil, errors.Unauthorized("refresh token expired")
	}

	// Verify user ID matches
	if storedToken.UserID != claims.UserID {
		return nil, errors.Unauthorized("invalid refresh token")
	}

	// Get the user
	user, err := db.Builder[*fs.User](as.DB()).
		Where(db.EQ("id", claims.UserID)).
		Select("id", "username", "email", "provider", "provider_id", "provider_username", "active", "roles").
		First(c)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, errors.Unauthorized("user not found")
		}
		return nil, errors.InternalServerError("failed to get user")
	}

	// Check if user is active
	if !user.Active {
		return nil, errors.Unauthorized("user is inactive")
	}

	// Delete the old refresh token (token rotation for security)
	if _, err = db.Builder[*fs.Token](as.DB()).
		Where(db.EQ("id", storedToken.ID)).
		Delete(c); err != nil {
		c.Logger().Errorf("failed to delete old refresh token: %v", err)
	}

	// Generate new token pair
	return as.GenerateJWTTokens(c, user)
}

// Logout invalidates the refresh token
func (as *AuthService) Logout(c fs.Context, req *RefreshTokenRequest) (bool, error) {
	// If no refresh token provided, just return success
	// (client-side logout is still valid)
	if req == nil || req.RefreshToken == "" {
		return true, nil
	}

	// Parse the refresh token to get the user ID
	claims, err := jwt.ParseRefreshToken(req.RefreshToken, as.AppKey())
	if err != nil {
		return true, errors.Unauthorized("invalid refresh token")
	}

	// Delete the refresh token from database using JTI
	if _, err = db.Builder[*fs.Token](as.DB()).
		Where(db.EQ("jti", claims.ID)).
		Where(db.EQ("user_id", claims.UserID)).
		Delete(c); err != nil {
		c.Logger().Errorf("failed to delete refresh token: %v", err)
	}

	return true, nil
}

// LogoutAll invalidates all refresh tokens for the current user
func (as *AuthService) LogoutAll(c fs.Context, _ any) (bool, error) {
	user := c.User()
	if user == nil {
		return false, errors.Unauthorized("user not authenticated")
	}

	// Delete all refresh tokens for this user
	if _, err := db.Builder[*fs.Token](as.DB()).
		Where(db.EQ("user_id", user.ID)).
		Delete(c); err != nil {
		c.Logger().Errorf("failed to delete all refresh tokens: %v", err)
		return false, errors.InternalServerError("failed to logout from all sessions")
	}

	return true, nil
}

// CleanupExpiredTokens removes expired tokens from the database
func (as *AuthService) CleanupExpiredTokens(c fs.Context) (int, error) {
	result, err := db.Builder[*fs.Token](as.DB()).
		Where(db.LT("expires_at", time.Now())).
		Delete(c)
	if err != nil {
		return 0, err
	}
	return result, nil
}

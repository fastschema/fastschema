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

	// Create session in database first to get the session ID
	refreshExpiration := as.GetRefreshTokenExpiration()
	refreshExpiresAt := now.Add(refreshExpiration)

	// Get device info from User-Agent header, fallback to query param
	deviceInfo := c.Header("User-Agent")
	if deviceInfo == "" {
		deviceInfo = c.Arg("device_info")
	}

	session, err := db.Builder[*fs.Session](as.DB()).Create(c, entity.New().
		Set("user_id", user.ID).
		Set("device_info", deviceInfo).
		Set("ip_address", c.IP()).
		Set("last_activity_at", now).
		Set("status", string(fs.SessionStatusActive)).
		Set("expires_at", refreshExpiresAt),
	)
	if err != nil {
		c.Logger().Errorf("failed to create session: %v", err)
		return nil, errors.InternalServerError("failed to create session")
	}

	// Generate refresh token with session ID
	refreshToken, err := jwt.GenerateRefreshToken(user.ID, session.ID, as.AppKey(), refreshExpiresAt)
	if err != nil {
		// Clean up the session if token generation fails
		_, _ = db.Builder[*fs.Session](as.DB()).Where(db.EQ("id", session.ID)).Delete(c)
		return nil, err
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

	// Check if the session exists in database using session ID
	storedSession, err := db.Builder[*fs.Session](as.DB()).
		Where(db.EQ("id", claims.SessionID)).
		First(c)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, errors.Unauthorized("invalid refresh token")
		}
		c.Logger().Errorf("failed to lookup session: %v", err)
		return nil, errors.InternalServerError("failed to validate token")
	}

	// Check if the session is active
	if storedSession.Status != string(fs.SessionStatusActive) {
		return nil, errors.Unauthorized("session is not active")
	}

	// Check if the stored session is expired
	if storedSession.ExpiresAt != nil && storedSession.ExpiresAt.Before(time.Now()) {
		// Mark session as inactive
		_, _ = db.Builder[*fs.Session](as.DB()).
			Where(db.EQ("id", storedSession.ID)).
			Update(c, entity.New().Set("status", string(fs.SessionStatusInactive)))
		return nil, errors.Unauthorized("refresh token expired")
	}

	// Verify user ID matches
	if storedSession.UserID != claims.UserID {
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

	// Delete the old session (token rotation for security)
	if _, err = db.Builder[*fs.Session](as.DB()).
		Where(db.EQ("id", storedSession.ID)).
		Delete(c); err != nil {
		c.Logger().Errorf("failed to delete old session: %v", err)
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

	// Parse the refresh token to get the session ID
	claims, err := jwt.ParseRefreshToken(req.RefreshToken, as.AppKey())
	if err != nil {
		return true, errors.Unauthorized("invalid refresh token")
	}

	// Delete the session from database using session ID
	if _, err = db.Builder[*fs.Session](as.DB()).
		Where(db.EQ("id", claims.SessionID)).
		Where(db.EQ("user_id", claims.UserID)).
		Delete(c); err != nil {
		c.Logger().Errorf("failed to delete session: %v", err)
	}

	return true, nil
}

// LogoutAll invalidates all sessions for the current user
func (as *AuthService) LogoutAll(c fs.Context, _ any) (bool, error) {
	user := c.User()
	if user == nil {
		return false, errors.Unauthorized("user not authenticated")
	}

	// Delete all sessions for this user
	if _, err := db.Builder[*fs.Session](as.DB()).
		Where(db.EQ("user_id", user.ID)).
		Delete(c); err != nil {
		c.Logger().Errorf("failed to delete all sessions: %v", err)
		return false, errors.InternalServerError("failed to logout from all sessions")
	}

	return true, nil
}

// CleanupExpiredSessions removes expired sessions from the database
func (as *AuthService) CleanupExpiredSessions(c fs.Context) (int, error) {
	result, err := db.Builder[*fs.Session](as.DB()).
		Where(db.LT("expires_at", time.Now())).
		Delete(c)
	if err != nil {
		return 0, err
	}
	return result, nil
}

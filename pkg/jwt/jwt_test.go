package jwt_test

import (
	"testing"
	"time"

	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/jwt"
	jwtlib "github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateAccessToken(t *testing.T) {
	userClaims := &jwt.UserClaims{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
		Active:   true,
		RoleIDs:  []uint64{2},
	}
	key := "test-secret-key-32-characters!!"
	expiresAt := time.Now().Add(time.Hour)

	token, exp, err := jwt.GenerateAccessToken(userClaims, key, expiresAt, nil)
	require.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.Equal(t, expiresAt.Unix(), exp.Unix())
}

func TestGenerateAccessTokenDefaultExpiration(t *testing.T) {
	userClaims := &jwt.UserClaims{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
		Active:   true,
		RoleIDs:  []uint64{2},
	}
	key := "test-secret-key-32-characters!!"

	token, exp, err := jwt.GenerateAccessToken(userClaims, key, time.Time{}, nil)
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	// Should be approximately 7 days from now
	expectedExp := time.Now().Add(time.Hour * 24 * 7)
	assert.WithinDuration(t, expectedExp, exp, time.Minute)
}

func TestGenerateAccessTokenMissingKey(t *testing.T) {
	userClaims := &jwt.UserClaims{
		ID:       1,
		Username: "testuser",
	}

	_, _, err := jwt.GenerateAccessToken(userClaims, "", time.Now().Add(time.Hour), nil)
	assert.Error(t, err)
}

func TestGenerateAccessTokenWithCustomClaims(t *testing.T) {
	userClaims := &jwt.UserClaims{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
		Active:   true,
		RoleIDs:  []uint64{2},
	}
	key := "test-secret-key-32-characters!!"
	expiresAt := time.Now().Add(time.Hour)

	customClaimsFunc := func(claims *jwt.AccessTokenClaims) (jwtlib.Claims, error) {
		return claims, nil
	}

	token, _, err := jwt.GenerateAccessToken(userClaims, key, expiresAt, customClaimsFunc)
	require.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestGenerateRefreshToken(t *testing.T) {
	userID := uint64(1)
	jti := jwt.GenerateJTI()
	key := "test-secret-key-32-characters!!"
	expiresAt := time.Now().Add(7 * 24 * time.Hour)

	token, err := jwt.GenerateRefreshToken(userID, jti, key, expiresAt)
	require.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestGenerateRefreshTokenMissingKey(t *testing.T) {
	userID := uint64(1)
	jti := jwt.GenerateJTI()
	expiresAt := time.Now().Add(7 * 24 * time.Hour)

	_, err := jwt.GenerateRefreshToken(userID, jti, "", expiresAt)
	assert.Error(t, err)
}

func TestGenerateRefreshTokenAutoJTI(t *testing.T) {
	userID := uint64(1)
	key := "test-secret-key-32-characters!!"
	expiresAt := time.Now().Add(7 * 24 * time.Hour)

	// Empty JTI should auto-generate
	token, err := jwt.GenerateRefreshToken(userID, "", key, expiresAt)
	require.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestParseAccessToken(t *testing.T) {
	userClaims := &jwt.UserClaims{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
		Active:   true,
		RoleIDs:  []uint64{2},
	}
	key := "test-secret-key-32-characters!!"
	expiresAt := time.Now().Add(time.Hour)

	token, _, err := jwt.GenerateAccessToken(userClaims, key, expiresAt, nil)
	require.NoError(t, err)

	claims, err := jwt.ParseAccessToken(token, key)
	require.NoError(t, err)
	assert.Equal(t, userClaims.ID, claims.User.ID)
	assert.Equal(t, userClaims.Username, claims.User.Username)
	assert.Equal(t, userClaims.Email, claims.User.Email)
}

func TestParseAccessTokenInvalid(t *testing.T) {
	key := "test-secret-key-32-characters!!"

	_, err := jwt.ParseAccessToken("invalid-token", key)
	assert.Error(t, err)

	_, err = jwt.ParseAccessToken("", key)
	assert.Error(t, err)
}

func TestParseRefreshToken(t *testing.T) {
	userID := uint64(1)
	jti := jwt.GenerateJTI()
	key := "test-secret-key-32-characters!!"
	expiresAt := time.Now().Add(7 * 24 * time.Hour)

	token, err := jwt.GenerateRefreshToken(userID, jti, key, expiresAt)
	require.NoError(t, err)

	claims, err := jwt.ParseRefreshToken(token, key)
	require.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, jti, claims.ID)
}

func TestParseRefreshTokenInvalid(t *testing.T) {
	key := "test-secret-key-32-characters!!"

	_, err := jwt.ParseRefreshToken("invalid-token", key)
	assert.Error(t, err)

	_, err = jwt.ParseRefreshToken("", key)
	assert.Error(t, err)
}

func TestGenerateJTI(t *testing.T) {
	jti1 := jwt.GenerateJTI()
	jti2 := jwt.GenerateJTI()

	assert.Len(t, jti1, 32)
	assert.Len(t, jti2, 32)
	assert.NotEqual(t, jti1, jti2) // Should be unique
}

func TestDefaultConfig(t *testing.T) {
	config := jwt.DefaultConfig()

	assert.Equal(t, 15*time.Minute, config.AccessTokenExpiration)
	assert.Equal(t, 7*24*time.Hour, config.RefreshTokenExpiration)
}

func TestTokenPair(t *testing.T) {
	refreshTokenExpires := time.Now().Add(7 * 24 * time.Hour)
	pair := &jwt.JWTTokens{
		AccessToken:           "access-token",
		AccessTokenExpiresAt:  time.Now().Add(15 * time.Minute),
		RefreshToken:          "refresh-token",
		RefreshTokenExpiresAt: &refreshTokenExpires,
	}

	assert.Equal(t, "access-token", pair.AccessToken)
	assert.Equal(t, "refresh-token", pair.RefreshToken)
}

func TestUserToJwtClaims(t *testing.T) {
	user := &fs.User{
		ID:                   1,
		Username:             "testuser",
		Email:                "test@example.com",
		FirstName:            "Test",
		LastName:             "User",
		Active:               true,
		Provider:             "local",
		ProviderProfileImage: "http://example.com/image.jpg",
		Roles: []*fs.Role{
			{ID: 1, Name: "admin"},
			{ID: 2, Name: "user"},
		},
	}

	claims := jwt.UserToJwtClaims(user)

	assert.Equal(t, user.ID, claims.ID)
	assert.Equal(t, user.Username, claims.Username)
	assert.Equal(t, user.Email, claims.Email)
	assert.Equal(t, user.FirstName, claims.FirstName)
	assert.Equal(t, user.LastName, claims.LastName)
	assert.Equal(t, user.Active, claims.Active)
	assert.Equal(t, user.Provider, claims.Provider)
	assert.Equal(t, user.ProviderProfileImage, claims.ProviderProfileImage)
	assert.Equal(t, []uint64{1, 2}, claims.RoleIDs)
}

func TestJwtClaimsToUser(t *testing.T) {
	claims := &jwt.UserClaims{
		ID:                   1,
		Username:             "testuser",
		Email:                "test@example.com",
		FirstName:            "Test",
		LastName:             "User",
		Active:               true,
		Provider:             "local",
		ProviderProfileImage: "http://example.com/image.jpg",
		RoleIDs:              []uint64{1, 2},
	}

	user := jwt.JwtClaimsToUser(claims)

	assert.Equal(t, claims.ID, user.ID)
	assert.Equal(t, claims.Username, user.Username)
	assert.Equal(t, claims.Email, user.Email)
	assert.Equal(t, claims.FirstName, user.FirstName)
	assert.Equal(t, claims.LastName, user.LastName)
	assert.Equal(t, claims.Active, user.Active)
	assert.Equal(t, claims.Provider, user.Provider)
	assert.Equal(t, claims.ProviderProfileImage, user.ProviderProfileImage)
	assert.Equal(t, claims.RoleIDs, user.RoleIDs)

	// Test nil claims
	assert.Nil(t, jwt.JwtClaimsToUser(nil))
}

func TestWrapCustomClaimsFunc(t *testing.T) {
	// Test nil function
	result := jwt.WrapCustomClaimsFunc(nil, nil)
	assert.Nil(t, result)

	// Test function that returns nil
	result = jwt.WrapCustomClaimsFunc(nil, func() fs.JwtCustomClaimsFunc { return nil })
	assert.Nil(t, result)
}

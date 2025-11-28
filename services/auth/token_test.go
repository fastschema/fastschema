package authservice_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/entity"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/logger"
	"github.com/fastschema/fastschema/pkg/entdbadapter"
	"github.com/fastschema/fastschema/pkg/jwt"
	rr "github.com/fastschema/fastschema/pkg/restfulresolver"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
	as "github.com/fastschema/fastschema/services/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testTokenApp struct {
	db         db.Client
	config     *fs.Config
	authConfig *fs.AuthConfig
	resources  *fs.ResourcesManager
	resolver   *rr.RestfulResolver
}

func (s testTokenApp) JwtCustomClaimsFunc() fs.JwtCustomClaimsFunc {
	return nil
}

func (s testTokenApp) DB() db.Client {
	return s.db
}

func (s testTokenApp) Key() string {
	return "test-secret-key-32-characters!!"
}

func (s testTokenApp) Config() *fs.Config {
	return s.config
}

func (s testTokenApp) Roles() []*fs.Role {
	return []*fs.Role{fs.RoleAdmin, fs.RoleUser, fs.RoleGuest}
}

func (s testTokenApp) GetAuthProvider(name string) fs.AuthProvider {
	return nil
}

func createTestTokenApp(t *testing.T, enableRefreshToken bool) *testTokenApp {
	schemaDir := utils.Must(os.MkdirTemp("", "schemas"))
	migrationDir := utils.Must(os.MkdirTemp("", "migrations"))
	sb := utils.Must(schema.NewBuilderFromDir(schemaDir, fs.SystemSchemaTypes...))
	dbc := utils.Must(entdbadapter.NewTestClient(migrationDir, sb))

	roleModel := utils.Must(dbc.Model("role"))
	userModel := utils.Must(dbc.Model("user"))

	for _, r := range []*fs.Role{fs.RoleAdmin, fs.RoleUser, fs.RoleGuest} {
		utils.Must(roleModel.Create(context.Background(), entity.New().
			Set("name", r.Name).
			Set("root", r.Root),
		))
	}

	utils.Must(userModel.Create(context.Background(), entity.New().
		Set("username", "testuser").
		Set("email", "test@example.com").
		Set("password", "testpassword").
		Set("provider", "local").
		Set("provider_id", "1").
		Set("active", true).
		Set("roles", []*entity.Entity{entity.New(2)}),
	))

	utils.Must(userModel.Create(context.Background(), entity.New().
		Set("username", "inactiveuser").
		Set("email", "inactive@example.com").
		Set("password", "testpassword").
		Set("provider", "local").
		Set("provider_id", "2").
		Set("active", false).
		Set("roles", []*entity.Entity{entity.New(2)}),
	))

	authConfig := &fs.AuthConfig{
		EnableRefreshToken:   enableRefreshToken,
		AccessTokenLifetime:  900,    // 15 minutes
		RefreshTokenLifetime: 604800, // 7 days
	}

	app := &testTokenApp{
		db: dbc,
		config: &fs.Config{
			AppKey:     "test-secret-key-32-characters!!",
			AuthConfig: authConfig,
		},
		authConfig: authConfig,
	}

	// Setup resources and resolver for context creation
	app.resources = fs.NewResourcesManager()
	app.resources.Hooks = func() *fs.Hooks {
		return &fs.Hooks{}
	}
	api := app.resources.Group("api", &fs.Meta{Prefix: "/api"})
	api.Add(fs.Get("test", func(c fs.Context, _ any) (any, error) {
		return "test", nil
	}, &fs.Meta{Public: true}))
	_ = app.resources.Init()

	app.resolver = rr.NewRestfulResolver(&rr.ResolverConfig{
		ResourceManager: app.resources,
		Logger:          logger.CreateMockLogger(false),
	})

	return app
}

func (app *testTokenApp) createMockContext(user *fs.User) fs.Context {
	return &mockContext{
		user: user,
	}
}

// mockContext is a minimal mock implementation of fs.Context for testing
type mockContext struct {
	user   *fs.User
	locals map[string]any
}

func (m *mockContext) TraceID() string                    { return "test-trace" }
func (m *mockContext) User() *fs.User                     { return m.user }
func (m *mockContext) Value(key any) any                  { return nil }
func (m *mockContext) Logger() logger.Logger              { return logger.CreateMockLogger(false) }
func (m *mockContext) AuthToken() string                  { return "" }
func (m *mockContext) Next() error                        { return nil }
func (m *mockContext) Result(...*fs.Result) *fs.Result    { return nil }
func (m *mockContext) Arg(string, ...string) string       { return "" }
func (m *mockContext) ArgInt(string, ...int) int          { return 0 }
func (m *mockContext) Args() map[string]string            { return nil }
func (m *mockContext) SetArg(key, val string) string      { return "" }
func (m *mockContext) Body() ([]byte, error)              { return nil, nil }
func (m *mockContext) Payload() (*entity.Entity, error)   { return nil, nil }
func (m *mockContext) BodyParser(out any) error           { return nil }
func (m *mockContext) Bind(out any) error                 { return nil }
func (m *mockContext) FormValue(string, ...string) string { return "" }
func (m *mockContext) Resource() *fs.Resource             { return nil }
func (m *mockContext) Redirect(string) error              { return nil }
func (m *mockContext) IP() string                         { return "127.0.0.1" }
func (m *mockContext) Local(key string, value ...any) any {
	if m.locals == nil {
		m.locals = make(map[string]any)
	}
	if len(value) > 0 {
		m.locals[key] = value[0]
		return value[0]
	}
	return m.locals[key]
}
func (m *mockContext) Files() ([]*fs.File, error)  { return nil, nil }
func (m *mockContext) WSClient() fs.WSClient       { return nil }
func (m *mockContext) Deadline() (time.Time, bool) { return time.Time{}, false }
func (m *mockContext) Done() <-chan struct{}       { return nil }
func (m *mockContext) Err() error                  { return nil }

func TestAuthServiceGetAccessTokenExpiration(t *testing.T) {
	t.Run("with explicit config value", func(t *testing.T) {
		// When AccessTokenLifetime is explicitly configured, it should be used regardless of EnableRefreshToken
		app := createTestTokenApp(t, false)
		authService := as.New(app)

		// Should be 15 minutes (900 seconds) as configured in createTestTokenApp
		expiration := authService.GetAccessTokenExpiration()
		assert.Equal(t, 900*time.Second, expiration)
	})

	t.Run("with refresh token enabled", func(t *testing.T) {
		app := createTestTokenApp(t, true)
		authService := as.New(app)

		// Should be 15 minutes (900 seconds) as configured
		expiration := authService.GetAccessTokenExpiration()
		assert.Equal(t, 900*time.Second, expiration)
	})
}

func TestAuthServiceGetRefreshTokenExpiration(t *testing.T) {
	app := createTestTokenApp(t, true)
	authService := as.New(app)

	expiration := authService.GetRefreshTokenExpiration()
	assert.Equal(t, 604800*time.Second, expiration) // 7 days
}

func TestAuthServiceIsRefreshTokenEnabled(t *testing.T) {
	t.Run("enabled", func(t *testing.T) {
		app := createTestTokenApp(t, true)
		authService := as.New(app)
		assert.True(t, authService.IsRefreshTokenEnabled())
	})

	t.Run("disabled", func(t *testing.T) {
		app := createTestTokenApp(t, false)
		authService := as.New(app)
		assert.False(t, authService.IsRefreshTokenEnabled())
	})
}

func TestAuthServiceGenerateTokenPair(t *testing.T) {
	app := createTestTokenApp(t, true)
	authService := as.New(app)

	user := &fs.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
		Active:   true,
		Roles:    []*fs.Role{fs.RoleUser},
	}

	ctx := app.createMockContext(nil)
	tokenPair, err := authService.GenerateJWTTokens(ctx, user)
	require.NoError(t, err)

	assert.NotEmpty(t, tokenPair.AccessToken)
	assert.NotEmpty(t, tokenPair.RefreshToken)
	assert.False(t, tokenPair.AccessTokenExpiresAt.IsZero())
	assert.False(t, tokenPair.RefreshTokenExpiresAt.IsZero())

	// Access token should expire before refresh token
	assert.True(t, tokenPair.AccessTokenExpiresAt.Before(*tokenPair.RefreshTokenExpiresAt))

	// Verify refresh token is stored in database using JTI from the token
	claims, err := jwt.ParseRefreshToken(tokenPair.RefreshToken, app.Key())
	require.NoError(t, err)
	storedToken, err := db.Builder[*fs.Token](app.db).
		Where(db.EQ("jti", claims.ID)).
		First(context.Background())
	require.NoError(t, err)
	assert.Equal(t, user.ID, storedToken.UserID)
}

func TestAuthServiceRefreshToken(t *testing.T) {
	app := createTestTokenApp(t, true)
	authService := as.New(app)

	user := &fs.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
		Active:   true,
		Roles:    []*fs.Role{fs.RoleUser},
	}

	ctx := app.createMockContext(nil)

	// Generate initial token pair
	tokenPair, err := authService.GenerateJWTTokens(ctx, user)
	require.NoError(t, err)

	// Refresh the token
	refreshReq := &as.RefreshTokenRequest{
		RefreshToken: tokenPair.RefreshToken,
	}

	newTokenPair, err := authService.RefreshToken(ctx, refreshReq)
	require.NoError(t, err)

	assert.NotEmpty(t, newTokenPair.AccessToken)
	assert.NotEmpty(t, newTokenPair.RefreshToken)

	// New tokens should be different
	assert.NotEqual(t, tokenPair.AccessToken, newTokenPair.AccessToken)
	assert.NotEqual(t, tokenPair.RefreshToken, newTokenPair.RefreshToken)

	// Old refresh token should be invalidated (deleted from DB)
	oldClaims, err := jwt.ParseRefreshToken(tokenPair.RefreshToken, app.Key())
	require.NoError(t, err)
	_, err = db.Builder[*fs.Token](app.db).
		Where(db.EQ("jti", oldClaims.ID)).
		First(context.Background())
	assert.True(t, db.IsNotFound(err))

	// New refresh token should be in DB
	newClaims, err := jwt.ParseRefreshToken(newTokenPair.RefreshToken, app.Key())
	require.NoError(t, err)
	storedToken, err := db.Builder[*fs.Token](app.db).
		Where(db.EQ("jti", newClaims.ID)).
		First(context.Background())
	require.NoError(t, err)
	assert.Equal(t, user.ID, storedToken.UserID)
}

func TestAuthServiceRefreshTokenInvalid(t *testing.T) {
	app := createTestTokenApp(t, true)
	authService := as.New(app)
	ctx := app.createMockContext(nil)

	t.Run("empty refresh token", func(t *testing.T) {
		_, err := authService.RefreshToken(ctx, &as.RefreshTokenRequest{})
		assert.Error(t, err)
	})

	t.Run("nil request", func(t *testing.T) {
		_, err := authService.RefreshToken(ctx, nil)
		assert.Error(t, err)
	})

	t.Run("invalid token format", func(t *testing.T) {
		_, err := authService.RefreshToken(ctx, &as.RefreshTokenRequest{
			RefreshToken: "invalid-token",
		})
		assert.Error(t, err)
	})

	t.Run("token not in database", func(t *testing.T) {
		// Generate a valid token but don't store it in DB
		jti := jwt.GenerateJTI()
		token, err := jwt.GenerateRefreshToken(1, jti, app.Key(), time.Now().Add(time.Hour))
		require.NoError(t, err)

		_, err = authService.RefreshToken(ctx, &as.RefreshTokenRequest{
			RefreshToken: token,
		})
		assert.Error(t, err)
	})
}

func TestAuthServiceRefreshTokenInactiveUser(t *testing.T) {
	app := createTestTokenApp(t, true)
	authService := as.New(app)
	ctx := app.createMockContext(nil)

	// Get the inactive user
	inactiveUser, err := db.Builder[*fs.User](app.db).
		Where(db.EQ("username", "inactiveuser")).
		Select("id", "username", "email", "active", "roles").
		First(context.Background())
	require.NoError(t, err)
	inactiveUser.Roles = []*fs.Role{fs.RoleUser}

	// Generate token pair for inactive user
	tokenPair, err := authService.GenerateJWTTokens(ctx, inactiveUser)
	require.NoError(t, err)

	// Try to refresh - should fail because user is inactive
	_, err = authService.RefreshToken(ctx, &as.RefreshTokenRequest{
		RefreshToken: tokenPair.RefreshToken,
	})
	assert.Error(t, err)
}

func TestAuthServiceLogout(t *testing.T) {
	app := createTestTokenApp(t, true)
	authService := as.New(app)
	ctx := app.createMockContext(nil)

	user := &fs.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
		Active:   true,
		Roles:    []*fs.Role{fs.RoleUser},
	}

	// Generate token pair
	tokenPair, err := authService.GenerateJWTTokens(ctx, user)
	require.NoError(t, err)

	// Logout
	result, err := authService.Logout(ctx, &as.RefreshTokenRequest{
		RefreshToken: tokenPair.RefreshToken,
	})
	require.NoError(t, err)
	assert.True(t, result)

	// Verify token is deleted from DB
	claims, err := jwt.ParseRefreshToken(tokenPair.RefreshToken, app.Key())
	require.NoError(t, err)
	_, err = db.Builder[*fs.Token](app.db).
		Where(db.EQ("jti", claims.ID)).
		First(context.Background())
	assert.True(t, db.IsNotFound(err))

	// Try to refresh with deleted token - should fail
	_, err = authService.RefreshToken(ctx, &as.RefreshTokenRequest{
		RefreshToken: tokenPair.RefreshToken,
	})
	assert.Error(t, err)
}

func TestAuthServiceLogoutWithoutToken(t *testing.T) {
	app := createTestTokenApp(t, true)
	authService := as.New(app)
	ctx := app.createMockContext(nil)

	// Logout without token should succeed (client-side logout)
	result, err := authService.Logout(ctx, nil)
	require.NoError(t, err)
	assert.True(t, result)

	result, err = authService.Logout(ctx, &as.RefreshTokenRequest{})
	require.NoError(t, err)
	assert.True(t, result)
}

func TestAuthServiceLogoutAll(t *testing.T) {
	app := createTestTokenApp(t, true)
	authService := as.New(app)

	user := &fs.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
		Active:   true,
		Roles:    []*fs.Role{fs.RoleUser},
	}

	ctx := app.createMockContext(nil)

	// Generate multiple token pairs
	tokenPair1, err := authService.GenerateJWTTokens(ctx, user)
	require.NoError(t, err)

	tokenPair2, err := authService.GenerateJWTTokens(ctx, user)
	require.NoError(t, err)

	// Set user in context for LogoutAll
	ctx = app.createMockContext(user)

	// Logout all
	result, err := authService.LogoutAll(ctx, nil)
	require.NoError(t, err)
	assert.True(t, result)

	// Verify both tokens are deleted from DB
	for _, refreshToken := range []string{tokenPair1.RefreshToken, tokenPair2.RefreshToken} {
		claims, err := jwt.ParseRefreshToken(refreshToken, app.Key())
		require.NoError(t, err)
		_, err = db.Builder[*fs.Token](app.db).
			Where(db.EQ("jti", claims.ID)).
			First(context.Background())
		assert.True(t, db.IsNotFound(err))
	}
}

func TestAuthServiceLogoutAllUnauthenticated(t *testing.T) {
	app := createTestTokenApp(t, true)
	authService := as.New(app)
	ctx := app.createMockContext(nil)

	// LogoutAll without authenticated user should fail
	_, err := authService.LogoutAll(ctx, nil)
	assert.Error(t, err)
}

func TestAuthServiceCleanupExpiredTokens(t *testing.T) {
	app := createTestTokenApp(t, true)
	authService := as.New(app)
	ctx := app.createMockContext(nil)

	// Create some tokens - some expired, some not
	now := time.Now()
	expiredTime := now.Add(-1 * time.Hour)
	validTime := now.Add(1 * time.Hour)

	// Create expired token
	_, err := db.Builder[*fs.Token](app.db).Create(ctx, entity.New().
		Set("user_id", uint64(1)).
		Set("jti", "expired-token-jti").
		Set("expires_at", expiredTime))
	require.NoError(t, err)

	// Create valid token
	_, err = db.Builder[*fs.Token](app.db).Create(ctx, entity.New().
		Set("user_id", uint64(1)).
		Set("jti", "valid-token-jti").
		Set("expires_at", validTime))
	require.NoError(t, err)

	// Cleanup expired tokens
	deleted, err := authService.CleanupExpiredTokens(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, deleted)

	// Verify expired token is deleted
	_, err = db.Builder[*fs.Token](app.db).
		Where(db.EQ("jti", "expired-token-jti")).
		First(context.Background())
	assert.True(t, db.IsNotFound(err))

	// Verify valid token still exists
	_, err = db.Builder[*fs.Token](app.db).
		Where(db.EQ("jti", "valid-token-jti")).
		First(context.Background())
	require.NoError(t, err)
}

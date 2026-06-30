package auth_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/entity"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/logger"
	"github.com/fastschema/fastschema/pkg/auth"
	entdbadapter "github.com/fastschema/fastschema/pkg/entdbadapter"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// otpMockContext implements fs.Context for testing
type otpMockContext struct {
	context.Context
	args    map[string]string
	headers map[string]string
}

func (c *otpMockContext) TraceID() string                { return "" }
func (c *otpMockContext) User() *fs.User                 { return nil }
func (c *otpMockContext) Local(string, ...any) (val any) { return nil }
func (c *otpMockContext) Logger() logger.Logger          { return logger.CreateMockLogger(true) }
func (c *otpMockContext) Bind(any) error                 { return nil }
func (c *otpMockContext) SetArg(k, v string) string      { c.args[k] = v; return v }
func (c *otpMockContext) Args() map[string]string        { return c.args }
func (c *otpMockContext) Arg(name string, defaults ...string) string {
	if v, ok := c.args[name]; ok {
		return v
	}
	if len(defaults) > 0 {
		return defaults[0]
	}
	return ""
}
func (c *otpMockContext) ArgInt(name string, defaults ...int) int { return 0 }
func (c *otpMockContext) Header(key string, vals ...string) string {
	if v, ok := c.headers[key]; ok {
		return v
	}
	return ""
}
func (c *otpMockContext) Body() ([]byte, error)                               { return nil, nil }
func (c *otpMockContext) Payload() (*entity.Entity, error)                    { return nil, nil }
func (c *otpMockContext) BodyParser(out any) error                            { return nil }
func (c *otpMockContext) FormValue(key string, defaultValue ...string) string { return "" }
func (c *otpMockContext) Resource() *fs.Resource                              { return nil }
func (c *otpMockContext) AuthToken() string                                   { return "" }
func (c *otpMockContext) Next() error                                         { return nil }
func (c *otpMockContext) Result(...*fs.Result) *fs.Result                     { return nil }
func (c *otpMockContext) Files() ([]*fs.File, error)                          { return nil, nil }
func (c *otpMockContext) Redirect(string) error                               { return nil }
func (c *otpMockContext) Cookie(string, ...*fs.Cookie) string                 { return "" }
func (c *otpMockContext) WSClient() fs.WSClient                               { return nil }
func (c *otpMockContext) IP() string                                          { return "127.0.0.1" }

func createOTPProvider(config *testAppConfig) *auth.OTPProvider {
	if config.db == nil {
		schemasDir := utils.Must(os.MkdirTemp("", "schemas"))
		migrationsDir := utils.Must(os.MkdirTemp("", "migrations"))
		sb := utils.Must(schema.NewBuilderFromDir(schemasDir, fs.SystemSchemaTypes...))
		config.db = utils.Must(entdbadapter.NewTestClient(migrationsDir, sb))
	}

	if config.key == "" {
		config.key = testKey
	}

	var adminRoleID uuid.UUID
	if config.createData {
		roleModel, err := config.db.Model("role")
		if err == nil {
			// Check if roles exist
			roles, _ := roleModel.Query().Count(context.Background())
			if roles == 0 {
				adminRoleIDRaw, _ := roleModel.CreateFromJSON(context.Background(), `{"name": "admin"}`)
				adminRoleID = adminRoleIDRaw.(uuid.UUID)
				_, _ = roleModel.CreateFromJSON(context.Background(), `{"name": "user"}`)
			} else {
				// Get the admin role ID from existing role
				adminRole, _ := roleModel.Query().Where(db.EQ("name", "admin")).First(context.Background())
				if adminRole != nil {
					adminRoleID = adminRole.ID().(uuid.UUID)
				}
			}
		}

		userModel, err := config.db.Model("user")
		if err == nil {
			_, _ = userModel.CreateFromJSON(context.Background(), `{
				"username": "activeuser",
				"password": "activeuser",
				"email": "active@site.local",
				"provider": "local",
				"active": true,
				"roles": [{"id": "`+adminRoleID.String()+`"}]
			}`)

			_, _ = userModel.CreateFromJSON(context.Background(), `{
				"username": "inactiveuser",
				"password": "inactiveuser",
				"email": "inactive@site.local",
				"provider": "local",
				"active": false,
				"roles": [{"id": "`+adminRoleID.String()+`"}]
			}`)
		}
	}

	authProvider, _ := auth.NewOTPAuthProvider(fs.Map{}, "")
	otpProvider := authProvider.(*auth.OTPProvider)
	otpProvider.Init(
		func() db.Client { return config.db },
		func() string { return "testApp" },
		func(names ...string) fs.Mailer { return config.mailer },
		func() *fs.OTPConfig {
			return &fs.OTPConfig{
				Enabled:     true,
				Length:      6,
				Expiration:  300,
				MaxAttempts: 3,
			}
		},
	)

	return otpProvider
}

func TestOTPRequest(t *testing.T) {
	mailer := &MockMailer{}
	config := &testAppConfig{createData: true, mailer: mailer}
	provider := createOTPProvider(config)

	// Case 1: OTP is disabled
	{
		disabledProvider := &auth.OTPProvider{}
		disabledProvider.Init(
			func() db.Client { return config.db },
			func() string { return "testApp" },
			func(names ...string) fs.Mailer { return nil },
			func() *fs.OTPConfig { return &fs.OTPConfig{Enabled: false} },
		)
		resp, err := disabledProvider.RequestOTP(nil, &auth.OTPRequest{Email: "test@example.com"})
		assert.ErrorIs(t, err, auth.ERR_OTP_NOT_ENABLED)
		assert.Nil(t, resp)
	}

	// Case 2: Invalid email
	{
		resp, err := provider.RequestOTP(nil, &auth.OTPRequest{Email: "invalid"})
		assert.ErrorContains(t, err, auth.MSG_INVALID_EMAIL)
		assert.Nil(t, resp)
	}

	// Case 3: User checking error
	{
		assert.NoError(t, config.db.Close())
		c := &otpMockContext{Context: context.Background()}
		resp, err := provider.RequestOTP(c, &auth.OTPRequest{Email: "active@site.local"})
		assert.ErrorContains(t, err, auth.MSG_CHECKING_USER_ERROR)
		assert.Nil(t, resp)

		// Re-init DB for other tests
		config.db = nil
		provider = createOTPProvider(config)
	}

	// Case 4: User not found (should return success)
	{
		c := &otpMockContext{Context: context.Background()}
		resp, err := provider.RequestOTP(c, &auth.OTPRequest{Email: "notfound@example.com"})
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, auth.MSG_OTP_SENT, resp.Message)
		// No email should be sent
		assert.Empty(t, mailer.GetSentMails())
	}

	// Case 5: User inactive (should return success but no email)
	{
		mailer.Reset() // reset
		c := &otpMockContext{Context: context.Background()}
		resp, err := provider.RequestOTP(c, &auth.OTPRequest{Email: "inactive@site.local"})
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, auth.MSG_OTP_SENT, resp.Message)
		assert.Empty(t, mailer.GetSentMails())
	}

	// Case 6: Success (active user)
	{
		mailer.Reset()
		c := &otpMockContext{
			Context: context.Background(),
			headers: map[string]string{"User-Agent": "TestBrowser"},
			args:    map[string]string{"device_info": "TestDevice"},
		}

		resp, err := provider.RequestOTP(c, &auth.OTPRequest{Email: "active@site.local"})
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.NotEmpty(t, resp.SessionID)

		// Wait for goroutine to send email
		time.Sleep(100 * time.Millisecond)
		sentMails := mailer.GetSentMails()
		assert.Len(t, sentMails, 1)
		if len(sentMails) > 0 {
			assert.Equal(t, "active@site.local", sentMails[0].To[0])
		}

		// Verify session created
		sessionModel, err := config.db.Model("session")
		assert.NoError(t, err)

		sessionID, _ := uuid.Parse(resp.SessionID)
		sessionEntities, err := sessionModel.Query(db.EQ("id", sessionID)).Get(context.Background())
		assert.NoError(t, err)
		assert.Len(t, sessionEntities, 1)

		deviceInfo := sessionEntities[0].Get("device_info")
		assert.Equal(t, "TestBrowser", deviceInfo)
	}
}

func TestOTPVerify(t *testing.T) {
	mailer := &MockMailer{}
	config := &testAppConfig{createData: true, mailer: mailer}
	provider := createOTPProvider(config)

	// Case 1: OTP disabled
	{
		disabledProvider := &auth.OTPProvider{}
		disabledProvider.Init(
			func() db.Client { return config.db },
			func() string { return "testApp" },
			func(names ...string) fs.Mailer { return nil },
			func() *fs.OTPConfig { return &fs.OTPConfig{Enabled: false} },
		)
		user, err := disabledProvider.VerifyOTP(nil, &auth.OTPVerify{})
		assert.ErrorIs(t, err, auth.ERR_OTP_NOT_ENABLED)
		assert.Nil(t, user)
	}

	// Case 2: Invalid Input
	{
		// Empty SessionID
		user, err := provider.VerifyOTP(nil, &auth.OTPVerify{SessionID: "", OTP: "123"})
		assert.ErrorContains(t, err, auth.MSG_SESSION_ID_REQUIRED)
		assert.Nil(t, user)

		// Invalid UUID
		user, err = provider.VerifyOTP(nil, &auth.OTPVerify{SessionID: "invalid-uuid", OTP: "123"})
		assert.ErrorIs(t, err, auth.ERR_OTP_INVALID)
		assert.Nil(t, user)

		// Empty OTP
		user, err = provider.VerifyOTP(nil, &auth.OTPVerify{SessionID: uuid.NewString(), OTP: ""})
		assert.ErrorContains(t, err, auth.MSG_OTP_CODE_REQUIRED)
		assert.Nil(t, user)
	}

	// Session Setup
	activeUserEntity, _ := db.Builder[*fs.User](config.db).Where(db.EQ("email", "active@site.local")).First(context.Background())
	inactiveUserEntity, _ := db.Builder[*fs.User](config.db).Where(db.EQ("email", "inactive@site.local")).First(context.Background())
	require.NotNil(t, activeUserEntity)
	require.NotNil(t, inactiveUserEntity)

	activeUserID := activeUserEntity.ID
	inactiveUserID := inactiveUserEntity.ID

	createSession := func(userID uuid.UUID, otp string, attempts int, expiresAt time.Time) string {
		otpHash := utils.Must(auth.HashOTP(otp))
		sessionID := utils.Must(uuid.NewV7())
		sessionModel, _ := config.db.Model("session")
		_, err := sessionModel.Create(context.Background(), entity.New().
			Set("id", sessionID).
			Set("user_id", userID).
			Set("type", string(fs.SessionTypeOTPLogin)).
			Set("status", string(fs.SessionStatusPendingOTP)).
			Set("otp_hash", otpHash).
			Set("otp_attempts", attempts).
			Set("expires_at", expiresAt),
		)
		if err != nil {
			t.Fatal(err)
		}
		return sessionID.String()
	}

	// Case 3: Session Not Found
	{
		c := &otpMockContext{Context: context.Background()}
		user, err := provider.VerifyOTP(c, &auth.OTPVerify{SessionID: uuid.NewString(), OTP: "123"})
		assert.ErrorIs(t, err, auth.ERR_OTP_INVALID)
		assert.Nil(t, user)
	}

	// Case 4: Session Error (DB Error)
	{
		// Close DB temporarily
		config.db.Close()
		c := &otpMockContext{Context: context.Background()}
		user, err := provider.VerifyOTP(c, &auth.OTPVerify{SessionID: uuid.NewString(), OTP: "123"})
		assert.Error(t, err) // Should be DB error
		assert.Nil(t, user)

		// Restore DB
		config = &testAppConfig{createData: true, mailer: mailer}
		provider = createOTPProvider(config)
		activeUserEntity, _ = db.Builder[*fs.User](config.db).Where(db.EQ("email", "active@site.local")).First(context.Background())
		inactiveUserEntity, _ = db.Builder[*fs.User](config.db).Where(db.EQ("email", "inactive@site.local")).First(context.Background())
		activeUserID = activeUserEntity.ID
		inactiveUserID = inactiveUserEntity.ID
	}

	// Case 5: Session Expired
	{
		sessionID := createSession(activeUserID, "123456", 0, time.Now().Add(-1*time.Minute))
		c := &otpMockContext{Context: context.Background()}
		user, err := provider.VerifyOTP(c, &auth.OTPVerify{SessionID: sessionID, OTP: "123456"})
		assert.ErrorIs(t, err, auth.ERR_OTP_EXPIRED)
		assert.Nil(t, user)

		// Verify session inactive
		sessionModel, _ := config.db.Model("session")
		sessionIDUuid, _ := uuid.Parse(sessionID)
		session, _ := sessionModel.Query(db.EQ("id", sessionIDUuid)).First(context.Background())
		assert.Equal(t, string(fs.SessionStatusInactive), session.Get("status"))
	}

	// Case 6: Max Attempts Exceeded
	{
		sessionID := createSession(activeUserID, "123456", 3, time.Now().Add(5*time.Minute))
		c := &otpMockContext{Context: context.Background()}
		user, err := provider.VerifyOTP(c, &auth.OTPVerify{SessionID: sessionID, OTP: "123456"})
		assert.ErrorIs(t, err, auth.ERR_OTP_MAX_ATTEMPTS)
		assert.Nil(t, user)

		// Verify session inactive
		sessionModel, _ := config.db.Model("session")
		sessionIDUuid, _ := uuid.Parse(sessionID)
		session, _ := sessionModel.Query(db.EQ("id", sessionIDUuid)).First(context.Background())
		assert.Equal(t, string(fs.SessionStatusInactive), session.Get("status"))
	}

	// Case 7: Invalid OTP (Increment Attempts)
	{
		sessionID := createSession(activeUserID, "123456", 0, time.Now().Add(5*time.Minute))
		c := &otpMockContext{Context: context.Background()}
		user, err := provider.VerifyOTP(c, &auth.OTPVerify{SessionID: sessionID, OTP: "wrong"})
		assert.ErrorIs(t, err, auth.ERR_OTP_INVALID)
		assert.Nil(t, user)

		// Verify attempts incremented
		sessionModel, _ := config.db.Model("session")
		sessionIDUuid, _ := uuid.Parse(sessionID)
		session, _ := sessionModel.Query(db.EQ("id", sessionIDUuid)).First(context.Background())
		assert.Equal(t, 1, session.Get("otp_attempts"))
	}

	// Case 8: User Not Found (Deleted after session created)
	{
		c := &otpMockContext{Context: context.Background()}
		// Get role ID for user creation
		roleModel, _ := config.db.Model("role")
		roleEntity, _ := roleModel.Query(db.EQ("name", "admin")).First(context.Background())
		roleIDValue := roleEntity.Get("id").(uuid.UUID)

		// Create a user and delete it
		userModel, _ := config.db.Model("user")
		id, err := userModel.CreateFromJSON(context.Background(), `{
			"username": "todelete",
			"password": "todelete",
			"email": "todelete@site.local",
			"provider": "local",
			"active": true,
			"roles": [{"id": "`+roleIDValue.String()+`"}]
		}`)
		require.NoError(t, err)

		// Get ID safely
		var userID uuid.UUID
		if uID, ok := id.(uuid.UUID); ok {
			userID = uID
		} else if uIDStr, ok := id.(string); ok {
			userID, err = uuid.Parse(uIDStr)
			require.NoError(t, err)
		}

		sessionID := createSession(userID, "123456", 0, time.Now().Add(5*time.Minute))

		userModel.Mutation().Where(db.EQ("id", userID)).Delete(context.Background())

		user, err := provider.VerifyOTP(c, &auth.OTPVerify{SessionID: sessionID, OTP: "123456"})
		assert.ErrorContains(t, err, auth.MSG_OTP_USER_NOT_FOUND)
		assert.Nil(t, user)
	}

	// Case 9: User Inactive
	{
		sessionID := createSession(inactiveUserID, "123456", 0, time.Now().Add(5*time.Minute))
		c := &otpMockContext{Context: context.Background()}
		user, err := provider.VerifyOTP(c, &auth.OTPVerify{SessionID: sessionID, OTP: "123456"})
		assert.ErrorContains(t, err, auth.MSG_USER_IS_INACTIVE)
		assert.Nil(t, user)
	}

	// Case 10: Success
	{
		sessionID := createSession(activeUserID, "123456", 0, time.Now().Add(5*time.Minute))
		c := &otpMockContext{Context: context.Background()}
		user, err := provider.VerifyOTP(c, &auth.OTPVerify{SessionID: sessionID, OTP: "123456"})
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, activeUserID, user.ID)

		// Verify session deleted
		sessionModel, _ := config.db.Model("session")
		sessionIDUuid, _ := uuid.Parse(sessionID)
		session, err := sessionModel.Query(db.EQ("id", sessionIDUuid)).First(context.Background())
		assert.Error(t, err)
		assert.True(t, db.IsNotFound(err) || session == nil)
	}
}

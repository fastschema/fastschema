package auth

import (
	"strings"
	"time"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/entity"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/errors"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/google/uuid"
)

const ProviderOTP = "otp"

func init() {
	fs.RegisterAuthProviderMaker(ProviderOTP, NewOTPAuthProvider)
}

// OTPProvider represents the OTP passwordless authentication provider.
type OTPProvider struct {
	db        func() db.Client
	appName   func() string
	mailer    func(names ...string) fs.Mailer
	otpConfig func() *fs.OTPConfig
}

// NewOTPAuthProvider creates a new OTP auth provider
func NewOTPAuthProvider(config fs.Map, redirectURL string) (fs.AuthProvider, error) {
	return &OTPProvider{}, nil
}

// Init initializes the OTP provider with dependencies
func (op *OTPProvider) Init(
	db func() db.Client,
	appName func() string,
	mailer func(names ...string) fs.Mailer,
	otpConfig func() *fs.OTPConfig,
) {
	op.db = db
	op.appName = appName
	op.mailer = mailer
	op.otpConfig = otpConfig
}

// Name returns the provider name
func (op *OTPProvider) Name() string {
	return ProviderOTP
}

// Login is not used for OTP provider
func (op *OTPProvider) Login(c fs.Context) (any, error) {
	return nil, nil
}

// Callback is not used for OTP provider
func (op *OTPProvider) Callback(c fs.Context) (*fs.User, error) {
	return nil, nil
}

// VerifyIDToken is not used for OTP provider
func (op *OTPProvider) VerifyIDToken(c fs.Context, t fs.IDToken) (*fs.User, error) {
	return nil, nil
}

// IsEnabled checks if OTP is enabled
func (op *OTPProvider) IsEnabled() bool {
	if op.otpConfig == nil {
		return false
	}
	config := op.otpConfig()
	return config != nil && config.Enabled
}

// RequestOTP handles the OTP request for passwordless login
func (op *OTPProvider) RequestOTP(c fs.Context, req *OTPRequest) (*OTPResponse, error) {
	// Check if OTP is enabled
	if !op.IsEnabled() {
		return nil, ERR_OTP_NOT_ENABLED
	}

	otpConfig := op.otpConfig()

	// Validate email
	email := strings.TrimSpace(strings.ToLower(req.Email))
	if email == "" || !utils.IsValidEmail(email) {
		return nil, errors.UnprocessableEntity(MSG_INVALID_EMAIL)
	}

	// Check if user exists with this email (regardless of provider)
	user, err := db.Builder[*fs.User](op.db()).
		Where(db.EQ("email", email)).
		Select("id", "email", "active").
		First(c)
	if err != nil && !db.IsNotFound(err) {
		c.Logger().Errorf("Error checking user: %v", err)
		return nil, errors.InternalServerError(MSG_CHECKING_USER_ERROR)
	}

	// For security, always return success message even if user doesn't exist
	sessionID, err := uuid.NewV7()
	if err != nil {
		c.Logger().Errorf("Error generating session ID: %v", err)
		return nil, errors.InternalServerError(MSG_OTP_SESSION_CREATE_ERROR)
	}
	if user == nil {
		return &OTPResponse{
			Message:   MSG_OTP_SENT,
			SessionID: sessionID.String(),
			ExpiresIn: otpConfig.GetExpiration(),
		}, nil
	}

	// Check if user is active
	if !user.Active {
		return &OTPResponse{
			Message:   MSG_OTP_SENT,
			SessionID: sessionID.String(),
			ExpiresIn: otpConfig.GetExpiration(),
		}, nil
	}

	// Generate OTP
	otp, err := GenerateOTP(otpConfig.GetLength())
	if err != nil {
		c.Logger().Errorf("Error generating OTP: %v", err)
		return nil, errors.InternalServerError(MSG_OTP_GENERATION_ERROR)
	}

	// Hash OTP for storage
	otpHash, err := HashOTP(otp)
	if err != nil {
		c.Logger().Errorf("Error hashing OTP: %v", err)
		return nil, errors.InternalServerError(MSG_OTP_GENERATION_ERROR)
	}

	// Calculate expiration
	now := time.Now()
	expiresAt := now.Add(time.Duration(otpConfig.GetExpiration()) * time.Second)

	// Get device info
	deviceInfo := c.Header("User-Agent")
	if deviceInfo == "" {
		deviceInfo = c.Arg("device_info")
	}

	session, err := db.Builder[*fs.Session](op.db()).Create(c, entity.New().
		Set("id", sessionID).
		Set("user_id", user.ID).
		Set("type", string(fs.SessionTypeOTPLogin)).
		Set("status", string(fs.SessionStatusPendingOTP)).
		Set("otp_hash", otpHash).
		Set("otp_attempts", 0).
		Set("device_info", deviceInfo).
		Set("ip_address", c.IP()).
		Set("expires_at", expiresAt).
		Set("last_activity_at", now),
	)
	if err != nil {
		c.Logger().Errorf("Error creating OTP session: %v", err)
		return nil, errors.InternalServerError(MSG_OTP_SESSION_CREATE_ERROR)
	}

	// Send OTP email
	appName := op.appName()
	if appName == "" {
		appName = "FastSchema"
	}

	expirationMinutes := max(otpConfig.GetExpiration()/60, 1)
	mail := CreateOTPEmail(appName, email, otp, expirationMinutes)

	go func() {
		mailer := op.mailer()
		if mailer == nil {
			c.Logger().Error(MSG_MAILER_NOT_SET)
			return
		}
		if err := mailer.Send(mail); err != nil {
			c.Logger().Errorf("Error sending OTP email: %v", err)
		}
	}()

	return &OTPResponse{
		Message:   MSG_OTP_SENT,
		SessionID: session.ID.String(),
		ExpiresIn: otpConfig.GetExpiration(),
	}, nil
}

// VerifyOTP handles OTP verification for passwordless login
// Returns the authenticated user on success
func (op *OTPProvider) VerifyOTP(c fs.Context, req *OTPVerify) (*fs.User, error) {
	// Check if OTP is enabled
	if !op.IsEnabled() {
		return nil, ERR_OTP_NOT_ENABLED
	}

	otpConfig := op.otpConfig()

	// Validate input
	sessionID := strings.TrimSpace(req.SessionID)
	otp := strings.TrimSpace(req.OTP)

	if sessionID == "" {
		return nil, errors.UnprocessableEntity(MSG_SESSION_ID_REQUIRED)
	}

	// Parse session ID as UUID
	sessionUUID, err := uuid.Parse(sessionID)
	if err != nil {
		return nil, ERR_OTP_INVALID
	}

	if otp == "" {
		return nil, errors.UnprocessableEntity(MSG_OTP_CODE_REQUIRED)
	}

	// Find the OTP session by ID
	session, err := db.Builder[*fs.Session](op.db()).
		Where(db.EQ("id", sessionUUID)).
		Where(db.EQ("type", string(fs.SessionTypeOTPLogin))).
		Where(db.EQ("status", string(fs.SessionStatusPendingOTP))).
		First(c)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, ERR_OTP_INVALID
		}
		c.Logger().Errorf("Error finding OTP session: %v", err)
		return nil, errors.InternalServerError("Error verifying OTP")
	}

	// Check if session is expired
	if session.ExpiresAt != nil && session.ExpiresAt.Before(time.Now()) {
		// Mark session as inactive
		_, _ = db.Builder[*fs.Session](op.db()).
			Where(db.EQ("id", session.ID)).
			Update(c, entity.New().Set("status", string(fs.SessionStatusInactive)))
		return nil, ERR_OTP_EXPIRED
	}

	// Check max attempts
	maxAttempts := otpConfig.GetMaxAttempts()
	if session.OTPAttempts >= maxAttempts {
		// Mark session as inactive
		_, _ = db.Builder[*fs.Session](op.db()).
			Where(db.EQ("id", session.ID)).
			Update(c, entity.New().Set("status", string(fs.SessionStatusInactive)))
		return nil, ERR_OTP_MAX_ATTEMPTS
	}

	// Verify OTP
	if !VerifyOTP(otp, session.OTPHash) {
		// Increment attempt counter
		_, _ = db.Builder[*fs.Session](op.db()).
			Where(db.EQ("id", session.ID)).
			Update(c, entity.New().Set("otp_attempts", session.OTPAttempts+1))
		return nil, ERR_OTP_INVALID
	}

	// OTP is valid - delete the OTP session
	_, _ = db.Builder[*fs.Session](op.db()).
		Where(db.EQ("id", session.ID)).
		Delete(c)

	// Get the user
	user, err := db.Builder[*fs.User](op.db()).
		Where(db.EQ("id", session.UserID)).
		Select("id", "username", "email", "provider", "provider_id", "provider_username", "active", "roles").
		First(c)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, errors.Unauthorized(MSG_OTP_USER_NOT_FOUND)
		}
		c.Logger().Errorf("Error getting user: %v", err)
		return nil, errors.InternalServerError("Error getting user")
	}

	// Check if user is still active
	if !user.Active {
		return nil, errors.Unauthorized(MSG_USER_IS_INACTIVE)
	}

	return user, nil
}

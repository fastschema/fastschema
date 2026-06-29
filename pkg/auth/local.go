package auth

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/entity"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/errors"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/google/uuid"
)

const ProviderLocal = "local"

// LocalProvider represents the local authentication provider.
//
// config:
// activationMethod: auto, manual, email
//
//	auto: user is activated automatically
//	manual: user is activated manually by admin
//	email: user is activated by email
//
// verificationMethod: link, otp
//
//	link: verification via email link with token (default)
//	otp: verification via one-time password sent to email
type LocalProvider struct {
	db                  func() db.Client
	appKey              func() string
	appName             func() string
	appBaseURL          func() string
	mailer              func(names ...string) fs.Mailer
	config              fs.Map
	activationMethod    string
	activationURL       string
	recoveryURL         string
	verificationMethod  string // "link" or "otp"
	otpConfig           func() *fs.OTPConfig
	jwtCustomClaimsFunc func() fs.JwtCustomClaimsFunc
	emailTemplates      func() *fs.EmailTemplates
	// fireRegister runs the PreUserRegister hook chain (built-in policy + custom
	// hooks) before a user row is created. Nil when not wired (e.g. tests).
	fireRegister func(ctx context.Context, in *fs.RegistrationInput) error
	// registrationPolicy exposes the opt-in signup policy so LocalLogin can apply
	// the same email normalization used at registration. Nil when unset.
	registrationPolicy func() *fs.RegistrationPolicy
}

func NewLocalAuthProvider(config fs.Map, redirectURL string) (fs.AuthProvider, error) {
	la := &LocalProvider{
		config:             config,
		activationMethod:   fs.MapValue(config, "activation_method", "manual"),
		activationURL:      fs.MapValue(config, "activation_url", ""),
		recoveryURL:        fs.MapValue(config, "recovery_url", ""),
		verificationMethod: fs.MapValue(config, "verification_method", "link"),
	}

	return la, nil
}

func (la *LocalProvider) Init(
	db func() db.Client,
	appKey func() string,
	appName func() string,
	appBaseURL func() string,
	mailer func(names ...string) fs.Mailer,
	jwtCustomClaimsFunc func() fs.JwtCustomClaimsFunc,
	otpConfig func() *fs.OTPConfig,
	emailTemplates func() *fs.EmailTemplates,
	fireRegister func(ctx context.Context, in *fs.RegistrationInput) error,
	registrationPolicy func() *fs.RegistrationPolicy,
) {
	la.db = db
	la.appKey = appKey
	la.appName = appName
	la.mailer = mailer
	la.appBaseURL = appBaseURL
	la.jwtCustomClaimsFunc = jwtCustomClaimsFunc
	la.otpConfig = otpConfig
	la.emailTemplates = emailTemplates
	la.fireRegister = fireRegister
	la.registrationPolicy = registrationPolicy

	if la.activationURL == "" {
		la.activationURL = appBaseURL() + "/auth/local/activate"
	}

	if la.recoveryURL == "" {
		la.recoveryURL = appBaseURL() + "/auth/local/recover"
	}
}

// IsOTPVerification returns true if OTP verification method is enabled
func (la *LocalProvider) IsOTPVerification() bool {
	return la.verificationMethod == "otp"
}

// getOTPConfig returns the OTP config with defaults
func (la *LocalProvider) getOTPConfig() *fs.OTPConfig {
	if la.otpConfig == nil {
		return &fs.OTPConfig{
			Enabled:     true,
			Length:      6,
			Expiration:  300,
			MaxAttempts: 3,
		}
	}
	config := la.otpConfig()
	if config == nil {
		return &fs.OTPConfig{
			Enabled:     true,
			Length:      6,
			Expiration:  300,
			MaxAttempts: 3,
		}
	}
	return config
}

func (la *LocalProvider) Name() string {
	return ProviderLocal
}

func (la *LocalProvider) Login(c fs.Context) (_ any, err error) {
	return nil, nil
}

func (la *LocalProvider) Callback(c fs.Context) (user *fs.User, err error) {
	return nil, nil
}

func (la *LocalProvider) VerifyIDToken(c fs.Context, t fs.IDToken) (user *fs.User, err error) {
	return nil, nil
}

func (la *LocalProvider) Register(c fs.Context, payload *Register) (*Activation, error) {
	// Run the registration hook chain (built-in policy + custom hooks) FIRST so
	// the email is normalized before the uniqueness check below. Otherwise a
	// case/format variant (e.g. user@GMAIL.com vs user@gmail.com) slips past the
	// app-layer dedup and only trips the DB unique index — surfacing as a 500
	// instead of a clean "user already exists". Hooks may also mutate the
	// username; apply the result back so the persisted entity matches login.
	if la.fireRegister != nil {
		in := &fs.RegistrationInput{
			Email:    payload.Email,
			Username: payload.Username,
			Provider: la.Name(),
			IsOAuth:  false,
		}
		if err := la.fireRegister(c, in); err != nil {
			return nil, err
		}
		payload.Email = in.Email
		payload.Username = in.Username
	}

	if err := ValidateRegisterData(c, c.Logger(), la.db(), payload); err != nil {
		return nil, err
	}

	// Resolve the User role by name before opening the transaction so we can fail early.
	// Role name is the stable identifier; IDs are random per-deployment.
	userRole, err := db.Builder[*fs.Role](la.db()).
		Where(db.EQ("name", fs.RoleUser.Name)).
		First(c)
	if err != nil || userRole == nil {
		c.Logger().Errorf("role '%s' not found in database: %v", fs.RoleUser.Name, err)
		return nil, ERR_SAVE_USER
	}

	userEntity := payload.Entity(la.activationMethod, la.Name(), userRole.ID)
	if err := db.WithTx(la.db(), c, func(tx db.Client) error {
		user, err := db.Builder[*fs.User](tx).Create(c, userEntity)
		if err != nil {
			c.Logger().Errorf(MSG_USER_SAVE_ERROR+": %w", err)
			return ERR_SAVE_USER
		}

		if _, err = db.Builder[*fs.User](tx).
			Where(db.EQ("id", user.ID)).
			Update(c, entity.New().Set("provider_id", user.ID.String())); err != nil {
			c.Logger().Errorf(MSG_USER_UPDATE_PROVIDER_ID_ERROR, err)
			return ERR_SAVE_USER
		}

		user.ProviderID = user.ID.String()
		email, err := CreateActivationEmail(la, user)
		if err != nil {
			c.Logger().Errorf(MSG_CREATE_ACTIVATION_MAIL_ERROR, err)
			return ERR_SAVE_USER
		}

		go SendConfirmationEmail(la, c.Logger(), email)

		return nil
	}); err != nil {
		c.Logger().Errorf(MSG_USER_SAVE_ERROR+": %w", err)
		return nil, ERR_SAVE_USER
	}

	return &Activation{Activation: la.activationMethod}, nil
}

func (la *LocalProvider) Activate(c fs.Context, data *Confirmation) (*Activation, error) {
	var userID uuid.UUID
	var err error

	switch {
	case data.IsTokenBased():
		userID, err = ValidateConfirmationToken(data.Token, la.appKey())
		if err != nil {
			err = fmt.Errorf(MSG_INVALID_TOKEN+": %w", err)
			c.Logger().Error(err)
			return nil, err
		}
	case data.IsOTPBased():
		userID, err = la.verifyOTPSession(c, data.SessionID, data.OTP, fs.SessionTypeActivation, true)
		if err != nil {
			return nil, err
		}
	default:
		return nil, ERR_INVALID_TOKEN
	}

	var count int
	if count, err = db.Builder[*fs.User](la.db()).
		Where(db.EQ("id", userID)).
		Where(db.EQ("active", true)).
		Count(c); err != nil {
		c.Logger().Errorf(MSG_CHECKING_USER_ERROR+": %w", err)
		return nil, errors.InternalServerError(MSG_CHECKING_USER_ERROR)
	}
	if count > 0 {
		return nil, ERR_USER_ALREADY_ACTIVE
	}

	if _, err = db.Builder[*fs.User](la.db()).
		Where(db.EQ("id", userID)).
		Where(db.EQ("active", false)).
		Update(c, entity.New().Set("active", true)); err != nil {
		c.Logger().Errorf(MSG_USER_ACTIVATION_ERROR+": %w", err)
		return nil, errors.BadRequest(MSG_USER_ACTIVATION_ERROR)
	}

	return &Activation{Activation: "activated"}, nil
}

func (la *LocalProvider) SendActivationLink(c fs.Context, data *SendActivation) (*Activation, error) {
	if la.activationMethod != "email" {
		return nil, errors.BadRequest()
	}

	// OTP flow: request activation OTP via email
	if la.IsOTPVerification() && data.Email != "" {
		return la.sendActivationOTP(c, data.Email)
	}

	// Link flow: resend activation link
	// Only send the new activation link if:
	// - The confirmation token is valid
	// - The confirmation token is expired
	userID, err := ValidateConfirmationToken(data.Token, la.appKey())
	if err == nil || !errors.Is(err, ERR_TOKEN_EXPIRED) {
		return nil, ERR_INVALID_TOKEN
	}

	user, err := db.Builder[*fs.User](la.db()).
		Where(db.EQ("id", userID)).
		Where(db.EQ("active", false)).
		Select("id", "username", "email").
		First(c)
	if err != nil {
		return nil, ERR_INVALID_TOKEN
	}

	email, err := CreateActivationEmail(la, user)
	if err != nil {
		c.Logger().Error(MSG_CREATE_ACTIVATION_MAIL_ERROR, err)
		return nil, errors.BadRequest(MSG_CREATE_ACTIVATION_MAIL_ERROR)
	}

	go SendConfirmationEmail(la, c.Logger(), email)

	return &Activation{Activation: la.activationMethod}, nil
}

// sendActivationOTP sends an OTP for account activation
func (la *LocalProvider) sendActivationOTP(c fs.Context, email string) (*Activation, error) {
	otpConfig := la.getOTPConfig()

	// Validate email
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" || !utils.IsValidEmail(email) {
		return nil, errors.UnprocessableEntity(MSG_INVALID_EMAIL)
	}

	// Find inactive user with this email
	user, err := db.Builder[*fs.User](la.db()).
		Where(db.EQ("email", email)).
		Where(db.EQ("provider", la.Name())).
		Where(db.EQ("active", false)).
		Select("id", "email").
		First(c)
	if err != nil && !db.IsNotFound(err) {
		c.Logger().Errorf("Error checking user: %v", err)
		return nil, errors.InternalServerError(MSG_CHECKING_USER_ERROR)
	}

	// For security, always return success message even if user doesn't exist or is already active
	dummySessionID, _ := uuid.NewV7()
	if user == nil {
		return &Activation{
			Activation: la.activationMethod,
			SessionID:  dummySessionID.String(),
			ExpiresIn:  otpConfig.GetExpiration(),
		}, nil
	}

	// Create OTP session
	sessionID, otp, err := la.createOTPSession(c, user.ID, fs.SessionTypeActivation)
	if err != nil {
		c.Logger().Errorf("Error creating activation OTP session: %v", err)
		return nil, err
	}

	// Send OTP email
	appName := la.appName()
	if appName == "" {
		appName = "FastSchema"
	}
	expirationMinutes := max(otpConfig.GetExpiration()/60, 1)
	mail := CreateActivationOTPEmail(appName, email, otp, expirationMinutes)
	la.sendOTPEmail(c, mail)

	return &Activation{
		Activation: la.activationMethod,
		SessionID:  sessionID,
		ExpiresIn:  otpConfig.GetExpiration(),
	}, nil
}

// LocalLogin performs local login with username/email and password
func (la *LocalProvider) LocalLogin(c fs.Context, payload *LoginData) (*fs.User, error) {
	if payload == nil || strings.TrimSpace(payload.Login) == "" || payload.Password == "" {
		return nil, errors.UnprocessableEntity(MSG_INVALID_LOGIN_OR_PASSWORD)
	}

	login := strings.TrimSpace(payload.Login)
	// Apply the same email normalization used at registration so a normalized
	// stored email still matches at login. Only when login looks like an email
	// (usernames are left untouched) and the policy enables normalization.
	if strings.Contains(login, "@") && la.registrationPolicy != nil {
		if p := la.registrationPolicy(); p != nil && p.NormalizeEmail {
			login = NormalizeEmail(login)
		}
	}
	c.Local("keeppassword", "true")
	user, err := db.Builder[*fs.User](la.db()).
		Where(db.Or(
			db.EQ("username", login),
			db.EQ("email", login),
		)).
		Select(
			"id",
			"username",
			"email",
			"password",
			"provider",
			"provider_id",
			"provider_username",
			"active",
			"roles",
		).
		First(c)
	if err != nil && !db.IsNotFound(err) {
		c.Logger().Error(err)
		return nil, errors.InternalServerError(MSG_CHECKING_USER_ERROR)
	}

	if user == nil {
		return nil, errors.UnprocessableEntity(MSG_INVALID_LOGIN_OR_PASSWORD)
	}

	if !user.Active {
		return nil, errors.Unauthorized(MSG_USER_IS_INACTIVE)
	}

	if err := utils.CheckHash(payload.Password, user.Password); err != nil {
		return nil, errors.UnprocessableEntity(MSG_INVALID_LOGIN_OR_PASSWORD)
	}

	return user, nil
}

func (la *LocalProvider) Recover(c fs.Context, data *Recovery) (*Activation, error) {
	if !utils.IsValidEmail(data.Email) {
		return nil, errors.UnprocessableEntity(MSG_INVALID_EMAIL)
	}

	user, err := db.Builder[*fs.User](la.db()).
		Where(db.EQ("email", data.Email)).
		Where(db.EQ("provider", la.Name())).
		Select("id", "email", "username").
		First(c)
	if err != nil && !db.IsNotFound(err) {
		c.Logger().Error(err)
		return nil, errors.InternalServerError(MSG_CHECKING_USER_ERROR)
	}

	otpConfig := la.getOTPConfig()

	// For security, always return success even if user doesn't exist
	if user == nil {
		if la.IsOTPVerification() {
			dummySessionID, _ := uuid.NewV7()
			return &Activation{
				Activation: "sent",
				SessionID:  dummySessionID.String(),
				ExpiresIn:  otpConfig.GetExpiration(),
			}, nil
		}
		return &Activation{Activation: "sent"}, nil
	}

	// OTP flow
	if la.IsOTPVerification() {
		return la.sendRecoveryOTP(c, user)
	}

	// Link flow (existing behavior)
	email, err := CreateRecoveryEmail(la, user)
	if err != nil {
		c.Logger().Errorf(MSG_CREATEP_RECOVERY_MAIL_ERROR+": %w", err)
		return nil, errors.BadRequest(MSG_CREATEP_RECOVERY_MAIL_ERROR)
	}

	go SendConfirmationEmail(la, c.Logger(), email)

	return &Activation{Activation: "sent"}, nil
}

// sendRecoveryOTP sends an OTP for password recovery
func (la *LocalProvider) sendRecoveryOTP(c fs.Context, user *fs.User) (*Activation, error) {
	otpConfig := la.getOTPConfig()

	// Create OTP session
	sessionID, otp, err := la.createOTPSession(c, user.ID, fs.SessionTypeRecovery)
	if err != nil {
		c.Logger().Errorf("Error creating recovery OTP session: %v", err)
		return nil, err
	}

	// Send OTP email
	appName := la.appName()
	if appName == "" {
		appName = "FastSchema"
	}
	expirationMinutes := max(otpConfig.GetExpiration()/60, 1)
	mail := CreateRecoveryOTPEmail(appName, user.Email, otp, expirationMinutes)
	la.sendOTPEmail(c, mail)

	return &Activation{
		Activation: "sent",
		SessionID:  sessionID,
		ExpiresIn:  otpConfig.GetExpiration(),
	}, nil
}

func (la *LocalProvider) RecoverCheck(c fs.Context, data *Confirmation) (*Activation, error) {
	if data.IsTokenBased() {
		var emptyUUID uuid.UUID
		userID, err := ValidateConfirmationToken(data.Token, la.appKey())
		if err != nil {
			return nil, err
		}
		return &Activation{Activation: "valid", Verified: userID != emptyUUID}, nil
	}

	if data.IsOTPBased() {
		_, err := la.verifyOTPSession(c, data.SessionID, data.OTP, fs.SessionTypeRecovery, false)
		if err != nil {
			return nil, err
		}
		// Return the same session ID - it's now in 'verified' status
		return &Activation{
			Activation: "verified",
			SessionID:  data.SessionID,
			Verified:   true,
		}, nil
	}

	return nil, ERR_INVALID_TOKEN
}

func (la *LocalProvider) ResetPassword(c fs.Context, data *ResetPassword) (_ bool, err error) {
	var userID uuid.UUID

	switch {
	case data.Token != "":
		userID, err = ValidateConfirmationToken(data.Token, la.appKey())
		if err != nil {
			return false, err
		}
	case data.SessionID != "":
		userID, err = la.getUserFromVerifiedSession(c, data.SessionID, fs.SessionTypeRecovery)
		if err != nil {
			return false, err
		}
	default:
		return false, ERR_INVALID_TOKEN
	}

	if data.Password == "" || data.ConfirmPassword == "" || data.Password != data.ConfirmPassword {
		return false, errors.UnprocessableEntity(MSG_INVALID_PASSWORD)
	}

	if _, err := db.Builder[*fs.User](la.db()).
		Where(db.EQ("id", userID)).
		Update(c, entity.New().Set("password", data.Password)); err != nil {
		c.Logger().Errorf(MSG_USER_SAVE_ERROR+": %w", err)
		return false, ERR_SAVE_USER
	}

	// If session-based, mark session as inactive
	if data.SessionID != "" {
		sessionUUID, _ := uuid.Parse(data.SessionID)
		_, _ = db.Builder[*fs.Session](la.db()).
			Where(db.EQ("id", sessionUUID)).
			Update(c, entity.New().Set("status", string(fs.SessionStatusInactive)))
	}

	return true, nil
}

// invalidatePreviousSessions invalidates all pending OTP sessions for a user
func (la *LocalProvider) invalidatePreviousSessions(
	c fs.Context,
	userID uuid.UUID,
	sessionType fs.SessionType,
) error {
	_, err := db.Builder[*fs.Session](la.db()).
		Where(db.EQ("user_id", userID)).
		Where(db.EQ("type", string(sessionType))).
		Where(db.Or(
			db.EQ("status", string(fs.SessionStatusPendingOTP)),
			db.EQ("status", string(fs.SessionStatusVerified)),
		)).
		Update(c, entity.New().Set("status", string(fs.SessionStatusInactive)))
	return err
}

// createOTPSession creates a new OTP session after invalidating previous ones
func (la *LocalProvider) createOTPSession(
	c fs.Context,
	userID uuid.UUID,
	sessionType fs.SessionType,
) (sessionID string, otp string, err error) {
	otpConfig := la.getOTPConfig()

	// 1. Invalidate previous sessions
	_ = la.invalidatePreviousSessions(c, userID, sessionType)

	// 2. Generate OTP
	otp, err = GenerateOTP(otpConfig.GetLength())
	if err != nil {
		return "", "", errors.InternalServerError(MSG_OTP_GENERATION_ERROR)
	}

	// 3. Hash OTP
	otpHash, err := HashOTP(otp)
	if err != nil {
		return "", "", errors.InternalServerError(MSG_OTP_GENERATION_ERROR)
	}

	// 4. Create session
	sessionUUID, err := uuid.NewV7()
	if err != nil {
		return "", "", errors.InternalServerError(MSG_OTP_SESSION_CREATE_ERROR)
	}

	now := time.Now()
	expiresAt := now.Add(time.Duration(otpConfig.GetExpiration()) * time.Second)

	// Get device info
	deviceInfo := c.Header("User-Agent")
	if deviceInfo == "" {
		deviceInfo = c.Arg("device_info")
	}

	_, err = db.Builder[*fs.Session](la.db()).Create(c, entity.New().
		Set("id", sessionUUID).
		Set("user_id", userID).
		Set("type", string(sessionType)).
		Set("status", string(fs.SessionStatusPendingOTP)).
		Set("otp_hash", otpHash).
		Set("otp_attempts", 0).
		Set("device_info", deviceInfo).
		Set("ip_address", c.IP()).
		Set("expires_at", expiresAt).
		Set("last_activity_at", now),
	)
	if err != nil {
		return "", "", errors.InternalServerError(MSG_OTP_SESSION_CREATE_ERROR)
	}

	return sessionUUID.String(), otp, nil
}

// verifyOTPSession verifies an OTP and returns the user ID
// For activation: deletes session on success
// For recovery: updates status to 'verified' on success
func (la *LocalProvider) verifyOTPSession(
	c fs.Context,
	sessionID string,
	otp string,
	sessionType fs.SessionType,
	deleteOnSuccess bool,
) (userID uuid.UUID, err error) {
	otpConfig := la.getOTPConfig()
	var emptyUUID uuid.UUID

	// 1. Parse session UUID
	sessionUUID, err := uuid.Parse(sessionID)
	if err != nil {
		return emptyUUID, ERR_INVALID_TOKEN
	}

	// 2. Find session
	session, err := db.Builder[*fs.Session](la.db()).
		Where(db.EQ("id", sessionUUID)).
		Where(db.EQ("type", string(sessionType))).
		Where(db.EQ("status", string(fs.SessionStatusPendingOTP))).
		First(c)
	if err != nil {
		if db.IsNotFound(err) {
			return emptyUUID, ERR_INVALID_TOKEN
		}
		c.Logger().Errorf("Error finding OTP session: %v", err)
		return emptyUUID, errors.InternalServerError("Error verifying OTP")
	}

	// 3. Check expiration
	if session.ExpiresAt != nil && session.ExpiresAt.Before(time.Now()) {
		la.markSessionInactive(c, session.ID)
		return emptyUUID, ERR_OTP_EXPIRED
	}

	// 4. Check max attempts
	if session.OTPAttempts >= otpConfig.GetMaxAttempts() {
		la.markSessionInactive(c, session.ID)
		return emptyUUID, ERR_OTP_MAX_ATTEMPTS
	}

	// 5. Verify OTP
	if !VerifyOTP(otp, session.OTPHash) {
		// Increment attempts
		_, _ = db.Builder[*fs.Session](la.db()).
			Where(db.EQ("id", session.ID)).
			Update(c, entity.New().Set("otp_attempts", session.OTPAttempts+1))
		return emptyUUID, ERR_OTP_INVALID
	}

	// 6. Success - update or delete session
	if deleteOnSuccess {
		_, _ = db.Builder[*fs.Session](la.db()).
			Where(db.EQ("id", session.ID)).
			Delete(c)
	} else {
		_, _ = db.Builder[*fs.Session](la.db()).
			Where(db.EQ("id", session.ID)).
			Update(c, entity.New().Set("status", string(fs.SessionStatusVerified)))
	}

	return session.UserID, nil
}

// getUserFromVerifiedSession retrieves the user ID from a verified OTP session
func (la *LocalProvider) getUserFromVerifiedSession(
	c fs.Context,
	sessionID string,
	sessionType fs.SessionType,
) (uuid.UUID, error) {
	var emptyUUID uuid.UUID
	sessionUUID, err := uuid.Parse(sessionID)
	if err != nil {
		return emptyUUID, ERR_INVALID_TOKEN
	}

	session, err := db.Builder[*fs.Session](la.db()).
		Where(db.EQ("id", sessionUUID)).
		Where(db.EQ("type", string(sessionType))).
		Where(db.EQ("status", string(fs.SessionStatusVerified))).
		First(c)
	if err != nil {
		if db.IsNotFound(err) {
			return emptyUUID, errors.BadRequest(MSG_OTP_VERIFICATION_REQUIRED)
		}
		return emptyUUID, errors.InternalServerError("Error verifying session")
	}

	// Check expiration
	if session.ExpiresAt != nil && session.ExpiresAt.Before(time.Now()) {
		la.markSessionInactive(c, session.ID)
		return emptyUUID, ERR_OTP_EXPIRED
	}

	return session.UserID, nil
}

// markSessionInactive marks a session as inactive
func (la *LocalProvider) markSessionInactive(c fs.Context, sessionID uuid.UUID) {
	_, _ = db.Builder[*fs.Session](la.db()).
		Where(db.EQ("id", sessionID)).
		Update(c, entity.New().Set("status", string(fs.SessionStatusInactive)))
}

// sendOTPEmail sends an OTP email asynchronously
func (la *LocalProvider) sendOTPEmail(c fs.Context, email *fs.Mail) {
	go func() {
		mailer := la.mailer()
		if mailer == nil {
			c.Logger().Error(MSG_MAILER_NOT_SET)
			return
		}
		if err := mailer.Send(email); err != nil {
			c.Logger().Errorf("Error sending OTP email: %v", err)
		}
	}()
}

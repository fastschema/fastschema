package auth

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/entity"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/errors"
	"github.com/fastschema/fastschema/pkg/utils"
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
	jwtCustomClaimsFunc func() fs.JwtCustomClaimsFunc
}

func NewLocalAuthProvider(config fs.Map, redirectURL string) (fs.AuthProvider, error) {
	la := &LocalProvider{
		config:           config,
		activationMethod: fs.MapValue(config, "activation_method", "manual"),
		activationURL:    fs.MapValue(config, "activation_url", ""),
		recoveryURL:      fs.MapValue(config, "recovery_url", ""),
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
) {
	la.db = db
	la.appKey = appKey
	la.appName = appName
	la.mailer = mailer
	la.appBaseURL = appBaseURL
	la.jwtCustomClaimsFunc = jwtCustomClaimsFunc

	if la.activationURL == "" {
		la.activationURL = appBaseURL() + "/auth/local/activate"
	}

	if la.recoveryURL == "" {
		la.recoveryURL = appBaseURL() + "/auth/local/recover"
	}
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
	if err := ValidateRegisterData(c, c.Logger(), la.db(), payload); err != nil {
		return nil, err
	}

	userEntity := payload.Entity(la.activationMethod, la.Name())
	if err := db.WithTx(la.db(), c, func(tx db.Client) error {
		user, err := db.Builder[*fs.User](tx).Create(c, userEntity)
		if err != nil {
			c.Logger().Errorf(MSG_USER_SAVE_ERROR+": %w", err)
			return ERR_SAVE_USER
		}

		if _, err = db.Builder[*fs.User](tx).
			Where(db.EQ("id", user.ID)).
			Update(c, entity.New().Set("provider_id", strconv.FormatUint(user.ID, 10))); err != nil {
			c.Logger().Errorf(MSG_USER_UPDATE_PROVIDER_ID_ERROR, err)
			return ERR_SAVE_USER
		}

		user.ProviderID = strconv.FormatUint(user.ID, 10)
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
	userID, err := ValidateConfirmationToken(data.Token, la.appKey())
	if err != nil {
		err = fmt.Errorf(MSG_INVALID_TOKEN+": %w", err)
		c.Logger().Error(err)
		return nil, err
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

func (la *LocalProvider) SendActivationLink(c fs.Context, data *Confirmation) (*Activation, error) {
	if la.activationMethod != "email" {
		return nil, errors.BadRequest()
	}
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

func (la *LocalProvider) LocalLogin(c fs.Context, payload *LoginData) (_ *LoginResponse, err error) {
	if payload == nil || strings.TrimSpace(payload.Login) == "" || payload.Password == "" {
		return nil, errors.UnprocessableEntity(MSG_INVALID_LOGIN_OR_PASSWORD)
	}

	login := strings.TrimSpace(payload.Login)
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
			entity.FieldCreatedAt,
			entity.FieldUpdatedAt,
			entity.FieldDeletedAt,
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

	jwtToken, exp, err := user.JwtClaim(c, &fs.UserJwtConfig{
		Key:              la.appKey(),
		CustomClaimsFunc: nil,
	})
	if err != nil {
		return nil, err
	}

	return &LoginResponse{Token: jwtToken, Expires: exp}, nil
}

func (la *LocalProvider) Recover(c fs.Context, data *Recovery) (_ bool, err error) {
	if !utils.IsValidEmail(data.Email) {
		return false, errors.UnprocessableEntity(MSG_INVALID_EMAIL)
	}

	user, err := db.Builder[*fs.User](la.db()).
		Where(db.EQ("email", data.Email)).
		Where(db.EQ("provider", la.Name())).
		Select("id", "email").
		First(c)
	if err != nil && !db.IsNotFound(err) {
		c.Logger().Error(err)
		return false, errors.InternalServerError(MSG_CHECKING_USER_ERROR)
	}

	if user == nil {
		return true, nil
	}

	email, err := CreateRecoveryEmail(la, user)
	if err != nil {
		c.Logger().Errorf(MSG_CREATEP_RECOVERY_MAIL_ERROR+": %w", err)
		return false, errors.BadRequest(MSG_CREATEP_RECOVERY_MAIL_ERROR)
	}

	go SendConfirmationEmail(la, c.Logger(), email)

	return true, nil
}

func (la *LocalProvider) RecoverCheck(c fs.Context, data *Confirmation) (_ bool, err error) {
	userID, err := ValidateConfirmationToken(data.Token, la.appKey())
	return userID > 0, err
}

func (la *LocalProvider) ResetPassword(c fs.Context, data *ResetPassword) (_ bool, err error) {
	userID, err := ValidateConfirmationToken(data.Token, la.appKey())
	if err != nil {
		return false, err
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

	return true, nil
}

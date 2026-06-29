package auth

import (
	"encoding/json"
	"net/url"
	"time"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/entity"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/logger"
	"github.com/fastschema/fastschema/pkg/errors"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/google/uuid"
)

// emailChangeToken is the payload carried by the (encrypted) confirmation token.
// The new email travels inside the token (so no extra DB column is needed); the
// Session row tracks single-use state.
type emailChangeToken struct {
	SID   string `json:"sid"`   // pending session id
	Email string `json:"email"` // proposed new email
	Exp   int64  `json:"exp"`   // expiry (UnixMicro)
}

func encodeEmailChangeToken(sid, email string, exp time.Time, key string) (string, error) {
	raw, err := json.Marshal(emailChangeToken{SID: sid, Email: email, Exp: exp.UnixMicro()})
	if err != nil {
		return "", err
	}
	return utils.Encrypt(string(raw), key)
}

func decodeEmailChangeToken(token, key string) (*emailChangeToken, error) {
	decrypted, err := utils.Decrypt(token, key)
	if err != nil {
		return nil, ERR_INVALID_TOKEN
	}
	var data emailChangeToken
	if err := json.Unmarshal([]byte(decrypted), &data); err != nil {
		return nil, ERR_INVALID_TOKEN
	}
	return &data, nil
}

// ChangeEmail starts an authenticated email change: re-authenticate with the
// current password, validate the new address (format + policy + uniqueness),
// persist a pending email-change session, and email the old (notify) and new
// (confirm) addresses. The change is NOT applied until confirmed.
func (la *LocalProvider) ChangeEmail(c fs.Context, payload *ChangeEmailRequest) (*EmailChangeResponse, error) {
	authUser := c.User()
	if authUser == nil {
		return nil, errors.Unauthorized()
	}
	if payload == nil || payload.CurrentPassword == "" || !utils.IsValidEmail(payload.NewEmail) {
		return nil, errors.UnprocessableEntity(MSG_INVALID_EMAIL)
	}

	// keeppassword keeps the hashed password in the result (it is redacted by
	// default) so we can re-authenticate the caller.
	c.Local("keeppassword", "true")
	user, err := db.Builder[*fs.User](la.db()).
		Where(db.EQ("id", authUser.ID)).
		Select("id", "email", "password", "provider", "username").
		First(c)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, errors.Unauthorized()
		}
		c.Logger().Error(err)
		return nil, errors.InternalServerError(MSG_CHECKING_USER_ERROR)
	}

	// Only local accounts have a password to change their email with.
	if user.Provider != ProviderLocal {
		return nil, errors.BadRequest(MSG_EMAIL_CHANGE_ONLY_LOCAL)
	}

	// Re-authentication: verify current password before a sensitive change.
	if err := utils.CheckHash(payload.CurrentPassword, user.Password); err != nil {
		return nil, errors.UnprocessableEntity(MSG_INCORRECT_PASSWORD)
	}

	// Validate the new email through the built-in policy (normalize + domain
	// rules) - same gate as registration. Custom OnPreUserRegister hooks are NOT
	// fired here (those are signup side-effects, not email changes).
	newEmail := payload.NewEmail
	if la.registrationPolicy != nil {
		if p := la.registrationPolicy(); p != nil {
			in := &fs.RegistrationInput{Email: newEmail, Provider: ProviderLocal}
			if err := BuiltinPolicyValidator(p)(c, in); err != nil {
				return nil, err
			}
			newEmail = in.Email
		}
	}

	if newEmail == user.Email {
		return nil, errors.BadRequest(MSG_EMAIL_CHANGE_SAME)
	}

	if taken, err := la.emailTaken(c, newEmail, user.ID); err != nil {
		return nil, err
	} else if taken {
		return nil, errors.BadRequest(MSG_EMAIL_NOT_AVAILABLE)
	}

	// Invalidate any prior pending change for this user, then create a fresh one.
	la.invalidateEmailChangeSessions(c, user.ID)

	sessionID, err := uuid.NewV7()
	if err != nil {
		return nil, errors.InternalServerError(MSG_OTP_SESSION_CREATE_ERROR)
	}
	expiresAt := time.Now().Add(24 * time.Hour)
	if _, err := db.Builder[*fs.Session](la.db()).Create(c, entity.New().
		Set("id", sessionID).
		Set("user_id", user.ID).
		Set("type", string(fs.SessionTypeEmailChange)).
		Set("status", string(fs.SessionStatusActive)).
		Set("ip_address", c.IP()).
		Set("expires_at", expiresAt).
		Set("last_activity_at", time.Now()),
	); err != nil {
		c.Logger().Error(err)
		return nil, errors.InternalServerError(MSG_OTP_SESSION_CREATE_ERROR)
	}

	token, err := encodeEmailChangeToken(sessionID.String(), newEmail, expiresAt, la.appKey())
	if err != nil {
		c.Logger().Error(err)
		return nil, errors.InternalServerError(MSG_OTP_SESSION_CREATE_ERROR)
	}

	go la.sendEmailChangeMails(c.Logger(), user.Email, newEmail, token)

	return &EmailChangeResponse{Message: MSG_EMAIL_CHANGE_REQUESTED}, nil
}

// ConfirmEmailChange completes a pending change using the single-use token sent
// to the new address. The user UPDATE fires the standard audit hook.
func (la *LocalProvider) ConfirmEmailChange(c fs.Context, payload *ConfirmEmailChange) (*EmailChangeResponse, error) {
	if payload == nil || payload.Token == "" {
		return nil, ERR_INVALID_TOKEN
	}

	data, err := decodeEmailChangeToken(payload.Token, la.appKey())
	if err != nil {
		return nil, ERR_INVALID_TOKEN
	}
	if time.Now().UnixMicro() > data.Exp {
		return nil, ERR_TOKEN_EXPIRED
	}
	sessionUUID, err := uuid.Parse(data.SID)
	if err != nil {
		return nil, ERR_INVALID_TOKEN
	}

	session, err := db.Builder[*fs.Session](la.db()).
		Where(db.EQ("id", sessionUUID)).
		Where(db.EQ("type", string(fs.SessionTypeEmailChange))).
		Where(db.EQ("status", string(fs.SessionStatusActive))).
		First(c)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, ERR_INVALID_TOKEN
		}
		c.Logger().Error(err)
		return nil, errors.InternalServerError(MSG_CHECKING_USER_ERROR)
	}
	if session.ExpiresAt != nil && session.ExpiresAt.Before(time.Now()) {
		la.markSessionInactive(c, session.ID)
		return nil, ERR_TOKEN_EXPIRED
	}

	if taken, err := la.emailTaken(c, data.Email, session.UserID); err != nil {
		return nil, err
	} else if taken {
		la.markSessionInactive(c, session.ID)
		return nil, errors.BadRequest(MSG_EMAIL_NOT_AVAILABLE)
	}

	// Commit the new email and consume the session atomically. The session flip
	// is a conditional update (status=active -> inactive) checked via the
	// affected-row count from the model-layer mutation: if two requests race the
	// same token, only one matches active and the loser aborts. (db.Builder.Update
	// can't be used here — it re-queries with the same predicate after updating,
	// so a status filter would always report 0 rows.)
	err = db.WithTx(la.db(), c, func(tx db.Client) error {
		sessionModel, e := tx.Model("session")
		if e != nil {
			return e
		}
		consumed, e := sessionModel.Mutation().
			Where(
				db.EQ("id", session.ID),
				db.EQ("status", string(fs.SessionStatusActive)),
			).
			Update(c, entity.New().Set("status", string(fs.SessionStatusInactive)))
		if e != nil {
			return e
		}
		if consumed == 0 {
			return ERR_INVALID_TOKEN // already consumed by a concurrent request
		}
		if _, e := db.Builder[*fs.User](tx).
			Where(db.EQ("id", session.UserID)).
			Update(c, entity.New().Set("email", data.Email)); e != nil {
			return e
		}
		return nil
	})
	if err != nil {
		if errors.Is(err, ERR_INVALID_TOKEN) {
			return nil, ERR_INVALID_TOKEN
		}
		c.Logger().Error(err)
		return nil, ERR_SAVE_USER
	}

	return &EmailChangeResponse{Message: MSG_EMAIL_CHANGED}, nil
}

// emailTaken reports whether another user already owns the given email.
func (la *LocalProvider) emailTaken(c fs.Context, email string, excludeUserID uuid.UUID) (bool, error) {
	existing, err := db.Builder[*fs.User](la.db()).
		Where(db.EQ("email", email)).
		Where(db.NEQ("id", excludeUserID)).
		Select("id").
		First(c)
	if err != nil && !db.IsNotFound(err) {
		c.Logger().Error(err)
		return false, errors.InternalServerError(MSG_CHECKING_USER_ERROR)
	}
	return existing != nil, nil
}

// invalidateEmailChangeSessions marks all active pending email-change sessions
// for a user as inactive (best-effort).
func (la *LocalProvider) invalidateEmailChangeSessions(c fs.Context, userID uuid.UUID) {
	_, _ = db.Builder[*fs.Session](la.db()).
		Where(db.EQ("user_id", userID)).
		Where(db.EQ("type", string(fs.SessionTypeEmailChange))).
		Where(db.EQ("status", string(fs.SessionStatusActive))).
		Update(c, entity.New().Set("status", string(fs.SessionStatusInactive)))
}

// sendEmailChangeMails notifies the current address and sends a confirmation
// link to the new address. Tokens/URLs are never logged. Takes a logger (not the
// request Context) since it runs in a goroutine after the handler returns.
func (la *LocalProvider) sendEmailChangeMails(log logger.Logger, oldEmail, newEmail, token string) {
	mailer := la.mailer()
	if mailer == nil {
		log.Error(MSG_MAILER_NOT_SET)
		return
	}

	appName := la.appName()
	if appName == "" {
		appName = "FastSchema"
	}

	changeURL := fs.MapValue(la.config, "email_change_url", la.appBaseURL()+"/auth/local/email/confirm")
	confirmLink := changeURL + "?token=" + url.QueryEscape(token)

	notify := &fs.Mail{
		To:      []string{oldEmail},
		Subject: appName + ": email change requested",
		Body: "A request was made to change the email address on your " + appName +
			" account to " + newEmail + ". If you did not request this, please secure your account immediately.",
	}
	confirm := &fs.Mail{
		To:      []string{newEmail},
		Subject: appName + ": confirm your new email",
		Body: "Confirm your new email address for " + appName +
			" by opening this link: " + confirmLink,
	}

	if err := mailer.Send(notify); err != nil {
		log.Errorf("error sending email-change notification: %v", err)
	}
	if err := mailer.Send(confirm); err != nil {
		log.Errorf("error sending email-change confirmation: %v", err)
	}
}

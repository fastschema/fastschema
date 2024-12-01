package auth

import (
	"context"
	"net/url"
	"time"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/logger"
	"github.com/fastschema/fastschema/pkg/errors"
	"github.com/fastschema/fastschema/pkg/utils"
)

func SendConfirmationEmail(la *LocalProvider, logger logger.Logger, mail *fs.Mail) {
	if la.activationMethod != "email" {
		return
	}

	if mailer := la.mailer(); mailer == nil {
		logger.Error(MSG_MAILER_NOT_SET)
	} else if err := mailer.Send(mail); err != nil {
		logger.Error(MSG_SEND_ACTIVATION_EMAIL_ERROR, err)
	}
}

func CreateConfirmationURL(baseURL, appKey string, user *fs.User) (string, error) {
	if token, err := utils.CreateConfirmationToken(user.ID, appKey); err != nil {
		return "", err
	} else {
		url, err := url.Parse(baseURL)
		if err != nil {
			return "", err
		}

		q := url.Query()
		q.Set("token", token)
		url.RawQuery = q.Encode()

		return url.String(), nil
	}
}

func ValidateConfirmationToken(token, key string) (uint64, error) {
	if token == "" {
		return 0, ERR_INVALID_TOKEN
	}

	parsed, err := utils.ParseConfirmationToken(token, key)
	if err != nil {
		return 0, err
	}

	if time.Now().UnixMicro() > parsed.Exp {
		return parsed.UserID, ERR_TOKEN_EXPIRED
	}

	return parsed.UserID, nil
}

func ValidateRegisterData(
	c context.Context,
	logger logger.Logger,
	dbClient db.Client,
	payload *Register,
) (err error) {
	if !utils.IsValidEmail(payload.Email) ||
		payload.Username == "" ||
		payload.Password == "" ||
		payload.ConfirmPassword == "" {
		return errors.UnprocessableEntity(MSG_INVALID_REGISTRATION)
	}

	if payload.Password != payload.ConfirmPassword {
		return errors.BadRequest(MSG_INVALID_PASSWORD)
	}

	existedUser, err := db.Builder[*fs.User](dbClient).
		Where(db.Or(
			db.EQ("username", payload.Username),
			db.EQ("email", payload.Email),
		)).
		Select("id").
		First(c)

	if err != nil && !db.IsNotFound(err) {
		logger.Error(MSG_CHECKING_USER_ERROR, err)
		return errors.BadRequest(MSG_CHECKING_USER_ERROR)
	}

	if existedUser != nil {
		return errors.BadRequest(MSG_USER_EXISTS)
	}

	return nil
}

package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/logger"
	"github.com/fastschema/fastschema/pkg/errors"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/google/uuid"
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

func ValidateConfirmationToken(token, key string) (uuid.UUID, error) {
	var emptyUUID uuid.UUID
	if token == "" {
		return emptyUUID, ERR_INVALID_TOKEN
	}

	parsed, err := utils.ParseConfirmationToken(token, key)
	if err != nil {
		// A malformed/undecryptable token is a client error, not a server error.
		// ParseConfirmationToken returns a plain error which would otherwise map to 500.
		return emptyUUID, ERR_INVALID_TOKEN
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
		payload.Password == "" ||
		payload.ConfirmPassword == "" {
		return errors.UnprocessableEntity(MSG_INVALID_REGISTRATION)
	}

	if payload.Password != payload.ConfirmPassword {
		return errors.BadRequest(MSG_INVALID_PASSWORD)
	}

	existingUser, err := db.Builder[*fs.User](dbClient).
		Where(db.EQ("email", payload.Email)).
		Select("id").
		First(c)

	if err != nil && !db.IsNotFound(err) {
		logger.Error(MSG_CHECKING_USER_ERROR, err)
		return errors.BadRequest(MSG_CHECKING_USER_ERROR)
	}

	if existingUser != nil {
		msg := MSG_USER_EXISTS
		if existingUser.Provider != ProviderLocal {
			msg = MSG_EXISTING_USER_WITH_EMAIL
		}
		return errors.BadRequest(msg)
	}

	return nil
}

func SendRequest[T any](method, url string, headers map[string]string, requestBody io.Reader) (T, error) {
	var t T
	req, err := http.NewRequest(method, url, requestBody)
	if err != nil {
		return t, err
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return t, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return t, fmt.Errorf("request failed with status code %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return t, err
	}

	if err := json.Unmarshal(body, &t); err != nil {
		return t, err
	}

	return t, nil
}

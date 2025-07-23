package auth

import (
	"fmt"
	"strings"

	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/errors"
	"github.com/fastschema/fastschema/pkg/utils"
)

var (
	MSG_USER_SAVE_ERROR               = "error saving user"
	MSG_USER_ACTIVATION_ERROR         = "error activating user"
	MSG_INVALID_TOKEN                 = "invalid token"
	MSG_TOKEN_EXPIRED                 = "token expired"
	MSG_CREATE_ACTIVATION_MAIL_ERROR  = "error while creating activation email: %w"
	MSG_USER_UPDATE_PROVIDER_ID_ERROR = "error while update provider id: %w"
	MSG_CREATEP_RECOVERY_MAIL_ERROR   = "error while creating recovery email"
	MSG_INVALID_EMAIL                 = "invalid email"
	MSG_INVALID_PASSWORD              = "invalid password"
	MSG_INVALID_LOGIN_OR_PASSWORD     = "invalid login or password"
	MSG_USER_IS_INACTIVE              = "user is inactive"
	MSG_INVALID_REGISTRATION          = "username, email, password and confirm_password are required"
	MSG_SEND_ACTIVATION_EMAIL_ERROR   = "error while sending activation email"
	MSG_MAILER_NOT_SET                = "mailer is not set"
	MSG_CHECKING_USER_ERROR           = "error checking user"
	MSG_USER_EXISTS                   = "user already exists"

	ERR_SAVE_USER     = errors.InternalServerError(MSG_USER_SAVE_ERROR)
	ERR_INVALID_TOKEN = errors.BadRequest(MSG_INVALID_TOKEN)
	ERR_TOKEN_EXPIRED = errors.BadRequest(MSG_TOKEN_EXPIRED)
	ERR_INVALID_LOGIN = errors.UnprocessableEntity(MSG_INVALID_LOGIN_OR_PASSWORD)
)

func CreateActivationEmail(la *LocalProvider, user *fs.User) (*fs.Mail, error) {
	activationURL, err := CreateConfirmationURL(la.activationURL, la.appKey(), user)
	if err != nil {
		return nil, err
	}

	bodyLines := []string{
		fmt.Sprintf(`Hey %s,`, user.Username),
		fmt.Sprintf(`Welcome to %s! We’re excited to have you on board. To complete your account setup, please click the link below to verify your email address:`, la.appName()),
		fmt.Sprintf(`<a href="%s">%s</a>`, activationURL, "Verify Email"),
		`In case the link doesn’t work, please copy and paste the following URL in your browser:`,
		activationURL,
		"Welcome aboard!",
		"Sincerely,",
		la.appName(),
	}

	return &fs.Mail{
		To:      []string{user.Email},
		Subject: "Welcome to " + la.appName(),
		Body: strings.Join(utils.Map(bodyLines, func(l string) string {
			return fmt.Sprintf("<p>%s</p>", l)
		}), "\r\n"),
	}, nil
}

func CreateRecoveryEmail(la *LocalProvider, user *fs.User) (*fs.Mail, error) {
	recoveryURL, err := CreateConfirmationURL(la.recoveryURL, la.appKey(), user)
	if err != nil {
		return nil, err
	}

	bodyLines := []string{
		fmt.Sprintf(`Hey %s,`, user.Username),
		fmt.Sprintf(`You recently requested to reset your password for your %s account. Click the link below to reset it.`, la.appName()),
		fmt.Sprintf(`<a href="%s">%s</a>`, recoveryURL, "Reset Password"),
		`In case the link doesn’t work, please copy and paste the following URL in your browser:`,
		recoveryURL,
		"If you did not request a password reset, please ignore this email.",
		"Thanks,",
		la.appName(),
	}

	return &fs.Mail{
		To:      []string{user.Email},
		Subject: fmt.Sprintf("Reset your %s password", la.appName()),
		Body: strings.Join(utils.Map(bodyLines, func(l string) string {
			return fmt.Sprintf("<p>%s</p>", l)
		}), "\r\n"),
	}, nil
}

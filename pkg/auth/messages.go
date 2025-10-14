package auth

import (
	"fmt"
	"strings"

	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/errors"
	"github.com/fastschema/fastschema/pkg/utils"
)

var (
	MSG_USER_SAVE_ERROR               = "Error saving user"
	MSG_USER_ACTIVATION_ERROR         = "Error activating user"
	MSG_USER_ALREADY_ACTIVE           = "Your account is already activated"
	MSG_INVALID_TOKEN                 = "Invalid token"
	MSG_TOKEN_EXPIRED                 = "Token expired"
	MSG_CREATE_ACTIVATION_MAIL_ERROR  = "Error while creating activation email: %w"
	MSG_USER_UPDATE_PROVIDER_ID_ERROR = "Error while update provider id: %w"
	MSG_CREATEP_RECOVERY_MAIL_ERROR   = "Error while creating recovery email"
	MSG_INVALID_EMAIL                 = "Invalid email"
	MSG_INVALID_PASSWORD              = "Invalid password"
	MSG_INVALID_LOGIN_OR_PASSWORD     = "Invalid login or password" //nolint:gosec // G101: This is an error message, not a hardcoded credential
	MSG_USER_IS_INACTIVE              = "User is inactive"
	MSG_INVALID_REGISTRATION          = "Email, password and confirm_password are required"
	MSG_SEND_ACTIVATION_EMAIL_ERROR   = "Error while sending activation email"
	MSG_MAILER_NOT_SET                = "Mailer is not set"
	MSG_CHECKING_USER_ERROR           = "Error checking user"
	MSG_USER_EXISTS                   = "User already exists"
	MSG_EXISTING_USER_WITH_EMAIL      = "Looks like you already have an account with this email. Please log in using your existing sign-in method, or try signing up with a different email."

	ERR_SAVE_USER           = errors.InternalServerError(MSG_USER_SAVE_ERROR)
	ERR_INVALID_TOKEN       = errors.BadRequest(MSG_INVALID_TOKEN)
	ERR_TOKEN_EXPIRED       = errors.BadRequest(MSG_TOKEN_EXPIRED)
	ERR_INVALID_LOGIN       = errors.UnprocessableEntity(MSG_INVALID_LOGIN_OR_PASSWORD)
	ERR_USER_ALREADY_ACTIVE = errors.BadRequest(MSG_USER_ALREADY_ACTIVE)
)

func CreateActivationEmail(la *LocalProvider, user *fs.User) (*fs.Mail, error) {
	activationURL, err := CreateConfirmationURL(la.activationURL, la.appKey(), user)
	if err != nil {
		return nil, err
	}

	name := user.FirstName
	if name == "" {
		name = user.Username
	}

	if name == "" {
		parts := strings.Split(user.Email, "@")
		if len(parts) > 0 {
			name = parts[0]
		} else {
			name = "there"
		}
	}

	bodyLines := []string{
		fmt.Sprintf(`Hey %s,`, name),
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

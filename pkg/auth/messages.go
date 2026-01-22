package auth

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/errors"
)

// Default email templates
const (
	DefaultActivationSubject = "Welcome to {{.AppName}}"
	DefaultActivationBody    = `<p>Hey {{.UserName}},</p>
<p>Welcome to {{.AppName}}! We're excited to have you on board. To complete your account setup, please click the link below to verify your email address:</p>
<p><a href="{{.ActionURL}}">{{.ActionLabel}}</a></p>
<p>In case the link doesn't work, please copy and paste the following URL in your browser:</p>
<p>{{.ActionURL}}</p>
<p>Welcome aboard!</p>
<p>Sincerely,</p>
<p>{{.AppName}}</p>`

	DefaultRecoverySubject = "Reset your {{.AppName}} password"
	DefaultRecoveryBody    = `<p>Hey {{.UserName}},</p>
<p>You recently requested to reset your password for your {{.AppName}} account. Click the link below to reset it.</p>
<p><a href="{{.ActionURL}}">{{.ActionLabel}}</a></p>
<p>In case the link doesn't work, please copy and paste the following URL in your browser:</p>
<p>{{.ActionURL}}</p>
<p>If you did not request a password reset, please ignore this email.</p>
<p>Thanks,</p>
<p>{{.AppName}}</p>`
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

// resolveTemplate parses and executes a template string.
// Uses defaultValue if configValue is empty.
func resolveTemplate(configValue, defaultValue string, data *fs.EmailTemplateData) (string, error) {
	tmplStr := configValue
	if tmplStr == "" {
		tmplStr = defaultValue
	}

	tmpl, err := template.New("email").Parse(tmplStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse email template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute email template: %w", err)
	}

	return buf.String(), nil
}

// deriveUserName extracts display name from user
func deriveUserName(user *fs.User) string {
	if user.FirstName != "" {
		return user.FirstName
	}
	if user.Username != "" {
		return user.Username
	}
	parts := strings.Split(user.Email, "@")
	if len(parts) > 0 {
		return parts[0]
	}
	return "there"
}

func CreateActivationEmail(la *LocalProvider, user *fs.User) (*fs.Mail, error) {
	activationURL, err := CreateConfirmationURL(la.activationURL, la.appKey(), user)
	if err != nil {
		return nil, err
	}

	data := &fs.EmailTemplateData{
		AppName:     la.appName(),
		UserName:    deriveUserName(user),
		UserEmail:   user.Email,
		ActionURL:   activationURL,
		ActionLabel: "Verify Email",
	}

	var templates *fs.EmailTemplates
	if la.emailTemplates != nil {
		templates = la.emailTemplates()
	}
	var subjectTmpl, bodyTmpl string
	if templates != nil {
		subjectTmpl = templates.ActivationSubject
		bodyTmpl = templates.ActivationBody
	}

	subject, err := resolveTemplate(subjectTmpl, DefaultActivationSubject, data)
	if err != nil {
		return nil, err
	}

	body, err := resolveTemplate(bodyTmpl, DefaultActivationBody, data)
	if err != nil {
		return nil, err
	}

	return &fs.Mail{
		To:      []string{user.Email},
		Subject: subject,
		Body:    body,
	}, nil
}

func CreateRecoveryEmail(la *LocalProvider, user *fs.User) (*fs.Mail, error) {
	recoveryURL, err := CreateConfirmationURL(la.recoveryURL, la.appKey(), user)
	if err != nil {
		return nil, err
	}

	data := &fs.EmailTemplateData{
		AppName:     la.appName(),
		UserName:    deriveUserName(user),
		UserEmail:   user.Email,
		ActionURL:   recoveryURL,
		ActionLabel: "Reset Password",
	}

	var templates *fs.EmailTemplates
	if la.emailTemplates != nil {
		templates = la.emailTemplates()
	}
	var subjectTmpl, bodyTmpl string
	if templates != nil {
		subjectTmpl = templates.RecoverySubject
		bodyTmpl = templates.RecoveryBody
	}

	subject, err := resolveTemplate(subjectTmpl, DefaultRecoverySubject, data)
	if err != nil {
		return nil, err
	}

	body, err := resolveTemplate(bodyTmpl, DefaultRecoveryBody, data)
	if err != nil {
		return nil, err
	}

	return &fs.Mail{
		To:      []string{user.Email},
		Subject: subject,
		Body:    body,
	}, nil
}

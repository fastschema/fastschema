package auth

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/errors"
	"github.com/fastschema/fastschema/pkg/utils"
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

	// OTP-related messages
	MSG_OTP_NOT_ENABLED           = "OTP passwordless login is not enabled"
	MSG_OTP_SENT                  = "If an account exists with this email, a verification code has been sent"
	MSG_OTP_INVALID               = "Invalid or expired verification code"
	MSG_OTP_EXPIRED               = "Verification code has expired"
	MSG_OTP_MAX_ATTEMPTS          = "Maximum verification attempts exceeded. Please request a new code"
	MSG_OTP_GENERATION_ERROR      = "Error generating verification code"
	MSG_OTP_SEND_ERROR            = "Error sending verification code"
	MSG_OTP_SESSION_CREATE_ERROR  = "Error creating OTP session"
	MSG_OTP_USER_NOT_FOUND        = "No account found with this email"
	MSG_OTP_EMAIL_REQUIRED        = "Email is required"
	MSG_OTP_CODE_REQUIRED         = "Verification code is required"
	MSG_SESSION_ID_REQUIRED       = "Session ID is required"
	MSG_OTP_SESSION_INVALID       = "Invalid or expired session"
	MSG_OTP_VERIFICATION_REQUIRED = "OTP verification required before this action"

	ERR_SAVE_USER           = errors.InternalServerError(MSG_USER_SAVE_ERROR)
	ERR_INVALID_TOKEN       = errors.BadRequest(MSG_INVALID_TOKEN)
	ERR_TOKEN_EXPIRED       = errors.BadRequest(MSG_TOKEN_EXPIRED)
	ERR_INVALID_LOGIN       = errors.UnprocessableEntity(MSG_INVALID_LOGIN_OR_PASSWORD)
	ERR_USER_ALREADY_ACTIVE = errors.BadRequest(MSG_USER_ALREADY_ACTIVE)

	// OTP-related errors
	ERR_OTP_NOT_ENABLED  = errors.BadRequest(MSG_OTP_NOT_ENABLED)
	ERR_OTP_INVALID      = errors.UnprocessableEntity(MSG_OTP_INVALID)
	ERR_OTP_EXPIRED      = errors.UnprocessableEntity(MSG_OTP_EXPIRED)
	ERR_OTP_MAX_ATTEMPTS = errors.TooManyRequests(MSG_OTP_MAX_ATTEMPTS)
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

// CreateActivationOTPEmail creates an email with OTP for account activation
func CreateActivationOTPEmail(appName, email, otp string, expirationMinutes int) *fs.Mail {
	name := "there"
	parts := strings.Split(email, "@")
	if len(parts) > 0 && parts[0] != "" {
		name = parts[0]
	}

	bodyLines := []string{
		fmt.Sprintf(`Hey %s,`, name),
		fmt.Sprintf(`Welcome to %s! To complete your account setup, please use the verification code below:`, appName),
		fmt.Sprintf(`<div style="font-size: 32px; font-weight: bold; letter-spacing: 8px; text-align: center; padding: 20px; background-color: #f5f5f5; border-radius: 8px; margin: 20px 0;">%s</div>`, otp),
		fmt.Sprintf(`This code will expire in %d minutes.`, expirationMinutes),
		`If you didn't create an account, you can safely ignore this email.`,
		`For security, never share this code with anyone.`,
		"Welcome aboard!",
		appName,
	}

	return &fs.Mail{
		To:      []string{email},
		Subject: fmt.Sprintf("Your %s activation code: %s", appName, otp),
		Body: strings.Join(utils.Map(bodyLines, func(l string) string {
			return fmt.Sprintf("<p>%s</p>", l)
		}), "\r\n"),
	}
}

// CreateRecoveryOTPEmail creates an email with OTP for password recovery
func CreateRecoveryOTPEmail(appName, email, otp string, expirationMinutes int) *fs.Mail {
	name := "there"
	parts := strings.Split(email, "@")
	if len(parts) > 0 && parts[0] != "" {
		name = parts[0]
	}

	bodyLines := []string{
		fmt.Sprintf(`Hey %s,`, name),
		fmt.Sprintf(`You requested to reset your password for your %s account. Use the verification code below:`, appName),
		fmt.Sprintf(`<div style="font-size: 32px; font-weight: bold; letter-spacing: 8px; text-align: center; padding: 20px; background-color: #f5f5f5; border-radius: 8px; margin: 20px 0;">%s</div>`, otp),
		fmt.Sprintf(`This code will expire in %d minutes.`, expirationMinutes),
		`If you didn't request a password reset, please ignore this email.`,
		`For security, never share this code with anyone.`,
		"Thanks,",
		appName,
	}

	return &fs.Mail{
		To:      []string{email},
		Subject: fmt.Sprintf("Your %s password reset code: %s", appName, otp),
		Body: strings.Join(utils.Map(bodyLines, func(l string) string {
			return fmt.Sprintf("<p>%s</p>", l)
		}), "\r\n"),
	}
}

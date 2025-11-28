package mailer

import (
	"fmt"
	"net/mail"
	"net/smtp"
	"strings"
	"time"

	"github.com/fastschema/fastschema/fs"
)

type SMTPConfig struct {
	DefaultSenderName string
	DefaultSenderMail string
	Host              string
	Port              int
	Username          string
	Password          string
	Insecure          bool
}

type SMTPMailer struct {
	name   string
	config *SMTPConfig
}

func NewSMTPMailer(name string, config *SMTPConfig) (fs.Mailer, error) {
	if config.Host == "" {
		return nil, fmt.Errorf("mailer %s: host is required", name)
	}

	return &SMTPMailer{
		name:   name,
		config: config,
	}, nil
}

func (s *SMTPMailer) Driver() string {
	return "smtp"
}

func (s *SMTPMailer) Name() string {
	return s.name
}

func (s *SMTPMailer) Send(content *fs.Mail, froms ...mail.Address) error {
	if len(content.To) == 0 {
		return fmt.Errorf("mailer %s: no recipient", s.name)
	}

	from := append(froms, mail.Address{
		Name:    s.config.DefaultSenderName,
		Address: s.config.DefaultSenderMail,
	})[0]

	header := make(map[string]string)
	header["From"] = from.String()
	header["To"] = strings.Join(content.To, ", ")

	if len(content.CC) > 0 {
		header["Cc"] = strings.Join(content.CC, ", ")
	}

	if len(content.BCC) > 0 {
		header["Bcc"] = strings.Join(content.BCC, ", ")
	}

	header["Subject"] = content.Subject
	header["MIME-Version"] = "1.0"
	header["Content-Type"] = `text/html; charset="UTF-8"`
	header["Date"] = time.Now().Format(time.RFC1123Z)

	// Build the message
	var messageBuilder strings.Builder
	for k, v := range header {
		fmt.Fprintf(&messageBuilder, "%s: %s\r\n", k, v)
	}
	messageBuilder.WriteString("\r\n<html><body>")
	messageBuilder.WriteString(content.Body)
	messageBuilder.WriteString("</body></html>")
	message := messageBuilder.String()

	// Combine all recipients (To, Cc, Bcc)
	recipients := content.To
	recipients = append(recipients, content.CC...)
	recipients = append(recipients, content.BCC...)

	// Authentication
	var auth smtp.Auth
	if s.config.Username != "" {
		auth = smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.Host)
	}

	// Send the email
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	return smtp.SendMail(addr, auth, from.Address, recipients, []byte(message))
}

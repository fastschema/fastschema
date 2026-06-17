package fs

import (
	"maps"
	"net/mail"
)

// EmailTemplateData contains all data available to email templates
type EmailTemplateData struct {
	AppName     string // Application name
	UserName    string // Recipient's display name
	UserEmail   string // Recipient's email address
	ActionURL   string // The target link (activation or recovery URL)
	ActionLabel string // Descriptive text for the link
}

// EmailTemplates holds string templates for email customization
// All templates use Go text/template syntax
type EmailTemplates struct {
	ActivationSubject string `json:"activation_subject"`
	ActivationBody    string `json:"activation_body"`
	RecoverySubject   string `json:"recovery_subject"`
	RecoveryBody      string `json:"recovery_body"`
}

// Clone creates a deep copy of EmailTemplates
func (e *EmailTemplates) Clone() *EmailTemplates {
	if e == nil {
		return nil
	}
	return &EmailTemplates{
		ActivationSubject: e.ActivationSubject,
		ActivationBody:    e.ActivationBody,
		RecoverySubject:   e.RecoverySubject,
		RecoveryBody:      e.RecoveryBody,
	}
}

type MailConfig struct {
	SenderName        string          `json:"sender_name"`
	SenderMail        string          `json:"sender_mail"`
	DefaultClientName string          `json:"default_client"`
	Clients           []Map           `json:"clients"`
	Templates         *EmailTemplates `json:"templates"`
}

func (m *MailConfig) Clone() *MailConfig {
	c := &MailConfig{
		SenderName:        m.SenderName,
		SenderMail:        m.SenderMail,
		DefaultClientName: m.DefaultClientName,
		Clients:           make([]Map, len(m.Clients)),
		Templates:         m.Templates.Clone(),
	}

	for i, client := range m.Clients {
		newClient := Map{}
		maps.Copy(newClient, client)
		c.Clients[i] = newClient
	}

	return c
}

type Mail struct {
	Subject string
	Body    string
	To      []string
	CC      []string
	BCC     []string
}

type Mailer interface {
	Send(mail *Mail, froms ...mail.Address) error
	Name() string
	Driver() string
}

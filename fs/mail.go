package fs

import "net/mail"

type MailConfig struct {
	SenderName        string `json:"sender_name"`
	SenderMail        string `json:"sender_mail"`
	DefaultClientName string `json:"default_client"`
	Clients           []Map  `json:"clients"`
}

func (m *MailConfig) Clone() *MailConfig {
	c := &MailConfig{
		SenderName:        m.SenderName,
		SenderMail:        m.SenderMail,
		DefaultClientName: m.DefaultClientName,
		Clients:           make([]Map, len(m.Clients)),
	}

	for i, client := range m.Clients {
		newClient := Map{}
		for k, v := range client {
			newClient[k] = v
		}
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

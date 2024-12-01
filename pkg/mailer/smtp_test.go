package mailer_test

import (
	"net/mail"
	"testing"

	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/mailer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSMTPMailer(t *testing.T) {
	tests := []struct {
		name    string
		config  *mailer.SMTPConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: &mailer.SMTPConfig{
				Host:     "smtp.example.com",
				Port:     587,
				Username: "user",
				Password: "pass",
			},
			wantErr: false,
		},
		{
			name: "missing host",
			config: &mailer.SMTPConfig{
				Host:     "",
				Port:     587,
				Username: "user",
				Password: "pass",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mailer, err := mailer.NewSMTPMailer("testMailer", tt.config)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, mailer)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, mailer)
			}
		})
	}
}

func TestSmtpMethods(t *testing.T) {
	// Invalid mock server address
	mock, err := mailer.CreateMockSMTPServer("invalid")
	assert.Error(t, err)
	assert.Nil(t, mock)

	mock, err = mailer.CreateMockSMTPServer()
	require.NoError(t, err)
	defer func() {
		assert.NoError(t, mock.Close())
	}()
	config := &mailer.SMTPConfig{
		Host:     mock.Host,
		Port:     mock.Port,
		Username: "user",
		Password: "pass",
	}

	smtpMailer, err := mailer.NewSMTPMailer("testMailer", config)
	assert.NoError(t, err)

	assert.Equal(t, "smtp", smtpMailer.Driver())
	assert.Equal(t, "testMailer", smtpMailer.Name())

	sampleMail := &fs.Mail{
		Subject: "Test",
		Body:    "Test body",
		CC:      []string{"ccuser@site.local"},
		BCC:     []string{"bccuser@site.local"},
	}

	// No recipient
	err = smtpMailer.Send(sampleMail)
	assert.Error(t, err)

	// With recipient
	sampleMail.To = []string{"touser@site.local"}
	from := mail.Address{Address: "admin@somesite.com"}
	err = smtpMailer.Send(sampleMail, from)
	assert.NoError(t, err)
	assert.Contains(t, string(mock.Backend.Messages[0]), "Test body")
}

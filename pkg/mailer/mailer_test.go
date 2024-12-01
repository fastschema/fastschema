package mailer_test

import (
	"testing"

	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/mailer"
	"github.com/stretchr/testify/assert"
)

func TestNewMailersFromConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  *fs.MailConfig
		wantErr bool
	}{
		{
			name: "valid smtp config",
			config: &fs.MailConfig{
				SenderName: "Test Sender",
				SenderMail: "sender@example.com",
				Clients: []map[string]interface{}{
					{
						"name":     "smtp1",
						"driver":   "smtp",
						"host":     "smtp.example.com",
						"port":     587,
						"username": "user",
						"password": "pass",
						"insecure": false,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "missing name",
			config: &fs.MailConfig{
				SenderName: "Test Sender",
				SenderMail: "sender@example.com",
				Clients: []map[string]interface{}{
					{
						"driver":   "smtp",
						"host":     "smtp.example.com",
						"port":     587,
						"username": "user",
						"password": "pass",
						"insecure": false,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "missing driver",
			config: &fs.MailConfig{
				SenderName: "Test Sender",
				SenderMail: "sender@example.com",
				Clients: []map[string]interface{}{
					{
						"name":     "smtp1",
						"host":     "smtp.example.com",
						"port":     587,
						"username": "user",
						"password": "pass",
						"insecure": false,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "unknown driver",
			config: &fs.MailConfig{
				SenderName: "Test Sender",
				SenderMail: "sender@example.com",
				Clients: []map[string]interface{}{
					{
						"name":     "unknown1",
						"driver":   "unknown",
						"host":     "smtp.example.com",
						"port":     587,
						"username": "user",
						"password": "pass",
						"insecure": false,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "missing smtp config fields",
			config: &fs.MailConfig{
				SenderName: "Test Sender",
				SenderMail: "sender@example.com",
				Clients: []map[string]interface{}{
					{
						"name":   "smtp1",
						"driver": "smtp",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "missing smtp host",
			config: &fs.MailConfig{
				SenderName: "Test Sender",
				SenderMail: "sender@example.com",
				Clients: []map[string]interface{}{
					{
						"name":     "smtp1",
						"driver":   "smtp",
						"port":     587,
						"username": "user",
						"password": "pass",
						"insecure": false,
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mailers, err := mailer.NewMailersFromConfig(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, mailers)
			}
		})
	}
}

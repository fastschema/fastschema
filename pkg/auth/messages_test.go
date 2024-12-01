package auth_test

import (
	"testing"

	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/auth"
	"github.com/stretchr/testify/assert"
)

func TestCreateActivationEmail(t *testing.T) {
	user := &fs.User{ID: 123}
	mailer := &MockMailer{}

	// Create error (invalid key)
	{
		provider := createLocalAuthProvider(&testAppConfig{
			activation: "manual",
			mailer:     mailer,
			key:        "invalid",
		})
		mail, err := auth.CreateActivationEmail(provider, user)
		assert.Error(t, err)
		assert.Nil(t, mail)
	}

	// Success
	{
		provider := createLocalAuthProvider(&testAppConfig{
			activation: "manual",
			mailer:     mailer,
		})
		mail, err := auth.CreateActivationEmail(provider, user)
		assert.NoError(t, err)
		assert.Equal(t, "Welcome to testApp", mail.Subject)
		assert.Contains(t, mail.Body, "?token=")
	}
}

func TestCreateRecoveryEmail(t *testing.T) {
	user := &fs.User{ID: 123}
	mailer := &MockMailer{}

	// Create error (invalid key)
	{
		provider := createLocalAuthProvider(&testAppConfig{
			activation: "manual",
			mailer:     mailer,
			key:        "invalid",
		})
		mail, err := auth.CreateRecoveryEmail(provider, user)
		assert.Error(t, err)
		assert.Nil(t, mail)
	}

	// Success
	{
		provider := createLocalAuthProvider(&testAppConfig{
			activation: "manual",
			mailer:     mailer,
		})
		mail, err := auth.CreateRecoveryEmail(provider, user)
		assert.NoError(t, err)
		assert.Equal(t, "Reset your testApp password", mail.Subject)
		assert.Contains(t, mail.Body, "?token=")
	}
}

package auth_test

import (
	"context"
	"net/mail"
	"testing"
	"time"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/entity"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/logger"
	"github.com/fastschema/fastschema/pkg/auth"
	"github.com/fastschema/fastschema/pkg/entdbadapter"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/stretchr/testify/assert"
)

type MockMailer struct {
	sent int
	err  error
}

const testKey = "rLnWcTEFhTNEeEenhnfZEJahGaTrLnWa"

func (m *MockMailer) Send(mail *fs.Mail, froms ...mail.Address) error {
	if m.err != nil {
		return m.err
	}
	m.sent++
	return nil
}

func (m *MockMailer) Name() string {
	return "mock"
}

func (m *MockMailer) Driver() string {
	return "mock"
}

func TestSendConfirmationEmail(t *testing.T) {
	logger := logger.CreateMockLogger(true)

	// There is no mailer
	{
		provider := createLocalAuthProvider(&testAppConfig{activation: "email"})
		auth.SendConfirmationEmail(provider, logger, &fs.Mail{
			To:      []string{"test@site.local"},
			Subject: "Test",
			Body:    "Test",
		})

		assert.Contains(t, logger.Last().String(), auth.MSG_MAILER_NOT_SET)
	}

	// Activation method not email
	{
		mailer := &MockMailer{}
		provider := createLocalAuthProvider(&testAppConfig{
			activation: "manual",
			mailer:     mailer,
		})
		auth.SendConfirmationEmail(provider, logger, &fs.Mail{
			To:      []string{"test@site.local"},
			Subject: "Test",
			Body:    "Test",
		})
		assert.Equal(t, 0, mailer.sent)
	}

	// Send mail error
	{
		mailer := &MockMailer{err: assert.AnError}
		provider := createLocalAuthProvider(&testAppConfig{
			activation: "email",
			mailer:     mailer,
		})
		auth.SendConfirmationEmail(provider, logger, &fs.Mail{
			To:      []string{"test@site.local"},
			Subject: "Test",
			Body:    "Test",
		})

		assert.Contains(t, logger.Last().String(), auth.MSG_SEND_ACTIVATION_EMAIL_ERROR)
	}

	// Success
	{
		mailer := &MockMailer{}
		provider := createLocalAuthProvider(&testAppConfig{
			activation: "email",
			mailer:     mailer,
		})
		auth.SendConfirmationEmail(provider, logger, &fs.Mail{
			To:      []string{"test@site.local"},
			Subject: "Test",
			Body:    "Test",
		})

		assert.Equal(t, 1, mailer.sent)
	}
}
func TestCreateConfirmationUrl(t *testing.T) {
	user := &fs.User{ID: 123}

	// Invalid key size
	{
		url, err := auth.CreateConfirmationURL("http://localhost:8080/confirm", "", user)
		assert.Error(t, err)
		assert.Empty(t, url)
	}

	// Invalid base URL
	{
		url, err := auth.CreateConfirmationURL(":", testKey, user)
		assert.Error(t, err)
		assert.Empty(t, url)
	}

	// Success
	{
		url, err := auth.CreateConfirmationURL("http://localhost:8080/confirm", testKey, user)
		assert.NoError(t, err)
		assert.Contains(t, url, "http://localhost:8080/confirm?token=")
	}
}
func TestValidateConfirmationToken(t *testing.T) {
	// Empty token
	{
		userID, err := auth.ValidateConfirmationToken("", testKey)
		assert.Error(t, err)
		assert.Equal(t, auth.ERR_INVALID_TOKEN, err)
		assert.Equal(t, uint64(0), userID)
	}

	// Invalid token
	{
		userID, err := auth.ValidateConfirmationToken("invalidToken", testKey)
		assert.Error(t, err)
		assert.Equal(t, uint64(0), userID)
	}

	// Expired token
	{
		expiresAt := time.Now().Add(-time.Hour)
		expiredToken, _ := utils.CreateConfirmationToken(123, testKey, expiresAt)
		userID, err := auth.ValidateConfirmationToken(expiredToken, testKey)
		assert.Error(t, err)
		assert.Equal(t, auth.ERR_TOKEN_EXPIRED, err)
		assert.Equal(t, uint64(123), userID)
	}

	// Valid token
	{
		validToken, _ := utils.CreateConfirmationToken(123, testKey)
		userID, err := auth.ValidateConfirmationToken(validToken, testKey)
		assert.NoError(t, err)
		assert.Equal(t, uint64(123), userID)
	}
}

type MockDBQuery struct {
	*entdbadapter.Query
	err      error
	entities []*entity.Entity
}

func (q *MockDBQuery) Limit(limit uint) db.Querier {
	return q
}

func (q *MockDBQuery) Offset(offset uint) db.Querier {
	return q
}

func (q *MockDBQuery) Order(order ...string) db.Querier {
	return q
}

func (q *MockDBQuery) Select(columns ...string) db.Querier {
	return q
}

func (q *MockDBQuery) Get(ctx context.Context) ([]*entity.Entity, error) {
	if q.err != nil {
		return nil, q.err
	}

	return q.entities, nil
}

type MockDBModel struct {
	*entdbadapter.Model
	entities []*entity.Entity
}

func (m *MockDBModel) Query(predicates ...*db.Predicate) db.Querier {
	return &MockDBQuery{
		entities: m.entities,
	}
}

type MockDBClient struct {
	*db.NoopClient
	model db.Model
}

func (d *MockDBClient) Model(model any) (db.Model, error) {
	if d.model == nil {
		return nil, assert.AnError
	}

	return d.model, nil
}

func TestValidateRegisterData(t *testing.T) {
	logger := logger.CreateMockLogger(true)

	type args struct {
		payload *auth.Register
	}
	tests := []struct {
		name    string
		db      db.Client
		args    args
		wantErr string
	}{
		{
			name: "missing fields",
			db:   &db.NoopClient{},
			args: args{
				payload: &auth.Register{
					Username:        "",
					Email:           "",
					Password:        "",
					ConfirmPassword: "",
				},
			},
			wantErr: auth.MSG_INVALID_REGISTRATION,
		},
		{
			name: "passwords do not match",
			db:   &db.NoopClient{},
			args: args{
				payload: &auth.Register{
					Username:        "newUser",
					Email:           "new@site.local",
					Password:        "password",
					ConfirmPassword: "differentPassword",
				},
			},
			wantErr: auth.MSG_INVALID_PASSWORD,
		},
		{
			name: "check user error",
			db:   &MockDBClient{},
			args: args{
				payload: &auth.Register{
					Username:        "newUser",
					Email:           "newUser@site.local",
					Password:        "password",
					ConfirmPassword: "password",
				},
			},
			wantErr: "error checking user",
		},
		{
			name: "user exists",
			db: &MockDBClient{
				model: &MockDBModel{
					entities: []*entity.Entity{entity.New(5)},
				},
			},
			args: args{
				payload: &auth.Register{
					Username:        "existingUser",
					Email:           "existing@site.local",
					Password:        "password",
					ConfirmPassword: "password",
				},
			},
			wantErr: auth.MSG_EXISTING_USER_WITH_EMAIL,
		},
		{
			name: "successful validation",
			db: &MockDBClient{
				model: &MockDBModel{
					entities: []*entity.Entity{},
				},
			},
			args: args{
				payload: &auth.Register{
					Username:        "newUser",
					Email:           "new@site.local",
					Password:        "password",
					ConfirmPassword: "password",
				},
			},
			wantErr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := auth.ValidateRegisterData(context.Background(), logger, tt.db, tt.args.payload)
			if tt.wantErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

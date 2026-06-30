package auth_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/logger"
	"github.com/fastschema/fastschema/pkg/auth"
	"github.com/fastschema/fastschema/pkg/entdbadapter"
	"github.com/fastschema/fastschema/pkg/restfulresolver"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
	"github.com/google/uuid"
)

type restfulResolverContext = restfulresolver.Context
type mockContext struct {
	*restfulResolverContext
	redirectURL string
	args        map[string]string
}

func (m *mockContext) Logger() logger.Logger {
	return logger.CreateMockLogger(true)
}

func (m *mockContext) Redirect(url string) error {
	m.redirectURL = url
	return nil
}

func (m *mockContext) Cookie(string, ...*fs.Cookie) string { return "" }

func (m *mockContext) FormValue(key string, defaultValue ...string) string {
	return ""
}

func (m *mockContext) Arg(name string, defaultValues ...string) string {
	if value, ok := m.args[name]; ok {
		return value
	}

	if len(defaultValues) > 0 {
		return defaultValues[0]
	}

	return ""
}

func createGithubAuth(configs ...fs.Map) fs.AuthProvider {
	config := append(configs, fs.Map{})[0]
	config["client_id"] = "mockClientID"
	config["client_secret"] = "mockClient"
	return utils.Must(auth.NewGithubAuthProvider(config, "http://localhost:8080/callback"))
}

func createGoogleAuth(configs ...fs.Map) fs.AuthProvider {
	config := append(configs, fs.Map{})[0]
	config["client_id"] = "mockClientID"
	config["client_secret"] = "mockClient"
	return utils.Must(auth.NewGoogleAuthProvider(config, "http://localhost:8080/callback"))
}

type RW = http.ResponseWriter

func createAuthProviderTestSever(handler func(w RW)) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w RW, r *http.Request) {
		handler(w)
	}))
}

type testAppConfig struct {
	key        string
	mailer     fs.Mailer
	db         db.Client
	createData bool
	activation string
	// Optional registration hook + policy wiring for the local provider.
	preUserRegister    fs.PreUserRegisterHook
	registrationPolicy *fs.RegistrationPolicy
	// Track created user IDs for testing
	user01ID uuid.UUID
	user02ID uuid.UUID
}

func createLocalAuthProvider(config *testAppConfig) *auth.LocalProvider {
	redirectURL := "http://localhost:8080/api/auth/local/callback"

	if config.key == "" {
		config.key = testKey
	}

	if config.createData {
		schemasDir := utils.Must(os.MkdirTemp("", "schemas"))
		migrationsDir := utils.Must(os.MkdirTemp("", "migrations"))
		sb := utils.Must(schema.NewBuilderFromDir(schemasDir, fs.SystemSchemaTypes...))
		config.db = utils.Must(entdbadapter.NewTestClient(migrationsDir, sb))
		roleModel := utils.Must(config.db.Model("role"))
		userModel := utils.Must(config.db.Model("user"))
		adminRoleIDRaw := utils.Must(roleModel.CreateFromJSON(context.Background(), `{"name": "admin"}`))
		adminRoleID := adminRoleIDRaw.(uuid.UUID)
		userRoleIDRaw := utils.Must(roleModel.CreateFromJSON(context.Background(), `{"name": "User"}`))
		userRoleID := userRoleIDRaw.(uuid.UUID)
		// Update the global fs.RoleUser.ID so that Register.Entity() can reference it
		fs.RoleUser.ID = userRoleID
		user01IDRaw := utils.Must(userModel.CreateFromJSON(context.Background(), `{
			"username": "user01",
			"password": "user01",
			"email": "user01@site.local",
			"provider": "local",
			"active": false,
			"roles": [{"id": "`+adminRoleID.String()+`"}]
	}`))
		config.user01ID = user01IDRaw.(uuid.UUID)
		user02IDRaw := utils.Must(userModel.CreateFromJSON(context.Background(), `{
			"username": "user02",
			"password": "user02",
			"email": "user02@site.local",
			"provider": "local",
			"active": true,
			"roles": [{"id": "`+adminRoleID.String()+`"}]
	}`))
		config.user02ID = user02IDRaw.(uuid.UUID)
	}

	authProvider := utils.Must(auth.NewLocalAuthProvider(fs.Map{
		"activation_method": config.activation,
	}, redirectURL))
	localAuthProvider := authProvider.(*auth.LocalProvider)
	localAuthProvider.Init(
		func() db.Client { return config.db },
		func() string { return config.key },
		func() string { return "testApp" },
		func() string { return "http://localhost:8080" },
		func(names ...string) fs.Mailer { return config.mailer },
		nil,
		nil,
		nil,
		func(ctx context.Context, in *fs.RegistrationInput) error {
			if config.preUserRegister != nil {
				return config.preUserRegister(ctx, in)
			}
			return nil
		},
		func() *fs.RegistrationPolicy { return config.registrationPolicy },
	)

	return localAuthProvider
}

package auth_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/auth"
	"github.com/fastschema/fastschema/pkg/entdbadapter"
	"github.com/fastschema/fastschema/pkg/restfulresolver"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
)

type restfulResolverContext = restfulresolver.Context
type mockContext struct {
	*restfulResolverContext
	rediectURL string
	args       map[string]string
}

func (m *mockContext) Redirect(url string) error {
	m.rediectURL = url
	return nil
}

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
		utils.Must(roleModel.CreateFromJSON(context.Background(), `{"name": "admin"}`))
		utils.Must(roleModel.CreateFromJSON(context.Background(), `{"name": "user"}`))
		utils.Must(userModel.CreateFromJSON(context.Background(), `{
			"username": "user01",
			"password": "user01",
			"email": "user01@site.local",
			"provider": "local",
			"active": false,
			"roles": [{"id": 1}]
	}`))
		utils.Must(userModel.CreateFromJSON(context.Background(), `{
			"username": "user02",
			"password": "user02",
			"email": "user02@site.local",
			"provider": "local",
			"active": true,
			"roles": [{"id": 1}]
	}`))
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
	)

	return localAuthProvider
}

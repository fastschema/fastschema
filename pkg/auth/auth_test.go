package auth_test

import (
	"net/http"
	"net/http/httptest"

	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/auth"
	"github.com/fastschema/fastschema/pkg/restfulresolver"
	"github.com/fastschema/fastschema/pkg/utils"
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

func (m *mockContext) Arg(name string, defaultValues ...string) string {
	if value, ok := m.args[name]; ok {
		return value
	}

	if len(defaultValues) > 0 {
		return defaultValues[0]
	}

	return ""
}

func createGithubAuth(configs ...map[string]string) fs.AuthProvider {
	config := append(configs, map[string]string{})[0]
	config["client_id"] = "mockClientID"
	config["client_secret"] = "mockClient"
	return utils.Must(auth.NewGithubAuthProvider(config, "http://localhost:8080/callback"))
}

func createGoogleAuth(configs ...map[string]string) fs.AuthProvider {
	config := append(configs, map[string]string{})[0]
	config["client_id"] = "mockClientID"
	config["client_secret"] = "mockClient"
	return utils.Must(auth.NewGoogleAuthProvider(config, "http://localhost:8080/callback"))
}

type RW = http.ResponseWriter

func createTestSever(handler func(w RW)) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w RW, r *http.Request) {
		handler(w)
	}))
}

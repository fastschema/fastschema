package auth_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/logger"
	"github.com/fastschema/fastschema/pkg/auth"
	"github.com/fastschema/fastschema/pkg/entdbadapter"
	"github.com/fastschema/fastschema/pkg/restfulresolver"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
	"github.com/stretchr/testify/assert"
)

type testApp struct {
	sb *schema.Builder
	db db.Client
}

func (s testApp) DB() db.Client {
	return s.db
}

func (s testApp) Key() string {
	return "test"
}

func TestNewLocalAuthProvider(t *testing.T) {
	redirectURL := "http://localhost:8080/api/auth/local/callback"
	authProvider, err := auth.NewLocalAuthProvider(map[string]string{}, redirectURL)
	assert.NoError(t, err)
	assert.NotNil(t, authProvider)
	assert.Equal(t, "local", authProvider.Name())

	user, err := authProvider.Callback(nil)
	assert.Nil(t, user)
	assert.Nil(t, err)

	sb := utils.Must(schema.NewBuilderFromDir(t.TempDir(), fs.SystemSchemaTypes...))
	dbc := utils.Must(entdbadapter.NewTestClient(utils.Must(os.MkdirTemp("", "migrations")), sb))
	testApp := &testApp{sb: sb, db: dbc}

	authProvider.(*auth.LocalAuthProvider).Init(testApp.DB, testApp.Key)
	roleModel := utils.Must(dbc.Model("role"))
	userModel := utils.Must(dbc.Model("user"))
	utils.Must(roleModel.CreateFromJSON(context.Background(), `{"name": "user"}`))
	utils.Must(userModel.CreateFromJSON(context.Background(), fmt.Sprintf(
		`{
			"username": "user01",
			"password": "%s",
			"active": false,
			"roles": [{"id": 1}]
		}`,
		utils.Must(utils.GenerateHash("user01")),
	)))
	utils.Must(userModel.CreateFromJSON(context.Background(), fmt.Sprintf(
		`{
			"username": "user02",
			"password": "%s",
			"active": true,
			"roles": [{"id": 1}]
		}`,
		utils.Must(utils.GenerateHash("user02")),
	)))

	resources := fs.NewResourcesManager()
	resources.Add(fs.Post("/user/login", func(ctx fs.Context, _ any) (any, error) {
		return authProvider.Login(ctx)
	}, &fs.Meta{Public: true}))
	assert.NoError(t, resources.Init())
	server := restfulresolver.NewRestfulResolver(&restfulresolver.ResolverConfig{
		ResourceManager: resources,
		Logger:          logger.CreateMockLogger(true),
	}).Server()

	// Case 1: No login data
	req := httptest.NewRequest("POST", "/user/login", nil)
	resp, _ := server.Test(req)
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 400, resp.StatusCode)
	assert.Contains(t, utils.Must(utils.ReadCloserToString(resp.Body)), `login and password are required`)

	// Case 2: Empty login data
	// Case 1: Login User not found
	req = httptest.NewRequest("POST", "/user/login", bytes.NewReader([]byte(`{}`)))
	resp, _ = server.Test(req)
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 400, resp.StatusCode)
	assert.Contains(t, utils.Must(utils.ReadCloserToString(resp.Body)), `login and password are required`)

	// Case 1: Login User not found
	req = httptest.NewRequest("POST", "/user/login", bytes.NewReader([]byte(`{"login": "user", "password": "user"}`)))
	resp, _ = server.Test(req)
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 401, resp.StatusCode)
	assert.Contains(t, utils.Must(utils.ReadCloserToString(resp.Body)), `invalid login or password`)

	// Case 2: Login Invalid password
	req = httptest.NewRequest("POST", "/user/login", bytes.NewReader([]byte(`{"login": "user01", "password": "123"}`)))
	resp, _ = server.Test(req)
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 401, resp.StatusCode)
	assert.Contains(t, utils.Must(utils.ReadCloserToString(resp.Body)), `invalid login or password`)

	// Case 3: Login User is not active
	req = httptest.NewRequest("POST", "/user/login", bytes.NewReader([]byte(`{"login": "user01", "password": "user01"}`)))
	resp, _ = server.Test(req)
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 401, resp.StatusCode)
	assert.Contains(t, utils.Must(utils.ReadCloserToString(resp.Body)), `user is not active`)

	// Case 4: Login Success
	req = httptest.NewRequest("POST", "/user/login", bytes.NewReader([]byte(`{"login": "user02", "password": "user02"}`)))
	resp, _ = server.Test(req)
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 200, resp.StatusCode)
	type LoginResponse struct {
		Data *auth.LoginResponse `json:"data"`
	}
	loginResponse := LoginResponse{}
	assert.NoError(t, json.Unmarshal([]byte(utils.Must(utils.ReadCloserToString(resp.Body))), &loginResponse))
	assert.NotEmpty(t, loginResponse.Data.Token)
	assert.NotEmpty(t, loginResponse.Data.Expires)
}

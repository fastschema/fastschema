package userservice_test

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
	"github.com/fastschema/fastschema/pkg/entdbadapter"
	"github.com/fastschema/fastschema/pkg/restresolver"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
	userservice "github.com/fastschema/fastschema/services/user"
	jwt "github.com/golang-jwt/jwt/v4"
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

func TestUserService(t *testing.T) {
	sb := utils.Must(schema.NewBuilderFromDir(t.TempDir(), fs.SystemSchemaTypes...))
	db := utils.Must(entdbadapter.NewTestClient(utils.Must(os.MkdirTemp("", "migrations")), sb))
	testApp := &testApp{sb: sb, db: db}
	userService := userservice.New(testApp)

	roleModel := utils.Must(db.Model("role"))
	userModel := utils.Must(db.Model("user"))
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
	resources.Middlewares = []fs.Middleware{func(c fs.Context) error {
		authToken := c.AuthToken()
		jwtToken, err := jwt.ParseWithClaims(
			authToken,
			&fs.UserJwtClaims{},
			func(token *jwt.Token) (any, error) {
				return []byte("test"), nil
			},
		)

		if err == nil {
			if claims, ok := jwtToken.Claims.(*fs.UserJwtClaims); ok && jwtToken.Valid {
				user := claims.User
				user.Roles = []*fs.Role{{ID: 1, Name: "user"}}
				c.Value("user", user)
			}
		}

		return c.Next()
	}}
	resources.Group("user").
		Add(fs.NewResource("logout", userService.Logout, &fs.Meta{
			Post:   "/logout",
			Public: true,
		})).
		Add(fs.NewResource("me", userService.Me, &fs.Meta{Public: true})).
		Add(fs.NewResource("login", userService.Login, &fs.Meta{
			Post:   "/login",
			Public: true,
		}))

	assert.NoError(t, resources.Init())
	server := restresolver.NewRestResolver(resources, logger.CreateMockLogger(true)).Server()

	// Case 1: Login User not found
	req := httptest.NewRequest("POST", "/user/login", bytes.NewReader([]byte(`{"login": "user", "password": "user"}`)))
	resp := utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 401, resp.StatusCode)
	response := utils.Must(utils.ReadCloserToString(resp.Body))

	assert.Contains(t, response, `invalid login or password`)

	// Case 2: Login Invalid password
	req = httptest.NewRequest("POST", "/user/login", bytes.NewReader([]byte(`{"login": "user01", "password": "123"}`)))
	resp = utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 401, resp.StatusCode)
	response = utils.Must(utils.ReadCloserToString(resp.Body))
	assert.Contains(t, response, `invalid login or password`)

	// Case 3: Login User is not active
	req = httptest.NewRequest("POST", "/user/login", bytes.NewReader([]byte(`{"login": "user01", "password": "user01"}`)))
	resp = utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 401, resp.StatusCode)
	response = utils.Must(utils.ReadCloserToString(resp.Body))
	assert.Contains(t, response, `user is not active`)

	// Case 4: Login Success
	req = httptest.NewRequest("POST", "/user/login", bytes.NewReader([]byte(`{"login": "user02", "password": "user02"}`)))
	resp = utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 200, resp.StatusCode)
	response = utils.Must(utils.ReadCloserToString(resp.Body))
	type LoginResponse struct {
		Data *userservice.LoginResponse `json:"data"`
	}
	loginResponse := LoginResponse{}
	assert.NoError(t, json.Unmarshal([]byte(response), &loginResponse))
	assert.NotEmpty(t, loginResponse.Data.Token)
	assert.NotEmpty(t, loginResponse.Data.Expires)

	// Case 5: Logout
	req = httptest.NewRequest("POST", "/user/logout", nil)
	resp = utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 200, resp.StatusCode)

	// Case 6: Me Unauthorized
	req = httptest.NewRequest("GET", "/user/me", nil)
	resp = utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 401, resp.StatusCode)

	// Case 7: Me Success
	req = httptest.NewRequest("GET", "/user/me", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", loginResponse.Data.Token))
	req.Header.Set("Authorization", "Bearer "+loginResponse.Data.Token)
	resp = utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 200, resp.StatusCode)
	response = utils.Must(utils.ReadCloserToString(resp.Body))
	assert.Contains(t, response, `"username":"user02"`)
}

package auth_test

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/logger"
	"github.com/fastschema/fastschema/pkg/auth"
	"github.com/fastschema/fastschema/pkg/restfulresolver"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestNewLocalAuthProviderMethods(t *testing.T) {
	localAuthProvider := createLocalAuthProvider(&testAppConfig{
		activation: "manual",
	})

	assert.Equal(t, "local", localAuthProvider.Name())

	user := utils.Must(localAuthProvider.Callback(nil))
	assert.Nil(t, user)

	res := utils.Must(localAuthProvider.Login(nil))
	assert.Nil(t, res)
}

func createServer(t *testing.T, resource *fs.Resource) *restfulresolver.Server {
	resources := fs.NewResourcesManager()
	resources.Add(resource)
	assert.NoError(t, resources.Init())
	server := restfulresolver.NewRestfulResolver(&restfulresolver.ResolverConfig{
		ResourceManager: resources,
		Logger:          logger.CreateMockLogger(true),
	}).Server()

	return server
}

func TestLocalAuthLogin(t *testing.T) {
	config := &testAppConfig{activation: "manual", createData: true}
	provider := createLocalAuthProvider(config)
	server := createServer(t, fs.Post("/user/login", provider.LocalLogin, &fs.Meta{Public: true}))

	// Case 1: No login data
	req := httptest.NewRequest("POST", "/user/login", nil)
	resp, _ := server.Test(req)
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 400, resp.StatusCode)

	// Case 2: Empty login data
	// Case 1: Login User not found
	req = httptest.NewRequest("POST", "/user/login", bytes.NewReader([]byte(`{}`)))
	resp, _ = server.Test(req)
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 422, resp.StatusCode)
	assert.Contains(t, utils.Must(utils.ReadCloserToString(resp.Body)), `invalid login or password`)

	// Case 1: Login User not found
	req = httptest.NewRequest("POST", "/user/login", bytes.NewReader([]byte(`{"login": "user", "password": "user"}`)))
	resp, _ = server.Test(req)
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 422, resp.StatusCode)
	assert.Contains(t, utils.Must(utils.ReadCloserToString(resp.Body)), `invalid login or password`)

	// Case 2: Login User is not active
	req = httptest.NewRequest("POST", "/user/login", bytes.NewReader([]byte(`{"login": "user01", "password": "user01"}`)))
	resp, _ = server.Test(req)
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 401, resp.StatusCode)
	assert.Contains(t, utils.Must(utils.ReadCloserToString(resp.Body)), `user is inactive`)

	// Case 3: Login Invalid password
	req = httptest.NewRequest("POST", "/user/login", bytes.NewReader([]byte(`{"login": "user02", "password": "123"}`)))
	resp, _ = server.Test(req)
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 422, resp.StatusCode)
	assert.Contains(t, utils.Must(utils.ReadCloserToString(resp.Body)), `invalid login or password`)

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

	// Case 5: Error checking user
	{
		assert.NoError(t, config.db.Close())
		req = httptest.NewRequest(
			"POST", "/user/login",
			bytes.NewReader([]byte(`{"login": "user02", "password": "user02"}`)),
		)
		resp, _ = server.Test(req)
		defer func() { assert.NoError(t, resp.Body.Close()) }()
		assert.Equal(t, 500, resp.StatusCode)
	}
}

func TestLocalAuthRegister(t *testing.T) {
	config := &testAppConfig{activation: "manual", createData: true}
	provider := createLocalAuthProvider(config)
	server := createServer(t, fs.Post("/user/register", provider.Register, &fs.Meta{Public: true}))

	// Case 1: The request body is empty
	{
		req := httptest.NewRequest("POST", "/user/register", nil)
		resp, _ := server.Test(req)
		defer func() { assert.NoError(t, resp.Body.Close()) }()
		assert.Equal(t, 400, resp.StatusCode)
	}

	// Case 2: The request body is invalid
	{
		req := httptest.NewRequest("POST", "/user/register", bytes.NewReader([]byte(`{}`)))
		resp, _ := server.Test(req)
		defer func() { assert.NoError(t, resp.Body.Close()) }()
		assert.Equal(t, 422, resp.StatusCode)
		assert.Contains(t, utils.Must(utils.ReadCloserToString(resp.Body)), auth.MSG_INVALID_REGISTRATION)
	}

	validData := []byte(`{
		"username": "user03",
		"email": "user03@local.ltd",
		"provider": "local",
		"password": "user03",
		"confirm_password": "user03"
	}`)

	// Case 3: Register failed because of invalid app key
	{
		config.key = "invalid"
		req := httptest.NewRequest("POST", "/user/register", bytes.NewReader(validData))
		resp, _ := server.Test(req)
		defer func() { assert.NoError(t, resp.Body.Close()) }()
		assert.Equal(t, 500, resp.StatusCode)
		config.key = testKey
	}

	// Case 4: The request body is valid
	{
		req := httptest.NewRequest("POST", "/user/register", bytes.NewReader(validData))
		resp, _ := server.Test(req)
		defer func() { assert.NoError(t, resp.Body.Close()) }()
		assert.Equal(t, 200, resp.StatusCode)
		assert.Contains(t, utils.Must(utils.ReadCloserToString(resp.Body)), `{"activation":"manual"}`)
	}
}

func TestLocalAuthActivation(t *testing.T) {
	config := &testAppConfig{activation: "manual", createData: true}
	provider := createLocalAuthProvider(config)
	server := createServer(t, fs.Post("/user/activate", provider.Activate, &fs.Meta{Public: true}))

	// Case 1: Invalid token
	{
		req := httptest.NewRequest("POST", "/user/activate", bytes.NewReader([]byte("{}")))
		resp, _ := server.Test(req)
		defer func() { assert.NoError(t, resp.Body.Close()) }()
		assert.Equal(t, 400, resp.StatusCode)
		assert.Contains(t, utils.Must(utils.ReadCloserToString(resp.Body)), auth.MSG_INVALID_TOKEN)
	}

	// Case 2: Success
	token := utils.Must(utils.CreateConfirmationToken(1, config.key))
	{
		req := httptest.NewRequest("POST", "/user/activate?token="+token, bytes.NewReader([]byte(`{"token": "`+token+`"}`)))
		resp, _ := server.Test(req)
		defer func() { assert.NoError(t, resp.Body.Close()) }()
		assert.Equal(t, 200, resp.StatusCode)
		assert.Contains(t, utils.Must(utils.ReadCloserToString(resp.Body)), `{"activation":"activated"}`)
	}

	// Case 3: User already activated
	{
		req := httptest.NewRequest("POST", "/user/activate", bytes.NewReader([]byte(`{"token": "`+token+`"}`)))
		resp, _ := server.Test(req)
		defer func() { assert.NoError(t, resp.Body.Close()) }()
		assert.Equal(t, 200, resp.StatusCode)
	}

	// Case 4: Update status failed
	{
		assert.NoError(t, config.db.Close())
		req := httptest.NewRequest("POST", "/user/activate", bytes.NewReader([]byte(`{"token": "`+token+`"}`)))
		resp, _ := server.Test(req)
		defer func() { assert.NoError(t, resp.Body.Close()) }()
		assert.Equal(t, 400, resp.StatusCode)
		assert.Contains(t, utils.Must(utils.ReadCloserToString(resp.Body)), auth.MSG_USER_ACTIVATION_ERROR)
	}
}

func TestLocalAuthSendActivationLink(t *testing.T) {
	// Case 1: Invalid activation method
	{
		provider := createLocalAuthProvider(&testAppConfig{activation: "manual", createData: true})
		server := createServer(t, fs.Post(
			"/user/activate/send",
			provider.SendActivationLink,
			&fs.Meta{Public: true},
		))
		req := httptest.NewRequest(
			"POST", "/user/activate/send",
			bytes.NewReader([]byte(`{"token": "123"}`)),
		)
		resp, _ := server.Test(req)
		defer func() { assert.NoError(t, resp.Body.Close()) }()
		assert.Equal(t, 400, resp.StatusCode)
	}

	config := &testAppConfig{activation: "email", createData: true}
	provider := createLocalAuthProvider(config)
	server := createServer(t, fs.Post(
		"/user/activate/send",
		provider.SendActivationLink,
		&fs.Meta{Public: true},
	))

	// Case 2: Invalid token
	{
		req := httptest.NewRequest(
			"POST", "/user/activate/send",
			bytes.NewReader([]byte(`{"token": "123"}`)),
		)
		resp, _ := server.Test(req)
		defer func() { assert.NoError(t, resp.Body.Close()) }()
		assert.Equal(t, 400, resp.StatusCode)
		assert.Contains(t, utils.Must(utils.ReadCloserToString(resp.Body)), auth.MSG_INVALID_TOKEN)
	}

	// Case 3: Token user is not found
	{
		expiredTime := time.Now().Add(-time.Hour * 48)
		token := utils.Must(utils.CreateConfirmationToken(123, config.key, expiredTime))
		req := httptest.NewRequest(
			"POST", "/user/activate/send",
			bytes.NewReader([]byte(`{"token": "`+token+`"}`)),
		)
		resp, _ := server.Test(req)
		defer func() { assert.NoError(t, resp.Body.Close()) }()
		assert.Equal(t, 400, resp.StatusCode)
		assert.Contains(t, utils.Must(utils.ReadCloserToString(resp.Body)), auth.MSG_INVALID_TOKEN)
	}

	// Case 4: Success
	{
		expiredTime := time.Now().Add(-time.Hour * 48)
		token := utils.Must(utils.CreateConfirmationToken(1, config.key, expiredTime))
		req := httptest.NewRequest(
			"POST", "/user/activate/send",
			bytes.NewReader([]byte(`{"token": "`+token+`"}`)),
		)
		resp, _ := server.Test(req)
		defer func() { assert.NoError(t, resp.Body.Close()) }()
		assert.Equal(t, 200, resp.StatusCode)
		assert.Contains(t, utils.Must(utils.ReadCloserToString(resp.Body)), `{"activation":"email"}`)
	}
}

func TestLocalAuthRecover(t *testing.T) {
	mailer := &MockMailer{}
	config := &testAppConfig{activation: "manual", createData: true, mailer: mailer}
	provider := createLocalAuthProvider(config)
	server := createServer(t, fs.Post(
		"/user/recover",
		provider.Recover,
		&fs.Meta{Public: true},
	))

	// Case 1: Invalid email
	{
		req := httptest.NewRequest(
			"POST", "/user/recover",
			bytes.NewReader([]byte(`{"email": ""}`)),
		)
		resp, _ := server.Test(req)
		defer func() { assert.NoError(t, resp.Body.Close()) }()
		assert.Equal(t, 422, resp.StatusCode)
	}

	// Case 2: Email not found
	{
		req := httptest.NewRequest(
			"POST", "/user/recover",
			bytes.NewReader([]byte(`{"email": "notfound@site.local"}`)),
		)
		resp, _ := server.Test(req)
		defer func() { assert.NoError(t, resp.Body.Close()) }()
		assert.Equal(t, 200, resp.StatusCode)
	}

	// Case 3: Create confirmation email failed
	{
		config.key = "invalid"
		req := httptest.NewRequest(
			"POST", "/user/recover",
			bytes.NewReader([]byte(`{"email": "user01@site.local"}`)),
		)
		resp, _ := server.Test(req)
		defer func() { assert.NoError(t, resp.Body.Close()) }()
		assert.Equal(t, 400, resp.StatusCode)
		assert.Contains(t, utils.Must(utils.ReadCloserToString(resp.Body)), auth.MSG_CREATEP_RECOVERY_MAIL_ERROR)
		config.key = testKey
	}

	// Case 4: Success
	{
		req := httptest.NewRequest(
			"POST", "/user/recover",
			bytes.NewReader([]byte(`{"email": "user01@site.local"}`)),
		)
		resp, _ := server.Test(req)
		defer func() { assert.NoError(t, resp.Body.Close()) }()
		assert.Equal(t, 200, resp.StatusCode)
		assert.Contains(t, utils.Must(utils.ReadCloserToString(resp.Body)), `{"data":true}`)
	}

	// Case 4: Check user failed
	{
		assert.NoError(t, config.db.Close())
		req := httptest.NewRequest(
			"POST", "/user/recover",
			bytes.NewReader([]byte(`{"email": "user01@site.local"}`)),
		)
		resp, _ := server.Test(req)
		defer func() { assert.NoError(t, resp.Body.Close()) }()
		assert.Equal(t, 500, resp.StatusCode)
		assert.Contains(t, utils.Must(utils.ReadCloserToString(resp.Body)), auth.MSG_CHECKING_USER_ERROR)
	}
}

func TestLocalAuthRecoverCheck(t *testing.T) {
	mailer := &MockMailer{}
	config := &testAppConfig{activation: "manual", createData: true, mailer: mailer}
	provider := createLocalAuthProvider(config)
	server := createServer(t, fs.Post(
		"/user/recover/check",
		provider.RecoverCheck,
		&fs.Meta{Public: true},
	))

	// Case 1: Invalid token
	{
		req := httptest.NewRequest(
			"POST", "/user/recover/check",
			bytes.NewReader([]byte(`{"token": "123"}`)),
		)
		resp, _ := server.Test(req)
		defer func() { assert.NoError(t, resp.Body.Close()) }()
		assert.Equal(t, 400, resp.StatusCode)
		assert.Contains(t, utils.Must(utils.ReadCloserToString(resp.Body)), auth.MSG_INVALID_TOKEN)
	}

	// Case 2: Success
	{
		token := utils.Must(utils.CreateConfirmationToken(1, config.key))
		req := httptest.NewRequest(
			"POST", "/user/recover/check",
			bytes.NewReader([]byte(`{"token": "`+token+`"}`)),
		)
		resp, _ := server.Test(req)
		defer func() { assert.NoError(t, resp.Body.Close()) }()
		assert.Equal(t, 200, resp.StatusCode)
	}
}

func TestLocalAuthResetPassword(t *testing.T) {
	mailer := &MockMailer{}
	config := &testAppConfig{activation: "manual", createData: true, mailer: mailer}
	provider := createLocalAuthProvider(config)
	server := createServer(t, fs.Post(
		"/user/recover/reset",
		provider.ResetPassword,
		&fs.Meta{Public: true},
	))

	// Case 1: Invalid token
	{
		req := httptest.NewRequest(
			"POST", "/user/recover/reset",
			bytes.NewReader([]byte(`{"token": "123"}`)),
		)
		resp, _ := server.Test(req)
		defer func() { assert.NoError(t, resp.Body.Close()) }()
		assert.Equal(t, 400, resp.StatusCode)
	}

	// Case 2: Invalid password (not match)
	{
		token := utils.Must(utils.CreateConfirmationToken(1, config.key))
		req := httptest.NewRequest(
			"POST", "/user/recover/reset",
			bytes.NewReader([]byte(`{"token": "`+token+`", "password": "123", "confirm_password": "1234"}`)),
		)
		resp, _ := server.Test(req)
		defer func() { assert.NoError(t, resp.Body.Close()) }()
		assert.Equal(t, 422, resp.StatusCode)
		assert.Contains(t, utils.Must(utils.ReadCloserToString(resp.Body)), auth.MSG_INVALID_PASSWORD)
	}

	// Case 3: Success
	{
		token := utils.Must(utils.CreateConfirmationToken(1, config.key))
		req := httptest.NewRequest(
			"POST", "/user/recover/reset",
			bytes.NewReader([]byte(`{"token": "`+token+`", "password": "123", "confirm_password": "123"}`)),
		)
		resp, _ := server.Test(req)
		defer func() { assert.NoError(t, resp.Body.Close()) }()
		assert.Equal(t, 200, resp.StatusCode)
	}

	// Case 4: Update password failed
	{
		token := utils.Must(utils.CreateConfirmationToken(1, config.key))
		assert.NoError(t, config.db.Close())
		req := httptest.NewRequest(
			"POST", "/user/recover/reset",
			bytes.NewReader([]byte(`{"token": "`+token+`", "password": "123", "confirm_password": "123"}`)),
		)
		resp, _ := server.Test(req)
		defer func() { assert.NoError(t, resp.Body.Close()) }()
		assert.Equal(t, 500, resp.StatusCode)
		assert.Contains(t, utils.Must(utils.ReadCloserToString(resp.Body)), auth.MSG_USER_SAVE_ERROR)
	}
}

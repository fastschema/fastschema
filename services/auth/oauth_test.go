package authservice_test

import (
	"net/http/httptest"
	"testing"

	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestOauthProvider(t *testing.T) {
	testApp := createTestApp(t)
	server := testApp.restResolver.Server()
	assert.NotNil(t, testApp)
	assert.NotNil(t, server)

	// Case 1: provider not found
	req := httptest.NewRequest("GET", "/api/auth/invalidprovider/login", nil)
	resp := utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, "404 Not Found", resp.Status)
	assert.Contains(t, utils.Must(utils.ReadCloserToString(resp.Body)), `"message":"invalid auth provider"`)

	// Case 2: login should redirect to the auth provider
	req = httptest.NewRequest("GET", "/api/auth/testauthprovider/login", nil)
	resp = utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, "302 Found", resp.Status)
	assert.Equal(t, "http://auth.example.local?callback=http://localhost:8000/auth/testauthprovider/callback", resp.Header.Get("Location"))

	// Case 3: callback error invalid auth provider
	req = httptest.NewRequest("GET", "/api/auth/invalidprovider/callback", nil)
	resp = utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, "404 Not Found", resp.Status)
	assert.Contains(t, utils.Must(utils.ReadCloserToString(resp.Body)), `"message":"invalid auth provider"`)

	// Case 4: callback error due to provider error
	req = httptest.NewRequest("GET", "/api/auth/testauthprovider/callback?error=1", nil)
	resp = utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, "500 Internal Server Error", resp.Status)

	// Case 5: callback success
	req = httptest.NewRequest("GET", "/api/auth/testauthprovider/callback", nil)
	resp = utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	response := utils.Must(utils.ReadCloserToString(resp.Body))
	assert.Equal(t, "200 OK", resp.Status)
	assert.Contains(t, response, `"token":`)
	assert.Contains(t, response, `"expires":`)

	// Case 6: callback with nil user
	req = httptest.NewRequest("GET", "/api/auth/testauthprovider/callback?niluser=1", nil)
	resp = utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, "401 Unauthorized", resp.Status)
}
func TestMe(t *testing.T) {
	testApp := createTestApp(t)
	server := testApp.restResolver.Server()
	assert.NotNil(t, testApp)
	assert.NotNil(t, server)

	// Case 1: Unauthorized request
	req := httptest.NewRequest("GET", "/api/auth/me", nil)
	resp := utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, "401 Unauthorized", resp.Status)

	// Case 2: Not found user
	req = httptest.NewRequest("GET", "/api/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+testApp.notFoundUserToken)
	resp = utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, "401 Unauthorized", resp.Status)

	// Case 3: Valid user
	req = httptest.NewRequest("GET", "/api/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+testApp.normalUserToken)
	resp = utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	response := utils.Must(utils.ReadCloserToString(resp.Body))
	assert.Equal(t, "200 OK", resp.Status)
	assert.Contains(t, response, `"username":"normaluser"`)
}

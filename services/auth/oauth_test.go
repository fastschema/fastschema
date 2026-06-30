package authservice_test

import (
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/stretchr/testify/assert"
)

// TestOauthProvider exercises the social login flow end to end. Since the CSRF
// fix, every callback verifies the signed `state` carrier: a missing or
// tampered state is rejected with 401, and a valid round-tripped state lets the
// callback complete (acceptance: normal social login works AND now verifies
// state).
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

	// Case 2: login redirects to the provider, carrying a signed state carrier
	req = httptest.NewRequest("GET", "/api/auth/testauthprovider/login", nil)
	resp = utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, "302 Found", resp.Status)
	location := resp.Header.Get("Location")
	assert.Contains(t, location, "http://auth.example.local?callback=")
	state := utils.Must(url.Parse(location)).Query().Get("state")
	assert.NotEmpty(t, state)

	// Case 3: callback error invalid auth provider
	req = httptest.NewRequest("GET", "/api/auth/invalidprovider/callback", nil)
	resp = utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, "404 Not Found", resp.Status)
	assert.Contains(t, utils.Must(utils.ReadCloserToString(resp.Body)), `"message":"invalid auth provider"`)

	// Case 4: valid state, provider error -> 500 (state passes, exchange fails)
	req = httptest.NewRequest("GET", "/api/auth/testauthprovider/callback?error=1&state="+url.QueryEscape(state), nil)
	resp = utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, "500 Internal Server Error", resp.Status)

	// Case 5: callback success with a valid state -> JWT JSON (legacy mode)
	req = httptest.NewRequest("GET", "/api/auth/testauthprovider/callback?state="+url.QueryEscape(state), nil)
	resp = utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	response := utils.Must(utils.ReadCloserToString(resp.Body))
	assert.Equal(t, "200 OK", resp.Status)
	assert.Contains(t, response, `"token":`)
	assert.Contains(t, response, `"expires":`)

	// Case 6: callback with nil user (valid state) -> 401
	req = httptest.NewRequest("GET", "/api/auth/testauthprovider/callback?niluser=1&state="+url.QueryEscape(state), nil)
	resp = utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, "401 Unauthorized", resp.Status)

	// Case 7: callback WITHOUT state -> 401 (CSRF hole closed)
	req = httptest.NewRequest("GET", "/api/auth/testauthprovider/callback", nil)
	resp = utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, "401 Unauthorized", resp.Status)
	assert.Contains(t, utils.Must(utils.ReadCloserToString(resp.Body)), "invalid auth state")

	// Case 8: callback with a tampered state -> 401
	req = httptest.NewRequest("GET", "/api/auth/testauthprovider/callback?state=deadbeef", nil)
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

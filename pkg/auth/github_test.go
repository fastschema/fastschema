package auth_test

import (
	"net/http"
	"testing"

	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/auth"
	"github.com/stretchr/testify/assert"
)

func TestNewGithubAuthProvider(t *testing.T) {
	redirectURL := "http://localhost:8080/api/auth/github/callback"
	authProvider, err := auth.NewGithubAuthProvider(fs.Map{}, redirectURL)
	assert.Error(t, err)
	assert.Nil(t, authProvider)

	authProvider = createGithubAuth()
	assert.NotNil(t, authProvider)
	assert.Equal(t, "github", authProvider.Name())
}

func TestGithubLogin(t *testing.T) {
	mockContext := &mockContext{}
	ga := createGithubAuth()
	_, err := ga.Login(mockContext)
	assert.NoError(t, err)
	assert.Contains(t, mockContext.rediectURL, "github.com/login/oauth/authorize")
	assert.Contains(t, mockContext.rediectURL, "client_id")
	assert.Contains(t, mockContext.rediectURL, "redirect_uri")
}

func TestGithubCallbackNoCode(t *testing.T) {
	ga := createGithubAuth()
	_, err := ga.Callback(&mockContext{})
	assert.ErrorContains(t, err, "callback code is empty")
}

func TestGithubAuthCallbackError(t *testing.T) {
	// Case 1: Access token server error
	accessTokenServer := createAuthProviderTestSever(func(w RW) {
		w.WriteHeader(http.StatusBadRequest)
	})
	defer accessTokenServer.Close()

	ga := createGithubAuth(fs.Map{"access_token_url": accessTokenServer.URL})
	_, err := ga.Callback(&mockContext{args: map[string]string{"code": "mockCode"}})
	assert.ErrorContains(t, err, "request failed with status code")

	// Case 2: Access token success and get user error
	accessTokenServer = createAuthProviderTestSever(func(w RW) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"access_token": "mock"}`))
	})
	defer accessTokenServer.Close()

	config := fs.Map{"access_token_url": accessTokenServer.URL}
	getUserServer := createAuthProviderTestSever(func(w RW) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	defer getUserServer.Close()
	config["user_info_url"] = getUserServer.URL
	ga = createGithubAuth(config)
	_, err = ga.Callback(&mockContext{args: map[string]string{"code": "mockCode"}})
	assert.ErrorContains(t, err, "request failed with status code")
}

func TestGithubAuthCallbackSuccess(t *testing.T) {
	accessTokenServer := createAuthProviderTestSever(func(w RW) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"access_token": "mock"}`))
	})
	defer accessTokenServer.Close()
	config := fs.Map{"access_token_url": accessTokenServer.URL}

	getUserServer := createAuthProviderTestSever(func(w RW) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"login": "testuser",
			"id": 12345,
			"avatar_url": "https://example.com/avatar.jpg",
			"name": "Test User",
			"blog": "https://example.com",
			"email": "testuser@site.local",
			"bio": "Test user bio"
		}`))
	})
	defer getUserServer.Close()
	config["user_info_url"] = getUserServer.URL
	ga := createGithubAuth(config)
	user, err := ga.Callback(&mockContext{args: map[string]string{"code": "mockCode"}})
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "testuser", user.Username)
}

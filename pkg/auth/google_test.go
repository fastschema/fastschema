package auth_test

import (
	"net/http"
	"testing"

	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/auth"
	"github.com/stretchr/testify/assert"
)

func TestNewGoogleAuthProvider(t *testing.T) {
	redirectURL := "http://localhost:8080/callback"
	authProvider, err := auth.NewGoogleAuthProvider(fs.Map{}, redirectURL)
	assert.Error(t, err)
	assert.Nil(t, authProvider)

	authProvider = createGoogleAuth()
	assert.NotNil(t, authProvider)
	assert.Equal(t, "google", authProvider.Name())
}

func TestGoogleLogin(t *testing.T) {
	mockContext := &mockContext{}
	ga := createGoogleAuth()
	_, err := ga.Login(mockContext)
	assert.NoError(t, err)
	assert.Contains(t, mockContext.rediectURL, "accounts.google.com/o/oauth2/auth")
	assert.Contains(t, mockContext.rediectURL, "client_id")
	assert.Contains(t, mockContext.rediectURL, "redirect_uri")
}

func TestGoogleCallbackNoCode(t *testing.T) {
	ga := createGoogleAuth()
	_, err := ga.Callback(&mockContext{})
	assert.ErrorContains(t, err, "callback code is empty")
}

func TestGoogleAuthCallbackError(t *testing.T) {
	// Case 1: Access token server error
	accessTokenServer := createAuthProviderTestSever(func(w RW) {
		w.WriteHeader(http.StatusBadRequest)
	})
	defer accessTokenServer.Close()
	ga := createGoogleAuth(fs.Map{"access_token_url": accessTokenServer.URL})
	_, err := ga.Callback(&mockContext{args: map[string]string{"code": "mockCode"}})
	assert.ErrorContains(t, err, "cannot fetch token")

	// Case 2: Access token success and get user error
	accessTokenServer = createAuthProviderTestSever(func(w RW) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"access_token": "mock"}`))
	})
	defer accessTokenServer.Close()
	config := fs.Map{"access_token_url": accessTokenServer.URL}

	getUserServer := createAuthProviderTestSever(func(w RW) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	defer getUserServer.Close()
	config["user_info_url"] = getUserServer.URL + "?access_token="
	ga = createGoogleAuth(config)
	_, err = ga.Callback(&mockContext{args: map[string]string{"code": "mockCode"}})
	assert.ErrorContains(t, err, "request failed with status code")
}

func TestGoogleAuthCallbackSuccess(t *testing.T) {
	accessTokenServer := createAuthProviderTestSever(func(w RW) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"access_token": "mock"}`))
	})
	defer accessTokenServer.Close()
	config := fs.Map{"access_token_url": accessTokenServer.URL}

	getUserServer := createAuthProviderTestSever(func(w RW) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"email": "testuser@site.local",
			"id": "12345",
			"name": "Test User"
		}`))
	})
	defer getUserServer.Close()
	config["user_info_url"] = getUserServer.URL + "?access_token="

	ga := createGoogleAuth(config)
	user, err := ga.Callback(&mockContext{args: map[string]string{"code": "mockCode"}})
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "testuser@site.local", user.Username)
	assert.Equal(t, "testuser@site.local", user.Email)
	assert.Equal(t, "testuser@site.local", user.ProviderUsername)
	assert.Equal(t, "12345", user.ProviderID)
	assert.Equal(t, "google", user.Provider)
	assert.True(t, user.Active)
}

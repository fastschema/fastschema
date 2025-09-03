package fs_test

import (
	"testing"

	"github.com/fastschema/fastschema/fs"
	"github.com/stretchr/testify/assert"
)

func TestRegisterAuthProviderMakerErrorNil(t *testing.T) {
	defer func() {
		r := recover()
		assert.Equal(t, "auth: Register auth provider is nil", r)
	}()
	fs.RegisterAuthProviderMaker("test", nil)
}

func TestRegisterAuthProviderMakerErrorDup(t *testing.T) {
	defer func() {
		r := recover()
		assert.Equal(t, "auth: Register called twice for auth provider test", r)
	}()
	fs.RegisterAuthProviderMaker("test", func(fs.Map, string) (fs.AuthProvider, error) {
		return nil, nil
	})
	fs.RegisterAuthProviderMaker("test", func(fs.Map, string) (fs.AuthProvider, error) {
		return nil, nil
	})
}

func TestCreateAuthProviderSuccess(t *testing.T) {
	fs.RegisterAuthProviderMaker("emptyauthprovider", func(fs.Map, string) (fs.AuthProvider, error) {
		return nil, nil
	})
	assert.Contains(t, fs.AuthProviders(), "emptyauthprovider")
}

type TestAuthProvider struct{}

func (ta *TestAuthProvider) Name() string {
	return "testauthprovider"
}

func (ta *TestAuthProvider) Login(fs.Context) (any, error) {
	return nil, nil
}

func (ta *TestAuthProvider) Callback(fs.Context) (*fs.User, error) {
	return nil, nil
}

func (ta *TestAuthProvider) Form(fs.Context) (*fs.User, error) {
	return nil, nil
}

func TestCreateAuthProvider(t *testing.T) {
	invalidProvider, err := fs.CreateAuthProvider("invalidprovider", nil, "")
	assert.Nil(t, invalidProvider)
	assert.Equal(t, "auth: unknown auth provider \"invalidprovider\"", err.Error())

	fs.RegisterAuthProviderMaker("testauthprovider", func(fs.Map, string) (fs.AuthProvider, error) {
		return &TestAuthProvider{}, nil
	})
	provider, err := fs.CreateAuthProvider("testauthprovider", nil, "")
	assert.Nil(t, err)
	assert.Equal(t, "testauthprovider", provider.Name())
}

func TestAuthConfigClone(t *testing.T) {
	ac := &fs.AuthConfig{
		EnabledProviders: []string{"test"},
		Providers: map[string]fs.Map{
			"test": {"key": "value"},
		},
	}
	clone := ac.Clone()
	assert.Equal(t, ac, clone)

	var nilAuthConfig *fs.AuthConfig
	assert.Nil(t, nilAuthConfig.Clone())
}

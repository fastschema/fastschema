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

func (ta *TestAuthProvider) VerifyIDToken(c fs.Context, p fs.IDToken) (*fs.User, error) {
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

func TestOTPConfigClone(t *testing.T) {
	// Test nil clone
	var nilOTP *fs.OTPConfig
	assert.Nil(t, nilOTP.Clone())

	// Test non-nil clone
	otp := &fs.OTPConfig{
		Enabled:     true,
		Length:      8,
		Expiration:  600,
		MaxAttempts: 5,
	}
	clone := otp.Clone()
	assert.Equal(t, otp.Enabled, clone.Enabled)
	assert.Equal(t, otp.Length, clone.Length)
	assert.Equal(t, otp.Expiration, clone.Expiration)
	assert.Equal(t, otp.MaxAttempts, clone.MaxAttempts)
}

func TestOTPConfigGetLength(t *testing.T) {
	// Test nil receiver - should return default 6
	var nilOTP *fs.OTPConfig
	assert.Equal(t, 6, nilOTP.GetLength())

	// Test zero length - should return default 6
	otp := &fs.OTPConfig{Length: 0}
	assert.Equal(t, 6, otp.GetLength())

	// Test negative length - should return default 6
	otp = &fs.OTPConfig{Length: -1}
	assert.Equal(t, 6, otp.GetLength())

	// Test custom length
	otp = &fs.OTPConfig{Length: 8}
	assert.Equal(t, 8, otp.GetLength())
}

func TestOTPConfigGetExpiration(t *testing.T) {
	// Test nil receiver - should return default 300
	var nilOTP *fs.OTPConfig
	assert.Equal(t, 300, nilOTP.GetExpiration())

	// Test zero expiration - should return default 300
	otp := &fs.OTPConfig{Expiration: 0}
	assert.Equal(t, 300, otp.GetExpiration())

	// Test negative expiration - should return default 300
	otp = &fs.OTPConfig{Expiration: -1}
	assert.Equal(t, 300, otp.GetExpiration())

	// Test custom expiration
	otp = &fs.OTPConfig{Expiration: 600}
	assert.Equal(t, 600, otp.GetExpiration())
}

func TestOTPConfigGetMaxAttempts(t *testing.T) {
	// Test nil receiver - should return default 3
	var nilOTP *fs.OTPConfig
	assert.Equal(t, 3, nilOTP.GetMaxAttempts())

	// Test zero max attempts - should return default 3
	otp := &fs.OTPConfig{MaxAttempts: 0}
	assert.Equal(t, 3, otp.GetMaxAttempts())

	// Test negative max attempts - should return default 3
	otp = &fs.OTPConfig{MaxAttempts: -1}
	assert.Equal(t, 3, otp.GetMaxAttempts())

	// Test custom max attempts
	otp = &fs.OTPConfig{MaxAttempts: 5}
	assert.Equal(t, 5, otp.GetMaxAttempts())
}

func TestAuthConfigCloneWithOTP(t *testing.T) {
	ac := &fs.AuthConfig{
		EnabledProviders: []string{"test"},
		Providers: map[string]fs.Map{
			"test": {"key": "value"},
		},
		OTP: &fs.OTPConfig{
			Enabled:     true,
			Length:      8,
			Expiration:  600,
			MaxAttempts: 5,
		},
	}
	clone := ac.Clone()

	assert.NotNil(t, clone.OTP)
	assert.Equal(t, ac.OTP.Enabled, clone.OTP.Enabled)
	assert.Equal(t, ac.OTP.Length, clone.OTP.Length)
	assert.Equal(t, ac.OTP.Expiration, clone.OTP.Expiration)
	assert.Equal(t, ac.OTP.MaxAttempts, clone.OTP.MaxAttempts)
}

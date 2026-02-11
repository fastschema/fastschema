package auth

import (
	"testing"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/fs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateOTP(t *testing.T) {
	// Test default length (6)
	otp, err := GenerateOTP(0)
	require.NoError(t, err)
	assert.Len(t, otp, 6)

	// Test custom lengths
	for _, length := range []int{4, 6, 8, 10} {
		otp, err := GenerateOTP(length)
		require.NoError(t, err)
		assert.Len(t, otp, length)

		// Verify OTP contains only digits
		for _, c := range otp {
			assert.True(t, c >= '0' && c <= '9', "OTP should contain only digits")
		}
	}

	// Test that different OTPs are generated
	otps := make(map[string]bool)
	for i := 0; i < 100; i++ {
		otp, err := GenerateOTP(6)
		require.NoError(t, err)
		otps[otp] = true
	}
	// Should have at least 90 unique OTPs out of 100
	assert.Greater(t, len(otps), 90, "OTPs should be random")
}

func TestHashOTP(t *testing.T) {
	otp := "123456"

	hash, err := HashOTP(otp)
	require.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, otp, hash)

	// Different OTPs should produce different hashes
	hash2, err := HashOTP("654321")
	require.NoError(t, err)
	assert.NotEqual(t, hash, hash2)
}

func TestVerifyOTP(t *testing.T) {
	otp := "123456"
	hash, err := HashOTP(otp)
	require.NoError(t, err)

	// Correct OTP should verify
	assert.True(t, VerifyOTP(otp, hash))

	// Wrong OTP should not verify
	assert.False(t, VerifyOTP("654321", hash))
	assert.False(t, VerifyOTP("", hash))
	assert.False(t, VerifyOTP("12345", hash))   // Wrong length
	assert.False(t, VerifyOTP("1234567", hash)) // Wrong length
}

func TestCreateOTPEmail(t *testing.T) {
	mail := CreateOTPEmail("TestApp", "test@example.com", "123456", 5)

	assert.Equal(t, []string{"test@example.com"}, mail.To)
	assert.Contains(t, mail.Subject, "TestApp")
	assert.Contains(t, mail.Subject, "123456")
	assert.Contains(t, mail.Body, "123456")
	assert.Contains(t, mail.Body, "5 minutes")
	assert.Contains(t, mail.Body, "test") // Greeting extracted from email
}

func TestCreateOTPEmailWithFullEmail(t *testing.T) {
	mail := CreateOTPEmail("MyApp", "john.doe@company.org", "987654", 10)

	assert.Equal(t, []string{"john.doe@company.org"}, mail.To)
	assert.Contains(t, mail.Body, "john.doe") // Name from email
	assert.Contains(t, mail.Body, "987654")
	assert.Contains(t, mail.Body, "10 minutes")
}

func TestOTPRequestStruct(t *testing.T) {
	req := &OTPRequest{
		Email: "test@example.com",
	}
	assert.Equal(t, "test@example.com", req.Email)
}

func TestOTPResponseStruct(t *testing.T) {
	resp := &OTPResponse{
		Message:   "OTP sent",
		ExpiresIn: 300,
	}
	assert.Equal(t, "OTP sent", resp.Message)
	assert.Equal(t, 300, resp.ExpiresIn)
}

func TestOTPProviderName(t *testing.T) {
	provider := &OTPProvider{}
	assert.Equal(t, ProviderOTP, provider.Name())
	assert.Equal(t, "otp", provider.Name())
}

func TestOTPProviderIsEnabled(t *testing.T) {
	// Provider with nil otpConfig
	provider := &OTPProvider{}
	assert.False(t, provider.IsEnabled())

	// Provider with otpConfig returning nil
	provider2 := &OTPProvider{
		otpConfig: func() *fs.OTPConfig { return nil },
	}
	assert.False(t, provider2.IsEnabled())

	// Provider with disabled OTP
	provider3 := &OTPProvider{
		otpConfig: func() *fs.OTPConfig { return &fs.OTPConfig{Enabled: false} },
	}
	assert.False(t, provider3.IsEnabled())

	// Provider with enabled OTP
	provider4 := &OTPProvider{
		otpConfig: func() *fs.OTPConfig { return &fs.OTPConfig{Enabled: true} },
	}
	assert.True(t, provider4.IsEnabled())
}

func TestOTPProviderAuthProviderMethods(t *testing.T) {
	provider := &OTPProvider{}

	// Login should return nil, nil
	result, err := provider.Login(nil)
	assert.Nil(t, result)
	assert.NoError(t, err)

	// Callback should return nil, nil
	user, err := provider.Callback(nil)
	assert.Nil(t, user)
	assert.NoError(t, err)

	// VerifyIDToken should return nil, nil
	user, err = provider.VerifyIDToken(nil, fs.IDToken{})
	assert.Nil(t, user)
	assert.NoError(t, err)
}

func TestNewOTPAuthProvider(t *testing.T) {
	provider, err := NewOTPAuthProvider(nil, "")
	require.NoError(t, err)
	assert.NotNil(t, provider)
	assert.IsType(t, &OTPProvider{}, provider)
}

func TestOTPProviderInit(t *testing.T) {
	provider := &OTPProvider{}

	// Initialize with mock functions
	dbFunc := func() db.Client { return nil }
	appNameFunc := func() string { return "TestApp" }
	mailerFunc := func(names ...string) fs.Mailer { return nil }
	otpConfigFunc := func() *fs.OTPConfig { return &fs.OTPConfig{Enabled: true} }

	provider.Init(dbFunc, appNameFunc, mailerFunc, otpConfigFunc)

	// Verify functions are set
	assert.Equal(t, "TestApp", provider.appName())
	assert.True(t, provider.IsEnabled())
}

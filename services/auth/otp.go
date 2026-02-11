package authservice

import (
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/auth"
)

// OTPRequestWrapper wraps the OTPProvider's RequestOTP method
func (as *AuthService) OTPRequestWrapper(provider *auth.OTPProvider) func(c fs.Context, req *auth.OTPRequest) (*auth.OTPResponse, error) {
	return func(c fs.Context, req *auth.OTPRequest) (*auth.OTPResponse, error) {
		return provider.RequestOTP(c, req)
	}
}

// OTPVerifyWrapper wraps the OTPProvider's VerifyOTP method and handles JWT token generation
func (as *AuthService) OTPVerifyWrapper(provider *auth.OTPProvider) func(c fs.Context, req *auth.OTPVerify) (*fs.JWTTokens, error) {
	return func(c fs.Context, req *auth.OTPVerify) (*fs.JWTTokens, error) {
		// Verify OTP using the provider
		user, err := provider.VerifyOTP(c, req)
		if err != nil {
			return nil, err
		}

		// Generate JWT tokens
		return as.GenerateJWTTokens(c, user)
	}
}

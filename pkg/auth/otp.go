package auth

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"

	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/utils"
)

// GenerateOTP generates a cryptographically secure numeric OTP of the specified length.
func GenerateOTP(length int) (string, error) {
	if length <= 0 {
		length = 6
	}

	// Maximum value for the OTP (e.g., 999999 for 6 digits)
	max := new(big.Int)
	max.Exp(big.NewInt(10), big.NewInt(int64(length)), nil)

	// Generate random number
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", fmt.Errorf("failed to generate OTP: %w", err)
	}

	// Format with leading zeros if necessary
	format := fmt.Sprintf("%%0%dd", length)
	return fmt.Sprintf(format, n), nil
}

// HashOTP creates a secure hash of the OTP for storage.
func HashOTP(otp string) (string, error) {
	return utils.GenerateHash(otp)
}

// VerifyOTP verifies an OTP against its hash.
// Returns true if the OTP matches, false otherwise.
func VerifyOTP(otp, hash string) bool {
	err := utils.CheckHash(otp, hash)
	return err == nil
}

// CreateOTPEmail creates an email with the OTP code for passwordless login.
func CreateOTPEmail(appName, email, otp string, expirationMinutes int) *fs.Mail {
	name := "there"
	parts := strings.Split(email, "@")
	if len(parts) > 0 && parts[0] != "" {
		name = parts[0]
	}

	bodyLines := []string{
		fmt.Sprintf(`Hey %s,`, name),
		fmt.Sprintf(`You requested to sign in to %s using a one-time password.`, appName),
		`Your verification code is:`,
		fmt.Sprintf(`<div style="font-size: 32px; font-weight: bold; letter-spacing: 8px; text-align: center; padding: 20px; background-color: #f5f5f5; border-radius: 8px; margin: 20px 0;">%s</div>`, otp),
		fmt.Sprintf(`This code will expire in %d minutes.`, expirationMinutes),
		`If you didn't request this code, you can safely ignore this email.`,
		`For security, never share this code with anyone.`,
		"Thanks,",
		appName,
	}

	return &fs.Mail{
		To:      []string{email},
		Subject: fmt.Sprintf("Your %s verification code: %s", appName, otp),
		Body: strings.Join(utils.Map(bodyLines, func(l string) string {
			return fmt.Sprintf("<p>%s</p>", l)
		}), "\r\n"),
	}
}

// OTPRequest represents a request for OTP passwordless login
type OTPRequest struct {
	Email string `json:"email"`
}

// OTPVerify represents the OTP verification request
type OTPVerify struct {
	SessionID string `json:"session_id"` // UUID session ID returned from OTP request
	OTP       string `json:"otp"`
}

// OTPResponse represents the response after requesting an OTP
type OTPResponse struct {
	Message   string `json:"message"`
	SessionID string `json:"session_id"` // UUID session ID for verification
	ExpiresIn int    `json:"expires_in"` // Expiration time in seconds
}

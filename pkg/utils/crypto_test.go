package utils

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRandomString(t *testing.T) {
	length := 10
	randomString := RandomString(length)
	assert.Equal(t, length, len(randomString))
}

func TestSecureRandomBytes(t *testing.T) {
	length := 10
	randomBytes, err := SecureRandomBytes(length)
	assert.NoError(t, err)
	assert.Equal(t, length, len(randomBytes))
}

func TestGenerateHash(t *testing.T) {
	// Test case 1: Non-empty input
	input1 := "password123"
	hash1, err1 := GenerateHash(input1)
	assert.NoError(t, err1)
	assert.NotEmpty(t, hash1)

	// Test case 2: Empty input
	input2 := ""
	_, err2 := GenerateHash(input2)
	assert.Error(t, err2)
}

func TestCheckHash(t *testing.T) {
	input := "password"
	hash, err := GenerateHash(input)
	assert.NoError(t, err)
	err = CheckHash(input, hash)
	assert.NoError(t, err)

	// Test case with invalid hash
	invalidHash := "invalid_hash"
	err = CheckHash(input, invalidHash)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid hash")

	// Test case with incorrect input
	incorrectInput := "incorrect_password"
	err = CheckHash(incorrectInput, hash)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid hash")

	// Test case with wrong format
	wrongFormat := "$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s"
	err = CheckHash(input, wrongFormat)
	assert.Error(t, err)

	// Test case with invalid base64 in salt
	base64Error := "$argon2id$v=19$m=65536,t=3,p=4$invalidBase64$TSj7ccmNoowqotWys56GYKAIlzA8opVHU7icgWLI9jk"
	err = CheckHash(input, base64Error)
	assert.Error(t, err)

	// Test case with invalid base64 in hash
	base64Error = "$argon2id$v=19$m=65536,t=3,p=4$TSj7ccmNoowqotWys56GYKAIlzA8opVHU7icgWLI9jk$invalidBase64"
	err = CheckHash(input, base64Error)
	assert.Error(t, err)
}

func TestEncrypt(t *testing.T) {
	// Test case 1: Valid encryption
	plaintext := "Hello, World!"
	key := "0123456789abcdef" // 16 bytes key for AES-128
	encryptedText, err := Encrypt(plaintext, key)
	assert.NoError(t, err)
	assert.NotEmpty(t, encryptedText)

	// Test case 2: Invalid key length
	invalidKey := "short"
	_, err = Encrypt(plaintext, invalidKey)
	assert.Error(t, err)

	// Test case 3: Empty plaintext
	emptyPlaintext := ""
	encryptedText, err = Encrypt(emptyPlaintext, key)
	assert.NoError(t, err)
	assert.NotEmpty(t, encryptedText)
}

func TestDecrypt(t *testing.T) {
	// Test case 1: Valid decryption
	plaintext := "Hello, World!"
	key := "0123456789abcdef" // 16 bytes key for AES-128
	encryptedText, err := Encrypt(plaintext, key)
	assert.NoError(t, err)
	assert.NotEmpty(t, encryptedText)

	decryptedText, err := Decrypt(encryptedText, key)
	assert.NoError(t, err)
	assert.Equal(t, plaintext, decryptedText)

	// Test case 2: Invalid key length
	invalidKey := "short"
	_, err = Decrypt(encryptedText, invalidKey)
	assert.Error(t, err)

	// Test case 3: Empty encrypted string
	emptyEncryptedText := ""
	_, err = Decrypt(emptyEncryptedText, key)
	assert.Error(t, err)

	// Test case 4: Invalid hex string
	invalidHex := "invalid_hex"
	_, err = Decrypt(invalidHex, key)
	assert.Error(t, err)

	// Test case 5: Ciphertext too short
	shortCiphertext := "00"
	_, err = Decrypt(shortCiphertext, key)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ciphertext too short")
}

func TestCreateConfirmationToken(t *testing.T) {
	// Test case 1: Valid token creation
	userID := uint64(12345)
	key := "0123456789abcdef" // 16 bytes key for AES-128
	token, err := CreateConfirmationToken(userID, key)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// Test case 2: Token creation with expiration time
	expirationTime := time.Now().Add(time.Hour * 48)
	token, err = CreateConfirmationToken(userID, key, expirationTime)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// Test case 3: Invalid key length
	invalidKey := "short"
	_, err = CreateConfirmationToken(userID, invalidKey)
	assert.Error(t, err)
}
func TestParseConfirmationToken(t *testing.T) {
	// Test case 1: Valid token
	userID := uint64(12345)
	key := "0123456789abcdef" // 16 bytes key for AES-128
	expirationTime := time.Now().Add(time.Hour * 24)
	token, err := CreateConfirmationToken(userID, key, expirationTime)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	confirmationData, err := ParseConfirmationToken(token, key)
	assert.NoError(t, err)
	assert.Equal(t, userID, confirmationData.UserID)
	assert.Equal(t, expirationTime.UnixMicro(), confirmationData.Exp)

	// Test case 2: Invalid token (decryption error)
	invalidToken := "invalid_token"
	_, err = ParseConfirmationToken(invalidToken, key)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid token")

	// Test case 3: Invalid token format (missing parts)
	invalidTokenFormat := "invalid_format"
	_, err = ParseConfirmationToken(invalidTokenFormat, key)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid token")

	// Test case 4: Invalid userID in token
	invalidUserIDToken, err := Encrypt("invalid_userID_1234567890", key)
	assert.NoError(t, err)
	_, err = ParseConfirmationToken(invalidUserIDToken, key)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid token")

	// Test case 5: Invalid expiration time in token
	invalidExpTimeToken, err := Encrypt(fmt.Sprintf("%d_invalid_exp", userID), key)
	assert.NoError(t, err)
	_, err = ParseConfirmationToken(invalidExpTimeToken, key)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid token")

	// Test case 6: Invalid userID format
	token, err = Encrypt(
		fmt.Sprintf("invalid_%d", time.Now().Add(time.Hour*24).UnixMicro()),
		key,
	)
	assert.NoError(t, err)
	_, err = ParseConfirmationToken(token, key)
	assert.Error(t, err)

	// Test case 7: Invalid expiration time format
	token, err = Encrypt(
		fmt.Sprintf("%d_invalid", userID),
		key,
	)
	assert.NoError(t, err)
	_, err = ParseConfirmationToken(token, key)
	assert.Error(t, err)
}

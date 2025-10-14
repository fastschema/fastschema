package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/fastschema/fastschema/pkg/errors"
	"golang.org/x/crypto/argon2"
)

// Random string generation
// source: https://stackoverflow.com/a/35615565/2422005
const (
	letterBytes   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ" // 52 possibilities
	letterIdxBits = 6                                                      // 6 bits to represent 64 possibilities / indexes
	letterIdxMask = 1<<letterIdxBits - 1                                   // All 1-bits, as many as letterIdxBits
)

// SecureRandomBytes returns the requested number of bytes using crypto/rand
func SecureRandomBytes(length int) ([]byte, error) {
	var randomBytes = make([]byte, length)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, err
	}
	return randomBytes, nil
}

func RandomString(length int) string {
	result := make([]byte, length)
	bufferSize := int(float64(length) * 1.3)
	for i, j, randomBytes := 0, 0, []byte{}; i < length; j++ {
		if j%bufferSize == 0 {
			randomBytes, _ = SecureRandomBytes(bufferSize)
		}
		if idx := int(randomBytes[j%length] & letterIdxMask); idx < len(letterBytes) {
			result[i] = letterBytes[idx]
			i++
		}
	}

	return string(result)
}

type HashConfig struct {
	Iterations uint32
	Memory     uint32
	KeyLen     uint32
	Threads    uint8
}

func GenerateHash(input string) (string, error) {
	if input == "" {
		return "", errors.BadRequest("hash: input cannot be empty")
	}

	salt := make([]byte, 16)
	cfg := &HashConfig{
		Iterations: 3,
		Memory:     64 * 1024,
		Threads:    4,
		KeyLen:     32,
	}

	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	hash := argon2.IDKey([]byte(input), salt, cfg.Iterations, cfg.Memory, cfg.Threads, cfg.KeyLen)
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)
	format := "$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s"
	full := fmt.Sprintf(format, argon2.Version, cfg.Memory, cfg.Iterations, cfg.Threads, b64Salt, b64Hash)

	return full, nil
}

func CheckHash(input, hash string) error {
	var err error
	var salt []byte
	var decodedHash []byte
	parts := strings.Split(hash, "$")
	cfg := &HashConfig{}

	if len(parts) != 6 {
		return errors.New("invalid hash")
	}

	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &cfg.Memory, &cfg.Iterations, &cfg.Threads); err != nil {
		return err
	}

	if salt, err = base64.RawStdEncoding.DecodeString(parts[4]); err != nil {
		return err
	}

	if decodedHash, err = base64.RawStdEncoding.DecodeString(parts[5]); err != nil {
		return err
	}

	cfg.KeyLen = uint32(len(decodedHash))
	comparisonHash := argon2.IDKey([]byte(input), salt, cfg.Iterations, cfg.Memory, cfg.Threads, cfg.KeyLen)
	valid := subtle.ConstantTimeCompare(decodedHash, comparisonHash) == 1

	if !valid {
		return errors.New("invalid hash")
	}

	return nil
}

func Encrypt(stringToEncrypt string, key string) (string, error) {
	plaintext := []byte(stringToEncrypt)
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := aesGCM.Seal(nonce, nonce, plaintext, nil)
	return hex.EncodeToString(ciphertext), nil
}

func Decrypt(encryptedString string, key string) (string, error) {
	enc, err := hex.DecodeString(encryptedString)

	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	if len(enc) < aesGCM.NonceSize() {
		return "", errors.New("ciphertext too short")
	}

	nonceSize := aesGCM.NonceSize()
	nonce, ciphertext := enc[:nonceSize], enc[nonceSize:]

	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

type ConfirmationData struct {
	UserID uint64 `json:"user_id"`
	Exp    int64  `json:"exp"`
}

func CreateConfirmationToken(
	userID uint64,
	key string,
	expiresAt ...time.Time,
) (string, error) {
	var exp time.Time
	if len(expiresAt) == 0 {
		exp = time.Now().Add(time.Hour * 24)
	} else {
		exp = expiresAt[0]
	}

	activationToken, err := Encrypt(
		fmt.Sprintf("%d_%d", userID, exp.UnixMicro()),
		key,
	)
	if err != nil {
		return "", err
	}

	return activationToken, nil
}

func ParseConfirmationToken(token, key string) (*ConfirmationData, error) {
	decryptedData, err := Decrypt(token, key)
	if err != nil {
		return nil, errors.BadRequest("Invalid token")
	}

	parts := strings.Split(decryptedData, "_")
	if len(parts) != 2 {
		return nil, errors.BadRequest("Invalid token")
	}

	userID, err := strconv.ParseUint(parts[0], 10, 64)
	if err != nil {
		return nil, errors.BadRequest("Invalid token")
	}

	expTime, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return nil, errors.BadRequest("Invalid token")
	}

	return &ConfirmationData{
		UserID: userID,
		Exp:    expTime,
	}, nil
}

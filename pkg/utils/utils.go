package utils

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/argon2"
)

// Random string generation
// source: https://stackoverflow.com/a/35615565/2422005
const (
	letterBytes   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ" // 52 possibilities
	letterIdxBits = 6                                                      // 6 bits to represent 64 possibilities / indexes
	letterIdxMask = 1<<letterIdxBits - 1                                   // All 1-bits, as many as letterIdxBits
)

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

// SecureRandomBytes returns the requested number of bytes using crypto/rand
func SecureRandomBytes(length int) ([]byte, error) {
	var randomBytes = make([]byte, length)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, err
	}
	return randomBytes, nil
}

func Map[T any, R any](slice []T, mapper func(T) R) []R {
	var result = If(
		cap(slice) > len(slice),
		make([]R, 0, cap(slice)),
		make([]R, 0, len(slice)),
	)

	for _, e := range slice {
		result = append(result, mapper(e))
	}
	return result
}

func Filter[T any](slice []T, predicate func(T) bool) []T {
	var result []T
	for _, e := range slice {
		if predicate(e) {
			result = append(result, e)
		}
	}
	return result
}

func Contains[T comparable](slice []T, element T) bool {
	for _, e := range slice {
		if e == element {
			return true
		}
	}
	return false
}

func SliceEqual[T comparable](slice1 []T, slice2 []T) bool {
	if len(slice1) != len(slice2) {
		return false
	}

	for i, e := range slice1 {
		if e != slice2[i] {
			return false
		}
	}

	return true
}

func SliceInsertBeforeElement[T comparable](slice []T, newElement T, checkIndexFn func(element T) bool) []T {
	var index = -1
	for i, e := range slice {
		if checkIndexFn(e) {
			index = i
		}
	}

	if index == -1 {
		return append(slice, newElement)
	}

	return append(slice[:index], append([]T{newElement}, slice[index:]...)...)
}

func If[T any](condition bool, ifTrue T, ifFalse T) T {
	if condition {
		return ifTrue
	}
	return ifFalse
}

func IfFn[T any](condition bool, ifTrue func() T, ifFalse func() T) T {
	if condition {
		return ifTrue()
	}
	return ifFalse()
}

func GetMapKeys[K comparable, V any](m map[K]V) []K {
	var keys []K
	for k := range m {
		keys = append(keys, k)
	}

	return keys
}

func GetMapValues[K comparable, V any](m map[K]V) []V {
	var values []V
	for _, v := range m {
		values = append(values, v)
	}

	return values
}

// Pick returns value in the nested map by the given paths in format: "path.to.value"
func Pick(obj any, path string, defaultValues ...any) any {
	var value = obj
	defaultValues = append(defaultValues, nil)
	parts := strings.Split(path, ".")

	for _, part := range parts {
		switch v := value.(type) {
		case []any:
			index, err := strconv.Atoi(part)
			if err != nil || index < 0 || index >= len(v) {
				return defaultValues[0]
			}
			value = v[index]
		case map[string]any:
			value = v[part]
		default:
			return defaultValues[0]
		}
	}

	if value == nil {
		return defaultValues[0]
	}

	return value
}

// EscapeQuery escapes the query string to be used in a regular expression.
// copied from ent test
func EscapeQuery(query string) string {
	rows := strings.Split(query, "\n")
	for i := range rows {
		rows[i] = strings.TrimPrefix(rows[i], " ")
	}
	query = strings.Join(rows, " ")
	return strings.TrimSpace(regexp.QuoteMeta(query)) + "$"
}

// IsNumber checks if the given any	is a number
func IsNumber(value any) bool {
	switch value.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		return true
	default:
		return false
	}
}

func Must[T any](value T, err error) T {
	if err != nil {
		panic(err)
	}
	return value
}

// IsValidBool check if the given value is a valid boolean.
func IsValidBool(v any) bool {
	switch v.(type) {
	case bool:
		return true
	default:
		return false
	}
}

// IsValidTime check if the given value is a valid time.
func IsValidTime(v any) bool {
	timeStringValue, ok := v.(string)
	if !ok {
		return false
	}

	_, err1 := time.Parse(time.RFC3339, timeStringValue)
	_, err2 := time.Parse(time.DateTime, timeStringValue)

	return err1 == nil || err2 == nil
}

// IsValidString check if the given value is a valid string.
func IsValidString(v any) bool {
	switch v.(type) {
	case string:
		return true
	default:
		return false
	}
}

// IsValidFloat check if the given value is a valid float.
func IsValidFloat(v any) bool {
	switch v.(type) {
	case float32, float64:
		return true
	default:
		if _, err := strconv.ParseFloat(fmt.Sprintf("%v", v), 64); err != nil {
			return false
		}
		return true
	}
}

// IsValidInt check if the given value is a valid integer.
func IsValidInt(v any) bool {
	switch v.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return true
	default:
		floatValue, err := strconv.ParseFloat(fmt.Sprintf("%v", v), 64)
		if err != nil {
			return false
		}
		if math.Trunc(floatValue) != floatValue {
			return false
		}
		return true
	}
}

// IsValidUInt check if the given value is a valid unsigned integer.
func IsValidUInt(v any) bool {
	if ok := IsValidInt(v); !ok {
		return false
	}

	switch v.(type) {
	case uint, uint8, uint16, uint32, uint64:
		return true
	case int, int8, int16, int32, int64:
		intValue, err := strconv.ParseInt(fmt.Sprintf("%v", v), 10, 64)
		if err != nil {
			return false
		}
		return intValue >= 0
	default:
		if floatValue, err := strconv.ParseFloat(fmt.Sprintf("%v", v), 64); err == nil {
			return floatValue >= 0
		}
	}

	return false
}

// WriteFile writes the given content to the given file path.
func WriteFile(filePath string, content string) error {
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err2 := f.WriteString(content)
	return err2
}

// AppendFile appends the given content to the given file path.
// If the file does not exist, it creates the file.
func AppendFile(filePath string, content string) error {
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err2 := f.WriteString(content)
	return err2
}

// IsFileExists checks if the given file path exists.
func IsFileExists(filePath string) bool {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return false
	}
	return true
}

func CopyFile(src string, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	return os.WriteFile(dst, data, 0600)
}

type HashConfig struct {
	Iterations uint32
	Memory     uint32
	KeyLen     uint32
	Threads    uint8
}

func GenerateHash(input string) (string, error) {
	if input == "" {
		return "", errors.New("hash: input cannot be empty")
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

func MkDirs(dirs ...string) error {
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0777); err != nil {
			return err
		}
	}
	return nil
}

func Env(name string, defaultValues ...string) string {
	value := os.Getenv(name)
	if value == "" && len(defaultValues) > 0 {
		return defaultValues[0]
	}

	return value
}

func ReadCloserToString(rc io.ReadCloser) (string, error) {
	defer rc.Close()
	data, err := io.ReadAll(rc)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

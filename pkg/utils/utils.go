package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/hjson/hjson-go/v4"
	"github.com/iancoleman/strcase"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

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

func Unique[T comparable](slice []T) []T {
	uniqueMap := make(map[T]struct{})
	var uniqueSlice []T

	for _, e := range slice {
		if _, exists := uniqueMap[e]; !exists {
			uniqueMap[e] = struct{}{}
			uniqueSlice = append(uniqueSlice, e)
		}
	}
	return uniqueSlice
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

func MapKeys[K comparable, V any](m map[K]V) []K {
	var keys []K
	for k := range m {
		keys = append(keys, k)
	}

	return keys
}

func MapValues[K comparable, V any](m map[K]V) []V {
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
	switch v := v.(type) {
	case uint, uint8, uint16, uint32, uint64:
		return true
	case int:
		return v >= 0
	case int8:
		return v >= 0
	case int16:
		return v >= 0
	case int32:
		return v >= 0
	case int64:
		return v >= 0
	case float32:
		return v >= 0 && math.Trunc(float64(v)) == float64(v)
	case float64:
		return v >= 0 && math.Trunc(v) == v
	default:
		floatValue, err := strconv.ParseFloat(fmt.Sprintf("%v", v), 64)
		return err == nil && floatValue >= 0 && math.Trunc(floatValue) == floatValue
	}
}

// IsValidEmail check if the given value is a valid email.
func IsValidEmail(v any) bool {
	email, ok := v.(string)
	if !ok {
		return false
	}

	return regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`).MatchString(email)
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

func EnvInt(name string, defaultValues ...int) int {
	value := os.Getenv(name)
	if value == "" && len(defaultValues) > 0 {
		return defaultValues[0]
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		if len(defaultValues) > 0 {
			return defaultValues[0]
		}
		return 0
	}

	return intValue
}

func ReadCloserToString(rc io.ReadCloser) (string, error) {
	defer rc.Close()
	data, err := io.ReadAll(rc)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// Capitalize the first letter of the given string
func Capitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// Title converts a string to title case
func Title(s string) string {
	s = Capitalize(s)
	// replace underscores with spaces
	s = strings.ReplaceAll(s, "_", " ")
	// replace dashes with spaces
	s = strings.ReplaceAll(s, "-", " ")
	// replace multiple spaces with a single space
	s = regexp.MustCompile(`\s+`).ReplaceAllString(s, " ")

	return cases.Title(language.English).String(s)
}

// ToSnakeCase converts a string from capital case to snake case
//
//		For example:
//	 	"ToSnakeCase" -> "create_snake_case"
//	 	"userID" -> "user_id"
//	 	"HTTPResponse" -> "http_response"
func ToSnakeCase(s string) string {
	return strcase.ToSnake(s)
}

// GetStructFieldName returns the name of a struct field based on the provided reflect.StructField and tag name.
//
//	If no tag name is provided, it defaults to "json".
//	If the field name is "-", it returns an empty string.
func GetStructFieldName(field reflect.StructField, fromTags ...string) string {
	if len(fromTags) == 0 {
		fromTags = []string{"json"}
	}
	fieldName := field.Name
	fieldTag := field.Tag.Get(fromTags[0])

	if fieldTag != "" {
		fieldName = strings.TrimSpace(strings.Split(fieldTag, ",")[0])
		if fieldName == "-" {
			fieldName = ""
		}
	}

	return fieldName
}

// ParseStructFieldTag parses the struct field tag and returns a map of tag key-value pairs.
//
//	For example, if the struct tag is:
//	`fs:"name=title;label=Title;multiple;unique;optional;sortable;filterable;size=255"`,
//	then ParseStructFieldTag(field, "fs") will return:
//	{
//		"name": "title",
//		"label": "Title",
//		"multiple": "",
//		"unique": "",
//		"optional": "",
//		"sortable": "",
//		"filterable": "",
//		"size": "255",
//	}
func ParseStructFieldTag(field reflect.StructField, tagName string) map[string]string {
	tag := field.Tag.Get(tagName)
	tagMap := map[string]string{}
	props := Map(strings.Split(tag, ";"), strings.TrimSpace)

	for _, prop := range props {
		parts := strings.Split(prop, "=")
		propName := strings.TrimSpace(parts[0])

		if propName == "" {
			continue
		}

		if len(parts) == 1 {
			parts = append(parts, "")
		}

		tagMap[propName] = strings.TrimSpace(parts[1])
	}

	return tagMap
}

// ParseHJSON parses the given input byte slice as HJSON and returns the result as type T.
func ParseHJSON[T any](input []byte) (T, error) {
	var v T
	err := hjson.Unmarshal(input, &v)
	return v, err
}

// CreateSwaggerUIPage creates a simple HTML page that serves the Swagger UI
func CreateSwaggerUIPage(specURL string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="utf-8" />
	<meta name="viewport" content="width=device-width, initial-scale=1" />
	<title>Fastschema Open OAS 3.1</title>
	<link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5.11.0/swagger-ui.css" />
</head>
<body>
<div id="swagger-ui"></div>
<script src="https://unpkg.com/swagger-ui-dist@5.11.0/swagger-ui-bundle.js" crossorigin></script>
<script>window.onload = () => {
	window.ui = SwaggerUIBundle({url: '%s',dom_id: '#swagger-ui'});
};</script>
</body>
</html>`, specURL)
}

func SendRequest[T any](method, url string, headers map[string]string, requestBody io.Reader) (T, error) {
	var t T
	req, err := http.NewRequest(method, url, requestBody)
	if err != nil {
		return t, err
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return t, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return t, fmt.Errorf("request failed with status code %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return t, err
	}

	if err := json.Unmarshal(body, &t); err != nil {
		return t, err
	}

	return t, nil
}

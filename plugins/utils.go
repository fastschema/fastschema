package plugins

import (
	"regexp"
	"strings"
)

func ExtractInlineJsFunc(handlerContent, handlerName string) string {
	newContent := ExtractInlineJsNormalFunc(handlerContent, handlerName)
	newContent = ExtractInlineJsArrowFunc(newContent, handlerName)
	return newContent
}

func ExtractInlineJsNormalFunc(handlerContent, handlerName string) string {
	handlerLines := strings.Split(handlerContent, "\n")

	// Regex to match JavaScript function definitions and capture the function name
	re := regexp.MustCompile(`^(async\s+)?function\s+(\w*)\s*\(`)

	// Replace all occurrences of the old function name with the new function name
	handlerLines[0] = re.ReplaceAllString(handlerLines[0], `${1}function `+handlerName+`(`)

	return strings.Join(handlerLines, "\n")
}

func ExtractInlineJsArrowFunc(handlerContent, handlerName string) string {
	handlerLines := strings.Split(handlerContent, "\n")

	// Regex to match JavaScript arrow functions
	re := regexp.MustCompile(`^(async\s+)?(\([^)]*\)|\w+)\s*=>\s*`)

	// Replace all matched arrow functions with a named function using `const`
	handlerLines[0] = re.ReplaceAllString(handlerLines[0], `const `+handlerName+` = ${1}${2} => `)

	return strings.Join(handlerLines, "\n")
}

func IsValidJSFuncName(name string) bool {
	// Regex for a valid JavaScript function name
	re := regexp.MustCompile(`^[a-zA-Z_$][a-zA-Z0-9_$]*$`)

	// Test if the name matches the regex
	return re.MatchString(name)
}

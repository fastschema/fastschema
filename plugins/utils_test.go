package plugins_test

import (
	"testing"

	"github.com/fastschema/fastschema/plugins"
	"github.com/stretchr/testify/assert"
)

func TestIsValidJSFuncName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"ValidName", "validFunctionName", true},
		{"ValidNameWithUnderscore", "_validFunctionName", true},
		{"ValidNameWithDollar", "$validFunctionName", true},
		{"InvalidNameStartingWithNumber", "1invalidFunctionName", false},
		{"InvalidNameWithSpace", "invalid FunctionName", false},
		{"InvalidNameWithSpecialChar", "invalid@FunctionName", false},
		{"EmptyName", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := plugins.IsValidJSFuncName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
func TestExtractInlineJsArrowFunc(t *testing.T) {
	tests := []struct {
		name           string
		handlerContent string
		handlerName    string
		expected       string
	}{
		{
			"ArrowOneLineFunction",
			"() => {}",
			"newName",
			"const newName = () => {}",
		},
		{
			"ArrowMultiLineFunction",
			`() => {
				console.log("Hello World");
			}`,
			"newName",
			`const newName = () => {
				console.log("Hello World");
			}`,
		},
		{
			"ArrowFunctionWithParams",
			`(param1, param2) => {
				console.log("Hello World");
			}`,
			"newName",
			`const newName = (param1, param2) => {
				console.log("Hello World");
			}`,
		},
		{
			"ArrowFunctionWithAsync",
			`async (param1, param2) => {
				console.log("Hello World");
			}`,
			"newName",
			`const newName = async (param1, param2) => {
				console.log("Hello World");
			}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := plugins.ExtractInlineJsArrowFunc(tt.handlerContent, tt.handlerName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractInlineJsNormalFunc(t *testing.T) {
	tests := []struct {
		name           string
		handlerContent string
		handlerName    string
		expected       string
	}{
		{
			"NormalFunction",
			"function oldName() {}",
			"newName",
			"function newName() {}",
		},
		{
			"NormalFunctionWithParams",
			"function oldName(param1, param2) {}",
			"newName",
			"function newName(param1, param2) {}",
		},
		{
			"AsyncFunction",
			"async function oldName() {}",
			"newName",
			"async function newName() {}",
		},
		{
			"AsyncFunctionWithParams",
			"async function oldName(param1, param2) {}",
			"newName",
			"async function newName(param1, param2) {}",
		},
		{
			"FunctionWithDifferentName",
			"function anotherName() {}",
			"newName",
			"function newName() {}",
		},
		{
			"FunctionWithNoName",
			"function () {}",
			"newName",
			"function newName() {}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := plugins.ExtractInlineJsNormalFunc(tt.handlerContent, tt.handlerName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractInlineJsFunc(t *testing.T) {
	tests := []struct {
		name           string
		handlerContent string
		handlerName    string
		expected       string
	}{
		{
			"AsyncArrowFunction",
			`async (param1, param2) => {
				console.log("Hello World");
			}`,
			"newName",
			`const newName = async (param1, param2) => {
				console.log("Hello World");
			}`,
		},
		{
			"AsyncNormalFunction",
			`async function oldName(param1, param2) {
				console.log("Hello World");
			}`,
			"newName",
			`async function newName(param1, param2) {
				console.log("Hello World");
			}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := plugins.ExtractInlineJsFunc(tt.handlerContent, tt.handlerName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

package activityservice

import "strings"

// redactedPlaceholder replaces the value of any field deemed sensitive.
const redactedPlaceholder = "[REDACTED]"

// defaultRedactFields are case-insensitive substrings that mark a field as
// sensitive. Substring matching catches variants like "password_hash" or
// "user_api_key". Operators can override the list via config (see config phase).
var defaultRedactFields = []string{
	"password",
	"token",
	"secret",
	"api_key",
	"apikey",
	"hash",
	"salt",
	"private_key",
}

// redactor decides which field values to mask before they are persisted.
type redactor struct {
	fields []string
}

func newRedactor(fields []string) *redactor {
	if len(fields) == 0 {
		fields = defaultRedactFields
	}

	return &redactor{fields: fields}
}

// shouldRedact reports whether a field name matches the sensitive list.
func (r *redactor) shouldRedact(key string) bool {
	k := strings.ToLower(key)
	for _, f := range r.fields {
		if f != "" && strings.Contains(k, f) {
			return true
		}
	}

	return false
}

// redactValue returns the placeholder for sensitive keys, the value otherwise.
func (r *redactor) redactValue(key string, value any) any {
	if r.shouldRedact(key) {
		return redactedPlaceholder
	}

	return value
}

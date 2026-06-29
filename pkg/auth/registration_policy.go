package auth

import (
	"context"
	"strings"

	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/errors"
	"golang.org/x/net/idna"
)

// BuiltinPolicyValidator turns a RegistrationPolicy into a PreUserRegister hook.
// It runs as chain entry [0] (registered in createServices) so custom hooks see
// the already-normalized email. Order: normalize -> allowlist -> blocklist ->
// reserved username. Anything more specific (disposable/free-webmail/...) is
// expected to be a custom OnPreUserRegister hook.
func BuiltinPolicyValidator(p *fs.RegistrationPolicy) fs.PreUserRegisterHook {
	return func(_ context.Context, in *fs.RegistrationInput) error {
		if p == nil {
			return nil
		}

		if p.NormalizeEmail {
			in.Email = NormalizeEmail(in.Email)
		}

		domain := domainOf(in.Email)

		if len(p.AllowedEmailDomains) > 0 && !containsFold(p.AllowedEmailDomains, domain) {
			return errors.BadRequest(MSG_EMAIL_DOMAIN_NOT_ALLOWED)
		}

		if containsFold(p.BlockedEmailDomains, domain) {
			return errors.BadRequest(MSG_EMAIL_DOMAIN_NOT_ALLOWED)
		}

		if in.Username != "" && containsFold(p.ReservedUsernames, in.Username) {
			return errors.BadRequest(MSG_USERNAME_NOT_AVAILABLE)
		}

		return nil
	}
}

// NormalizeEmail lowercases and punycode-encodes the domain part while
// preserving the local-part casing (per OWASP guidance). Applied consistently
// at registration and login so stored/queried values match. Returns the input
// trimmed if it has no domain part.
func NormalizeEmail(email string) string {
	email = strings.TrimSpace(email)
	at := strings.LastIndex(email, "@")
	if at < 0 {
		return email
	}
	local, domain := email[:at], normalizeDomain(email[at+1:])
	return local + "@" + domain
}

// normalizeDomain lowercases the domain and converts IDN to punycode (ASCII).
// On conversion error it falls back to the lowercased domain.
func normalizeDomain(domain string) string {
	domain = strings.ToLower(strings.TrimSpace(domain))
	if ascii, err := idna.ToASCII(domain); err == nil {
		return ascii
	}
	return domain
}

// domainOf returns the lowercased domain part of an email (empty if none).
func domainOf(email string) string {
	at := strings.LastIndex(email, "@")
	if at < 0 {
		return ""
	}
	return strings.ToLower(strings.TrimSpace(email[at+1:]))
}

// containsFold reports whether list contains target, case-insensitively.
func containsFold(list []string, target string) bool {
	for _, item := range list {
		if strings.EqualFold(strings.TrimSpace(item), target) {
			return true
		}
	}
	return false
}

package authservice

import (
	"net/url"
	"strings"

	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/errors"
)

// loopbackHosts are the only hosts treated as a native-app loopback target.
// RFC 8252 section 7.3: the loopback port is variable, so any port is accepted.
var loopbackHosts = map[string]bool{
	"127.0.0.1": true,
	"::1":       true,
	"localhost": true,
}

// cliLoginConfig returns the effective CLI-login config (never nil-derefs).
func (as *AuthService) cliLoginConfig() *fs.CLILoginConfig {
	cfg := as.AppConfig()
	if cfg == nil || cfg.AuthConfig == nil {
		return nil
	}
	return cfg.AuthConfig.CLILogin
}

// cliLoginEnabled reports whether the CLI / native-app login feature is on.
// Nil config (the default) means disabled.
func cliLoginEnabled(cfg *fs.CLILoginConfig) bool {
	return cfg != nil && cfg.Enabled
}

// isLoopbackHost reports whether host is a loopback address per RFC 8252.
func isLoopbackHost(host string) bool {
	return loopbackHosts[strings.ToLower(host)]
}

// validateRedirectTarget enforces the redirect-target rules before any code is
// ever issued, so the endpoint can never be turned into an open redirector:
//   - loopback host (127.0.0.1 / ::1 / localhost): any port, scheme http|https.
//   - non-loopback host: scheme MUST be https AND the host must match an entry
//     of allowedHosts exactly (case-insensitive, no wildcard).
//
// Anything else is an error; the caller returns 400 and never redirects.
func validateRedirectTarget(raw string, allowedHosts []string) error {
	if strings.TrimSpace(raw) == "" {
		return errors.BadRequest("redirect_uri is required")
	}

	u, err := url.Parse(raw)
	if err != nil {
		return errors.BadRequest("redirect_uri is not a valid URL")
	}

	host := u.Hostname()
	if host == "" {
		return errors.BadRequest("redirect_uri must include a host")
	}

	scheme := strings.ToLower(u.Scheme)

	if isLoopbackHost(host) {
		if scheme != "http" && scheme != "https" {
			return errors.BadRequest("loopback redirect_uri must use http or https")
		}
		return nil
	}

	if scheme != "https" {
		return errors.BadRequest("non-loopback redirect_uri must use https")
	}

	for _, allowed := range allowedHosts {
		if strings.EqualFold(strings.TrimSpace(allowed), host) {
			return nil
		}
	}

	return errors.BadRequest("redirect_uri host is not allowed")
}

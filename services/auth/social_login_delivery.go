package authservice

import (
	"net/url"
	"strings"
	"time"

	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/errors"
)

// deliverSocialLogin routes the minted credential per carrier mode. The token is
// never returned to the browser for web/cli; both mint a one-time code that the
// recipient exchanges server-to-server (the dash for web, the loopback listener
// for cli), keeping token storage on the client side.
//   - cli: 302 to the loopback carrying the code.
//   - web: 302 to the dash login route with the code in the URL fragment.
//   - legacy/absent: return the JWT as JSON.
func (as *AuthService) deliverSocialLogin(c fs.Context, carrier *authCarrier, tokens *fs.JWTTokens) (*fs.JWTTokens, error) {
	switch carrier.Mode {
	case carrierModeCLI:
		cfg := as.cliLoginConfig()
		if !cliLoginEnabled(cfg) {
			return nil, errors.Forbidden("cli login is disabled")
		}
		if err := validateRedirectTarget(carrier.RedirectURI, cfg.AllowedRedirectHosts); err != nil {
			return nil, err
		}
		code := as.otcStore.mint(tokens, carrier.CodeChallenge, time.Now())
		return nil, c.Redirect(buildLoopbackRedirect(carrier, code))

	case carrierModeWeb:
		code := as.otcStore.mint(tokens, carrier.CodeChallenge, time.Now())
		return nil, c.Redirect(as.dashLoginURL() + "#code=" + url.QueryEscape(code))

	default:
		return tokens, nil
	}
}

// dashLoginURL is the dash login route the web social callback redirects to with
// the one-time code; the dash exchanges it and stores the token via the SDK.
func (as *AuthService) dashLoginURL() string {
	cfg := as.AppConfig()
	if cfg != nil && cfg.DashURL != "" {
		return strings.TrimRight(cfg.DashURL, "/") + "/login"
	}
	return "/dash/login"
}

// mintOTCRedirect stashes the credential under a one-time code and returns the
// loopback redirect that carries the code back to the CLI listener.
func (as *AuthService) mintOTCRedirect(carrier *authCarrier, tokens *fs.JWTTokens) *cliRedirectResponse {
	code := as.otcStore.mint(tokens, carrier.CodeChallenge, time.Now())
	return &cliRedirectResponse{Redirect: buildLoopbackRedirect(carrier, code)}
}

// buildLoopbackRedirect appends the one-time code and the echoed correlation to
// the validated loopback redirect target.
func buildLoopbackRedirect(carrier *authCarrier, code string) string {
	sep := "?"
	if strings.Contains(carrier.RedirectURI, "?") {
		sep = "&"
	}
	redirect := carrier.RedirectURI + sep + "code=" + url.QueryEscape(code)
	if carrier.Correlation != "" {
		redirect += "&state=" + url.QueryEscape(carrier.Correlation)
	}
	return redirect
}

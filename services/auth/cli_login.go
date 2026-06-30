package authservice

import (
	"net/url"
	"time"

	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/auth"
	"github.com/fastschema/fastschema/pkg/errors"
)

// otcTTL bounds how long a minted one-time code stays redeemable. Short enough
// to make interception-then-replay impractical, long enough for the browser to
// 302 to the loopback and the CLI to exchange.
const otcTTL = 60 * time.Second

type cliInitiateRequest struct {
	RedirectURI   string `json:"redirect_uri"`
	Correlation   string `json:"correlation"`
	CodeChallenge string `json:"code_challenge"`
}

type cliInitiateResponse struct {
	Carrier      string `json:"carrier"`
	AuthorizeURL string `json:"authorize_url"`
}

type cliLocalLoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
	Carrier  string `json:"carrier"`
}

type cliOTPLoginRequest struct {
	SessionID string `json:"session_id"`
	OTP       string `json:"otp"`
	Carrier   string `json:"carrier"`
}

type cliRedirectResponse struct {
	Redirect string `json:"redirect"`
}

// CLIInitiate validates the requested loopback target and returns a signed
// carrier plus the dash authorize URL. No credential or code is issued here;
// fail-fast on a disabled feature or a bad redirect before anything is minted.
func (as *AuthService) CLIInitiate(c fs.Context, req *cliInitiateRequest) (*cliInitiateResponse, error) {
	cfg := as.cliLoginConfig()
	if !cliLoginEnabled(cfg) {
		return nil, errors.Forbidden("cli login is disabled")
	}
	if req == nil {
		return nil, errors.BadRequest("invalid request")
	}
	// PKCE is mandatory for native apps (RFC 8252 section 8.1): loopback
	// redirects are interceptable by other local processes, so every minted code
	// must be bound to a verifier the interceptor does not hold.
	if req.CodeChallenge == "" {
		return nil, errors.BadRequest("code_challenge is required")
	}
	if err := validateRedirectTarget(req.RedirectURI, cfg.AllowedRedirectHosts); err != nil {
		return nil, err
	}

	carrier, err := buildAuthCarrier(authCarrier{
		Mode:          carrierModeCLI,
		RedirectURI:   req.RedirectURI,
		Correlation:   req.Correlation,
		CodeChallenge: req.CodeChallenge,
	}, as.AppKey(), time.Now())
	if err != nil {
		return nil, err
	}

	return &cliInitiateResponse{
		Carrier:      carrier,
		AuthorizeURL: "/dash/login?cli=" + url.QueryEscape(carrier),
	}, nil
}

// CLILocalLogin authenticates with username/password through the dash browser
// and returns a loopback redirect carrying a one-time code. The JWT is minted
// server-side and stashed in the OTC store; it never reaches the browser.
func (as *AuthService) CLILocalLogin(c fs.Context, req *cliLocalLoginRequest) (*cliRedirectResponse, error) {
	cfg := as.cliLoginConfig()
	if !cliLoginEnabled(cfg) {
		return nil, errors.Forbidden("cli login is disabled")
	}
	if req == nil {
		return nil, errors.BadRequest("invalid request")
	}

	carrier, err := as.verifyCLICarrier(req.Carrier, cfg)
	if err != nil {
		return nil, err
	}

	provider, ok := as.GetAuthProvider(auth.ProviderLocal).(*auth.LocalProvider)
	if !ok {
		return nil, errors.InternalServerError("local provider unavailable")
	}

	user, err := provider.LocalLogin(c, &auth.LoginData{Login: req.Login, Password: req.Password})
	if err != nil {
		return nil, err
	}

	tokens, err := as.GenerateJWTTokens(c, user)
	if err != nil {
		return nil, err
	}

	return as.mintOTCRedirect(carrier, tokens), nil
}

// CLIOTPLogin completes a passwordless OTP login (after /auth/otp/request) and
// returns a loopback redirect carrying a one-time code. JWT stays server-side.
func (as *AuthService) CLIOTPLogin(c fs.Context, req *cliOTPLoginRequest) (*cliRedirectResponse, error) {
	cfg := as.cliLoginConfig()
	if !cliLoginEnabled(cfg) {
		return nil, errors.Forbidden("cli login is disabled")
	}
	if req == nil {
		return nil, errors.BadRequest("invalid request")
	}

	carrier, err := as.verifyCLICarrier(req.Carrier, cfg)
	if err != nil {
		return nil, err
	}

	provider, ok := as.GetAuthProvider(auth.ProviderOTP).(*auth.OTPProvider)
	if !ok || !provider.IsEnabled() {
		return nil, errors.BadRequest("otp login is not enabled")
	}

	user, err := provider.VerifyOTP(c, &auth.OTPVerify{SessionID: req.SessionID, OTP: req.OTP})
	if err != nil {
		return nil, err
	}

	tokens, err := as.GenerateJWTTokens(c, user)
	if err != nil {
		return nil, err
	}

	return as.mintOTCRedirect(carrier, tokens), nil
}

// verifyCLICarrier parses a cli-mode carrier and re-checks the redirect
// allowlist at delivery time (defense-in-depth).
func (as *AuthService) verifyCLICarrier(raw string, cfg *fs.CLILoginConfig) (*authCarrier, error) {
	carrier, err := parseAuthCarrier(raw, as.AppKey(), time.Now())
	if err != nil {
		return nil, err
	}
	if carrier.Mode != carrierModeCLI {
		return nil, errors.BadRequest("invalid cli carrier")
	}
	if err := validateRedirectTarget(carrier.RedirectURI, cfg.AllowedRedirectHosts); err != nil {
		return nil, err
	}
	return carrier, nil
}

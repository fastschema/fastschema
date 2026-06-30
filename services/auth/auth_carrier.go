package authservice

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/errors"
	"github.com/fastschema/fastschema/pkg/utils"
)

// authCarrierTTL bounds validity of a signed auth carrier; caps replay while
// leaving time to complete the provider consent screen.
const authCarrierTTL = 5 * time.Minute

// authStateArg is the context key carrying the carrier from the service to a
// provider's Login, which relays it as the OAuth `state` value.
const authStateArg = "auth_state"

// Carrier modes select how the social callback delivers the credential.
const (
	carrierModeLegacy = ""    // callback returns the JWT as JSON
	carrierModeWeb    = "web" // mint a one-time code + 302 to the dash to exchange
	carrierModeCLI    = "cli" // mint a one-time code + 302 to the loopback
)

// authCarrier is the value passed as the OAuth `state` on every social login. It
// carries the login mode and, for cli mode, the loopback intent through the
// provider redirect. AEAD-encrypted (AES-GCM via APP_KEY): the client can
// neither read nor tamper with it. The nonce backs the per-browser state
// binding (see bindStateCookie).
type authCarrier struct {
	Nonce         string `json:"n"`            // backs the per-browser state binding
	Mode          string `json:"m,omitempty"`  // "" legacy | "web" | "cli"
	RedirectURI   string `json:"r,omitempty"`  // loopback target (cli mode only)
	Correlation   string `json:"c,omitempty"`  // opaque caller value, echoed back (cli mode)
	CodeChallenge string `json:"cc,omitempty"` // PKCE S256 challenge, base64url (cli mode)
	ExpiresAt     int64  `json:"e"`            // unix seconds
}

// buildAuthCarrier serializes and encrypts a carrier, stamping the TTL relative
// to now. A random nonce is generated when the caller leaves it empty.
func buildAuthCarrier(payload authCarrier, key string, now time.Time) (string, error) {
	if payload.Nonce == "" {
		payload.Nonce = utils.RandomString(16)
	}
	payload.ExpiresAt = now.Add(authCarrierTTL).Unix()

	raw, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	return utils.Encrypt(string(raw), key)
}

// parseAuthCarrier decrypts and validates a carrier. A decrypt/unmarshal
// failure means the value was tampered with (or signed by a different key); an
// elapsed ExpiresAt means it is stale. Both reject as Unauthorized.
func parseAuthCarrier(opaque, key string, now time.Time) (*authCarrier, error) {
	if opaque == "" {
		return nil, errors.Unauthorized("invalid auth state")
	}

	raw, err := utils.Decrypt(opaque, key)
	if err != nil {
		return nil, errors.Unauthorized("invalid auth state")
	}

	var carrier authCarrier
	if err := json.Unmarshal([]byte(raw), &carrier); err != nil {
		return nil, errors.Unauthorized("invalid auth state")
	}

	if carrier.ExpiresAt < now.Unix() {
		return nil, errors.Unauthorized("auth state expired")
	}

	return &carrier, nil
}

// injectAuthState stores the carrier on the request context for the provider's
// Login to relay as the OAuth `state` parameter.
func injectAuthState(c fs.Context, carrier string) {
	c.SetArg(authStateArg, carrier)
}

// buildLoginCarrier resolves the login mode from the request and returns both
// the encrypted carrier (the OAuth `state`) and its parsed form:
//   - ?cli=<carrier>: a pre-signed cli carrier from /cli/initiate; its integrity,
//     TTL, gating, and redirect allowlist are re-checked, then it is relayed
//     unchanged (already signed with the same key).
//   - ?web=1: a fresh web-mode carrier.
//   - neither: a legacy carrier (mode "").
func (as *AuthService) buildLoginCarrier(c fs.Context) (string, *authCarrier, error) {
	key := as.AppKey()
	now := time.Now()

	if pre := c.Arg("cli"); pre != "" {
		carrier, err := parseAuthCarrier(pre, key, now)
		if err != nil {
			return "", nil, err
		}
		if carrier.Mode != carrierModeCLI {
			return "", nil, errors.BadRequest("invalid cli carrier")
		}
		cfg := as.cliLoginConfig()
		if !cliLoginEnabled(cfg) {
			return "", nil, errors.Forbidden("cli login is disabled")
		}
		if err := validateRedirectTarget(carrier.RedirectURI, cfg.AllowedRedirectHosts); err != nil {
			return "", nil, err
		}
		return pre, carrier, nil
	}

	mode := carrierModeLegacy
	if web := strings.ToLower(c.Arg("web")); web == "1" || web == "true" {
		mode = carrierModeWeb
	}

	// Set the nonce here so the same value can be written to the binding cookie
	// (see bindStateCookie); buildAuthCarrier keeps a pre-set nonce.
	payload := authCarrier{Mode: mode, Nonce: utils.RandomString(16)}
	encrypted, err := buildAuthCarrier(payload, key, now)
	if err != nil {
		return "", nil, err
	}
	return encrypted, &payload, nil
}

// readVerifiedCarrier parses and validates the `state` from a provider callback;
// a missing, tampered, or expired state is rejected.
func (as *AuthService) readVerifiedCarrier(c fs.Context) (*authCarrier, error) {
	return parseAuthCarrier(c.Arg("state"), as.AppKey(), time.Now())
}

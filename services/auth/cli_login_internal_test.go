package authservice

import (
	"crypto/sha256"
	"encoding/base64"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/fastschema/fastschema/entity"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// validTestKey is a 32-byte AES-256 key (production APP_KEY is RandomString(32)).
const validTestKey = "0123456789abcdef0123456789abcdef"

// captureContext is a minimal fs.Context that records the side effects the cli
// delivery path produces (redirect target + cookies) so they can be asserted
// without a live HTTP server.
type captureContext struct {
	args         map[string]string
	headers      map[string]string
	reqCookies   map[string]string // incoming request cookies (get path)
	redirectedTo string
	cookies      []*fs.Cookie // cookies set on the response (set path)
}

func newCaptureContext() *captureContext {
	return &captureContext{args: map[string]string{}, headers: map[string]string{}, reqCookies: map[string]string{}}
}

func (c *captureContext) Redirect(path string) error { c.redirectedTo = path; return nil }
func (c *captureContext) Cookie(name string, values ...*fs.Cookie) string {
	if len(values) > 0 {
		c.cookies = append(c.cookies, values[0])
		return values[0].Value
	}
	return c.reqCookies[name]
}
func (c *captureContext) SetArg(k, v string) string { c.args[k] = v; return v }
func (c *captureContext) Arg(k string, d ...string) string {
	if v, ok := c.args[k]; ok {
		return v
	}
	if len(d) > 0 {
		return d[0]
	}
	return ""
}
func (c *captureContext) Header(k string, v ...string) string {
	if len(v) > 0 {
		c.headers[k] = v[0]
		return v[0]
	}
	return c.headers[k]
}
func (c *captureContext) Args() map[string]string            { return c.args }
func (c *captureContext) ArgInt(string, ...int) int          { return 0 }
func (c *captureContext) TraceID() string                    { return "test" }
func (c *captureContext) User() *fs.User                     { return nil }
func (c *captureContext) Local(string, ...any) any           { return nil }
func (c *captureContext) Logger() logger.Logger              { return logger.CreateMockLogger(false) }
func (c *captureContext) Bind(any) error                     { return nil }
func (c *captureContext) Body() ([]byte, error)              { return nil, nil }
func (c *captureContext) Payload() (*entity.Entity, error)   { return nil, nil }
func (c *captureContext) BodyParser(any) error               { return nil }
func (c *captureContext) FormValue(string, ...string) string { return "" }
func (c *captureContext) Resource() *fs.Resource             { return nil }
func (c *captureContext) AuthToken() string                  { return "" }
func (c *captureContext) Next() error                        { return nil }
func (c *captureContext) Result(...*fs.Result) *fs.Result    { return nil }
func (c *captureContext) Files() ([]*fs.File, error)         { return nil, nil }
func (c *captureContext) WSClient() fs.WSClient              { return nil }
func (c *captureContext) IP() string                         { return "127.0.0.1" }
func (c *captureContext) Deadline() (time.Time, bool)        { return time.Time{}, false }
func (c *captureContext) Done() <-chan struct{}              { return nil }
func (c *captureContext) Err() error                         { return nil }
func (c *captureContext) Value(any) any                      { return nil }

// newTestService builds an AuthService whose only configured surface is the
// app key and the CLI-login config, enough to exercise gating/initiate/exchange/
// delivery directly.
func newTestService(cli *fs.CLILoginConfig) *AuthService {
	return &AuthService{
		AppKey: func() string { return validTestKey },
		AppConfig: func() *fs.Config {
			return &fs.Config{AppKey: validTestKey, DashURL: "http://localhost:8000/dash", AuthConfig: &fs.AuthConfig{CLILogin: cli}}
		},
		otcStore: newOTCStore(),
	}
}

func enabledCLI(hosts ...string) *fs.CLILoginConfig {
	return &fs.CLILoginConfig{Enabled: true, AllowedRedirectHosts: hosts}
}

func TestCLIRedirectValidator(t *testing.T) {
	allowed := []string{"app.example.com", "Allowed.Com"}
	cases := []struct {
		name string
		raw  string
		ok   bool
	}{
		{"loopback ipv4 any port", "http://127.0.0.1:54321/cb", true},
		{"loopback ipv6 any port", "http://[::1]:9/x", true},
		{"loopback localhost", "http://localhost:1/", true},
		{"loopback https ok", "https://127.0.0.1:8443/cb", true},
		{"allowlisted https", "https://app.example.com/cb", true},
		{"allowlisted case-insensitive", "https://allowed.com/cb", true},
		{"allowlisted https with port", "https://app.example.com:8443/cb", true},
		{"foreign http", "http://evil.com", false},
		{"foreign https not allowlisted", "https://evil.com", false},
		{"non-loopback non-https", "http://app.example.com", false},
		{"non-loopback scheme ftp", "ftp://127.0.0.1", false},
		{"subdomain not wildcard", "https://sub.app.example.com/cb", false},
		{"empty", "", false},
		{"no host", "/just/a/path", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateRedirectTarget(tc.raw, allowed)
			if tc.ok {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestAuthCarrierRoundTrip(t *testing.T) {
	now := time.Now()
	in := authCarrier{Mode: carrierModeCLI, RedirectURI: "http://127.0.0.1:5000/cb", Correlation: "abc", CodeChallenge: "chal"}
	raw, err := buildAuthCarrier(in, validTestKey, now)
	require.NoError(t, err)

	out, err := parseAuthCarrier(raw, validTestKey, now)
	require.NoError(t, err)
	assert.Equal(t, carrierModeCLI, out.Mode)
	assert.Equal(t, "http://127.0.0.1:5000/cb", out.RedirectURI)
	assert.Equal(t, "abc", out.Correlation)
	assert.Equal(t, "chal", out.CodeChallenge)
	assert.NotEmpty(t, out.Nonce)
}

func TestAuthCarrierTampered(t *testing.T) {
	now := time.Now()
	raw, err := buildAuthCarrier(authCarrier{Mode: carrierModeWeb}, validTestKey, now)
	require.NoError(t, err)

	// Drop the final hex byte pair to corrupt the AEAD ciphertext/tag.
	tampered := raw[:len(raw)-2]
	_, err = parseAuthCarrier(tampered, validTestKey, now)
	assert.Error(t, err)

	// A carrier signed with a different key must not verify.
	_, err = parseAuthCarrier(raw, "ffffffffffffffffffffffffffffffff", now)
	assert.Error(t, err)
}

func TestAuthCarrierExpired(t *testing.T) {
	built := time.Now()
	raw, err := buildAuthCarrier(authCarrier{Mode: carrierModeLegacy}, validTestKey, built)
	require.NoError(t, err)

	// Parse well past the TTL.
	_, err = parseAuthCarrier(raw, validTestKey, built.Add(authCarrierTTL+time.Second))
	assert.Error(t, err)
}

func TestCLICodeStoreSingleUseAndTTL(t *testing.T) {
	store := newOTCStore()
	now := time.Now()
	tokens := &fs.JWTTokens{AccessToken: "jwt"}

	code := store.mint(tokens, "", now)
	require.NotEmpty(t, code)

	got, ok := store.take(code, now)
	require.True(t, ok)
	assert.Equal(t, "jwt", got.tokens.AccessToken)

	// Second take fails (single-use).
	_, ok = store.take(code, now)
	assert.False(t, ok)

	// Expired entry is a miss.
	expiredCode := store.mint(tokens, "", now)
	_, ok = store.take(expiredCode, now.Add(2*time.Minute))
	assert.False(t, ok)
}

func TestVerifyPKCES256(t *testing.T) {
	verifier := "this-is-a-code-verifier-string"
	sum := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(sum[:])

	assert.True(t, verifyPKCES256(challenge, verifier))
	assert.False(t, verifyPKCES256(challenge, "wrong-verifier"))
	assert.False(t, verifyPKCES256(challenge, ""))
}

func TestCLILoginDisabledRejected(t *testing.T) {
	as := newTestService(nil) // CLILogin nil => disabled

	_, err := as.CLIInitiate(newCaptureContext(), &cliInitiateRequest{RedirectURI: "http://127.0.0.1:5000/cb"})
	assert.Error(t, err)

	_, err = as.CLILocalLogin(newCaptureContext(), &cliLocalLoginRequest{})
	assert.Error(t, err)

	_, err = as.CLIOTPLogin(newCaptureContext(), &cliOTPLoginRequest{})
	assert.Error(t, err)
	// Exchange is intentionally ungated (the gate is at mint time); an unknown
	// code still fails as invalid.
	_, err = as.Exchange(newCaptureContext(), &exchangeRequest{Code: "x"})
	assert.Error(t, err)
}

func TestCLIInitiate(t *testing.T) {
	as := newTestService(enabledCLI("app.example.com"))

	// Missing PKCE challenge: rejected (RFC 8252 requires PKCE for native apps).
	_, err := as.CLIInitiate(newCaptureContext(), &cliInitiateRequest{RedirectURI: "http://127.0.0.1:5000/cb"})
	assert.Error(t, err)

	// Bad redirect: rejected before any carrier is issued.
	_, err = as.CLIInitiate(newCaptureContext(), &cliInitiateRequest{RedirectURI: "https://evil.com/cb", CodeChallenge: "chal"})
	assert.Error(t, err)

	// Good loopback redirect: returns a carrier + dash authorize URL.
	resp, err := as.CLIInitiate(newCaptureContext(), &cliInitiateRequest{
		RedirectURI:   "http://127.0.0.1:5000/cb",
		Correlation:   "corr-1",
		CodeChallenge: "chal-1",
	})
	require.NoError(t, err)
	assert.NotEmpty(t, resp.Carrier)
	assert.True(t, strings.HasPrefix(resp.AuthorizeURL, "/dash/login?cli="))

	// The carrier round-trips back into a cli-mode carrier.
	carrier, err := parseAuthCarrier(resp.Carrier, validTestKey, time.Now())
	require.NoError(t, err)
	assert.Equal(t, carrierModeCLI, carrier.Mode)
	assert.Equal(t, "corr-1", carrier.Correlation)
}

func TestVerifyStateBinding(t *testing.T) {
	// Legacy mode is not browser-bound: no cookie required.
	assert.NoError(t, verifyStateBinding(newCaptureContext(), &authCarrier{Mode: carrierModeLegacy, Nonce: "n1"}))

	for _, mode := range []string{carrierModeWeb, carrierModeCLI} {
		carrier := &authCarrier{Mode: mode, Nonce: "secret-nonce"}

		// Missing cookie -> rejected (login CSRF / session fixation defense).
		assert.Error(t, verifyStateBinding(newCaptureContext(), carrier), mode)

		// Mismatched cookie -> rejected.
		bad := newCaptureContext()
		bad.reqCookies[stateCookieName] = "other"
		assert.Error(t, verifyStateBinding(bad, carrier), mode)

		// Matching cookie -> accepted.
		good := newCaptureContext()
		good.reqCookies[stateCookieName] = "secret-nonce"
		assert.NoError(t, verifyStateBinding(good, carrier), mode)
	}
}

func TestBindStateCookie(t *testing.T) {
	as := newTestService(nil)

	// Legacy mode sets no binding cookie.
	legacy := newCaptureContext()
	as.bindStateCookie(legacy, &authCarrier{Mode: carrierModeLegacy, Nonce: "n"})
	assert.Empty(t, legacy.cookies)

	// Web mode sets an HttpOnly binding cookie carrying the nonce.
	web := newCaptureContext()
	as.bindStateCookie(web, &authCarrier{Mode: carrierModeWeb, Nonce: "n-web"})
	require.Len(t, web.cookies, 1)
	assert.Equal(t, stateCookieName, web.cookies[0].Name)
	assert.Equal(t, "n-web", web.cookies[0].Value)
	assert.True(t, web.cookies[0].HTTPOnly)
}

func TestCLIExchangeSingleUseAndPKCE(t *testing.T) {
	as := newTestService(enabledCLI("app.example.com"))
	verifier := "verifier-abc-123"
	sum := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(sum[:])

	code := as.otcStore.mint(&fs.JWTTokens{AccessToken: "jwt-1"}, challenge, time.Now())

	// Wrong verifier rejected.
	_, err := as.Exchange(newCaptureContext(), &exchangeRequest{Code: code, CodeVerifier: "nope"})
	assert.Error(t, err)

	// Correct verifier rejected too, because the failed attempt already consumed
	// the single-use code (atomic take happens before PKCE check).
	_, err = as.Exchange(newCaptureContext(), &exchangeRequest{Code: code, CodeVerifier: verifier})
	assert.Error(t, err)

	// Fresh code: correct verifier returns the credential exactly once.
	code2 := as.otcStore.mint(&fs.JWTTokens{AccessToken: "jwt-2"}, challenge, time.Now())
	tokens, err := as.Exchange(newCaptureContext(), &exchangeRequest{Code: code2, CodeVerifier: verifier})
	require.NoError(t, err)
	assert.Equal(t, "jwt-2", tokens.AccessToken)

	// Second exchange of the same code fails.
	_, err = as.Exchange(newCaptureContext(), &exchangeRequest{Code: code2, CodeVerifier: verifier})
	assert.Error(t, err)
}

func TestDeliverSocialLoginCLIMode(t *testing.T) {
	as := newTestService(enabledCLI("app.example.com"))
	carrier := &authCarrier{Mode: carrierModeCLI, RedirectURI: "http://127.0.0.1:5000/cb", Correlation: "corr-x"}
	c := newCaptureContext()

	tokens, err := as.deliverSocialLogin(c, carrier, &fs.JWTTokens{AccessToken: "secret-jwt"})
	require.NoError(t, err)
	assert.Nil(t, tokens) // token never returned to the browser in cli mode

	parsed, err := url.Parse(c.redirectedTo)
	require.NoError(t, err)
	assert.Equal(t, "127.0.0.1:5000", parsed.Host)
	assert.NotEmpty(t, parsed.Query().Get("code"))
	assert.Equal(t, "corr-x", parsed.Query().Get("state"))
	// The redirect carries only the one-time code, never the JWT.
	assert.NotContains(t, c.redirectedTo, "secret-jwt")
	assert.Empty(t, c.cookies)
}

func TestDeliverSocialLoginWebMode(t *testing.T) {
	as := newTestService(nil) // web mode is independent of the cli flag
	c := newCaptureContext()

	tokens, err := as.deliverSocialLogin(c, &authCarrier{Mode: carrierModeWeb}, &fs.JWTTokens{AccessToken: "access-jwt"})
	require.NoError(t, err)
	assert.Nil(t, tokens)      // token never returned to the browser
	assert.Empty(t, c.cookies) // no token cookie set server-side

	// Redirect to the dash login route with the one-time code in the fragment.
	assert.True(t, strings.HasPrefix(c.redirectedTo, "http://localhost:8000/dash/login#code="))
	assert.NotContains(t, c.redirectedTo, "access-jwt")

	// The code redeems exactly once for the credential.
	code := c.redirectedTo[strings.Index(c.redirectedTo, "#code=")+len("#code="):]
	got, ok := as.otcStore.take(code, time.Now())
	require.True(t, ok)
	assert.Equal(t, "access-jwt", got.tokens.AccessToken)
}

func TestDeliverSocialLoginLegacy(t *testing.T) {
	as := newTestService(nil)
	c := newCaptureContext()

	tokens, err := as.deliverSocialLogin(c, &authCarrier{Mode: carrierModeLegacy}, &fs.JWTTokens{AccessToken: "jwt"})
	require.NoError(t, err)
	require.NotNil(t, tokens)
	assert.Equal(t, "jwt", tokens.AccessToken) // legacy returns the JWT as JSON
	assert.Empty(t, c.redirectedTo)
	assert.Empty(t, c.cookies)
}

func TestAuthMethodsWhitelistedDTO(t *testing.T) {
	as := &AuthService{
		AppConfig: func() *fs.Config {
			return &fs.Config{AuthConfig: &fs.AuthConfig{
				EnabledProviders: []string{"local", "github", "google"},
				OTP:              &fs.OTPConfig{Enabled: true},
			}}
		},
	}

	dto, err := as.AuthMethods(newCaptureContext(), nil)
	require.NoError(t, err)
	assert.Equal(t, []string{"local", "github", "google"}, dto.Providers)
	assert.True(t, dto.OTP)

	// OTP disabled => false.
	as.AppConfig = func() *fs.Config {
		return &fs.Config{AuthConfig: &fs.AuthConfig{EnabledProviders: []string{"local"}}}
	}
	dto, err = as.AuthMethods(newCaptureContext(), nil)
	require.NoError(t, err)
	assert.False(t, dto.OTP)
}

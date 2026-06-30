package authservice

import (
	"crypto/subtle"
	"strings"
	"time"

	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/errors"
)

// stateCookieName is the cookie holding the carrier nonce; the callback matches
// it to bind the flow to the browser that began it. A signed carrier alone is
// integrity-checked but replayable across browsers, so the cookie is what
// prevents another browser's carrier from completing the callback.
const stateCookieName = "fs_auth_state"

// stateBoundMode reports whether a mode delivers its outcome into the browser
// (web cookie or cli loopback) and so must be bound to it. Legacy returns JSON
// and needs no binding.
func stateBoundMode(mode string) bool {
	return mode == carrierModeWeb || mode == carrierModeCLI
}

// bindStateCookie sets the binding cookie for state-bound modes. SameSite=Lax so
// the top-level provider callback navigation still sends it.
func (as *AuthService) bindStateCookie(c fs.Context, carrier *authCarrier) {
	if carrier == nil || !stateBoundMode(carrier.Mode) {
		return
	}
	c.Cookie(stateCookieName, &fs.Cookie{
		Name:     stateCookieName,
		Value:    carrier.Nonce,
		Path:     "/",
		Expires:  time.Now().Add(authCarrierTTL),
		Secure:   strings.EqualFold(c.Header("X-Forwarded-Proto"), "https"),
		HTTPOnly: true,
		SameSite: "Lax",
	})
}

// verifyStateBinding rejects a state-bound callback whose binding cookie is
// missing or does not match the carrier nonce. The cookie is not cleared here;
// it lapses by its short TTL, so the callback issues no Set-Cookie.
func verifyStateBinding(c fs.Context, carrier *authCarrier) error {
	if !stateBoundMode(carrier.Mode) {
		return nil
	}
	got := c.Cookie(stateCookieName)
	if got == "" || subtle.ConstantTimeCompare([]byte(got), []byte(carrier.Nonce)) != 1 {
		return errors.Unauthorized("invalid auth state")
	}
	return nil
}

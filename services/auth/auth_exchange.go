package authservice

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"time"

	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/errors"
)

type exchangeRequest struct {
	Code         string `json:"code"`
	CodeVerifier string `json:"code_verifier"`
}

// Exchange redeems a one-time code for the credential. It serves both the cli
// loopback (server-to-server, with a PKCE verifier) and the web social callback
// (the dash exchanges the code the callback put in its URL fragment). Ungated:
// codes only exist once minted, so the feature gate sits at mint time. The code
// is atomically consumed (any attempt, valid or not, spends it); a registered
// PKCE challenge must match.
func (as *AuthService) Exchange(c fs.Context, req *exchangeRequest) (*fs.JWTTokens, error) {
	if req == nil || req.Code == "" {
		return nil, errors.BadRequest("code is required")
	}

	entry, ok := as.otcStore.take(req.Code, time.Now())
	if !ok {
		return nil, errors.Unauthorized("invalid or expired code")
	}

	if entry.codeChallenge != "" && !verifyPKCES256(entry.codeChallenge, req.CodeVerifier) {
		return nil, errors.Unauthorized("invalid code verifier")
	}

	return entry.tokens, nil
}

// verifyPKCES256 reports whether base64url(sha256(verifier)) equals the stored
// challenge, using a constant-time compare. Only S256 is supported (no plain).
func verifyPKCES256(challenge, verifier string) bool {
	if verifier == "" {
		return false
	}
	sum := sha256.Sum256([]byte(verifier))
	computed := base64.RawURLEncoding.EncodeToString(sum[:])
	return subtle.ConstantTimeCompare([]byte(computed), []byte(challenge)) == 1
}

package authservice

import (
	"sync"
	"time"

	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/utils"
)

// otcEntry holds a minted credential awaiting server-to-server exchange.
type otcEntry struct {
	tokens        *fs.JWTTokens
	codeChallenge string // PKCE S256 challenge (base64url); empty for web (no PKCE)
	expiresAt     time.Time
}

// otcStore is an in-process store of one-time codes (OTC) for the CLI login
// flow. Codes are single-use (atomic take) and short-lived. Single-instance is
// acceptable here: the OTC window is tiny (~60s) and the credential never leaves
// this process except through the authenticated exchange. A shared/multi-node
// store is a documented follow-up.
type otcStore struct {
	m sync.Map // key: OTC string -> *otcEntry
}

func newOTCStore() *otcStore {
	return &otcStore{}
}

// mint stores the credential under a fresh high-entropy code and returns the
// code. The code is ~190 bits of entropy (RandomString(32) over a 52-char
// alphabet), far above the 128-bit floor. Entries live for otcTTL.
func (s *otcStore) mint(tokens *fs.JWTTokens, codeChallenge string, now time.Time) string {
	code := utils.RandomString(32)
	s.m.Store(code, &otcEntry{
		tokens:        tokens,
		codeChallenge: codeChallenge,
		expiresAt:     now.Add(otcTTL),
	})
	return code
}

// take atomically removes and returns the entry for code. An expired entry is
// treated as a miss. The atomic LoadAndDelete guarantees a code is redeemable at
// most once even under concurrent exchange attempts (no double-spend).
func (s *otcStore) take(code string, now time.Time) (*otcEntry, bool) {
	v, ok := s.m.LoadAndDelete(code)
	if !ok {
		return nil, false
	}

	entry, ok := v.(*otcEntry)
	if !ok || now.After(entry.expiresAt) {
		return nil, false
	}

	return entry, true
}

package auth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Internal test for the encrypted email-change token (unexported helpers).
func TestEmailChangeToken_RoundTrip(t *testing.T) {
	key := "0123456789abcdef0123456789abcdef" // 32 bytes
	exp := time.Now().Add(time.Hour)

	token, err := encodeEmailChangeToken("sid-123", "new@example.com", exp, key)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	data, err := decodeEmailChangeToken(token, key)
	require.NoError(t, err)
	assert.Equal(t, "sid-123", data.SID)
	assert.Equal(t, "new@example.com", data.Email)
	assert.Equal(t, exp.UnixMicro(), data.Exp)
}

func TestEmailChangeToken_TamperedOrWrongKey(t *testing.T) {
	key := "0123456789abcdef0123456789abcdef"
	token, err := encodeEmailChangeToken("sid", "a@b.com", time.Now().Add(time.Hour), key)
	require.NoError(t, err)

	// Wrong key fails to decrypt.
	_, err = decodeEmailChangeToken(token, "ffffffffffffffffffffffffffffffff")
	require.Error(t, err)

	// Tampered ciphertext fails.
	_, err = decodeEmailChangeToken(token+"xx", key)
	require.Error(t, err)

	// Garbage fails.
	_, err = decodeEmailChangeToken("not-a-token", key)
	require.Error(t, err)
}

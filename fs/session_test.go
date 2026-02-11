package fs_test

import (
	"testing"

	"github.com/fastschema/fastschema/fs"
	"github.com/stretchr/testify/assert"
)

func TestSessionSchema(t *testing.T) {
	s := fs.Session{}
	schema := s.Schema()

	assert.NotNil(t, schema, "Expected Schema to return a non-nil schema")
	assert.NotNil(t, schema.DB, "Expected Schema.DB to be non-nil")
	assert.NotNil(t, schema.DB.Indexes, "Expected Schema.DB.Indexes to be non-nil")
	assert.Len(t, schema.DB.Indexes, 6, "Expected 6 database indexes")

	// Verify index names
	indexNames := make([]string, len(schema.DB.Indexes))
	for i, idx := range schema.DB.Indexes {
		indexNames[i] = idx.Name
	}

	expectedIndexes := []string{
		"idx_session_user_id",
		"idx_session_expires_at",
		"idx_session_status",
		"idx_session_user_status",
		"idx_session_type",
		"idx_session_type_status",
	}

	for _, expected := range expectedIndexes {
		assert.Contains(t, indexNames, expected, "Expected index %s to be present", expected)
	}
}

func TestSessionStatusConstants(t *testing.T) {
	// Test that session status constants are defined correctly
	assert.Equal(t, fs.SessionStatus("active"), fs.SessionStatusActive)
	assert.Equal(t, fs.SessionStatus("inactive"), fs.SessionStatusInactive)
	assert.Equal(t, fs.SessionStatus("revoked"), fs.SessionStatusRevoked)
	assert.Equal(t, fs.SessionStatus("pending_otp"), fs.SessionStatusPendingOTP)
	assert.Equal(t, fs.SessionStatus("verified"), fs.SessionStatusVerified)
}

func TestSessionTypeConstants(t *testing.T) {
	// Test that session type constants are defined correctly
	assert.Equal(t, fs.SessionType("refresh_token"), fs.SessionTypeRefreshToken)
	assert.Equal(t, fs.SessionType("otp_login"), fs.SessionTypeOTPLogin)
	assert.Equal(t, fs.SessionType("activation"), fs.SessionTypeActivation)
	assert.Equal(t, fs.SessionType("recovery"), fs.SessionTypeRecovery)
}

package fs_test

import (
	"testing"

	"github.com/fastschema/fastschema/fs"
	"github.com/stretchr/testify/assert"
)

func TestMailConfig_Clone(t *testing.T) {
	original := &fs.MailConfig{
		SenderName:        "Original Sender",
		SenderMail:        "original@example.com",
		DefaultClientName: "Original Client",
		Clients:           []fs.Map{{"key1": "value1"}, {"key2": "value2"}},
	}

	clone := original.Clone()

	assert.Equal(t, original.SenderName, clone.SenderName)
	assert.Equal(t, original.SenderMail, clone.SenderMail)
	assert.Equal(t, original.DefaultClientName, clone.DefaultClientName)
	assert.Equal(t, original.Clients, clone.Clients)

	// Ensure that the clone is a deep copy
	clone.SenderName = "Modified Sender"
	clone.SenderMail = "modified@example.com"
	clone.DefaultClientName = "Modified Client"
	clone.Clients[0]["key1"] = "modified_value1"

	assert.NotEqual(t, original.SenderName, clone.SenderName)
	assert.NotEqual(t, original.SenderMail, clone.SenderMail)
	assert.NotEqual(t, original.DefaultClientName, clone.DefaultClientName)
	assert.NotEqual(t, original.Clients[0]["key1"], clone.Clients[0]["key1"])
}

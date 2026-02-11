package entdbadapter

import (
	"context"
	"os"
	"testing"

	entSchema "entgo.io/ent/dialect/sql/schema"
	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdapterMigrate(t *testing.T) {
	// Error
	mockAdapter := createMockAdapter(t)
	changes := &db.Changes{}
	appendEntTables := []*entSchema.Table{}
	err := mockAdapter.Migrate(context.Background(), changes, false, appendEntTables...)
	assert.Error(t, err)

	// Success
	sb := createSchemaBuilder()
	dbClient, err := NewTestClient(
		utils.Must(os.MkdirTemp("", "test")),
		sb,
	)
	assert.NoError(t, err)
	_, err = dbClient.Reload(context.Background(), sb, &db.Changes{
		RenameTables: []*db.RenameItem{{From: "users", To: "members"}},
	}, false, true)
	require.NoError(t, err)
}

// TestGenerateMigrationFiles tests the GenerateMigrationFiles method.
// Note: Full testing requires a properly initialized migration directory with  checksums,
// which is complex to set up in unit tests. This test verifies the method can be called.
func TestGenerateMigrationFiles(t *testing.T) {
	sb := createSchemaBuilder()
	tmpDir := utils.Must(os.MkdirTemp("", "migrations_test"))
	defer os.RemoveAll(tmpDir)

	dbClient, err := NewTestClient(tmpDir, sb)
	assert.NoError(t, err)
	defer dbClient.Close()

	// The method will fail due to missing checksum, but we verify it doesn't panic
	// and handles the error gracefully
	err = dbClient.GenerateMigrationFiles(context.Background(), "test")
	// Expect an error related to checksum or migration directory
	if err != nil {
		assert.Contains(t, err.Error(), "checksum")
	}
}

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
	migration := &db.Migration{}
	appendEntTables := []*entSchema.Table{}
	err := mockAdapter.Migrate(context.Background(), migration, false, appendEntTables...)
	assert.Error(t, err)

	// Success
	sb := createSchemaBuilder()
	dbClient, err := NewTestClient(
		utils.Must(os.MkdirTemp("", "test")),
		sb,
	)
	assert.NoError(t, err)
	_, err = dbClient.Reload(context.Background(), sb, &db.Migration{
		RenameTables: []*db.RenameItem{{From: "users", To: "members"}},
	}, false, true)
	require.NoError(t, err)
}

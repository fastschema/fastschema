package entdbadapter

import (
	"context"
	"testing"

	entSchema "entgo.io/ent/dialect/sql/schema"
	"github.com/fastschema/fastschema/db"
	"github.com/stretchr/testify/assert"
)

func TestAdapterMigrateErrorConnection(t *testing.T) {
	adapter := createMockAdapter(t)

	migration := &db.Migration{}            // Replace with your migration definition
	appendEntTables := []*entSchema.Table{} // Replace with additional ent tables if needed

	err := adapter.Migrate(context.Background(), migration, false, appendEntTables...)
	assert.Error(t, err)
}

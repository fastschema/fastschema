package entdbadapter

import (
	"testing"

	entSchema "entgo.io/ent/dialect/sql/schema"
	"github.com/fastschema/fastschema/app"
	"github.com/stretchr/testify/assert"
)

func TestAdapterMigrateErrorConnection(t *testing.T) {
	adapter := createMockAdapter(t)

	migration := &app.Migration{}           // Replace with your migration definition
	appendEntTables := []*entSchema.Table{} // Replace with additional ent tables if needed

	err := adapter.Migrate(migration, appendEntTables...)
	assert.Error(t, err)
}

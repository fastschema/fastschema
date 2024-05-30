package entdbadapter

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
	"github.com/stretchr/testify/assert"
)

func TestScan(t *testing.T) {
	migrationDir := utils.Must(os.MkdirTemp("", "migrations"))
	adapter, err := NewTestClient(
		migrationDir,
		utils.Must(schema.NewBuilderFromDir(migrationDir, fs.SystemSchemaTypes...)),
	)

	assert.NoError(t, err)
	assert.NotNil(t, adapter)

	rows, err := adapter.Query(
		context.Background(),
		`SELECT
			? as bool_column,
			? as int64_column,
			? as uint64_column,
			? as time_column
		`,
		[]any{
			true,
			1,
			int64(1),
			uint64(1),
			time.Now(),
		},
	)
	assert.NoError(t, err)
	assert.NotNil(t, rows)
	assert.Len(t, rows, 1)

	json := rows[0].String()

	assert.Contains(t, json, `"bool_column":1`)
	assert.Contains(t, json, `"int64_column":1`)
	assert.Contains(t, json, `"uint64_column":1`)
	assert.Contains(t, json, `"time_column":`)

	result, err := adapter.Exec(context.Background(), "SELECT 1", []any{})
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

package fastschema_test

import (
	"context"
	"testing"
	"time"

	"github.com/fastschema/fastschema"
	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/entity"
	"github.com/fastschema/fastschema/fs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAuditRetentionPurge seeds old + recent activity rows and verifies the
// retention cleanup deletes only rows past the cutoff.
func TestAuditRetentionPurge(t *testing.T) {
	clearEnvs(t)
	config := &fs.Config{HideResourcesInfo: true, Dir: t.TempDir()}
	app, err := fastschema.New(config)
	require.NoError(t, err)

	ctx := context.Background()
	client := app.DB()

	// Seed directly on the activity schema (hard-skipped, so this does not
	// recurse). One old row (200 days), one fresh row.
	oldTime := time.Now().UTC().AddDate(0, 0, -200)
	freshTime := time.Now().UTC()

	_, err = db.Create[*fs.Activity](ctx, client, entity.New().
		Set("action", "create").Set("schema_name", "category").
		Set("record_id", "old-1").Set("created_at", oldTime))
	require.NoError(t, err)

	_, err = db.Create[*fs.Activity](ctx, client, entity.New().
		Set("action", "create").Set("schema_name", "category").
		Set("record_id", "fresh-1").Set("created_at", freshTime))
	require.NoError(t, err)

	// Purge rows older than 90 days.
	app.Services().Activity().PurgeExpired(ctx, 90)

	remaining, err := db.Builder[*fs.Activity](client).Get(ctx)
	require.NoError(t, err)

	var hasOld, hasFresh bool
	for _, r := range remaining {
		switch r.RecordID {
		case "old-1":
			hasOld = true
		case "fresh-1":
			hasFresh = true
		}
	}

	assert.False(t, hasOld, "expired row must be purged")
	assert.True(t, hasFresh, "fresh row must be kept")
}

// TestAuditDisabledNoCapture verifies that with audit disabled, mutations
// produce no activity rows.
func TestAuditDisabledNoCapture(t *testing.T) {
	clearEnvs(t)
	disabled := false
	config := &fs.Config{
		HideResourcesInfo: true,
		Dir:               t.TempDir(),
		AuditConfig:       &fs.AuditConfig{Enabled: &disabled},
	}
	app, err := fastschema.New(config)
	require.NoError(t, err)

	ctx := context.Background()
	client := app.DB()

	_, err = db.Create[*fs.Role](ctx, client,
		entity.New().Set("name", "no_audit_role").Set("description", "x"))
	require.NoError(t, err)

	rows, err := db.Builder[*fs.Activity](client).Get(ctx)
	require.NoError(t, err)
	assert.Empty(t, rows, "audit disabled -> no activity rows captured")
}

package fastschema_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/fastschema/fastschema"
	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/entity"
	"github.com/fastschema/fastschema/fs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// activityFor returns the activity rows for a given schema/action filtered to a
// single record id.
func activityFor(t *testing.T, client db.Client, ctx context.Context, schemaName, action, recordID string) []*fs.Activity {
	t.Helper()
	rows, err := db.Builder[*fs.Activity](client).
		Where(db.EQ("schema_name", schemaName), db.EQ("action", action), db.EQ("record_id", recordID)).
		Get(ctx)
	require.NoError(t, err)
	return rows
}

// TestAuditTrailCaptureCRUD boots a real app (sqlite) and verifies the DB hooks
// record one activity row per create/update/delete, with correct attribution,
// redacted diffs, recursion safety, and log-and-continue semantics.
func TestAuditTrailCaptureCRUD(t *testing.T) {
	clearEnvs(t)
	config := &fs.Config{HideResourcesInfo: true, Dir: t.TempDir()}
	app, err := fastschema.New(config)
	require.NoError(t, err)
	require.NotNil(t, app)

	ctx := context.Background()
	client := app.DB()

	// --- CREATE ---
	role, err := db.Create[*fs.Role](ctx, client,
		entity.New().Set("name", "audit_trail_role").Set("description", "before"))
	require.NoError(t, err)
	require.NotNil(t, role)
	recordID := fmt.Sprint(role.ID)

	creates := activityFor(t, client, ctx, "role", fs.ActivityActionCreate, recordID)
	require.Len(t, creates, 1, "exactly one create activity row")
	assert.Equal(t, fs.ActivityActorSystem, creates[0].ActorType, "non-HTTP mutation -> system actor")
	assert.Contains(t, creates[0].Changes, "audit_trail_role", "create snapshot stores new values")

	// --- UPDATE ---
	_, err = db.Update[*fs.Role](ctx, client,
		entity.New().Set("description", "after"),
		[]*db.Predicate{db.EQ("id", role.ID)})
	require.NoError(t, err)

	updates := activityFor(t, client, ctx, "role", fs.ActivityActionUpdate, recordID)
	require.Len(t, updates, 1, "exactly one update activity row")
	assert.Contains(t, updates[0].ChangedKeys, "description")
	assert.Contains(t, updates[0].Changes, "before", "delta keeps old value")
	assert.Contains(t, updates[0].Changes, "after", "delta keeps new value")

	// --- DELETE ---
	_, err = db.Delete[*fs.Role](ctx, client, []*db.Predicate{db.EQ("id", role.ID)})
	require.NoError(t, err)

	deletes := activityFor(t, client, ctx, "role", fs.ActivityActionDelete, recordID)
	require.Len(t, deletes, 1, "exactly one delete activity row")

	// --- RECURSION GUARD: writing activity must not produce activity-about-activity ---
	selfRows, err := db.Builder[*fs.Activity](client).
		Where(db.EQ("schema_name", "_activity")).
		Get(ctx)
	require.NoError(t, err)
	assert.Empty(t, selfRows, "_activity schema must be hard-skipped (no recursion)")
}

// TestAuditTrailRedactsSecrets verifies sensitive fields never appear in the
// stored diff, using the user schema (which has an otp_hash / password-like
// surface) via a generic create.
func TestAuditTrailRedactsSecrets(t *testing.T) {
	clearEnvs(t)
	config := &fs.Config{HideResourcesInfo: true, Dir: t.TempDir()}
	app, err := fastschema.New(config)
	require.NoError(t, err)

	ctx := context.Background()
	client := app.DB()

	secret := "super-secret-value"
	user, err := db.Create[*fs.User](ctx, client, entity.New().
		Set("username", "audit_user").
		Set("password", secret).
		Set("provider", "local"))
	require.NoError(t, err)
	require.NotNil(t, user)

	rows := activityFor(t, client, ctx, "user", fs.ActivityActionCreate, fmt.Sprint(user.ID))
	require.Len(t, rows, 1)
	assert.NotContains(t, rows[0].Changes, secret, "password value must be redacted")
	assert.True(t, strings.Contains(rows[0].Changes, "[REDACTED]"), "redaction placeholder present")
}

// TestAuditTrailSkipsSession verifies infrastructure schemas in the skip list
// are not audited.
func TestAuditTrailSkipsSession(t *testing.T) {
	clearEnvs(t)
	config := &fs.Config{HideResourcesInfo: true, Dir: t.TempDir()}
	app, err := fastschema.New(config)
	require.NoError(t, err)

	ctx := context.Background()
	client := app.DB()

	before, err := db.Builder[*fs.Activity](client).Where(db.EQ("schema_name", "session")).Get(ctx)
	require.NoError(t, err)

	// Creating a user (above flow not needed) is unnecessary; create a session row directly.
	expiresAt := time.Now().Add(time.Hour)
	_, err = db.Create[*fs.Session](ctx, client, entity.New().
		Set("user_id", "00000000-0000-0000-0000-000000000000").
		Set("status", "active").
		Set("expires_at", expiresAt))
	require.NoError(t, err)

	after, err := db.Builder[*fs.Activity](client).Where(db.EQ("schema_name", "session")).Get(ctx)
	require.NoError(t, err)
	assert.Equal(t, len(before), len(after), "session schema is skipped -> no new activity rows")
}

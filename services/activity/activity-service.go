// Package activityservice captures an append-only audit trail: one activity row
// per successful create/update/delete on an in-scope schema, attributed to the
// request actor and carrying a redacted before/after diff.
//
// Capture runs in the PostDB* hooks (after the mutation), so the audit write is
// NOT atomic with the mutation. The hook therefore follows a log-and-continue
// policy: a failed audit write is logged and swallowed, never failing the
// underlying mutation. A recursion guard skips the activity schema itself (and
// other infrastructure schemas) so writing a row cannot trigger another row.
package activityservice

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/entity"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/logger"
	"github.com/fastschema/fastschema/pkg/auditlog"
	"github.com/google/uuid"
)

// hardSkipSchema is always skipped regardless of config: auditing the activity
// table itself would recurse infinitely. Must match the Activity schema name
// (see fs/activity.go `name=_activity`).
const hardSkipSchema = "_activity"

// defaultSkipSchemas are infrastructure schemas excluded from auditing by
// default. The config phase can extend this set (but never un-skip activity).
var defaultSkipSchemas = []string{"session", "migration"}

type AppLike interface {
	DB() db.Client
	Logger() logger.Logger
}

// ActivityService writes audit-trail rows from the DB mutation hooks.
type ActivityService struct {
	DB            func() db.Client
	Logger        func() logger.Logger
	skip          map[string]bool
	redactor      *redactor
	cleanupCancel context.CancelFunc
}

func New(app AppLike) *ActivityService {
	return &ActivityService{
		DB:       app.DB,
		Logger:   app.Logger,
		skip:     buildSkipSet(defaultSkipSchemas),
		redactor: newRedactor(nil),
	}
}

// Configure applies operator config: the extra schemas to skip (the activity
// schema is always hard-skipped) and the sensitive-field redaction list. Empty
// inputs fall back to defaults.
func (s *ActivityService) Configure(skipSchemas, redactFields []string) {
	if len(skipSchemas) == 0 {
		skipSchemas = defaultSkipSchemas
	}

	s.skip = buildSkipSet(skipSchemas)
	s.redactor = newRedactor(redactFields)
}

// buildSkipSet builds the skip lookup, always including the hard-skip schema.
func buildSkipSet(schemas []string) map[string]bool {
	skip := map[string]bool{hardSkipSchema: true}
	for _, name := range schemas {
		skip[name] = true
	}

	return skip
}

// shouldSkip reports whether a schema is excluded from auditing.
func (s *ActivityService) shouldSkip(schemaName string) bool {
	return s.skip[schemaName]
}

// writeActivity persists one audit row. It never returns an error: any failure
// is logged and swallowed so the originating mutation is unaffected.
func (s *ActivityService) writeActivity(
	ctx context.Context,
	action, schemaName, recordID string,
	changes any,
	changedKeys []string,
) {
	id, err := uuid.NewV7()
	if err != nil {
		s.Logger().Warn(fmt.Sprintf("audit: failed to generate activity id: %v", err))
		return
	}

	e := entity.New().
		Set("id", id).
		Set("action", action).
		Set("schema_name", schemaName).
		Set("record_id", recordID).
		Set("created_at", time.Now().UTC())

	s.applyActor(e, auditlog.ActorFromContext(ctx))

	if changes != nil {
		if b, mErr := json.Marshal(changes); mErr == nil {
			e.Set("changes", string(b))
		}
	}

	if len(changedKeys) > 0 {
		if b, mErr := json.Marshal(changedKeys); mErr == nil {
			e.Set("changed_keys", string(b))
		}
	}

	if _, err := db.Builder[*fs.Activity](s.DB()).Create(ctx, e); err != nil {
		s.Logger().Warn(fmt.Sprintf(
			"audit: failed to write activity %s on %s record=%s: %v",
			action, schemaName, recordID, err,
		))
	}
}

// applyActor copies the request actor onto the row, falling back to the system
// actor for non-HTTP mutations (background jobs, migrations).
func (s *ActivityService) applyActor(e *entity.Entity, actor *auditlog.ActorContext) {
	if actor == nil {
		e.Set("actor_type", fs.ActivityActorSystem)
		return
	}

	if actor.UserID != nil {
		e.Set("actor_id", *actor.UserID)
	}

	actorType := actor.UserType
	if actorType == "" {
		actorType = fs.ActivityActorSystem
	}

	e.Set("actor_type", actorType).
		Set("ip", actor.IP).
		Set("method", actor.Method).
		Set("path", actor.Path).
		Set("trace_id", actor.TraceID)
}

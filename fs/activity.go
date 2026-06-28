package fs

import (
	"time"

	"github.com/fastschema/fastschema/schema"
	"github.com/google/uuid"
)

// Activity actions captured by the audit trail.
const (
	ActivityActionCreate = "create"
	ActivityActionUpdate = "update"
	ActivityActionDelete = "delete"
)

// Activity actor types: a real authenticated user, an anonymous guest, or the
// system itself (background jobs, migrations, non-HTTP mutations).
const (
	ActivityActorUser   = "user"
	ActivityActorGuest  = "guest"
	ActivityActorSystem = "system"
)

// Activity is a single audit-trail entry: one row per successful create/update/
// delete mutation on an in-scope schema. It is append-only at the API layer.
//
// The schema is named "_activity" (leading underscore) so the common word
// "activity" stays free for user-defined content schemas; the prefix also marks
// it as an internal/system collection (compare session, migration).
//
// The primary key is a v7 UUID so rows stay time-sortable by insertion while
// matching the uuid convention of the other system schemas. The Changes /
// ChangedKeys payloads are stored as portable text (JSON) so they remain
// driver-agnostic (mysql LONGTEXT / postgres text / sqlite text).
type Activity struct {
	_           any        `json:"-" fs:"name=_activity;namespace=_activities;label_field=id"`
	ID          uuid.UUID  `json:"id" fs:"type=uuid"`
	ActorID     *uuid.UUID `json:"actor_id,omitempty" fs:"type=uuid;optional;filterable;sortable"`
	ActorType   string     `json:"actor_type,omitempty" fs:"size=20;optional;filterable"`
	Action      string     `json:"action,omitempty" fs:"size=20;optional;filterable"`
	SchemaName  string     `json:"schema_name,omitempty" fs:"size=255;optional;filterable"`
	RecordID    string     `json:"record_id,omitempty" fs:"size=255;optional;filterable"`
	IP          string     `json:"ip,omitempty" fs:"size=64;optional"`
	Method      string     `json:"method,omitempty" fs:"size=16;optional"`
	Path        string     `json:"path,omitempty" fs:"size=512;optional"`
	TraceID     string     `json:"trace_id,omitempty" fs:"size=64;optional"`
	ChangedKeys string     `json:"changed_keys,omitempty" fs:"type=text;optional"`
	Changes     string     `json:"changes,omitempty" fs:"type=text;optional"`
	CreatedAt   *time.Time `json:"created_at,omitempty" fs:"filterable"`
}

func (a Activity) Schema() *schema.Schema {
	return &schema.Schema{
		Fields: []*schema.Field{},
		DB: &schema.SchemaDB{
			Indexes: []*schema.SchemaDBIndex{
				// Lookup all activity for a given record (entity history view).
				{
					Name:    "idx_activity_schema_record",
					Columns: []string{"schema_name", "record_id"},
				},
				// "What did this actor do, most recent first" queries.
				{
					Name:    "idx_activity_actor_created",
					Columns: []string{"actor_id", "created_at"},
				},
				// Global feed ordered by time + retention-cleanup range scans.
				{
					Name:    "idx_activity_created",
					Columns: []string{"created_at"},
				},
			},
		},
	}
}

package toolservice

import (
	"math"
	"time"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/errors"
)

const (
	activityDefaultLimit = 20
	activityMaxLimit     = 100
)

// ActivityList is the paginated response for the audit-trail query endpoint.
type ActivityList struct {
	Total       int            `json:"total"`
	PerPage     int            `json:"per_page"`
	CurrentPage int            `json:"current_page"`
	LastPage    int            `json:"last_page"`
	Items       []*fs.Activity `json:"items"`
}

// Activity returns audit-trail entries, newest first, with optional filters:
//
//	?schema=  exact schema name
//	?action=  create | update | delete
//	?actor=   actor (user) id
//	?from=    inclusive lower bound on created_at (RFC3339 or YYYY-MM-DD)
//	?to=      inclusive upper bound on created_at
//	?page=    1-based page (default 1)
//	?limit=   page size (default 20, max 100)
//
// The endpoint is read-only; there is no route to mutate activity rows
// (append-only enforced at the API layer). Access is gated by the standard
// resource permission (api.tool.activity), so only admins / explicitly granted
// roles can read it.
func (s *ToolService) Activity(c fs.Context, _ any) (*ActivityList, error) {
	predicates := s.activityPredicates(c)

	total, err := db.Builder[*fs.Activity](s.DB()).Where(predicates...).Count(c)
	if err != nil {
		return nil, errors.BadRequest(err.Error())
	}

	page := c.ArgInt("page", 1)
	if page < 1 {
		page = 1
	}

	limit := c.ArgInt("limit", activityDefaultLimit)
	if limit < 1 {
		limit = activityDefaultLimit
	}
	if limit > activityMaxLimit {
		limit = activityMaxLimit
	}

	// Order by the v7 UUID primary key: it is monotonic by creation time, so
	// "-id" yields newest-first. (created_at is a managed timestamp column and
	// is not a sortable field; the PK is always sortable.)
	items, err := db.Builder[*fs.Activity](s.DB()).
		Where(predicates...).
		Order("-id").
		Limit(uint(limit)).
		Offset(uint((page - 1) * limit)).
		Get(c)
	if err != nil {
		return nil, errors.BadRequest(err.Error())
	}

	lastPage := 0
	if total > 0 {
		lastPage = int(math.Ceil(float64(total) / float64(limit)))
	}

	return &ActivityList{
		Total:       total,
		PerPage:     limit,
		CurrentPage: page,
		LastPage:    lastPage,
		Items:       items,
	}, nil
}

// activityPredicates builds the filter predicates from query arguments.
func (s *ToolService) activityPredicates(c fs.Context) []*db.Predicate {
	predicates := []*db.Predicate{}

	if v := c.Arg("schema", ""); v != "" {
		predicates = append(predicates, db.EQ("schema_name", v))
	}

	if v := c.Arg("action", ""); v != "" {
		predicates = append(predicates, db.EQ("action", v))
	}

	if v := c.Arg("actor", ""); v != "" {
		predicates = append(predicates, db.EQ("actor_id", v))
	}

	if t, ok := parseActivityTime(c.Arg("from", "")); ok {
		predicates = append(predicates, db.GTE("created_at", t))
	}

	if t, ok := parseActivityTime(c.Arg("to", "")); ok {
		predicates = append(predicates, db.LTE("created_at", t))
	}

	return predicates
}

// parseActivityTime accepts RFC3339 timestamps or plain YYYY-MM-DD dates.
func parseActivityTime(v string) (time.Time, bool) {
	if v == "" {
		return time.Time{}, false
	}

	if t, err := time.Parse(time.RFC3339, v); err == nil {
		return t, true
	}

	if t, err := time.Parse("2006-01-02", v); err == nil {
		return t, true
	}

	return time.Time{}, false
}

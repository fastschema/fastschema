package toolservice

import (
	"fmt"
	"sort"
	"time"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/entity"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/schema"
)

type AppLike interface {
	DB() db.Client
}

type ToolService struct {
	DB func() db.Client
}

func New(app AppLike) *ToolService {
	return &ToolService{
		DB: app.DB,
	}
}

func (s *ToolService) CreateResource(api *fs.Resource) {
	api.Group("tool").
		Add(fs.NewResource("stats", s.Stats, &fs.Meta{Get: "/stats"})).
		Add(fs.NewResource("recent", s.Recent, &fs.Meta{Get: "/recent"}))
}

// SchemaCount holds the record count for a single content schema.
type SchemaCount struct {
	Schema string `json:"schema"`
	Label  string `json:"label"`
	Count  int    `json:"count"`
}

type StatsData struct {
	// TotalSchemas counts ALL schemas (including system schemas) for backward compatibility.
	TotalSchemas int `json:"totalSchemas"`
	TotalUsers   int `json:"totalUsers"`
	TotalRoles   int `json:"totalRoles"`
	TotalFiles   int `json:"totalFiles"`
	// TotalContent is the sum of record counts across content schemas only.
	TotalContent int `json:"totalContent"`
	// ContentCounts is the per-content-schema record count, sorted desc by count.
	ContentCounts []SchemaCount `json:"contentCounts"`
}

func (s *ToolService) Stats(c fs.Context, _ any) (_ *StatsData, err error) {
	totalSchemas := len(s.DB().SchemaBuilder().Schemas())
	totalUsers := 0
	totalRoles := 0
	totalFiles := 0

	if totalUsers, err = db.Builder[*fs.User](s.DB()).Count(c); err != nil {
		return nil, err
	}

	if totalRoles, err = db.Builder[*fs.Role](s.DB()).Count(c); err != nil {
		return nil, err
	}

	if totalFiles, err = db.Builder[*fs.File](s.DB()).Count(c); err != nil {
		return nil, err
	}

	contentCounts, totalContent, err := s.contentCounts(c)
	if err != nil {
		return nil, err
	}

	return &StatsData{
		TotalSchemas:  totalSchemas,
		TotalUsers:    totalUsers,
		TotalRoles:    totalRoles,
		TotalFiles:    totalFiles,
		TotalContent:  totalContent,
		ContentCounts: contentCounts,
	}, nil
}

// contentCounts returns the record count per content schema (excluding system and
// junction schemas) and their sum. Counting per schema via the query builder keeps
// soft-delete and cross-dialect handling correct without hand-built SQL.
func (s *ToolService) contentCounts(c fs.Context) ([]SchemaCount, int, error) {
	counts := []SchemaCount{}
	total := 0

	for _, sch := range s.DB().SchemaBuilder().Schemas() {
		if sch.IsSystemSchema || sch.IsJunctionSchema {
			continue
		}

		model, err := s.DB().Model(sch.Name)
		if err != nil {
			// Skip schemas without a backing model rather than failing the endpoint.
			continue
		}

		count, err := model.Query().Count(c, &db.QueryOption{})
		if err != nil {
			return nil, 0, err
		}

		counts = append(counts, SchemaCount{
			Schema: sch.Name,
			Label:  schemaLabel(sch),
			Count:  count,
		})
		total += count
	}

	sort.SliceStable(counts, func(i, j int) bool {
		return counts[i].Count > counts[j].Count
	})

	return counts, total, nil
}

// schemaLabel returns a display label for a schema, falling back to its name.
func schemaLabel(sch *schema.Schema) string {
	if sch.LabelFieldName != "" {
		return sch.LabelFieldName
	}
	return sch.Name
}

const (
	recentDefaultLimit = 10
	recentMaxLimit     = 50
)

// RecentItem describes a recently updated content record for the dashboard.
type RecentItem struct {
	Schema    string    `json:"schema"`
	Label     string    `json:"label"`
	ID        any       `json:"id"`
	Title     string    `json:"title"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// Recent returns the most recently updated content records across all content
// schemas that track timestamps, merged and sorted by update time. Aggregating
// here avoids the dashboard issuing one request per schema.
func (s *ToolService) Recent(c fs.Context, _ any) ([]*RecentItem, error) {
	limit := c.ArgInt("limit", recentDefaultLimit)
	if limit < 1 {
		limit = 1
	}
	if limit > recentMaxLimit {
		limit = recentMaxLimit
	}

	items := []*RecentItem{}

	for _, sch := range s.DB().SchemaBuilder().Schemas() {
		// Skip system/junction schemas and schemas without timestamp columns
		// (no updated_at to order by).
		if sch.IsSystemSchema || sch.IsJunctionSchema || sch.DisableTimestamp {
			continue
		}

		model, err := s.DB().Model(sch.Name)
		if err != nil {
			continue
		}

		// updated_at is not a sortable column, so order by the primary key
		// (always sortable) to fetch the freshest candidate rows per schema.
		// The merged result is then ordered by updated_at in Go below.
		records, err := model.Query().
			Limit(uint(limit)).
			Order("-" + model.Schema().PrimaryKeyName()).
			Get(c)
		if err != nil {
			return nil, err
		}

		for _, record := range records {
			items = append(items, &RecentItem{
				Schema:    sch.Name,
				Label:     schemaLabel(sch),
				ID:        record.ID(),
				Title:     recordTitle(record, sch),
				UpdatedAt: recordUpdatedAt(record),
			})
		}
	}

	sort.SliceStable(items, func(i, j int) bool {
		return items[i].UpdatedAt.After(items[j].UpdatedAt)
	})

	if len(items) > limit {
		items = items[:limit]
	}

	return items, nil
}

// recordTitle picks a human-friendly title using the schema's label field,
// falling back to the record id.
func recordTitle(e *entity.Entity, sch *schema.Schema) string {
	if sch.LabelFieldName != "" {
		if title := e.GetString(sch.LabelFieldName); title != "" {
			return title
		}
	}
	if id := e.ID(); id != nil {
		return fmt.Sprintf("%v", id)
	}
	return ""
}

// recordUpdatedAt returns the record's update time, falling back to created_at.
// A freshly created record has a null updated_at until first edited, so without
// the fallback its timestamp would be the zero value.
func recordUpdatedAt(e *entity.Entity) time.Time {
	if t := timeFieldValue(e, entity.FieldUpdatedAt); !t.IsZero() {
		return t
	}
	return timeFieldValue(e, entity.FieldCreatedAt)
}

// timeFieldValue reads a time field, handling both time.Time and *time.Time.
func timeFieldValue(e *entity.Entity, name string) time.Time {
	switch t := e.Get(name).(type) {
	case time.Time:
		return t
	case *time.Time:
		if t != nil {
			return *t
		}
	}
	return time.Time{}
}

package activityservice

import (
	"testing"

	"github.com/fastschema/fastschema/entity"
	"github.com/fastschema/fastschema/schema"
	"github.com/stretchr/testify/assert"
)

func TestBuildSkipSetIncludesHardSkip(t *testing.T) {
	skip := buildSkipSet([]string{"session", "migration"})

	// activity is always hard-skipped to prevent recursion.
	assert.True(t, skip["_activity"])
	assert.True(t, skip["session"])
	assert.True(t, skip["migration"])
	assert.False(t, skip["user"])
}

func TestBuildSkipSetEmptyStillHardSkipsActivity(t *testing.T) {
	skip := buildSkipSet(nil)
	assert.True(t, skip["_activity"])
}

func TestShouldSkip(t *testing.T) {
	s := &ActivityService{skip: buildSkipSet(defaultSkipSchemas)}

	assert.True(t, s.shouldSkip("_activity"))
	assert.True(t, s.shouldSkip("session"))
	assert.False(t, s.shouldSkip("user"))
	assert.False(t, s.shouldSkip("category"))
}

func TestRecordIDUsesPrimaryKey(t *testing.T) {
	sch := &schema.Schema{Name: "user", PrimaryFieldName: "id"}
	e := entity.New().Set("id", "abc-123").Set("name", "alice")

	assert.Equal(t, "abc-123", recordID(sch, e))
}

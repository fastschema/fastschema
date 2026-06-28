package activityservice

import (
	"fmt"
	"testing"

	"github.com/fastschema/fastschema/entity"
	"github.com/stretchr/testify/assert"
)

func TestRedactorShouldRedact(t *testing.T) {
	r := newRedactor(nil)

	assert.True(t, r.shouldRedact("password"))
	assert.True(t, r.shouldRedact("Password"))
	assert.True(t, r.shouldRedact("password_hash"))
	assert.True(t, r.shouldRedact("user_api_key"))
	assert.True(t, r.shouldRedact("otp_hash"))
	assert.False(t, r.shouldRedact("name"))
	assert.False(t, r.shouldRedact("email"))
}

func TestRedactorCustomFields(t *testing.T) {
	r := newRedactor([]string{"ssn"})

	assert.True(t, r.shouldRedact("ssn"))
	// custom list replaces defaults entirely
	assert.False(t, r.shouldRedact("password"))
}

func TestSnapshotRedactsSensitive(t *testing.T) {
	r := newRedactor(nil)
	e := entity.New().Set("name", "alice").Set("password", "s3cret")

	snap := r.snapshot(e)

	assert.Equal(t, "alice", snap["name"])
	assert.Equal(t, redactedPlaceholder, snap["password"])
}

func TestSnapshotNilEntity(t *testing.T) {
	r := newRedactor(nil)
	assert.Empty(t, r.snapshot(nil))
}

func TestComputeDeltaOnlyChangedKeys(t *testing.T) {
	r := newRedactor(nil)
	old := entity.New().Set("name", "alice").Set("age", 30).Set("password", "old")
	upd := entity.New().Set("name", "alice").Set("age", 31).Set("password", "new")

	changes, keys := r.computeDelta(old, upd)

	// name unchanged -> excluded; age + password changed.
	assert.ElementsMatch(t, []string{"age", "password"}, keys)
	assert.Len(t, changes, 2)

	age := changes["age"].(map[string]any)
	assert.Equal(t, "30", fmt.Sprint(age["old"]))
	assert.Equal(t, "31", fmt.Sprint(age["new"]))

	// sensitive field is redacted on both sides.
	pw := changes["password"].(map[string]any)
	assert.Equal(t, redactedPlaceholder, pw["old"])
	assert.Equal(t, redactedPlaceholder, pw["new"])
}

func TestComputeDeltaNoChange(t *testing.T) {
	r := newRedactor(nil)
	old := entity.New().Set("name", "alice")
	upd := entity.New().Set("name", "alice")

	changes, keys := r.computeDelta(old, upd)

	assert.Empty(t, keys)
	assert.Empty(t, changes)
}

func TestValuesEqual(t *testing.T) {
	assert.True(t, valuesEqual(nil, nil))
	assert.True(t, valuesEqual(1, 1))
	assert.True(t, valuesEqual("a", "a"))
	assert.False(t, valuesEqual(1, 2))
	assert.False(t, valuesEqual(nil, ""))
}

package activityservice

import (
	"fmt"

	"github.com/fastschema/fastschema/entity"
)

// snapshot returns a redacted copy of every field on an entity. Used for the
// create payload (new values) and the delete payload (last-known values).
func (r *redactor) snapshot(e *entity.Entity) map[string]any {
	out := make(map[string]any)
	if e == nil {
		return out
	}

	for _, k := range e.Keys() {
		out[k] = r.redactValue(k, e.Get(k))
	}

	return out
}

// computeDelta diffs the old row against the update set, restricted to the keys
// actually being written (newData). It returns a {key: {"old","new"}} map of
// redacted changes plus the ordered list of changed keys.
func (r *redactor) computeDelta(old, newData *entity.Entity) (map[string]any, []string) {
	changes := make(map[string]any)
	changedKeys := make([]string, 0)
	if newData == nil {
		return changes, changedKeys
	}

	for _, k := range newData.Keys() {
		newVal := newData.Get(k)

		var oldVal any
		if old != nil {
			oldVal = old.Get(k)
		}

		if valuesEqual(oldVal, newVal) {
			continue
		}

		changedKeys = append(changedKeys, k)
		changes[k] = map[string]any{
			"old": r.redactValue(k, oldVal),
			"new": r.redactValue(k, newVal),
		}
	}

	return changes, changedKeys
}

// valuesEqual compares two field values loosely. Audit diffs only need to know
// "did this change"; a string-form comparison sidesteps type mismatches between
// DB-loaded originals (e.g. time.Time) and update payloads without panicking on
// non-comparable values.
func valuesEqual(a, b any) bool {
	if a == nil && b == nil {
		return true
	}

	return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
}

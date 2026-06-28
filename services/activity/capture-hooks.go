package activityservice

import (
	"context"
	"fmt"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/entity"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/schema"
)

// CaptureCreateHook logs a create. The full (redacted) new entity is stored as
// the change payload; there are no "changed keys" for a creation.
func (s *ActivityService) CaptureCreateHook(
	ctx context.Context,
	sch *schema.Schema,
	dataCreate *entity.Entity,
	id any,
) error {
	if s.shouldSkip(sch.Name) {
		return nil
	}

	changes := s.redactor.snapshot(dataCreate)
	s.writeActivity(ctx, fs.ActivityActionCreate, sch.Name, fmt.Sprint(id), changes, nil)

	return nil
}

// CaptureUpdateHook logs an update per affected row, storing the before/after
// delta for the fields that actually changed.
func (s *ActivityService) CaptureUpdateHook(
	ctx context.Context,
	sch *schema.Schema,
	predicates *[]*db.Predicate,
	updateData *entity.Entity,
	originalEntities []*entity.Entity,
	affected int,
) error {
	if s.shouldSkip(sch.Name) {
		return nil
	}

	for _, original := range originalEntities {
		changes, changedKeys := s.redactor.computeDelta(original, updateData)
		if len(changedKeys) == 0 {
			continue
		}

		s.writeActivity(
			ctx,
			fs.ActivityActionUpdate,
			sch.Name,
			recordID(sch, original),
			changes,
			changedKeys,
		)
	}

	return nil
}

// CaptureDeleteHook logs a delete per affected row, storing the last-known
// (redacted) snapshot of each deleted entity.
func (s *ActivityService) CaptureDeleteHook(
	ctx context.Context,
	sch *schema.Schema,
	predicates *[]*db.Predicate,
	originalEntities []*entity.Entity,
	affected int,
) error {
	if s.shouldSkip(sch.Name) {
		return nil
	}

	for _, original := range originalEntities {
		changes := s.redactor.snapshot(original)
		s.writeActivity(ctx, fs.ActivityActionDelete, sch.Name, recordID(sch, original), changes, nil)
	}

	return nil
}

// recordID resolves the primary-key value of an entity as a string. RecordID is
// stored as text so it accommodates uuid, integer, and string primary keys.
func recordID(sch *schema.Schema, e *entity.Entity) string {
	if pk := sch.PrimaryKeyName(); pk != "" {
		if v := e.Get(pk); v != nil {
			return fmt.Sprint(v)
		}
	}

	return fmt.Sprint(e.ID())
}

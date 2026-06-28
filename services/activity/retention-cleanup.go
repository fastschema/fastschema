package activityservice

import (
	"context"
	"fmt"
	"time"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/fs"
)

// cleanupInterval is how often expired audit rows are purged. Daily is enough:
// each run only deletes the day's worth of newly-expired rows.
const cleanupInterval = 24 * time.Hour

// StartRetentionCleanup launches a background loop that periodically deletes
// audit rows older than retentionDays. A non-positive retentionDays keeps rows
// forever (no cleanup). Safe to call once; call StopRetentionCleanup to stop.
func (s *ActivityService) StartRetentionCleanup(retentionDays int) {
	if retentionDays <= 0 {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	s.cleanupCancel = cancel

	go s.runRetentionLoop(ctx, retentionDays)
}

// StopRetentionCleanup stops the background cleanup loop (no-op if not running).
func (s *ActivityService) StopRetentionCleanup() {
	if s.cleanupCancel != nil {
		s.cleanupCancel()
		s.cleanupCancel = nil
	}
}

func (s *ActivityService) runRetentionLoop(ctx context.Context, retentionDays int) {
	// Purge once on start so a restart reclaims a backlog, then on each tick.
	s.PurgeExpired(ctx, retentionDays)

	ticker := time.NewTicker(cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.PurgeExpired(ctx, retentionDays)
		}
	}
}

// PurgeExpired deletes audit rows created before the retention cutoff. Errors
// are logged and swallowed; cleanup is best-effort and must never crash the app.
func (s *ActivityService) PurgeExpired(ctx context.Context, retentionDays int) {
	cutoff := time.Now().UTC().AddDate(0, 0, -retentionDays)

	affected, err := db.Builder[*fs.Activity](s.DB()).
		Where(db.LT("created_at", cutoff)).
		Delete(ctx)
	if err != nil {
		s.Logger().Warn(fmt.Sprintf("audit: retention cleanup failed: %v", err))
		return
	}

	if affected > 0 {
		s.Logger().Info(fmt.Sprintf("audit: retention cleanup removed %d expired rows", affected))
	}
}

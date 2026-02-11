package fs

import "time"

// Migration tracks applied database migrations
type Migration struct {
	_         any       `json:"-" fs:"namespace=migrations;label_field=version"`
	ID        uint64    `json:"id,omitempty" fs:"type=uuid"`
	Version   string    `json:"version" fs:"unique"`          // Timestamp: "20251212093000"
	Name      string    `json:"name,omitempty" fs:"optional"` // Human-readable name
	AppliedAt time.Time `json:"applied_at"`
}

// MigrationFile represents a single migration file pair (up/down)
type MigrationFile struct {
	Version   string     // Timestamp: "20251212093000"
	Name      string     // Human-readable name
	UpFile    string     // Path to .up.sql file
	DownFile  string     // Path to .down.sql file
	UpSQL     string     // Content of up migration
	DownSQL   string     // Content of down migration
	AppliedAt *time.Time // When migration was applied (nil if pending)
}

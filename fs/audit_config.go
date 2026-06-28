package fs

// DefaultAuditRetentionDays is the default age (in days) after which audit
// entries are eligible for cleanup. Operators with compliance needs should
// raise this (PCI ~365, HIPAA ~2190).
const DefaultAuditRetentionDays = 90

// DefaultAuditSkipSchemas are infrastructure schemas excluded from auditing by
// default (in addition to the always-skipped `activity` schema).
var DefaultAuditSkipSchemas = []string{"session", "migration"}

// AuditConfig controls the activity / audit-trail feature.
type AuditConfig struct {
	// Enabled toggles audit capture. nil means "use default" (enabled).
	Enabled *bool `json:"enabled,omitempty"`
	// RetentionDays: rows older than this are purged. <= 0 disables cleanup
	// (keep forever).
	RetentionDays int `json:"retention_days,omitempty"`
	// SkipSchemas are additional schema names excluded from auditing. The
	// `activity` schema is always skipped regardless of this list.
	SkipSchemas []string `json:"skip_schemas,omitempty"`
	// RedactFields overrides the default sensitive-field substring list used to
	// mask values in stored diffs.
	RedactFields []string `json:"redact_fields,omitempty"`
}

// IsEnabled reports whether audit capture should run (defaults to true).
func (c *AuditConfig) IsEnabled() bool {
	if c == nil || c.Enabled == nil {
		return true
	}

	return *c.Enabled
}

// Clone returns a deep copy of the audit config.
func (c *AuditConfig) Clone() *AuditConfig {
	if c == nil {
		return nil
	}

	clone := &AuditConfig{RetentionDays: c.RetentionDays}
	if c.Enabled != nil {
		enabled := *c.Enabled
		clone.Enabled = &enabled
	}
	clone.SkipSchemas = append([]string{}, c.SkipSchemas...)
	clone.RedactFields = append([]string{}, c.RedactFields...)

	return clone
}

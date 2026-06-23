package toolservice

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/fs"
)

func MigrationNew(c context.Context, client db.Client, name string) (*fs.MigrationFile, error) {
	version := GenerateVersion()
	mf := &fs.MigrationFile{
		Version: version,
		Name:    SanitizeMigrationName(name),
		UpSQL:   "-- Write your UP migration SQL here\n",
		DownSQL: "-- Write your DOWN migration SQL here\n",
	}

	if err := WriteMigrationFiles(client.Config().MigrationDir, mf); err != nil {
		return nil, err
	}

	fmt.Printf("Created migration files:\n")
	fmt.Printf("  UP:   %s\n", mf.UpFile)
	fmt.Printf("  DOWN: %s\n", mf.DownFile)
	return mf, nil
}

func MigrationGenerate(c context.Context, client db.Client, name string) error {
	if err := client.GenerateMigrationFiles(c, name); err != nil {
		return err
	}

	// Find the generated one: The latest migration file
	files, err := LoadMigrationFiles(client.Config().MigrationDir)
	if err != nil {
		return fmt.Errorf("failed to reload migration files: %w", err)
	}

	if len(files) == 0 {
		return fmt.Errorf("migration reported success but no files found")
	}

	mf := files[len(files)-1]
	fmt.Printf("Generated migration files:\n")
	fmt.Printf("  UP:   %s\n", mf.UpFile)
	fmt.Printf("  DOWN: %s\n", mf.DownFile)
	return nil
}

func MigrationUp(c context.Context, client db.Client, count int) ([]*fs.MigrationFile, error) {
	_, pending, err := MigrationStatus(c, client)
	if err != nil {
		return nil, err
	}

	if len(pending) == 0 {
		return []*fs.MigrationFile{}, nil
	}

	if count > 0 && count < len(pending) {
		pending = pending[:count]
	}

	// Read all migration SQL files before starting transaction
	for _, mf := range pending {
		if err := ReadMigrationSQL(mf); err != nil {
			return nil, fmt.Errorf("failed to read migration %s: %w", mf.Version, err)
		}
	}

	applied := make([]*fs.MigrationFile, 0, len(pending))

	// Execute all migrations within a single transaction
	err = db.WithTx(client, c, func(tx db.Client) error {
		for _, mf := range pending {
			if err := executeMigration(c, mf.UpSQL, tx); err != nil {
				return fmt.Errorf("failed to apply migration %s: %w", mf.Version, err)
			}

			if err := RecordAppliedMigration(c, tx, mf); err != nil {
				return fmt.Errorf("failed to record migration %s: %w", mf.Version, err)
			}

			now := time.Now()
			mf.AppliedAt = &now
			applied = append(applied, mf)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	if len(applied) == 0 {
		fmt.Println("No pending migrations to apply")
		return nil, nil
	}

	fmt.Printf("Applied %d migration(s):\n", len(applied))
	for _, mf := range applied {
		fmt.Printf("  - %s_%s\n", mf.Version, mf.Name)
	}

	return applied, nil
}

func MigrationDown(c context.Context, client db.Client, count int) ([]*fs.MigrationFile, error) {
	appliedMigs, _, err := MigrationStatus(c, client)
	if err != nil {
		return nil, err
	}

	if len(appliedMigs) == 0 {
		return []*fs.MigrationFile{}, nil
	}

	if count <= 0 {
		count = 1
	}

	if count > len(appliedMigs) {
		count = len(appliedMigs)
	}

	// Get the migrations to roll back (from newest to oldest)
	toRollback := make([]*fs.MigrationFile, 0, count)
	for i := len(appliedMigs) - 1; i >= len(appliedMigs)-count; i-- {
		toRollback = append(toRollback, appliedMigs[i])
	}

	// Read all migration SQL files and validate before starting transaction
	for _, mf := range toRollback {
		if err := ReadMigrationSQL(mf); err != nil {
			return nil, fmt.Errorf("failed to read migration %s: %w", mf.Version, err)
		}

		if strings.TrimSpace(mf.DownSQL) == "" {
			return nil, fmt.Errorf("migration %s has no down script", mf.Version)
		}
	}

	rolledBack := make([]*fs.MigrationFile, 0, count)

	// Execute all rollbacks within a single transaction
	err = db.WithTx(client, c, func(tx db.Client) error {
		for _, mf := range toRollback {
			if err := executeMigration(c, mf.DownSQL, tx); err != nil {
				return fmt.Errorf("failed to rollback migration %s: %w", mf.Version, err)
			}

			if err := RemoveAppliedMigration(c, tx, mf.Version); err != nil {
				return fmt.Errorf("failed to remove migration record %s: %w", mf.Version, err)
			}

			mf.AppliedAt = nil
			rolledBack = append(rolledBack, mf)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	if len(rolledBack) == 0 {
		fmt.Println("No migrations to roll back")
		return nil, nil
	}

	fmt.Printf("Rolled back %d migration(s):\n", len(rolledBack))
	for _, mf := range rolledBack {
		fmt.Printf("  - %s_%s\n", mf.Version, mf.Name)
	}
	return rolledBack, nil
}

func MigrationStatus(c context.Context, client db.Client) (applied, pending []*fs.MigrationFile, err error) {
	allFiles, err := LoadMigrationFiles(client.Config().MigrationDir)
	if err != nil {
		return nil, nil, err
	}

	appliedMap, err := GetAppliedMigrations(c, client)
	if err != nil {
		return nil, nil, err
	}

	applied = make([]*fs.MigrationFile, 0)
	pending = make([]*fs.MigrationFile, 0)

	for _, mf := range allFiles {
		if dbMig, ok := appliedMap[mf.Version]; ok {
			mf.AppliedAt = &dbMig.AppliedAt
			applied = append(applied, mf)
		} else {
			pending = append(pending, mf)
		}
	}

	fmt.Printf("Migration Status:\n\n")

	if len(applied) > 0 {
		fmt.Printf("Applied (%d):\n", len(applied))
		for _, mf := range applied {
			fmt.Printf("  - %s_%s (applied: %s)\n",
				mf.Version, mf.Name,
				mf.AppliedAt.Format("2006-01-02 15:04:05"))
		}
	} else {
		fmt.Println("Applied (0): none")
	}

	fmt.Println()

	if len(pending) > 0 {
		fmt.Printf("Pending (%d):\n", len(pending))
		for _, mf := range pending {
			fmt.Printf("  ○ %s_%s\n", mf.Version, mf.Name)
		}
	} else {
		fmt.Println("Pending (0): none")
	}

	return applied, pending, nil
}

// GenerateVersion creates a timestamp-based version string
func GenerateVersion() string {
	return time.Now().Format("20060102150405")
}

// ParseMigrationFilename extracts version and name from filename
// Expected format: {version}_{name}.up.sql or {version}_{name}.down.sql
func ParseMigrationFilename(filename string) (version, name, direction string, err error) {
	base := filepath.Base(filename)

	re := regexp.MustCompile(`^(\d+)_(.+)\.(up|down)\.sql$`)
	matches := re.FindStringSubmatch(base)
	if len(matches) != 4 {
		return "", "", "", fmt.Errorf("invalid migration filename: %s", base)
	}

	return matches[1], matches[2], matches[3], nil
}

// LoadMigrationFiles loads all migration files from directory
func LoadMigrationFiles(dir string) ([]*fs.MigrationFile, error) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return []*fs.MigrationFile{}, nil
	}

	files, err := filepath.Glob(filepath.Join(dir, "*.up.sql"))
	if err != nil {
		return nil, fmt.Errorf("failed to read migration directory: %w", err)
	}

	migrations := make(map[string]*fs.MigrationFile)
	for _, upFile := range files {
		version, name, _, err := ParseMigrationFilename(upFile)
		if err != nil {
			fmt.Printf("Skipping invalid migration file: %s\n", upFile)
			continue
		}

		downFile := strings.TrimSuffix(upFile, ".up.sql") + ".down.sql"
		migrations[version] = &fs.MigrationFile{
			Version:  version,
			Name:     name,
			UpFile:   upFile,
			DownFile: downFile,
		}
	}

	// Sort by version
	result := make([]*fs.MigrationFile, 0, len(migrations))
	for _, m := range migrations {
		result = append(result, m)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Version < result[j].Version
	})

	return result, nil
}

// WriteMigrationFiles writes up/down SQL files to the migration directory
func WriteMigrationFiles(dir string, mf *fs.MigrationFile) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create migration directory: %w", err)
	}

	baseName := fmt.Sprintf("%s_%s", mf.Version, mf.Name)
	upPath := filepath.Join(dir, baseName+".up.sql")
	downPath := filepath.Join(dir, baseName+".down.sql")

	if err := os.WriteFile(upPath, []byte(mf.UpSQL), 0600); err != nil {
		return fmt.Errorf("failed to write up migration: %w", err)
	}

	if err := os.WriteFile(downPath, []byte(mf.DownSQL), 0600); err != nil {
		return fmt.Errorf("failed to write down migration: %w", err)
	}

	mf.UpFile = upPath
	mf.DownFile = downPath

	return nil
}

// ReadMigrationSQL reads the SQL content from migration files
func ReadMigrationSQL(mf *fs.MigrationFile) error {
	if mf.UpFile != "" {
		content, err := os.ReadFile(mf.UpFile)
		if err != nil {
			return fmt.Errorf("failed to read up migration: %w", err)
		}
		mf.UpSQL = string(content)
	}

	if mf.DownFile != "" {
		if _, err := os.Stat(mf.DownFile); err == nil {
			content, err := os.ReadFile(mf.DownFile)
			if err != nil {
				return fmt.Errorf("failed to read down migration: %w", err)
			}
			mf.DownSQL = string(content)
		}
	}

	return nil
}

// GetAppliedMigrations retrieves applied migrations from the database
func GetAppliedMigrations(ctx context.Context, client db.Client) (map[string]*fs.Migration, error) {
	model, err := client.Model("migration")
	if err != nil {
		return nil, fmt.Errorf("migration model not found: %w", err)
	}

	entities, err := model.Query().Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query migrations: %w", err)
	}

	result := make(map[string]*fs.Migration)
	for _, e := range entities {
		version, _ := e.Get("version").(string)
		name, _ := e.Get("name").(string)
		appliedAt, _ := e.Get("applied_at").(time.Time)

		id, _ := e.ID().(uint64)
		result[version] = &fs.Migration{
			ID:        id,
			Version:   version,
			Name:      name,
			AppliedAt: appliedAt,
		}
	}

	return result, nil
}

// RecordAppliedMigration marks a migration as applied in the database
func RecordAppliedMigration(ctx context.Context, client db.Client, mf *fs.MigrationFile) error {
	model, err := client.Model("migration")
	if err != nil {
		return fmt.Errorf("migration model not found: %w", err)
	}

	_, err = model.CreateFromJSON(ctx, fmt.Sprintf(
		`{"version": "%s", "name": "%s", "applied_at": "%s"}`,
		mf.Version, mf.Name, time.Now().UTC().Format("2006-01-02 15:04:05"),
	))
	if err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	return nil
}

// RemoveAppliedMigration removes a migration record from the database
func RemoveAppliedMigration(ctx context.Context, client db.Client, version string) error {
	model, err := client.Model("migration")
	if err != nil {
		return fmt.Errorf("migration model not found: %w", err)
	}

	_, err = model.Mutation().Where(db.EQ("version", version)).Delete(ctx)
	if err != nil {
		return fmt.Errorf("failed to remove migration record: %w", err)
	}

	return nil
}

// executeMigration runs the SQL in the migration
func executeMigration(ctx context.Context, sqlContent string, client db.Client) error {
	if strings.TrimSpace(sqlContent) == "" {
		return nil
	}

	// Split by semicolons and execute each statement
	statements := SplitSQLStatements(sqlContent)
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" || strings.HasPrefix(stmt, "--") {
			continue
		}
		if _, err := client.Exec(ctx, stmt); err != nil {
			return fmt.Errorf("failed to execute SQL: %w\nStatement: %s", err, stmt)
		}
	}

	return nil
}

// SplitSQLStatements splits SQL content by semicolons, handling edge cases
func SplitSQLStatements(content string) []string {
	// Simple split - handles most cases
	// For complex cases with stored procedures, users should use transactions
	statements := strings.Split(content, ";")
	result := make([]string, 0, len(statements))

	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt != "" {
			result = append(result, stmt)
		}
	}

	return result
}

// SanitizeMigrationName converts a name to a valid migration filename component
func SanitizeMigrationName(name string) string {
	// Replace spaces and special characters with underscores
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, "-", "_")

	// Remove any non-alphanumeric characters except underscores
	var result strings.Builder
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
			result.WriteRune(r)
		}
	}

	// Trim leading/trailing underscores
	return strings.Trim(result.String(), "_")
}

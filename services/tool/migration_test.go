package toolservice_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/fastschema/fastschema/fs"
	toolservice "github.com/fastschema/fastschema/services/tool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateVersion(t *testing.T) {
	version := toolservice.GenerateVersion()
	assert.Len(t, version, 14, "Version should be 14 characters (YYYYMMDDHHmmss)")

	// Verify it's a valid timestamp format
	_, err := time.Parse("20060102150405", version)
	assert.NoError(t, err, "Version should be a valid timestamp")
}

func TestParseMigrationFilename(t *testing.T) {
	tests := []struct {
		name          string
		filename      string
		wantVersion   string
		wantName      string
		wantDirection string
		wantErr       bool
	}{
		{
			name:          "valid up file",
			filename:      "20251212093000_add_users_table.up.sql",
			wantVersion:   "20251212093000",
			wantName:      "add_users_table",
			wantDirection: "up",
			wantErr:       false,
		},
		{
			name:          "valid down file",
			filename:      "20251212093000_add_users_table.down.sql",
			wantVersion:   "20251212093000",
			wantName:      "add_users_table",
			wantDirection: "down",
			wantErr:       false,
		},
		{
			name:          "full path",
			filename:      "/path/to/migrations/20251212093000_create_posts.up.sql",
			wantVersion:   "20251212093000",
			wantName:      "create_posts",
			wantDirection: "up",
			wantErr:       false,
		},
		{
			name:     "invalid format - no underscore",
			filename: "20251212093000.up.sql",
			wantErr:  true,
		},
		{
			name:          "valid format - short version",
			filename:      "2025121209_test.up.sql",
			wantVersion:   "2025121209",
			wantName:      "test",
			wantDirection: "up",
			wantErr:       false,
		},
		{
			name:     "invalid format - no direction",
			filename: "20251212093000_test.sql",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			version, name, direction, err := toolservice.ParseMigrationFilename(tt.filename)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantVersion, version)
			assert.Equal(t, tt.wantName, name)
			assert.Equal(t, tt.wantDirection, direction)
		})
	}
}

func TestWriteAndLoadMigrationFiles(t *testing.T) {
	dir := t.TempDir()

	// Write migration files
	mf := &fs.MigrationFile{
		Version: "20251212100000",
		Name:    "test_migration",
		UpSQL:   "CREATE TABLE test (id INT);",
		DownSQL: "DROP TABLE test;",
	}

	err := toolservice.WriteMigrationFiles(dir, mf)
	require.NoError(t, err)

	// Verify files were created
	assert.FileExists(t, mf.UpFile)
	assert.FileExists(t, mf.DownFile)

	// Load migration files
	migrations, err := toolservice.LoadMigrationFiles(dir)
	require.NoError(t, err)
	require.Len(t, migrations, 1)

	assert.Equal(t, "20251212100000", migrations[0].Version)
	assert.Equal(t, "test_migration", migrations[0].Name)

	// Read SQL content
	err = toolservice.ReadMigrationSQL(migrations[0])
	require.NoError(t, err)
	assert.Equal(t, "CREATE TABLE test (id INT);", migrations[0].UpSQL)
	assert.Equal(t, "DROP TABLE test;", migrations[0].DownSQL)
}

func TestLoadMigrationFiles_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	migrations, err := toolservice.LoadMigrationFiles(dir)
	require.NoError(t, err)
	assert.Empty(t, migrations)
}

func TestLoadMigrationFiles_NonExistentDir(t *testing.T) {
	migrations, err := toolservice.LoadMigrationFiles("/non/existent/path")
	require.NoError(t, err)
	assert.Empty(t, migrations)
}

func TestLoadMigrationFiles_Ordering(t *testing.T) {
	dir := t.TempDir()

	// Create migrations in non-chronological order
	migrations := []*fs.MigrationFile{
		{Version: "20251212120000", Name: "third", UpSQL: "-- third", DownSQL: "-- third"},
		{Version: "20251212100000", Name: "first", UpSQL: "-- first", DownSQL: "-- first"},
		{Version: "20251212110000", Name: "second", UpSQL: "-- second", DownSQL: "-- second"},
	}

	for _, mf := range migrations {
		require.NoError(t, toolservice.WriteMigrationFiles(dir, mf))
	}

	// Load and verify order
	loaded, err := toolservice.LoadMigrationFiles(dir)
	require.NoError(t, err)
	require.Len(t, loaded, 3)

	assert.Equal(t, "20251212100000", loaded[0].Version)
	assert.Equal(t, "20251212110000", loaded[1].Version)
	assert.Equal(t, "20251212120000", loaded[2].Version)
}

func TestSanitizeName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"add users table", "add_users_table"},
		{"Add-Users-Table", "add_users_table"},
		{"create_posts", "create_posts"},
		{"test!@#$%name", "testname"},
		{"__leading__trailing__", "leading__trailing"},
		{"MixedCase123", "mixedcase123"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := toolservice.SanitizeMigrationName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSplitSQLStatements(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "single statement",
			input:    "CREATE TABLE test (id INT);",
			expected: []string{"CREATE TABLE test (id INT)"},
		},
		{
			name:     "multiple statements",
			input:    "CREATE TABLE a (id INT); CREATE TABLE b (id INT);",
			expected: []string{"CREATE TABLE a (id INT)", "CREATE TABLE b (id INT)"},
		},
		{
			name:     "with newlines",
			input:    "CREATE TABLE a (id INT);\n\nCREATE TABLE b (id INT);",
			expected: []string{"CREATE TABLE a (id INT)", "CREATE TABLE b (id INT)"},
		},
		{
			name:     "empty statements filtered",
			input:    ";;CREATE TABLE a (id INT);;;",
			expected: []string{"CREATE TABLE a (id INT)"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toolservice.SplitSQLStatements(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestReadMigrationSQL_MissingDownFile(t *testing.T) {
	dir := t.TempDir()

	// Create only up file
	upFile := filepath.Join(dir, "20251212100000_test.up.sql")
	require.NoError(t, os.WriteFile(upFile, []byte("CREATE TABLE test;"), 0644))

	mf := &fs.MigrationFile{
		Version:  "20251212100000",
		Name:     "test",
		UpFile:   upFile,
		DownFile: filepath.Join(dir, "20251212100000_test.down.sql"),
	}

	// Should not error, just leave DownSQL empty
	err := toolservice.ReadMigrationSQL(mf)
	require.NoError(t, err)
	assert.Equal(t, "CREATE TABLE test;", mf.UpSQL)
	assert.Empty(t, mf.DownSQL)
}

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"
)

// migrateJSONToYAML migrates all JSON schema files to YAML format
func migrateJSONToYAML(c *cli.Context) error {
	schemaDir := filepath.Join(c.Args().Get(0), "data", "schemas")

	// Check if schemas directory exists
	if _, err := os.Stat(schemaDir); os.IsNotExist(err) {
		return fmt.Errorf("schemas directory not found at %s", schemaDir)
	}

	entries, err := os.ReadDir(schemaDir)
	if err != nil {
		return fmt.Errorf("failed to read schemas directory: %w", err)
	}

	var (
		migrated    int
		failed      int
		skipped     int
		failedFiles []string
	)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Only process JSON files
		if !strings.HasSuffix(entry.Name(), ".json") {
			skipped++
			continue
		}

		jsonPath := filepath.Join(schemaDir, entry.Name())
		yamlPath := filepath.Join(schemaDir, strings.TrimSuffix(entry.Name(), ".json")+".yaml")

		// Check if YAML file already exists
		if _, err := os.Stat(yamlPath); err == nil {
			fmt.Printf("⚠️  YAML file already exists: %s (skipping)\n", yamlPath)
			skipped++
			continue
		}

		// Read JSON file
		jsonData, err := os.ReadFile(jsonPath)
		if err != nil {
			fmt.Printf("❌ Failed to read JSON file: %s - %v\n", jsonPath, err)
			failed++
			failedFiles = append(failedFiles, entry.Name())
			continue
		}

		// Parse JSON to validate and get structured data
		var data map[string]any
		if err := json.Unmarshal(jsonData, &data); err != nil {
			fmt.Printf("❌ Invalid JSON format: %s - %v\n", jsonPath, err)
			failed++
			failedFiles = append(failedFiles, entry.Name())
			continue
		}

		// Convert to YAML
		yamlData, err := yaml.Marshal(data)
		if err != nil {
			fmt.Printf("❌ Failed to convert to YAML: %s - %v\n", jsonPath, err)
			failed++
			failedFiles = append(failedFiles, entry.Name())
			continue
		}

		// Write YAML file
		if err := os.WriteFile(yamlPath, yamlData, 0600); err != nil {
			fmt.Printf("❌ Failed to write YAML file: %s - %v\n", yamlPath, err)
			failed++
			failedFiles = append(failedFiles, entry.Name())
			continue
		}

		fmt.Printf("✅ Migrated: %s → %s\n", entry.Name(), filepath.Base(yamlPath))
		migrated++
	}

	// Print summary
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Printf("Migration Summary:\n")
	fmt.Printf("  ✅ Successfully migrated: %d files\n", migrated)
	fmt.Printf("  ⚠️  Skipped: %d files\n", skipped)
	fmt.Printf("  ❌ Failed: %d files\n", failed)

	if failed > 0 {
		fmt.Printf("\nFailed files:\n")
		for _, f := range failedFiles {
			fmt.Printf("  - %s\n", f)
		}
		return fmt.Errorf("migration completed with %d errors", failed)
	}

	if migrated == 0 {
		fmt.Println("\nNo JSON schema files found to migrate.")
		return nil
	}

	fmt.Println("\n✨ Migration completed successfully!")
	fmt.Println("\nNext steps:")
	fmt.Println("  1. Verify the converted YAML files: " + schemaDir)
	fmt.Println("  2. Test your application to ensure schemas work correctly")
	fmt.Println("  3. Delete old JSON files if migration is successful")
	fmt.Println("  4. Update any schema loading code if necessary")

	return nil
}

// NewMigrateJSONToYAMLCommand creates the migrate-json-to-yaml command
func NewMigrateJSONToYAMLCommand() *cli.Command {
	return &cli.Command{
		Name:  "migrate-json-to-yaml",
		Usage: "Migrate all JSON schema files to YAML format",
		Description: `Convert all JSON schema files in ./data/schemas/ to YAML format.

This command reads JSON schema files and converts them to YAML format.
Original JSON files are preserved.

Note: This command will be removed in a future release after the migration period.

For more information, see: https://github.com/fastschema/fastschema/blob/develop/SCHEMA_MIGRATION_GUIDE.md`,
		Action: migrateJSONToYAML,
	}
}

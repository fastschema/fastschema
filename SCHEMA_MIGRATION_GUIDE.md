# Schema Format Migration: JSON to YAML

## Overview

This document outlines the breaking change in FastSchema schema file format. Schema definitions have been refactored from **JSON** to **YAML**, improving readability and reducing boilerplate while maintaining complete feature parity.

## What Changed

### File Format Migration

- **Old Format:** JSON files (`.json`)
- **New Format:** YAML files (`.yaml`)

All schema definitions now use YAML format exclusively. JSON format support has been deprecated.

### Example: Side-by-Side Comparison

#### Before (JSON)
```json
{
  "name": "user",
  "namespace": "users",
  "label_field": "name",
  "db": {
    "indexes": [{
      "name": "idx_name",
      "columns": ["name"]
    }]
  },
  "fields": [
    {
      "name": "name",
      "label": "Name",
      "type": "string",
      "sortable": true,
      "filterable": true,
      "unique": true
    }
  ]
}
```

#### After (YAML)
```yaml
name: user
namespace: users
label_field: name
db:
  indexes:
  - name: idx_name
    columns:
    - name
fields:
- name: name
  label: Name
  type: string
  sortable: true
  filterable: true
  unique: true
```

## Migration Guide

### For Application Users

#### Automated Migration (Recommended)

> Note: This command will be removed in a future release after the migration period. Back up `./data/schemas/` before running and see `./fastschema migrate-json-to-yaml --help` for usage and safety options.

FastSchema provides an automated migration command to convert all JSON schema files to YAML:

```bash
./fastschema migrate-json-to-yaml <directory>
```

**Usage Examples:**

```bash
# Migrate schemas in current directory
./fastschema migrate-json-to-yaml .

# Migrate schemas in specific directory
./fastschema migrate-json-to-yaml /path/to/app

# Show help
./fastschema migrate-json-to-yaml --help
```

**What the Command Does:**
- ✅ Automatically converts all `.json` schema files to `.yaml` format
- ✅ Preserves original JSON files (safe - no files deleted)
- ✅ Validates JSON format before conversion
- ✅ Skips if YAML file already exists
- ✅ Provides detailed progress and summary report

**Example Output:**
```
✅ Migrated: user.json → user.yaml
✅ Migrated: post.json → post.yaml

============================================================
Migration Summary:
  ✅ Successfully migrated: 2 files
  ⚠️  Skipped: 1 files
  ❌ Failed: 0 files

✨ Migration completed successfully!

Next steps:
  1. Verify the converted YAML files: data/schemas
  2. Test your application to ensure schemas work correctly
  3. Delete old JSON files if migration is successful
  4. Update any schema loading code if necessary
```

#### Manual Migration (Alternative)

If you prefer to migrate schemas manually:

##### 1. Update Schema File Extensions
Rename all schema files from `.json` to `.yaml`:

```bash
# Example: migrating a single schema
mv ./schemas/user.json ./schemas/user.yaml

# Batch rename all schema files
find ./schemas -name "*.json" -exec sh -c 'mv "$1" "${1%.json}.yaml"' _ {} \;
```

##### 2. Convert JSON Content to YAML
Convert the content of each schema file from JSON to YAML format. Use online conversion tools or the following patterns:

**Key Conversion Patterns:**

| JSON | YAML |
|------|------|
| `{` | (no brace needed) |
| `"key": value` | `key: value` |
| `[` | list items with `-` prefix |
| `true/false` | `true/false` (same) |
| `null` | `null` (same) |

##### 3. Validate Your Schemas
After migration, verify that your schema files are properly formatted:

```bash
# Test your schema loading to ensure it's valid YAML
# Run your application and check schema initialization logs
```

### For Plugin Developers

If you are developing plugins with embedded schemas, follow these steps:

1. **Update Plugin Schema Files**
   - Convert all `.json` schema files to `.yaml`
   - Place them in the plugin's schema directory

2. **Update Schema Loading Code**
   - Replace `NewSchemaFromJSONFile()` calls with `NewSchemaFromYAMLFile()`
   - Update any file path references from `.json` to `.yaml`

#### Example Code Update
```go
// Before
schema, err := schema.NewSchemaFromJSONFile("./schemas/product.json")

// After
schema, err := schema.NewSchemaFromYAMLFile("./schemas/product.yaml")
```

### For Schema API Integration

#### Programmatic Schema Creation

JSON and YAML strings can still be used programmatically with the respective functions:

```go
// Creating from YAML string
schemaYAML := `
name: user
namespace: users
fields:
- name: email
  type: string
  unique: true
`
schema, err := schema.NewSchemaFromYAML(schemaYAML)

// Creating from map (works the same as before)
schemaMap := map[string]interface{}{
  "name": "user",
  "namespace": "users",
  // ... fields
}
schema, err := schema.NewSchemaFromMap(schemaMap)
```

## Breaking Changes

### Removed Functions

| Function | Alternative |
|----------|-------------|
| `NewSchemaFromJSONFile()` | `NewSchemaFromYAMLFile()` |
| Direct `.json` schema file loading | Use `.yaml` extension |

### API Endpoints

If you are using schema import/export APIs, note that:
- **Schema Export:** Now returns YAML format by default
- **Schema Import:** Expects YAML format files

Update any automation scripts that handle schema import/export operations.

## Backward Compatibility

**Important:** The JSON to YAML migration is a **breaking change**. Applications using JSON schema files must be updated before upgrading to this version.

**No automatic conversion** is performed. You must manually convert existing schemas or use conversion tools.

## Benefits of YAML Format

1. **Improved Readability:** YAML's whitespace-based syntax is more human-readable
2. **Reduced Boilerplate:** No need for quotes around keys or commas between entries
3. **Better Maintainability:** Cleaner code for complex nested structures
4. **Industry Standard:** YAML is widely used in configuration files across various frameworks
5. **Enhanced Developer Experience:** Easier to write and review schema definitions

## Migration Checklist

### Using Automated Migration Command (Recommended)
- [ ] Ensure FastSchema is built: `go build ./cmd/...`
- [ ] Run migration command: `./fastschema migrate-json-to-yaml .`
- [ ] Verify converted YAML files in `./data/schemas/`
- [ ] Test application with YAML schemas
- [ ] Review migration summary for any failures
- [ ] Delete old JSON files after verification
- [ ] Update documentation and deployment scripts

### Using Manual Migration
- [ ] Identify all schema files in your project (`.json` extension)
- [ ] Back up existing schema files
- [ ] Convert JSON schemas to YAML format
- [ ] Rename files from `.json` to `.yaml`
- [ ] Update schema loading code to use YAML functions
- [ ] Update schema import/export automation scripts
- [ ] Test schema loading and validation
- [ ] Verify all schema-dependent features (CRUD, relationships, etc.)
- [ ] Update documentation and deployment scripts

## Troubleshooting

### Migration Command Issues

#### Issue: "schemas directory not found"

**Solution:** Ensure you're running the command from the correct directory or provide the correct path:
```bash
# Verify the path exists
ls -la ./data/schemas/

# Run with explicit path
./fastschema migrate-json-to-yaml /full/path/to/app
```

#### Issue: "Failed to read JSON file"

**Solution:** Check file permissions:
```bash
# Verify read permissions
ls -la ./data/schemas/

# Fix permissions if needed
chmod 644 ./data/schemas/*.json
```

#### Issue: Migration completed with errors

**Solution:** Review the failed files list and check:
1. File format is valid JSON
2. File is not corrupted
3. Disk space is available
4. Write permissions exist in the schemas directory

### Schema Format Issues

#### Issue: "Schema file not found" or "Failed to unmarshal schema"

**Solution:** Verify that:
1. Schema file extension is `.yaml` (not `.json`)
2. YAML formatting is valid (proper indentation, no mixed tabs/spaces)
3. File paths in code reference the new `.yaml` extension

#### Issue: Invalid YAML syntax

**Solution:** Check for common YAML mistakes:
- Use spaces (not tabs) for indentation
- Ensure consistent indentation levels
- Quote string values containing special characters
- Use `-` for list items (not `[]`)

#### Issue: Schema validation fails after conversion

**Solution:** Ensure all required fields are present:
- `name` (string)
- `namespace` (string)
- `fields` (array)

Refer to the [FastSchema Documentation](https://fastschema.com) for complete schema structure requirements.

## Support

For issues or questions regarding this migration:

1. Check the [FastSchema Documentation](https://fastschema.com)
2. Review [Example Schemas](./tests/data/schemas/) in the repository
3. Open an [Issue on GitHub](https://github.com/fastschema/fastschema/issues)
4. Refer to [Contributing Guidelines](./CONTRIBUTING.md)

## Version Information

- **Breaking Change Introduced:** v1.x.x (YAML Migration)
- **Migration Required:** Yes
- **Rollback Possible:** No (permanent format change)

---

For comprehensive information on FastSchema schema structure and features, visit the [official documentation](https://fastschema.com).

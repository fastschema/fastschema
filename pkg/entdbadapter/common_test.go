package entdbadapter

import (
	"context"
	"fmt"
	"os"
	"testing"

	atlasSchema "ariga.io/atlas/sql/schema"
	entSchema "entgo.io/ent/dialect/sql/schema"
	"entgo.io/ent/schema/field"
	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/logger"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
	"github.com/stretchr/testify/assert"
)

func TestCreateEntColumn(t *testing.T) {
	type args struct {
		name   string
		field  *schema.Field
		column *entSchema.Column
	}

	tests := []args{
		{
			name: "testIDColumn",
			field: &schema.Field{
				Name:   "id",
				Type:   schema.TypeUint64,
				Unique: true,
				DB: &schema.FieldDB{
					Increment: true,
				},
			},
			column: &entSchema.Column{
				Name:      "id",
				Type:      field.TypeUint64,
				Increment: true,
				Unique:    true,
			},
		},
		{
			name: "testTextColumn",
			field: &schema.Field{
				Name: "content",
				Type: schema.TypeText,
				Size: 100,
				DB: &schema.FieldDB{
					Collation: "utf8mb4_unicode_ci",
					Key:       "MUL",
					Attr:      "UNIQUE",
				},
			},
			column: &entSchema.Column{
				Name:      "content",
				Type:      field.TypeString,
				Size:      100,
				Collation: "utf8mb4_unicode_ci",
				Key:       "MUL",
				Attr:      "UNIQUE",
			},
		},
		{
			name: "testNormalColumn",
			field: &schema.Field{
				Name:     "name",
				Type:     schema.TypeString,
				Default:  "test",
				Optional: true,
			},
			column: &entSchema.Column{
				Name:     "name",
				Type:     field.TypeString,
				Default:  "test",
				Nullable: true,
			},
		},
		{
			name: "testEnumColumn",
			field: &schema.Field{
				Name: "status",
				Type: schema.TypeEnum,
				Enums: []*schema.FieldEnum{
					{
						Label: "Active",
						Value: "active",
					},
					{
						Label: "Inactive",
						Value: "inactive",
					},
				},
			},
			column: &entSchema.Column{
				Name:  "status",
				Type:  field.TypeEnum,
				Enums: []string{"active", "inactive"},
			},
		},
		{
			name: "testTimeColumn",
			field: &schema.Field{
				Name: "created_at",
				Type: schema.TypeTime,
			},
			column: &entSchema.Column{
				Name: "created_at",
				Type: field.TypeTime,
				SchemaType: map[string]string{
					"mysql": "datetime",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			column := createEntColumn(tt.field)
			assert.Equal(t, tt.column, column)
		})
	}
}

func TestCreateDBDSN(t *testing.T) {
	config := &db.Config{
		Driver: "mysql",
		User:   "user",
		Pass:   "pass",
		Host:   "localhost",
		Port:   "3306",
		Name:   "database",
	}

	expectedMySQLDSN := "user:pass@tcp(localhost:3306)/database?charset=utf8mb4&collation=utf8mb4_unicode_ci&parseTime=true&multiStatements=true"
	assert.Equal(t, expectedMySQLDSN, CreateDBDSN(config))

	config.Driver = "pgx"
	expectedPGXDSN := "host=localhost port=3306 user=user dbname=database password=pass sslmode=disable"
	assert.Equal(t, expectedPGXDSN, CreateDBDSN(config))

	config.Driver = "sqlite"
	expectedSQLiteDSN := "file:database?cache=shared&_fk=1&_pragma=foreign_keys(1)"
	assert.Equal(t, expectedSQLiteDSN, CreateDBDSN(config))
	config.Name = ":memory:"
	expectedSQLiteMemoryDSN := "file:/fastschema_.db?vfs=memdb&_fk=1&_pragma=foreign_keys(1)"
	assert.Equal(t, expectedSQLiteMemoryDSN, CreateDBDSN(config))
}

func TestGetEntDialect(t *testing.T) {
	tests := []struct {
		name            string
		config          *db.Config
		expectedDialect string
		expectedError   error
	}{
		{
			name: "Supported driver",
			config: &db.Config{
				Driver: "mysql",
			},
			expectedDialect: "mysql",
			expectedError:   nil,
		},
		{
			name: "Unsupported driver",
			config: &db.Config{
				Driver: "mongodb",
			},
			expectedDialect: "",
			expectedError:   fmt.Errorf("unsupported driver: mongodb"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dialect, err := GetEntDialect(tt.config)
			assert.Equal(t, tt.expectedDialect, dialect)
			assert.Equal(t, tt.expectedError, err)
		})
	}
}

func TestCreateRenameColumnsHook(t *testing.T) {
	// Define sample rename tables and rename columns
	renameTables := []*db.RenameItem{
		{From: "table_1_old_name", To: "table_1_new_name"},
		{From: "table_2_old_name", To: "table_2_new_name"},
	}
	renameColumns := []*db.RenameItem{
		{From: "t1_c1_old_name", To: "t1_c1_new_name"},
		{From: "t1_c2_old_name", To: "t1_c2_new_name"},
		{From: "t3_c1_old_name", To: "t3_c1_new_name", SchemaNamespace: "table_3"},
	}

	// Create a sample current schema
	currentSchema := &atlasSchema.Schema{
		Tables: []*atlasSchema.Table{
			{
				Name: "table_1_old_name",
				Columns: []*atlasSchema.Column{
					{Name: "t1_c1_old_name"},
					{Name: "t1_c2_old_name"},
				},
			},
			{
				Name: "table_2_old_name",
				Columns: []*atlasSchema.Column{
					{Name: "t2_c1"},
				},
			},
			{
				Name: "table_3",
				Columns: []*atlasSchema.Column{
					{Name: "t3_c1_old_name"},
				},
			},
		},
	}

	// Create a sample desired schema
	desiredSchema := &atlasSchema.Schema{
		Tables: []*atlasSchema.Table{
			{
				Name: "table_1_new_name",
				Columns: []*atlasSchema.Column{
					{Name: "t1_c1_new_name"},
				},
			},
			{
				Name: "table_2_new_name",
				Columns: []*atlasSchema.Column{
					{Name: "t2_c1"},
					{Name: "t1_c2_new_name"},
					{Name: "another_new_column"},
				},
			},
			{
				Name: "table_3",
				Columns: []*atlasSchema.Column{
					{Name: "t3_c1_new_name"},
				},
			},
		},
	}

	// Create the diff hook
	diffHook := createRenameColumnsHook(renameTables, renameColumns)
	expectedChanges := []atlasSchema.Change{
		&atlasSchema.AddTable{
			T: desiredSchema.Tables[0],
		},
		&atlasSchema.RenameColumn{
			From: currentSchema.Tables[0].Columns[0],
			To:   desiredSchema.Tables[0].Columns[0],
		},
		&atlasSchema.RenameColumn{
			From: currentSchema.Tables[0].Columns[1],
			To:   desiredSchema.Tables[1].Columns[1],
		},
		&atlasSchema.ModifyTable{
			T: desiredSchema.Tables[2],
			Changes: []atlasSchema.Change{
				&atlasSchema.AddColumn{
					C: desiredSchema.Tables[2].Columns[0],
				},
			},
		},
	}

	// Create the differ with the diff hook
	var next entSchema.Differ = entSchema.DiffFunc(func(current, desired *atlasSchema.Schema) ([]atlasSchema.Change, error) {
		return expectedChanges, nil
	})

	differ := diffHook(next)

	// Calculate the diff between current and desired schema
	changes, err := differ.Diff(currentSchema, desiredSchema)
	assert.NoError(t, err)

	assert.Equal(t, expectedChanges[1:], changes)
}

func TestNOW(t *testing.T) {
	// Test for MySQL dialect
	mysqlResult := NOW("mysql")
	assert.NotNil(t, mysqlResult)
	// Add assertions for the expected MySQL result

	// Test for PostgreSQL dialect
	pgxResult := NOW("pgx")
	assert.NotNil(t, pgxResult)
	// Add assertions for the expected PostgreSQL result

	// Test for SQLite dialect
	sqliteResult := NOW("sqlite")
	assert.NotNil(t, sqliteResult)
	// Add assertions for the expected SQLite result

	// Test for unsupported dialect
	unsupportedResult := NOW("unsupported")
	assert.NotNil(t, unsupportedResult)
	// Add assertions for the expected unsupported result
}

type testContext struct {
	context.Context
	traceID string
}

func (t *testContext) TraceID() string {
	return t.traceID
}

func TestCreateDebugFN(t *testing.T) {
	mockLogger := logger.CreateMockLogger(true)
	config := &db.Config{
		LogQueries: true,
		Logger:     mockLogger,
	}

	ctx := &testContext{traceID: "12345", Context: context.Background()}
	debugFn := CreateDebugFN(config)

	debugFn(ctx, 1, 2, 3)
	assert.Contains(t, mockLogger.Last().String(), "[1 2 3]")

	ctx2 := context.WithValue(context.Background(), fs.ContextKeyTraceID, "12345")
	debugFn(ctx2, 1, 2, 3)
	assert.Contains(t, mockLogger.Last().String(), "[1 2 3]")
}

func TestCloneMigrateTableWithNewName(t *testing.T) {
	// Create a sample table
	table := &atlasSchema.Table{
		Name:        "table",
		Schema:      &atlasSchema.Schema{},
		Columns:     []*atlasSchema.Column{},
		Indexes:     []*atlasSchema.Index{},
		PrimaryKey:  &atlasSchema.Index{},
		ForeignKeys: []*atlasSchema.ForeignKey{},
		Attrs:       []atlasSchema.Attr{},
	}

	// Clone the table with a new name
	clone := cloneMigrateTableWithNewName(table, "newTable")

	// Verify that the cloned table has the same values as the original
	assert.Equal(t, "newTable", clone.Name)
	assert.Equal(t, table.Schema, clone.Schema)
	assert.Equal(t, table.Columns, clone.Columns)
	assert.Equal(t, table.Indexes, clone.Indexes)
	assert.Equal(t, table.PrimaryKey, clone.PrimaryKey)
	assert.Equal(t, table.ForeignKeys, clone.ForeignKeys)
	assert.Equal(t, table.Attrs, clone.Attrs)
}

func TestGetPlanForRenameTables(t *testing.T) {
	ctx := context.Background()
	sb := createSchemaBuilder()
	dbClient, err := NewTestClient(
		utils.Must(os.MkdirTemp("", "test")),
		sb,
	)
	assert.NoError(t, err)
	adapter := dbClient.(EntAdapter)

	migrateDriver, err := getAtlasMigrateDriver(adapter.Driver().Dialect(), adapter.DB())
	assert.NoError(t, err)

	// Rename table not found
	_, err = getPlanForRenameTables(ctx, migrateDriver, []*db.RenameItem{
		{From: "notfound", To: "notfound2"},
	})
	assert.Error(t, err)

	// Rename table found
	plan, err := getPlanForRenameTables(ctx, migrateDriver, []*db.RenameItem{
		{From: "users", To: "members"},
	})
	assert.NoError(t, err)
	assert.Len(t, plan.Changes, 1)

	// Error db connection
	assert.NoError(t, adapter.Close())
	_, err = getPlanForRenameTables(ctx, migrateDriver, []*db.RenameItem{
		{From: "users", To: "members"},
	})
	assert.Error(t, err)
}

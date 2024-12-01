package entdbadapter

import (
	"context"
	"database/sql/driver"
	"os"
	"reflect"
	"testing"

	entSchema "entgo.io/ent/dialect/sql/schema"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
	"github.com/stretchr/testify/assert"
)

func TestAdapterInit(t *testing.T) {
	config := &db.Config{Driver: "sqlmock"}
	adapter := &Adapter{
		driver:        nil,
		sqldb:         nil,
		config:        config,
		migrationDir:  config.MigrationDir,
		schemaBuilder: createSchemaBuilder(),
		models:        make([]*Model, 0),
		tables:        make([]*entSchema.Table, 0),
		edgeSpec:      make(map[string]sqlgraph.EdgeSpec),
	}

	err := adapter.init()
	assert.NoError(t, err)
	assert.Nil(t, adapter.DB())
}

func TestAdapterNewEdgeSpecError(t *testing.T) {
	adapter := createMockAdapter(t)
	relation := &schema.Relation{Name: "relation"}
	edgeSpec, err := adapter.NewEdgeSpec(relation, []driver.Value{})
	assert.Error(t, err)
	assert.Nil(t, edgeSpec)
}

func TestNewEdgeStepOptionError(t *testing.T) {
	adapter := createMockAdapter(t)
	relation := &schema.Relation{Name: "relation"}
	edgeStepOption, err := adapter.NewEdgeStepOption(relation)
	assert.Error(t, err)
	assert.Nil(t, edgeStepOption)
}

func TestModel(t *testing.T) {
	adapter := createMockAdapter(t)

	// case 1: nil model
	_, err := adapter.Model(nil)
	assert.Error(t, err)

	// case 1: model not found
	type testStruct struct{}
	_, err = adapter.Model(&testStruct{})
	assert.Error(t, err)

	// case 2: model found
	roleModel, err := adapter.Model(&fs.Role{})
	assert.NoError(t, err)
	assert.NotNil(t, roleModel)

	// case 3: using reflection
	rtype := reflect.TypeOf(fs.Role{})
	roleModel, err = adapter.Model(rtype)
	assert.NoError(t, err)
	assert.NotNil(t, roleModel)
}

func TestExec(t *testing.T) {
	ctx := context.Background()
	migrationDir := utils.Must(os.MkdirTemp("", "migrations"))
	sb := createSchemaBuilder()

	// Case 1: Pre hook error
	client, err := NewTestClient(migrationDir, sb, func() *db.Hooks {
		return &db.Hooks{
			PreDBExec: []db.PreDBExec{
				func(ctx context.Context, option *db.QueryOption) error {
					return assert.AnError
				},
			},
			PreDBQuery: []db.PreDBQuery{
				func(ctx context.Context, option *db.QueryOption) error {
					return assert.AnError
				},
			},
		}
	})
	assert.NoError(t, err)

	result, err := client.Exec(ctx, "SELECT ?, ?", 1, 2)
	assert.Error(t, err)
	assert.Nil(t, result)

	entities, err := client.Query(ctx, "SELECT ?, ?", 1, 2)
	assert.Error(t, err)
	assert.Nil(t, entities)
}

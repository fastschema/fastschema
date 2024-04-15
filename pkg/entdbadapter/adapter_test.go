package entdbadapter

import (
	"database/sql/driver"
	"testing"

	entSchema "entgo.io/ent/dialect/sql/schema"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/fastschema/fastschema/app"
	"github.com/fastschema/fastschema/schema"
	"github.com/stretchr/testify/assert"
)

func TestAdapterInit(t *testing.T) {
	config := &app.DBConfig{Driver: "sqlmock"}
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
func TestAdapterCreateDBModel(t *testing.T) {
	adapter := createMockAdapter(t)
	s := &schema.Schema{}
	relations := []*schema.Relation{{Name: "relation1"}, {Name: "relation2"}}

	model := adapter.CreateDBModel(s, relations...)

	assert.NotNil(t, model)
}

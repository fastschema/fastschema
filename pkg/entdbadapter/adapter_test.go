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
	"github.com/stretchr/testify/require"
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

func TestAdapterRelationOnDeleteOptions(t *testing.T) {
	sb := createOnDeleteSchemaBuilder(t)
	migrationDir := utils.Must(os.MkdirTemp("", "migrations"))
	t.Cleanup(func() { _ = os.RemoveAll(migrationDir) })

	client, err := NewTestClient(migrationDir, sb)
	require.NoError(t, err)
	adapter, ok := client.(*Adapter)
	require.True(t, ok)
	t.Cleanup(func() { _ = adapter.Close() })

	childModel, err := adapter.Model("child")
	require.NoError(t, err)
	entChildModel, ok := childModel.(*Model)
	require.True(t, ok)

	expectations := map[string]entSchema.ReferenceOption{
		"strict_parent_id":   entSchema.NoAction,
		"nullable_parent_id": entSchema.SetNull,
		"cascade_parent_id":  entSchema.Cascade,
	}

	for _, fk := range entChildModel.entTable.ForeignKeys {
		if len(fk.Columns) == 0 {
			continue
		}
		column := fk.Columns[0].Name
		expected, ok := expectations[column]
		if !ok {
			continue
		}
		assert.Equal(t, expected, fk.OnDelete, "unexpected on delete for column %s", column)
		delete(expectations, column)
	}

	assert.Empty(t, expectations, "missing foreign key expectations: %v", expectations)
}

func TestAdapterRelationOnUpdateOptions(t *testing.T) {
	sb := createOnUpdateSchemaBuilder(t)
	migrationDir := utils.Must(os.MkdirTemp("", "migrations"))
	t.Cleanup(func() { _ = os.RemoveAll(migrationDir) })

	client, err := NewTestClient(migrationDir, sb)
	require.NoError(t, err)
	adapter, ok := client.(*Adapter)
	require.True(t, ok)
	t.Cleanup(func() { _ = adapter.Close() })

	childModel, err := adapter.Model("child")
	require.NoError(t, err)
	entChildModel, ok := childModel.(*Model)
	require.True(t, ok)

	expectations := map[string]entSchema.ReferenceOption{
		"default_parent_id":  entSchema.NoAction,
		"setnull_parent_id":  entSchema.SetNull,
		"cascade_parent_id":  entSchema.Cascade,
		"restrict_parent_id": entSchema.Restrict,
	}

	for _, fk := range entChildModel.entTable.ForeignKeys {
		if len(fk.Columns) == 0 {
			continue
		}
		column := fk.Columns[0].Name
		expected, ok := expectations[column]
		if !ok {
			continue
		}
		assert.Equal(t, expected, fk.OnUpdate, "unexpected on update for column %s", column)
		delete(expectations, column)
	}

	assert.Empty(t, expectations, "missing foreign key expectations: %v", expectations)
}

func createOnDeleteSchemaBuilder(t *testing.T) *schema.Builder {
	t.Helper()
	parentSchema := &schema.Schema{
		Name:           "parent",
		Namespace:      "parents",
		LabelFieldName: "name",
		Fields: []*schema.Field{
			{
				Name:  "name",
				Label: "Name",
				Type:  schema.TypeString,
			},
			{
				Name:  "strict_children",
				Label: "Strict Children",
				Type:  schema.TypeRelation,
				Relation: &schema.Relation{
					Type:             schema.O2M,
					Owner:            true,
					TargetSchemaName: "child",
					TargetFieldName:  "strict_parent",
				},
			},
			{
				Name:  "nullable_children",
				Label: "Nullable Children",
				Type:  schema.TypeRelation,
				Relation: &schema.Relation{
					Type:             schema.O2M,
					Owner:            true,
					TargetSchemaName: "child",
					TargetFieldName:  "nullable_parent",
				},
			},
			{
				Name:  "cascade_children",
				Label: "Cascade Children",
				Type:  schema.TypeRelation,
				Relation: &schema.Relation{
					Type:             schema.O2M,
					Owner:            true,
					TargetSchemaName: "child",
					TargetFieldName:  "cascade_parent",
				},
			},
		},
	}

	childSchema := &schema.Schema{
		Name:           "child",
		Namespace:      "children",
		LabelFieldName: "name",
		Fields: []*schema.Field{
			{
				Name:  "name",
				Label: "Name",
				Type:  schema.TypeString,
			},
			{
				Name:  "strict_parent",
				Label: "Strict Parent",
				Type:  schema.TypeRelation,
				Relation: &schema.Relation{
					Type:             schema.O2M,
					TargetSchemaName: "parent",
					TargetFieldName:  "strict_children",
				},
			},
			{
				Name:     "nullable_parent",
				Label:    "Nullable Parent",
				Type:     schema.TypeRelation,
				Optional: true,
				Relation: &schema.Relation{
					Type:             schema.O2M,
					TargetSchemaName: "parent",
					TargetFieldName:  "nullable_children",
				},
			},
			{
				Name:  "cascade_parent",
				Label: "Cascade Parent",
				Type:  schema.TypeRelation,
				Relation: &schema.Relation{
					Type:             schema.O2M,
					TargetSchemaName: "parent",
					TargetFieldName:  "cascade_children",
					OnDelete:         schema.Cascade,
				},
			},
		},
	}

	schemas := map[string]*schema.Schema{
		parentSchema.Name: parentSchema,
		childSchema.Name:  childSchema,
	}

	sb, err := schema.NewBuilderFromSchemas("", schemas)
	require.NoError(t, err)

	return sb
}

func createOnUpdateSchemaBuilder(t *testing.T) *schema.Builder {
	t.Helper()
	parentSchema := &schema.Schema{
		Name:           "parent",
		Namespace:      "parents",
		LabelFieldName: "name",
		Fields: []*schema.Field{
			{
				Name:  "name",
				Label: "Name",
				Type:  schema.TypeString,
			},
			{
				Name:  "default_children",
				Label: "Default Children",
				Type:  schema.TypeRelation,
				Relation: &schema.Relation{
					Type:             schema.O2M,
					Owner:            true,
					TargetSchemaName: "child",
					TargetFieldName:  "default_parent",
				},
			},
			{
				Name:  "setnull_children",
				Label: "SetNull Children",
				Type:  schema.TypeRelation,
				Relation: &schema.Relation{
					Type:             schema.O2M,
					Owner:            true,
					TargetSchemaName: "child",
					TargetFieldName:  "setnull_parent",
				},
			},
			{
				Name:  "cascade_children",
				Label: "Cascade Children",
				Type:  schema.TypeRelation,
				Relation: &schema.Relation{
					Type:             schema.O2M,
					Owner:            true,
					TargetSchemaName: "child",
					TargetFieldName:  "cascade_parent",
				},
			},
			{
				Name:  "restrict_children",
				Label: "Restrict Children",
				Type:  schema.TypeRelation,
				Relation: &schema.Relation{
					Type:             schema.O2M,
					Owner:            true,
					TargetSchemaName: "child",
					TargetFieldName:  "restrict_parent",
				},
			},
		},
	}

	childSchema := &schema.Schema{
		Name:           "child",
		Namespace:      "children",
		LabelFieldName: "name",
		Fields: []*schema.Field{
			{
				Name:  "name",
				Label: "Name",
				Type:  schema.TypeString,
			},
			{
				Name:  "default_parent",
				Label: "Default Parent",
				Type:  schema.TypeRelation,
				Relation: &schema.Relation{
					Type:             schema.O2M,
					TargetSchemaName: "parent",
					TargetFieldName:  "default_children",
				},
			},
			{
				Name:     "setnull_parent",
				Label:    "SetNull Parent",
				Type:     schema.TypeRelation,
				Optional: true,
				Relation: &schema.Relation{
					Type:             schema.O2M,
					TargetSchemaName: "parent",
					TargetFieldName:  "setnull_children",
					OnUpdate:         schema.SetNull,
				},
			},
			{
				Name:  "cascade_parent",
				Label: "Cascade Parent",
				Type:  schema.TypeRelation,
				Relation: &schema.Relation{
					Type:             schema.O2M,
					TargetSchemaName: "parent",
					TargetFieldName:  "cascade_children",
					OnUpdate:         schema.Cascade,
				},
			},
			{
				Name:  "restrict_parent",
				Label: "Restrict Parent",
				Type:  schema.TypeRelation,
				Relation: &schema.Relation{
					Type:             schema.O2M,
					TargetSchemaName: "parent",
					TargetFieldName:  "restrict_children",
					OnUpdate:         schema.Restrict,
				},
			},
		},
	}

	schemas := map[string]*schema.Schema{
		parentSchema.Name: parentSchema,
		childSchema.Name:  childSchema,
	}

	sb, err := schema.NewBuilderFromSchemas("", schemas)
	require.NoError(t, err)

	return sb
}

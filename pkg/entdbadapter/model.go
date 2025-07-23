package entdbadapter

import (
	"context"
	"errors"
	"fmt"

	entSchema "entgo.io/ent/dialect/sql/schema"
	"entgo.io/ent/dialect/sql/sqlgraph"
	ef "entgo.io/ent/schema/field"
	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/entity"
	"github.com/fastschema/fastschema/schema"
)

// Column is a wrapper around a schema field and an ent column
type Column struct {
	field     *schema.Field
	entColumn *entSchema.Column
}

// Model holds the model data
type Model struct {
	name        string
	client      db.Client
	schema      *schema.Schema
	entTable    *entSchema.Table  `json:"-"`
	entIDColumn *entSchema.Column `json:"-"`
	columns     []*Column         `json:"-"`
}

func (m *Model) Clone() db.Model {
	return &Model{
		name:        m.name,
		client:      m.client,
		schema:      m.schema,
		entTable:    m.entTable,
		entIDColumn: m.entIDColumn,
		columns:     m.columns,
	}
}

func (m *Model) Name() string {
	return m.name
}

func (m *Model) GetEntTable() *entSchema.Table {
	return m.entTable
}

func (m *Model) SetClient(client db.Client) db.Model {
	m.client = client
	return m
}

// Schema returns the schema of the model
func (m *Model) Schema() *schema.Schema {
	return m.schema
}

// Column returns a column by name
func (m *Model) Column(name string) (*Column, error) {
	for _, column := range m.columns {
		if column.field.Name == name {
			return column, nil
		}
	}

	return nil, fmt.Errorf("column %s.%s not found", m.name, name)
}

// DBColumns returns the database columns
func (m *Model) DBColumns() []string {
	columns := make([]string, 0)
	for _, column := range m.columns {
		if column.entColumn != nil {
			columns = append(columns, column.entColumn.Name)
		}
	}
	return columns
}

// Query returns a new query builder for the model
func (m *Model) Query(predicates ...*db.Predicate) db.Querier {
	q := &Query{
		model:           m,
		client:          m.client,
		predicates:      predicates,
		entities:        []*entity.Entity{},
		withEdgesFields: []*schema.Field{},
	}

	q.querySpec = &sqlgraph.QuerySpec{
		Node: &sqlgraph.NodeSpec{
			Table: m.schema.Namespace,
			ID: &sqlgraph.FieldSpec{
				Type:   ef.TypeUint64,
				Column: entity.FieldID,
			},
		},
		ScanValues: func(columns []string) ([]any, error) {
			q.entities = append(q.entities, entity.New())
			return schemaScanValues(m.schema, columns)
		},
		Assign: func(columns []string, values []any) error {
			if len(q.entities) == 0 {
				return errors.New("assign called without calling ScanValues")
			}
			entity := q.entities[len(q.entities)-1]
			return schemaAssignValues(m.schema, entity, columns, values)
		},
	}

	return q
}

// Mutation returns a new mutation builder for the model
func (m *Model) Mutation() db.Mutator {
	return &Mutation{
		client:     m.client,
		autoCommit: false,
		model:      m,
	}
}

// Create creates a new entity
func (m *Model) Create(ctx context.Context, e *entity.Entity) (_ uint64, err error) {
	return m.Mutation().Create(ctx, e)
}

// CreateFromJSON creates a new entity from JSON
func (m *Model) CreateFromJSON(ctx context.Context, json string) (_ uint64, err error) {
	entity, err := entity.NewEntityFromJSON(json)

	if err != nil {
		return 0, err
	}

	return m.Mutation().Create(ctx, entity)
}

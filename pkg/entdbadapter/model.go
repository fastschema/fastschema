package entdbadapter

import (
	"context"
	"fmt"

	entSchema "entgo.io/ent/dialect/sql/schema"
	"entgo.io/ent/dialect/sql/sqlgraph"
	ef "entgo.io/ent/schema/field"
	"github.com/fastschema/fastschema/app"
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
	ctx         context.Context
	client      app.DBClient
	schema      *schema.Schema
	entTable    *entSchema.Table  `json:"-"`
	entIDColumn *entSchema.Column `json:"-"`
	columns     []*Column         `json:"-"`
}

func (m *Model) Clone() app.Model {
	return &Model{
		name:        m.name,
		ctx:         m.ctx,
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

func (m *Model) SetClient(client app.DBClient) app.Model {
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
func (m *Model) Query(predicates ...*app.Predicate) app.Query {
	q := &Query{
		model:           m,
		client:          m.client,
		predicates:      predicates,
		entities:        []*schema.Entity{},
		withEdgesFields: []*schema.Field{},
	}

	q.querySpec = &sqlgraph.QuerySpec{
		Node: &sqlgraph.NodeSpec{
			Table: m.schema.Namespace,
			ID: &sqlgraph.FieldSpec{
				Type:   ef.TypeUint64,
				Column: schema.FieldID,
			},
		},
		ScanValues: func(columns []string) ([]any, error) {
			q.entities = append(q.entities, schema.NewEntity())
			return scanValues(m.schema, columns)
		},
		Assign: func(columns []string, values []any) error {
			if len(q.entities) == 0 {
				return fmt.Errorf("assign called without calling ScanValues")
			}
			entity := q.entities[len(q.entities)-1]
			return assignValues(m.schema, entity, columns, values)
		},
	}

	return q
}

// Mutation returns a new mutation builder for the model
func (m *Model) Mutation(skipTxs ...bool) app.Mutation {
	return &Mutation{
		client: m.client,
		ctx:    m.ctx,
		skipTx: true,
		model:  m,
	}
}

// Create creates a new entity
func (m *Model) Create(e *schema.Entity) (_ uint64, err error) {
	return m.Mutation().Create(e)
}

// CreateFromJSON creates a new entity from JSON
func (m *Model) CreateFromJSON(json string) (_ uint64, err error) {
	entity, err := schema.NewEntityFromJSON(json)

	if err != nil {
		return 0, err
	}

	return m.Mutation().Create(entity)
}

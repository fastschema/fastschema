package entdbadapter

import (
	"context"
	"fmt"

	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/fastschema/fastschema/schema"
)

// Create creates a new entity in the database
func (m *Mutation) Create(ctx context.Context, e *schema.Entity) (_ uint64, err error) {
	if m.model == nil || m.model.schema == nil {
		return 0, fmt.Errorf("model or schema %s not found", m.model.name)
	}

	// Prevent creating an entity with an existing ID
	if e.ID() != 0 {
		return 0, fmt.Errorf("cannot create entity with existing ID %d", e.ID())
	}

	if err := runPreDBCreateHooks(ctx, m.client, m.model.schema, e); err != nil {
		return 0, err
	}

	createSpec := &sqlgraph.CreateSpec{
		Table: m.model.schema.Namespace,
		ID: &sqlgraph.FieldSpec{
			Column: m.model.entIDColumn.Name,
			Type:   field.TypeUint64,
		},
		Fields: []*sqlgraph.FieldSpec{},
		Edges:  []*sqlgraph.EdgeSpec{},
	}

	entAdapter, ok := m.client.(EntAdapter)
	if !ok {
		return 0, fmt.Errorf("client is not an ent adapter")
	}

	var c *Column
	for pair := e.First(); pair != nil; pair = pair.Next() {
		fieldName := pair.Key
		fieldValue := pair.Value

		c, err = m.model.Column(fieldName)
		if err != nil {
			return 0, fmt.Errorf("column error: %w", err)
		}

		// Non-relation fields
		if !c.field.Type.IsRelationType() {
			createSpec.Fields = append(createSpec.Fields, &sqlgraph.FieldSpec{
				Column: c.entColumn.Name,
				Value:  fieldValue,
				Type:   c.entColumn.Type,
			})
			continue
		}

		// Relation fields
		relationEntityIDs, err := m.GetRelationEntityIDs(c.field.Name, fieldValue)
		if err != nil {
			return 0, err
		}

		if len(relationEntityIDs) > 0 {
			edge, err := entAdapter.NewEdgeSpec(c.field.Relation, relationEntityIDs)
			if err != nil {
				return 0, fmt.Errorf("edge error %s.%s: %w", m.model.name, fieldName, err)
			}

			createSpec.Edges = append(createSpec.Edges, edge)
		}
	}

	if err = sqlgraph.CreateNode(ctx, entAdapter.Driver(), createSpec); err != nil {
		return 0, err
	}

	e.SetID(createSpec.ID.Value)

	if m.autoCommit {
		if err = m.client.Commit(); err != nil {
			return 0, err
		}
	}

	if err := runPostDBCreateHooks(ctx, m.client, m.model.schema, e, e.ID()); err != nil {
		return 0, err
	}

	return e.ID(), nil
}

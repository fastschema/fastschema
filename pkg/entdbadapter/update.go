package entdbadapter

import (
	"database/sql/driver"
	"fmt"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
)

// Update updates the entity and returns the updated entity
func (m *Mutation) Update(e *schema.Entity) (affected int, err error) {
	if m.model == nil || m.model.schema == nil {
		return 0, fmt.Errorf("model or schema %s not found", m.model.name)
	}

	entAdapter, ok := m.client.(EntAdapter)
	if !ok {
		return 0, fmt.Errorf("client is not an ent adapter")
	}

	m.updateSpec = &sqlgraph.UpdateSpec{
		Node: &sqlgraph.NodeSpec{
			Table: m.model.schema.Namespace,
			// Columns: []string{},
			ID: &sqlgraph.FieldSpec{
				Column: m.model.entIDColumn.Name,
				Type:   m.model.entIDColumn.Type,
			},
		},
		Fields: sqlgraph.FieldMut{
			Set:   []*sqlgraph.FieldSpec{},
			Add:   []*sqlgraph.FieldSpec{},
			Clear: []*sqlgraph.FieldSpec{},
		},
		Edges: sqlgraph.EdgeMut{
			Add:   []*sqlgraph.EdgeSpec{},
			Clear: []*sqlgraph.EdgeSpec{},
		},
	}

	if len(m.predicates) > 0 {
		sqlPredicatesFn, err := createEntPredicates(entAdapter, m.model, m.predicates)
		if err != nil {
			return 0, err
		}
		m.updateSpec.Predicate = func(s *sql.Selector) {
			s.Where(sql.And(sqlPredicatesFn(s)...))
		}
	}

	for pair := e.First(); pair != nil; pair = pair.Next() {
		switch pair.Key {
		case "$add":
			if err := m.ProcessUpdateBlockAdd(entAdapter, pair.Value); err != nil {
				return 0, err
			}
			continue
		case "$clear":
			if err := m.ProcessUpdateBlockClear(entAdapter, pair.Value); err != nil {
				return 0, err
			}
			continue
		case "$expr":
			if err := m.ProcessUpdateBlockExpr(entAdapter, pair.Value); err != nil {
				return 0, err
			}
			continue
		case "$set":
			if err := m.ProcessUpdateBlockSet(entAdapter, pair.Value); err != nil {
				return 0, err
			}
			continue
		}

		if err := m.ProcessUpdateFieldSet(entAdapter, pair.Key, pair.Value); err != nil {
			return 0, err
		}
	}

	hasFieldsUpdate := len(m.updateSpec.Fields.Set) > 0 ||
		len(m.updateSpec.Fields.Add) > 0 ||
		len(m.updateSpec.Fields.Clear) > 0 ||
		len(m.updateSpec.Modifiers) > 0 ||
		m.shouldUpdateTimestamps
	if hasFieldsUpdate && !m.model.schema.DisableTimestamp {
		m.updateSpec.Modifiers = append(m.updateSpec.Modifiers, func(u *sql.UpdateBuilder) {
			u.Set(schema.FieldUpdatedAt, NOW(m.client.Dialect()))
		})
	}

	if affected, err = sqlgraph.UpdateNodes(m.ctx, entAdapter.Driver(), m.updateSpec); err != nil {
		return 0, err
	}

	if !m.skipTx {
		if err = m.client.Commit(); err != nil {
			return 0, err
		}
	}

	return affected, nil
}

// ProcessUpdateBlockExpr processes the $expr block
func (m *Mutation) ProcessUpdateBlockExpr(entAdapter EntAdapter, fieldValue any) error {
	if expr, ok := fieldValue.(*schema.Entity); ok {
		for pair := expr.First(); pair != nil; pair = pair.Next() {
			if err := m.ProcessFieldExpr(entAdapter, pair.Key, pair.Value); err != nil {
				return err
			}
		}
	}

	return nil
}

// ProcessFieldExpr processes a field in the $expr block
func (m *Mutation) ProcessFieldExpr(entAdapter EntAdapter, k string, v any) error {
	m.updateSpec.Modifiers = append(m.updateSpec.Modifiers, func(u *sql.UpdateBuilder) {
		u.Set(k, sql.Expr(v.(string)))
	})
	return nil
}

// ProcessUpdateBlockAdd processes the $add block
func (m *Mutation) ProcessUpdateBlockAdd(entAdapter EntAdapter, fieldValue any) error {
	if expr, ok := fieldValue.(*schema.Entity); ok {
		for pair := expr.First(); pair != nil; pair = pair.Next() {
			if err := m.ProcessFieldAdd(entAdapter, pair.Key, pair.Value); err != nil {
				return err
			}
		}
	}

	return nil
}

// ProcessFieldAdd processes a field in the $add block
func (m *Mutation) ProcessFieldAdd(entAdapter EntAdapter, k string, v any) error {
	c, err := m.model.Column(k)

	if err != nil {
		return fmt.Errorf("field $add.%s error: %w", k, err)
	}

	relation := c.field.Relation

	if relation == nil {
		if utils.IsNumber(v) {
			m.updateSpec.Fields.Add = append(m.updateSpec.Fields.Add, &sqlgraph.FieldSpec{
				Column: c.entColumn.Name,
				Type:   c.entColumn.Type,
				Value:  v,
			})
		} else {
			return fmt.Errorf("field %s=%v is not a number", k, v)
		}
	} else {
		entitiesID, err := m.GetRelationEntityIDs(k, v)
		if err != nil {
			return err
		}

		// if relation is not m2m and is not owner, update timestamps
		if !relation.Type.IsM2M() && !relation.Owner {
			m.shouldUpdateTimestamps = true
		}

		addSpec, err := entAdapter.NewEdgeSpec(relation, entitiesID)
		if err != nil {
			return fmt.Errorf("field $add.%s error: %w", k, err)
		}

		m.updateSpec.Edges.Add = append(m.updateSpec.Edges.Add, addSpec)
	}

	return nil
}

// ProcessUpdateBlockClear processes the $clear block
func (m *Mutation) ProcessUpdateBlockClear(entAdapter EntAdapter, fieldValue any) error {
	if expr, ok := fieldValue.(*schema.Entity); ok {
		for pair := expr.First(); pair != nil; pair = pair.Next() {
			if err := m.ProcessFieldClear(entAdapter, pair.Key, pair.Value); err != nil {
				return err
			}
		}
	}

	return nil
}

// ProcessFieldClear processes a field in the $clear block
func (m *Mutation) ProcessFieldClear(entAdapter EntAdapter, k string, v any) error {
	c, err := m.model.Column(k)
	if err != nil {
		return fmt.Errorf("field $clear.%s error: %w", k, err)
	}

	relation := c.field.Relation

	if relation == nil {
		m.updateSpec.Fields.Clear = append(m.updateSpec.Fields.Clear, &sqlgraph.FieldSpec{
			Column: c.entColumn.Name,
			Type:   c.entColumn.Type,
		})
	} else {
		var entitiesID []driver.Value
		var err error
		if valueBool, ok := v.(bool); ok {
			if !valueBool {
				return fmt.Errorf("field $clear.%s: if boolean is used, it must be true", k)
			}
		} else {
			entitiesID, err = m.GetRelationEntityIDs(k, v)
			if err != nil {
				return err
			}
		}

		// if relation is not m2m and is not owner, update timestamps
		if !relation.Type.IsM2M() && !relation.Owner {
			m.shouldUpdateTimestamps = true
		}

		clearSpec, err := entAdapter.NewEdgeSpec(relation, entitiesID)
		if err != nil {
			return fmt.Errorf("field $clear.%s error: %w", k, err)
		}

		m.updateSpec.Edges.Clear = append(m.updateSpec.Edges.Clear, clearSpec)
	}

	return nil
}

// ProcessUpdateBlockSet processes the $set block
func (m *Mutation) ProcessUpdateBlockSet(entAdapter EntAdapter, fieldValue any) error {
	if expr, ok := fieldValue.(*schema.Entity); ok {
		for pair := expr.First(); pair != nil; pair = pair.Next() {
			if err := m.ProcessUpdateFieldSet(entAdapter, pair.Key, pair.Value); err != nil {
				return err
			}
		}
	}

	return nil
}

// ProcessUpdateFieldSet processes a field in the $set block
func (m *Mutation) ProcessUpdateFieldSet(entAdapter EntAdapter, k string, v any) error {
	c, err := m.model.Column(k)
	if err != nil {
		return fmt.Errorf("field $set.%s error: %w", k, err)
	}

	relation := c.field.Relation

	if relation == nil {
		m.updateSpec.Fields.Set = append(m.updateSpec.Fields.Set, &sqlgraph.FieldSpec{
			Column: c.entColumn.Name,
			Type:   c.entColumn.Type,
			Value:  v,
		})
	} else {
		entityIDs, err := m.GetRelationEntityIDs(k, v)
		if err != nil {
			return err
		}

		if len(entityIDs) == 0 {
			return nil
		}

		// if relation is not m2m and is not owner, update timestamps
		if !relation.Type.IsM2M() && !relation.Owner {
			m.shouldUpdateTimestamps = true
		}

		// currently, ent does not support setting edges.
		// so we need to clear and add the edges
		clearSpec, err := entAdapter.NewEdgeSpec(relation, nil)
		if err != nil {
			return fmt.Errorf("field $set.%s clearSpec error: %w", k, err)
		}

		addSpec, err := entAdapter.NewEdgeSpec(relation, entityIDs)
		if err != nil {
			return fmt.Errorf("field $set.%s addSpec error: %w", k, err)
		}

		m.updateSpec.Edges.Clear = append(m.updateSpec.Edges.Clear, clearSpec)
		m.updateSpec.Edges.Add = append(m.updateSpec.Edges.Add, addSpec)
	}

	return nil
}

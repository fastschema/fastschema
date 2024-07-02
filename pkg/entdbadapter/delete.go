package entdbadapter

import (
	"context"
	"fmt"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/schema"
)

// Delete deletes entities from the database
func (m *Mutation) Delete(ctx context.Context) (affected int, err error) {
	var hooks = &db.Hooks{}
	var originalEntities []*schema.Entity
	if m.client != nil {
		hooks = m.client.Hooks()
	}
	deleteSpec := &sqlgraph.DeleteSpec{
		Node: &sqlgraph.NodeSpec{
			Table: m.model.schema.Namespace,
			ID: &sqlgraph.FieldSpec{
				Column: m.model.entIDColumn.Name,
				Type:   m.model.entIDColumn.Type,
			},
		},
	}

	entAdapter, ok := m.client.(EntAdapter)
	if !ok {
		return 0, fmt.Errorf("client is not an ent adapter")
	}

	if len(m.predicates) > 0 {
		sqlPredicatesFn, err := createEntPredicates(entAdapter, m.model, m.predicates)
		if err != nil {
			return 0, err
		}
		deleteSpec.Predicate = func(s *sql.Selector) {
			s.Where(sql.And(sqlPredicatesFn(s)...))
		}

		if len(hooks.PostDBDelete) > 0 {
			originalEntities, err = m.model.Query(m.predicates...).Get(ctx)
			if err != nil {
				return 0, err
			}
		}
	}

	affected, err = sqlgraph.DeleteNodes(ctx, entAdapter.Driver(), deleteSpec)
	if err != nil {
		return 0, fmt.Errorf("delete nodes: %w", err)
	}

	if len(originalEntities) > 0 && len(hooks.PostDBDelete) > 0 {
		for _, hook := range hooks.PostDBDelete {
			if err := hook(m.model.schema, m.predicates, originalEntities, affected); err != nil {
				return 0, fmt.Errorf("post delete hook: %w", err)
			}
		}
	}

	return affected, nil
}

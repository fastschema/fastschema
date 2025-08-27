package entdbadapter

import (
	"context"
	"errors"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/entity"
)

// Delete deletes entities from the database
func (m *Mutation) Delete(ctx context.Context) (affected int, err error) {
	var hooks = &db.Hooks{}
	var originalEntities []*entity.Entity
	if m.client != nil {
		hooks = m.client.Hooks()
		if len(hooks.PostDBDelete) > 0 {
			originalEntities, err = m.model.Query(*m.predicates...).Get(ctx)
			if err != nil {
				return 0, err
			}
		}
	}

	if err := runPreDBDeleteHooks(ctx, m.client, m.model.schema, m.predicates); err != nil {
		return 0, err
	}

	runPostDBDeleteHooks := func() error {
		return runPostDBDeleteHooks(
			ctx,
			m.client,
			m.model.schema,
			m.predicates,
			originalEntities,
			affected,
		)
	}

	if m.client != nil && m.client.Config().UseSoftDeletes {
		if affected, err = m.model.
			Mutation().
			Where(*m.predicates...).
			Update(ctx, entity.New().Set("deleted_at", time.Now())); err != nil {
			return 0, fmt.Errorf("soft delete error: %w", err)
		}

		return affected, runPostDBDeleteHooks()
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
		return 0, errors.New("client is not an ent adapter")
	}

	if len(*m.predicates) > 0 {
		sqlPredicatesFn, err := createEntPredicates(entAdapter, m.model, *m.predicates)
		if err != nil {
			return 0, err
		}
		deleteSpec.Predicate = func(s *sql.Selector) {
			s.Where(sql.And(sqlPredicatesFn(s)...))
		}
	}

	affected, err = sqlgraph.DeleteNodes(ctx, entAdapter.Driver(), deleteSpec)
	if err != nil {
		return 0, fmt.Errorf("delete nodes error: %w", err)
	}

	return affected, runPostDBDeleteHooks()
}

package entdbadapter

import (
	"fmt"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
)

// Delete deletes entities from the database
func (m *Mutation) Delete() (affected int, err error) {
	deleteSpec := &sqlgraph.DeleteSpec{
		Node: &sqlgraph.NodeSpec{
			Table: m.model.schema.Namespace,
			ID: &sqlgraph.FieldSpec{
				Column: m.model.entIDColumn.Name,
				Type:   m.model.entIDColumn.Type,
			},
		},
	}

	entAdapter, ok := m.client.(*Adapter)
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
	}

	return sqlgraph.DeleteNodes(m.ctx, entAdapter.Driver(), deleteSpec)
}

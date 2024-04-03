package entdbadapter

import (
	"context"
	"database/sql/driver"
	"fmt"

	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/schema"
)

// Mutation holds the entity mutation data
type Mutation struct {
	ctx    context.Context
	skipTx bool
	client db.Client
	// client     *Adapter
	model                  *Model
	updateSpec             *sqlgraph.UpdateSpec
	predicates             []*db.Predicate
	shouldUpdateTimestamps bool
}

// Where adds a predicate to the mutation
func (m *Mutation) Where(predicates ...*db.Predicate) db.Mutation {
	m.predicates = append(m.predicates, predicates...)
	return m
}

// GetRelationEntityIDs return the relation IDs from the given field value
func (m *Mutation) GetRelationEntityIDs(fieldName string, fieldValue any) ([]driver.Value, error) {
	if fieldValue == nil {
		return nil, nil
	}

	relationEntities := make([]*schema.Entity, 0)
	relationEntity, ok := fieldValue.(*schema.Entity)
	if ok {
		relationEntities = append(relationEntities, relationEntity)
	} else {
		relationEntities, ok = fieldValue.([]*schema.Entity)
		if !ok {
			return nil, fmt.Errorf(
				"relation value for %s.%s is invalid",
				m.model.name,
				fieldName,
			)
		}
	}

	relationEntityIDs := make([]driver.Value, 0)
	for _, relationEntity := range relationEntities {
		if relationEntity == nil || relationEntity.Empty() {
			continue
		}

		// Create the relation entity if it doesn't exist
		// if relationEntity.ID() == 0 {
		// 	relationModel := m.client.Model(relation.TargetSchemaName)
		// 	relationEntity, err = relationModel.Mutation(true).Create(relationEntity)
		// 	if err != nil {
		// 		return nil, err
		// 	}
		// }

		if relationEntity.ID() == 0 {
			return nil, fmt.Errorf(
				"relation entity for %s.%s has no ID",
				m.model.name,
				fieldName,
			)
		}

		// Add the relation entity id to the list of ids
		relationEntityIDs = append(relationEntityIDs, relationEntity.ID())
	}

	return relationEntityIDs, nil
}

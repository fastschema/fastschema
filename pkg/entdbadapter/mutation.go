package entdbadapter

import (
	"database/sql/driver"
	"fmt"

	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/entity"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
)

// Mutation holds the entity mutation data
type Mutation struct {
	autoCommit             bool
	client                 db.Client
	model                  *Model
	updateSpec             *sqlgraph.UpdateSpec
	predicates             *[]*db.Predicate
	shouldUpdateTimestamps bool
}

// Where adds a predicate to the mutation
func (m *Mutation) Where(predicates ...*db.Predicate) db.Mutator {
	*m.predicates = append(*m.predicates, predicates...)
	return m
}

// GetRelationEntityIDs return the relation IDs from the given field value
func (m *Mutation) GetRelationEntityIDs(fieldName string, fieldValue any) ([]driver.Value, error) {
	if fieldValue == nil {
		return nil, nil
	}

	var relation *schema.Relation
	if m.model != nil && m.model.schema != nil {
		if field := m.model.schema.Field(fieldName); field != nil {
			relation = field.Relation
		}
	}

	relationEntities := make([]*entity.Entity, 0)
	relationEntity, ok := fieldValue.(*entity.Entity)
	if ok {
		relationEntities = append(relationEntities, relationEntity)
	} else {
		relationEntities, ok = fieldValue.([]*entity.Entity)
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

		value, err := m.targetReferenceValue(fieldName, relation, relationEntity)
		if err != nil {
			return nil, err
		}

		relationEntityIDs = append(relationEntityIDs, value)
	}

	return relationEntityIDs, nil
}

func (m *Mutation) targetReferenceValue(
	fieldName string,
	relation *schema.Relation,
	relationEntity *entity.Entity,
) (driver.Value, error) {
	if relation == nil {
		return nil, fmt.Errorf("relation for %s.%s not found", m.model.name, fieldName)
	}

	if relation.Type.IsM2M() {
		if relationEntity.ID() == 0 {
			return nil, fmt.Errorf("relation entity for %s.%s has no ID", m.model.name, fieldName)
		}

		return relationEntity.ID(), nil
	}

	refColumn := utils.If(relation.TargetColumn != "", relation.TargetColumn, entity.FieldID)
	value, err := relationEntity.GetUint64(refColumn, false)
	if err != nil || value == 0 {
		return nil, fmt.Errorf(
			"relation entity for %s.%s target column '%s' is invalid, value=%d, err=%w",
			m.model.name,
			fieldName,
			refColumn,
			value,
			err,
		)
	}

	return value, nil
}

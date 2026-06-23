package entdbadapter

import (
	"fmt"
	"strings"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
)

type PredicateFN func(*sql.Selector) *sql.Predicate

// parseFieldPath parses a field path with dot notation (e.g., "teams.slug")
// and returns the relation field names and the final field name.
// If the field path does not contain a dot, it returns nil for relation field names.
// Example:
//   - "teams.slug" -> (["teams"], "slug")
//   - "teams.project.name" -> (["teams", "project"], "name")
//   - "name" -> (nil, "name")
func parseFieldPath(field string) (relationFieldNames []string, fieldName string) {
	if !strings.Contains(field, ".") {
		return nil, field
	}

	parts := strings.Split(field, ".")
	if len(parts) < 2 {
		return nil, field
	}

	// Filter out empty parts (handles cases like ".field" or "field.")
	filteredParts := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			filteredParts = append(filteredParts, trimmed)
		}
	}

	if len(filteredParts) < 2 {
		return nil, field
	}

	return filteredParts[:len(filteredParts)-1], filteredParts[len(filteredParts)-1]
}

// createEntPredicates creates ent sql predicates from the given predicates
func createEntPredicates(
	entAdapter EntAdapter,
	model *Model,
	predicates []*db.Predicate,
) (func(*sql.Selector) []*sql.Predicate, error) {
	var predicateFns = []PredicateFN{}

	for _, p := range predicates {
		if p == nil {
			continue
		}

		// Parse dot notation in field name to extract relation field names
		// This allows using db.EQ("teams.slug", value) for relation filtering
		relationFields, fieldName := parseFieldPath(p.Field)

		// Check if this is a relation field without dot notation (implicit PK filter)
		// E.g. db.EQ("tags", 1) should be transformed to db.EQ("tags.id", 1)
		if len(relationFields) == 0 && p.Field != "" {
			field := model.schema.Field(p.Field)
			if field != nil && field.Type.IsRelationType() {
				targetSchema, err := entAdapter.SchemaBuilder().Schema(field.Relation.TargetSchemaName)
				if err != nil {
					return nil, fmt.Errorf("invalid relation schema %s: %w", field.Relation.TargetSchemaName, err)
				}

				pkFieldName := targetSchema.PrimaryKeyName()
				if pkFieldName == "" {
					return nil, fmt.Errorf("target schema %s has no primary key field", field.Relation.TargetSchemaName)
				}

				// Transform: "tags" -> relation filter on "tags" with PK field
				relationFields = []string{p.Field}
				fieldName = pkFieldName
			}
		}

		if len(relationFields) > 0 {
			relationPredicateFn, err := createRelationsPredicate(
				entAdapter,
				model,
				&db.Predicate{Field: fieldName, Operator: p.Operator, Value: p.Value},
				relationFields...,
			)

			if err != nil {
				return nil, err
			}

			predicateFns = append(predicateFns, relationPredicateFn)
			continue
		}

		if p.Field != "" {
			predicateFn, err := CreateFieldPredicate(p)
			if err != nil {
				return nil, err
			}

			predicateFns = append(predicateFns, predicateFn)
			continue
		}

		if p.And != nil {
			andPredicatesFn, err := createEntPredicates(entAdapter, model, p.And)
			if err != nil {
				return nil, err
			}

			predicateFns = append(predicateFns, func(s *sql.Selector) *sql.Predicate {
				return sql.And(andPredicatesFn(s)...)
			})
			continue
		}

		if p.Or != nil {
			orPredicatesFn, err := createEntPredicates(entAdapter, model, p.Or)
			if err != nil {
				return nil, err
			}

			predicateFns = append(predicateFns, func(s *sql.Selector) *sql.Predicate {
				return sql.Or(orPredicatesFn(s)...)
			})
			continue
		}
	}

	return func(s *sql.Selector) []*sql.Predicate {
		var entPredicates = []*sql.Predicate{}
		for _, predicateFn := range predicateFns {
			entPredicates = append(entPredicates, predicateFn(s))
		}

		return entPredicates
	}, nil
}

// createRelationsPredicate creates the relation predicate
func createRelationsPredicate(
	entAdapter EntAdapter,
	model *Model,
	lastFieldPredicate *db.Predicate,
	relationFieldNames ...string,
) (PredicateFN, error) {
	relationFieldName := relationFieldNames[0]
	relationFieldNames = relationFieldNames[1:]
	hasNestedRelations := len(relationFieldNames) > 0
	relationField := model.schema.Field(relationFieldName)

	if relationField == nil {
		return nil, schema.ErrFieldNotFound(model.schema.Name, relationFieldName)
	}

	relation := relationField.Relation

	targetModel, err := model.client.Model(relation.TargetSchemaName)
	if err != nil {
		return nil, err
	}

	entTargetModel, ok := targetModel.(*Model)
	if !ok {
		return nil, fmt.Errorf("model %s is not an ent model", targetModel.Schema().Name)
	}

	stepOption, err := entAdapter.NewEdgeStepOption(relation)
	if err != nil {
		return nil, fmt.Errorf("invalid edge step option '%s': %w", relationFieldName, err)
	}

	fromColumn := model.entPrimaryColumn.Name
	toColumn := entTargetModel.entPrimaryColumn.Name

	if column := relationStepFromColumn(model, relation); column != "" {
		fromColumn = column
	}

	if column := relationStepToColumn(entTargetModel, relation); column != "" {
		toColumn = column
	}

	relationStep := sqlgraph.NewStep(
		sqlgraph.From(model.schema.Namespace, fromColumn),
		sqlgraph.To(entTargetModel.schema.Namespace, toColumn),
		stepOption,
	)

	var pred func(*sql.Selector)
	useNegatedExists := false
	if hasNestedRelations {
		p, err := createRelationsPredicate(
			entAdapter,
			entTargetModel,
			lastFieldPredicate,
			relationFieldNames...,
		)

		if err != nil {
			return nil, err
		}

		pred = func(s2 *sql.Selector) {
			s2.Where(p(s2))
		}
	} else {
		useNegatedExists = relationOperatorNeedsNegation(lastFieldPredicate.Operator)
		targetPredicate := lastFieldPredicate
		if useNegatedExists {
			if positiveOperator, ok := inverseRelationOperator(lastFieldPredicate.Operator); ok {
				predicateCopy := *lastFieldPredicate
				predicateCopy.Operator = positiveOperator
				targetPredicate = &predicateCopy
			}
		}

		predFn, err := createEntPredicates(entAdapter, model, []*db.Predicate{targetPredicate})
		if err != nil {
			return nil, err
		}

		pred = func(s2 *sql.Selector) {
			s2.Where(sql.And(predFn(s2)...))
		}
	}

	return func(selector *sql.Selector) *sql.Predicate {
		s1 := selector.Clone().SetP(nil)
		sqlgraph.HasNeighborsWith(s1, relationStep, pred)
		predicate := s1.P()
		if !hasNestedRelations && useNegatedExists {
			return sql.Not(predicate)
		}
		return predicate
	}, nil
}

// =============================================================================
// Predicate Helper Functions
// =============================================================================

// columnWrap wraps a column name with selector context if available.
func columnWrap(field string, selectors ...*sql.Selector) string {
	if len(selectors) > 0 {
		return selectors[0].C(field)
	}
	return field
}

// validateStringValue validates that the predicate value is a string and returns it.
// Returns an error with field context if the value is not a string.
func validateStringValue(predicate *db.Predicate) (string, error) {
	stringValue, ok := predicate.Value.(string)
	if !ok {
		return "", fmt.Errorf(
			"value of field %s.%s = %v (%T) must be string",
			predicate.Field,
			predicate.Operator,
			predicate.Value,
			predicate.Value,
		)
	}
	return stringValue, nil
}

// validateArrayValue validates that the predicate value is an array and returns it.
// Returns an error with field context if the value is not an array.
func validateArrayValue(predicate *db.Predicate) ([]any, error) {
	arrayValue, ok := predicate.Value.([]any)
	if !ok {
		return nil, fmt.Errorf(
			"value of field %s.%s = %v (%T) must be an array",
			predicate.Field,
			predicate.Operator,
			predicate.Value,
			predicate.Value,
		)
	}
	return arrayValue, nil
}

// simplePredicateBuilder is a function type for simple comparison predicates.
type simplePredicateBuilder func(column string, value any) *sql.Predicate

// simplePredicateMap maps operators to their simple predicate builders.
var simplePredicateMap = map[db.OperatorType]simplePredicateBuilder{
	db.OpEQ:  sql.EQ,
	db.OpNEQ: sql.NEQ,
	db.OpGT:  sql.GT,
	db.OpGTE: sql.GTE,
	db.OpLT:  sql.LT,
	db.OpLTE: sql.LTE,
}

// =============================================================================
// CreateFieldPredicate
// =============================================================================

// CreateFieldPredicate convert a predicate to ent predicate
func CreateFieldPredicate(predicate *db.Predicate) (PredicateFN, error) {
	// Check for simple comparison operators first
	if builder, ok := simplePredicateMap[predicate.Operator]; ok {
		return func(s *sql.Selector) *sql.Predicate {
			return builder(columnWrap(predicate.Field, s), predicate.Value)
		}, nil
	}

	switch predicate.Operator {
	case db.OpLike:
		stringValue, err := validateStringValue(predicate)
		if err != nil {
			return nil, err
		}
		return func(s *sql.Selector) *sql.Predicate {
			return sql.Like(columnWrap(predicate.Field, s), stringValue)
		}, nil

	case db.OpNotLike:
		stringValue, err := validateStringValue(predicate)
		if err != nil {
			return nil, err
		}
		return func(s *sql.Selector) *sql.Predicate {
			return sql.Not(sql.Like(columnWrap(predicate.Field, s), stringValue))
		}, nil

	case db.OpContains:
		stringValue, err := validateStringValue(predicate)
		if err != nil {
			return nil, err
		}
		return func(s *sql.Selector) *sql.Predicate {
			return sql.Contains(columnWrap(predicate.Field, s), stringValue)
		}, nil

	case db.OpNotContains:
		stringValue, err := validateStringValue(predicate)
		if err != nil {
			return nil, err
		}
		return func(s *sql.Selector) *sql.Predicate {
			return sql.Not(sql.Contains(columnWrap(predicate.Field, s), stringValue))
		}, nil

	case db.OpContainsFold:
		stringValue, err := validateStringValue(predicate)
		if err != nil {
			return nil, err
		}
		return func(s *sql.Selector) *sql.Predicate {
			return sql.ContainsFold(columnWrap(predicate.Field, s), stringValue)
		}, nil

	case db.OpNotContainsFold:
		stringValue, err := validateStringValue(predicate)
		if err != nil {
			return nil, err
		}
		return func(s *sql.Selector) *sql.Predicate {
			return sql.Not(sql.ContainsFold(columnWrap(predicate.Field, s), stringValue))
		}, nil

	case db.OpIN, db.OpNIN:
		arrayValue, err := validateArrayValue(predicate)
		if err != nil {
			return nil, err
		}
		return func(s *sql.Selector) *sql.Predicate {
			op := utils.If(predicate.Operator == db.OpIN, sql.In, sql.NotIn)
			return op(columnWrap(predicate.Field, s), arrayValue...)
		}, nil

	case db.OpNULL:
		return func(s *sql.Selector) *sql.Predicate {
			op := utils.If(predicate.Value == true, sql.IsNull, sql.NotNull)
			return op(columnWrap(predicate.Field, s))
		}, nil

	default:
		return nil, fmt.Errorf("operator %s not supported", predicate.Operator)
	}
}

func relationStepFromColumn(model *Model, relation *schema.Relation) string {
	if relation == nil || relation.Type.IsM2M() || !relation.Owner || relation.BackRef == nil || model == nil {
		return ""
	}

	targetColumn := relation.BackRef.TargetColumn
	if targetColumn == "" || targetColumn == model.entPrimaryColumn.Name {
		return ""
	}

	return targetColumn
}

func relationStepToColumn(targetModel *Model, relation *schema.Relation) string {
	if relation == nil || relation.Type.IsM2M() || relation.Owner || targetModel == nil {
		return ""
	}

	targetColumn := relation.TargetColumn
	if targetColumn == "" || targetColumn == targetModel.entPrimaryColumn.Name {
		return ""
	}

	return targetColumn
}

func relationOperatorNeedsNegation(operator db.OperatorType) bool {
	_, ok := inverseRelationOperator(operator)
	return ok
}

func inverseRelationOperator(operator db.OperatorType) (db.OperatorType, bool) {
	switch operator {
	case db.OpNEQ:
		return db.OpEQ, true
	case db.OpNIN:
		return db.OpIN, true
	case db.OpNotLike:
		return db.OpLike, true
	case db.OpNotContains:
		return db.OpContains, true
	case db.OpNotContainsFold:
		return db.OpContainsFold, true
	default:
		return operator, false
	}
}

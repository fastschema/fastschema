package db

import (
	"errors"
	"fmt"
	"strings"

	"github.com/fastschema/fastschema/entity"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
)

// Predicate represents a filter condition for database queries.
// The Field can be a simple field name (e.g., "name") or a dot notation path
// for relation fields (e.g., "teams.slug" where "teams" is the relation field
// and "slug" is the target field in the related schema).
type Predicate struct {
	Field    string       `json:"field"`
	Operator OperatorType `json:"operator"`
	Value    any          `json:"value"`
	And      []*Predicate `json:"and,omitempty"`
	Or       []*Predicate `json:"or,omitempty"`
}

func (p *Predicate) Clone() *Predicate {
	cloned := &Predicate{
		Field:    p.Field,
		Operator: p.Operator,
		Value:    p.Value,
	}

	if p.And != nil {
		cloned.And = utils.Map(p.And, func(ap *Predicate) *Predicate {
			return ap.Clone()
		})
	}

	if p.Or != nil {
		cloned.Or = utils.Map(p.Or, func(ap *Predicate) *Predicate {
			return ap.Clone()
		})
	}

	return cloned
}

func filterError(err error) error {
	return fmt.Errorf("filter error: %w", err)
}

// var operators = []string{"$eq", "$neq", "$gt", "$gte", "$lt", "$lte", "$like", "$in", "$nin", "$null"}
// var rootOperators = []string{"$and", "$or"}

// CreatePredicatesFromFilterObject creates a predicate from a filter object
// A filter object is a JSON object that contains the filter for the query
// E.g.
//
//	{
//		"approved": true,
//		"status": "online",
//		"name": {
//			"$like": "test%",
//			"$neq": "test2"
//		},
//		"age": {
//			"$gt": 1,
//			"$lt": 10
//		},
//		"$or": [
//			{
//				"age": 1
//			},
//			{
//				"bio": {
//					"$like": "test%",
//					"$neq": "test2"
//				}
//			},
//			{
//				"status": "offline"
//			},
//			{
//				"$and": [
//					{
//						"bio": {
//							"$neq": "test",
//							"$like": "%a"
//						}
//					},
//					{
//						"age": {
//							"$gt": 1
//						}
//					}
//				]
//			}
//		]
//	}
//
// will be converted to
// "entgo.io/ent/dialect/sql"
// sql.And(
//
//	sql.EQ("approved", true),
//	sql.EQ("status", "online"),
//	sql.And(
//		sql.Like("name", "test%"),
//		sql.NEQ("name", "test2"),
//	),
//	sql.And(
//		sql.GT("age", 1),
//		sql.LT("age", 10),
//	),
//	sql.Or(
//		sql.EQ("age", 1),
//		sql.And(
//			sql.Like("bio", "test%"),
//			sql.NEQ("bio", "test2"),
//		),
//		sql.EQ("status", "offline"),
//		sql.And(
//			sql.NEQ("bio", "test"),
//			sql.Like("bio", "%a"),
//			sql.GT("age", 1),
//		),
//	),
//
// )
func CreatePredicatesFromFilterObject(
	sb *schema.Builder,
	s *schema.Schema,
	filterObject string,
) ([]*Predicate, error) {
	if filterObject == "" {
		return []*Predicate{}, nil
	}

	filterEntity, err := entity.NewEntityFromJSON(filterObject)
	if err != nil {
		return nil, filterError(err)
	}

	return createObjectPredicates(sb, s, filterEntity)
}

func CreatePredicatesFromFilterMap(
	sb *schema.Builder,
	s *schema.Schema,
	filterObject map[string]any,
) ([]*Predicate, error) {
	if filterObject == nil {
		return []*Predicate{}, nil
	}

	filterEntity := entity.NewEntityFromMap(filterObject)

	return createObjectPredicates(sb, s, filterEntity)
}

// createObjectPredicates creates predicates from a filter object.
// Each field in the filter object is connected by AND.
// If the field is $or or $and, the field value is an array of filter objects (array of entities)
// Otherwise, the field value is a single entity that
// contains the filter for the field (e.g. { "age": { "$gt": 1, "$lt": 10 } }).
// Returns an array of predicates.
func createObjectPredicates(
	sb *schema.Builder,
	s *schema.Schema,
	filterObject *entity.Entity,
) ([]*Predicate, error) {
	var predicates = make([]*Predicate, 0)

	for pair := filterObject.First(); pair != nil; pair = pair.Next() {
		if pair.Key == "$or" || pair.Key == "$and" {
			opEntities, ok := pair.Value.([]*entity.Entity)
			if !ok {
				return nil, errors.New("invalid $or/$and value")
			}

			var opPredicates []*Predicate
			for _, opEntity := range opEntities {
				/**
				In a single entity, all field are connected by AND
				E.g. { "age": { "$gt": 1, "$lt": 10 } } --> age > 1 AND age < 10
				*/
				objectPredicates, err := createObjectPredicates(sb, s, opEntity)
				if err != nil {
					return nil, err
				}

				if len(objectPredicates) > 1 {
					opPredicates = append(opPredicates, And(objectPredicates...))
				} else {
					opPredicates = append(opPredicates, objectPredicates...)
				}
				// opPredicates = append(opPredicates, objectPredicates...)
			}

			op := utils.If(pair.Key == "$or", Or, And)
			predicates = append(predicates, op(opPredicates...))
			continue
		}

		var fieldPredicates []*Predicate
		var err error
		var fieldPath string // The full field path (e.g., "teams.slug" or "name")

		// If the field contains ".", it is a relation filter
		// E.g. "owner.groups.name": { "$like": "group_or_the_pet_owner%" }
		if strings.Contains(pair.Key, ".") {
			relationFields := utils.Filter(
				strings.Split(pair.Key, "."),
				func(s string) bool {
					return strings.TrimSpace(s) != ""
				},
			)

			if len(relationFields) < 2 {
				return nil, filterError(fmt.Errorf("%s is not a valid relation field", pair.Key))
			}

			// last field is the last relation filter column
			lastRelationColumn := relationFields[len(relationFields)-1]
			relationFieldsPath := relationFields[:len(relationFields)-1]

			currentSchema := s
			var targetSchema *schema.Schema = nil

			for _, relationField := range relationFieldsPath {
				currentField := currentSchema.Field(relationField)
				if currentField == nil {
					return nil, filterError(schema.ErrFieldNotFound(currentSchema.Name, relationField))
				}

				if !currentField.Type.IsRelationType() {
					return nil, filterError(fmt.Errorf("%s is not a relation field", pair.Key))
				}

				targetSchema, err = sb.Schema(currentField.Relation.TargetSchemaName)
				if err != nil {
					return nil, filterError(fmt.Errorf("invalid relation schema %s: %w", relationField, err))
				}

				currentSchema = targetSchema
			}

			lastRelationField := targetSchema.Field(lastRelationColumn)
			if lastRelationField == nil {
				return nil, filterError(schema.ErrFieldNotFound(targetSchema.Name, lastRelationColumn))
			}

			// Store the full dot notation path (e.g., "teams.slug")
			fieldPath = pair.Key
			if fieldPredicates, err = createFieldPredicate(
				lastRelationField,
				pair.Value,
			); err != nil {
				return nil, err
			}
		} else {
			f := s.Field(pair.Key)
			if f == nil {
				return nil, filterError(schema.ErrFieldNotFound(s.Name, pair.Key))
			}
			fieldPath = pair.Key

			// For relation fields without dot notation (implicit PK filter),
			// create predicates directly without type validation.
			// The transformation to target schema's PK happens in createEntPredicates.
			// E.g. {"country": 1} or {"country": {"$in": [1, 2, 3]}}
			if f.Type.IsRelationType() {
				if fieldPredicates, err = createRelationFieldPredicates(f.Name, pair.Value); err != nil {
					return nil, err
				}
			} else {
				if fieldPredicates, err = createFieldPredicate(f, pair.Value); err != nil {
					return nil, err
				}
			}
		}

		// Update the field path in predicates to include relation path if any
		if len(fieldPredicates) > 1 {
			// Update each predicate's field to use the full path
			for _, fp := range fieldPredicates {
				if strings.Contains(fieldPath, ".") {
					// Replace the field name with the full path
					parts := strings.Split(fieldPath, ".")
					parts[len(parts)-1] = fp.Field
					fp.Field = strings.Join(parts, ".")
				}
			}
			predicates = append(predicates, And(fieldPredicates...))
		} else {
			if len(fieldPredicates) == 0 {
				return nil, errors.New("invalid field predicates")
			}
			p := fieldPredicates[0]
			if strings.Contains(fieldPath, ".") {
				// Replace the field name with the full path
				parts := strings.Split(fieldPath, ".")
				parts[len(parts)-1] = p.Field
				p.Field = strings.Join(parts, ".")
			}
			predicates = append(predicates, p)
		}
	}

	return predicates, nil
}

// createRelationFieldPredicates creates predicates for a relation field without dot notation.
// This handles implicit PK filtering like {"country": 1} or {"country": {"$in": [1, 2, 3]}}.
// Unlike createFieldPredicate, this doesn't validate the value type since the actual
// field type (the target schema's PK) is resolved later in createEntPredicates.
func createRelationFieldPredicates(fieldName string, value any) ([]*Predicate, error) {
	return createPredicatesFromValue(fieldName, value, nil)
}

// createFieldPredicate creates predicates for a single field
// A field can have multiple operators, e.g.
// { "age": { "$gt": 1, "$lt": 10 } } --> age > 1 AND age < 10
// A field can have a single value, e.g. { "age": 1 } --> age = 1
// If a field value is an entity, it means there may be some operators. E.g.
// { "age": { "$gt": 1, "$lt": 10 } }
// We need to loop through the entity and create predicate for each operator
// If a field value is a primitive, using the default operator $eq
// Returns an array of predicates for a single field
func createFieldPredicate(
	field *schema.Field,
	value any,
) ([]*Predicate, error) {
	return createPredicatesFromValue(field.Name, value, field)
}

// createPredicatesFromValue creates predicates from a value for a given field name.
// If field is provided, it validates the value type against the field type.
// If field is nil, no type validation is performed (used for relation fields).
func createPredicatesFromValue(fieldName string, value any, field *schema.Field) ([]*Predicate, error) {
	switch fieldValue := value.(type) {
	// If the value is an entity, it means there are some operators. E.g.
	// { "age": { "$gt": 1, "$lt": 10 } }
	// --> create predicates for each operator
	case *entity.Entity:
		predicates := []*Predicate{}
		for p := fieldValue.First(); p != nil; p = p.Next() {
			op := stringToOperatorTypes[p.Key]
			if op == OpInvalid {
				return nil, filterError(fmt.Errorf("invalid operator %s for field %s", p.Key, fieldName))
			}

			// Validate value type if field is provided (skip for relation fields)
			if field != nil && op != OpNULL && !field.IsValidValue(p.Value) {
				return nil, filterError(fmt.Errorf(
					"invalid value for field %s.%s (%s) = %v (%T)",
					fieldName,
					p.Key,
					field.Type,
					p.Value,
					p.Value,
				))
			}

			predicate, err := createOperatorPredicate(fieldName, op, p.Value)
			if err != nil {
				return nil, err
			}
			predicates = append(predicates, predicate)
		}
		return predicates, nil
	// If the value is primitive
	// --> create a simple EQ predicate (string, int, uint, bool, etc.)
	default:
		// Validate value type if field is provided (skip for relation fields)
		if field != nil && !field.IsValidValue(fieldValue) {
			return nil, filterError(fmt.Errorf(
				"invalid value for field %s (%s) = %v (%T)",
				fieldName,
				field.Type,
				fieldValue,
				fieldValue,
			))
		}
		return []*Predicate{EQ(fieldName, fieldValue)}, nil
	}
}

// createOperatorPredicate creates a single predicate for a given operator and value.
func createOperatorPredicate(fieldName string, op OperatorType, value any) (*Predicate, error) {
	switch op {
	case OpEQ:
		return EQ(fieldName, value), nil
	case OpNEQ:
		return NEQ(fieldName, value), nil
	case OpGT:
		return GT(fieldName, value), nil
	case OpGTE:
		return GTE(fieldName, value), nil
	case OpLT:
		return LT(fieldName, value), nil
	case OpLTE:
		return LTE(fieldName, value), nil
	case OpLike:
		stringVal, ok := value.(string)
		if !ok {
			return nil, filterError(errors.New("$like operator must be a string"))
		}
		return Like(fieldName, stringVal), nil
	case OpNotLike:
		stringVal, ok := value.(string)
		if !ok {
			return nil, filterError(errors.New("$notlike operator must be a string"))
		}
		return NotLike(fieldName, stringVal), nil
	case OpContains:
		stringVal, ok := value.(string)
		if !ok {
			return nil, filterError(errors.New("$contains operator must be a string"))
		}
		return Contains(fieldName, stringVal), nil
	case OpNotContains:
		stringVal, ok := value.(string)
		if !ok {
			return nil, filterError(errors.New("$notcontains operator must be a string"))
		}
		return NotContains(fieldName, stringVal), nil
	case OpContainsFold:
		stringVal, ok := value.(string)
		if !ok {
			return nil, filterError(errors.New("$containsfold operator must be a string"))
		}
		return ContainsFold(fieldName, stringVal), nil
	case OpNotContainsFold:
		stringVal, ok := value.(string)
		if !ok {
			return nil, filterError(errors.New("$notcontainsfold operator must be a string"))
		}
		return NotContainsFold(fieldName, stringVal), nil
	case OpIN:
		arrayVal, ok := value.([]any)
		if !ok {
			return nil, filterError(errors.New("$in operator must be an array"))
		}
		return In(fieldName, arrayVal), nil
	case OpNIN:
		arrayVal, ok := value.([]any)
		if !ok {
			return nil, filterError(errors.New("$nin operator must be an array"))
		}
		return NotIn(fieldName, arrayVal), nil
	case OpNULL:
		boolVal, ok := value.(bool)
		if !ok {
			return nil, filterError(errors.New("$null operator must be a boolean"))
		}
		return Null(fieldName, boolVal), nil
	default:
		return nil, filterError(fmt.Errorf("unsupported operator %s", op))
	}
}

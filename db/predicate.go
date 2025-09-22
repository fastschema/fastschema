package db

import (
	"errors"
	"fmt"
	"strings"

	"github.com/fastschema/fastschema/entity"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
)

type Predicate struct {
	Field              string       `json:"field"`
	Operator           OperatorType `json:"operator"`
	Value              any          `json:"value"`
	RelationFieldNames []string     `json:"relationFieldNames"`
	And                []*Predicate `json:"and"`
	Or                 []*Predicate `json:"or"`
}

func (p *Predicate) Clone() *Predicate {
	cloned := &Predicate{
		Field:              p.Field,
		Operator:           p.Operator,
		Value:              p.Value,
		RelationFieldNames: p.RelationFieldNames,
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

		var fieldRelations []string
		var fieldPredicates []*Predicate
		var err error

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
			relationFields = relationFields[:len(relationFields)-1]

			currentSchema := s
			var targetSchema *schema.Schema = nil

			for _, relationField := range relationFields {
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

			fieldRelations = relationFields
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
			if fieldPredicates, err = createFieldPredicate(f, pair.Value); err != nil {
				return nil, err
			}
		}

		if len(fieldPredicates) > 1 {
			p := And(fieldPredicates...)
			p.RelationFieldNames = fieldRelations
			predicates = append(predicates, p)
		} else {
			if len(fieldPredicates) == 0 {
				return nil, errors.New("invalid field predicates")
			}
			p := fieldPredicates[0]
			p.RelationFieldNames = fieldRelations
			predicates = append(predicates, p)
		}
	}

	return predicates, nil
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
	switch fieldValue := value.(type) {
	// If the value is an entity, it means there are some operators. E.g.
	// { "age": { "$gt": 1, "$lt": 10 } }
	// --> create predicates for each operator
	case *entity.Entity:
		predicates := []*Predicate{}
		for p := fieldValue.First(); p != nil; p = p.Next() {
			op := stringToOperatorTypes[p.Key]
			if op != OpNULL && !field.IsValidValue(p.Value) {
				return nil, filterError(fmt.Errorf(
					"invalid value for field %s.%s (%s) = %v (%T)",
					field.Name,
					p.Key,
					field.Type,
					p.Value,
					p.Value,
				))
			}

			switch stringToOperatorTypes[p.Key] {
			case OpEQ:
				predicates = append(predicates, EQ(field.Name, p.Value))
			case OpNEQ:
				predicates = append(predicates, NEQ(field.Name, p.Value))
			case OpGT:
				predicates = append(predicates, GT(field.Name, p.Value))
			case OpGTE:
				predicates = append(predicates, GTE(field.Name, p.Value))
			case OpLT:
				predicates = append(predicates, LT(field.Name, p.Value))
			case OpLTE:
				predicates = append(predicates, LTE(field.Name, p.Value))
			case OpLike:
				stringVal, ok := p.Value.(string)
				if !ok {
					return nil, filterError(errors.New("$like operator must be a string"))
				}
				predicates = append(predicates, Like(field.Name, stringVal))
			case OpNotLike:
				stringVal, ok := p.Value.(string)
				if !ok {
					return nil, filterError(errors.New("$notlike operator must be a string"))
				}
				predicates = append(predicates, NotLike(field.Name, stringVal))
			case OpContains:
				stringVal, ok := p.Value.(string)
				if !ok {
					return nil, filterError(errors.New("$contains operator must be a string"))
				}
				predicates = append(predicates, Contains(field.Name, stringVal))
			case OpNotContains:
				stringVal, ok := p.Value.(string)
				if !ok {
					return nil, filterError(errors.New("$notcontains operator must be a string"))
				}
				predicates = append(predicates, NotContains(field.Name, stringVal))
			case OpContainsFold:
				stringVal, ok := p.Value.(string)
				if !ok {
					return nil, filterError(errors.New("$containsfold operator must be a string"))
				}
				predicates = append(predicates, ContainsFold(field.Name, stringVal))
			case OpNotContainsFold:
				stringVal, ok := p.Value.(string)
				if !ok {
					return nil, filterError(errors.New("$notcontainsfold operator must be a string"))
				}
				predicates = append(predicates, NotContainsFold(field.Name, stringVal))
			case OpIN:
				arrayVal, ok := p.Value.([]any)
				if !ok {
					return nil, filterError(errors.New("$in operator must be an array"))
				}
				predicates = append(predicates, In(field.Name, arrayVal))
			case OpNIN:
				arrayVal, ok := p.Value.([]any)
				if !ok {
					return nil, filterError(errors.New("$nin operator must be an array"))
				}
				predicates = append(predicates, NotIn(field.Name, arrayVal))
			case OpNULL:
				boolVal, ok := p.Value.(bool)
				if !ok {
					return nil, filterError(errors.New("$null operator must be a boolean"))
				}
				predicates = append(predicates, Null(field.Name, boolVal))
			}
		}

		return predicates, nil
	// If the value is primitive
	// --> create a simple EQ predicate (string, int, uint, bool, etc.)
	default:
		if !field.IsValidValue(fieldValue) {
			return nil, filterError(fmt.Errorf(
				"invalid value for field %s (%s) = %v (%T)",
				field.Name,
				field.Type,
				fieldValue,
				fieldValue,
			))
		}
		return []*Predicate{EQ(field.Name, fieldValue)}, nil
	}
}

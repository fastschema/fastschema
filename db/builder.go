package db

import (
	"errors"
	"reflect"

	"github.com/fastschema/fastschema/pkg/utils"
)

type QueryBuilder[T any] struct {
	// Base struct for Query and Mutation
	rType      reflect.Type
	schemaName string
	client     Client
	predicates []*Predicate

	// Query specific fields
	limit  uint
	offset uint
	fields []string
	order  []string
}

func Builder[T any](client Client, schemas ...string) *QueryBuilder[T] {
	query := &QueryBuilder[T]{client: client}
	// if the schema name is specified, use the schema name
	// otherwise, use the reflect type of the schema
	if len(schemas) > 0 && schemas[0] != "" {
		query.schemaName = schemas[0]
	} else {
		query.rType = utils.GetDereferencedType(new(T))
	}

	return query
}

// model returns the actual model of the builder.
func (q *QueryBuilder[T]) model() (Model, error) {
	if q.rType != nil && q.rType.String() == "entity.Entity" && q.schemaName == "" {
		return nil, errors.New("schema name is required for type entity.Entity")
	}

	// if the schema name is not empty, use the schema name
	if q.schemaName != "" {
		return q.client.Model(q.schemaName)
	}

	// if the schema name is empty, use the reflect type of the schema
	return q.client.Model(q.rType)
}

// Where adds the given predicates to the builder.
func (q *QueryBuilder[T]) Where(predicates ...*Predicate) *QueryBuilder[T] {
	q.predicates = append(q.predicates, predicates...)
	return q
}

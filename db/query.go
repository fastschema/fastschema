package db

import (
	"context"
	"fmt"
	"reflect"

	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
	"github.com/mitchellh/mapstructure"
)

type DBQuery[T any] struct {
	rType      reflect.Type
	client     Client
	predicates []*Predicate
	limit      uint
	offset     uint
	fields     []string
	order      []string
}

func BindStruct(data any, target any) error {
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName: "json",
		Result:  target,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(func(from, to reflect.Type, data any) (any, error) {
			if e, ok := data.(*schema.Entity); ok {
				return e.ToMap(), nil
			}

			return data, nil
		}),
	})
	if err != nil {
		return err
	}

	if err := decoder.Decode(data); err != nil {
		return err
	}

	return nil
}

func Query[T any](client Client) *DBQuery[T] {
	var t T
	return &DBQuery[T]{
		client: client,
		rType:  utils.GetDereferencedType(t),
	}
}

// Limit sets the limit of the query.
func (q *DBQuery[T]) Limit(limit uint) *DBQuery[T] {
	q.limit = limit
	return q
}

// Offset sets the offset of the query.
func (q *DBQuery[T]) Offset(offset uint) *DBQuery[T] {
	q.offset = offset
	return q
}

// Order sets the order of the query.
func (q *DBQuery[T]) Order(order ...string) *DBQuery[T] {
	q.order = append(q.order, order...)
	return q
}

// Select sets the columns of the query.
func (q *DBQuery[T]) Select(fields ...string) *DBQuery[T] {
	q.fields = append(q.fields, fields...)
	return q
}

// Where adds the given predicates to the query.
func (q *DBQuery[T]) Where(predicates ...*Predicate) *DBQuery[T] {
	q.predicates = append(q.predicates, predicates...)
	return q
}

// Count returns the number of entities that match the query.
func (q *DBQuery[T]) Count(ctx context.Context, options *CountOption) (int, error) {
	model, err := q.client.Model("", q.rType)
	if err != nil {
		return 0, err
	}

	count, err := model.Query(q.predicates...).Count(ctx, options)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// Get returns the list of entities that match the query.
func (q *DBQuery[T]) Get(ctx context.Context) ([]T, error) {
	model, err := q.client.Model("", q.rType)
	if err != nil {
		return nil, err
	}

	entities, err := model.Query(q.predicates...).
		Limit(q.limit).
		Offset(q.offset).
		Order(q.order...).
		Select(q.fields...).
		Get(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]T, 0)
	for _, e := range entities {
		var record T
		if err := BindStruct(e, &record); err != nil {
			return nil, err
		}

		result = append(result, record)
	}

	return result, nil
}

// First returns the first entity that matches the query.
func (q *DBQuery[T]) First(ctx context.Context) (t T, err error) {
	q.Limit(1)
	entities, err := q.Get(ctx)

	if err != nil {
		return t, err
	}

	if len(entities) == 0 {
		return t, &NotFoundError{Message: "no entities found"}
	}

	return entities[0], nil
}

// Only returns the matched entity or an error if there is more than one.
func (q *DBQuery[T]) Only(ctx context.Context) (t T, err error) {
	entities, err := q.Get(ctx)

	if err != nil {
		return t, err
	}

	if len(entities) > 1 {
		return t, fmt.Errorf("more than one entity found")
	}

	if len(entities) == 0 {
		return t, &NotFoundError{Message: "no entities found"}
	}

	return entities[0], nil
}

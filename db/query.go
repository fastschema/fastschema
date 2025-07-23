package db

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/fastschema/fastschema/entity"
	"github.com/mitchellh/mapstructure"
)

/** Query related methods **/

// Limit sets the limit of the query.
func (q *QueryBuilder[T]) Limit(limit uint) *QueryBuilder[T] {
	q.limit = limit
	return q
}

// Offset sets the offset of the query.
func (q *QueryBuilder[T]) Offset(offset uint) *QueryBuilder[T] {
	q.offset = offset
	return q
}

// Order sets the order of the query.
func (q *QueryBuilder[T]) Order(order ...string) *QueryBuilder[T] {
	q.order = append(q.order, order...)
	return q
}

// Select sets the columns of the query.
func (q *QueryBuilder[T]) Select(fields ...string) *QueryBuilder[T] {
	q.fields = append(q.fields, fields...)
	return q
}

// Count returns the number of entities that match the query.
func (q *QueryBuilder[T]) Count(ctx context.Context, options ...*QueryOption) (int, error) {
	model, err := q.model()
	if err != nil {
		return 0, err
	}

	count, err := model.Query(q.predicates...).Count(ctx, options...)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// Get returns the list of entities that match the query.
func (q *QueryBuilder[T]) Get(ctx context.Context) ([]T, error) {
	model, err := q.model()
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

	var t T
	_, tIsEntity := any(t).(*entity.Entity)

	result := make([]T, 0)
	for _, e := range entities {
		if tIsEntity {
			if converted, ok := any(e).(T); ok {
				result = append(result, converted)
			} else {
				return nil, fmt.Errorf("failed to convert entity to type %T", t)
			}
			continue
		}

		var record T
		if err := BindStruct(e, &record); err != nil {
			return nil, err
		}

		result = append(result, record)
	}

	return result, nil
}

// First returns the first entity that matches the query.
func (q *QueryBuilder[T]) First(ctx context.Context) (t T, err error) {
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
func (q *QueryBuilder[T]) Only(ctx context.Context) (t T, err error) {
	entities, err := q.Get(ctx)

	if err != nil {
		return t, err
	}

	if len(entities) > 1 {
		return t, errors.New("more than one entity found")
	}

	if len(entities) == 0 {
		return t, &NotFoundError{Message: "no entities found"}
	}

	return entities[0], nil
}

func BindStruct(src any, target any) error {
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName: "json",
		Result:  target,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(func(from, to reflect.Type, data any) (any, error) {
			if e, ok := data.(*entity.Entity); ok {
				return e.ToMap(), nil
			}

			return data, nil
		}),
	})
	if err != nil {
		return err
	}

	if err := decoder.Decode(src); err != nil {
		return err
	}

	return nil
}

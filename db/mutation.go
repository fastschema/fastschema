package db

import (
	"context"
	"reflect"

	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
)

type DBMutation[T any] struct {
	rType      reflect.Type
	client     Client
	predicates []*Predicate
}

// Mutation is a helper function to create a new mutation for a given model
//
//	The model is inferred from the type of the entity
//	T must be a pointer to a struct or a struct type
func Mutation[T any](client Client) *DBMutation[T] {
	var t T
	return &DBMutation[T]{
		client:     client,
		rType:      utils.GetDereferencedType(t),
		predicates: make([]*Predicate, 0),
	}
}

// Where adds a predicate to the mutation
func (m *DBMutation[T]) Where(predicates ...*Predicate) *DBMutation[T] {
	m.predicates = append(m.predicates, predicates...)
	return m
}

// Create creates a new entity and return the newly created entity
func (m *DBMutation[T]) Create(ctx context.Context, e *schema.Entity) (t T, err error) {
	model, err := m.client.Model("", m.rType)
	if err != nil {
		return t, err
	}

	id, err := model.Mutation().Create(ctx, e)
	if err != nil {
		return t, err
	}

	return Query[T](m.client).Where(EQ(schema.FieldID, id)).First(ctx)
}

// Update updates the entity and returns the updated entities
func (m *DBMutation[T]) Update(ctx context.Context, e *schema.Entity) (t []T, err error) {
	model, err := m.client.Model("", m.rType)
	if err != nil {
		return t, err
	}

	if _, err = model.Mutation().
		Where(m.predicates...).
		Update(ctx, e); err != nil {
		return t, err
	}

	return Query[T](m.client).Where(m.predicates...).Get(ctx)
}

// CreateFromJSON creates a new entity from JSON
func (m *DBMutation[T]) CreateFromJSON(ctx context.Context, json string) (t T, err error) {
	entity, err := schema.NewEntityFromJSON(json)

	if err != nil {
		return t, err
	}

	return m.Create(ctx, entity)
}

// Delete deletes entities from the database
func (m *DBMutation[T]) Delete(ctx context.Context) (affected int, err error) {
	model, err := m.client.Model("", m.rType)
	if err != nil {
		return 0, err
	}

	return model.Mutation().Where(m.predicates...).Delete(ctx)
}

// Create is a shortcut to mutation.Create
func Create[T any](ctx context.Context, client Client, e *schema.Entity) (T, error) {
	return Mutation[T](client).Create(ctx, e)
}

// Update is a shortcut to mutation.Update
func Update[T any](ctx context.Context, client Client, e *schema.Entity, predicates ...*Predicate) ([]T, error) {
	return Mutation[T](client).Where(predicates...).Update(ctx, e)
}

// Delete is a shortcut to mutation.Delete
func Delete[T any](ctx context.Context, client Client, predicates ...*Predicate) (int, error) {
	return Mutation[T](client).Where(predicates...).Delete(ctx)
}

// CreateFromJSON is a shortcut to mutation.CreateFromJSON
func CreateFromJSON[T any](ctx context.Context, client Client, json string) (T, error) {
	return Mutation[T](client).CreateFromJSON(ctx, json)
}

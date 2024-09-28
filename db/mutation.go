package db

import (
	"context"
	"fmt"

	"github.com/fastschema/fastschema/pkg/errors"
	"github.com/fastschema/fastschema/schema"
)

/** Mutation related methods **/

// Create creates a new entity and return the newly created entity
func (m *QueryBuilder[T]) Create(ctx context.Context, dataCreate any) (t T, err error) {
	model, err := m.model()
	if err != nil {
		return t, err
	}

	entityCreate, err := typesToEntity(dataCreate)
	if err != nil {
		return t, fmt.Errorf("cannot create entity: %w", err)
	}

	id, err := model.Mutation().Create(ctx, entityCreate)
	if err != nil {
		return t, err
	}

	q := Builder[T](m.client, m.schemaName)

	return q.Where(EQ(schema.FieldID, id)).First(ctx)
}

// CreateFromJSON creates a new entity from JSON
func (m *QueryBuilder[T]) CreateFromJSON(ctx context.Context, json string) (t T, err error) {
	entity, err := schema.NewEntityFromJSON(json)

	if err != nil {
		return t, err
	}

	return m.Create(ctx, entity)
}

// Update updates the entity and returns the updated entities
func (m *QueryBuilder[T]) Update(ctx context.Context, updateData any) (ts []T, err error) {
	entityUpdate, err := typesToEntity(updateData)
	if err != nil {
		return ts, fmt.Errorf("cannot create entity: %w", err)
	}

	model, err := m.model()
	if err != nil {
		return ts, err
	}

	if _, err = model.Mutation().
		Where(m.predicates...).
		Update(ctx, entityUpdate); err != nil {
		return ts, err
	}

	q := Builder[T](m.client, m.schemaName)

	return q.Where(m.predicates...).Get(ctx)
}

// Delete deletes entities from the database
func (m *QueryBuilder[T]) Delete(ctx context.Context) (affected int, err error) {
	model, err := m.model()
	if err != nil {
		return 0, err
	}

	return model.Mutation().Where(m.predicates...).Delete(ctx)
}

// Create is a shortcut to mutation.Create
func Create[T any](
	ctx context.Context,
	client Client,
	dataCreate any,
	schemas ...string,
) (T, error) {
	return Builder[T](client, schemas...).Create(ctx, dataCreate)
}

// CreateFromJSON is a shortcut to mutation.CreateFromJSON
func CreateFromJSON[T any](
	ctx context.Context,
	client Client,
	json string,
	schemas ...string,
) (T, error) {
	return Builder[T](client, schemas...).CreateFromJSON(ctx, json)
}

// Update is a shortcut to mutation.Update
func Update[T any](
	ctx context.Context,
	client Client,
	dataUpdate any,
	predicates []*Predicate,
	schemas ...string,
) ([]T, error) {
	return Builder[T](client, schemas...).Where(predicates...).Update(ctx, dataUpdate)
}

// Delete is a shortcut to mutation.Delete
func Delete[T any](
	ctx context.Context,
	client Client,
	predicates []*Predicate,
	schemas ...string,
) (int, error) {
	return Builder[T](client, schemas...).Where(predicates...).Delete(ctx)
}

func typesToEntity(t any) (*schema.Entity, error) {
	var entity *schema.Entity
	switch data := t.(type) {
	case *schema.Entity:
		entity = data
	case map[string]any:
		entity = schema.NewEntityFromMap(data)
	default:
		return nil, errors.BadRequest("mutation data must be an entity or a map")
	}

	return entity, nil
}

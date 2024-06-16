package db

import (
	"context"
	"fmt"
	"reflect"

	"github.com/fastschema/fastschema/pkg/errors"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
)

type DBMutation[T any] struct {
	rType      reflect.Type
	schemaName string
	client     Client
	predicates []*Predicate
}

// Mutation is a helper function to create a new mutation for a given schema
//
//	The schema is inferred from the type of the entity
//	T must be a pointer to a struct or a struct type
func Mutation[T any](client Client, schemas ...string) *DBMutation[T] {
	mutation := &DBMutation[T]{
		client:     client,
		predicates: make([]*Predicate, 0),
	}

	// if the schema name is specified, use the schema name
	// otherwise, use the reflect type of the schema
	if len(schemas) > 0 && schemas[0] != "" {
		mutation.schemaName = schemas[0]
	} else {
		mutation.rType = utils.GetDereferencedType(new(T))
	}

	return mutation
}

// Model returns the actual model of the mutation.
func (q *DBMutation[T]) Model() (Model, error) {
	if q.rType != nil && q.rType.String() == "schema.Entity" && q.schemaName == "" {
		return nil, fmt.Errorf("schema name is required for type schema.Entity")
	}

	// if the schema name is not empty, use the schema name
	if q.schemaName != "" {
		return q.client.Model(q.schemaName)
	}

	// if the schema name is empty, use the reflect type of the schema
	return q.client.Model("", q.rType)
}

// Where adds a predicate to the mutation
func (m *DBMutation[T]) Where(predicates ...*Predicate) *DBMutation[T] {
	m.predicates = append(m.predicates, predicates...)
	return m
}

// Create creates a new entity and return the newly created entity
func (m *DBMutation[T]) Create(ctx context.Context, dataCreate any) (t T, err error) {
	model, err := m.Model()
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

	q := Query[T](m.client, m.schemaName)

	return q.Where(EQ(schema.FieldID, id)).First(ctx)
}

// CreateFromJSON creates a new entity from JSON
func (m *DBMutation[T]) CreateFromJSON(ctx context.Context, json string) (t T, err error) {
	entity, err := schema.NewEntityFromJSON(json)

	if err != nil {
		return t, err
	}

	return m.Create(ctx, entity)
}

// Update updates the entity and returns the updated entities
func (m *DBMutation[T]) Update(ctx context.Context, updateData any) (ts []T, err error) {
	entityUpdate, err := typesToEntity(updateData)
	if err != nil {
		return ts, fmt.Errorf("cannot create entity: %w", err)
	}

	model, err := m.Model()
	if err != nil {
		return ts, err
	}

	if _, err = model.Mutation().
		Where(m.predicates...).
		Update(ctx, entityUpdate); err != nil {
		return ts, err
	}

	q := Query[T](m.client, m.schemaName)

	return q.Where(m.predicates...).Get(ctx)
}

// Delete deletes entities from the database
func (m *DBMutation[T]) Delete(ctx context.Context) (affected int, err error) {
	model, err := m.Model()
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
	return Mutation[T](client, schemas...).Create(ctx, dataCreate)
}

// CreateFromJSON is a shortcut to mutation.CreateFromJSON
func CreateFromJSON[T any](
	ctx context.Context,
	client Client,
	json string,
	schemas ...string,
) (T, error) {
	return Mutation[T](client, schemas...).CreateFromJSON(ctx, json)
}

// Update is a shortcut to mutation.Update
func Update[T any](
	ctx context.Context,
	client Client,
	dataUpdate any,
	predicates []*Predicate,
	schemas ...string,
) ([]T, error) {
	return Mutation[T](client, schemas...).Where(predicates...).Update(ctx, dataUpdate)
}

// Delete is a shortcut to mutation.Delete
func Delete[T any](
	ctx context.Context,
	client Client,
	predicates []*Predicate,
	schemas ...string,
) (int, error) {
	return Mutation[T](client, schemas...).Where(predicates...).Delete(ctx)
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

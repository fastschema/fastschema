package entdbadapter

import (
	"context"
	"errors"
	"fmt"

	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/fastschema/fastschema/entity"
	"github.com/fastschema/fastschema/expr"
	"github.com/fastschema/fastschema/schema"
	"github.com/google/uuid"
)

// Create creates a new entity in the database
func (m *Mutation) Create(ctx context.Context, e *entity.Entity) (_ any, err error) {
	if m.model == nil || m.model.schema == nil {
		return nil, fmt.Errorf("model or schema %s not found", m.model.name)
	}

	if err := m.model.schema.ApplySetters(ctx, e, expr.Config{
		DB: func() expr.DBLike {
			return m.client
		},
	}); err != nil {
		return nil, err
	}

	if err := runPreDBCreateHooks(ctx, m.client, m.model.schema, e); err != nil {
		return nil, err
	}

	// Auto-generate UUID v7 for primary key if not provided
	if err := m.autoGenerateUUID(e); err != nil {
		return nil, err
	}

	createSpec := &sqlgraph.CreateSpec{
		Table: m.model.schema.Namespace,
		ID: &sqlgraph.FieldSpec{
			Column: m.model.entPrimaryColumn.Name,
			Type:   m.model.entPrimaryColumn.Type,
		},
		Fields: []*sqlgraph.FieldSpec{},
		Edges:  []*sqlgraph.EdgeSpec{},
	}

	entAdapter, ok := m.client.(EntAdapter)
	if !ok {
		return nil, errors.New("client is not an ent adapter")
	}

	var c *Column
	for pair := e.First(); pair != nil; pair = pair.Next() {
		fieldName := pair.Key
		fieldValue := pair.Value

		c, err = m.model.Column(fieldName)
		if err != nil {
			return nil, fmt.Errorf("column error: %w", err)
		}

		// Non-relation fields
		if !c.field.Type.IsRelationType() {
			createSpec.Fields = append(createSpec.Fields, &sqlgraph.FieldSpec{
				Column: c.entColumn.Name,
				Value:  fieldValue,
				Type:   c.entColumn.Type,
			})
			continue
		}

		// Relation fields
		relationEntityIDs, err := m.GetRelationEntityIDs(c.field.Name, fieldValue)
		if err != nil {
			return nil, err
		}

		if len(relationEntityIDs) > 0 {
			edge, err := entAdapter.NewEdgeSpec(c.field.Relation, relationEntityIDs)
			if err != nil {
				return nil, fmt.Errorf("edge error %s.%s: %w", m.model.name, fieldName, err)
			}

			createSpec.Edges = append(createSpec.Edges, edge)
		}
	}

	if err = sqlgraph.CreateNode(ctx, entAdapter.Driver(), createSpec); err != nil {
		return nil, err
	}

	pkField := m.model.schema.IDField()
	insertedID := createSpec.ID.Value
	if insertedID == nil && pkField != nil {
		insertedID = e.Get(pkField.Name)
	}
	if insertedID == nil {
		insertedID = e.Get(entity.FieldID)
	}

	if insertedID == nil {
		return nil, fmt.Errorf("create mutation for %s returned no ID", m.model.name)
	}

	if pkField != nil {
		normalizedID, err := normalizeIDValue(pkField, insertedID)
		if err != nil {
			return nil, err
		}
		insertedID = normalizedID
	}

	primaryFieldName := entity.FieldID
	if pkField != nil && pkField.Name != "" {
		primaryFieldName = pkField.Name
	}
	e.SetIDField(primaryFieldName)

	if err := e.SetID(insertedID); err != nil {
		return nil, err
	}

	if m.autoCommit {
		if err = m.client.Commit(); err != nil {
			return nil, err
		}
	}

	if err := runPostDBCreateHooks(ctx, m.client, m.model.schema, e, insertedID); err != nil {
		return nil, err
	}

	return insertedID, nil
}

// autoGenerateUUID generates a UUID v7 for the primary key field if:
// - The primary key field is of type UUID
// - The primary key value is not already set in the entity
func (m *Mutation) autoGenerateUUID(e *entity.Entity) error {
	pkField := m.model.schema.IDField()
	if pkField == nil {
		return nil
	}

	// Only auto-generate for UUID type primary keys
	if pkField.Type != schema.TypeUUID {
		return nil
	}

	// Check if the primary key value is already set
	existingValue := e.Get(pkField.Name)
	if !isZeroValue(existingValue) {
		return nil
	}

	// Generate a new UUID v7
	newUUID, err := uuid.NewV7()
	if err != nil {
		return fmt.Errorf("failed to generate UUID v7 for field %s: %w", pkField.Name, err)
	}

	e.Set(pkField.Name, newUUID)
	return nil
}

package entdbadapter

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"entgo.io/ent/dialect"
	dialectSql "entgo.io/ent/dialect/sql"
	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/entity"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// edgeLoader unit tests
// =============================================================================

func TestNewEdgeLoader(t *testing.T) {
	adapter := createMockAdapter(t)
	userModel, err := adapter.Model("user")
	require.NoError(t, err)
	userModelEnt, ok := userModel.(*Model)
	require.True(t, ok)

	petModel, err := adapter.Model("pet")
	require.NoError(t, err)
	petModelEnt, ok := petModel.(*Model)
	require.True(t, ok)

	parentQuery := &Query{
		client: adapter,
		model:  userModelEnt,
	}

	petsField := userModelEnt.schema.Field("pets")
	require.NotNil(t, petsField)

	relOpt := &db.RelationOption{
		Limit:  5,
		Offset: 2,
		Sort:   "-name",
	}

	loader := parentQuery.newEdgeLoader(
		context.Background(),
		petsField,
		petModelEnt,
		[]string{"name", "age"},
		relOpt,
	)

	assert.NotNil(t, loader)
	assert.Equal(t, parentQuery, loader.q)
	assert.Equal(t, petsField, loader.field)
	assert.Equal(t, petModelEnt, loader.edgeModel)
	assert.Equal(t, []string{"name", "age"}, loader.edgeColumns)
	assert.Equal(t, relOpt, loader.relOpt)
}

func TestBuildDirectEdgeConfig(t *testing.T) {
	adapter := createMockAdapter(t)
	userModel, err := adapter.Model("user")
	require.NoError(t, err)
	userModelEnt, ok := userModel.(*Model)
	require.True(t, ok)

	petModel, err := adapter.Model("pet")
	require.NoError(t, err)
	petModelEnt, ok := petModel.(*Model)
	require.True(t, ok)

	cardModel, err := adapter.Model("card")
	require.NoError(t, err)
	cardModelEnt, ok := cardModel.(*Model)
	require.True(t, ok)

	tests := []struct {
		name               string
		parentModel        *Model
		field              string
		edgeModel          *Model
		expectedWhereCol   string
		expectedIsArray    bool
		expectedRefFieldID bool
	}{
		{
			name:               "O2M owner side (user.pets)",
			parentModel:        userModelEnt,
			field:              "pets",
			edgeModel:          petModelEnt,
			expectedWhereCol:   "owner_id",
			expectedIsArray:    true,
			expectedRefFieldID: true,
		},
		{
			name:               "O2M non-owner side (pet.owner)",
			parentModel:        petModelEnt,
			field:              "owner",
			edgeModel:          userModelEnt,
			expectedWhereCol:   "id",
			expectedIsArray:    false,
			expectedRefFieldID: false,
		},
		{
			name:               "O2O owner side (user.card)",
			parentModel:        userModelEnt,
			field:              "card",
			edgeModel:          cardModelEnt,
			expectedWhereCol:   "owner_id",
			expectedIsArray:    false,
			expectedRefFieldID: true,
		},
		{
			name:               "O2O non-owner side (card.owner)",
			parentModel:        cardModelEnt,
			field:              "owner",
			edgeModel:          userModelEnt,
			expectedWhereCol:   "id",
			expectedIsArray:    false,
			expectedRefFieldID: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parentQuery := &Query{
				client: adapter,
				model:  tt.parentModel,
			}

			field := tt.parentModel.schema.Field(tt.field)
			require.NotNil(t, field)

			loader := parentQuery.newEdgeLoader(
				context.Background(),
				field,
				tt.edgeModel,
				nil,
				nil,
			)

			cfg, err := loader.buildDirectEdgeConfig()
			require.NoError(t, err)
			assert.NotNil(t, cfg)
			assert.Equal(t, tt.expectedWhereCol, cfg.whereColumn)
			assert.Equal(t, tt.expectedIsArray, cfg.isArray)
			if tt.expectedRefFieldID {
				assert.Equal(t, "id", cfg.parentRefField.Name)
			}
		})
	}
}

func TestNeedsPerParentLimitOffsetExtended(t *testing.T) {
	adapter := createMockAdapter(t)
	userModel, err := adapter.Model("user")
	require.NoError(t, err)
	userModelEnt, ok := userModel.(*Model)
	require.True(t, ok)

	petModel, err := adapter.Model("pet")
	require.NoError(t, err)
	petModelEnt, ok := petModel.(*Model)
	require.True(t, ok)

	groupModel, err := adapter.Model("group")
	require.NoError(t, err)
	groupModelEnt, ok := groupModel.(*Model)
	require.True(t, ok)

	parentQuery := &Query{
		client: adapter,
		model:  userModelEnt,
	}

	tests := []struct {
		name     string
		field    string
		model    *Model
		relOpt   *db.RelationOption
		expected bool
	}{
		{
			name:     "O2M owner with limit - should need per-parent limit",
			field:    "pets",
			model:    petModelEnt,
			relOpt:   &db.RelationOption{Limit: 2},
			expected: true,
		},
		{
			name:     "O2M owner with offset - should need per-parent limit",
			field:    "pets",
			model:    petModelEnt,
			relOpt:   &db.RelationOption{Offset: 1},
			expected: true,
		},
		{
			name:     "O2M owner with limit and offset - should need per-parent limit",
			field:    "pets",
			model:    petModelEnt,
			relOpt:   &db.RelationOption{Limit: 2, Offset: 1},
			expected: true,
		},
		{
			name:     "O2M owner without limit/offset - should not need per-parent limit",
			field:    "pets",
			model:    petModelEnt,
			relOpt:   &db.RelationOption{Sort: "-name"},
			expected: false,
		},
		{
			name:     "O2M owner with nil relOpt - should not need per-parent limit",
			field:    "pets",
			model:    petModelEnt,
			relOpt:   nil,
			expected: false,
		},
		{
			name:     "O2M owner with zero limit and offset - should not need per-parent limit",
			field:    "pets",
			model:    petModelEnt,
			relOpt:   &db.RelationOption{Limit: 0, Offset: 0},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := userModelEnt.schema.Field(tt.field)
			require.NotNil(t, field, "field %s not found", tt.field)

			loader := parentQuery.newEdgeLoader(
				context.Background(),
				field,
				tt.model,
				nil,
				tt.relOpt,
			)

			result := loader.needsPerParentLimitOffset()
			assert.Equal(t, tt.expected, result)
		})
	}

	// Test M2M relation
	t.Run("M2M with limit - should need per-parent limit", func(t *testing.T) {
		groupField := userModelEnt.schema.Field("groups")
		require.NotNil(t, groupField)

		loader := parentQuery.newEdgeLoader(
			context.Background(),
			groupField,
			groupModelEnt,
			nil,
			&db.RelationOption{Limit: 2},
		)

		result := loader.needsPerParentLimitOffset()
		assert.True(t, result)
	})
}

func TestAssignToParent(t *testing.T) {
	adapter := createMockAdapter(t)
	userModel, err := adapter.Model("user")
	require.NoError(t, err)
	userModelEnt, ok := userModel.(*Model)
	require.True(t, ok)

	petModel, err := adapter.Model("pet")
	require.NoError(t, err)
	petModelEnt, ok := petModel.(*Model)
	require.True(t, ok)

	parentQuery := &Query{
		client: adapter,
		model:  userModelEnt,
	}

	petsField := userModelEnt.schema.Field("pets")
	require.NotNil(t, petsField)

	loader := parentQuery.newEdgeLoader(
		context.Background(),
		petsField,
		petModelEnt,
		nil,
		nil,
	)

	t.Run("assign single neighbor to non-array field", func(t *testing.T) {
		parent := entity.New(1).Set("name", "John")
		neighbor := entity.New(10).Set("name", "Fido")

		err := loader.assignToParent(parent, neighbor, false)
		require.NoError(t, err)
		assert.Equal(t, neighbor, parent.Get(petsField.Name))
	})

	t.Run("assign single neighbor to array field (empty)", func(t *testing.T) {
		parent := entity.New(1).Set("name", "John")
		neighbor := entity.New(10).Set("name", "Fido")

		err := loader.assignToParent(parent, neighbor, true)
		require.NoError(t, err)

		pets := parent.Get(petsField.Name).([]*entity.Entity)
		assert.Len(t, pets, 1)
		assert.Equal(t, neighbor, pets[0])
	})

	t.Run("assign multiple neighbors to array field", func(t *testing.T) {
		parent := entity.New(1).Set("name", "John")
		neighbor1 := entity.New(10).Set("name", "Fido")
		neighbor2 := entity.New(11).Set("name", "Buddy")

		err := loader.assignToParent(parent, neighbor1, true)
		require.NoError(t, err)

		err = loader.assignToParent(parent, neighbor2, true)
		require.NoError(t, err)

		pets := parent.Get(petsField.Name).([]*entity.Entity)
		assert.Len(t, pets, 2)
		assert.Equal(t, neighbor1, pets[0])
		assert.Equal(t, neighbor2, pets[1])
	})

	t.Run("error when existing value is not entity array", func(t *testing.T) {
		parent := entity.New(1).Set("name", "John").Set(petsField.Name, "invalid")
		neighbor := entity.New(10).Set("name", "Fido")

		err := loader.assignToParent(parent, neighbor, true)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "is not []*entity.Entity")
	})
}

// =============================================================================
// buildEdgeColumns unit tests
// =============================================================================

func TestBuildEdgeColumnsExtended(t *testing.T) {
	adapter := createMockAdapter(t)
	petModel, err := adapter.Model("pet")
	require.NoError(t, err)
	petModelEnt, ok := petModel.(*Model)
	require.True(t, ok)

	tests := []struct {
		name            string
		edgeColumns     []string
		selectFullEdge  bool
		requiredColumns []string
		wantDirect      []string
		wantNested      []string
		wantRelation    []string
		wantErr         bool
	}{
		{
			name:            "empty edgeColumns with selectFullEdge true",
			edgeColumns:     []string{},
			selectFullEdge:  true,
			requiredColumns: []string{"id"},
			wantDirect:      nil,
			wantNested:      []string{},
			wantRelation:    []string{},
		},
		{
			name:            "specific columns",
			edgeColumns:     []string{"name", "id"},
			selectFullEdge:  false,
			requiredColumns: []string{"id"},
			wantDirect:      []string{"name", "id"},
			wantNested:      []string{},
			wantRelation:    []string{},
		},
		{
			name:            "columns with nested field",
			edgeColumns:     []string{"name", "owner.name"},
			selectFullEdge:  false,
			requiredColumns: []string{"id", "owner_id"},
			wantDirect:      []string{"name", "id", "owner_id"},
			wantNested:      []string{"owner.name"},
			wantRelation:    []string{},
		},
		{
			name:            "columns with relation field",
			edgeColumns:     []string{"name", "owner"},
			selectFullEdge:  false,
			requiredColumns: []string{"id", "owner_id"},
			wantDirect:      []string{"name", "id", "owner_id"},
			wantNested:      []string{},
			wantRelation:    []string{"owner"},
		},
		{
			name:            "mixed columns",
			edgeColumns:     []string{"name", "owner", "owner.age"},
			selectFullEdge:  false,
			requiredColumns: []string{"id"},
			wantDirect:      []string{"name", "id"},
			wantNested:      []string{"owner.age"},
			wantRelation:    []string{"owner"},
		},
		{
			name:            "duplicate required columns",
			edgeColumns:     []string{"name", "id"},
			selectFullEdge:  false,
			requiredColumns: []string{"id", "id"},
			wantDirect:      []string{"name", "id"},
			wantNested:      []string{},
			wantRelation:    []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := buildEdgeColumns(petModelEnt, tt.edgeColumns, tt.selectFullEdge, tt.requiredColumns)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)

			if tt.selectFullEdge {
				assert.Nil(t, result.directColumns)
			} else if tt.wantDirect != nil {
				for _, col := range tt.wantDirect {
					assert.Contains(t, result.directColumns, col)
				}
			}
			assert.ElementsMatch(t, tt.wantNested, result.nestedFields)
			assert.ElementsMatch(t, tt.wantRelation, result.relationFields)
		})
	}
}

// =============================================================================
// collectParentRefs tests
// =============================================================================

func TestCollectParentRefs(t *testing.T) {
	idField := &schema.Field{
		Name: "id",
		Type: schema.TypeUint64,
	}

	fkField := &schema.Field{
		Name: "owner_id",
		Type: schema.TypeUint64,
	}

	t.Run("collect primary key refs", func(t *testing.T) {
		entities := []*entity.Entity{
			entity.New(1),
			entity.New(2),
			entity.New(3),
		}

		refs, parentMap, err := collectParentRefs(entities, "id", idField, "user", false)
		require.NoError(t, err)
		assert.Len(t, refs, 3)
		assert.Len(t, parentMap, 3)
		assert.Contains(t, parentMap, "uint64:1")
		assert.Contains(t, parentMap, "uint64:2")
		assert.Contains(t, parentMap, "uint64:3")
	})

	t.Run("collect FK refs skipping nulls", func(t *testing.T) {
		entities := []*entity.Entity{
			entity.New(1).Set("owner_id", uint64(10)),
			entity.New(2), // no owner_id
			entity.New(3).Set("owner_id", uint64(20)),
		}

		refs, parentMap, err := collectParentRefs(entities, "owner_id", fkField, "pet", true)
		require.NoError(t, err)
		assert.Len(t, refs, 2)
		assert.Len(t, parentMap, 2)
		assert.Contains(t, parentMap, "uint64:10")
		assert.Contains(t, parentMap, "uint64:20")
	})

	t.Run("multiple entities with same FK", func(t *testing.T) {
		entities := []*entity.Entity{
			entity.New(1).Set("owner_id", uint64(10)),
			entity.New(2).Set("owner_id", uint64(10)),
			entity.New(3).Set("owner_id", uint64(20)),
		}

		refs, parentMap, err := collectParentRefs(entities, "owner_id", fkField, "pet", true)
		require.NoError(t, err)
		// Should only have 2 unique refs
		assert.Len(t, refs, 2)
		assert.Len(t, parentMap, 2)
		// But parentMap[10] should have 2 entities
		assert.Len(t, parentMap["uint64:10"], 2)
	})
}

// =============================================================================
// collectEntityIDs tests
// =============================================================================

func TestCollectEntityIDs(t *testing.T) {
	idField := &schema.Field{
		Name: "id",
		Type: schema.TypeUint64,
	}

	t.Run("collect IDs from entities", func(t *testing.T) {
		entities := []*entity.Entity{
			entity.New(1),
			entity.New(2),
			entity.New(3),
		}

		ids, entityMap, err := collectEntityIDs("user", idField, entities)
		require.NoError(t, err)
		assert.Len(t, ids, 3)
		assert.Len(t, entityMap, 3)
	})

	t.Run("error for entity without ID", func(t *testing.T) {
		entities := []*entity.Entity{
			entity.New(1),
			entity.New(), // no ID - this should cause an error
			entity.New(3),
		}

		_, _, err := collectEntityIDs("user", idField, entities)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid id")
	})

	t.Run("empty entities", func(t *testing.T) {
		ids, entityMap, err := collectEntityIDs("user", idField, []*entity.Entity{})
		require.NoError(t, err)
		assert.Len(t, ids, 0)
		assert.Len(t, entityMap, 0)
	})
}

// =============================================================================
// Edge loading error scenarios
// =============================================================================

func TestEdgeLoadingErrors(t *testing.T) {
	sb := createSchemaBuilder()

	tests := []MockTestQueryData{
		{
			Name:    "Edge_loading_query_error",
			Schema:  "user",
			Columns: []string{"name", "pets"},
			Expect: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(utils.EscapeQuery("SELECT `users`.`id`, `users`.`name` FROM `users`")).
					WillReturnRows(mock.NewRows([]string{"id", "name"}).
						AddRow(testUserUUID1, "John"))
				mock.ExpectQuery(utils.EscapeQuery("SELECT * FROM `pets` WHERE `pets`.`owner_id` IN (?)")).
					WithArgs(testUserUUID1).
					WillReturnError(errors.New("database error"))
			},
			ExpectError: "database error",
		},
	}

	MockRunQueryTests(func(d *sql.DB) db.Client {
		client := utils.Must(NewEntClient(&db.Config{
			Driver: "sqlmock",
		}, sb, dialectSql.OpenDB(dialect.MySQL, d)))
		return client
	}, sb, t, tests)
}

// =============================================================================
// Error helper function tests
// =============================================================================

func TestInvalidFKErrorMessage(t *testing.T) {
	err := invalidFKError("pets", "owner_id", uint64(123), errors.New("some error"))
	assert.EqualError(t, err, "invalid FK value pets.owner_id for node id=123: some error")
}

func TestNoFKNodeErrorMessage(t *testing.T) {
	err := noFKNodeError("user", "pet", "owner_id", uint64(123), uint64(456))
	assert.EqualError(t, err, "no FK node (user) found for (pet=123).owner_id=456")
}

func TestInvalidEntityArrayErrorMessage(t *testing.T) {
	err := invalidEntityArrayError("user", "pets", "invalid")
	assert.Contains(t, err.Error(), "edge values user.pets=invalid")
	assert.Contains(t, err.Error(), "is not []*entity.Entity")
}

// =============================================================================
// valueKey function tests
// =============================================================================

func TestValueKey(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected string
	}{
		{"uint64", uint64(123), "uint64:123"},
		{"int", 123, "int:123"},
		{"string", "abc", "string:abc"},
		{"nil", nil, "<nil>"},
		{"float64", 123.456, "float64:123.456"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := valueKey(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// =============================================================================
// isZeroValue function tests
// =============================================================================

func TestIsZeroValue(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected bool
	}{
		{"nil", nil, true},
		{"zero uint64", uint64(0), true},
		{"non-zero uint64", uint64(1), false},
		{"zero int", 0, true},
		{"non-zero int", 1, false},
		{"empty string", "", true},
		{"non-empty string", "abc", false},
		{"zero float64", float64(0), true},
		{"non-zero float64", float64(1.5), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isZeroValue(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

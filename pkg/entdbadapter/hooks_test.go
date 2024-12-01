package entdbadapter

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/entity"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
	"github.com/stretchr/testify/assert"
)

func TestHooksError(t *testing.T) {
	migrationDir := utils.Must(os.MkdirTemp("", "migrations"))
	sb := createSchemaBuilder()

	client, err := NewTestClient(migrationDir, sb, func() *db.Hooks {
		return &db.Hooks{
			PreDBQuery: []db.PreDBQuery{
				func(ctx context.Context, option *db.QueryOption) error {
					return assert.AnError
				},
			},
			PostDBQuery: []db.PostDBQuery{
				func(
					ctx context.Context,
					option *db.QueryOption,
					entities []*entity.Entity,
				) ([]*entity.Entity, error) {
					return nil, assert.AnError
				},
			},
			PreDBExec: []db.PreDBExec{
				func(ctx context.Context, option *db.QueryOption) error {
					return assert.AnError
				},
			},
			PostDBExec: []db.PostDBExec{
				func(ctx context.Context, option *db.QueryOption, result sql.Result) error {
					return assert.AnError
				},
			},
			PreDBCreate: []db.PreDBCreate{
				func(ctx context.Context, schema *schema.Schema, createData *entity.Entity) error {
					return assert.AnError
				},
			},
			PostDBCreate: []db.PostDBCreate{
				func(ctx context.Context, schema *schema.Schema, createData *entity.Entity, id uint64) error {
					return assert.AnError
				},
			},
			PreDBUpdate: []db.PreDBUpdate{
				func(
					ctx context.Context,
					schema *schema.Schema,
					predicates []*db.Predicate,
					updateData *entity.Entity,
				) error {
					return assert.AnError
				},
			},
			PostDBUpdate: []db.PostDBUpdate{
				func(
					ctx context.Context,
					schema *schema.Schema,
					predicates []*db.Predicate,
					updateData *entity.Entity,
					originalEntities []*entity.Entity,
					affected int,
				) error {
					return assert.AnError
				},
			},
			PreDBDelete: []db.PreDBDelete{
				func(ctx context.Context, schema *schema.Schema, predicates []*db.Predicate) error {
					return assert.AnError
				},
			},
			PostDBDelete: []db.PostDBDelete{
				func(
					ctx context.Context,
					schema *schema.Schema,
					predicates []*db.Predicate,
					originalEntities []*entity.Entity,
					affected int,
				) error {
					return assert.AnError
				},
			},
		}
	})
	assert.NoError(t, err)

	ctx := context.Background()

	// Hooks error
	err = runPreDBQueryHooks(ctx, client, &db.QueryOption{})
	assert.Error(t, err)

	_, err = runPostDBQueryHooks(ctx, client, &db.QueryOption{}, nil)
	assert.Error(t, err)

	err = runPreDBExecHooks(ctx, client, &db.QueryOption{})
	assert.Error(t, err)

	err = runPostDBExecHooks(ctx, client, &db.QueryOption{}, nil)
	assert.Error(t, err)

	err = runPreDBCreateHooks(ctx, client, nil, &entity.Entity{})
	assert.Error(t, err)

	err = runPostDBCreateHooks(ctx, client, nil, &entity.Entity{}, 0)
	assert.Error(t, err)

	err = runPreDBUpdateHooks(ctx, client, nil, nil, &entity.Entity{})
	assert.Error(t, err)

	err = runPostDBUpdateHooks(ctx, client, nil, nil, &entity.Entity{}, nil, 0)
	assert.Error(t, err)

	err = runPreDBDeleteHooks(ctx, client, nil, nil)
	assert.Error(t, err)

	err = runPostDBDeleteHooks(ctx, client, nil, nil, nil, 0)
	assert.Error(t, err)

	// Client is nil
	err = runPreDBQueryHooks(ctx, nil, &db.QueryOption{})
	assert.NoError(t, err)

	_, err = runPostDBQueryHooks(ctx, nil, &db.QueryOption{}, nil)
	assert.NoError(t, err)

	err = runPreDBExecHooks(ctx, nil, &db.QueryOption{})
	assert.NoError(t, err)

	err = runPostDBExecHooks(ctx, nil, &db.QueryOption{}, nil)
	assert.NoError(t, err)

	err = runPreDBCreateHooks(ctx, nil, nil, &entity.Entity{})
	assert.NoError(t, err)

	err = runPostDBCreateHooks(ctx, nil, nil, &entity.Entity{}, 0)
	assert.NoError(t, err)

	err = runPreDBUpdateHooks(ctx, nil, nil, nil, &entity.Entity{})
	assert.NoError(t, err)

	err = runPostDBUpdateHooks(ctx, nil, nil, nil, &entity.Entity{}, nil, 0)
	assert.NoError(t, err)

	err = runPreDBDeleteHooks(ctx, nil, nil, nil)
	assert.NoError(t, err)

	err = runPostDBDeleteHooks(ctx, nil, nil, nil, nil, 0)
	assert.NoError(t, err)
}

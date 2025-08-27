package db_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/entity"
	"github.com/fastschema/fastschema/pkg/entdbadapter"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
	"github.com/stretchr/testify/assert"
)

type TestCategory struct {
	_         any        `fs:"name=category;namespace=categories"`
	ID        uint64     `json:"id,omitempty"`
	Name      string     `json:"name"`
	CreatedAt *time.Time `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
}

type testPost struct {
	_     any    `fs:"namespace=posts"`
	Title string `json:"title"`
}

func prepareTest() (db.Client, context.Context) {
	sb := utils.Must(schema.NewBuilderFromDir(
		utils.Must(os.MkdirTemp("", "schemas")),
		TestCategory{},
	))
	client := utils.Must(entdbadapter.NewTestClient(
		utils.Must(os.MkdirTemp("", "migrations")),
		sb,
	))

	return client, context.Background()
}

func TestDBConfigClone(t *testing.T) {
	c := &db.Config{
		Driver:       "mysql",
		Name:         "mydb",
		Host:         "localhost",
		Port:         "3306",
		User:         "root",
		Pass:         "password",
		Logger:       nil,
		LogQueries:   true,
		MigrationDir: "/path/to/migrations",
	}

	clone := c.Clone()
	assert.Equal(t, c.Driver, clone.Driver)
	assert.Equal(t, c.Name, clone.Name)
	assert.Equal(t, c.Host, clone.Host)
	assert.Equal(t, c.Port, clone.Port)
	assert.Equal(t, c.User, clone.User)
	assert.Equal(t, c.Pass, clone.Pass)
	assert.Equal(t, c.Logger, clone.Logger)
	assert.Equal(t, c.LogQueries, clone.LogQueries)
	assert.Equal(t, c.MigrationDir, clone.MigrationDir)
}

func TestHooksClone(t *testing.T) {
	hooks := &db.Hooks{
		PostDBQuery: []db.PostDBQuery{func(ctx context.Context, option *db.QueryOption, entities []*entity.Entity) ([]*entity.Entity, error) {
			return nil, nil
		}},
		PostDBCreate: []db.PostDBCreate{func(ctx context.Context, schema *schema.Schema, dataCreate *entity.Entity, id uint64) error {
			return nil
		}},
		PostDBUpdate: []db.PostDBUpdate{func(ctx context.Context, schema *schema.Schema, predicates *[]*db.Predicate, updateData *entity.Entity, originalEntities []*entity.Entity, affected int) error {
			return nil
		}},
		PostDBDelete: []db.PostDBDelete{func(ctx context.Context, schema *schema.Schema, predicates *[]*db.Predicate, originalEntities []*entity.Entity, affected int) error {
			return nil
		}},
		PreDBQuery:  []db.PreDBQuery{func(ctx context.Context, option *db.QueryOption) error { return nil }},
		PreDBCreate: []db.PreDBCreate{func(ctx context.Context, schema *schema.Schema, createData *entity.Entity) error { return nil }},
		PreDBUpdate: []db.PreDBUpdate{func(ctx context.Context, schema *schema.Schema, predicates *[]*db.Predicate, updateData *entity.Entity) error {
			return nil
		}},
		PreDBDelete: []db.PreDBDelete{func(ctx context.Context, schema *schema.Schema, predicates *[]*db.Predicate) error { return nil }},
	}

	clonedHooks := hooks.Clone()

	assert.Equal(t, len(hooks.PostDBQuery), len(clonedHooks.PostDBQuery))
	assert.Equal(t, len(hooks.PostDBCreate), len(clonedHooks.PostDBCreate))
	assert.Equal(t, len(hooks.PostDBUpdate), len(clonedHooks.PostDBUpdate))
	assert.Equal(t, len(hooks.PostDBDelete), len(clonedHooks.PostDBDelete))
	assert.Equal(t, len(hooks.PreDBQuery), len(clonedHooks.PreDBQuery))
	assert.Equal(t, len(hooks.PreDBCreate), len(clonedHooks.PreDBCreate))
	assert.Equal(t, len(hooks.PreDBUpdate), len(clonedHooks.PreDBUpdate))
	assert.Equal(t, len(hooks.PreDBDelete), len(clonedHooks.PreDBDelete))
}

package entdbadapter

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/schema"
)

func runPreDBQueryHooks(ctx context.Context, client db.Client, option *db.QueryOption) error {
	if client == nil {
		return nil
	}

	hooks := client.Hooks()
	if len(hooks.PreDBQuery) > 0 {
		for _, hook := range hooks.PreDBQuery {
			if err := hook(ctx, option); err != nil {
				return fmt.Errorf("pre query hook: %w", err)
			}
		}
	}

	return nil
}

func runPostDBQueryHooks(
	ctx context.Context,
	client db.Client,
	option *db.QueryOption,
	entities []*schema.Entity,
) (_ []*schema.Entity, err error) {
	if client == nil {
		return entities, nil
	}

	hooks := client.Hooks()
	if len(hooks.PostDBQuery) > 0 {
		for _, hook := range hooks.PostDBQuery {
			entities, err = hook(ctx, option, entities)
			if err != nil {
				return nil, fmt.Errorf("post query hook: %w", err)
			}
		}
	}

	return entities, nil
}

func runPreDBExecHooks(ctx context.Context, client db.Client, option *db.QueryOption) error {
	if client == nil {
		return nil
	}

	hooks := client.Hooks()
	if len(hooks.PreDBExec) > 0 {
		for _, hook := range hooks.PreDBExec {
			if err := hook(ctx, option); err != nil {
				return fmt.Errorf("pre exec hook: %w", err)
			}
		}
	}

	return nil
}

func runPostDBExecHooks(
	ctx context.Context,
	client db.Client,
	option *db.QueryOption,
	result sql.Result,
) error {
	if client == nil {
		return nil
	}

	hooks := client.Hooks()
	if len(hooks.PostDBExec) > 0 {
		for _, hook := range hooks.PostDBExec {
			if err := hook(ctx, option, result); err != nil {
				return fmt.Errorf("post exec hook: %w", err)
			}
		}
	}

	return nil
}

func runPreDBCreateHooks(
	ctx context.Context,
	client db.Client,
	schema *schema.Schema,
	createData *schema.Entity,
) error {
	if client == nil {
		return nil
	}

	hooks := client.Hooks()
	if len(hooks.PreDBCreate) > 0 {
		for _, hook := range hooks.PreDBCreate {
			if err := hook(ctx, schema, createData); err != nil {
				return fmt.Errorf("pre create hook: %w", err)
			}
		}
	}

	return nil
}

func runPostDBCreateHooks(
	ctx context.Context,
	client db.Client,
	schema *schema.Schema,
	createData *schema.Entity,
	createdID uint64,
) error {
	if client == nil {
		return nil
	}

	hooks := client.Hooks()
	if len(hooks.PostDBCreate) > 0 {
		for _, hook := range hooks.PostDBCreate {
			if err := hook(ctx, schema, createData, createdID); err != nil {
				return fmt.Errorf("post create hook: %w", err)
			}
		}
	}

	return nil
}

func runPreDBUpdateHooks(
	ctx context.Context,
	client db.Client,
	schema *schema.Schema,
	predicates []*db.Predicate,
	updateData *schema.Entity,
) error {
	if client == nil {
		return nil
	}

	hooks := client.Hooks()
	if len(hooks.PreDBUpdate) > 0 {
		for _, hook := range hooks.PreDBUpdate {
			if err := hook(ctx, schema, predicates, updateData); err != nil {
				return fmt.Errorf("pre update hook: %w", err)
			}
		}
	}

	return nil
}

func runPostDBUpdateHooks(
	ctx context.Context,
	client db.Client,
	schema *schema.Schema,
	predicates []*db.Predicate,
	updateData *schema.Entity,
	originalEntities []*schema.Entity,
	affected int,
) error {
	if client == nil {
		return nil
	}

	hooks := client.Hooks()
	if len(hooks.PostDBUpdate) > 0 {
		for _, hook := range hooks.PostDBUpdate {
			if err := hook(ctx, schema, predicates, updateData, originalEntities, affected); err != nil {
				return fmt.Errorf("post update hook: %w", err)
			}
		}
	}

	return nil
}

func runPreDBDeleteHooks(
	ctx context.Context,
	client db.Client,
	schema *schema.Schema,
	predicates []*db.Predicate,
) error {
	if client == nil {
		return nil
	}

	hooks := client.Hooks()
	if len(hooks.PreDBDelete) > 0 {
		for _, hook := range hooks.PreDBDelete {
			if err := hook(ctx, schema, predicates); err != nil {
				return fmt.Errorf("pre delete hook: %w", err)
			}
		}
	}

	return nil
}

func runPostDBDeleteHooks(
	ctx context.Context,
	client db.Client,
	schema *schema.Schema,
	predicates []*db.Predicate,
	originalEntities []*schema.Entity,
	affected int,
) error {
	if client == nil {
		return nil
	}

	hooks := client.Hooks()
	if len(hooks.PostDBDelete) > 0 {
		for _, hook := range hooks.PostDBDelete {
			if err := hook(ctx, schema, predicates, originalEntities, affected); err != nil {
				return fmt.Errorf("post delete hook: %w", err)
			}
		}
	}

	return nil
}

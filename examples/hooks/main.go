package main

import (
	"context"
	"fmt"
	"log"

	"github.com/fastschema/fastschema"
	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/entity"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/utils"
)

type Blog struct {
	Name string `json:"name"`
	Desc string `json:"desc" fs:"optional"`
}

func main() {
	app := utils.Must(fastschema.New(&fs.Config{
		Port:          "8000",
		SystemSchemas: []any{Blog{}},
	}))

	// Create a blog without description
	ctx := context.Background()
	_ = utils.Must(db.Builder[Blog](app.DB()).Create(ctx, fs.Map{
		"name": "My Blog",
	}))

	// Add a pre resolve hook
	app.OnPreResolve(func(c fs.Context) error {
		// This hook will be executed before resolving the request
		// You can use it to skip, modify the request or to add some extra logic
		// For example, we add a custom value to the request context

		if c.Resource().ID() != "/pre-resolve" {
			return nil
		}

		c.Local("custom", "pre resolve hook")
		return nil
	})

	// Add a post resolve hook
	app.OnPostResolve(func(c fs.Context) error {
		// This hook will be executed after resolving the request
		// You can use it to modify the response or to add some extra logic
		// For example, we modify the response

		if c.Resource().ID() != "/post-resolve" {
			return nil
		}

		result := c.Result()
		result.Data = "This is a custom response from the post resolve hook"
		c.Result(result)
		return nil
	})

	// Add a post db get hook
	app.OnPostDBQuery(func(
		ctx context.Context,
		query *db.QueryOption,
		entities []*entity.Entity,
	) ([]*entity.Entity, error) {
		// This hook will be executed after getting data from the database
		// You can use it to modify the data or to add some extra logic
		// For example, we add a description to the tags

		schemaName := query.Schema.Name
		if schemaName != "blog" {
			return entities, nil
		}

		for _, entity := range entities {
			entity.Set("desc", fmt.Sprintf("Description for %s", entity.Get("name")))
		}

		return entities, nil
	})

	// Add a resource for pre resolve hook
	app.AddResource(fs.Get("/pre-resolve", func(c fs.Context, _ any) (any, error) {
		custom := c.Local("custom").(string)
		return fmt.Sprintf("Custom value from pre resolve hook: %s", custom), nil
	}, &fs.Meta{Public: true}))
	// response: {"data":"This is a custom value from the pre resolve hook"}

	// Add a resource for post resolve hook
	app.AddResource(fs.Get("/post-resolve", func(c fs.Context, _ any) (string, error) {
		return "response from resource", nil
	}, &fs.Meta{Public: true}))
	// response: {"data":"This is a custom response from the post resolve hook"}

	// Add a resource for post db get hook
	app.AddResource(fs.Get("/post-db-get", func(c fs.Context, _ any) ([]Blog, error) {
		return db.Builder[Blog](app.DB()).Get(c)
	}, &fs.Meta{Public: true}))

	log.Fatal(app.Start())
}

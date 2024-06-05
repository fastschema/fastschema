package main

import (
	"context"
	"fmt"

	"github.com/davecgh/go-spew/spew"
	"github.com/fastschema/fastschema"
	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
)

// In this example, we use utils.Must to handle errors.
// You can use the standard error handling if you prefer.
func main() {
	ctx := context.Background()
	app := utils.Must(fastschema.New(&fs.Config{
		Port:          "8000",
		SystemSchemas: []any{Tag{}, Blog{}},
		DBConfig: &db.Config{
			Driver:     "sqlite",
			LogQueries: false,
		},
	}))

	defer func() {
		utils.Must(0, app.Shutdown())
	}()

	// Delete all relations
	utils.Must(db.Exec(ctx, app.DB(), "DELETE FROM blogs_tags"))

	// Delete all tags
	utils.Must(db.Exec(ctx, app.DB(), "DELETE FROM tags; DELETE FROM SQLITE_SEQUENCE WHERE name='tags';"))

	// Delete all blogs
	utils.Must(db.Exec(ctx, app.DB(), "DELETE FROM blogs; DELETE FROM SQLITE_SEQUENCE WHERE name='blogs';"))

	// Create tag1 using system schema
	tag1 := utils.Must(db.Mutation[Tag](app.DB()).Create(ctx, fs.Map{
		"name": "Technology",
		"desc": "Technology related blogs",
	}))
	fmt.Printf("> Create tag using system schema: %+s\n\n", spew.Sdump(tag1))

	// Create tag using schema name
	tag2 := utils.Must(db.Mutation[*schema.Entity](app.DB(), "tag").Create(ctx, fs.Map{
		"name": "Science",
		"desc": "Science related blogs",
	}))
	fmt.Printf("> Create tag using schema name: %s\n\n", tag2)

	// Create tag using exec
	result := utils.Must(db.Exec(
		ctx, app.DB(),
		"INSERT INTO tags (name, desc) VALUES ($1, $2)",
		"Health", "Health related blogs",
	))
	fmt.Printf("> Create tag using exec, last insert ID: %d\n\n", utils.Must(result.LastInsertId()))

	// Create blog with tags
	blog1 := utils.Must(db.Mutation[Blog](app.DB()).Create(ctx, fs.Map{
		"title": "Blog 1",
		"body":  "Blog 1 body",
		"tags": []*schema.Entity{
			schema.NewEntity(tag1.ID),
			tag2,
		},
	}))
	fmt.Printf("> Create blog with tags: %s\n\n", spew.Sdump(blog1))

	// Query blog with tags
	blog1 = utils.Must(db.Query[Blog](app.DB()).
		Where(db.EQ("id", blog1.ID)).
		Select("tags").
		First(ctx))
	fmt.Printf("> Query blog with tags: %+v\n\n", spew.Sdump(blog1))

	// Raw query
	blog1Tags := utils.Must(db.RawQuery[*schema.Entity](
		ctx, app.DB(),
		"SELECT `t1`.`blogs` AS blogs_id, `tags`.`id`, `tags`.`name`, `tags`.`desc`, `tags`.`created_at`, `tags`.`updated_at`, `tags`.`deleted_at` FROM `tags` JOIN `blogs_tags` AS `t1` ON `t1`.`tags` = `tags`.`id` WHERE `t1`.`blogs` IN (?)",
		blog1.ID,
	))
	fmt.Printf("> Raw query: %s\n\n", spew.Sdump(blog1Tags))

	// Update blog
	updatedBlog1 := utils.Must(db.Mutation[Blog](app.DB()).
		Where(db.EQ("id", blog1.ID)).
		Update(ctx, fs.Map{
			"vote": 10,
		}))
	fmt.Printf("> Update blog: %s\n\n", spew.Sdump(updatedBlog1))

	// Delete blog
	affected := utils.Must(db.Mutation[Blog](app.DB()).
		Where(db.EQ("id", blog1.ID)).
		Delete(ctx))
	fmt.Printf("> Delete blog, affected: %d\n\n", affected)

	// Verify delete
	_, err := db.Query[Blog](app.DB()).
		Where(db.EQ("id", blog1.ID)).
		First(ctx)
	fmt.Printf("> Verify delete: %t\n\n", db.IsNotFound(err))
}

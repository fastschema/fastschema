package selectoptions

import (
	"fmt"
	"testing"

	u "github.com/fastschema/fastschema/pkg/utils"
	h "github.com/fastschema/fastschema/tests/integration/helpers"
)

func setupTestData(t *testing.T, client h.DBClient) {
	t.Helper()
	h.ClearDBData(
		client.C,
		"posts_tags", "categories_posts", "posts", "tags", "categories", "authors", "countries",
	)

	// Create countries
	countryModel := u.Must(client.C.Model("country"))
	usaID := u.Must(countryModel.CreateFromJSON(
		h.Ctx(),
		`{"name": "USA", "code": "US"}`,
	))
	ukID := u.Must(countryModel.CreateFromJSON(
		h.Ctx(),
		`{"name": "UK", "code": "GB"}`,
	))

	// Create authors with countries
	authorModel := u.Must(client.C.Model("author"))
	author1ID := u.Must(authorModel.CreateFromJSON(
		h.Ctx(),
		fmt.Sprintf(
			`{"name": "Author 1", "bio": "Bio 1", "country": {"id": %d}}`,
			usaID,
		),
	))
	author2ID := u.Must(authorModel.CreateFromJSON(
		h.Ctx(),
		fmt.Sprintf(
			`{"name": "Author 2", "bio": "Bio 2", "country": {"id": %d}}`,
			ukID,
		),
	))

	// Create tags
	tagModel := u.Must(client.C.Model("tag"))
	tag1ID := u.Must(tagModel.CreateFromJSON(
		h.Ctx(),
		`{"name": "Tag A", "priority": 1, "status": "active"}`,
	))
	tag2ID := u.Must(tagModel.CreateFromJSON(
		h.Ctx(),
		`{"name": "Tag B", "priority": 2, "status": "active"}`,
	))
	tag3ID := u.Must(tagModel.CreateFromJSON(
		h.Ctx(),
		`{"name": "Tag C", "priority": 3, "status": "inactive"}`,
	))
	tag4ID := u.Must(tagModel.CreateFromJSON(
		h.Ctx(),
		`{"name": "Tag D", "priority": 4, "status": "active"}`,
	))
	tag5ID := u.Must(tagModel.CreateFromJSON(
		h.Ctx(),
		`{"name": "Tag E", "priority": 5, "status": "inactive"}`,
	))

	// Create categories
	categoryModel := u.Must(client.C.Model("category"))
	cat1ID := u.Must(categoryModel.CreateFromJSON(
		h.Ctx(),
		`{"name": "Cat 1", "priority": 10, "status": "active"}`,
	))
	cat2ID := u.Must(categoryModel.CreateFromJSON(
		h.Ctx(),
		`{"name": "Cat 2", "priority": 20, "status": "active"}`,
	))
	cat3ID := u.Must(categoryModel.CreateFromJSON(
		h.Ctx(),
		`{"name": "Cat 3", "priority": 30, "status": "inactive"}`,
	))

	// Create posts with tags and categories
	postModel := u.Must(client.C.Model("post"))
	u.Must(postModel.CreateFromJSON(
		h.Ctx(),
		fmt.Sprintf(`{
				"title": "Post 1", 
				"content": "Content 1", 
				"priority": 1, 
				"status": "published",
				"author": {"id": %d},
				"tags": [{"id": %d}, {"id": %d}, {"id": %d}, {"id": %d}, {"id": %d}],
				"categories": [{"id": %d}, {"id": %d}]
			}`,
			author1ID, tag1ID, tag2ID, tag3ID, tag4ID, tag5ID, cat1ID, cat2ID,
		),
	))

	u.Must(postModel.CreateFromJSON(
		h.Ctx(),
		fmt.Sprintf(`{
				"title": "Post 2", 
				"content": "Content 2", 
				"priority": 2, 
				"status": "draft",
				"author": {"id": %d},
				"tags": [{"id": %d}, {"id": %d}],
				"categories": [{"id": %d}, {"id": %d}, {"id": %d}]
			}`,
			author2ID, tag1ID, tag3ID, cat1ID, cat2ID, cat3ID,
		),
	))
}

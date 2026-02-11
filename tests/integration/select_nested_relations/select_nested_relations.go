package selectnestedrelations

import (
	"fmt"
	"testing"

	u "github.com/fastschema/fastschema/pkg/utils"
	h "github.com/fastschema/fastschema/tests/integration/helpers"
)

// Helper function to setup base test data
func setupBaseTestData(t *testing.T, client h.DBClient) map[string]any {
	t.Helper()
	h.ClearDBData(client.C,
		"posts_tags",
		"categories_posts",
		"comments",
		"posts",
		"tags",
		"categories",
		"authors",
		"commenters",
	)
	c := h.Ctx()

	// Create commenters
	commenterModel := u.Must(client.C.Model("commenter"))
	commenter1ID := u.Must(commenterModel.CreateFromJSON(
		c,
		`{"name": "Commenter 1", "email": "c1@test.com"}`,
	))
	commenter2ID := u.Must(commenterModel.CreateFromJSON(
		c,
		`{"name": "Commenter 2", "email": "c2@test.com"}`,
	))

	// Create authors
	authorModel := u.Must(client.C.Model("author"))
	author1ID := u.Must(authorModel.CreateFromJSON(
		c,
		`{"name": "Author 1", "bio": "Bio 1"}`,
	))
	author2ID := u.Must(authorModel.CreateFromJSON(
		c,
		`{"name": "Author 2", "bio": "Bio 2"}`,
	))

	// Create tags
	tagModel := u.Must(client.C.Model("tag"))
	tag1ID := u.Must(tagModel.CreateFromJSON(
		c,
		`{"name": "Tag 1", "priority": 1}`,
	))
	tag2ID := u.Must(tagModel.CreateFromJSON(
		c,
		`{"name": "Tag 2", "priority": 2}`,
	))
	tag3ID := u.Must(tagModel.CreateFromJSON(
		c,
		`{"name": "Tag 3", "priority": 3}`,
	))

	// Create categories (with parent-child relationship)
	categoryModel := u.Must(client.C.Model("category"))
	parentCatID := u.Must(categoryModel.CreateFromJSON(
		c,
		`{"name": "Parent Category", "description": "Parent", "priority": 1}`,
	))
	cat1ID := u.Must(categoryModel.CreateFromJSON(
		c,
		fmt.Sprintf(
			`{"name": "Category 1", "description": "Cat 1", "priority": 10, "parent": {"id": %v}}`,
			parentCatID,
		),
	))
	cat2ID := u.Must(categoryModel.CreateFromJSON(
		c,
		fmt.Sprintf(
			`{"name": "Category 2", "description": "Cat 2", "priority": 20, "parent": {"id": %v}}`,
			parentCatID,
		),
	))
	cat3ID := u.Must(categoryModel.CreateFromJSON(
		c,
		`{"name": "Category 3", "description": "Cat 3 - empty", "priority": 30}`,
	))

	// Create posts
	postModel := u.Must(client.C.Model("post"))
	// Post 1: belongs to both Cat1 and Cat2, has multiple tags
	post1ID := u.Must(postModel.CreateFromJSON(c, fmt.Sprintf(`{
		"title": "Post 1",
		"content": "Content 1",
		"priority": 1,
		"author": {"id": %v},
		"tags": [{"id": %v}, {"id": %v}],
		"categories": [{"id": %v}, {"id": %v}]
	}`, author1ID, tag1ID, tag2ID, cat1ID, cat2ID)))

	// Post 2: belongs to Cat1 only
	post2ID := u.Must(postModel.CreateFromJSON(c, fmt.Sprintf(`{
		"title": "Post 2",
		"content": "Content 2",
		"priority": 2,
		"author": {"id": %v},
		"tags": [{"id": %v}],
		"categories": [{"id": %v}]
	}`, author1ID, tag3ID, cat1ID)))

	// Post 3: belongs to Cat2 only, no comments
	post3ID := u.Must(postModel.CreateFromJSON(c, fmt.Sprintf(`{
		"title": "Post 3",
		"content": "Content 3",
		"priority": 3,
		"author": {"id": %v},
		"tags": [{"id": %v}, {"id": %v}],
		"categories": [{"id": %v}]
	}`, author2ID, tag1ID, tag3ID, cat2ID)))

	// Create comments for posts
	commentModel := u.Must(client.C.Model("comment"))
	// Comments for Post 1
	comment1ID := u.Must(commentModel.CreateFromJSON(
		c,
		fmt.Sprintf(
			`{"content": "Comment 1 on Post 1", "rating": 5, "post": {"id": %v}, "commenter": {"id": %v}}`,
			post1ID, commenter1ID,
		),
	))
	comment2ID := u.Must(commentModel.CreateFromJSON(
		c,
		fmt.Sprintf(
			`{"content": "Comment 2 on Post 1", "rating": 4, "post": {"id": %v}, "commenter": {"id": %v}}`,
			post1ID, commenter2ID,
		),
	))
	comment3ID := u.Must(commentModel.CreateFromJSON(
		c,
		fmt.Sprintf(
			`{"content": "Comment 3 on Post 1", "rating": 3, "post": {"id": %v}, "commenter": {"id": %v}}`,
			post1ID, commenter1ID,
		),
	))

	// Comments for Post 2
	comment4ID := u.Must(commentModel.CreateFromJSON(
		c,
		fmt.Sprintf(
			`{"content": "Comment 1 on Post 2", "rating": 5, "post": {"id": %v}, "commenter": {"id": %v}}`,
			post2ID, commenter2ID,
		),
	))

	return map[string]any{
		"commenter1ID": commenter1ID,
		"commenter2ID": commenter2ID,
		"author1ID":    author1ID,
		"author2ID":    author2ID,
		"tag1ID":       tag1ID,
		"tag2ID":       tag2ID,
		"tag3ID":       tag3ID,
		"parentCatID":  parentCatID,
		"cat1ID":       cat1ID,
		"cat2ID":       cat2ID,
		"cat3ID":       cat3ID,
		"post1ID":      post1ID,
		"post2ID":      post2ID,
		"post3ID":      post3ID,
		"comment1ID":   comment1ID,
		"comment2ID":   comment2ID,
		"comment3ID":   comment3ID,
		"comment4ID":   comment4ID,
	}
}

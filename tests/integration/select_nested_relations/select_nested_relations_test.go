package selectnestedrelations

import (
	"testing"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/entity"
	u "github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
	h "github.com/fastschema/fastschema/tests/integration/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	schemaDir    = "../../../tests/integration/select_nested_relations/data/schemas"
	migrationDir = "../../../tests/integration/select_nested_relations/data/migrations"
	sqliteDSN    = "../../../tests/integration/select_nested_relations/data/nested_relations_test.db"
)

func TestMySQL(t *testing.T) {
	runNestedRelationTests(t, u.Map(h.MysqlConfigs, func(cfg h.DBConfig) h.DBClient {
		sb := u.Must(schema.NewBuilderFromDir(schemaDir))
		return h.NewMySQLClient(t, cfg, sb, migrationDir)
	}))
}

func TestPostgres(t *testing.T) {
	runNestedRelationTests(t, u.Map(h.PostgresConfigs, func(cfg h.DBConfig) h.DBClient {
		sb := u.Must(schema.NewBuilderFromDir(schemaDir))
		return h.NewPostgresClient(t, cfg, sb, migrationDir)
	}))
}

func TestSQLite(t *testing.T) {
	sb := u.Must(schema.NewBuilderFromDir(schemaDir))
	client := h.NewSQLiteClient(t, "sqlite", sqliteDSN, migrationDir, sb)
	runNestedRelationTests(t, []h.DBClient{client})
}

func runNestedRelationTests(t *testing.T, clients []h.DBClient) {
	for _, client := range clients {
		client := client
		t.Run(client.Name, func(t *testing.T) {
			// Test Case 1: M2M -> O2M chain (category.posts.comments)
			// When same post appears in multiple categories, comments should load for all instances
			t.Run("M2M_O2M_DuplicateEntityWithNestedRelation", func(t *testing.T) {
				testM2M_O2M_DuplicateEntityWithNestedRelation(t, client)
			})

			// Test Case 2: Deep nesting (3+ levels)
			// category.posts.comments.commenter
			t.Run("DeepNesting_ThreeLevels", func(t *testing.T) {
				testDeepNesting_ThreeLevels(t, client)
			})

			// Test Case 3: Multiple nested relations at same level
			// category.posts.comments, category.posts.tags
			t.Run("MultipleNestedRelationsAtSameLevel", func(t *testing.T) {
				testMultipleNestedRelationsAtSameLevel(t, client)
			})

			// Test Case 4: O2M -> M2M chain (author.posts.tags)
			t.Run("O2M_M2M_Chain", func(t *testing.T) {
				testO2M_M2M_Chain(t, client)
			})

			// Test Case 5: Non-owner side with nested relations
			// comment.post.categories
			t.Run("NonOwnerSideNestedRelations", func(t *testing.T) {
				testNonOwnerSideNestedRelations(t, client)
			})

			// Test Case 6: Empty nested relations
			// Should return empty arrays, not nil
			t.Run("EmptyNestedRelations", func(t *testing.T) {
				testEmptyNestedRelations(t, client)
			})

			// Test Case 7: Circular reference (category.posts.categories)
			// Should not cause infinite loop
			t.Run("CircularReference", func(t *testing.T) {
				testCircularReference(t, client)
			})

			// Test Case 8: Self-referencing relation with nesting
			// category.children.posts
			t.Run("SelfReferencingWithNesting", func(t *testing.T) {
				testSelfReferencingWithNesting(t, client)
			})

			// Test Case 9: Select specific fields from nested relations
			// posts.comments.content (only content field)
			t.Run("SelectSpecificFieldsFromNestedRelation", func(t *testing.T) {
				testSelectSpecificFieldsFromNestedRelation(t, client)
			})

			// Test Case 10: M2M -> O2M with multiple parent entities
			// Multiple categories, each with overlapping posts, each post with comments
			t.Run("M2M_O2M_MultipleParentEntities", func(t *testing.T) {
				testM2M_O2M_MultipleParentEntities(t, client)
			})

			// Test Case 11: Non-owner M2M -> O2M chain (tag.posts.comments)
			// Tag is non-owner of M2M with posts
			t.Run("NonOwnerM2M_O2M_Chain", func(t *testing.T) {
				testNonOwnerM2M_O2M_Chain(t, client)
			})

			// Test Case 12: O2M owner -> O2M non-owner chain
			// author.posts.author (back to author through different path)
			t.Run("O2M_Owner_NonOwner_Chain", func(t *testing.T) {
				testO2M_Owner_NonOwner_Chain(t, client)
			})
		})
	}
}

// Test Case 1: M2M -> O2M chain with duplicate entities
// When same post appears in multiple categories, comments should load for all instances
func testM2M_O2M_DuplicateEntityWithNestedRelation(t *testing.T, client h.DBClient) {
	ids := setupBaseTestData(t, client)
	ctx := h.Ctx()
	categoryModel := u.Must(client.C.Model("category"))

	// Post 1 belong to both Cat1 and Cat2

	// Query Cat1: Verify Post 1 is loaded with comments
	results := u.Must(categoryModel.Query().
		Select("name", "posts.title", "posts.comments.content").
		Where(db.EQ("id", ids["cat1ID"])).
		Get(ctx))
	require.Len(t, results, 1)

	// Verify Cat1 has posts with comments
	cat1Posts := results[0].Get("posts").([]*entity.Entity)
	require.GreaterOrEqual(t, len(cat1Posts), 1, "Cat1 should have at least 1 post")

	// Find Post 1 in Cat1's posts
	var post1InCat1 *entity.Entity
	for _, p := range cat1Posts {
		if p.Get("title") == "Post 1" {
			post1InCat1 = p
			break
		}
	}
	require.NotNil(t, post1InCat1, "Post 1 should be in Cat1")
	post1Comments := post1InCat1.Get("comments").([]*entity.Entity)
	assert.Len(t, post1Comments, 3, "Post 1 should have 3 comments in Cat1")

	// Query Cat2: Verify Post 1 is loaded with comments
	results2 := u.Must(categoryModel.Query().
		Select("name", "posts.title", "posts.comments.content").
		Where(db.EQ("id", ids["cat2ID"])).
		Get(ctx))
	require.Len(t, results2, 1)

	cat2Posts := results2[0].Get("posts").([]*entity.Entity)
	var post1InCat2 *entity.Entity
	for _, p := range cat2Posts {
		if p.Get("title") == "Post 1" {
			post1InCat2 = p
			break
		}
	}
	require.NotNil(t, post1InCat2, "Post 1 should be in Cat2")
	post1CommentsInCat2 := post1InCat2.Get("comments").([]*entity.Entity)
	assert.Len(t, post1CommentsInCat2, 3, "Post 1 should have 3 comments in Cat2")

	// Query BOTH categories: Verify both categories load Post 1 with comments
	results3 := u.Must(categoryModel.Query().
		Select("name", "posts.title", "posts.comments.content").
		Order("id").
		Get(ctx))
	require.GreaterOrEqual(t, len(results3), 2)

	for _, cat := range results3 {
		catName := cat.Get("name").(string)
		posts := cat.Get("posts")
		if posts == nil {
			continue
		}
		postList := posts.([]*entity.Entity)
		for _, post := range postList {
			if post.Get("title") == "Post 1" {
				comments := post.Get("comments")
				require.NotNil(t, comments, "Post 1 in %s should have comments loaded", catName)
				commentList := comments.([]*entity.Entity)
				assert.Len(t, commentList, 3, "Post 1 in %s should have 3 comments", catName)
			}
		}
	}
}

// Test Case 2: Deep nesting (3+ levels)
func testDeepNesting_ThreeLevels(t *testing.T, client h.DBClient) {
	ids := setupBaseTestData(t, client)
	ctx := h.Ctx()
	categoryModel := u.Must(client.C.Model("category"))

	// category.posts.comments.commenter (4 levels)
	results := u.Must(categoryModel.Query().
		Select("name", "posts.title", "posts.comments.content", "posts.comments.commenter.name").
		Where(db.EQ("id", ids["cat1ID"])).
		Get(ctx))
	require.Len(t, results, 1)

	posts := results[0].Get("posts").([]*entity.Entity)
	require.GreaterOrEqual(t, len(posts), 1)

	// Find post with comments
	var foundCommentWithCommenter bool
	for _, post := range posts {
		comments := post.Get("comments")
		if comments == nil {
			continue
		}
		commentList := comments.([]*entity.Entity)
		for _, comment := range commentList {
			commenter := comment.Get("commenter")
			if commenter != nil {
				commenterEntity := commenter.(*entity.Entity)
				assert.NotEmpty(t, commenterEntity.Get("name"), "Commenter should have name")
				foundCommentWithCommenter = true
			}
		}
	}
	assert.True(t, foundCommentWithCommenter, "Should find at least one comment with commenter loaded")
}

// Test Case 3: Multiple nested relations at same level
func testMultipleNestedRelationsAtSameLevel(t *testing.T, client h.DBClient) {
	ids := setupBaseTestData(t, client)
	ctx := h.Ctx()
	categoryModel := u.Must(client.C.Model("category"))

	// Select both comments and tags from posts
	results := u.Must(categoryModel.Query().
		Select("name", "posts.title", "posts.comments.content", "posts.tags.name").
		Where(db.EQ("id", ids["cat1ID"])).
		Get(ctx))
	require.Len(t, results, 1)

	posts := results[0].Get("posts").([]*entity.Entity)
	require.GreaterOrEqual(t, len(posts), 1)

	// Find Post 1 which has both comments and tags
	var post1 *entity.Entity
	for _, p := range posts {
		if p.Get("title") == "Post 1" {
			post1 = p
			break
		}
	}
	require.NotNil(t, post1, "Should find Post 1")

	// Verify both nested relations are loaded
	comments := post1.Get("comments").([]*entity.Entity)
	assert.Len(t, comments, 3, "Post 1 should have 3 comments")

	tags := post1.Get("tags").([]*entity.Entity)
	assert.Len(t, tags, 2, "Post 1 should have 2 tags")
}

// Test Case 4: O2M -> M2M chain
func testO2M_M2M_Chain(t *testing.T, client h.DBClient) {
	ids := setupBaseTestData(t, client)
	ctx := h.Ctx()
	authorModel := u.Must(client.C.Model("author"))

	// author.posts.tags
	results := u.Must(authorModel.Query().
		Select("name", "posts.title", "posts.tags.name").
		Where(db.EQ("id", ids["author1ID"])).
		Get(ctx))
	require.Len(t, results, 1)

	posts := results[0].Get("posts").([]*entity.Entity)
	require.Len(t, posts, 2, "Author 1 should have 2 posts")

	// Verify tags are loaded for each post
	for _, post := range posts {
		tags := post.Get("tags")
		require.NotNil(t, tags, "Post %s should have tags loaded", post.Get("title"))
		tagList := tags.([]*entity.Entity)
		assert.GreaterOrEqual(t, len(tagList), 1, "Post %s should have at least 1 tag", post.Get("title"))
	}
}

// Test Case 5: Non-owner side with nested relations
func testNonOwnerSideNestedRelations(t *testing.T, client h.DBClient) {
	ids := setupBaseTestData(t, client)
	ctx := h.Ctx()
	commentModel := u.Must(client.C.Model("comment"))

	// comment.post.categories (comment is non-owner of post relation)
	results := u.Must(commentModel.Query().
		Select("content", "post.title", "post.categories.name").
		Where(db.EQ("id", ids["comment1ID"])).
		Get(ctx))
	require.Len(t, results, 1)

	post := results[0].Get("post").(*entity.Entity)
	require.NotNil(t, post, "Comment should have post loaded")
	assert.Equal(t, "Post 1", post.Get("title"))

	categories := post.Get("categories").([]*entity.Entity)
	assert.Len(t, categories, 2, "Post 1 should have 2 categories")
}

// Test Case 6: Empty nested relations
func testEmptyNestedRelations(t *testing.T, client h.DBClient) {
	ids := setupBaseTestData(t, client)
	ctx := h.Ctx()
	categoryModel := u.Must(client.C.Model("category"))

	// Cat3 has no posts, should return empty array
	results := u.Must(categoryModel.Query().
		Select("name", "posts.title", "posts.comments.content").
		Where(db.EQ("id", ids["cat3ID"])).
		Get(ctx))
	require.Len(t, results, 1)

	posts := results[0].Get("posts")
	require.NotNil(t, posts, "posts should not be nil, should be empty array")
	postList := posts.([]*entity.Entity)
	assert.Len(t, postList, 0, "Cat3 should have 0 posts")

	// Post 3 has no comments
	postModel := u.Must(client.C.Model("post"))
	postResults := u.Must(postModel.Query().
		Select("title", "comments.content").
		Where(db.EQ("id", ids["post3ID"])).
		Get(ctx))
	require.Len(t, postResults, 1)

	comments := postResults[0].Get("comments")
	require.NotNil(t, comments, "comments should not be nil, should be empty array")
	commentList := comments.([]*entity.Entity)
	assert.Len(t, commentList, 0, "Post 3 should have 0 comments")
}

// Test Case 7: Circular reference
func testCircularReference(t *testing.T, client h.DBClient) {
	ids := setupBaseTestData(t, client)
	ctx := h.Ctx()
	categoryModel := u.Must(client.C.Model("category"))

	// category.posts.categories (back to categories)
	results := u.Must(categoryModel.Query().
		Select("name", "posts.title", "posts.categories.name").
		Where(db.EQ("id", ids["cat1ID"])).
		Get(ctx))
	require.Len(t, results, 1)

	posts := results[0].Get("posts").([]*entity.Entity)
	require.GreaterOrEqual(t, len(posts), 1)

	// Post 1 belongs to both Cat1 and Cat2, so when querying from Cat1,
	// Post 1's categories should include both
	var post1 *entity.Entity
	for _, p := range posts {
		if p.Get("title") == "Post 1" {
			post1 = p
			break
		}
	}
	require.NotNil(t, post1)

	categories := post1.Get("categories").([]*entity.Entity)
	assert.Len(t, categories, 2, "Post 1 should show 2 categories in circular reference")
}

// Test Case 8: Self-referencing with nesting
func testSelfReferencingWithNesting(t *testing.T, client h.DBClient) {
	ids := setupBaseTestData(t, client)
	ctx := h.Ctx()
	categoryModel := u.Must(client.C.Model("category"))

	// parent category with children that have posts
	results := u.Must(categoryModel.Query().
		Select("name", "children.name", "children.posts.title").
		Where(db.EQ("id", ids["parentCatID"])).
		Get(ctx))
	require.Len(t, results, 1)

	children := results[0].Get("children").([]*entity.Entity)
	assert.Len(t, children, 2, "Parent category should have 2 children")

	// Each child should have posts loaded
	for _, child := range children {
		posts := child.Get("posts")
		require.NotNil(t, posts, "Child category %s should have posts loaded", child.Get("name"))
		postList := posts.([]*entity.Entity)
		assert.GreaterOrEqual(t, len(postList), 1, "Child category %s should have at least 1 post", child.Get("name"))
	}
}

// Test Case 9: Select specific fields from nested relation
func testSelectSpecificFieldsFromNestedRelation(t *testing.T, client h.DBClient) {
	ids := setupBaseTestData(t, client)
	ctx := h.Ctx()
	postModel := u.Must(client.C.Model("post"))

	// Only select content from comments, not other fields
	results := u.Must(postModel.Query().
		Select("title", "comments.content").
		Where(db.EQ("id", ids["post1ID"])).
		Get(ctx))
	require.Len(t, results, 1)

	comments := results[0].Get("comments").([]*entity.Entity)
	require.Len(t, comments, 3)

	for _, comment := range comments {
		// content should be loaded
		assert.NotEmpty(t, comment.Get("content"), "Comment content should be loaded")
		// id should always be loaded
		assert.NotZero(t, comment.ID(), "Comment ID should be loaded")
	}
}

// Test Case 10: M2M -> O2M with multiple parent entities
func testM2M_O2M_MultipleParentEntities(t *testing.T, client h.DBClient) {
	_ = setupBaseTestData(t, client)
	ctx := h.Ctx()
	categoryModel := u.Must(client.C.Model("category"))

	// Query all categories with nested posts.comments
	results := u.Must(categoryModel.Query().
		Select("name", "posts.title", "posts.comments.content").
		Order("priority").
		Get(ctx))

	// We have 4 categories (parent, cat1, cat2, cat3)
	require.GreaterOrEqual(t, len(results), 3)

	// Count how many times Post 1 appears across all categories
	post1Count := 0
	post1CommentsLoadedCount := 0
	for _, cat := range results {
		posts := cat.Get("posts")
		if posts == nil {
			continue
		}
		postList := posts.([]*entity.Entity)
		for _, post := range postList {
			if post.Get("title") == "Post 1" {
				post1Count++
				comments := post.Get("comments")
				if comments != nil {
					commentList := comments.([]*entity.Entity)
					if len(commentList) == 3 {
						post1CommentsLoadedCount++
					}
				}
			}
		}
	}

	// Post 1 should appear in Cat1 and Cat2
	assert.Equal(t, 2, post1Count, "Post 1 should appear in 2 categories")
	// Comments should be loaded for both instances
	assert.Equal(t, 2, post1CommentsLoadedCount, "Post 1 should have comments loaded in both categories")
}

// Test Case 11: Non-owner M2M -> O2M chain
func testNonOwnerM2M_O2M_Chain(t *testing.T, client h.DBClient) {
	ids := setupBaseTestData(t, client)
	ctx := h.Ctx()
	tagModel := u.Must(client.C.Model("tag"))

	// tag.posts.comments (tag is non-owner of M2M with posts)
	results := u.Must(tagModel.Query().
		Select("name", "posts.title", "posts.comments.content").
		Where(db.EQ("id", ids["tag1ID"])).
		Get(ctx))
	require.Len(t, results, 1)

	posts := results[0].Get("posts").([]*entity.Entity)
	// Tag 1 is on Post 1 and Post 3
	require.Len(t, posts, 2, "Tag 1 should have 2 posts")

	// Find Post 1 which has comments
	var post1 *entity.Entity
	for _, p := range posts {
		if p.Get("title") == "Post 1" {
			post1 = p
			break
		}
	}
	require.NotNil(t, post1, "Should find Post 1")

	comments := post1.Get("comments").([]*entity.Entity)
	assert.Len(t, comments, 3, "Post 1 should have 3 comments when accessed through tag")
}

// Test Case 12: O2M owner -> O2M non-owner chain
func testO2M_Owner_NonOwner_Chain(t *testing.T, client h.DBClient) {
	ids := setupBaseTestData(t, client)
	ctx := h.Ctx()
	authorModel := u.Must(client.C.Model("author"))

	// author.posts.author (back to author - should load the author of each post)
	results := u.Must(authorModel.Query().
		Select("name", "posts.title", "posts.author.name").
		Where(db.EQ("id", ids["author1ID"])).
		Get(ctx))
	require.Len(t, results, 1)

	posts := results[0].Get("posts").([]*entity.Entity)
	require.Len(t, posts, 2, "Author 1 should have 2 posts")

	for _, post := range posts {
		author := post.Get("author")
		require.NotNil(t, author, "Post should have author loaded")
		authorEntity := author.(*entity.Entity)
		assert.Equal(t, "Author 1", authorEntity.Get("name"), "Post author should be Author 1")
	}
}

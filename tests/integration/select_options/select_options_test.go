package selectoptions

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
	schemaDir    = "../../../tests/integration/select_options/data/schemas"
	migrationDir = "../../../tests/integration/select_options/data/migrations"
	sqliteDSN    = "../../../tests/integration/select_options/data/select_options_test.db"
)

func TestMySQL(t *testing.T) {
	runSelectOptionsTests(t, u.Map(h.MysqlConfigs, func(cfg h.DBConfig) h.DBClient {
		sb := u.Must(schema.NewBuilderFromDir(schemaDir))
		return h.NewMySQLClient(t, cfg, sb, migrationDir)
	}))
}

func TestPostgres(t *testing.T) {
	runSelectOptionsTests(t, u.Map(h.PostgresConfigs, func(cfg h.DBConfig) h.DBClient {
		sb := u.Must(schema.NewBuilderFromDir(schemaDir))
		return h.NewPostgresClient(t, cfg, sb, migrationDir)
	}))
}

func TestSQLite(t *testing.T) {
	sb := u.Must(schema.NewBuilderFromDir(schemaDir))
	client := h.NewSQLiteClient(t, "sqlite", sqliteDSN, migrationDir, sb)
	runSelectOptionsTests(t, []h.DBClient{client})
}

func runSelectOptionsTests(t *testing.T, clients []h.DBClient) {
	for _, client := range clients {
		client := client
		t.Run(client.Name, func(t *testing.T) {
			t.Run("M2M_Limit", func(t *testing.T) {
				testM2MLimit(t, client)
			})

			t.Run("M2M_Offset", func(t *testing.T) {
				testM2MOffset(t, client)
			})

			t.Run("M2M_Sort", func(t *testing.T) {
				testM2MSort(t, client)
			})

			t.Run("M2M_SortAscending", func(t *testing.T) {
				testM2MSortAscending(t, client)
			})

			t.Run("M2M_Filter", func(t *testing.T) {
				testM2MFilter(t, client)
			})

			t.Run("M2M_Combined", func(t *testing.T) {
				testM2MCombined(t, client)
			})

			t.Run("M2M_LimitWithSort", func(t *testing.T) {
				testM2MLimitWithSort(t, client)
			})

			t.Run("O2M_Limit", func(t *testing.T) {
				testO2MLimit(t, client)
			})

			t.Run("O2M_Offset", func(t *testing.T) {
				testO2MOffset(t, client)
			})

			t.Run("O2M_Sort", func(t *testing.T) {
				testO2MSort(t, client)
			})

			t.Run("O2M_Filter", func(t *testing.T) {
				testO2MFilter(t, client)
			})

			t.Run("M2M_OffsetBeyondCount", func(t *testing.T) {
				testM2MOffsetBeyondCount(t, client)
			})

			t.Run("M2M_FilterNoMatch", func(t *testing.T) {
				testM2MFilterNoMatch(t, client)
			})

			t.Run("M2M_MultipleRelations", func(t *testing.T) {
				testM2MMultipleRelations(t, client)
			})

			t.Run("BackwardCompatibility", func(t *testing.T) {
				testBackwardCompatibility(t, client)
			})
		})
	}
}

func testM2MLimit(t *testing.T, client h.DBClient) {
	setupTestData(t, client)
	postModel := u.Must(client.C.Model("post"))

	relOpts := db.RelationOptions{
		"tags": {Limit: 2},
	}

	results := u.Must(postModel.Query().
		Select("title", "tags.name", "tags.priority").
		WithRelationOptions(relOpts).
		Order("id").
		Get(h.Ctx()))
	require.Len(t, results, 2)

	// Post 1 has 5 tags (Tag A, B, C, D, E with priorities 1, 2, 3, 4, 5)
	// With limit 2, should get only the first 2 tags by default order (id)
	post1Tags := results[0].Get("tags").([]*entity.Entity)
	require.Len(t, post1Tags, 2, "Post 1 should have exactly 2 tags with limit")
	assert.Equal(t, "Tag A", post1Tags[0].Get("name"), "First tag should be Tag A")
	assert.Equal(t, "Tag B", post1Tags[1].Get("name"), "Second tag should be Tag B")

	// Verify the tags have correct priority values
	p1, _ := u.AnyToInt[int](post1Tags[0].Get("priority"))
	p2, _ := u.AnyToInt[int](post1Tags[1].Get("priority"))
	assert.Equal(t, 1, p1, "Tag A should have priority 1")
	assert.Equal(t, 2, p2, "Tag B should have priority 2")

	// Post 2 has 2 tags (Tag A, Tag C)
	// With limit 2, should get both
	post2Tags := results[1].Get("tags").([]*entity.Entity)
	require.Len(t, post2Tags, 2, "Post 2 should have exactly 2 tags")
	assert.Equal(t, "Tag A", post2Tags[0].Get("name"), "First tag should be Tag A")
	assert.Equal(t, "Tag C", post2Tags[1].Get("name"), "Second tag should be Tag C")
}

func testM2MOffset(t *testing.T, client h.DBClient) {
	setupTestData(t, client)
	postModel := u.Must(client.C.Model("post"))

	relOpts := db.RelationOptions{
		"tags": {Offset: 1, Limit: 2},
	}

	results := u.Must(postModel.Query().
		Select("title", "tags.name", "tags.priority").
		WithRelationOptions(relOpts).
		Order("id").
		Get(h.Ctx()))

	require.Len(t, results, 2)

	// Post 1 has 5 tags (Tag A, B, C, D, E)
	// With offset 1 and limit 2, should get tags at positions 1, 2 (Tag B, Tag C)
	post1Tags := results[0].Get("tags").([]*entity.Entity)
	require.Len(t, post1Tags, 2, "Post 1 should have exactly 2 tags after offset")
	assert.Equal(t, "Tag B", post1Tags[0].Get("name"), "First tag after offset should be Tag B")
	assert.Equal(t, "Tag C", post1Tags[1].Get("name"), "Second tag after offset should be Tag C")

	// Verify correct priorities
	p1, _ := u.AnyToInt[int](post1Tags[0].Get("priority"))
	p2, _ := u.AnyToInt[int](post1Tags[1].Get("priority"))
	assert.Equal(t, 2, p1, "Tag B should have priority 2")
	assert.Equal(t, 3, p2, "Tag C should have priority 3")

	// Post 2 has 2 tags (Tag A, Tag C)
	// With offset 1, should get only 1 tag (Tag C)
	post2Tags := results[1].Get("tags").([]*entity.Entity)
	require.Len(t, post2Tags, 1, "Post 2 should have 1 tag after offset")
	assert.Equal(t, "Tag C", post2Tags[0].Get("name"), "Only tag after offset should be Tag C")
}

func testM2MSort(t *testing.T, client h.DBClient) {
	setupTestData(t, client)
	postModel := u.Must(client.C.Model("post"))

	// Test descending sort by priority
	relOpts := db.RelationOptions{
		"tags": {Sort: "-priority"},
	}

	results := u.Must(postModel.Query().
		Select("title", "tags.name", "tags.priority").
		WithRelationOptions(relOpts).
		Order("id").
		Get(h.Ctx()))

	require.Len(t, results, 2)

	// Post 1 has 5 tags with priorities 1, 2, 3, 4, 5
	// Sorted descending, should be: Tag E (5), Tag D (4), Tag C (3), Tag B (2), Tag A (1)
	post1Tags := results[0].Get("tags").([]*entity.Entity)
	require.Len(t, post1Tags, 5)

	expectedNamesDesc := []string{"Tag E", "Tag D", "Tag C", "Tag B", "Tag A"}
	expectedPrioritiesDesc := []int{5, 4, 3, 2, 1}
	for i, tag := range post1Tags {
		assert.Equal(t, expectedNamesDesc[i], tag.Get("name"), "Tag at position %d should be %s", i, expectedNamesDesc[i])
		p, _ := u.AnyToInt[int](tag.Get("priority"))
		assert.Equal(t, expectedPrioritiesDesc[i], p, "Priority at position %d should be %d", i, expectedPrioritiesDesc[i])
	}

	// Post 2 has tags: Tag A (priority 1), Tag C (priority 3)
	// Sorted descending: Tag C, Tag A
	post2Tags := results[1].Get("tags").([]*entity.Entity)
	require.Len(t, post2Tags, 2)
	assert.Equal(t, "Tag C", post2Tags[0].Get("name"))
	assert.Equal(t, "Tag A", post2Tags[1].Get("name"))
}

func testM2MSortAscending(t *testing.T, client h.DBClient) {
	setupTestData(t, client)
	postModel := u.Must(client.C.Model("post"))

	// Test ascending sort by priority (explicit)
	relOpts := db.RelationOptions{
		"tags": {Sort: "priority"},
	}

	results := u.Must(postModel.Query().
		Select("title", "tags.name", "tags.priority").
		WithRelationOptions(relOpts).
		Order("id").
		Get(h.Ctx()))

	require.Len(t, results, 2)

	// Post 1 has 5 tags with priorities 1, 2, 3, 4, 5
	// Sorted ascending, should be: Tag A (1), Tag B (2), Tag C (3), Tag D (4), Tag E (5)
	post1Tags := results[0].Get("tags").([]*entity.Entity)
	require.Len(t, post1Tags, 5)

	expectedNamesAsc := []string{"Tag A", "Tag B", "Tag C", "Tag D", "Tag E"}
	expectedPrioritiesAsc := []int{1, 2, 3, 4, 5}
	for i, tag := range post1Tags {
		assert.Equal(t, expectedNamesAsc[i], tag.Get("name"), "Tag at position %d should be %s", i, expectedNamesAsc[i])
		p, _ := u.AnyToInt[int](tag.Get("priority"))
		assert.Equal(t, expectedPrioritiesAsc[i], p, "Priority at position %d should be %d", i, expectedPrioritiesAsc[i])
	}

	// Post 2 has tags: Tag A (priority 1), Tag C (priority 3)
	// Sorted ascending: Tag A, Tag C
	post2Tags := results[1].Get("tags").([]*entity.Entity)
	require.Len(t, post2Tags, 2)
	assert.Equal(t, "Tag A", post2Tags[0].Get("name"))
	assert.Equal(t, "Tag C", post2Tags[1].Get("name"))
}

func testM2MFilter(t *testing.T, client h.DBClient) {
	setupTestData(t, client)
	postModel := u.Must(client.C.Model("post"))

	relOpts := db.RelationOptions{
		"tags": {Filter: map[string]any{"status": "active"}},
	}

	results := u.Must(postModel.Query().
		Select("title", "tags.name", "tags.status", "tags.priority").
		WithRelationOptions(relOpts).
		Order("id").
		Get(h.Ctx()))

	require.Len(t, results, 2)

	// Post 1 has 5 tags, 3 are active (Tag A, B, D)
	post1Tags := results[0].Get("tags").([]*entity.Entity)
	require.Len(t, post1Tags, 3, "Post 1 should have 3 active tags")

	activeTagNames := []string{"Tag A", "Tag B", "Tag D"}
	for i, tag := range post1Tags {
		assert.Equal(t, activeTagNames[i], tag.Get("name"), "Active tag at position %d should be %s", i, activeTagNames[i])
		assert.Equal(t, "active", tag.Get("status"), "Tag %s should have status 'active'", tag.Get("name"))
	}

	// Post 2 has 2 tags (Tag A active, Tag C inactive)
	// Only Tag A should be returned
	post2Tags := results[1].Get("tags").([]*entity.Entity)
	require.Len(t, post2Tags, 1, "Post 2 should have 1 active tag")
	assert.Equal(t, "Tag A", post2Tags[0].Get("name"))
	assert.Equal(t, "active", post2Tags[0].Get("status"))
}

func testM2MCombined(t *testing.T, client h.DBClient) {
	setupTestData(t, client)
	postModel := u.Must(client.C.Model("post"))

	// Combined: filter active, sort by priority desc, offset 1, limit 2
	relOpts := db.RelationOptions{
		"tags": {
			Filter: map[string]any{"status": "active"},
			Sort:   "-priority",
			Limit:  2,
			Offset: 1,
		},
	}

	results := u.Must(postModel.Query().
		Select("title", "tags.name", "tags.priority", "tags.status").
		WithRelationOptions(relOpts).
		Order("id").
		Get(h.Ctx()))

	require.Len(t, results, 2)

	// Post 1 active tags: Tag A (1), Tag B (2), Tag D (4)
	// Sorted desc by priority: Tag D (4), Tag B (2), Tag A (1)
	// After offset 1, limit 2: Tag B (2), Tag A (1)
	post1Tags := results[0].Get("tags").([]*entity.Entity)
	require.Len(t, post1Tags, 2, "Post 1 should have 2 tags after combined options")

	assert.Equal(t, "Tag B", post1Tags[0].Get("name"), "First tag should be Tag B")
	assert.Equal(t, "Tag A", post1Tags[1].Get("name"), "Second tag should be Tag A")

	for _, tag := range post1Tags {
		assert.Equal(t, "active", tag.Get("status"), "All tags should be active")
	}

	// Verify sort order is correct (priorities should be descending: 2, 1)
	p1, _ := u.AnyToInt[int](post1Tags[0].Get("priority"))
	p2, _ := u.AnyToInt[int](post1Tags[1].Get("priority"))
	assert.Equal(t, 2, p1, "Tag B should have priority 2")
	assert.Equal(t, 1, p2, "Tag A should have priority 1")
	assert.Greater(t, p1, p2, "Tags should be sorted by priority descending")

	// Post 2 has only 1 active tag (Tag A), so with offset 1, should have 0 tags
	post2Tags := results[1].Get("tags").([]*entity.Entity)
	assert.Len(t, post2Tags, 0, "Post 2 should have 0 tags after offset (only had 1 active tag)")
}

func testM2MLimitWithSort(t *testing.T, client h.DBClient) {
	setupTestData(t, client)
	postModel := u.Must(client.C.Model("post"))

	// Limit with sort - get top 2 tags by priority (descending)
	relOpts := db.RelationOptions{
		"tags": {
			Sort:  "-priority",
			Limit: 2,
		},
	}

	results := u.Must(postModel.Query().
		Select("title", "tags.name", "tags.priority").
		WithRelationOptions(relOpts).
		Order("id").
		Get(h.Ctx()))

	require.Len(t, results, 2)

	// Post 1 has 5 tags with priorities 1, 2, 3, 4, 5
	// Sorted descending, top 2: Tag E (5), Tag D (4)
	post1Tags := results[0].Get("tags").([]*entity.Entity)
	require.Len(t, post1Tags, 2, "Post 1 should have 2 tags")
	assert.Equal(t, "Tag E", post1Tags[0].Get("name"), "First tag should be Tag E")
	assert.Equal(t, "Tag D", post1Tags[1].Get("name"), "Second tag should be Tag D")

	p1, _ := u.AnyToInt[int](post1Tags[0].Get("priority"))
	p2, _ := u.AnyToInt[int](post1Tags[1].Get("priority"))
	assert.Equal(t, 5, p1, "Tag E should have priority 5")
	assert.Equal(t, 4, p2, "Tag D should have priority 4")

	// Post 2 has tags: Tag A (priority 1), Tag C (priority 3)
	// Sorted descending, top 2: Tag C (3), Tag A (1)
	post2Tags := results[1].Get("tags").([]*entity.Entity)
	require.Len(t, post2Tags, 2, "Post 2 should have 2 tags")
	assert.Equal(t, "Tag C", post2Tags[0].Get("name"), "First tag should be Tag C")
	assert.Equal(t, "Tag A", post2Tags[1].Get("name"), "Second tag should be Tag A")
}

func testM2MOffsetBeyondCount(t *testing.T, client h.DBClient) {
	setupTestData(t, client)
	postModel := u.Must(client.C.Model("post"))

	// Offset beyond the number of related records
	relOpts := db.RelationOptions{
		"tags": {Offset: 100},
	}

	results := u.Must(postModel.Query().
		Select("title", "tags.name").
		WithRelationOptions(relOpts).
		Order("id").
		Get(h.Ctx()))

	require.Len(t, results, 2)

	// Both posts should have no tags when offset is beyond count
	post1Tags := results[0].Get("tags").([]*entity.Entity)
	assert.Len(t, post1Tags, 0, "Post 1 should have 0 tags with offset beyond count")

	post2Tags := results[1].Get("tags").([]*entity.Entity)
	assert.Len(t, post2Tags, 0, "Post 2 should have 0 tags with offset beyond count")
}

func testM2MFilterNoMatch(t *testing.T, client h.DBClient) {
	setupTestData(t, client)
	postModel := u.Must(client.C.Model("post"))

	// Filter that matches no tags
	relOpts := db.RelationOptions{
		"tags": {Filter: map[string]any{"status": "nonexistent"}},
	}

	results := u.Must(postModel.Query().
		Select("title", "tags.name").
		WithRelationOptions(relOpts).
		Order("id").
		Get(h.Ctx()))

	require.Len(t, results, 2)

	// Both posts should have 0 tags with nonexistent filter value
	post1Tags := results[0].Get("tags").([]*entity.Entity)
	assert.Len(t, post1Tags, 0, "Post 1 should have 0 tags with no matching filter")

	post2Tags := results[1].Get("tags").([]*entity.Entity)
	assert.Len(t, post2Tags, 0, "Post 2 should have 0 tags with no matching filter")
}

func testM2MMultipleRelations(t *testing.T, client h.DBClient) {
	setupTestData(t, client)
	postModel := u.Must(client.C.Model("post"))

	// Apply different options to different relations
	relOpts := db.RelationOptions{
		"tags":       {Limit: 2, Sort: "-priority"},
		"categories": {Limit: 1, Sort: "priority"},
	}

	results := u.Must(postModel.Query().
		Select("title", "tags.name", "tags.priority", "categories.name", "categories.priority").
		WithRelationOptions(relOpts).
		Order("id").
		Get(h.Ctx()))

	require.Len(t, results, 2)

	// Post 1: tags sorted desc by priority, limit 2 = Tag E (5), Tag D (4)
	post1Tags := results[0].Get("tags").([]*entity.Entity)
	require.Len(t, post1Tags, 2, "Post 1 should have 2 tags")
	assert.Equal(t, "Tag E", post1Tags[0].Get("name"))
	assert.Equal(t, "Tag D", post1Tags[1].Get("name"))

	// Post 1: categories sorted asc by priority, limit 1 = Cat 1 (priority 10)
	post1Categories := results[0].Get("categories").([]*entity.Entity)
	require.Len(t, post1Categories, 1, "Post 1 should have 1 category")
	assert.Equal(t, "Cat 1", post1Categories[0].Get("name"))

	// Post 2: tags sorted desc by priority, limit 2 = Tag C (3), Tag A (1)
	post2Tags := results[1].Get("tags").([]*entity.Entity)
	require.Len(t, post2Tags, 2, "Post 2 should have 2 tags")
	assert.Equal(t, "Tag C", post2Tags[0].Get("name"))
	assert.Equal(t, "Tag A", post2Tags[1].Get("name"))

	// Post 2: categories sorted asc by priority, limit 1 = Cat 1 (priority 10)
	post2Categories := results[1].Get("categories").([]*entity.Entity)
	require.Len(t, post2Categories, 1, "Post 2 should have 1 category")
	assert.Equal(t, "Cat 1", post2Categories[0].Get("name"))
}

func testO2MLimit(t *testing.T, client h.DBClient) {
	setupTestData(t, client)
	authorModel := u.Must(client.C.Model("author"))

	relOpts := db.RelationOptions{
		"posts": {Limit: 1},
	}

	results := u.Must(authorModel.Query().
		Select("name", "posts.title", "posts.priority").
		WithRelationOptions(relOpts).
		Order("id").
		Get(h.Ctx()))

	require.Len(t, results, 2)

	// Author 1 has 1 post (Post 1), limit 1 should return it
	author1 := results[0]
	assert.Equal(t, "Author 1", author1.Get("name"))
	author1Posts := author1.Get("posts").([]*entity.Entity)
	require.Len(t, author1Posts, 1, "Author 1 should have exactly 1 post with limit")
	assert.Equal(t, "Post 1", author1Posts[0].Get("title"), "Post should be 'Post 1'")
	p1, _ := u.AnyToInt[int](author1Posts[0].Get("priority"))
	assert.Equal(t, 1, p1, "Post 1 should have priority 1")

	// Author 2 has 1 post (Post 2), limit 1 should return it
	author2 := results[1]
	assert.Equal(t, "Author 2", author2.Get("name"))
	author2Posts := author2.Get("posts").([]*entity.Entity)
	require.Len(t, author2Posts, 1, "Author 2 should have exactly 1 post with limit")
	assert.Equal(t, "Post 2", author2Posts[0].Get("title"), "Post should be 'Post 2'")
	p2, _ := u.AnyToInt[int](author2Posts[0].Get("priority"))
	assert.Equal(t, 2, p2, "Post 2 should have priority 2")
}

func testO2MOffset(t *testing.T, client h.DBClient) {
	setupTestData(t, client)
	authorModel := u.Must(client.C.Model("author"))

	// Each author has only 1 post, so offset 1 should return 0 posts
	relOpts := db.RelationOptions{
		"posts": {Offset: 1},
	}

	results := u.Must(authorModel.Query().
		Select("name", "posts.title").
		WithRelationOptions(relOpts).
		Order("id").
		Get(h.Ctx()))

	require.Len(t, results, 2)

	// Author 1 has 1 post, offset 1 means no posts
	author1Posts := results[0].Get("posts").([]*entity.Entity)
	assert.Len(t, author1Posts, 0, "Author 1 should have 0 posts with offset 1")

	// Author 2 has 1 post, offset 1 means no posts
	author2Posts := results[1].Get("posts").([]*entity.Entity)
	assert.Len(t, author2Posts, 0, "Author 2 should have 0 posts with offset 1")
}

func testO2MSort(t *testing.T, client h.DBClient) {
	setupTestData(t, client)
	authorModel := u.Must(client.C.Model("author"))

	relOpts := db.RelationOptions{
		"posts": {Sort: "-priority"},
	}

	results := u.Must(authorModel.Query().
		Select("name", "posts.title", "posts.priority").
		WithRelationOptions(relOpts).
		Order("id").
		Get(h.Ctx()))

	require.Len(t, results, 2)

	// Each author has only 1 post, so sort doesn't affect the result
	// But we verify the post is still correct
	author1 := results[0]
	assert.Equal(t, "Author 1", author1.Get("name"))
	author1Posts := author1.Get("posts").([]*entity.Entity)
	require.Len(t, author1Posts, 1)
	assert.Equal(t, "Post 1", author1Posts[0].Get("title"))

	author2 := results[1]
	assert.Equal(t, "Author 2", author2.Get("name"))
	author2Posts := author2.Get("posts").([]*entity.Entity)
	require.Len(t, author2Posts, 1)
	assert.Equal(t, "Post 2", author2Posts[0].Get("title"))
}

func testO2MFilter(t *testing.T, client h.DBClient) {
	setupTestData(t, client)
	authorModel := u.Must(client.C.Model("author"))

	// Filter by status = published (only Post 1 has this status)
	relOpts := db.RelationOptions{
		"posts": {Filter: map[string]any{"status": "published"}},
	}

	results := u.Must(authorModel.Query().
		Select("name", "posts.title", "posts.status").
		WithRelationOptions(relOpts).
		Order("id").
		Get(h.Ctx()))

	require.Len(t, results, 2)

	// Author 1 has Post 1 (published), should return it
	author1 := results[0]
	assert.Equal(t, "Author 1", author1.Get("name"))
	author1Posts := author1.Get("posts").([]*entity.Entity)
	require.Len(t, author1Posts, 1, "Author 1 should have 1 published post")
	assert.Equal(t, "Post 1", author1Posts[0].Get("title"))
	assert.Equal(t, "published", author1Posts[0].Get("status"))

	// Author 2 has Post 2 (draft), should return 0 posts
	author2 := results[1]
	assert.Equal(t, "Author 2", author2.Get("name"))
	author2Posts := author2.Get("posts").([]*entity.Entity)
	assert.Len(t, author2Posts, 0, "Author 2 should have 0 published posts")
}

func testBackwardCompatibility(t *testing.T, client h.DBClient) {
	setupTestData(t, client)
	postModel := u.Must(client.C.Model("post"))

	// Query without relation options - should return all related records
	results := u.Must(postModel.Query().
		Select("title", "tags.name", "tags.priority").
		Order("id").
		Get(h.Ctx()))

	require.Len(t, results, 2)

	// Post 1 should have all 5 tags
	post1Tags := results[0].Get("tags").([]*entity.Entity)
	require.Len(t, post1Tags, 5, "Post 1 should have all 5 tags")

	expectedTagNames := []string{"Tag A", "Tag B", "Tag C", "Tag D", "Tag E"}
	for i, tag := range post1Tags {
		assert.Equal(t, expectedTagNames[i], tag.Get("name"), "Tag at position %d should be %s", i, expectedTagNames[i])
	}

	// Post 2 should have 2 tags
	post2Tags := results[1].Get("tags").([]*entity.Entity)
	require.Len(t, post2Tags, 2, "Post 2 should have 2 tags")
	assert.Equal(t, "Tag A", post2Tags[0].Get("name"))
	assert.Equal(t, "Tag C", post2Tags[1].Get("name"))
}

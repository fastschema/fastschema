package pk_test

import (
	"fmt"
	"testing"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/entity"
	u "github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
	h "github.com/fastschema/fastschema/tests/integration/helpers"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	schemaDir    = "../../../tests/integration/pk/data/schemas"
	migrationDir = "../../../tests/integration/pk/data/migrations"
	sqliteDSN    = "../../../tests/integration/pk/data/pk_test.db"
)

var pkTestCases = []struct {
	name string
	fn   func(t *testing.T, client h.DBClient)
}{
	{"SchemaMetadata", testSchemaMetadata},
	{"CustomPKCrud", testCustomPKCrud},
	{"InvalidPKInputs", testInvalidPKInputs},
	{"AutoIncrementCompatibility", testAutoIncrementCompatibility},
	{"MixedRelationsAndQueries", testMixedRelationsAndQueries},
	{"SystemSchemasOnly", testSystemSchemasOnly},
	{"SystemM2MRelations", testSystemM2MRelations},
	{"MixedSystemAndJSONSchemas", testMixedSystemAndJSONSchemas},
	{"MixedSystemM2MUuidUint", testMixedSystemM2MUuidUint},
	{"M2MStringUUID", testM2MStringUUID},
	{"M2MUuidUint", testM2MUuidUint},
}

func newPKSchemaBuilder(t *testing.T) *schema.Builder {
	t.Helper()
	return u.Must(schema.NewBuilderFromDir(
		schemaDir,
		systemSeries{},
		systemEpisode{},
		systemTopic{},
		systemProfile{},
	))
}

func TestPKMySQL(t *testing.T) {
	runPKTests(t, u.Map(h.MysqlConfigs, func(cfg h.DBConfig) h.DBClient {
		sb := newPKSchemaBuilder(t)
		return h.NewMySQLClient(t, cfg, sb, migrationDir)
	}))
}

func TestPKPostgres(t *testing.T) {
	runPKTests(t, u.Map(h.PostgresConfigs, func(cfg h.DBConfig) h.DBClient {
		sb := newPKSchemaBuilder(t)
		return h.NewPostgresClient(t, cfg, sb, migrationDir)
	}))
}

func TestPKSQLite(t *testing.T) {
	sb := newPKSchemaBuilder(t)
	client := h.NewSQLiteClient(t, "sqlite", sqliteDSN, migrationDir, sb)
	runPKTests(t, []h.DBClient{client})
}

func runPKTests(t *testing.T, clients []h.DBClient) {
	for _, client := range clients {
		clientCopy := client
		t.Run(clientCopy.Name, func(t *testing.T) {
			for _, tc := range pkTestCases {
				testCase := tc
				t.Run(testCase.name, func(t *testing.T) {
					testCase.fn(t, clientCopy)
				})
			}
		})
	}
}

func testSchemaMetadata(t *testing.T, client h.DBClient) {
	builder := client.C.SchemaBuilder()
	require.NotNil(t, builder)

	category := u.Must(builder.Schema("category"))
	categoryID := category.IDField()
	require.NotNil(t, categoryID)
	assert.Equal(t, schema.TypeString, categoryID.Type)
	assert.False(t, categoryID.DB.Increment)
	assert.True(t, categoryID.Filterable)
	assert.True(t, categoryID.Sortable)

	tag := u.Must(builder.Schema("tag"))
	tagID := tag.IDField()
	require.NotNil(t, tagID)
	assert.Equal(t, schema.TypeUint64, tagID.Type)
	assert.False(t, tagID.DB.Increment)

	user := u.Must(builder.Schema("user"))
	userID := user.IDField()
	require.NotNil(t, userID)
	assert.Equal(t, schema.TypeUUID, userID.Type)

	post := u.Must(builder.Schema("post"))
	postID := post.IDField()
	require.NotNil(t, postID)
	assert.Equal(t, schema.TypeUint64, postID.Type)
	assert.True(t, postID.DB.Increment)
}

func testCustomPKCrud(t *testing.T, client h.DBClient) {
	f := seedBlogGraph(t, client)
	categoryModel := u.Must(client.C.Model("category"))
	tagModel := u.Must(client.C.Model("tag"))

	// Update category description using string PK
	affected := u.Must(categoryModel.
		Mutation().
		Where(db.EQ("id", f.categorySlug)).
		Update(h.Ctx(), entity.New().Set("description", "updated")))
	require.Equal(t, 1, affected)

	// Refetch category and verify update using string PK
	category := u.Must(categoryModel.
		Query(db.EQ("id", f.categorySlug)).
		First(h.Ctx()))
	assert.Equal(t, "updated", category.Get("description"))

	// Attempt to create duplicate category with same PK should fail
	_, err := categoryModel.CreateFromJSON(
		h.Ctx(),
		fmt.Sprintf(`{"id":"%s","title":"Duplicate"}`, f.categorySlug),
	)
	require.Error(t, err)

	// Create new tag with uint64 PK
	newTagID := uint64(9100)
	u.Must(tagModel.CreateFromJSON(
		h.Ctx(),
		fmt.Sprintf(`{"id":%d,"name":"apis"}`, newTagID),
	))

	// Relate new tag to existing post using uint64 PK
	u.Must(tagModel.
		Mutation().
		Where(db.EQ("id", newTagID)).
		Update(h.Ctx(), entity.New().Set("posts", []*entity.Entity{
			entity.New(f.postID),
		})))

	// Update tag name using uint64 PK
	affected = u.Must(tagModel.
		Mutation().
		Where(db.EQ("id", newTagID)).
		Update(h.Ctx(), entity.New().Set("name", "rest")))
	require.Equal(t, 1, affected)

	// Refetch tag and verify update
	tag := u.Must(tagModel.
		Query(db.EQ("id", newTagID)).
		First(h.Ctx()))
	assert.Equal(t, "rest", tag.Get("name"))

	// Delete tag using uint64 PK
	affected = u.Must(tagModel.
		Mutation().
		Where(db.EQ("id", f.tagID)).
		Delete(h.Ctx()))
	require.Equal(t, 1, affected)
	remaining := u.Must(tagModel.Query(db.EQ("id", f.tagID)).Get(h.Ctx()))
	assert.Len(t, remaining, 0)
}

func testInvalidPKInputs(t *testing.T, client h.DBClient) {
	h.ClearDBData(client.C, pkTables...)
	categoryModel := u.Must(client.C.Model("category"))
	tagModel := u.Must(client.C.Model("tag"))
	postModel := u.Must(client.C.Model("post"))

	// Missing string PK
	_, err := categoryModel.CreateFromJSON(
		h.Ctx(),
		`{"title":"Missing"}`,
	)
	require.Error(t, err)

	// Valid uint64 PK
	_, err = tagModel.CreateFromJSON(
		h.Ctx(),
		`{"id":1,"name":"ok"}`,
	)
	require.NoError(t, err)

	// Duplicate uint64 PK
	_, err = tagModel.CreateFromJSON(
		h.Ctx(),
		`{"id":1,"name":"duplicate"}`,
	)
	require.Error(t, err)

	// Invalid string PK type
	_, err = tagModel.CreateFromJSON(
		h.Ctx(),
		`{"id":"abc","name":"invalid"}`,
	)
	require.Error(t, err)

	// Create error for invalid foreign key references
	_, err = postModel.CreateFromJSON(
		h.Ctx(),
		`{"title":"orphan","category":{"id":"ghost"},"author":{"id":"ghost"}}`,
	)
	require.Error(t, err)
}

func testAutoIncrementCompatibility(t *testing.T, client h.DBClient) {
	f := seedBlogGraph(t, client)
	commentModel := u.Must(client.C.Model("comment"))

	comment := u.Must(commentModel.
		Query(db.EQ("post_id", f.postID)).
		First(h.Ctx()))
	commentID := coerceUint(t, comment.ID())
	assert.Greater(t, commentID, uint64(0))

	updated := u.Must(commentModel.
		Mutation().
		Where(db.EQ("id", commentID)).
		Update(h.Ctx(), entity.New().Set("content", "edited")))
	require.Equal(t, 1, updated)

	refetched := u.Must(commentModel.
		Query(db.EQ("id", commentID)).
		First(h.Ctx()))
	assert.Equal(t, "edited", refetched.Get("content"))
}

func testMixedRelationsAndQueries(t *testing.T, client h.DBClient) {
	f := seedBlogGraph(t, client)
	postModel := u.Must(client.C.Model("post"))
	commentModel := u.Must(client.C.Model("comment"))
	tagModel := u.Must(client.C.Model("tag"))

	// Create second post
	secondPostID := normalizeUint(t, u.Must(postModel.CreateFromJSON(
		h.Ctx(),
		fmt.Sprintf(
			`{"title":"Second","body":"more","category":{"id":"%s"},"author":{"id":"%s"}}`,
			f.categorySlug, f.userID,
		),
	)))

	// Create comment for second post
	u.Must(commentModel.CreateFromJSON(
		h.Ctx(),
		fmt.Sprintf(
			`{"content":"another","post":{"id":%d},"author":{"id":"%s"}}`,
			secondPostID, f.userID,
		),
	))

	// Create and relate tag to second post
	secondTagID := uint64(9200)
	u.Must(tagModel.CreateFromJSON(
		h.Ctx(),
		fmt.Sprintf(`{"id":%d,"name":"cli"}`, secondTagID),
	))
	u.Must(tagModel.
		Mutation().
		Where(db.EQ("id", secondTagID)).
		Update(h.Ctx(), entity.New().Set("posts", []*entity.Entity{
			entity.New(secondPostID),
		})))

	// Query by string PK relation
	postsByCategory := u.Must(postModel.
		Query(db.EQ("category_id", f.categorySlug)).
		Get(h.Ctx()))
	require.Len(t, postsByCategory, 2)

	// Query by uuid PK relation
	postsByAuthor := u.Must(postModel.
		Query(db.EQ("author_id", f.userID)).
		Get(h.Ctx()))
	require.Len(t, postsByAuthor, 2)

	// Query by uint64 PK relation
	comments := u.Must(commentModel.
		Query(db.EQ("post_id", f.postID)).
		Get(h.Ctx()))
	require.Len(t, comments, 1)
	assert.Equal(t, "first!", comments[0].Get("content"))

	// Query by uint64 PK relation
	commentsSecond := u.Must(commentModel.
		Query(db.EQ("post_id", secondPostID)).
		Get(h.Ctx()))
	require.Len(t, commentsSecond, 1)
	assert.Equal(t, "another", commentsSecond[0].Get("content"))
}

func testSystemSchemasOnly(t *testing.T, client h.DBClient) {
	h.ClearDBData(client.C, pkTables...)

	seriesModel := u.Must(client.C.Model("system_series"))
	episodeModel := u.Must(client.C.Model("system_episode"))

	seriesID := uuid.NewString()

	// Create series
	u.Must(seriesModel.CreateFromJSON(
		h.Ctx(),
		fmt.Sprintf(`{"id":"%s","title":"Docs","synopsis":"system"}`, seriesID),
	))

	// Create episode "Pilot"
	pilotID := normalizeUint(t, u.Must(episodeModel.CreateFromJSON(
		h.Ctx(),
		fmt.Sprintf(`{"title":"Pilot","duration":45,"series":{"id":"%s"}}`, seriesID),
	)))

	// Create episode: "Finale"
	finaleID := normalizeUint(t, u.Must(episodeModel.CreateFromJSON(
		h.Ctx(),
		fmt.Sprintf(`{"title":"Finale","duration":60,"series":{"id":"%s"}}`, seriesID),
	)))

	// Get all episodes from series
	episodes := u.Must(episodeModel.
		Query(db.EQ("series_id", seriesID)).
		Get(h.Ctx()))
	require.Len(t, episodes, 2)
	ids := u.Map(episodes, func(e *entity.Entity) uint64 {
		return coerceUint(t, e.Get("id"))
	})
	assert.ElementsMatch(t, []uint64{pilotID, finaleID}, ids)

	// Update record using uuid PK
	updated := u.Must(episodeModel.
		Mutation().
		Where(db.EQ("id", pilotID)).
		Update(h.Ctx(), entity.New().Set("duration", 55)))
	require.Equal(t, 1, updated)

	// Query record using uint64 PK
	pilot := u.Must(episodeModel.
		Query(db.EQ("id", pilotID)).
		First(h.Ctx()))
	assert.EqualValues(t, 55, coerceInt(t, pilot.Get("duration")))

	// Delete record using uint64 PK
	deleted := u.Must(episodeModel.
		Mutation().
		Where(db.EQ("id", finaleID)).
		Delete(h.Ctx()))
	require.Equal(t, 1, deleted)

	// Query remaining episodes using uuid FK
	remaining1 := u.Must(episodeModel.
		Query(db.EQ("series_id", seriesID)).
		Get(h.Ctx()))
	require.Len(t, remaining1, 1)

	// Query remaining episodes using relation uuid PK
	remaining2 := u.Must(episodeModel.
		Query(db.EQ("series.id", seriesID)).
		Get(h.Ctx()))
	require.Len(t, remaining2, 1)

	// Delete episode using uint64 PK
	deleted = u.Must(episodeModel.
		Mutation().
		Where(db.EQ("id", pilotID)).
		Delete(h.Ctx()))
	require.Equal(t, 1, deleted)

	// Delete series using uuid PK
	cleared := u.Must(seriesModel.
		Mutation().
		Where(db.EQ("id", seriesID)).
		Delete(h.Ctx()))
	require.Equal(t, 1, cleared)
}

func testSystemM2MRelations(t *testing.T, client h.DBClient) {
	h.ClearDBData(client.C, pkTables...)

	seriesModel := u.Must(client.C.Model("system_series"))
	topicModel := u.Must(client.C.Model("system_topic"))
	joinModel := u.Must(client.C.Model("series_topics"))

	// Create test topics
	topicIDs := []string{
		fmt.Sprintf("topic-%s", uuid.NewString()),
		fmt.Sprintf("topic-%s", uuid.NewString()),
		fmt.Sprintf("topic-%s", uuid.NewString()),
	}

	for idx, id := range topicIDs {
		u.Must(topicModel.CreateFromJSON(
			h.Ctx(),
			fmt.Sprintf(`{"id":"%s","label":"Topic %d"}`, id, idx),
		))
	}

	seriesID := uuid.NewString()

	// Create series
	u.Must(seriesModel.CreateFromJSON(
		h.Ctx(),
		fmt.Sprintf(`{"id":"%s","title":"System M2M","synopsis":"links"}`, seriesID),
	))

	// Add topics to series
	u.Must(seriesModel.
		Mutation().
		Where(db.EQ("id", seriesID)).
		Update(h.Ctx(), entity.New().Set("topics", []*entity.Entity{
			entity.New(topicIDs[0]),
			entity.New(topicIDs[1]),
		})))

	// Select relation and filter by relation
	serieses := u.Must(seriesModel.
		Query(db.EQ("topics.label", "Topic 1")).
		Select("topics").
		Get(h.Ctx()))

	// "Topic 1" has 1 series that belong to "Topic 1" and "Topic 2"
	assert.Equal(t, "System M2M", serieses[0].Get("title"))
	topics, ok := serieses[0].Get("topics").([]*entity.Entity)
	assert.True(t, ok)
	topicNames := u.Map(topics, func(e *entity.Entity) string {
		return coerceString(t, e.Get("label"))
	})
	assert.ElementsMatch(t, []string{"Topic 0", "Topic 1"}, topicNames)

	// Verify junction table data
	rows := u.Must(joinModel.
		Query(db.EQ("series", seriesID)).
		Get(h.Ctx()))
	require.Len(t, rows, 2)
	attached := u.Map(rows, func(e *entity.Entity) string {
		return coerceString(t, e.Get("topics"))
	})
	assert.ElementsMatch(t, topicIDs[:2], attached)

	// Replace topics in series
	u.Must(seriesModel.
		Mutation().
		Where(db.EQ("id", seriesID)).
		Update(h.Ctx(), entity.New().Set("topics", []*entity.Entity{
			entity.New(topicIDs[2]),
		})))

	// Verify junction table data after replacement
	rows = u.Must(joinModel.
		Query(db.EQ("series", seriesID)).
		Get(h.Ctx()))
	require.Len(t, rows, 1)
	assert.Equal(t, topicIDs[2], coerceString(t, rows[0].Get("topics")))

	// Verify inverse relation from topics to series
	inverse := u.Must(joinModel.
		Query(db.EQ("topics", topicIDs[2])).
		Get(h.Ctx()))
	require.Len(t, inverse, 1)
	assert.Equal(t, seriesID, coerceString(t, inverse[0].Get("series")))
}

func testMixedSystemAndJSONSchemas(t *testing.T, client h.DBClient) {
	f := seedBlogGraph(t, client)
	profileModel := u.Must(client.C.Model("system_profile"))
	userModel := u.Must(client.C.Model("user"))

	primaryProfileID := fmt.Sprintf("profile-%s", uuid.NewString())

	// Create primary profile for user with string PK
	u.Must(profileModel.CreateFromJSON(
		h.Ctx(),
		fmt.Sprintf(`
		{"id":"%s","display_name":"Author","bio":"writer","owner":{"id":"%s"}}`,
			primaryProfileID, f.userID,
		),
	))

	// Query profile by user relation uuid PK
	primaryOwnerProfiles := u.Must(profileModel.
		Query(db.EQ("owner_id", f.userID)).
		Get(h.Ctx()))
	require.Len(t, primaryOwnerProfiles, 1)
	assert.Equal(t, primaryProfileID, coerceString(t, primaryOwnerProfiles[0].Get("id")))

	// Create orphan profile with string PK
	orphanProfileID := fmt.Sprintf("profile-%s", uuid.NewString())
	u.Must(profileModel.CreateFromJSON(
		h.Ctx(),
		fmt.Sprintf(
			`{"id":"%s","display_name":"Editor","bio":"pending"}`,
			orphanProfileID,
		),
	))

	// Create second user with uuid PK
	secondUserID := uuid.NewString()
	u.Must(userModel.CreateFromJSON(
		h.Ctx(),
		fmt.Sprintf(`{"id":"%s","name":"Barry","email":"barry_%s@example.com"}`,
			secondUserID, secondUserID[:8],
		),
	))

	// Assign orphan profile to second user using string PK
	updated := u.Must(userModel.
		Mutation().
		Where(db.EQ("id", secondUserID)).
		Update(h.Ctx(), entity.New().Set("profile", entity.New(orphanProfileID))))
	require.Equal(t, 1, updated)

	// Verify reassignment
	reassigned := u.Must(profileModel.
		Query(db.EQ("owner_id", secondUserID)).
		First(h.Ctx()))
	assert.Equal(t, orphanProfileID, coerceString(t, reassigned.Get("id")))

	// Verify primary profile still linked to first user
	primaryOwnerProfiles = u.Must(profileModel.
		Query(db.EQ("owner_id", f.userID)).
		Get(h.Ctx()))
	require.Len(t, primaryOwnerProfiles, 1)
	assert.Equal(t, primaryProfileID, coerceString(t, primaryOwnerProfiles[0].Get("id")))

	// Update orphan profile bio using string PK
	u.Must(profileModel.
		Mutation().
		Where(db.EQ("id", orphanProfileID)).
		Update(h.Ctx(), entity.New().Set("bio", "edited")))

	// Refetch orphan profile and verify update
	refetched := u.Must(profileModel.
		Query(db.EQ("id", orphanProfileID)).
		First(h.Ctx()))
	assert.Equal(t, "edited", refetched.Get("bio"))
}

func testMixedSystemM2MUuidUint(t *testing.T, client h.DBClient) {
	h.ClearDBData(client.C, pkTables...)

	seriesModel := u.Must(client.C.Model("system_series"))
	episodeModel := u.Must(client.C.Model("system_episode"))
	userModel := u.Must(client.C.Model("user"))
	joinModel := u.Must(client.C.Model("favorite_episodes_viewers"))

	seriesID := uuid.NewString()

	// Create series with uuid PK
	u.Must(seriesModel.CreateFromJSON(
		h.Ctx(),
		fmt.Sprintf(
			`{"id":"%s","title":"Mixed","synopsis":"bridge"}`,
			seriesID,
		),
	))

	// Create episodes with uint64 PKs
	firstEpisodeID := normalizeUint(t, u.Must(episodeModel.CreateFromJSON(
		h.Ctx(),
		fmt.Sprintf(
			`{"title":"Pilot","duration":30,"series":{"id":"%s"}}`,
			seriesID,
		),
	)))
	secondEpisodeID := normalizeUint(t, u.Must(episodeModel.CreateFromJSON(
		h.Ctx(),
		fmt.Sprintf(
			`{"title":"Encore","duration":35,"series":{"id":"%s"}}`,
			seriesID,
		),
	)))

	// Query episodes by series uuid FK
	episodes := u.Must(episodeModel.
		Query(db.EQ("series.id", seriesID)).
		Get(h.Ctx()))
	require.Len(t, episodes, 2)
	ids := u.Map(episodes, func(e *entity.Entity) uint64 {
		return coerceUint(t, e.Get("id"))
	})
	assert.ElementsMatch(t, []uint64{firstEpisodeID, secondEpisodeID}, ids)

	// Query series by episode uint64 FK
	serieses := u.Must(seriesModel.
		Query(db.EQ("episodes.id", firstEpisodeID)).
		Get(h.Ctx()))
	require.Len(t, serieses, 1)
	assert.Equal(t, seriesID, coerceString(t, serieses[0].Get("id")))

	// Create viewer users with uuid PKs
	viewerIDs := []string{uuid.NewString(), uuid.NewString()}
	for idx, viewer := range viewerIDs {
		u.Must(userModel.CreateFromJSON(
			h.Ctx(),
			fmt.Sprintf(
				`{"id":"%s","name":"Viewer %d","email":"viewer_%d@example.com"}`,
				viewer, idx, idx,
			),
		))
	}

	// Function to set viewers for an episode
	setViewers := func(episodeID uint64, ids ...string) {
		entities := u.Map(ids, func(id string) *entity.Entity {
			return entity.New(id)
		})
		u.Must(episodeModel.
			Mutation().
			Where(db.EQ("id", episodeID)).
			Update(h.Ctx(), entity.New().Set("viewers", entities)))
	}

	setViewers(firstEpisodeID, viewerIDs[0])

	// Verify junction table entries
	rows := u.Must(joinModel.
		Query(db.EQ("favorite_episodes", firstEpisodeID)).
		Get(h.Ctx()))
	require.Len(t, rows, 1)
	assert.Equal(t, viewerIDs[0], coerceString(t, rows[0].Get("viewers")))
	setViewers(firstEpisodeID, viewerIDs[0], viewerIDs[1])

	// Verify both viewers are linked
	rows = u.Must(joinModel.
		Query(db.EQ("favorite_episodes", firstEpisodeID)).
		Get(h.Ctx()))
	require.Len(t, rows, 2)
	viewers := u.Map(rows, func(e *entity.Entity) string {
		return coerceString(t, e.Get("viewers"))
	})
	assert.ElementsMatch(t, viewerIDs, viewers)

	setViewers(firstEpisodeID, viewerIDs[1])
	setViewers(secondEpisodeID, viewerIDs[0])

	// Verify updated junction table entries
	firstRows := u.Must(joinModel.
		Query(db.EQ("favorite_episodes", firstEpisodeID)).
		Get(h.Ctx()))
	require.Len(t, firstRows, 1)
	assert.Equal(t, viewerIDs[1], coerceString(t, firstRows[0].Get("viewers")))

	// Verify second episode viewers
	secondRows := u.Must(joinModel.
		Query(db.EQ("favorite_episodes", secondEpisodeID)).
		Get(h.Ctx()))
	require.Len(t, secondRows, 1)
	assert.Equal(t, viewerIDs[0], coerceString(t, secondRows[0].Get("viewers")))

	// Verify inverse relations
	inverse := u.Must(joinModel.
		Query(db.EQ("viewers", viewerIDs[0])).
		Get(h.Ctx()))
	require.Len(t, inverse, 1)
	assert.Equal(t, secondEpisodeID, coerceUint(t, inverse[0].Get("favorite_episodes")))
}

func testM2MStringUUID(t *testing.T, client h.DBClient) {
	f := seedBlogGraph(t, client)
	categoryModel := u.Must(client.C.Model("category"))
	userModel := u.Must(client.C.Model("user"))
	joinModel := u.Must(client.C.Model("followed_categories_followers"))

	toString := func(value any) string { return coerceString(t, value) }
	secondUserID := uuid.NewString()

	// Create second user with uuid PK
	u.Must(userModel.CreateFromJSON(
		h.Ctx(),
		fmt.Sprintf(
			`{"id":"%s","name":"Bob","email":"bob@example.com"}`,
			secondUserID,
		),
	))

	// Create third user with uuid PK
	thirdUserID := uuid.NewString()
	u.Must(userModel.CreateFromJSON(
		h.Ctx(),
		fmt.Sprintf(
			`{"id":"%s","name":"Cara","email":"cara@example.com"}`,
			thirdUserID,
		),
	))

	// Add followers to category using uuid PKs
	_, err := categoryModel.Mutation().
		Where(db.EQ("id", f.categorySlug)).
		Update(h.Ctx(), entity.New().Set("followers", []*entity.Entity{
			entity.New(f.userID),
			entity.New(secondUserID),
		}))
	require.NoError(t, err)

	// Query using relation uuid PK
	categories := u.Must(categoryModel.
		Query(db.EQ("followers.id", f.userID)).
		Select("followers").
		Get(h.Ctx()))
	require.Len(t, categories, 1)
	assert.Equal(t, f.categorySlug, toString(categories[0].Get("id")))
	followers, ok := categories[0].Get("followers").([]*entity.Entity)
	assert.True(t, ok)
	followerIDs := u.Map(followers, func(e *entity.Entity) string {
		return toString(e.Get("id"))
	})
	assert.ElementsMatch(t, []string{f.userID, secondUserID}, followerIDs)

	// Query using filter on relation string PK
	followers2 := u.Must(userModel.
		Query(db.EQ("followed_categories.id", f.categorySlug)).
		Select("followed_categories").
		Get(h.Ctx()))
	require.Len(t, followers2, 2)
	categories, ok = followers2[0].Get("followed_categories").([]*entity.Entity)
	assert.True(t, ok)
	categoryIDs := u.Map(categories, func(e *entity.Entity) string {
		return toString(e.Get("id"))
	})
	assert.ElementsMatch(t, []string{f.categorySlug}, categoryIDs)

	// Verify junction table entries
	rows := u.Must(joinModel.
		Query(db.EQ("followed_categories", f.categorySlug)).
		Get(h.Ctx()))
	require.Len(t, rows, 2)
	followers3 := u.Map(rows, func(e *entity.Entity) string {
		return toString(e.Get("followers"))
	})
	assert.ElementsMatch(t, []string{f.userID, secondUserID}, followers3)

	// Replace followers with third user
	_, err = categoryModel.
		Mutation().
		Where(db.EQ("id", f.categorySlug)).
		Update(h.Ctx(), entity.New().Set("followers", []*entity.Entity{
			entity.New(thirdUserID),
		}))
	require.NoError(t, err)

	// Verify junction table after replacement
	rows = u.Must(joinModel.
		Query(db.EQ("followed_categories", f.categorySlug)).
		Get(h.Ctx()))
	require.Len(t, rows, 1)
	assert.Equal(t, thirdUserID, toString(rows[0].Get("followers")))

	// Verify inverse relation from user to categories
	inverse := u.Must(joinModel.
		Query(db.EQ("followers", thirdUserID)).
		Get(h.Ctx()))
	require.Len(t, inverse, 1)
	assert.Equal(t, f.categorySlug, toString(inverse[0].Get("followed_categories")))
}

func testM2MUuidUint(t *testing.T, client h.DBClient) {
	f := seedBlogGraph(t, client)
	postModel := u.Must(client.C.Model("post"))
	userModel := u.Must(client.C.Model("user"))
	joinModel := u.Must(client.C.Model("liked_posts_likes"))

	toString := func(value any) string { return coerceString(t, value) }
	secondUserID := uuid.NewString()

	// Create second user with uuid PK
	u.Must(userModel.CreateFromJSON(
		h.Ctx(),
		fmt.Sprintf(
			`{"id":"%s","name":"Dana","email":"dana@example.com"}`,
			secondUserID,
		),
	))

	// Create third user with uuid PK
	thirdUserID := uuid.NewString()
	u.Must(userModel.CreateFromJSON(
		h.Ctx(),
		fmt.Sprintf(`{"id":"%s","name":"Eli","email":"eli@example.com"}`, thirdUserID),
	))

	// Add likes to post using user uuid PKs
	_, err := postModel.
		Mutation().
		Where(db.EQ("id", f.postID)).
		Update(h.Ctx(), entity.New().Set("likes", []*entity.Entity{
			entity.New(f.userID),
			entity.New(secondUserID),
		}))
	require.NoError(t, err)

	// Query using relation uuid PK
	posts := u.Must(postModel.
		Query(db.EQ("likes.id", f.userID)).
		Select("likes").
		Get(h.Ctx()))
	require.Len(t, posts, 1)
	assert.Equal(t, f.postID, coerceUint(t, posts[0].Get("id")))
	likers, ok := posts[0].Get("likes").([]*entity.Entity)
	assert.True(t, ok)
	likerIDs := u.Map(likers, func(e *entity.Entity) string {
		return toString(e.Get("id"))
	})
	assert.ElementsMatch(t, []string{f.userID, secondUserID}, likerIDs)

	// Query using filter on relation uint64 PK
	likers2 := u.Must(userModel.
		Query(db.EQ("liked_posts.id", f.postID)).
		Select("liked_posts").
		Get(h.Ctx()))
	require.Len(t, likers2, 2)
	posts2, ok := likers2[0].Get("liked_posts").([]*entity.Entity)
	assert.True(t, ok)
	postIDs := u.Map(posts2, func(e *entity.Entity) uint64 {
		return coerceUint(t, e.Get("id"))
	})
	assert.ElementsMatch(t, []uint64{f.postID}, postIDs)

	// Verify junction table entries
	rows := u.Must(joinModel.
		Query(db.EQ("liked_posts", f.postID)).
		Get(h.Ctx()))
	require.Len(t, rows, 2)
	likers3 := u.Map(rows, func(e *entity.Entity) string {
		return toString(e.Get("likes"))
	})
	assert.ElementsMatch(t, []string{f.userID, secondUserID}, likers3)

	// Replace likes with third user
	_, err = postModel.
		Mutation().
		Where(db.EQ("id", f.postID)).
		Update(h.Ctx(), entity.New().Set("likes", []*entity.Entity{
			entity.New(thirdUserID),
		}))
	require.NoError(t, err)

	// Verify junction table after replacement
	rows = u.Must(joinModel.
		Query(db.EQ("liked_posts", f.postID)).
		Get(h.Ctx()))
	require.Len(t, rows, 1)
	assert.Equal(t, thirdUserID, toString(rows[0].Get("likes")))

	// Verify inverse relation from user to liked posts
	inverse := u.Must(joinModel.
		Query(db.EQ("likes", thirdUserID)).
		Get(h.Ctx()))
	require.Len(t, inverse, 1)
	assert.Equal(t, f.postID, coerceUint(t, inverse[0].Get("liked_posts")))
}

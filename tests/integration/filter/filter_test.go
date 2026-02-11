package filter

import (
	"fmt"
	"sort"
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
	schemaDir    = "../../../tests/integration/filter/data/schemas"
	migrationDir = "../../../tests/integration/filter/data/migrations"
	sqliteDSN    = "../../../tests/integration/filter/data/filter_test.db"
)

func TestMySQL(t *testing.T) {
	runFilterTests(t, u.Map(h.MysqlConfigs, func(cfg h.DBConfig) h.DBClient {
		sb := u.Must(schema.NewBuilderFromDir(schemaDir))
		return h.NewMySQLClient(t, cfg, sb, migrationDir)
	}))
}

func TestPostgres(t *testing.T) {
	runFilterTests(t, u.Map(h.PostgresConfigs, func(cfg h.DBConfig) h.DBClient {
		sb := u.Must(schema.NewBuilderFromDir(schemaDir))
		return h.NewPostgresClient(t, cfg, sb, migrationDir)
	}))
}

func TestSQLite(t *testing.T) {
	sb := u.Must(schema.NewBuilderFromDir(schemaDir))
	client := h.NewSQLiteClient(t, "sqlite", sqliteDSN, migrationDir, sb)
	runFilterTests(t, []h.DBClient{client})
}

func runFilterTests(t *testing.T, clients []h.DBClient) {
	for _, client := range clients {
		client := client
		t.Run(client.Name, func(t *testing.T) {
			t.Run("M2MDefaultNEQ", func(t *testing.T) {
				testM2MNEQDefault(t, client)
			})

			t.Run("M2MCustomNEQ", func(t *testing.T) {
				testM2MNEQCustom(t, client)
			})

			t.Run("ImplicitPKFilter", func(t *testing.T) {
				testImplicitPKFilter(t, client)
			})

			t.Run("ImplicitPKFilterWithOperators", func(t *testing.T) {
				testImplicitPKFilterWithOperators(t, client)
			})
		})
	}
}

func testM2MNEQDefault(t *testing.T, client h.DBClient) {
	h.ClearDBData(client.C, "post_legacy_tags", "posts_tags", "posts", "tags")

	postModel := u.Must(client.C.Model("post"))
	tagModel := u.Must(client.C.Model("tag"))

	targetTagID := u.Must(tagModel.CreateFromJSON(h.Ctx(), `{"label": "Tag 1"}`))
	otherTagID := u.Must(tagModel.CreateFromJSON(h.Ctx(), `{"label": "Tag 2"}`))

	u.Must(postModel.CreateFromJSON(h.Ctx(), `{"title": "Post 1"}`))
	u.Must(postModel.CreateFromJSON(h.Ctx(), fmt.Sprintf(`{"title": "Post 2", "tags": [{"id": %d}]}`, otherTagID)))
	u.Must(postModel.CreateFromJSON(h.Ctx(), fmt.Sprintf(`{"title": "Post 3", "tags": [{"id": %d}]}`, targetTagID)))
	u.Must(postModel.CreateFromJSON(h.Ctx(), fmt.Sprintf(`{"title": "Post 4", "tags": [{"id": %d}, {"id": %d}]}`, targetTagID, otherTagID)))

	results := u.Must(postModel.Query(db.NEQ("tags.id", targetTagID)).Select("title").Get(h.Ctx()))
	require.Len(t, results, 2)
	titles := extractAndSortTitles(t, results)
	assert.Equal(t, []string{"Post 1", "Post 2"}, titles)

	resultsByTitle := u.Must(postModel.Query(db.NEQ("tags.label", "Tag 1")).Select("title").Get(h.Ctx()))
	require.Len(t, resultsByTitle, 2)
	titlesByLabel := extractAndSortTitles(t, resultsByTitle)
	assert.Equal(t, []string{"Post 1", "Post 2"}, titlesByLabel)
}

func testM2MNEQCustom(t *testing.T, client h.DBClient) {
	h.ClearDBData(client.C, "post_legacy_tags", "posts_tags", "posts", "tags")

	postModel := u.Must(client.C.Model("post"))
	tagModel := u.Must(client.C.Model("tag"))

	targetTagID := u.Must(tagModel.CreateFromJSON(h.Ctx(), `{"label": "Tag 3"}`))
	otherTagID := u.Must(tagModel.CreateFromJSON(h.Ctx(), `{"label": "Tag 4"}`))

	u.Must(postModel.CreateFromJSON(h.Ctx(), `{"title": "Post 5"}`))
	u.Must(postModel.CreateFromJSON(h.Ctx(), fmt.Sprintf(`{"title": "Post 6", "legacy_tags": [{"id": %d}]}`, otherTagID)))
	u.Must(postModel.CreateFromJSON(h.Ctx(), fmt.Sprintf(`{"title": "Post 7", "legacy_tags": [{"id": %d}]}`, targetTagID)))
	u.Must(postModel.CreateFromJSON(h.Ctx(), fmt.Sprintf(`{"title": "Post 8", "legacy_tags": [{"id": %d}, {"id": %d}]}`, targetTagID, otherTagID)))

	results := u.Must(postModel.Query(db.NEQ("legacy_tags.id", targetTagID)).Select("title").Get(h.Ctx()))
	require.Len(t, results, 2)
	titles := extractAndSortTitles(t, results)
	assert.Equal(t, []string{"Post 5", "Post 6"}, titles)

	resultsByLabel := u.Must(postModel.Query(db.NEQ("legacy_tags.label", "Tag 3")).Select("title").Get(h.Ctx()))
	require.Len(t, resultsByLabel, 2)
	titlesByLabel := extractAndSortTitles(t, resultsByLabel)
	assert.Equal(t, []string{"Post 5", "Post 6"}, titlesByLabel)
}

func extractAndSortTitles(t *testing.T, entities []*entity.Entity) []string {
	t.Helper()
	titles := make([]string, len(entities))
	for i, e := range entities {
		title, ok := e.Get("title").(string)
		require.True(t, ok)
		titles[i] = title
	}
	sort.Strings(titles)
	return titles
}

// testImplicitPKFilter tests filtering by relation field without dot notation
// E.g. {"tags": 1} should be interpreted as {"tags.id": 1}
func testImplicitPKFilter(t *testing.T, client h.DBClient) {
	h.ClearDBData(client.C, "post_legacy_tags", "posts_tags", "posts", "tags")

	postModel := u.Must(client.C.Model("post"))
	tagModel := u.Must(client.C.Model("tag"))

	// Create tags
	tag1ID := u.Must(tagModel.CreateFromJSON(h.Ctx(), `{"label": "Implicit Tag 1"}`))
	tag2ID := u.Must(tagModel.CreateFromJSON(h.Ctx(), `{"label": "Implicit Tag 2"}`))
	tag3ID := u.Must(tagModel.CreateFromJSON(h.Ctx(), `{"label": "Implicit Tag 3"}`))

	// Create posts with different tag combinations
	u.Must(postModel.CreateFromJSON(h.Ctx(), `{"title": "Implicit Post 1"}`))                                                                // No tags
	u.Must(postModel.CreateFromJSON(h.Ctx(), fmt.Sprintf(`{"title": "Implicit Post 2", "tags": [{"id": %d}]}`, tag1ID)))                     // Tag 1
	u.Must(postModel.CreateFromJSON(h.Ctx(), fmt.Sprintf(`{"title": "Implicit Post 3", "tags": [{"id": %d}]}`, tag2ID)))                     // Tag 2
	u.Must(postModel.CreateFromJSON(h.Ctx(), fmt.Sprintf(`{"title": "Implicit Post 4", "tags": [{"id": %d}, {"id": %d}]}`, tag1ID, tag2ID))) // Tag 1 & 2

	// Test implicit PK filter with $eq (primitive value)
	// {"tags": tag1ID} should be equivalent to {"tags.id": tag1ID}
	results := u.Must(postModel.Query(db.EQ("tags", tag1ID)).Select("title").Get(h.Ctx()))
	require.Len(t, results, 2, "Should find posts with tag1")
	titles := extractAndSortTitles(t, results)
	assert.Equal(t, []string{"Implicit Post 2", "Implicit Post 4"}, titles)

	// Test implicit PK filter with $in operator
	// {"tags": {"$in": [tag1ID, tag3ID]}} should be equivalent to {"tags.id": {"$in": [tag1ID, tag3ID]}}
	resultsIn := u.Must(postModel.Query(db.In("tags", []any{tag1ID, tag3ID})).Select("title").Get(h.Ctx()))
	require.Len(t, resultsIn, 2, "Should find posts with tag1 or tag3")
	titlesIn := extractAndSortTitles(t, resultsIn)
	assert.Equal(t, []string{"Implicit Post 2", "Implicit Post 4"}, titlesIn)

	// Test implicit PK filter with $neq operator
	// {"tags": {"$neq": tag1ID}} should exclude posts that have tag1
	resultsNeq := u.Must(postModel.Query(db.NEQ("tags", tag1ID)).Select("title").Get(h.Ctx()))
	require.Len(t, resultsNeq, 2, "Should find posts without tag1")
	titlesNeq := extractAndSortTitles(t, resultsNeq)
	assert.Equal(t, []string{"Implicit Post 1", "Implicit Post 3"}, titlesNeq)

	// Backward compatibility: explicit PK field still works
	resultsExplicit := u.Must(postModel.Query(db.EQ("tags.id", tag2ID)).Select("title").Get(h.Ctx()))
	require.Len(t, resultsExplicit, 2, "Explicit tags.id should still work")
	titlesExplicit := extractAndSortTitles(t, resultsExplicit)
	assert.Equal(t, []string{"Implicit Post 3", "Implicit Post 4"}, titlesExplicit)
}

// testImplicitPKFilterWithOperators tests various operators with implicit PK filter
func testImplicitPKFilterWithOperators(t *testing.T, client h.DBClient) {
	h.ClearDBData(client.C, "post_legacy_tags", "posts_tags", "posts", "tags")

	postModel := u.Must(client.C.Model("post"))
	tagModel := u.Must(client.C.Model("tag"))

	// Create tags with sequential IDs
	tag1ID := u.Must(tagModel.CreateFromJSON(h.Ctx(), `{"label": "Op Tag 1"}`))
	tag2ID := u.Must(tagModel.CreateFromJSON(h.Ctx(), `{"label": "Op Tag 2"}`))
	_ = u.Must(tagModel.CreateFromJSON(h.Ctx(), `{"label": "Op Tag 3"}`))

	// Create posts
	u.Must(postModel.CreateFromJSON(h.Ctx(), `{"title": "Op Post 1"}`))
	u.Must(postModel.CreateFromJSON(h.Ctx(), fmt.Sprintf(`{"title": "Op Post 2", "tags": [{"id": %d}]}`, tag1ID)))
	u.Must(postModel.CreateFromJSON(h.Ctx(), fmt.Sprintf(`{"title": "Op Post 3", "tags": [{"id": %d}]}`, tag2ID)))

	// Test $nin (not in) operator
	resultsNin := u.Must(postModel.Query(db.NotIn("tags", []any{tag1ID})).Select("title").Get(h.Ctx()))
	require.Len(t, resultsNin, 2, "$nin should exclude posts with tag1")
	titlesNin := extractAndSortTitles(t, resultsNin)
	assert.Equal(t, []string{"Op Post 1", "Op Post 3"}, titlesNin)

	// Test $gt (greater than) operator - useful for int PKs
	resultsGt := u.Must(postModel.Query(db.GT("tags", tag1ID)).Select("title").Get(h.Ctx()))
	require.Len(t, resultsGt, 1, "$gt should find posts with tag id > tag1ID")
	assert.Equal(t, "Op Post 3", resultsGt[0].Get("title"))

	// Test $gte (greater than or equal) operator
	resultsGte := u.Must(postModel.Query(db.GTE("tags", tag1ID)).Select("title").Get(h.Ctx()))
	require.Len(t, resultsGte, 2, "$gte should find posts with tag id >= tag1ID")
	titlesGte := extractAndSortTitles(t, resultsGte)
	assert.Equal(t, []string{"Op Post 2", "Op Post 3"}, titlesGte)

	// Test $lt (less than) operator
	resultsLt := u.Must(postModel.Query(db.LT("tags", tag2ID)).Select("title").Get(h.Ctx()))
	require.Len(t, resultsLt, 1, "$lt should find posts with tag id < tag2ID")
	assert.Equal(t, "Op Post 2", resultsLt[0].Get("title"))

	// Test $lte (less than or equal) operator
	resultsLte := u.Must(postModel.Query(db.LTE("tags", tag2ID)).Select("title").Get(h.Ctx()))
	require.Len(t, resultsLte, 2, "$lte should find posts with tag id <= tag2ID")
	titlesLte := extractAndSortTitles(t, resultsLte)
	assert.Equal(t, []string{"Op Post 2", "Op Post 3"}, titlesLte)
}

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

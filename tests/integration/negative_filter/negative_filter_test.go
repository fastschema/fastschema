package negative_filter

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
	schemaDir    = "../../../tests/integration/negative_filter/data/schemas"
	migrationDir = "../../../tests/integration/negative_filter/data/migrations"
	sqliteDSN    = "../../../tests/integration/negative_filter/data/negative_filter_test.db"
)

func TestMySQL(t *testing.T) {
	runNegativeFilterTests(t, u.Map(h.MysqlConfigs, func(cfg h.DBConfig) h.DBClient {
		sb := u.Must(schema.NewBuilderFromDir(schemaDir))
		return h.NewMySQLClient(t, cfg, sb, migrationDir)
	}))
}

func TestPostgres(t *testing.T) {
	runNegativeFilterTests(t, u.Map(h.PostgresConfigs, func(cfg h.DBConfig) h.DBClient {
		sb := u.Must(schema.NewBuilderFromDir(schemaDir))
		return h.NewPostgresClient(t, cfg, sb, migrationDir)
	}))
}

func TestSQLite(t *testing.T) {
	sb := u.Must(schema.NewBuilderFromDir(schemaDir))
	client := h.NewSQLiteClient(t, "sqlite", sqliteDSN, migrationDir, sb)
	runNegativeFilterTests(t, []h.DBClient{client})
}

func runNegativeFilterTests(t *testing.T, clients []h.DBClient) {
	for _, client := range clients {
		client := client
		t.Run(client.Name, func(t *testing.T) {
			t.Run("O2MNEQUsesAntiJoin", func(t *testing.T) {
				testO2MNEQAntiExists(t, client)
			})

			t.Run("O2MNotLike", func(t *testing.T) {
				testO2MNotLike(t, client)
			})

			t.Run("O2MNotContainsFold", func(t *testing.T) {
				testO2MNotContainsFold(t, client)
			})

			t.Run("M2MNEQUsesAntiJoin", func(t *testing.T) {
				testM2MNEQAntiExists(t, client)
			})

			t.Run("M2MNotLike", func(t *testing.T) {
				testM2MNotLike(t, client)
			})

			t.Run("M2MNotContains", func(t *testing.T) {
				testM2MNotContains(t, client)
			})

			t.Run("M2MNIN", func(t *testing.T) {
				testM2MNIN(t, client)
			})

			t.Run("M2ONEQUsesAntiJoin", func(t *testing.T) {
				testM2ONEQAntiExists(t, client)
			})

			t.Run("M2ONotLike", func(t *testing.T) {
				testM2ONotLike(t, client)
			})

			t.Run("M2ONotContainsFold", func(t *testing.T) {
				testM2ONotContainsFold(t, client)
			})

			t.Run("M2ONIN", func(t *testing.T) {
				testM2ONIN(t, client)
			})
		})
	}
}

func testO2MNEQAntiExists(t *testing.T, client h.DBClient) {
	h.ClearDBData(client.C, "countries", "regions")

	regionModel := u.Must(client.C.Model("region"))
	countryModel := u.Must(client.C.Model("country"))

	usID := h.IDUint64(t, u.Must(regionModel.CreateFromJSON(h.Ctx(), `{"name": "US"}`)))
	asiaID := h.IDUint64(t, u.Must(regionModel.CreateFromJSON(h.Ctx(), `{"name": "Asia"}`)))
	_ = u.Must(regionModel.CreateFromJSON(h.Ctx(), `{"name": "Antarctica"}`))

	targetCountryID := h.IDUint64(t, u.Must(countryModel.CreateFromJSON(h.Ctx(), fmt.Sprintf(`{"name": "USA", "region_id": %d}`, usID))))
	_ = u.Must(countryModel.CreateFromJSON(h.Ctx(), fmt.Sprintf(`{"name": "Canada", "region_id": %d}`, usID)))
	_ = u.Must(countryModel.CreateFromJSON(h.Ctx(), fmt.Sprintf(`{"name": "Vietnam", "region_id": %d}`, asiaID)))

	results := u.Must(regionModel.Query(db.NEQ("countries.id", targetCountryID)).Select("name").Get(h.Ctx()))

	names := extractAndSortField(t, results, "name")
	require.Len(t, names, 2)
	assert.Equal(t, []string{"Antarctica", "Asia"}, names)
}

func testO2MNotLike(t *testing.T, client h.DBClient) {
	h.ClearDBData(client.C, "countries", "regions")

	regionModel := u.Must(client.C.Model("region"))
	countryModel := u.Must(client.C.Model("country"))

	usID := h.IDUint64(t, u.Must(regionModel.CreateFromJSON(h.Ctx(), `{"name": "US"}`)))
	_ = u.Must(regionModel.CreateFromJSON(h.Ctx(), `{"name": "Asia"}`))
	_ = u.Must(regionModel.CreateFromJSON(h.Ctx(), `{"name": "Antarctica"}`))

	_ = u.Must(countryModel.CreateFromJSON(h.Ctx(), fmt.Sprintf(`{"name": "USA", "region_id": %d}`, usID)))
	_ = u.Must(countryModel.CreateFromJSON(h.Ctx(), fmt.Sprintf(`{"name": "Canada", "region_id": %d}`, usID)))

	results := u.Must(regionModel.Query(db.NotLike("countries.name", "USA")).Select("name").Get(h.Ctx()))
	names := extractAndSortField(t, results, "name")
	require.Len(t, names, 2)
	assert.Equal(t, []string{"Antarctica", "Asia"}, names)
}

func testO2MNotContainsFold(t *testing.T, client h.DBClient) {
	h.ClearDBData(client.C, "countries", "regions")

	regionModel := u.Must(client.C.Model("region"))
	countryModel := u.Must(client.C.Model("country"))

	usID := h.IDUint64(t, u.Must(regionModel.CreateFromJSON(h.Ctx(), `{"name": "US"}`)))
	_ = u.Must(regionModel.CreateFromJSON(h.Ctx(), `{"name": "Asia"}`))
	_ = u.Must(regionModel.CreateFromJSON(h.Ctx(), `{"name": "Antarctica"}`))

	_ = u.Must(countryModel.CreateFromJSON(h.Ctx(), fmt.Sprintf(`{"name": "USA", "region_id": %d}`, usID)))
	_ = u.Must(countryModel.CreateFromJSON(h.Ctx(), fmt.Sprintf(`{"name": "Canada", "region_id": %d}`, usID)))

	results := u.Must(regionModel.Query(db.NotContainsFold("countries.name", "usa")).Select("name").Get(h.Ctx()))
	names := extractAndSortField(t, results, "name")
	require.Len(t, names, 2)
	assert.Equal(t, []string{"Antarctica", "Asia"}, names)
}

func testM2MNEQAntiExists(t *testing.T, client h.DBClient) {
	h.ClearDBData(client.C, "posts_tags", "posts", "tags")

	postModel := u.Must(client.C.Model("post"))
	tagModel := u.Must(client.C.Model("tag"))

	targetTagID := h.IDUint64(t, u.Must(tagModel.CreateFromJSON(h.Ctx(), `{"label": "Target"}`)))
	otherTagID := h.IDUint64(t, u.Must(tagModel.CreateFromJSON(h.Ctx(), `{"label": "Other"}`)))

	_ = u.Must(postModel.CreateFromJSON(h.Ctx(), `{"title": "Post 1"}`))
	_ = u.Must(postModel.CreateFromJSON(h.Ctx(), fmt.Sprintf(`{"title": "Post 2", "tags": [{"id": %d}]}`, otherTagID)))
	_ = u.Must(postModel.CreateFromJSON(h.Ctx(), fmt.Sprintf(`{"title": "Post 3", "tags": [{"id": %d}]}`, targetTagID)))
	_ = u.Must(postModel.CreateFromJSON(h.Ctx(), fmt.Sprintf(`{"title": "Post 4", "tags": [{"id": %d}, {"id": %d}]}`, targetTagID, otherTagID)))

	results := u.Must(postModel.Query(db.NEQ("tags.id", targetTagID)).Select("title").Get(h.Ctx()))
	titles := extractAndSortField(t, results, "title")
	require.Len(t, titles, 2)
	assert.Equal(t, []string{"Post 1", "Post 2"}, titles)
}

func testM2MNotLike(t *testing.T, client h.DBClient) {
	h.ClearDBData(client.C, "posts_tags", "posts", "tags")

	postModel := u.Must(client.C.Model("post"))
	tagModel := u.Must(client.C.Model("tag"))

	targetTagID := h.IDUint64(t, u.Must(tagModel.CreateFromJSON(h.Ctx(), `{"label": "Target"}`)))
	otherTagID := h.IDUint64(t, u.Must(tagModel.CreateFromJSON(h.Ctx(), `{"label": "Other"}`)))

	_ = u.Must(postModel.CreateFromJSON(h.Ctx(), `{"title": "Post 1"}`))
	_ = u.Must(postModel.CreateFromJSON(h.Ctx(), fmt.Sprintf(`{"title": "Post 2", "tags": [{"id": %d}]}`, otherTagID)))
	_ = u.Must(postModel.CreateFromJSON(h.Ctx(), fmt.Sprintf(`{"title": "Post 3", "tags": [{"id": %d}]}`, targetTagID)))

	results := u.Must(postModel.Query(db.NotLike("tags.label", "Target")).Select("title").Get(h.Ctx()))
	titles := extractAndSortField(t, results, "title")
	require.Len(t, titles, 2)
	assert.Equal(t, []string{"Post 1", "Post 2"}, titles)
}

func testM2MNotContains(t *testing.T, client h.DBClient) {
	h.ClearDBData(client.C, "posts_tags", "posts", "tags")

	postModel := u.Must(client.C.Model("post"))
	tagModel := u.Must(client.C.Model("tag"))

	targetTagID := h.IDUint64(t, u.Must(tagModel.CreateFromJSON(h.Ctx(), `{"label": "TargetLabel"}`)))
	otherTagID := h.IDUint64(t, u.Must(tagModel.CreateFromJSON(h.Ctx(), `{"label": "Other"}`)))

	_ = u.Must(postModel.CreateFromJSON(h.Ctx(), `{"title": "Post 1"}`))
	_ = u.Must(postModel.CreateFromJSON(h.Ctx(), fmt.Sprintf(`{"title": "Post 2", "tags": [{"id": %d}]}`, otherTagID)))
	_ = u.Must(postModel.CreateFromJSON(h.Ctx(), fmt.Sprintf(`{"title": "Post 3", "tags": [{"id": %d}]}`, targetTagID)))
	_ = u.Must(postModel.CreateFromJSON(h.Ctx(), fmt.Sprintf(`{"title": "Post 4", "tags": [{"id": %d}, {"id": %d}]}`, targetTagID, otherTagID)))

	results := u.Must(postModel.Query(db.NotContains("tags.label", "Target")).Select("title").Get(h.Ctx()))
	titles := extractAndSortField(t, results, "title")
	require.Len(t, titles, 2)
	assert.Equal(t, []string{"Post 1", "Post 2"}, titles)
}

func testM2MNIN(t *testing.T, client h.DBClient) {
	h.ClearDBData(client.C, "posts_tags", "posts", "tags")

	postModel := u.Must(client.C.Model("post"))
	tagModel := u.Must(client.C.Model("tag"))

	targetTagID := h.IDUint64(t, u.Must(tagModel.CreateFromJSON(h.Ctx(), `{"label": "Target"}`)))
	otherTagID := h.IDUint64(t, u.Must(tagModel.CreateFromJSON(h.Ctx(), `{"label": "Other"}`)))
	_ = h.IDUint64(t, u.Must(tagModel.CreateFromJSON(h.Ctx(), `{"label": "Third"}`)))

	_ = u.Must(postModel.CreateFromJSON(h.Ctx(), `{"title": "Post 1"}`))
	_ = u.Must(postModel.CreateFromJSON(h.Ctx(), fmt.Sprintf(`{"title": "Post 2", "tags": [{"id": %d}]}`, otherTagID)))
	_ = u.Must(postModel.CreateFromJSON(h.Ctx(), fmt.Sprintf(`{"title": "Post 3", "tags": [{"id": %d}]}`, targetTagID)))
	_ = u.Must(postModel.CreateFromJSON(h.Ctx(), fmt.Sprintf(`{"title": "Post 4", "tags": [{"id": %d}, {"id": %d}]}`, targetTagID, otherTagID)))

	results := u.Must(postModel.Query(db.NotIn("tags.id", []any{targetTagID})).Select("title").Get(h.Ctx()))
	titles := extractAndSortField(t, results, "title")
	require.Len(t, titles, 2)
	assert.Equal(t, []string{"Post 1", "Post 2"}, titles)
}

func testM2ONEQAntiExists(t *testing.T, client h.DBClient) {
	h.ClearDBData(client.C, "countries", "regions")

	regionModel := u.Must(client.C.Model("region"))
	countryModel := u.Must(client.C.Model("country"))

	usID := h.IDUint64(t, u.Must(regionModel.CreateFromJSON(h.Ctx(), `{"name": "US"}`)))
	asiaID := h.IDUint64(t, u.Must(regionModel.CreateFromJSON(h.Ctx(), `{"name": "Asia"}`)))

	_ = u.Must(countryModel.CreateFromJSON(h.Ctx(), fmt.Sprintf(`{"name": "USA", "region_id": %d}`, usID)))
	_ = u.Must(countryModel.CreateFromJSON(h.Ctx(), fmt.Sprintf(`{"name": "Canada", "region_id": %d}`, usID)))
	_ = u.Must(countryModel.CreateFromJSON(h.Ctx(), fmt.Sprintf(`{"name": "Vietnam", "region_id": %d}`, asiaID)))

	results := u.Must(countryModel.Query(db.NEQ("region.id", usID)).Select("name").Get(h.Ctx()))
	names := extractAndSortField(t, results, "name")
	require.Len(t, names, 1)
	assert.Equal(t, []string{"Vietnam"}, names)
}

func testM2ONotLike(t *testing.T, client h.DBClient) {
	h.ClearDBData(client.C, "countries", "regions")

	regionModel := u.Must(client.C.Model("region"))
	countryModel := u.Must(client.C.Model("country"))

	usID := h.IDUint64(t, u.Must(regionModel.CreateFromJSON(h.Ctx(), `{"name": "US"}`)))
	asiaID := h.IDUint64(t, u.Must(regionModel.CreateFromJSON(h.Ctx(), `{"name": "Asia"}`)))

	_ = u.Must(countryModel.CreateFromJSON(h.Ctx(), fmt.Sprintf(`{"name": "USA", "region_id": %d}`, usID)))
	_ = u.Must(countryModel.CreateFromJSON(h.Ctx(), fmt.Sprintf(`{"name": "Canada", "region_id": %d}`, usID)))
	_ = u.Must(countryModel.CreateFromJSON(h.Ctx(), fmt.Sprintf(`{"name": "Vietnam", "region_id": %d}`, asiaID)))

	results := u.Must(countryModel.Query(db.NotLike("region.name", "US")).Select("name").Get(h.Ctx()))
	names := extractAndSortField(t, results, "name")
	require.Len(t, names, 1)
	assert.Equal(t, []string{"Vietnam"}, names)
}

func testM2ONotContainsFold(t *testing.T, client h.DBClient) {
	h.ClearDBData(client.C, "countries", "regions")

	regionModel := u.Must(client.C.Model("region"))
	countryModel := u.Must(client.C.Model("country"))

	usID := h.IDUint64(t, u.Must(regionModel.CreateFromJSON(h.Ctx(), `{"name": "US"}`)))
	asiaID := h.IDUint64(t, u.Must(regionModel.CreateFromJSON(h.Ctx(), `{"name": "Asia"}`)))

	_ = u.Must(countryModel.CreateFromJSON(h.Ctx(), fmt.Sprintf(`{"name": "USA", "region_id": %d}`, usID)))
	_ = u.Must(countryModel.CreateFromJSON(h.Ctx(), fmt.Sprintf(`{"name": "Canada", "region_id": %d}`, usID)))
	_ = u.Must(countryModel.CreateFromJSON(h.Ctx(), fmt.Sprintf(`{"name": "Vietnam", "region_id": %d}`, asiaID)))

	results := u.Must(countryModel.Query(db.NotContainsFold("region.name", "us")).Select("name").Get(h.Ctx()))
	names := extractAndSortField(t, results, "name")
	require.Len(t, names, 1)
	assert.Equal(t, []string{"Vietnam"}, names)
}

func testM2ONIN(t *testing.T, client h.DBClient) {
	h.ClearDBData(client.C, "countries", "regions")

	regionModel := u.Must(client.C.Model("region"))
	countryModel := u.Must(client.C.Model("country"))

	usID := h.IDUint64(t, u.Must(regionModel.CreateFromJSON(h.Ctx(), `{"name": "US"}`)))
	asiaID := h.IDUint64(t, u.Must(regionModel.CreateFromJSON(h.Ctx(), `{"name": "Asia"}`)))
	_ = u.Must(regionModel.CreateFromJSON(h.Ctx(), `{"name": "Africa"}`))

	_ = u.Must(countryModel.CreateFromJSON(h.Ctx(), fmt.Sprintf(`{"name": "USA", "region_id": %d}`, usID)))
	_ = u.Must(countryModel.CreateFromJSON(h.Ctx(), fmt.Sprintf(`{"name": "Canada", "region_id": %d}`, usID)))
	_ = u.Must(countryModel.CreateFromJSON(h.Ctx(), fmt.Sprintf(`{"name": "Vietnam", "region_id": %d}`, asiaID)))

	results := u.Must(countryModel.Query(db.NotIn("region.id", []any{usID})).Select("name").Get(h.Ctx()))
	names := extractAndSortField(t, results, "name")
	require.Len(t, names, 1)
	assert.Equal(t, []string{"Vietnam"}, names)
}

func extractAndSortField(t *testing.T, entities []*entity.Entity, field string) []string {
	t.Helper()
	values := make([]string, len(entities))
	for i, e := range entities {
		v, ok := e.Get(field).(string)
		require.True(t, ok)
		values[i] = v
	}
	sort.Strings(values)
	return values
}

package relation

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fastschema/fastschema/db"
	u "github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
	h "github.com/fastschema/fastschema/tests/integration/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	schemaDir    = "../../../tests/integration/ondelete/data/schemas"
	migrationDir = "../../../tests/integration/ondelete/data/migrations"
	sqliteDSN    = "../../../tests/integration/ondelete/data/ondelete_test.db"
)

var testCases = []struct {
	name string
	fn   func(t *testing.T, client h.DBClient)
}{
	{"OnDeleteSetNull", testOnDeleteSetNull},
	{"OnDeleteRestrict", testOnDeleteRestrict},
	{"OnDeleteNoAction", testOnDeleteNoAction},
	{"OnDeleteCascade", testOnDeleteCascade},
	{"OnDeleteSetDefault", testOnDeleteSetDefault},
}

func newMysqlSchemaBuilder(t *testing.T) *schema.Builder {
	files := u.Must(filepath.Glob(path.Join(schemaDir, "*.json")))
	schemaMap := make(map[string]*schema.Schema)
	for _, file := range files {
		if strings.HasSuffix(file, "post_delete_setdefault.json") {
			continue
		}
		s := u.Must(schema.NewSchemaFromJSONFile(file))
		schemaMap[s.Name] = s
	}

	delete(schemaMap, "post_delete_setdefault")
	if categorySchema, ok := schemaMap["category"]; ok {
		filteredFields := u.Filter(categorySchema.Fields, func(f *schema.Field) bool {
			return f.Name != "post_delete_setdefault"
		})
		categorySchema.Fields = filteredFields
	}

	sb, err := schema.NewBuilderFromSchemas("", schemaMap)
	require.NoError(t, err)
	return sb
}

func TestOnDeleteMysql(t *testing.T) {
	runTests(t, u.Map(h.MysqlConfigs, func(sc h.DBConfig) h.DBClient {
		sb := newMysqlSchemaBuilder(t)
		return h.NewMySQLClient(t, sc, sb, migrationDir)
	}))
}

func TestOnDeletePostgres(t *testing.T) {
	runTests(t, u.Map(h.PostgresConfigs, func(sc h.DBConfig) h.DBClient {
		sb := u.Must(schema.NewBuilderFromDir(schemaDir))
		return h.NewPostgresClient(t, sc, sb, migrationDir)
	}))
}

func TestOnDeleteSQLite(t *testing.T) {
	sb := u.Must(schema.NewBuilderFromDir(schemaDir))
	client := h.NewSQLiteClient(t, "sqlite", sqliteDSN, migrationDir, sb)
	runTests(t, []h.DBClient{client})
}

func runTests(t *testing.T, clients []h.DBClient) {
	for _, client := range clients {
		for _, tc := range testCases {
			t.Run(client.Name+"/"+tc.name, func(t *testing.T) {
				tables := []string{
					"post_delete_setnull",
					"post_delete_strict",
					"post_delete_noaction",
					"post_delete_cascade",
					"categories",
				}

				if !h.IsMySQLFamily(client.Name) {
					tables = append(tables, "post_delete_setdefault")
				}
				h.ClearDBData(client.C, tables...)
				tc.fn(t, client)
			})
		}
	}
}

func testOnDeleteSetNull(t *testing.T, client h.DBClient) {
	catModel := u.Must(client.C.Model("category"))
	postModel := u.Must(client.C.Model("post_delete_setnull"))

	catID := u.Must(catModel.CreateFromJSON(
		h.Ctx(),
		`{"name":"Category SetNull"}`,
	))
	postID := u.Must(postModel.CreateFromJSON(
		h.Ctx(),
		fmt.Sprintf(`{"name":"Post SetNull","category":{"id": %d}}`, catID),
	))

	// Deleting the category should set the category_id in post to NULL
	affected := u.Must(catModel.Mutation().Where(db.EQ("id", catID)).Delete(h.Ctx()))
	require.Equal(t, 1, affected)

	// Verify that the post's category_id is now NULL
	post := u.Must(postModel.Query(db.EQ("id", postID)).First(h.Ctx()))
	assert.Nil(t, post.Get("category_id"))
}

func testOnDeleteRestrict(t *testing.T, client h.DBClient) {
	catModel := u.Must(client.C.Model("category"))
	postModel := u.Must(client.C.Model("post_delete_strict"))

	catID := u.Must(catModel.CreateFromJSON(
		h.Ctx(),
		`{"name":"Category Restrict"}`,
	))
	_ = u.Must(postModel.CreateFromJSON(
		h.Ctx(),
		fmt.Sprintf(`{"name":"Post Restrict","category":{"id": %d}}`, catID),
	))

	// Attempting to delete the category should fail due to RESTRICT constraint
	_, err := catModel.Mutation().Where(db.EQ("id", catID)).Delete(h.Ctx())
	require.Error(t, err)

	// Verify that the category still exists
	cat := u.Must(catModel.Query(db.EQ("id", catID)).First(h.Ctx()))
	assert.Equal(t, catID, cat.ID())

	// Verify that the post still exists with the correct category_id
	posts := u.Must(postModel.Query(db.EQ("category_id", catID)).Get(h.Ctx()))
	require.Len(t, posts, 1)
	assert.Equal(t, catID, posts[0].Get("category_id"))
}

func testOnDeleteNoAction(t *testing.T, client h.DBClient) {
	catModel := u.Must(client.C.Model("category"))
	postModel := u.Must(client.C.Model("post_delete_noaction"))
	catID := u.Must(catModel.CreateFromJSON(
		h.Ctx(),
		`{"name":"Category NoAction"}`,
	))
	postID := u.Must(postModel.CreateFromJSON(
		h.Ctx(),
		fmt.Sprintf(`{"name":"Post NoAction","category":{"id": %d}}`, catID),
	))

	// Attempting to delete the category should fail due to NO ACTION constraint
	_, err := catModel.Mutation().Where(db.EQ("id", catID)).Delete(h.Ctx())
	require.Error(t, err)

	// Verify that the category still exists
	cat := u.Must(catModel.Query(db.EQ("id", catID)).First(h.Ctx()))
	assert.Equal(t, catID, cat.ID())

	// Verify that the post still exists with the correct category_id
	post := u.Must(postModel.Query(db.EQ("id", postID)).First(h.Ctx()))
	assert.Equal(t, catID, post.Get("category_id"))
}

func testOnDeleteCascade(t *testing.T, client h.DBClient) {
	catModel := u.Must(client.C.Model("category"))
	postModel := u.Must(client.C.Model("post_delete_cascade"))

	catID := u.Must(catModel.CreateFromJSON(
		h.Ctx(),
		`{"name":"Category Cascade"}`,
	))
	postID := u.Must(postModel.CreateFromJSON(
		h.Ctx(),
		fmt.Sprintf(`{"name":"Post Cascade","category":{"id": %d}}`, catID),
	))

	// Deleting the category should also delete the associated post due to CASCADE constraint
	affected := u.Must(catModel.Mutation().Where(db.EQ("id", catID)).Delete(h.Ctx()))
	require.Equal(t, 1, affected)

	// Verify that the post has been deleted
	_, err := postModel.Query(db.EQ("id", postID)).First(h.Ctx())
	require.Error(t, err)
}

func testOnDeleteSetDefault(t *testing.T, client h.DBClient) {
	if h.IsMySQLFamily(client.Name) {
		t.Skip("ON DELETE SET DEFAULT is not supported by " + client.Name)
	}

	catModel := u.Must(client.C.Model("category"))
	postModel := u.Must(client.C.Model("post_delete_setdefault"))

	// post_delete_setdefault.json defines category_id with DEFAULT 5.
	// Create a category with ID 5 to be the default.
	defaultCatID := u.Must(catModel.CreateFromJSON(
		h.Ctx(),
		`{"id":5,"name":"Fallback Category"}`,
	))
	require.Equal(t, uint64(5), defaultCatID)

	catID := u.Must(catModel.CreateFromJSON(
		h.Ctx(),
		`{"id":6,"name":"Category SetDefault"}`,
	))
	postID := u.Must(postModel.CreateFromJSON(
		h.Ctx(),
		fmt.Sprintf(`{"name":"Post SetDefault","category":{"id": %d}}`, catID),
	))

	// Deleting the category should set the category_id in post to the default value (5)
	affected := u.Must(catModel.Mutation().Where(db.EQ("id", catID)).Delete(h.Ctx()))
	require.Equal(t, 1, affected)

	// Verify that the post's category_id is now set to the default category ID
	post := u.Must(postModel.Query(db.EQ("id", postID)).First(h.Ctx()))
	assert.Equal(t, defaultCatID, post.Get("category_id"))
}

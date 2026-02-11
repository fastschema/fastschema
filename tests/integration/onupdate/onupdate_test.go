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
	onUpdateSchemaDir    = "../../../tests/integration/onupdate/data/schemas"
	onUpdateMigrationDir = "../../../tests/integration/onupdate/data/migrations"
	sqliteDSN            = "../../../tests/integration/onupdate/data/onupdate_test.db"
)

var onUpdateTestCases = []struct {
	name string
	fn   func(t *testing.T, client h.DBClient)
}{
	{"OnUpdateSetNull", testOnUpdateSetNull},
	{"OnUpdateRestrict", testOnUpdateRestrict},
	{"OnUpdateNoAction", testOnUpdateNoAction},
	{"OnUpdateCascade", testOnUpdateCascade},
	{"OnUpdateSetDefault", testOnUpdateSetDefault},
}

func TestOnUpdateMysql(t *testing.T) {
	runOnUpdateTests(t, u.Map(h.MysqlConfigs, func(sc h.DBConfig) h.DBClient {
		sb := newMysqlSchemaBuilder(t)
		return h.NewMySQLClient(t, sc, sb, onUpdateMigrationDir)
	}))
}

func TestOnUpdatePostgres(t *testing.T) {
	runOnUpdateTests(t, u.Map(h.PostgresConfigs, func(sc h.DBConfig) h.DBClient {
		sb := u.Must(schema.NewBuilderFromDir(onUpdateSchemaDir))
		return h.NewPostgresClient(t, sc, sb, onUpdateMigrationDir)
	}))
}

func TestOnUpdateSQLite(t *testing.T) {
	sb := u.Must(schema.NewBuilderFromDir(onUpdateSchemaDir))
	client := h.NewSQLiteClient(t, "sqlite", sqliteDSN, onUpdateMigrationDir, sb)
	runOnUpdateTests(t, []h.DBClient{client})
}

func runOnUpdateTests(t *testing.T, clients []h.DBClient) {
	for _, client := range clients {
		for _, tc := range onUpdateTestCases {
			t.Run(client.Name+"/"+tc.name, func(t *testing.T) {
				tables := []string{
					"categories",
					"post_update_setnull",
					"post_update_restrict",
					"post_update_noaction",
					"post_update_cascade",
				}
				if !h.IsMySQLFamily(client.Name) {
					tables = append(tables, "post_update_setdefault")
				}
				h.ClearDBData(client.C, tables...)
				tc.fn(t, client)
			})
		}
	}
}

func newMysqlSchemaBuilder(t *testing.T) *schema.Builder {
	t.Helper()
	files := u.Must(filepath.Glob(path.Join(onUpdateSchemaDir, "*.json")))
	schemaMap := make(map[string]*schema.Schema)
	for _, file := range files {
		if strings.HasSuffix(file, "post_update_setdefault.json") {
			continue
		}
		s := u.Must(schema.NewSchemaFromJSONFile(file))
		schemaMap[s.Name] = s
	}

	delete(schemaMap, "post_update_setdefault")
	if categorySchema, ok := schemaMap["category"]; ok {
		filteredFields := u.Filter(categorySchema.Fields, func(f *schema.Field) bool {
			return f.Name != "post_update_setdefault"
		})
		categorySchema.Fields = filteredFields
	}

	sb, err := schema.NewBuilderFromSchemas("", schemaMap)
	require.NoError(t, err)
	return sb
}

func testOnUpdateSetNull(t *testing.T, client h.DBClient) {
	catModel := u.Must(client.C.Model("category"))
	postModel := u.Must(client.C.Model("post_update_setnull"))

	catID := u.Must(catModel.CreateFromJSON(
		h.Ctx(),
		`{"name":"Category SetNull"}`,
	))
	postID := u.Must(postModel.CreateFromJSON(
		h.Ctx(),
		fmt.Sprintf(`{"name":"Post SetNull","category":{"id": %d}}`, catID),
	))

	// Update the category primary key
	newID := catID + 100
	require.NoError(t, updateCategoryPrimaryKey(client.C, catID, newID))

	// Verify that the post's category_id is set to NULL
	post := u.Must(postModel.Query(db.EQ("id", postID)).First(h.Ctx()))
	assert.Nil(t, post.Get("category_id"))
}

func testOnUpdateRestrict(t *testing.T, client h.DBClient) {
	catModel := u.Must(client.C.Model("category"))
	postModel := u.Must(client.C.Model("post_update_restrict"))

	catID := u.Must(catModel.CreateFromJSON(
		h.Ctx(),
		`{"name":"Category Restrict"}`,
	))
	u.Must(postModel.CreateFromJSON(
		h.Ctx(),
		fmt.Sprintf(`{"name":"Post Restrict","category":{"id": %d}}`, catID),
	))

	// Attempt to update the category primary key (which should be restricted)
	err := updateCategoryPrimaryKey(client.C, catID, catID+50)
	require.Error(t, err)

	// Verify that the category still exists with the original ID
	cat := u.Must(catModel.Query(db.EQ("id", catID)).First(h.Ctx()))
	assert.Equal(t, catID, cat.ID())

	// Verify that the post still exists with the correct category_id
	posts := u.Must(postModel.Query(db.EQ("category_id", catID)).Get(h.Ctx()))
	require.Len(t, posts, 1)
	assert.Equal(t, catID, posts[0].Get("category_id"))
}

func testOnUpdateNoAction(t *testing.T, client h.DBClient) {
	catModel := u.Must(client.C.Model("category"))
	postModel := u.Must(client.C.Model("post_update_noaction"))

	categoryID := u.Must(catModel.CreateFromJSON(
		h.Ctx(),
		`{"name":"Category NoAction"}`,
	))
	u.Must(postModel.CreateFromJSON(
		h.Ctx(),
		fmt.Sprintf(`{"name":"Post NoAction","category":{"id": %d}}`, categoryID),
	))

	// Attempt to update the category primary key (which should do no action)
	err := updateCategoryPrimaryKey(client.C, categoryID, categoryID+25)
	require.Error(t, err)

	// Verify that the category still exists with the original ID
	cat := u.Must(catModel.Query(db.EQ("id", categoryID)).First(h.Ctx()))
	assert.Equal(t, categoryID, cat.ID())

	// Verify that the post still exists with the correct category_id
	posts := u.Must(postModel.Query(db.EQ("category_id", categoryID)).Get(h.Ctx()))
	require.Len(t, posts, 1)
	assert.Equal(t, categoryID, posts[0].Get("category_id"))
}

func testOnUpdateCascade(t *testing.T, client h.DBClient) {
	catModel := u.Must(client.C.Model("category"))
	postModel := u.Must(client.C.Model("post_update_cascade"))

	catID := u.Must(catModel.CreateFromJSON(
		h.Ctx(),
		`{"name":"Category Cascade"}`,
	))
	postID := u.Must(postModel.CreateFromJSON(
		h.Ctx(),
		fmt.Sprintf(`{"name":"Post Cascade","category":{"id": %d}}`, catID),
	))

	// Update the category primary key
	newID := catID + 75
	require.NoError(t, updateCategoryPrimaryKey(client.C, catID, newID))

	// Verify that the category has the new ID
	cat := u.Must(catModel.Query(db.EQ("id", newID)).First(h.Ctx()))
	assert.Equal(t, newID, cat.ID())

	// Verify that the post's category_id is updated to the new ID
	post := u.Must(postModel.Query(db.EQ("id", postID)).First(h.Ctx()))
	assert.EqualValues(t, newID, post.Get("category_id"))
}

func testOnUpdateSetDefault(t *testing.T, client h.DBClient) {
	if h.IsMySQLFamily(client.Name) {
		t.Skipf("ON UPDATE SET DEFAULT is not supported by %s", client.Name)
	}

	catModel := u.Must(client.C.Model("category"))
	postModel := u.Must(client.C.Model("post_update_setdefault"))

	// post_update_setdefault.json defines category_id with DEFAULT 5.
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

	// Update the category primary key
	newID := catID + 55
	require.NoError(t, updateCategoryPrimaryKey(client.C, catID, newID))

	// Verify that the post's category_id is set to the default category ID
	post := u.Must(postModel.Query(db.EQ("id", postID)).First(h.Ctx()))
	assert.EqualValues(t, defaultCatID, post.Get("category_id"))
}

func updateCategoryPrimaryKey(client db.Client, fromID, toID uint64) error {
	query := fmt.Sprintf("UPDATE categories SET id = %d WHERE id = %d", toID, fromID)
	_, err := client.Exec(h.Ctx(), query)
	return err
}

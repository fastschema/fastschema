package toolservice_test

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/entity"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/logger"
	"github.com/fastschema/fastschema/pkg/entdbadapter"
	"github.com/fastschema/fastschema/pkg/restfulresolver"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
	toolservice "github.com/fastschema/fastschema/services/tool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newActivityTestServer(t *testing.T) (*restfulresolver.Server, db.Client) {
	t.Helper()
	sb := utils.Must(schema.NewBuilderFromDir(t.TempDir(), fs.SystemSchemaTypes...))
	client := utils.Must(entdbadapter.NewTestClient(utils.Must(os.MkdirTemp("", "migrations")), sb))
	toolService := toolservice.New(&testApp{sb: sb, db: client})

	resources := fs.NewResourcesManager()
	resources.Group("tool").
		Add(fs.NewResource("activity", toolService.Activity, &fs.Meta{Get: "/activity"}))

	require.NoError(t, resources.Init())
	restResolver := restfulresolver.NewRestfulResolver(&restfulresolver.ResolverConfig{
		ResourceManager: resources,
		Logger:          logger.CreateMockLogger(true),
	})

	return restResolver.Server(), client
}

func seedActivity(t *testing.T, client db.Client, schemaName, action string, createdAt time.Time) {
	t.Helper()
	_, err := db.Create[*fs.Activity](context.Background(), client, entity.New().
		Set("action", action).
		Set("schema_name", schemaName).
		Set("record_id", schemaName+"-"+action).
		Set("created_at", createdAt))
	require.NoError(t, err)
}

func getActivityList(t *testing.T, server *restfulresolver.Server, url string) *toolservice.ActivityList {
	t.Helper()
	req := httptest.NewRequest("GET", url, nil)
	resp := utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	body := utils.Must(utils.ReadCloserToString(resp.Body))
	require.Equal(t, 200, resp.StatusCode, "body: %s", body)

	var wrapper struct {
		Data *toolservice.ActivityList `json:"data"`
	}
	require.NoError(t, json.Unmarshal([]byte(body), &wrapper))
	return wrapper.Data
}

func TestToolActivityFilterAndPaging(t *testing.T) {
	server, client := newActivityTestServer(t)

	now := time.Now().UTC()
	seedActivity(t, client, "role", fs.ActivityActionCreate, now.Add(-3*time.Hour))
	seedActivity(t, client, "user", fs.ActivityActionUpdate, now.Add(-2*time.Hour))
	seedActivity(t, client, "role", fs.ActivityActionDelete, now.Add(-1*time.Hour))

	// No filter -> all 3, newest first.
	all := getActivityList(t, server, "/tool/activity")
	assert.Equal(t, 3, all.Total)
	require.Len(t, all.Items, 3)
	assert.Equal(t, fs.ActivityActionDelete, all.Items[0].Action, "ordered by -created_at")
	assert.Equal(t, fs.ActivityActionCreate, all.Items[2].Action)

	// Filter by schema.
	roles := getActivityList(t, server, "/tool/activity?schema=role")
	assert.Equal(t, 2, roles.Total)
	for _, it := range roles.Items {
		assert.Equal(t, "role", it.SchemaName)
	}

	// Filter by action.
	creates := getActivityList(t, server, "/tool/activity?action=create")
	assert.Equal(t, 1, creates.Total)

	// Pagination.
	page1 := getActivityList(t, server, "/tool/activity?limit=1&page=1")
	assert.Equal(t, 3, page1.Total)
	assert.Equal(t, 1, page1.PerPage)
	assert.Equal(t, 3, page1.LastPage)
	require.Len(t, page1.Items, 1)
	assert.Equal(t, fs.ActivityActionDelete, page1.Items[0].Action)

	page2 := getActivityList(t, server, "/tool/activity?limit=1&page=2")
	require.Len(t, page2.Items, 1)
	assert.Equal(t, fs.ActivityActionUpdate, page2.Items[0].Action)
}

func TestToolActivityEmptyResult(t *testing.T) {
	server, _ := newActivityTestServer(t)

	list := getActivityList(t, server, "/tool/activity?schema=nonexistent")
	assert.Equal(t, 0, list.Total)
	assert.Equal(t, 0, list.LastPage)
	assert.Empty(t, list.Items)
}

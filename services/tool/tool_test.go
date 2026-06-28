package toolservice_test

import (
	"net/http/httptest"
	"os"
	"testing"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/logger"
	"github.com/fastschema/fastschema/pkg/entdbadapter"
	"github.com/fastschema/fastschema/pkg/restfulresolver"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
	toolservice "github.com/fastschema/fastschema/services/tool"
	"github.com/stretchr/testify/assert"
)

type testApp struct {
	sb *schema.Builder
	db db.Client
}

func (s testApp) DB() db.Client {
	return s.db
}

func TestToolServiceError(t *testing.T) {
	sb := &schema.Builder{}
	db := utils.Must(entdbadapter.NewTestClient(utils.Must(os.MkdirTemp("", "migrations")), sb))
	toolService := toolservice.New(&testApp{sb: sb, db: db})

	resources := fs.NewResourcesManager()
	resources.Group("tool").
		Add(fs.NewResource("stats", toolService.Stats, &fs.Meta{
			Get: "/stats",
		}))

	assert.NoError(t, resources.Init())
	restResolver := restfulresolver.NewRestfulResolver(&restfulresolver.ResolverConfig{
		ResourceManager: resources,
		Logger:          logger.CreateMockLogger(true),
	})
	server := restResolver.Server()

	req := httptest.NewRequest("GET", "/tool/stats", nil)
	resp := utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 500, resp.StatusCode)
	response := utils.Must(utils.ReadCloserToString(resp.Body))
	assert.Contains(t, response, `model User not found`)
}

func TestToolService(t *testing.T) {
	sb := utils.Must(schema.NewBuilderFromDir(t.TempDir(), fs.SystemSchemaTypes...))
	db := utils.Must(entdbadapter.NewTestClient(utils.Must(os.MkdirTemp("", "migrations")), sb))
	toolService := toolservice.New(&testApp{sb: sb, db: db})

	resources := fs.NewResourcesManager()
	resources.Group("tool").
		Add(fs.NewResource("stats", toolService.Stats, &fs.Meta{
			Get: "/stats",
		})).
		Add(fs.NewResource("recent", toolService.Recent, &fs.Meta{
			Get: "/recent",
		}))

	assert.NoError(t, resources.Init())
	restResolver := restfulresolver.NewRestfulResolver(&restfulresolver.ResolverConfig{
		ResourceManager: resources,
		Logger:          logger.CreateMockLogger(true),
	})
	server := restResolver.Server()

	req := httptest.NewRequest("GET", "/tool/stats", nil)
	resp := utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 200, resp.StatusCode)
	response := utils.Must(utils.ReadCloserToString(resp.Body))
	assert.Contains(t, response, `"totalSchemas":8`)
	assert.Contains(t, response, `"totalUsers":0`)
	assert.Contains(t, response, `"totalFiles":0`)
	// Only system schemas are registered here, so content stats are empty:
	// this confirms system/junction schemas are excluded from contentCounts.
	assert.Contains(t, response, `"totalContent":0`)
	assert.Contains(t, response, `"contentCounts":[]`)

	// Recent has no content schemas to draw from here, so it returns an empty list.
	recentReq := httptest.NewRequest("GET", "/tool/recent", nil)
	recentResp := utils.Must(server.Test(recentReq))
	defer func() { assert.NoError(t, recentResp.Body.Close()) }()
	assert.Equal(t, 200, recentResp.StatusCode)
	recentBody := utils.Must(utils.ReadCloserToString(recentResp.Body))
	assert.Contains(t, recentBody, `"data":[]`)

	api := fs.NewResourcesManager().Group("api")
	toolService.CreateResource(api)
	assert.NotNil(t, api.Find("api.tool.stats"))
	assert.NotNil(t, api.Find("api.tool.recent"))
}

// Category is defined as a Go struct, so the schema builder marks it
// IsSystemSchema=true even though it is user content. Stats must still count it.
type Category struct {
	Name string `json:"name" fs:"label_field;optional;sortable;filterable;size=255"`
}

func TestToolServiceCountsCodeDefinedSchemas(t *testing.T) {
	schemaTypes := append([]any{}, fs.SystemSchemaTypes...)
	schemaTypes = append(schemaTypes, Category{})

	sb := utils.Must(schema.NewBuilderFromDir(t.TempDir(), schemaTypes...))
	db := utils.Must(entdbadapter.NewTestClient(utils.Must(os.MkdirTemp("", "migrations")), sb))
	toolService := toolservice.New(&testApp{sb: sb, db: db})

	resources := fs.NewResourcesManager()
	resources.Group("tool").
		Add(fs.NewResource("stats", toolService.Stats, &fs.Meta{Get: "/stats"}))
	assert.NoError(t, resources.Init())
	restResolver := restfulresolver.NewRestfulResolver(&restfulresolver.ResolverConfig{
		ResourceManager: resources,
		Logger:          logger.CreateMockLogger(true),
	})
	server := restResolver.Server()

	req := httptest.NewRequest("GET", "/tool/stats", nil)
	resp := utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 200, resp.StatusCode)
	response := utils.Must(utils.ReadCloserToString(resp.Body))
	// The code-defined Category schema must appear in contentCounts despite
	// carrying IsSystemSchema=true.
	assert.Contains(t, response, `"schema":"category"`)
}

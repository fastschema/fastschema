package toolservice_test

import (
	"net/http/httptest"
	"testing"

	"github.com/fastschema/fastschema/app"
	"github.com/fastschema/fastschema/pkg/entdbadapter"
	"github.com/fastschema/fastschema/pkg/restresolver"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
	toolservice "github.com/fastschema/fastschema/services/tool"
	"github.com/stretchr/testify/assert"
)

type testApp struct {
	sb *schema.Builder
	db app.DBClient
}

func (s testApp) DB() app.DBClient {
	return s.db
}

func TestToolServiceError(t *testing.T) {
	sb := &schema.Builder{}
	db := utils.Must(entdbadapter.NewTestClient(t.TempDir(), sb))
	toolService := toolservice.New(&testApp{sb: sb, db: db})

	resources := app.NewResourcesManager()
	resources.Group("tool").
		Add(app.NewResource("stats", toolService.Stats, app.Meta{app.GET: "/stats"}))

	assert.NoError(t, resources.Init())
	restResolver := restresolver.NewRestResolver(resources).Init(app.CreateMockLogger(true))
	server := restResolver.Server()

	req := httptest.NewRequest("GET", "/tool/stats", nil)
	resp := utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 500, resp.StatusCode)
	response := utils.Must(utils.ReadCloserToString(resp.Body))
	assert.Contains(t, response, `model user not found`)
	assert.Contains(t, response, `model role not found`)
	assert.Contains(t, response, `model media not found`)
}

func TestToolService(t *testing.T) {
	sb := utils.Must(schema.NewBuilderFromDir(t.TempDir()))
	db := utils.Must(entdbadapter.NewTestClient(t.TempDir(), sb))
	toolService := toolservice.New(&testApp{sb: sb, db: db})

	resources := app.NewResourcesManager()
	resources.Group("tool").
		Add(app.NewResource("stats", toolService.Stats, app.Meta{app.GET: "/stats"}))

	assert.NoError(t, resources.Init())
	restResolver := restresolver.NewRestResolver(resources).Init(app.CreateMockLogger(true))
	server := restResolver.Server()

	req := httptest.NewRequest("GET", "/tool/stats", nil)
	resp := utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 200, resp.StatusCode)
	response := utils.Must(utils.ReadCloserToString(resp.Body))
	assert.Contains(t, response, `"totalSchemas":5`)
	assert.Contains(t, response, `"totalUsers":0`)
	assert.Contains(t, response, `"totalMedias":0`)
}

package toolservice_test

import (
	"net/http/httptest"
	"os"
	"testing"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/logger"
	"github.com/fastschema/fastschema/pkg/entdbadapter"
	"github.com/fastschema/fastschema/pkg/restresolver"
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
	restResolver := restresolver.NewRestResolver(resources, logger.CreateMockLogger(true))
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
		}))

	assert.NoError(t, resources.Init())
	restResolver := restresolver.NewRestResolver(resources, logger.CreateMockLogger(true))
	server := restResolver.Server()

	req := httptest.NewRequest("GET", "/tool/stats", nil)
	resp := utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 200, resp.StatusCode)
	response := utils.Must(utils.ReadCloserToString(resp.Body))
	assert.Contains(t, response, `"totalSchemas":5`)
	assert.Contains(t, response, `"totalUsers":0`)
	assert.Contains(t, response, `"totalFiles":0`)
}

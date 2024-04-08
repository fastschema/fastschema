package contentservice_test

import (
	"net/http"
	"testing"

	"github.com/fastschema/fastschema/app"
	"github.com/fastschema/fastschema/pkg/entdbadapter"
	"github.com/fastschema/fastschema/pkg/restresolver"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
	contentservice "github.com/fastschema/fastschema/services/content"
	"github.com/stretchr/testify/assert"
)

type contentService struct {
	t             *testing.T
	schemaBuilder *schema.Builder
}

func (s contentService) DB() app.DBClient {
	migrateDir := s.t.TempDir()
	dbClient, err := entdbadapter.NewClient(&app.DBConfig{
		Driver:       "sqlite",
		Name:         ":memory:",
		MigrationDir: migrateDir,
	}, s.schemaBuilder)
	assert.NoError(s.t, err)
	return dbClient
}

func createContentService(t *testing.T) (
	*contentservice.ContentService,
	*restresolver.Server,
) {
	schemaDir := t.TempDir()
	utils.WriteFile(schemaDir+"/post.json", `{
		"name": "blog",
		"namespace": "blogs",
		"label_field": "name",
		"fields": [
			{
				"type": "string",
				"name": "name",
				"label": "Name",
				"sortable": true
			}
		]
	}`)
	schemaBuilder, err := schema.NewBuilderFromDir(schemaDir)
	assert.NoError(t, err)
	config := &contentService{t: t, schemaBuilder: schemaBuilder}
	contentService := contentservice.New(config)
	resources := app.NewResourcesManager()
	resources.Group("content").
		Add(app.NewResource("list", contentService.List, app.Meta{app.GET: "/:schema"})).
		Add(app.NewResource("detail", contentService.Detail, app.Meta{app.GET: "/:schema/:id"})).
		Add(app.NewResource("create", contentService.Create, app.Meta{app.POST: "/:schema"})).
		Add(app.NewResource("update", contentService.Update, app.Meta{app.PUT: "/:schema/:id"})).
		Add(app.NewResource("delete", contentService.Delete, app.Meta{app.DELETE: "/:schema/:id"}))
	err = resources.Init()
	assert.NoError(t, err)
	restResolver := restresolver.NewRestResolver(resources)
	restResolver.Init(app.CreateMockLogger())
	server := restResolver.Server()

	return contentService, server
}

func closeResponse(t *testing.T, resp *http.Response) {
	err := resp.Body.Close()
	assert.NoError(t, err)
}

func TestNewContentService(t *testing.T) {
	service, server := createContentService(t)
	assert.NotNil(t, service)
	assert.NotNil(t, server)
}

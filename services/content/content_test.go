package contentservice_test

import (
	"os"
	"testing"

	"github.com/fastschema/fastschema/app"
	"github.com/fastschema/fastschema/pkg/entdbadapter"
	rr "github.com/fastschema/fastschema/pkg/restresolver"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
	cs "github.com/fastschema/fastschema/services/content"
	"github.com/stretchr/testify/assert"
)

type testApp struct {
	sb *schema.Builder
	db app.DBClient
}

func (s testApp) DB() app.DBClient {
	return s.db
}

func createContentService(t *testing.T) (*cs.ContentService, *rr.Server) {
	schemaDir := t.TempDir()
	utils.WriteFile(schemaDir+"/blog.json", `{
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
	sb := utils.Must(schema.NewBuilderFromDir(schemaDir))
	db := utils.Must(entdbadapter.NewTestClient(utils.Must(os.MkdirTemp("", "migrations")), sb))
	contentService := cs.New(&testApp{sb: sb, db: db})
	resources := app.NewResourcesManager()
	resources.Group("content").
		Add(app.NewResource("list", contentService.List, app.Meta{app.GET: "/:schema"})).
		Add(app.NewResource("detail", contentService.Detail, app.Meta{app.GET: "/:schema/:id"})).
		Add(app.NewResource("create", contentService.Create, app.Meta{app.POST: "/:schema"})).
		Add(app.NewResource("update", contentService.Update, app.Meta{app.PUT: "/:schema/:id"})).
		Add(app.NewResource("delete", contentService.Delete, app.Meta{app.DELETE: "/:schema/:id"}))

	assert.NoError(t, resources.Init())
	restResolver := rr.NewRestResolver(resources, app.CreateMockLogger(true))

	return contentService, restResolver.Server()
}

func TestNewContentService(t *testing.T) {
	service, server := createContentService(t)
	assert.NotNil(t, service)
	assert.NotNil(t, server)
}

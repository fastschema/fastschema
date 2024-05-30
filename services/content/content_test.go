package contentservice_test

import (
	"os"
	"testing"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/logger"
	"github.com/fastschema/fastschema/pkg/entdbadapter"
	rr "github.com/fastschema/fastschema/pkg/restfulresolver"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
	cs "github.com/fastschema/fastschema/services/content"
	"github.com/stretchr/testify/assert"
)

type testApp struct {
	sb *schema.Builder
	db db.Client
}

func (s testApp) DB() db.Client {
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
	sb := utils.Must(schema.NewBuilderFromDir(schemaDir, fs.SystemSchemaTypes...))
	db := utils.Must(entdbadapter.NewTestClient(utils.Must(os.MkdirTemp("", "migrations")), sb))
	contentService := cs.New(&testApp{sb: sb, db: db})
	resources := fs.NewResourcesManager()
	resources.Group("content").
		Add(fs.NewResource("list", contentService.List, &fs.Meta{
			Get: "/:schema",
		})).
		Add(fs.NewResource("detail", contentService.Detail, &fs.Meta{
			Get: "/:schema/:id",
		})).
		Add(fs.NewResource("create", contentService.Create, &fs.Meta{
			Post: "/:schema",
		})).
		Add(fs.NewResource("update", contentService.Update, &fs.Meta{
			Put: "/:schema/:id",
		})).
		Add(fs.NewResource("delete", contentService.Delete, &fs.Meta{
			Delete: "/:schema/:id",
		}))

	assert.NoError(t, resources.Init())
	restResolver := rr.NewRestfulResolver(resources, logger.CreateMockLogger(true))

	return contentService, restResolver.Server()
}

func TestNewContentService(t *testing.T) {
	service, server := createContentService(t)
	assert.NotNil(t, service)
	assert.NotNil(t, server)
}

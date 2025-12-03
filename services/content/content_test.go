package contentservice_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/logger"
	"github.com/fastschema/fastschema/pkg/entdbadapter"
	rr "github.com/fastschema/fastschema/pkg/restfulresolver"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
	cs "github.com/fastschema/fastschema/services/content"
)

type testApp struct {
	sb        *schema.Builder
	db        db.Client
	resources *fs.ResourcesManager
}

func (s testApp) DB() db.Client {
	return s.db
}

func createContentService(t *testing.T) (*cs.ContentService, *rr.Server) {
	schemaDir := t.TempDir()
	utils.WriteFile(schemaDir+"/blog.yaml", `name: blog
namespace: blogs
label_field: name
fields:
  - name: name
    label: Name
    type: string
    sortable: true
  - name: tags
    label: Tags
    type: relation
    optional: true
    sortable: true
    relation:
      schema: tag
      field: blogs
      type: o2m
`)
	utils.WriteFile(schemaDir+"/tag.yaml", `name: tag
namespace: tags
label_field: name
fields:
  - name: name
    label: Name
    type: string
    sortable: true
  - name: blogs
    label: Blogs
    type: relation
    relation:
      schema: blog
      field: tags
      type: o2m
      owner: true
`)
	sb := utils.Must(schema.NewBuilderFromDir(schemaDir, fs.SystemSchemaTypes...))
	db := utils.Must(entdbadapter.NewTestClient(utils.Must(os.MkdirTemp("", "migrations")), sb))
	testApp := &testApp{sb: sb, db: db}
	contentService := cs.New(testApp)
	testApp.resources = fs.NewResourcesManager()
	testApp.resources.Group("content").
		Add(fs.NewResource("list", contentService.List, &fs.Meta{
			Get: "/:schema",
		})).
		Add(fs.NewResource("detail", contentService.Detail, &fs.Meta{
			Get: "/:schema/:id",
		})).
		Add(fs.NewResource("create", contentService.Create, &fs.Meta{
			Post: "/:schema",
		})).
		Add(fs.NewResource("bulk-update", contentService.BulkUpdate, &fs.Meta{
			Put: "/:schema/update",
		})).
		Add(fs.NewResource("update", contentService.Update, &fs.Meta{
			Put: "/:schema/:id",
		})).
		Add(fs.NewResource("bulk-delete", contentService.BulkDelete, &fs.Meta{
			Delete: "/:schema/delete",
		})).
		Add(fs.NewResource("delete", contentService.Delete, &fs.Meta{
			Delete: "/:schema/:id",
		}))

	assert.NoError(t, testApp.resources.Init())
	restResolver := rr.NewRestfulResolver(&rr.ResolverConfig{
		ResourceManager: testApp.resources,
		Logger:          logger.CreateMockLogger(true),
	})

	return contentService, restResolver.Server()
}

func TestNewContentService(t *testing.T) {
	service, server := createContentService(t)
	assert.NotNil(t, service)
	assert.NotNil(t, server)
}

func TestCreateResource(t *testing.T) {
	api := fs.NewResourcesManager().Group("api")
	service, _ := createContentService(t)

	service.CreateResource(api)
	assert.NotNil(t, api.Find("api.content.list"))
	assert.NotNil(t, api.Find("api.content.detail"))
	assert.NotNil(t, api.Find("api.content.create"))
	assert.NotNil(t, api.Find("api.content.bulk-update"))
	assert.NotNil(t, api.Find("api.content.update"))
	assert.NotNil(t, api.Find("api.content.bulk-delete"))
	assert.NotNil(t, api.Find("api.content.delete"))
}

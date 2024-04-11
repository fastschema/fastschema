package schemaservice_test

import (
	"os"
	"testing"

	"github.com/fastschema/fastschema/app"
	"github.com/fastschema/fastschema/pkg/entdbadapter"
	"github.com/fastschema/fastschema/pkg/restresolver"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
	schemaservice "github.com/fastschema/fastschema/services/schema"
	"github.com/stretchr/testify/assert"
)

var (
	testCategoryJSON = `{
		"name": "category",
		"namespace": "categories",
		"label_field": "name",
		"fields": [
			{
				"type": "string",
				"name": "name",
				"label": "Name",
				"unique": true,
				"sortable": true
			}
		]
	}`
	testBlogJSON = `{
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
	}`
	testTagJSON = `{
		"name": "tag",
		"namespace": "tags",
		"label_field": "name",
		"fields": [
			{
				"type": "string",
				"name": "name",
				"label": "Name",
				"unique": true,
				"sortable": true
			}
		]
	}`
)

type testApp struct {
	sb        *schema.Builder
	db        app.DBClient
	schemaDir string
}

func (s *testApp) Schema(name string) *schema.Schema {
	return utils.Must(s.sb.Schema(name))
}

func (s *testApp) DB() app.DBClient {
	return s.db
}

func (s *testApp) SchemaBuilder() *schema.Builder {
	return s.sb
}

func (s *testApp) Reload(migration *app.Migration) error {
	s.sb = utils.Must(schema.NewBuilderFromDir(s.schemaDir))
	s.db = utils.Must(entdbadapter.NewTestClient(os.TempDir(), s.sb))

	return nil
}

type testSchemaSeviceConfig struct {
	extraSchemas map[string]string
	schemaDir    string
}

func createSchemaService(t *testing.T, config *testSchemaSeviceConfig) (
	*testApp,
	*schemaservice.SchemaService,
	*restresolver.Server,
) {
	schemaDir := t.TempDir()
	schemas := map[string]string{
		"category": testCategoryJSON,
	}

	if config != nil {
		if config.schemaDir != "" {
			schemaDir = config.schemaDir

			// remove all json file in schemaDir
			assert.NoError(t, os.RemoveAll(schemaDir))
			assert.NoError(t, os.MkdirAll(schemaDir, 0755))
		}

		for k, v := range config.extraSchemas {
			schemas[k] = v
		}
	}

	for name, content := range schemas {
		utils.WriteFile(schemaDir+"/"+name+".json", content)
	}
	sb := utils.Must(schema.NewBuilderFromDir(schemaDir))
	db := utils.Must(entdbadapter.NewTestClient(t.TempDir(), sb))
	testApp := &testApp{sb: sb, db: db, schemaDir: schemaDir}
	schemaService := schemaservice.New(testApp)

	resources := app.NewResourcesManager()
	resources.Group("schema").
		Add(app.NewResource("list", schemaService.List, app.Meta{app.GET: ""})).
		Add(app.NewResource("create", schemaService.Create, app.Meta{app.POST: ""})).
		Add(app.NewResource("detail", schemaService.Detail, app.Meta{app.GET: "/:name"})).
		Add(app.NewResource("update", schemaService.Update, app.Meta{app.PUT: "/:name"})).
		Add(app.NewResource("delete", schemaService.Delete, app.Meta{app.DELETE: "/:name"}))

	assert.NoError(t, resources.Init())
	restResolver := restresolver.NewRestResolver(resources).Init(app.CreateMockLogger(true))

	return testApp, schemaService, restResolver.Server()
}

func TestSchemaService(t *testing.T) {
	testApp, schemaService, server := createSchemaService(t, nil)
	assert.NotNil(t, testApp)
	assert.NotNil(t, schemaService)
	assert.NotNil(t, server)
}

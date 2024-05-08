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
			},
			{
				"type": "relation",
				"name": "pair",
				"label": "Pair",
				"optional": true,
				"relation": {
					"schema": "blog",
					"field": "pair",
					"type": "o2o",
					"owner": true,
					"optional": true
				}
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

	testBlogJSONFields = map[string]string{
		"description": `{"type": "string","name": "description","label": "Description","sortable": true}`,
		"categories": `{
			"type": "relation",
			"name": "categories",
			"label": "Categories",
			"relation": {
				"schema": "category",
				"field": "blogs",
				"type": "m2m",
				"owner": false,
				"optional": false
			}
		}`,
		"category": `{
			"type": "relation",
			"name": "category",
			"label": "Category",
			"relation": {
				"schema": "category",
				"field": "blogs",
				"type": "o2m",
				"owner": false,
				"optional": false
			}
		}`,
		"note": `{"type": "string","name": "note","label": "Note","sortable": true}`,
		"tags": `{
			"type": "relation",
			"name": "tags",
			"label": "Tags",
			"relation": {
				"schema": "tag",
				"field": "blogs",
				"type": "m2m",
				"owner": false,
				"optional": false
			}
		}`,
	}
)

type testApp struct {
	sb        *schema.Builder
	db        app.DBClient
	schemaDir string
	reloadFn  func(*app.Migration) error
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
	s.db = utils.Must(entdbadapter.NewTestClient(utils.Must(os.MkdirTemp("", "migrations")), s.sb))

	if s.reloadFn != nil {
		return s.reloadFn(migration)
	}

	return nil
}

type testSchemaSeviceConfig struct {
	extraSchemas map[string]string
	schemaDir    string
	reloadFn     func(*app.Migration) error
}

func createSchemaService(t *testing.T, config *testSchemaSeviceConfig) (
	*testApp,
	*schemaservice.SchemaService,
	*restresolver.Server,
) {
	var reloadFn func(*app.Migration) error
	schemaDir := t.TempDir()
	schemas := map[string]string{
		"category": testCategoryJSON,
	}

	if config != nil {
		if config.schemaDir != "" {
			schemaDir = config.schemaDir

			// remove all files in schemaDir
			assert.NoError(t, os.RemoveAll(schemaDir))
			assert.NoError(t, os.MkdirAll(schemaDir, 0755))
		}

		for k, v := range config.extraSchemas {
			schemas[k] = v
		}

		reloadFn = config.reloadFn
	}

	for name, content := range schemas {
		utils.WriteFile(schemaDir+"/"+name+".json", content)
	}

	migrationDir := utils.Must(os.MkdirTemp("", "migrations"))
	sb := utils.Must(schema.NewBuilderFromDir(schemaDir))
	db := utils.Must(entdbadapter.NewTestClient(migrationDir, sb))
	testApp := &testApp{sb: sb, db: db, schemaDir: schemaDir, reloadFn: reloadFn}
	schemaService := schemaservice.New(testApp)

	resources := app.NewResourcesManager()
	resources.Group("schema").
		Add(app.NewResource("list", schemaService.List, &app.Meta{
			Get: "/",
		})).
		Add(app.NewResource("create", schemaService.Create, &app.Meta{
			Post: "/",
		})).
		Add(app.NewResource("detail", schemaService.Detail, &app.Meta{
			Get: "/:name",
		})).
		Add(app.NewResource("update", schemaService.Update, &app.Meta{
			Put: "/:name",
		})).
		Add(app.NewResource("delete", schemaService.Delete, &app.Meta{
			Delete: "/:name",
		}))

	assert.NoError(t, resources.Init())
	restResolver := restresolver.NewRestResolver(resources, app.CreateMockLogger(true))

	return testApp, schemaService, restResolver.Server()
}

func TestSchemaService(t *testing.T) {
	testApp, schemaService, server := createSchemaService(t, nil)
	assert.NotNil(t, testApp)
	assert.NotNil(t, schemaService)
	assert.NotNil(t, server)
}

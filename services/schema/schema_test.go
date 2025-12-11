package schemaservice_test

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/logger"
	"github.com/fastschema/fastschema/pkg/entdbadapter"
	"github.com/fastschema/fastschema/pkg/rclonefs"
	"github.com/fastschema/fastschema/pkg/restfulresolver"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
	schemaservice "github.com/fastschema/fastschema/services/schema"
)

var (
	testCategoryYAML = `name: category
namespace: categories
label_field: name
fields:
- name: name
  label: Name
  type: string
  unique: true
  sortable: true`

	testCategoryYAMLToImport = `name: category_import
namespace: categories_import
label_field: name
fields:
- name: name
  label: Name
  type: string
  unique: true
  sortable: true`
	testBlogYAML = `name: blog
namespace: blogs
label_field: name
fields:
- name: name
  label: Name
  type: string
  sortable: true
- name: pair
  label: Pair
  type: relation
  optional: true
  relation:
    type: o2o
    schema: blog
    field: pair
    owner: true
    optional: true`
	testTagYAML = `name: tag
namespace: tags
label_field: name
fields:
- name: name
  label: Name
  type: string
  unique: true
  sortable: true`

	testBlogYAMLFields = map[string]string{
		"description": `- name: description
  label: Description
  type: string
  sortable: true`,
		"categories": `- name: categories
  label: Categories
  type: relation
  relation:
    type: m2m
    schema: category
    field: blogs
    owner: false
    optional: false`,
		"category": `- name: category
  label: Category
  type: relation
  relation:
    type: o2m
    schema: category
    field: blogs
    owner: false
    optional: false`,
		"note": `- name: note
  label: Note
  type: string
  sortable: true`,
		"tags": `- name: tags
  label: Tags
  type: relation
  relation:
    type: m2m
    schema: tag
    field: blogs
    owner: false
    optional: false`,
	}
)

// schemaToJSON converts a schema object to a JSON string for API requests
func schemaToJSON(s *schema.Schema) string {
	jsonBytes, err := json.Marshal(s)
	if err != nil {
		panic(err)
	}
	return string(jsonBytes)
}

// yamlSchemaToJSON converts a YAML schema string to a JSON schema string for API requests
func yamlSchemaToJSON(yamlData string) string {
	s, err := schema.NewSchemaFromYAML(yamlData)
	if err != nil {
		panic(err)
	}
	return schemaToJSON(s)
}

// modifySchema parses JSON schema, applies a modification function, and returns JSON
func modifySchema(schemaJSON string, modifier func(*schema.Schema)) string {
	var s schema.Schema
	if err := json.Unmarshal([]byte(schemaJSON), &s); err != nil {
		panic(err)
	}
	modifier(&s)
	return schemaToJSON(&s)
}

type testApp struct {
	sb        *schema.Builder
	db        db.Client
	disks     []fs.Disk
	schemaDir string
	reloadFn  func(*db.Migration) error
}

func (s *testApp) Schema(name string) *schema.Schema {
	return utils.Must(s.sb.Schema(name))
}

func (s *testApp) DB() db.Client {
	return s.db
}

func (s *testApp) SchemaBuilder() *schema.Builder {
	return s.sb
}

func (s *testApp) SystemSchemas() []any {
	return fs.SystemSchemaTypes
}

func (s testApp) Disk(names ...string) fs.Disk {
	return s.disks[0]
}

func (s *testApp) Reload(ctx context.Context, migration *db.Migration) error {
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
	reloadFn     func(*db.Migration) error
}

func createSchemaService(t *testing.T, config *testSchemaSeviceConfig) (
	*testApp,
	*schemaservice.SchemaService,
	*restfulresolver.Server,
) {
	var reloadFn func(*db.Migration) error
	schemaDir := t.TempDir()
	schemas := map[string]string{
		"category": testCategoryYAML,
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
		utils.WriteFile(schemaDir+"/"+name+".yaml", content)
	}

	migrationDir := utils.Must(os.MkdirTemp("", "migrations"))
	sb := utils.Must(schema.NewBuilderFromDir(schemaDir, fs.SystemSchemaTypes...))
	db := utils.Must(entdbadapter.NewTestClient(migrationDir, sb))
	disks := utils.Must(rclonefs.NewFromConfig([]*fs.DiskConfig{{
		Name:    "local_test",
		Driver:  "local",
		Root:    t.TempDir(),
		BaseURL: "http://localhost:3000/files",
	}}, t.TempDir()))
	testApp := &testApp{sb: sb, db: db, schemaDir: schemaDir, disks: disks, reloadFn: reloadFn}
	schemaService := schemaservice.New(testApp)

	resources := fs.NewResourcesManager()
	resources.Group("schema").
		Add(fs.NewResource("list", schemaService.List, &fs.Meta{
			Get: "/",
		})).
		Add(fs.NewResource("create", schemaService.Create, &fs.Meta{
			Post: "/",
		})).
		Add(fs.NewResource("detail", schemaService.Detail, &fs.Meta{
			Get: "/:name",
		})).
		Add(fs.NewResource("update", schemaService.Update, &fs.Meta{
			Put: "/:name",
		})).
		Add(fs.NewResource("delete", schemaService.Delete, &fs.Meta{
			Delete: "/:name",
		})).
		Add(fs.NewResource("import", schemaService.Import, &fs.Meta{
			Post: "/import",
		})).
		Add(fs.NewResource("export", schemaService.Export, &fs.Meta{
			Post: "/export",
		}))

	assert.NoError(t, resources.Init())
	restResolver := restfulresolver.NewRestfulResolver(&restfulresolver.ResolverConfig{
		ResourceManager: resources,
		Logger:          logger.CreateMockLogger(true),
	})

	return testApp, schemaService, restResolver.Server()
}

func TestSchemaService(t *testing.T) {
	testApp, schemaService, server := createSchemaService(t, nil)
	assert.NotNil(t, testApp)
	assert.NotNil(t, schemaService)
	assert.NotNil(t, server)
}

func TestCreateResource(t *testing.T) {
	_, schemaService, _ := createSchemaService(t, nil)
	api := fs.NewResourcesManager().Group("api")
	schemaService.CreateResource(api)
	assert.NotNil(t, api.Find("api.schema.list"))
	assert.NotNil(t, api.Find("api.schema.create"))
	assert.NotNil(t, api.Find("api.schema.detail"))
	assert.NotNil(t, api.Find("api.schema.update"))
	assert.NotNil(t, api.Find("api.schema.delete"))
	assert.NotNil(t, api.Find("api.schema.import"))
	assert.NotNil(t, api.Find("api.schema.export"))
}

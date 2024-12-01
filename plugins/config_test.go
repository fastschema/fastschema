package plugins_test

import (
	"os"
	"testing"

	"github.com/fastschema/fastschema"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/plugins"
	"github.com/stretchr/testify/assert"
)

const pluginContent = `
const schemaProduct = {
  "name": "product",
  "namespace": "products",
  "label_field": "name",
  "disable_timestamp": false,
  "fields": [
    {
      "type": "string",
      "name": "name",
      "label": "Name",
      "optional": true,
      "sortable": true
    },
    {
      "type": "string",
      "name": "description",
      "label": "Description",
      "optional": true,
      "sortable": true,
      "renderer": {
        "class": "textarea"
      }
    }
  ]
}

const testHook = () => {
}

const Config = config => {
	config.AddSchemas(schemaProduct);
	config.port = 9000;

	config.OnPreResolve(testHook);
  config.OnPostResolve(testHook);

  config.OnPreDBQuery(testHook);
  config.OnPostDBQuery(testHook);

  config.OnPreDBExec(testHook);
  config.OnPostDBExec(testHook);

  config.OnPreDBCreate(testHook);
  config.OnPostDBCreate(testHook);

  config.OnPreDBUpdate(testHook);
  config.OnPostDBUpdate(testHook);

  config.OnPreDBDelete(testHook);
  config.OnPostDBDelete(testHook);
}
`

func TestConfig(t *testing.T) {
	// Create application
	app, err := fastschema.New(&fs.Config{
		Dir: utils.Must(os.MkdirTemp("", "fastschema")),
	})
	assert.NoError(t, err)
	assert.NotNil(t, app)

	// Create program
	gojaProgram, _, err := plugins.CreateGoJaProgram("", []byte(pluginContent))
	assert.NoError(t, err)
	program := plugins.NewProgram(gojaProgram, "plugin.test")

	// Create config actions
	configActions := plugins.NewConfigActions(app, program, nil)
	assert.NotNil(t, configActions)

	// Call Config function
	_, err = program.CallFunc("Config", nil, plugins.NewConfigActions(
		app,
		program,
		nil,
	))
	assert.NoError(t, err)

	// Assert created system schemas
	assert.Len(t, app.Config().SystemSchemas, 1)

	// Assert updated app port
	assert.Equal(t, "9000", app.Config().Port)

	// Assert created pre resolve hooks

	// Pre resolve hooks should has 2 hooks, including the default one: authorize
	assert.Len(t, app.Hooks().PreResolve, 2)

	// Post resolve hooks
	assert.Len(t, app.Hooks().PostResolve, 1)

	// Pre DB Query hooks
	assert.Len(t, app.Hooks().DBHooks.PreDBQuery, 1)

	// Post DB Query hooks: should has 2 hooks, including the default one: FileListHook
	assert.Len(t, app.Hooks().DBHooks.PostDBQuery, 2)

	// Pre DB Exec hooks
	assert.Len(t, app.Hooks().DBHooks.PreDBExec, 1)

	// Post DB Exec hooks
	assert.Len(t, app.Hooks().DBHooks.PostDBExec, 1)

	// Pre DB Create hooks
	assert.Len(t, app.Hooks().DBHooks.PreDBCreate, 1)

	// Post DB Create hooks: should has 2 hooks, including the default one: realtime.ContentCreateHook
	assert.Len(t, app.Hooks().DBHooks.PostDBCreate, 2)

	// Pre DB Update hooks
	assert.Len(t, app.Hooks().DBHooks.PreDBUpdate, 1)

	// Post DB Update hooks: should has 2 hooks, including the default one: realtime.ContentUpdateHook
	assert.Len(t, app.Hooks().DBHooks.PostDBUpdate, 2)

	// Pre DB Delete hooks
	assert.Len(t, app.Hooks().DBHooks.PreDBDelete, 1)

	// Post DB Delete hooks: should has 2 hooks, including the default one: realtime.ContentDeleteHook
	assert.Len(t, app.Hooks().DBHooks.PostDBDelete, 2)
}

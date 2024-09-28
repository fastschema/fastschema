package plugins_test

import (
	"testing"

	"github.com/dop251/goja"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/plugins"
	"github.com/stretchr/testify/assert"
)

func TestNewResource(t *testing.T) {
	gojaProgram, _, err := plugins.CreateGoJaProgram("", []byte(`function Ping() {}`))
	assert.Nil(t, err)
	assert.NotNil(t, gojaProgram)

	mockFsResource := &fs.Resource{}
	mockProgram := plugins.NewProgram(gojaProgram, "plugin.test")
	mockSet := map[string]any{"key": "value"}

	resource := plugins.NewResource(mockFsResource, mockProgram, mockSet)
	assert.NotNil(t, resource)

	apiResource := resource.Group("api", &fs.Meta{
		Prefix: "api",
	})
	assert.NotNil(t, apiResource)

	vm := goja.New()
	_, err = vm.RunProgram(gojaProgram)
	assert.Nil(t, err)
	funcObj := vm.Get("Ping")

	apiResource, err = apiResource.Add(funcObj)
	assert.Nil(t, err)
	assert.NotNil(t, apiResource)
}

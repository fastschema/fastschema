package plugins_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dop251/goja"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/plugins"
	"github.com/stretchr/testify/assert"
)

func TestProgram(t *testing.T) {
	jsContent := []byte(`
		const a = 1;
		const b = {};
		const c = {name: 222};
		const d = {name: ""};
		const e = {name: "+"};
		const f = {name: "b"};
		const test = () => {
			return "test";
		}

		const testThrow = () => {
			throw new Error("test error");
		}

		const testPromiseOk = async() => {
		}

		const testPromiseError = async() => {
			throw new Error("test error");
		}
	`)

	// Create program
	gojaProgram, _, err := plugins.CreateGoJaProgram("", jsContent)
	assert.NoError(t, err)
	assert.NotNil(t, gojaProgram)

	vm := goja.New()
	_, err = vm.RunProgram(gojaProgram)
	assert.NoError(t, err)

	program := plugins.NewProgram(gojaProgram, "plugin.test")

	// Test VerifyJsFunc
	t.Run("VerifyJsFunc", func(t *testing.T) {
		// Invalid object value
		_, err = program.VerifyJsFunc(vm.Get("a"))
		assert.Error(t, err)

		// Object doesn't have name property
		_, err = program.VerifyJsFunc(vm.Get("b"))
		assert.Error(t, err)

		// Name property is not a string
		_, err = program.VerifyJsFunc(vm.Get("c"))
		assert.Error(t, err)

		// Name property is empty
		_, err = program.VerifyJsFunc(vm.Get("d"))
		assert.Error(t, err)

		// Name property is not a valid function name
		_, err = program.VerifyJsFunc(vm.Get("e"))
		assert.Error(t, err)

		// Valid function object but not a function
		_, err = program.VerifyJsFunc(vm.Get("f"))
		assert.Error(t, err)

		// Valid function object but function is not found
		fnObject := vm.ToValue(map[string]any{
			"name": "invalidFunction",
		})
		_, err = program.VerifyJsFunc(fnObject)
		assert.Error(t, err)

		// Valid function object
		fnName, err := program.VerifyJsFunc(vm.Get("test"))
		assert.NoError(t, err)
		assert.Equal(t, "test", fnName)
	})

	t.Run("WithFuncName", func(t *testing.T) {
		// Invalid object value
		err := program.WithFuncName(vm.Get("a"), func(name string) {
			assert.Fail(t, "Should not be called")
		})
		assert.Error(t, err)

		err = program.WithFuncName(vm.Get("test"), func(name string) {
			assert.Equal(t, "test", name)
		})
		assert.NoError(t, err)
	})

	t.Run("CallFunc", func(t *testing.T) {
		// Invalid function name
		_, err := program.CallFunc("invalidFunction", nil)
		assert.Error(t, err)

		// Not a function
		_, err = program.CallFunc("f", nil)
		assert.Error(t, err)

		// Function throws error
		_, err = program.CallFunc("testThrow", nil)
		assert.Error(t, err)

		// Function returns promise ok
		_, err = program.CallFunc("testPromiseOk", nil)
		assert.NoError(t, err)

		// Function returns promise error
		_, err = program.CallFunc("testPromiseError", nil)
		assert.Error(t, err)

		// Valid function name
		result, err := program.CallFunc("test", nil)
		assert.NoError(t, err)
		assert.Equal(t, "test", result)
	})
}

func TestCreateGoJaProgram(t *testing.T) {
	t.Run("No file and no script", func(t *testing.T) {
		program, content, err := plugins.CreateGoJaProgram("", nil)
		assert.Nil(t, program)
		assert.Empty(t, content)
		assert.Error(t, err)
		assert.Equal(t, "createvm: file or script is required", err.Error())
	})

	t.Run("Valid script", func(t *testing.T) {
		script := []byte(`const test = () => "test";`)
		program, content, err := plugins.CreateGoJaProgram("", script)
		assert.NotNil(t, program)
		assert.Equal(t, string(script), content)
		assert.NoError(t, err)
	})

	t.Run("Invalid script", func(t *testing.T) {
		script := []byte(`const test = () => {`)
		program, content, err := plugins.CreateGoJaProgram("", script)
		assert.Nil(t, program)
		assert.Empty(t, content)
		assert.Error(t, err)
	})

	pluginsDir := utils.Must(os.MkdirTemp("", "plugins"))
	t.Run("Valid file", func(t *testing.T) {
		fileContent := `const test = () => "test";`
		fileName := filepath.Join(pluginsDir, "test.js")
		err := os.WriteFile(fileName, []byte(fileContent), 0644)
		assert.NoError(t, err)
		defer os.Remove(fileName)

		program, content, err := plugins.CreateGoJaProgram(fileName, nil)
		assert.NotNil(t, program)
		assert.Equal(t, fileContent, content)
		assert.NoError(t, err)
	})

	t.Run("Invalid file", func(t *testing.T) {
		fileName := filepath.Join(pluginsDir, "nonexistent.js")
		program, content, err := plugins.CreateGoJaProgram(fileName, nil)
		assert.Nil(t, program)
		assert.Empty(t, content)
		assert.Error(t, err)
	})
}

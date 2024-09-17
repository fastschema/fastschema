package plugins_test

import (
	"strings"
	"testing"

	"github.com/fastschema/fastschema/plugins"
	"github.com/stretchr/testify/assert"
)

func TestPoolError(t *testing.T) {
	defer func() {
		r := recover()
		assert.Error(t, r.(error))
	}()

	jsContent := []byte(`
		console.error("error running program");
		throw new Error("error running program");
	`)

	gojaProgram, _, err := plugins.CreateGoJaProgram("", jsContent)
	assert.Nil(t, err)

	program := plugins.NewProgram(gojaProgram, "test")
	props := plugins.CreateVMProps(
		"add",
		program,
		nil,
	)
	assert.Equal(t, "test.add", props.Key())

	pool := plugins.Pools.Get(props)
	assert.NotNil(t, pool)
	vm := pool.Get()
	defer pool.Put(vm)
}

func TestPoolOk(t *testing.T) {
	jsContent := []byte(`
		function add(a, b) {
			return a + b + c;
		}
	`)

	gojaProgram, _, err := plugins.CreateGoJaProgram("", jsContent)
	assert.Nil(t, err)

	program := plugins.NewProgram(gojaProgram, "test")
	props := plugins.CreateVMProps(
		"add",
		program,
		map[string]any{
			"k1": "v1",
			"k2": "v2",
		},
	)
	assert.True(t, strings.HasPrefix(props.Key(), "test.add_"))

	pool := plugins.Pools.Get(props)
	assert.NotNil(t, pool)

	vm := pool.Get()
	defer pool.Put(vm)

	result, err := program.CallFunc("add", map[string]any{
		"c": 5,
	}, 1, 2)

	assert.Nil(t, err)
	assert.Equal(t, int64(8), result)
}

func TestPoolConcurrent(t *testing.T) {
	jsContent := []byte(`
		function add(a, b) {
			return a + b + c;
		}
	`)

	gojaProgram, _, err := plugins.CreateGoJaProgram("", jsContent)
	assert.Nil(t, err)

	program := plugins.NewProgram(gojaProgram, "test")
	props := plugins.CreateVMProps(
		"add",
		program,
		map[string]any{
			"k1": "v1",
			"k2": "v2",
		},
	)
	assert.True(t, strings.HasPrefix(props.Key(), "test.add_"))

	pool := plugins.Pools.Get(props)
	assert.NotNil(t, pool)

	for i := 0; i < 100; i++ {
		go func() {
			vm := pool.Get()
			defer pool.Put(vm)

			result, err := program.CallFunc("add", map[string]any{
				"c": 5,
			}, 1, 2)

			assert.Nil(t, err)
			assert.Equal(t, int64(8), result)
		}()
	}
}

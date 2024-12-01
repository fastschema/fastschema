package expr_test

import (
	"context"
	"os"
	"reflect"
	"testing"

	"github.com/expr-lang/expr/conf"
	fsexpr "github.com/fastschema/fastschema/expr"
	"github.com/fastschema/fastschema/pkg/entdbadapter"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
	"github.com/stretchr/testify/assert"
)

type TestResult struct {
	Value string
}

func TestResolveReturnType(t *testing.T) {
	tests := []struct {
		name       string
		resultType any
		expectKind reflect.Kind
	}{
		{
			name:       "any resultType",
			resultType: nil,
			expectKind: reflect.Invalid,
		},
		{
			name:       "bool resultType",
			resultType: true,
			expectKind: reflect.Bool,
		},
		{
			name:       "int resultType",
			resultType: 1,
			expectKind: reflect.Int,
		},
		{
			name:       "int64 resultType",
			resultType: int64(1),
			expectKind: reflect.Int64,
		},
		{
			name:       "float32 resultType",
			resultType: float32(1.0),
			expectKind: reflect.Float64,
		},
		{
			name:       "float64 resultType",
			resultType: float64(1.0),
			expectKind: reflect.Float64,
		},
		{
			name:       "struct resultType",
			resultType: TestResult{},
			expectKind: reflect.Struct,
		},
		{
			name:       "default case",
			resultType: "string",
			expectKind: reflect.String,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("testing %s", tt.name)
			config := &conf.Config{}
			option := fsexpr.ResolveReturnType(tt.resultType)
			option(config)
			assert.Equal(t, true, config.ExpectAny)
			assert.Equal(t, tt.expectKind, config.Expect)
		})
	}
}
func TestProgram(t *testing.T) {
	// Compile empty program
	compiled, err := fsexpr.Compile[any, any]("")
	assert.NoError(t, err)
	assert.Nil(t, compiled)

	// Compile simple program
	{
		program := "1 + 1"
		compiled, err := fsexpr.Compile[any, int](program)
		assert.NoError(t, err)
		assert.NotNil(t, compiled)
	}

	// Compile invalid program
	{
		program := "1 +"
		compiled, err := fsexpr.Compile[any, int](program)
		assert.Error(t, err)
		assert.Nil(t, compiled)
	}

	type contextKey string
	createContextKey := func(input string) contextKey {
		return contextKey(input)
	}
	ctx := context.WithValue(context.Background(), createContextKey("counter"), 5)
	type IntArgs struct {
		A                int
		CreateContextKey func(input string) contextKey
	}

	// Compile program with sampleArgs
	{
		program := "1 + $args.A"
		compiled, err := fsexpr.Compile[IntArgs, int](program, IntArgs{})
		assert.NoError(t, err)
		assert.NotNil(t, compiled)
	}

	// Compile program with invalid sampleArgs
	{
		program := "'a string' + $args.A"
		compiled, err := fsexpr.Compile[IntArgs, int](program, IntArgs{})
		assert.Error(t, err)
		assert.Nil(t, compiled)
	}

	// Compile program with invalid result type
	{
		program := "1 + $args.A"
		compiled, err := fsexpr.Compile[IntArgs, string](program, IntArgs{})
		assert.Error(t, err)
		assert.Nil(t, compiled)
	}

	// Run program
	{
		program := "1 + $args.A"
		compiled := utils.Must(fsexpr.Compile[IntArgs, int](program, IntArgs{}))
		result, err := compiled.Run(ctx, IntArgs{A: 1})
		assert.NoError(t, err)
		assert.Equal(t, 2, utils.Must(result.Value()))
	}

	// Run program with invalid result type
	{
		program := "$context.Value($args.CreateContextKey('counter'))"
		compiled := utils.Must(fsexpr.Compile[IntArgs, string](program))
		result, err := compiled.Run(ctx, IntArgs{A: 1, CreateContextKey: createContextKey})
		assert.NoError(t, err)
		assert.Equal(t, 5, result.Raw())
		val, err := result.Value()
		assert.Error(t, err)
		assert.Equal(t, "", val)
	}

	// Run program with error
	{
		program := "$context.InvalidMethod()"
		compiled := utils.Must(fsexpr.Compile[any, any](program))
		result, err := compiled.Run(ctx, nil)
		assert.Error(t, err)
		assert.Nil(t, result)
	}

	migrationDir := utils.Must(os.MkdirTemp("", "fastschemamigrations"))
	schemaDir := utils.Must(os.MkdirTemp("", "fastschemaschemas"))
	schemaBuilder := utils.Must(schema.NewBuilderFromDir(schemaDir))
	defer os.RemoveAll(migrationDir)
	defer os.RemoveAll(schemaDir)

	// Run program with DB operation
	{
		program := `
		let counter = $context.Value($args.CreateContextKey('counter'));
		let _ = $db.Exec($context, 'CREATE TABLE IF NOT EXISTS test (id INT PRIMARY KEY)');
		let records = $db.Query($context, 'SELECT 1 as counter');
		records[0].Get('counter') + counter
		`
		compiled := utils.Must(fsexpr.Compile[IntArgs, any](program))
		result, err := compiled.Run(ctx, IntArgs{
			CreateContextKey: createContextKey,
		}, fsexpr.Config{
			DB: func() fsexpr.DBLike {
				return utils.Must(entdbadapter.NewTestClient(migrationDir, schemaBuilder))
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, 6, utils.Must(result.Value()))
	}

	// Run program that return undefined
	{
		program := "$undefined"
		compiled := utils.Must(fsexpr.Compile[any, any](program))
		result, err := compiled.Run(ctx, nil)
		assert.NoError(t, err)
		assert.True(t, result.IsUndefined())
	}
}

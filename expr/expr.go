package expr

import (
	"context"
	"reflect"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
	"github.com/fastschema/fastschema/pkg/utils"
)

// Program is a typed wrapper around the expr VM program.
// It holds the compiled expression and ensures type safety for execution.
type Program[ArgsType any, ResultType any] struct {
	*vm.Program
}

// Config holds configuration options for expression execution,
// such as database access.
type Config struct {
	DB func() DBLike
}

// Compile parses and compiles the given expression string into a Program.
//
//	program: The source code of the expression.
//	sampleArgs: Optional instance of ArgsType to aid in type inference during compilation.
func Compile[ArgsType any, ResultType any](
	program string,
	sampleArgs ...ArgsType,
) (*Program[ArgsType, ResultType], error) {
	if program == "" {
		return nil, nil
	}
	var argsType ArgsType
	var resultType ResultType

	if len(sampleArgs) > 0 {
		argsType = sampleArgs[0]
	}

	exprEnv := NewEnv(context.Background(), argsType)
	p, err := expr.Compile(
		program,
		expr.Env(exprEnv),
		ResolveReturnType(resultType),
	)
	if err != nil {
		return nil, err
	}

	return &Program[ArgsType, ResultType]{Program: p}, nil
}

// Run executes the compiled program with the given context and arguments.
func (p *Program[ArgsType, ResultType]) Run(
	ctx context.Context,
	args ArgsType,
	configs ...Config,
) (*Result[ResultType], error) {
	exprEnv := NewEnv(ctx, args, configs...)
	result, err := vm.Run(p.Program, exprEnv)
	if err != nil {
		return nil, err
	}

	return &Result[ResultType]{raw: result}, nil
}

func ResolveReturnType(resultType any) expr.Option {
	dereferencedType := utils.GetDereferencedType(resultType)
	if dereferencedType == nil {
		return expr.AsAny()
	}

	switch dereferencedType.Kind() {
	case reflect.Bool:
		return expr.AsBool()
	case reflect.Int:
		return expr.AsInt()
	case reflect.Int64:
		return expr.AsInt64()
	case reflect.Float32, reflect.Float64:
		return expr.AsFloat64()
	default:
		reflectType := reflect.TypeOf(resultType)
		return expr.AsKind(reflectType.Kind())
	}
}

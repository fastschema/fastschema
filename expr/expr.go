package expr

import (
	"context"
	"reflect"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
	"github.com/fastschema/fastschema/pkg/utils"
)

type Program[ArgsType any, ResultType any] struct {
	*vm.Program
}

type Config struct {
	DB func() DBLike
}

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
	dereferecedType := utils.GetDereferencedType(resultType)

	if dereferecedType == nil {
		return expr.AsAny()
	}

	switch dereferecedType.Kind() {
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
		if reflectType == nil {
			return expr.AsAny()
		}
		return expr.AsKind(reflectType.Kind())
	}
}

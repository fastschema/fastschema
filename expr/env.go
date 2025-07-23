package expr

import (
	"context"
	"fmt"

	"github.com/fastschema/fastschema/pkg/utils"
)

type Undefined struct{}

type Result[T any] struct {
	raw any
}

func (r *Result[T]) IsUndefined() bool {
	_, ok := r.raw.(*Undefined)
	return ok
}

func (r *Result[T]) Raw() any {
	return r.raw
}

func (r *Result[T]) Value() (T, error) {
	t, ok := r.raw.(T)
	if !ok {
		return t, fmt.Errorf(`raw value "%v" is not of type %T`, r.raw, t)
	}

	return t, nil
}

type Env[ArgType any] struct {
	// TODO:
	// Using 'any' instead of context.Context to avoid interface method checks by expr-lang/expr.
	// The actual value passed is fs.Context which has additional methods beyond context.Context.
	Context   any                          `expr:"$context"`
	Args      ArgType                      `expr:"$args"`
	DB        *DB                          `expr:"$db"`
	Undefined *Undefined                   `expr:"$undefined"`
	Sprintf   func(string, ...any) string  `expr:"$sprintf"`
	Hash      func(string) (string, error) `expr:"$hash"`
}

func NewEnv[ArgType any](
	ctx context.Context,
	args ArgType,
	configs ...Config,
) *Env[ArgType] {
	env := &Env[ArgType]{
		Context:   ctx,
		Args:      args,
		Undefined: &Undefined{},
		Sprintf:   fmt.Sprintf,
		Hash:      utils.GenerateHash,
	}

	if len(configs) > 0 && configs[0].DB != nil {
		env.DB = NewDB(configs[0].DB)
	}

	return env
}

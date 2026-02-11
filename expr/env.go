package expr

import (
	"context"
	"fmt"

	"github.com/fastschema/fastschema/pkg/utils"
)

// Undefined represents a value that is not defined.
type Undefined struct{}

// Result holds the output of an executed program.
type Result[T any] struct {
	raw any
}

// IsUndefined checks if the result is an Undefined value.
func (r *Result[T]) IsUndefined() bool {
	_, ok := r.raw.(*Undefined)
	return ok
}

// Raw returns the underlying raw value of the result.
func (r *Result[T]) Raw() any {
	return r.raw
}

// Value returns the typed value of the result.
// It returns an error if the raw value cannot be asserted to type T.
func (r *Result[T]) Value() (T, error) {
	t, ok := r.raw.(T)
	if !ok {
		return t, fmt.Errorf(`raw value "%v" is not of type %T`, r.raw, t)
	}

	return t, nil
}

// Env is the execution environment for the expression.
// It exposes variables and functions to the script.
type Env[ArgType any] struct {
	// Context is defined as `any` instead of `context.Context` to avoid interface method checks by expr-lang/expr.
	// The actual value passed is fs.Context (or wrappers) which has additional methods beyond context.Context.
	// This allows the expression language to access methods on the specific context implementation.
	Context   any                          `expr:"$context"`
	Args      ArgType                      `expr:"$args"`
	DB        *DB                          `expr:"$db"`
	Undefined *Undefined                   `expr:"$undefined"`
	Sprintf   func(string, ...any) string  `expr:"$sprintf"`
	Hash      func(string) (string, error) `expr:"$hash"`
}

// NewEnv creates a new execution environment with the given context, arguments, and optional configuration.
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

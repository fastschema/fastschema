package fs

import (
	"context"
	"net/http"

	"github.com/fastschema/fastschema/logger"
	"github.com/fastschema/fastschema/pkg/errors"
	"github.com/fastschema/fastschema/schema"
)

// Handler is a function that resolves a request
type Handler func(c Context) (any, error)

// Context is the interface that defines the methods that a context must implement
type Context interface {
	ID() string
	User() *User
	Value(string, ...any) (val any)
	Logger() logger.Logger
	Parse(any) error
	Context() context.Context
	Args() map[string]string
	Arg(string, ...string) string
	ArgInt(string, ...int) int
	Entity() (*schema.Entity, error)
	Resource() *Resource
	AuthToken() string
	Next() error
	Result(...*Result) *Result
	Files() ([]*File, error)
}

type HTTPResponse struct {
	StatusCode int
	Body       []byte
	Header     http.Header
}

// Result is a struct that contains the result of a resolver
type Result struct {
	Error *errors.Error `json:"error,omitempty"`
	Data  any           `json:"data,omitempty"`
}

// NewResult creates a new result struct
func NewResult(data any, err error) *Result {
	result := &Result{Data: data}

	if err != nil {
		if !errors.Is(err, &errors.Error{}) {
			result.Error = errors.From(err)
		} else {
			result.Error = err.(*errors.Error)
		}

		return result
	}

	result.Data = data

	return result
}

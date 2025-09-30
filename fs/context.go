package fs

import (
	"bytes"
	"context"
	"net/http"

	"github.com/fastschema/fastschema/entity"
	"github.com/fastschema/fastschema/logger"
	"github.com/fastschema/fastschema/pkg/errors"
)

// Handler is a function that resolves a request
type Handler func(c Context) (any, error)

// Context is the interface that defines the methods that a context must implement
type Context interface {
	context.Context
	TraceID() string
	User() *User
	Local(string, ...any) (val any)
	Logger() logger.Logger
	Bind(any) error
	SetArg(string, string) string
	Args() map[string]string
	Arg(name string, defaults ...string) string
	ArgInt(name string, defaults ...int) int
	Body() ([]byte, error)
	Payload() (*entity.Entity, error)
	BodyParser(out any) error
	FormValue(key string, defaultValue ...string) string
	Resource() *Resource
	AuthToken() string
	Next() error
	Result(...*Result) *Result
	Files() ([]*File, error)
	Redirect(string) error
	WSClient() WSClient
	IP() string
}

type HTTPResponse struct {
	StatusCode int
	Body       []byte
	Header     http.Header
	File       string
	Stream     *bytes.Buffer
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
		var fsError *errors.Error
		if !errors.As(err, &fsError) {
			result.Error = errors.From(err)
		} else {
			result.Error = fsError
		}

		return result
	}

	result.Data = data

	return result
}

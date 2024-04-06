package app

import (
	"context"
	"math"

	"github.com/fastschema/fastschema/logger"
	"github.com/fastschema/fastschema/pkg/errors"
	"github.com/fastschema/fastschema/schema"
)

// Resolver is a function that resolves a request
type Resolver func(c Context) (any, error)

// Meta hold extra data, ex: request method, path, etc
type Meta map[string]any

// Map is a shortcut for map[string]any
type Map map[string]any

// Signature hold the input and output types of a resolver
type Signature = [2]any

const POST = "POST"
const GET = "GET"
const PUT = "PUT"
const DELETE = "DELETE"
const PATCH = "PATCH"
const OPTIONS = "OPTIONS"

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

// Result is a struct that contains the result of a resolver
type Result struct {
	Error *errors.Error `json:"error,omitempty"`
	Data  any           `json:"data,omitempty"`
}

// PaginationInfo is a struct that contains pagination data
type PaginationInfo struct {
	Total       uint `json:"total"`
	PerPage     uint `json:"per_page"`
	CurrentPage uint `json:"current_page"`
	LastPage    uint `json:"last_page"`
}

// Pagination is a struct that contains pagination info and the data
type Pagination struct {
	Pagination *PaginationInfo `json:"pagination"`
	Data       any             `json:"data"`
}

// NewPagination creates a new pagination struct
func NewPagination(total, perPage, currentPage uint, data any) *Pagination {
	return &Pagination{
		Pagination: &PaginationInfo{
			Total:       total,
			PerPage:     perPage,
			CurrentPage: currentPage,
			LastPage:    uint(math.Ceil(float64(total) / float64(perPage))),
		},
		Data: data,
	}
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

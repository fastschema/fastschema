package app

import (
	"context"
	"math"

	"github.com/fastschema/fastschema/logger"
	"github.com/fastschema/fastschema/pkg/errors"
	"github.com/fastschema/fastschema/schema"
)

type Resolver func(c Context) (any, error)

type Meta map[string]any

type Map map[string]any

type Signature = [2]any

const POST = "POST"
const GET = "GET"
const PUT = "PUT"
const DELETE = "DELETE"
const PATCH = "PATCH"
const OPTIONS = "OPTIONS"

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

type Result struct {
	Error *errors.Error `json:"error,omitempty"`
	Data  any           `json:"data,omitempty"`
}

type PaginationData struct {
	Total       uint `json:"total"`
	PerPage     uint `json:"per_page"`
	CurrentPage uint `json:"current_page"`
	LastPage    uint `json:"last_page"`
}
type Pagination struct {
	Pagination *PaginationData `json:"pagination"`
	Data       any             `json:"data"`
}

func NewPagination(total, perPage, currentPage uint, data any) *Pagination {
	return &Pagination{
		Pagination: &PaginationData{
			Total:       total,
			PerPage:     perPage,
			CurrentPage: currentPage,
			LastPage:    uint(math.Ceil(float64(total) / float64(perPage))),
		},
		Data: data,
	}
}

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

package contentservice

import (
	"math"
	"strings"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/entity"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/errors"
)

// Pagination is a struct that contains pagination info and the data.
type Pagination struct {
	Total       uint             `json:"total"`
	PerPage     uint             `json:"per_page"`
	CurrentPage uint             `json:"current_page"`
	LastPage    uint             `json:"last_page"`
	Items       []*entity.Entity `json:"items"`
}

// NewPagination creates a new pagination struct.
func NewPagination(total, perPage, currentPage uint, items []*entity.Entity) *Pagination {
	return &Pagination{
		Total:       total,
		PerPage:     perPage,
		CurrentPage: currentPage,
		LastPage:    uint(math.Ceil(float64(total) / float64(perPage))),
		Items:       items,
	}
}

func (cs *ContentService) List(c fs.Context, _ any) (*Pagination, error) {
	model, err := cs.DB().Model(c.Arg("schema"))
	if err != nil {
		return nil, errors.BadRequest(err.Error())
	}

	predicates, err := db.CreatePredicatesFromFilterObject(
		cs.DB().SchemaBuilder(),
		model.Schema(),
		c.Arg("filter", ""),
	)
	if err != nil {
		return nil, errors.BadRequest(err.Error())
	}

	columns := []string{}
	total, err := model.Query(predicates...).Count(c, &db.QueryOption{})
	if err != nil {
		return nil, errors.BadRequest(err.Error())
	}

	if fields := c.Arg("select", ""); fields != "" {
		columns = strings.Split(fields, ",")
	} else if model.Schema().Name == "user" {
		columns = []string{"roles"}
	}

	page := uint(c.ArgInt("page", 1))
	limit := uint(c.ArgInt("limit", 10))
	records, err := model.Query(predicates...).
		Select(columns...).
		Limit(uint(c.ArgInt("limit", 10))).
		Offset((page - 1) * limit).
		Order(c.Arg("sort", "-id")).
		Get(c)

	if err != nil {
		return nil, errors.InternalServerError(err.Error())
	}

	return NewPagination(uint(total), limit, page, records), nil
}

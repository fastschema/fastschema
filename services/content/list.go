package contentservice

import (
	"math"
	"strings"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/errors"
	"github.com/fastschema/fastschema/schema"
)

// Pagination is a struct that contains pagination info and the data
type Pagination struct {
	Total       uint             `json:"total"`
	PerPage     uint             `json:"per_page"`
	CurrentPage uint             `json:"current_page"`
	LastPage    uint             `json:"last_page"`
	Items       []*schema.Entity `json:"items"`
}

// NewPagination creates a new pagination struct
func NewPagination(total, perPage, currentPage uint, items []*schema.Entity) *Pagination {
	return &Pagination{
		Total:       total,
		PerPage:     perPage,
		CurrentPage: currentPage,
		LastPage:    uint(math.Ceil(float64(total) / float64(perPage))),
		Items:       items,
	}
}

func (cs *ContentService) List(c fs.Context, _ any) (*Pagination, error) {
	schemaName := c.Arg("schema")
	model, err := cs.DB().Model(schemaName)
	if err != nil {
		return nil, errors.BadRequest(err.Error())
	}

	filter := c.Arg("filter")
	predicates, err := db.CreatePredicatesFromFilterObject(
		cs.DB().SchemaBuilder(),
		model.Schema(),
		filter,
	)

	if err != nil {
		return nil, errors.BadRequest(err.Error())
	}

	sort := c.Arg("sort", "-id")
	columns := []string{}
	page := uint(c.ArgInt("page", 1))
	limit := uint(c.ArgInt("limit", 10))
	offset := (page - 1) * limit
	total, err := model.Query(predicates...).Count(c.Context(), &db.CountOption{})

	if err != nil {
		return nil, errors.BadRequest(err.Error())
	}

	if fields := c.Arg("select", ""); fields != "" {
		columns = strings.Split(fields, ",")
	} else if schemaName == "user" {
		columns = []string{"id", "username", "email", "roles", "active", "provider", "created_at", "updated_at"}
	}

	records, err := model.Query(predicates...).
		Select(columns...).
		Limit(limit).
		Offset(offset).
		Order(sort).
		Get(c.Context())

	if err != nil {
		return nil, errors.InternalServerError(err.Error())
	}

	return NewPagination(uint(total), limit, page, records), nil
}

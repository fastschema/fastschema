package contentservice

import (
	"strings"

	"github.com/fastschema/fastschema/app"
	"github.com/fastschema/fastschema/pkg/errors"
)

func (cs *ContentService) List(c app.Context, _ *any) (*app.Pagination, error) {
	schemaName := c.Arg("schema")
	model, err := cs.app.DB().Model(schemaName)
	if err != nil {
		return nil, errors.BadRequest(err.Error())
	}

	filter := c.Arg("filter")
	predicates, err := app.CreatePredicatesFromFilterObject(
		cs.app.DB().SchemaBuilder(),
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
	total, err := model.Query(predicates...).Count(&app.CountOption{})

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

	return app.NewPagination(uint(total), limit, page, records), nil
}

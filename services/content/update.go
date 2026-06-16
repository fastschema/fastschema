package contentservice

import (
	"strings"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/entity"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/errors"
)

// isValidationError checks if the error is a client input validation error
// that should return 400 Bad Request instead of 500 Internal Server Error
func isValidationError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	// Invalid column/field errors are client input errors
	return strings.Contains(msg, "column") && strings.Contains(msg, "not found")
}

func (cs *ContentService) Update(c fs.Context, _ any) (*entity.Entity, error) {
	schemaName := c.Arg("schema")
	model, err := cs.DB().Model(schemaName)
	if err != nil {
		return nil, errors.BadRequest(err.Error())
	}

	idValue, err := parseIDArg(model.Schema(), c.Arg("id"))
	if err != nil {
		return nil, errors.NotFound(err.Error())
	}

	pkName := model.Schema().PrimaryKeyName()
	entity, err := c.Payload()
	if err != nil {
		return nil, errors.BadRequest(err.Error())
	}

	entity.SetIDField(pkName)
	if _, err := model.Mutation().Where(db.EQ(pkName, idValue)).Update(c, entity); err != nil {
		if isValidationError(err) {
			return nil, errors.BadRequest(err.Error())
		}
		return nil, errors.InternalServerError(err.Error())
	}

	if err := entity.SetID(idValue); err != nil {
		return nil, errors.InternalServerError(err.Error())
	}

	return entity.Delete("password"), nil
}

func (cs *ContentService) BulkUpdate(c fs.Context, _ any) (int, error) {
	model, err := cs.DB().Model(c.Arg("schema"))
	if err != nil {
		return 0, errors.BadRequest(err.Error())
	}

	predicates, err := db.CreatePredicatesFromFilterObject(
		cs.DB().SchemaBuilder(),
		model.Schema(),
		c.Arg("filter"),
	)
	if err != nil {
		return 0, errors.BadRequest(err.Error())
	}

	entity, err := c.Payload()
	if err != nil {
		return 0, errors.BadRequest(err.Error())
	}
	if entity.Empty() {
		return 0, nil
	}

	updatedCount, err := model.Mutation().Where(predicates...).Update(c, entity)
	if err != nil {
		if isValidationError(err) {
			return 0, errors.BadRequest(err.Error())
		}
		return 0, errors.InternalServerError(err.Error())
	}

	return updatedCount, nil
}

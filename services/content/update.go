package contentservice

import (
	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/errors"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
)

func (cs *ContentService) Update(c fs.Context, _ any) (*schema.Entity, error) {
	schemaName := c.Arg("schema")
	id := c.ArgInt("id")
	model, err := cs.DB().Model(schemaName)
	if err != nil {
		return nil, errors.BadRequest(err.Error())
	}

	entity, err := c.Payload()
	if err != nil {
		return nil, errors.BadRequest(err.Error())
	}

	if schemaName == "user" {
		password := entity.GetString("password")
		if password == "" {
			entity.Delete("password")
		} else {
			hash, err := utils.GenerateHash(password)
			if err != nil {
				return nil, errors.BadRequest(err.Error())
			}

			entity.Set("password", hash)
		}
	}

	if _, err := model.Mutation().Where(db.EQ("id", id)).Update(c, entity); err != nil {
		return nil, errors.InternalServerError(err.Error())
	}

	return entity.SetID(id).Delete("password"), nil
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
		return 0, errors.InternalServerError(err.Error())
	}

	return updatedCount, nil
}

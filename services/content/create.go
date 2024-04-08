package contentservice

import (
	"github.com/fastschema/fastschema/app"
	"github.com/fastschema/fastschema/pkg/errors"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
)

func (cs *ContentService) Create(c app.Context, _ *any) (*schema.Entity, error) {
	schemaName := c.Arg("schema")
	model, err := cs.DB().Model(schemaName)
	if err != nil {
		return nil, errors.BadRequest(err.Error())
	}

	entity, err := c.Entity()
	if err != nil {
		return nil, errors.BadRequest(err.Error())
	}

	if schemaName == "user" {
		password := entity.GetString("password")
		if password == "" {
			return nil, errors.BadRequest("password is required")
		}

		hash, err := utils.GenerateHash(password)
		if err != nil {
			return nil, errors.BadRequest(err.Error())
		}

		entity.Set("password", hash)
	}

	if _, err := model.Create(entity); err != nil {
		return nil, errors.BadRequest(err.Error())
	}

	return entity, nil
}

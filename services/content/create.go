package contentservice

import (
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/errors"
	// "github.com/fastschema/fastschema/schema"
)

func (cs *ContentService) Create(c fs.Context, _ any) (any, error) {
	schemaName := c.Arg("schema")
	model, err := cs.DB().Model(schemaName)
	if err != nil {
		return nil, errors.BadRequest(err.Error())
	}

	entity, err := c.Payload()
	if err != nil {
		return nil, errors.BadRequest(err.Error())
	}

	if _, err := model.Create(c, entity); err != nil {
		return nil, errors.BadRequest(err.Error())
	}

	return entity.Delete("password"), nil
}

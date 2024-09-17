package contentservice

import (
	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/errors"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
)

func isDeletable(id int, schemaName string) error {
	if schemaName == "user" && id == 1 {
		return errors.BadRequest("Cannot delete root user.")
	}

	return nil
}

func (cs *ContentService) Delete(c fs.Context, _ any) (any, error) {
	schemaName := c.Arg("schema")
	id := c.ArgInt("id")

	if err := isDeletable(id, schemaName); err != nil {
		return nil, err
	}

	model, err := cs.DB().Model(schemaName)
	if err != nil {
		return nil, errors.BadRequest(err.Error())
	}

	_, err = model.Query(db.EQ("id", id)).Only(c)

	if err != nil {
		e := utils.If(db.IsNotFound(err), errors.NotFound, errors.InternalServerError)
		return nil, e(err.Error())
	}

	if _, err := model.Mutation().Where(db.EQ("id", id)).Delete(c); err != nil {
		return nil, errors.BadRequest(err.Error())
	}

	return schema.NewEntity(uint64(id)), nil
}

func (cs *ContentService) BulkDelete(c fs.Context, _ any) (int, error) {
	schemaName := c.Arg("schema")
	model, err := cs.DB().Model(schemaName)
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

	records, err := model.Query(predicates...).Get(c)
	if err != nil {
		return 0, errors.BadRequest(err.Error())
	}

	if len(records) == 0 {
		return 0, nil
	}

	var ids []any
	for _, record := range records {
		recordID := record.ID()
		if err := isDeletable(int(recordID), schemaName); err != nil {
			return 0, errors.InternalServerError("Cannot delete root user.")
		}
		ids = append(ids, int(recordID))
	}

	recordDelete, err := model.Mutation().Where(db.In("id", ids)).Delete(c)
	if err != nil {
		return 0, errors.InternalServerError(err.Error())
	}

	return recordDelete, nil
}

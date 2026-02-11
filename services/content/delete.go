package contentservice

import (
	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/entity"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/errors"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
)

func isDeletable(id any, schemaName string) error {
	if schemaName != "user" {
		return nil
	}

	rootID, err := utils.AnyToUint[uint64](id)
	if err == nil && rootID == 1 {
		return errors.BadRequest("Cannot delete root user.")
	}

	return nil
}

func (cs *ContentService) Delete(c fs.Context, _ any) (any, error) {
	schemaName := c.Arg("schema")
	model, err := cs.DB().Model(schemaName)
	if err != nil {
		return nil, errors.BadRequest(err.Error())
	}

	idValue, err := parseIDArg(model.Schema(), c.Arg("id"))
	if err != nil {
		return nil, errors.BadRequest(err.Error())
	}

	if err := isDeletable(idValue, schemaName); err != nil {
		return nil, err
	}

	_, err = model.Query(db.EQ(entity.FieldID, idValue)).Only(c)

	if err != nil {
		e := utils.If(db.IsNotFound(err), errors.NotFound, errors.InternalServerError)
		return nil, e(err.Error())
	}

	if _, err := model.Mutation().Where(db.EQ(entity.FieldID, idValue)).Delete(c); err != nil {
		return nil, errors.BadRequest(err.Error())
	}

	return entity.New(idValue), nil
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

	ids := make([]any, 0, len(records))
	for _, record := range records {
		recordID := record.ID()
		if err := isDeletable(recordID, schemaName); err != nil {
			return 0, errors.InternalServerError("Cannot delete root user.")
		}
		ids = append(ids, recordID)
	}

	recordDelete, err := model.Mutation().Where(db.In("id", ids)).Delete(c)
	if err != nil {
		return 0, errors.InternalServerError(err.Error())
	}

	return recordDelete, nil
}

func parseIDArg(s *schema.Schema, rawID string) (any, error) {
	if rawID == "" {
		return nil, errors.BadRequest("missing id")
	}

	if s == nil {
		return rawID, nil
	}

	idField := s.IDField()
	if idField == nil {
		return rawID, nil
	}

	return schema.StringToFieldValue[any](idField, rawID)
}

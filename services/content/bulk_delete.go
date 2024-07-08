package contentservice

import (
	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/errors"
)

func (cs *ContentService) BulkDelete(c fs.Context, _ any) (any, error) {
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

	records, err := model.Query(predicates...).
		Get(c.Context())

	if err != nil {
		return nil, errors.BadRequest(err.Error())
	}

	if len(records) == 0 {
		return nil, errors.NotFound("no entities found")
	}

	var ids []int
	for _, record := range records {
		recordID := record.ID()
		if err := isDeletable(int(recordID), schemaName); err != nil {
			return nil, errors.InternalServerError("Cannot delete root user.")
		}
		ids = append(ids, int(recordID))
	}

	anyIDs := make([]any, len(ids))
	for i, id := range ids {
		anyIDs[i] = any(id)
	}

	if _, err := model.Mutation().Where(db.In("id", anyIDs)).Delete(c.Context()); err != nil {
		return nil, errors.InternalServerError(err.Error())
	}

	return true, nil
}

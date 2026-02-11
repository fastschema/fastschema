package contentservice

import (
	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/entity"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/errors"
)

// isRootUser checks if the given entity is a root user by checking if any of their roles have Root: true
func isRootUser(e *entity.Entity) bool {
	if e == nil {
		return false
	}

	rolesValue := e.Get("roles")
	if rolesValue == nil {
		return false
	}

	roles, ok := rolesValue.([]*entity.Entity)
	if !ok {
		return false
	}

	for _, role := range roles {
		if root, ok := role.Get("root").(bool); ok && root {
			return true
		}
	}

	return false
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

	primaryKeyName := model.Schema().PrimaryKeyName()
	query := model.Query(db.EQ(primaryKeyName, idValue))
	// For users, we need to select roles to check if they are root users
	if schemaName == "user" {
		query = query.Select("roles")
	}

	record, err := query.Only(c)
	if err != nil {
		e := errors.NotFound
		if !db.IsNotFound(err) {
			e = errors.InternalServerError
		}
		return nil, e(err.Error())
	}

	// Check if trying to delete a root user
	if schemaName == "user" && isRootUser(record) {
		return nil, errors.BadRequest("Cannot delete root user.")
	}

	if _, err := model.Mutation().Where(db.EQ(primaryKeyName, idValue)).Delete(c); err != nil {
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

	query := model.Query(predicates...)
	// For users, we need to select roles to check if they are root users
	if schemaName == "user" {
		query = query.Select("roles")
	}

	records, err := query.Get(c)
	if err != nil {
		return 0, errors.BadRequest(err.Error())
	}

	if len(records) == 0 {
		return 0, nil
	}

	primaryKeyName := model.Schema().PrimaryKeyName()
	ids := make([]any, 0, len(records))
	for _, record := range records {
		// Check if trying to delete a root user
		if schemaName == "user" && isRootUser(record) {
			return 0, errors.BadRequest("Cannot delete root user.")
		}
		ids = append(ids, record.ID())
	}

	recordDelete, err := model.Mutation().
		Where(db.In(primaryKeyName, ids)).
		Delete(c)
	if err != nil {
		return 0, errors.InternalServerError(err.Error())
	}

	return recordDelete, nil
}

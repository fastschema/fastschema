package schemaservice

import (
	"os"

	"github.com/fastschema/fastschema/app"
	"github.com/fastschema/fastschema/pkg/errors"
)

func (ss *SchemaService) Delete(c app.Context, _ *any) (app.Map, error) {
	schemaName := c.Arg("name")
	s, err := ss.app.SchemaBuilder().Schema(schemaName)
	if err != nil {
		return nil, errors.NotFound(err.Error())
	}

	// check if the schema has any relation
	// if it has, then we can't delete it
	// skip relation type check if the relation type is bi-directional
	hasRelation := false
	for _, field := range s.Fields {
		if field.Type.IsRelationType() && field.Relation.TargetSchemaName != schemaName {
			hasRelation = true
			break
		}
	}

	if hasRelation {
		return nil, errors.BadRequest("schema has relation, can't delete")
	}

	// delete the schema file
	schemaFile := ss.app.SchemaBuilder().SchemaFile(schemaName)
	if err := os.Remove(schemaFile); err != nil {
		return nil, errors.InternalServerError(err.Error())
	}

	if err := ss.app.Reload(nil); err != nil {
		return nil, errors.InternalServerError(err.Error())
	}

	return app.Map{"message": "Schema deleted"}, nil
}

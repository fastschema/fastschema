package schemaservice

import (
	"errors"
	"os"

	"github.com/fastschema/fastschema/app"
)

func (ss *SchemaService) Delete(c app.Context, _ *any) (app.Map, error) {
	schemaName := c.Arg("name")
	s, err := ss.app.SchemaBuilder().Schema(schemaName)
	if err != nil {
		return nil, err
	}

	// check if the schema has any relation
	// if it has, then we can't delete it
	hasRelation := false
	for _, field := range s.Fields {
		if field.Type.IsRelationType() {
			hasRelation = true
			break
		}
	}

	if hasRelation {
		return nil, errors.New("schema has relation, can't delete")
	}

	// delete the schema file
	schemaFile := ss.app.SchemaBuilder().SchemaFile(schemaName)
	if err := os.Remove(schemaFile); err != nil {
		return nil, errors.New("could not delete schema file")
	}

	if err := ss.app.Reload(nil); err != nil {
		return nil, err
	}

	return app.Map{"message": "Schema deleted"}, nil
}

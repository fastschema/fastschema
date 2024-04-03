package schemaservice

import (
	"sort"

	"github.com/fastschema/fastschema/app"
	"github.com/fastschema/fastschema/schema"
)

func (ss *SchemaService) List(c app.Context, _ *any) ([]*schema.Schema, error) {
	builder, err := schema.NewBuilderFromDir(ss.app.SchemaBuilder().Dir())
	if err != nil {
		return nil, err
	}

	schemas := builder.Schemas()
	sort.Slice(schemas, func(i, j int) bool {
		if schemas[i].IsSystemSchema != schemas[j].IsSystemSchema {
			return schemas[i].IsSystemSchema
		}

		return schemas[i].Name > schemas[j].Name
	})

	for i, j := 0, len(schemas)-1; i < j; i, j = i+1, j-1 {
		schemas[i], schemas[j] = schemas[j], schemas[i]
	}

	return schemas, nil
}

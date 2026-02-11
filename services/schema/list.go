package schemaservice

import (
	"sort"

	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
)

var ignoreContentSchemas = []string{
	// "permission",
	"migration",
	"session",
}

func (ss *SchemaService) List(c fs.Context, _ any) ([]*schema.Schema, error) {
	schemas := ss.app.SchemaBuilder().Schemas()
	sort.Slice(schemas, func(i, j int) bool {
		if schemas[i].IsSystemSchema != schemas[j].IsSystemSchema {
			return schemas[i].IsSystemSchema
		}

		return schemas[i].Name > schemas[j].Name
	})

	for i, j := 0, len(schemas)-1; i < j; i, j = i+1, j-1 {
		schemas[i], schemas[j] = schemas[j], schemas[i]
	}

	// Clone schemas to filter out auto-generated fields
	clonedSchemas := make([]*schema.Schema, 0, len(schemas))
	for _, s := range schemas {
		if utils.Contains(ignoreContentSchemas, s.Name) || s.IsJunctionSchema {
			continue
		}

		clonedSchemas = append(clonedSchemas, s.Clone())
		// Filter out system fields
		// clonedSchemas[i].Fields = utils.Filter(clonedSchemas[i].Fields, func(field *schema.Field) bool {
		// 	return !field.IsSystemField
		// })
	}

	return clonedSchemas, nil
}

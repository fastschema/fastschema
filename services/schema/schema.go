package schemaservice

import (
	"github.com/fastschema/fastschema/app"
	"github.com/fastschema/fastschema/schema"
)

type CreateDBAdapterFunc func(
	config *app.DBConfig,
	schemaBuilder *schema.Builder,
) (app.DBClient, error)

type SchemaService struct {
	app app.App
}

func NewSchemaService(app app.App) *SchemaService {
	ss := &SchemaService{app}

	return ss
}

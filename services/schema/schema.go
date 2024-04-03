package schemaservice

import (
	"github.com/fastschema/fastschema/app"
	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/schema"
)

type CreateDBAdapterFunc func(
	config *db.DBConfig,
	schemaBuilder *schema.Builder,
) (db.Client, error)

type SchemaService struct {
	app app.App
}

func NewSchemaService(app app.App) *SchemaService {
	ss := &SchemaService{app}

	return ss
}

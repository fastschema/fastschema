package schemaservice

import (
	"github.com/fastschema/fastschema/app"
	"github.com/fastschema/fastschema/schema"
)

type AppLike interface {
	DB() app.DBClient
	Reload(migration *app.Migration) error
	SchemaBuilder() *schema.Builder
}

type SchemaService struct {
	app AppLike
}

func New(app AppLike) *SchemaService {
	ss := &SchemaService{app}

	return ss
}
